package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type searchOptions struct {
	json      bool
	config		string
	directory string
}

func newSearchCmd() *cobra.Command {
	opts := &searchOptions{}
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Searches a directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.json, "json", "j", false, "Return search results as JSON.")
	cmd.Flags().StringVarP(&opts.config, "config", "c", "", "The config file to use.")
	cmd.Flags().StringVarP(&opts.directory, "directory", "d", "", "The directory to search.")

	cmd.MarkFlagDirname("directory")
	cmd.MarkFlagFilename("config")
	cmd.MarkFlagsMutuallyExclusive("directory", "config")
	cmd.MarkFlagsOneRequired("directory", "config")

	return cmd
}

func runServer(opts *searchOptions) error {
	fmt.Printf("%v", opts)
	return nil
}
