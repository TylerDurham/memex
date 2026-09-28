/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/TylerDurham/memex/internal/repo"
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

		names, err := repo.Names(cfgDir)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "no repos found in %s\n", cfgDir)
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tAPPLICATION\tDIRECTORY")
		for _, name := range names {
			cfg, err := repo.Load(cfgDir, name)
			if err != nil {
				// Don't let one broken repo hide the rest.
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
				continue
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", cfg.Name, cfg.Application, cfg.Directory)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
