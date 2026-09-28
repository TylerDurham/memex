/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List memex repositories",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, err := configDir()
		if err != nil {
			return err
		}
		if verbose {
			fmt.Fprintln(cmd.OutOrStdout(), "list called")
			fmt.Fprintf(cmd.OutOrStdout(), "config dir: %s\n", cfgDir)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
