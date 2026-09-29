package cli

import (
	"github.com/spf13/cobra"
)

// repoCmd represents the repo command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Work with a memex repository",
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
