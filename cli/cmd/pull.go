package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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

	// TODO: implement pull
	//   1. Load config
	//   2. Fetch Wolllama manifest from Walrus aggregator
	//   3. For each blob: download from Walrus, write to ~/.ollama/models/blobs/
	//   4. Write Ollama manifest entry
	//   5. Cache Wolllama manifest locally
	//   6. Print success + restart notice

	fmt.Printf("pull %s\n", manifestObjID)
	return nil
}
