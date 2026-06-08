package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ubinix-warun/wolllama/cli/internal/ollama"
	"github.com/ubinix-warun/wolllama/pkg/manifest"
	"github.com/ubinix-warun/wolllama/pkg/storage"
)

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().Int("epochs", 0, "Number of storage epochs (overrides config)")
}

var pushCmd = &cobra.Command{
	Use:   "push <model:tag>",
	Short: "Upload a model from Ollama to Walrus",
	Long: `Push uploads a model from your local Ollama store to Walrus decentralized storage.

Reads the model's manifest and blobs from ~/.ollama/models/,
uploads each blob to the Walrus Publisher, constructs a Wolllama
manifest, and stores it on Walrus.

The printed manifest object ID is the unique identifier — share it
so others can pull the model, or submit it to the Wolllama site
to make it discoverable.

Examples:
  wolllama push llama3.2:latest
  wolllama push llama3.2:3b-q4_K_M
  wolllama push mistral:7b --epochs 20`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
	modelTag := args[0]

	// --- 1. Open Ollama store ---
	ollamaPath := viper.GetString("ollama_path")
	store, err := ollama.NewStore(ollamaPath)
	if err != nil {
		return fmt.Errorf("open ollama store: %w", err)
	}

	// --- 2. Read the Ollama manifest ---
	entry, err := store.ReadManifest(modelTag)
	if err != nil {
		return fmt.Errorf("read model %q: %w", modelTag, err)
	}

	verbose := viper.GetBool("verbose")

	// Collect all blobs to upload: config + layers
	var blobs []blobToUpload
	if entry.Config != nil {
		blobs = append(blobs, blobToUpload{
			digest:    entry.Config.Digest,
			mediaType: entry.Config.MediaType,
			size:      entry.Config.Size,
		})
	}
	for _, layer := range entry.Layers {
		blobs = append(blobs, blobToUpload{
			digest:    layer.Digest,
			mediaType: layer.MediaType,
			size:      layer.Size,
		})
	}

	// --- 3. Initialize storage provider ---
	providerName := viper.GetString("provider")
	if flagProvider, _ := cmd.Flags().GetString("provider"); flagProvider != "" {
		providerName = flagProvider
	}

	epochs := viper.GetInt("epochs")
	if flagEpochs, _ := cmd.Flags().GetInt("epochs"); flagEpochs > 0 {
		epochs = flagEpochs
	}

	pubURL, aggURL := walrusURLs()
	provider, err := storage.New(providerName, storage.Config{
		PublisherURL:  pubURL,
		AggregatorURL: aggURL,
		Epochs:        epochs,
		TatumAPIKey:   viper.GetString("tatum_api_key"),
		TatumAPIURL:   viper.GetString("tatum_api_url"),
	})
	if err != nil {
		return fmt.Errorf("init storage provider: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Pushing %s (%d blobs, %s) via %s...\n",
		modelTag, len(blobs), formatBytes(totalSize(blobs)), provider.Name())

	// --- 4. Upload blobs sequentially ---
	// Chunk size is determined by the provider's limit.
	chunkSize := provider.MaxChunkSize()
	if chunkSize <= 0 || chunkSize > 500_000_000 {
		chunkSize = 256_000_000
	}

	blobMap := make(map[string]manifest.BlobRef, len(blobs)) // digest → walrus reference
	for i, b := range blobs {
		label := fmt.Sprintf("[%d/%d] %s", i+1, len(blobs), shortDigest(b.digest))

		if b.size <= chunkSize {
			// Single upload for blobs within provider limit
			fmt.Fprintf(os.Stderr, "%s (%s)... ", label, formatBytes(b.size))
			data, err := store.ReadBlob(b.digest)
			if err != nil {
				return fmt.Errorf("read blob %s: %w", shortDigest(b.digest), err)
			}
			objID, err := provider.Upload(data)
			if err != nil {
				return fmt.Errorf("upload blob %s: %w", shortDigest(b.digest), err)
			}
			blobMap[b.digest] = manifest.BlobRef{Single: objID}
			fmt.Fprintf(os.Stderr, "✓ %s\n", shortObjID(objID))
		} else {
			// Chunked upload for large blobs exceeding provider limit
			numChunks := int((b.size + chunkSize - 1) / chunkSize)
			fmt.Fprintf(os.Stderr, "%s (%s, %d chunks) uploading...\n", label, formatBytes(b.size), numChunks)

			blobPath := store.BlobPath(b.digest)
			f, err := os.Open(blobPath)
			if err != nil {
				return fmt.Errorf("open blob %s: %w", shortDigest(b.digest), err)
			}

			var chunkIDs []string
			for chunkIdx := 0; chunkIdx < numChunks; chunkIdx++ {
				start := int64(chunkIdx) * chunkSize
				end := start + chunkSize
				if end > b.size {
					end = b.size
				}
				chunkLen := end - start

				buf := make([]byte, chunkLen)
				if _, err := f.ReadAt(buf, start); err != nil {
					f.Close()
					return fmt.Errorf("read chunk %d of %s: %w", chunkIdx, shortDigest(b.digest), err)
				}

				chunkLabel := fmt.Sprintf("  chunk %d/%d", chunkIdx+1, numChunks)
				fmt.Fprintf(os.Stderr, "%s (%s)... ", chunkLabel, formatBytes(chunkLen))
				startTime := time.Now()

				chunkID, err := provider.Upload(buf)
				if err != nil {
					f.Close()
					return fmt.Errorf("upload chunk %d of %s: %w", chunkIdx, shortDigest(b.digest), err)
				}
				chunkIDs = append(chunkIDs, chunkID)
				elapsed := time.Since(startTime).Round(time.Second)
				fmt.Fprintf(os.Stderr, "✓ %s (%s)\n", shortObjID(chunkID), elapsed)
			}
			f.Close()

			blobMap[b.digest] = manifest.BlobRef{Chunks: chunkIDs}
		}

		if verbose {
			ref := blobMap[b.digest]
			fmt.Printf("      digest: %s\n      ref: %v\n", b.digest, ref.IDs())
		}
	}

	// --- 5. Build Wolllama manifest ---
	wm := manifest.NewWithProvider(modelTag, entry.Raw, blobMap, provider.Name())
	manifestJSON, err := json.Marshal(wm)
	if err != nil {
		return fmt.Errorf("marshal wolllama manifest: %w", err)
	}

	// --- 6. Upload Wolllama manifest to the provider ---
	fmt.Fprintf(os.Stderr, "Uploading wolllama manifest... ")
	manifestObjID, err := provider.Upload(manifestJSON)
	if err != nil {
		return fmt.Errorf("upload manifest: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✓\n")

	// --- 7. Print success summary ---
	totalSz := totalSize(blobs)
	fmt.Println()
	fmt.Printf("  ✓ %s pushed to Walrus\n\n", modelTag)
	fmt.Printf("  Model:    %s\n", modelTag)
	fmt.Printf("  Blobs:    %d uploaded\n", len(blobs))
	fmt.Printf("  Size:     %s\n", formatBytes(totalSz))
	fmt.Printf("  Epochs:   %d\n", epochs)
	fmt.Printf("  Manifest: %s\n\n", manifestObjID)
	fmt.Printf("  Share: wolllama pull %s\n", manifestObjID)
	fmt.Printf("  List:  https://wolllama.dev/models  (submit to appear here)\n")

	return nil
}

type blobToUpload struct {
	digest    string
	mediaType string
	size      int64
}

func totalSize(blobs []blobToUpload) int64 {
	var s int64
	for _, b := range blobs {
		s += b.size
	}
	return s
}

func shortDigest(d string) string {
	const prefix = "sha256:"
	if strings.HasPrefix(d, prefix) {
		d = d[len(prefix):]
	}
	if len(d) > 12 {
		return d[:12]
	}
	return d
}

func shortObjID(id string) string {
	if len(id) > 16 {
		return id[:16] + "..."
	}
	return id
}

// splitURLs splits a comma or space-separated URL string into a slice.
func splitURLs(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' '
	})
	var urls []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			urls = append(urls, p)
		}
	}
	return urls
}

// persistManifest saves the wolllama manifest to the local push log.
func persistManifest(manifestObjID string, wm *manifest.WolllamaManifest) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := home + "/.wolllama/manifests"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(wm)
	if err != nil {
		return err
	}
	return os.WriteFile(dir+"/"+manifestObjID+".json", data, 0o644)
}
