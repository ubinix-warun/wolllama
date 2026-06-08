package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ubinix-warun/wolllama/cli/internal/ollama"
	"github.com/ubinix-warun/wolllama/pkg/manifest"
	wwalrus "github.com/ubinix-warun/wolllama/pkg/walrus"
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull <manifest-obj-id>",
	Short: "Download a model from Walrus into Ollama",
	Long: `Pull downloads a model from Walrus and reconstructs it in your
local Ollama models directory.

Fetches the Wolllama manifest from Walrus by object ID, downloads
each blob, writes them to ~/.ollama/models/blobs/, and writes
the Ollama manifest so the model appears in Ollama after restart.

Examples:
  wolllama pull O1ABCdef...xyz
  wolllama pull O1ABCdef...xyz --ollama-path /custom/ollama`,
	Args: cobra.ExactArgs(1),
	RunE: runPull,
}

func runPull(cmd *cobra.Command, args []string) error {
	manifestObjID := args[0]

	// --- 1. Initialize Walrus client ---
	_, aggURL := walrusURLs()
	walrusClient := wwalrus.NewClient(wwalrus.Config{
		AggregatorURLs: splitURLs(aggURL),
	})

	// --- 2. Fetch Wolllama manifest from Walrus ---
	fmt.Printf("Fetching wolllama manifest %s... ", shortObjID(manifestObjID))
	data, err := walrusClient.ReadBlob(manifestObjID)
	if err != nil {
		// Quilt-patch IDs (Tatum) are rejected by the regular blob endpoint.
		// Try quilt-patch download as fallback.
		data, err = walrusClient.ReadBlobByQuiltPatchID(manifestObjID)
		if err != nil {
			return fmt.Errorf("fetch manifest: %w", err)
		}
	}
	fmt.Println("✓")

	var wm manifest.WolllamaManifest
	if err := json.Unmarshal(data, &wm); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}

	if err := wm.Validate(); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	summary, _ := wm.Parse()

	// --- 3. Open Ollama store ---
	ollamaPath := viper.GetString("ollama_path")
	store, err := ollama.NewStore(ollamaPath)
	if err != nil {
		return fmt.Errorf("open ollama store: %w", err)
	}

	verbose := viper.GetBool("verbose")

	// --- 4. Build digest → size map from the embedded Ollama manifest ---
	blobSizes := make(map[string]int64)
	if summary != nil {
		for _, b := range summary.Blobs {
			blobSizes[b.Digest] = b.Size
		}
	}

	// --- 5. Download each blob and write to Ollama store ---
	blobCount := len(wm.Blobs)
	current := 0
	for digest, ref := range wm.Blobs {
		current++

		var displayID string
		if ref.IsChunked() {
			displayID = fmt.Sprintf("%d chunks", len(ref.Chunks))
		} else {
			displayID = objIDStr(ref.Single)
		}
		fmt.Printf("[%d/%d] %s (%s)... ", current, blobCount, shortDigest(digest), displayID)

		var blobData []byte
		var err error

		if ref.IsChunked() {
			// Reassemble from chunks with per-chunk progress
			blobSize := blobSizes[digest]
			blobData = make([]byte, 0, blobSize)
			fmt.Fprintf(os.Stderr, "\n")
			for ci, chunkID := range ref.Chunks {
				fmt.Fprintf(os.Stderr, "    chunk %d/%d %s... ", ci+1, len(ref.Chunks), objIDStr(chunkID))
				chunk, err := walrusClient.ReadBlob(chunkID)
				if err != nil {
					// Quilt-patch fallback (Tatum)
					chunk, err = walrusClient.ReadBlobByQuiltPatchID(chunkID)
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n")
					return fmt.Errorf("download chunk %d of %s: %w", ci, shortDigest(digest), err)
				}
				blobData = append(blobData, chunk...)
				fmt.Fprintf(os.Stderr, "✓ (%s)\n", formatBytes(int64(len(chunk))))
			}
		} else {
			blobData, err = walrusClient.ReadBlob(ref.Single)
			if err != nil {
				// Quilt-patch ID fallback (Tatum provider)
				blobData, err = walrusClient.ReadBlobByQuiltPatchID(ref.Single)
			}
			if err != nil {
				return fmt.Errorf("download blob %s: %w", shortDigest(digest), err)
			}
		}

		// Verify sha256 matches the digest key in the manifest
		expectedHex := strings.TrimPrefix(digest, "sha256:")
		actualHash := sha256.Sum256(blobData)
		actualHex := hex.EncodeToString(actualHash[:])
		if actualHex != expectedHex {
			return fmt.Errorf("checksum mismatch for %s:\n  expected sha256:%s\n  got      sha256:%s",
				shortDigest(digest), expectedHex[:16]+"...", actualHex[:16]+"...")
		}

		if err := store.WriteBlob(digest, blobData); err != nil {
			return fmt.Errorf("write blob %s: %w", shortDigest(digest), err)
		}

		if ref.IsChunked() {
			fmt.Println("✓ checksum verified")
		} else {
			fmt.Println("✓")
		}
		if verbose {
			fmt.Printf("      digest: %s (verified)\n      size:   %s\n", digest, formatBytes(int64(len(blobData))))
		}
	}

	// --- 5. Write Ollama manifest so the model appears in ollama list ---
	var ollamaEntry ollama.ManifestEntry
	if err := json.Unmarshal(wm.OllamaManifest, &ollamaEntry); err != nil {
		return fmt.Errorf("decode ollama manifest: %w", err)
	}

	if err := store.WriteManifest(wm.Name, &ollamaEntry); err != nil {
		return fmt.Errorf("write ollama manifest: %w", err)
	}

	// --- 6. Cache wolllama manifest locally ---
	_ = cacheManifest(manifestObjID, data)

	// --- 7. Success ---
	fmt.Println()
	fmt.Printf("  ✓ %s pulled from Walrus\n\n", wm.Name)
	if summary != nil {
		fmt.Printf("  Model:  %s\n", summary.Name)
		fmt.Printf("  Blobs:  %d\n", summary.BlobCount)
		fmt.Printf("  Size:   %s\n", formatBytes(summary.TotalSize))
	}
	fmt.Println()
	fmt.Println("  Restart Ollama to use the model:")
	fmt.Println("    ollama serve")
	fmt.Printf("  Then run: ollama run %s\n", wm.Name)

	return nil
}

func objIDStr(id string) string {
	if len(id) > 16 {
		return id[:16] + "..."
	}
	return id
}
