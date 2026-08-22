package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

type indexOptions struct {
	config    string
	directory string
}

func newIndexCmd() *cobra.Command {
	opts := &indexOptions{}

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Indexes a directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIndex(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.config, "config", "c", "", "Path to a config file.")
	cmd.Flags().StringVarP(&opts.directory, "directory", "d", "", "The directory to index.")

	cmd.MarkFlagDirname("directory")
	cmd.MarkFlagFilename("config")
	cmd.MarkFlagsMutuallyExclusive("directory", "config")
	cmd.MarkFlagsOneRequired("directory", "config")

	return cmd
}

func runIndex(opts *indexOptions) error {
	fmt.Printf("%v", opts)

	return nil
}
