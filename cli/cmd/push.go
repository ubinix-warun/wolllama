package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wolllama/cli/internal/ollama"
	"github.com/wolllama/pkg/manifest"
	wwalrus "github.com/wolllama/pkg/walrus"
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

	// --- 3. Initialize Walrus client ---
	publisherURL := viper.GetString("publisher_url")
	aggregatorURL := viper.GetString("aggregator_url")
	epochs := viper.GetInt("epochs")
	if flagEpochs, _ := cmd.Flags().GetInt("epochs"); flagEpochs > 0 {
		epochs = flagEpochs
	}

	walrusClient := wwalrus.NewClient(wwalrus.Config{
		PublisherURLs:  splitURLs(publisherURL),
		AggregatorURLs: splitURLs(aggregatorURL),
	})

	fmt.Fprintf(os.Stderr, "Pushing %s (%d blobs, %s) to Walrus...\n",
		modelTag, len(blobs), formatBytes(totalSize(blobs)))

	// --- 4. Upload blobs sequentially ---
	blobMap := make(map[string]string, len(blobs)) // digest → walrus object ID
	for i, b := range blobs {
		fmt.Printf("[%d/%d] %s (%s)... ", i+1, len(blobs), shortDigest(b.digest), formatBytes(b.size))

		label := fmt.Sprintf("[%d/%d] %s", i+1, len(blobs), shortDigest(b.digest))
		var objID string

		if b.size < 1_000_000 {
			// Small blobs (config, license, params): upload in-memory
			fmt.Fprintf(os.Stderr, "%s (%s)... ", label, formatBytes(b.size))
			data, err := store.ReadBlob(b.digest)
			if err != nil {
				return fmt.Errorf("read blob %s: %w", shortDigest(b.digest), err)
			}
			objID, err = walrusClient.StoreBlob(data, epochs)
			if err != nil {
				return fmt.Errorf("upload blob %s: %w", shortDigest(b.digest), err)
			}
			fmt.Fprintf(os.Stderr, "✓\n")
		} else {
			// Large blobs (model): upload with elapsed time
			fmt.Fprintf(os.Stderr, "%s (%s) uploading...", label, formatBytes(b.size))
			start := time.Now()

			blobPath := store.BlobPath(b.digest)
			f, err := os.Open(blobPath)
			if err != nil {
				return fmt.Errorf("open blob %s: %w", shortDigest(b.digest), err)
			}
			objID, err = walrusClient.StoreBlobFromReader(f, epochs)
			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n")
				return fmt.Errorf("upload blob %s: %w", shortDigest(b.digest), err)
			}
			elapsed := time.Since(start).Round(time.Second)
			fmt.Fprintf(os.Stderr, "\r%s ✓ %s (%s)\n", label, shortObjID(objID), elapsed)
		}

		blobMap[b.digest] = objID

		if verbose {
			fmt.Printf("      digest: %s\n      objID:  %s\n", b.digest, objID)
		}
	}

	// --- 5. Build Wolllama manifest ---
	wm := manifest.New(modelTag, entry.Raw, blobMap)
	manifestJSON, err := json.Marshal(wm)
	if err != nil {
		return fmt.Errorf("marshal wolllama manifest: %w", err)
	}

	// --- 6. Upload Wolllama manifest to Walrus ---
	fmt.Fprintf(os.Stderr, "Uploading wolllama manifest... ")
	manifestObjID, err := walrusClient.StoreBlob(manifestJSON, epochs)
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
