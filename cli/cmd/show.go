package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wolllama/pkg/manifest"
	wwalrus "github.com/wolllama/pkg/walrus"
)

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().Bool("local", false, "Show locally cached manifest instead of fetching from Walrus")
}

var showCmd = &cobra.Command{
	Use:   "show <manifest-obj-id>",
	Short: "Display model info from a Walrus manifest",
	Long: `Show fetches a Wolllama manifest from Walrus and displays
a human-friendly summary of the model: name, tag, blobs, and sizes.

Use --local to show a cached manifest instead of fetching from Walrus.

Examples:
  wolllama show O1ABCdef...xyz
  wolllama show O1ABCdef...xyz --local`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	manifestObjID := args[0]
	localOnly, _ := cmd.Flags().GetBool("local")

	var data []byte

	if localOnly {
		// Read from local cache
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		cachePath := filepath.Join(home, ".wolllama", "manifests", manifestObjID+".json")
		data, err = os.ReadFile(cachePath)
		if err != nil {
			return fmt.Errorf("manifest %s not found in local cache (use without --local to fetch from Walrus)", manifestObjID)
		}
	} else {
		// Fetch from Walrus
		aggregatorURL := viper.GetString("aggregator_url")
		walrusClient := wwalrus.NewClient(wwalrus.Config{
			AggregatorURLs: splitURLs(aggregatorURL),
		})

		var err error
		data, err = walrusClient.ReadBlob(manifestObjID)
		if err != nil {
			return fmt.Errorf("fetch manifest from Walrus: %w", err)
		}
	}

	var wm manifest.WolllamaManifest
	if err := json.Unmarshal(data, &wm); err != nil {
		return fmt.Errorf("parse wolllama manifest: %w", err)
	}

	summary, err := wm.Parse()
	if err != nil {
		return fmt.Errorf("parse manifest contents: %w", err)
	}

	// Print human-friendly summary
	fmt.Println()
	fmt.Printf("  Model:   %s\n", summary.Name)
	if summary.Tag != "" {
		fmt.Printf("  Tag:     %s\n", summary.Tag)
	}
	fmt.Printf("  Blobs:   %d\n", summary.BlobCount)
	fmt.Printf("  Size:    %s\n", formatBytes(summary.TotalSize))
	if !summary.CreatedAt.IsZero() {
		fmt.Printf("  Created: %s\n", summary.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()
	fmt.Println("  Blobs:")
	for _, b := range summary.Blobs {
		kind := "unknown"
		switch {
		case b.MediaType == "application/vnd.docker.container.image.v1+json":
			kind = "config"
		case b.MediaType == "application/vnd.ollama.image.model":
			kind = "model"
		case b.MediaType == "application/vnd.ollama.image.license":
			kind = "license"
		case b.MediaType == "application/vnd.ollama.image.params":
			kind = "params"
		}
		refStr := b.WalrusIDs[0]
		if len(b.WalrusIDs) > 1 {
			refStr = fmt.Sprintf("%d chunks", len(b.WalrusIDs))
		}
		fmt.Printf("    %s (%s)  %s  → %s\n",
			shortDigest(b.Digest), kind, formatBytes(b.Size), refStr)
	}

	fmt.Println()
	if !localOnly {
		fmt.Printf("  View on site: https://wolllama.dev/models\n")
	}

	// Cache locally for future --local use
	_ = cacheManifest(manifestObjID, data)

	return nil
}

// cacheManifest saves a raw manifest JSON to ~/.wolllama/manifests/<objID>.json
func cacheManifest(manifestObjID string, data []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".wolllama", "manifests")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, manifestObjID+".json"), data, 0o644)
}
