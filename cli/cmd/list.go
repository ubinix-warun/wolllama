package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List locally synced models",
	Long: `List displays models that have been pulled via wolllama and cached
in ~/.wolllama/manifests/.

Examples:
  wolllama list`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// TODO: implement list
	//   1. Read ~/.wolllama/manifests/ directory
	//   2. Parse each cached Wolllama manifest
	//   3. Display table: NAME, TAG, SIZE, MANIFEST ID

	fmt.Println("list")
	return nil
}
