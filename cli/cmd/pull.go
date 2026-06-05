package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wolllama/cli/internal/ollama"
	"github.com/wolllama/pkg/manifest"
	wwalrus "github.com/wolllama/pkg/walrus"
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
	aggregatorURL := viper.GetString("aggregator_url")
	walrusClient := wwalrus.NewClient(wwalrus.Config{
		AggregatorURLs: splitURLs(aggregatorURL),
	})

	// --- 2. Fetch Wolllama manifest from Walrus ---
	fmt.Printf("Fetching wolllama manifest %s... ", shortObjID(manifestObjID))
	data, err := walrusClient.ReadBlob(manifestObjID)
	if err != nil {
		return fmt.Errorf("fetch manifest: %w", err)
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

	// --- 4. Download each blob and write to Ollama store ---
	blobCount := len(wm.Blobs)
	current := 0
	for digest, objID := range wm.Blobs {
		current++
		fmt.Printf("[%d/%d] %s (%s)... ", current, blobCount, shortDigest(digest), objIDStr(objID))

		blobData, err := walrusClient.ReadBlob(objID)
		if err != nil {
			return fmt.Errorf("download blob %s: %w", shortDigest(digest), err)
		}

		if err := store.WriteBlob(digest, blobData); err != nil {
			return fmt.Errorf("write blob %s: %w", shortDigest(digest), err)
		}

		fmt.Println("✓")
		if verbose {
			fmt.Printf("      digest: %s\n      objID:  %s\n      size:   %s\n", digest, objID, formatBytes(int64(len(blobData))))
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
