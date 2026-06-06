package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
  list    List models available in Ollama
  config  Manage Wolllama configuration`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize config before any command runs
		if err := initViper(); err != nil {
			return fmt.Errorf("init config: %w", err)
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().String("ollama-path", "", "Path to Ollama home directory (default: ~/.ollama)")
	rootCmd.PersistentFlags().String("provider", "", "Storage provider: walrus, tatum, ipfs, s3")
	rootCmd.PersistentFlags().String("tatum-api-key", "", "Tatum API key (for --provider tatum)")

	// Bind persistent flags to viper
	viper.BindPFlag("ollama_path", rootCmd.PersistentFlags().Lookup("ollama-path"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))
	viper.BindPFlag("tatum_api_key", rootCmd.PersistentFlags().Lookup("tatum-api-key"))
}

// exitf prints an error and exits.
func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
