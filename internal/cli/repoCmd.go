package cli

import "github.com/spf13/cobra"

// newRepoCmd groups the commands that manage repos: their settings and
// indexes, as opposed to what's in them.
func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage memex repos.",
		Long:  "Manage memex repos: create or update a repo's settings, list repos, and remove them.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoRemoveCmd())

	return cmd
}
