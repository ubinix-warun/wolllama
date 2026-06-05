package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wolllama/cli/internal/ollama"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List models available in Ollama",
	Long: `List displays all models currently available in your local Ollama store.
These are the models you can push to Walrus with 'wolllama push'.

The output mirrors 'ollama list' format: NAME, ID, SIZE, MODIFIED.

Examples:
  wolllama list`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	ollamaPath := viper.GetString("ollama_path")

	store, err := ollama.NewStore(ollamaPath)
	if err != nil {
		return fmt.Errorf("open ollama store: %w", err)
	}

	models, err := store.ListModels()
	if err != nil {
		return fmt.Errorf("list models: %w", err)
	}

	if len(models) == 0 {
		fmt.Println("No models found in Ollama store.")
		fmt.Println()
		fmt.Println("Pull a model with Ollama first:")
		fmt.Println("  ollama pull llama3.2")
		return nil
	}

	// Print table matching 'ollama list' format: NAME, ID, SIZE, MODIFIED
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tID\tSIZE\tMODIFIED")
	for _, m := range models {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			m.ModelTag,
			m.ID,
			formatBytes(m.Size),
			timeAgo(m.Modified),
		)
	}
	w.Flush()

	// Also show cached wolllama manifests if any
	cached, _ := listCachedManifests()
	if len(cached) > 0 {
		fmt.Println()
		fmt.Println("Cached wolllama manifests (pulled models):")
		w2 := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w2, "NAME\tSIZE\tMANIFEST ID")
		for _, c := range cached {
			shortID := c.objID
		if len(shortID) > 20 {
			shortID = shortID[:20] + "..."
		}
		fmt.Fprintf(w2, "%s\t%s\t%s\n", c.name, formatBytes(c.size), shortID)
		}
		w2.Flush()
	}

	return nil
}

type cachedModel struct {
	name  string
	size  int64
	objID string
}

// listCachedManifests reads ~/.wolllama/manifests/ for pulled/pushed models.
func listCachedManifests() ([]cachedModel, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cacheDir := home + "/.wolllama/manifests"

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, nil
	}

	var cached []cachedModel
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		objID := strings.TrimSuffix(entry.Name(), ".json")

		data, err := os.ReadFile(cacheDir + "/" + entry.Name())
		if err != nil {
			continue
		}

		// Parse the full wolllama manifest
		var wm struct {
			Name           string `json:"name"`
			OllamaManifest json.RawMessage `json:"ollamaManifest"`
			Blobs          map[string]struct {
				Single string   `json:"single"`
				Chunks []string `json:"chunks"`
			} `json:"blobs"`
		}
		if err := json.Unmarshal(data, &wm); err != nil || wm.Name == "" {
			continue
		}

		// Extract size from embedded Ollama manifest
		var om struct {
			Config *struct{ Size int64 } `json:"config"`
			Layers []struct{ Size int64 } `json:"layers"`
		}
		var totalSize int64
		if err := json.Unmarshal(wm.OllamaManifest, &om); err == nil {
			if om.Config != nil {
				totalSize += om.Config.Size
			}
			for _, l := range om.Layers {
				totalSize += l.Size
			}
		}

		cached = append(cached, cachedModel{
			name:  wm.Name,
			size:  totalSize,
			objID: objID,
		})
	}
	return cached, nil
}

// formatBytes formats a byte count using decimal units (matches Ollama).
// Ollama uses base-1000: 1.0 GB = 1,000,000,000 bytes.
func formatBytes(bytes int64) string {
	switch {
	case bytes >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB", float64(bytes)/1_000_000_000)
	case bytes >= 1_000_000:
		return fmt.Sprintf("%.0f MB", float64(bytes)/1_000_000)
	case bytes >= 1_000:
		return fmt.Sprintf("%.0f KB", float64(bytes)/1_000)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// timeAgo returns a human-friendly relative time like "2 hours ago".
func timeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "a moment ago"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
