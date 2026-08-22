package cli

import "github.com/spf13/cobra"

var indexCmd = &cobra.Command{
	Use: "index",
	Short: "indexes a directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)	
}
