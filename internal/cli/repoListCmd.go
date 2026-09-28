package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/spf13/cobra"
)

func newRepoListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Short:   "List repos.",
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := config.ListRepos()
			if err != nil {
				// Some configs couldn't be read; still list the rest.
				logger.Warn("could not load every repo", "err", err)
			}

			out := cmd.OutOrStdout()
			if len(repos) == 0 {
				fmt.Fprintln(out, "no repos; create one with 'memex repo init <name> --directory <path>'")
				return nil
			}

			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tAPP\tMODEL\tDIRECTORY")
			for _, r := range repos {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.Name, r.App, r.Model, r.Directory)
			}
			return tw.Flush()
		},
	}

	return cmd
}
