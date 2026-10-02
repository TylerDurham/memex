// Package cli
package cli

import (
	"encoding/json"
	"fmt"
	"io"
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
			// TODO: Implement file logging?
			logger.ConsoleLevel.Set(slog.LevelDebug)
			logger.Debug("cmd: repo list", "config dir", cfgDir)
		}

		names, err := repo.Names(cfgDir)
		if err != nil {
			return err
		}

		if len(names) == 0 && !formatJSON {
			logger.Warn("no repos found", "config dir", cfgDir)
			return nil
		}

		repos := []*repo.RepoInfo{}
		for _, name := range names {
			cfg, err := repo.Load(cfgDir, name)
			if err != nil {
				// Don't let one broken repo hide the rest.
				logger.Warn("could not load repo", "repo", name, "err", err)
				continue
			}
			repos = append(repos, cfg)
		}

		if formatJSON {
			return printRepoListAsJSON(cmd.OutOrStdout(), repos)
		}

		return printRepoListAsTable(cmd.OutOrStdout(), repos)

	},
}

func printRepoListAsJSON(w io.Writer, repos []*repo.RepoInfo) error {

	// Marshal the whole slice so the output is a single JSON array.
	data, err := json.MarshalIndent(repos, "", "\t")
	if err != nil {
		return fmt.Errorf("could not serialize to json: %w", err)
	}
	fmt.Fprintln(w, string(data))
	return nil
}

func printRepoListAsTable(w io.Writer, repos []*repo.RepoInfo) error {
	tabW := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tabW, "NAME\tAPPLICATION\tDIRECTORY\tCONFIG\tDB")
	for _, cfg := range repos {
		fmt.Fprintf(tabW, "%s\t%s\t%s\t%s\t%s\n", cfg.Name, cfg.AppType, cfg.Directory, cfg.ConfigFile, cfg.Database)
	}
	return tabW.Flush()
}

func init() {

	listCmd.Flags().BoolVarP(&formatJSON, "json", "j", false, "enable JSON output")
	rootCmd.AddCommand(listCmd)
}
