// Package cli

package cli

import (
	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type initOptions struct {
	local     bool
	directory string
	name      string
}

func newInitCmd() *cobra.Command {

	var opts = initOptions{}

	cmd := &cobra.Command{
		Use:     "init <name>",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Short:   "Initialize a new memex repo.",
		Long:    "Initialize a new memex repo.",
		RunE: func(cmd *cobra.Command, args []string) error {

			logger.Debug("config: ", "ConfigDirectory", config.ConfigDir())
			logger.Debug("config: ", "LogDirectory", config.ConfigDir())

			opts.name = args[0]

			logger.Debug("flags: ", "directory", opts.directory)
			logger.Debug("flags: ", "local", opts.local)
			logger.Debug("flags: ", "name", opts.name)

			store, err := store.Init(opts.name)

			if err == nil {
				store.Close()
			}

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.directory, "directory", "d", config.ConfigDir(), "The working directory.")
	cmd.Flags().BoolVarP(&opts.local, "local", "l", false, "Use a different working directory than MEMEX_CONFIG_DIR")

	return cmd

}
