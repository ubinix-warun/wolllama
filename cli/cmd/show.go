package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show <manifest-obj-id>",
	Short: "Display model info from a Walrus manifest",
	Long: `Show fetches a Wolllama manifest from Walrus and displays
a human-friendly summary of the model: name, tag, blobs, and sizes.

Examples:
  wolllama show O1ABCdef...xyz`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	manifestObjID := args[0]

	// TODO: implement show
	//   1. Load config
	//   2. Fetch Wolllama manifest from Walrus aggregator
	//   3. Parse and display human-friendly summary

	fmt.Printf("show %s\n", manifestObjID)
	return nil
}
