package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pushCmd)
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

	// TODO: implement push
	//   1. Load config
	//   2. Open Ollama store, read manifest
	//   3. Read blobs, upload to Walrus (sequential, with progress)
	//   4. Build Wolllama manifest
	//   5. Store on Walrus
	//   6. Print success summary

	fmt.Printf("push %s\n", modelTag)
	return nil
}
