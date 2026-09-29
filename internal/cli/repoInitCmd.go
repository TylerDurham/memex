package cli

import (
	"fmt"

	"github.com/TylerDurham/memex/internal/repo"
	"github.com/spf13/cobra"
)

// repoInitName is set by the optional -n/--name flag.
var repoInitName string

// repoInitApplication is set by the optional -a/--application flag.
var repoInitApplication string

// repoInitCmd represents the repo init command
var repoInitCmd = &cobra.Command{
	Use:   "init <directory>",
	Short: "Initialize a memex repository in a directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		directory := args[0]
		cfgDir, err := configDir()
		if err != nil {
			return err
		}
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "repo init called with directory %q, name %q and application %q\n", directory, repoInitName, repoInitApplication)
			fmt.Fprintf(cmd.OutOrStdout(), "config dir: %s\n", cfgDir)
		}
		path, err := repo.Init(cfgDir, repo.InitOptions{
			Directory:   directory,
			Name:        repoInitName,
			Application: repoInitApplication,
		})
		if err != nil {
			return err
		}
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
		}
		return nil
	},
}

func init() {
	repoInitCmd.Flags().StringVarP(&repoInitName, "name", "n", "", "name for the repo")
	repoInitCmd.Flags().StringVarP(&repoInitApplication, "application", "a", "obsidian", "application the repo belongs to")
	repoCmd.AddCommand(repoInitCmd)
}
