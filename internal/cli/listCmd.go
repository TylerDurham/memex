// Package cli
package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"text/tabwriter"

	"github.com/TylerDurham/memex/internal/globals/logger"
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
			logger.LogLevel.Set(slog.LevelDebug)
			logger.Debug("cmd: repo list", "config dir", cfgDir)
		}

		names, err := repo.Names(cfgDir)
		if err != nil {
			return err
		}
		if len(names) == 0 && !formatJSON {
			fmt.Fprintf(cmd.OutOrStdout(), "no repos found in %s\n", cfgDir)
			return nil
		}

		repos := []*repo.RepoInfo{}
		for _, name := range names {
			cfg, err := repo.Load(cfgDir, name)
			if err != nil {
				// Don't let one broken repo hide the rest.
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
				continue
			}
			repos = append(repos, cfg)
		}

		if formatJSON {
			// Marshal the whole slice so the output is a single JSON array.
			data, err := json.MarshalIndent(repos, "", "\t")
			if err != nil {
				return fmt.Errorf("could not serialize to json: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tAPPLICATION\tDIRECTORY\tCONFIG\tDB")
		for _, cfg := range repos {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", cfg.Name, cfg.Application, cfg.Directory, cfg.ConfigFile, cfg.Database)
		}
		return w.Flush()
	},
}

func init() {

	listCmd.Flags().BoolVarP(&formatJSON, "json", "j", false, "enable JSON output")
	rootCmd.AddCommand(listCmd)
}
