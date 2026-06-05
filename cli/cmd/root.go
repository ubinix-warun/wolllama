package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wolllama",
	Short: "Decentralized model registry for Ollama",
	Long: `Wolllama pushes and pulls Ollama models to/from Walrus decentralized storage.

No central registry. No gatekeepers. Your models, your storage.

Commands:
  push    Upload a model from Ollama to Walrus
  pull    Download a model from Walrus into Ollama
  show    Display model info from a Walrus manifest
  list    List locally synced models
  config  Manage Wolllama configuration`,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}

// exitf prints an error and exits.
func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
