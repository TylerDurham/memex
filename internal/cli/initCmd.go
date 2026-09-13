// Package cli

package cli

import (
	"github.com/TylerDurham/memex/internal/globals"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type initOptions struct {
	local     bool
	directory string
	name      string
}

func newInitCmd(app globals.App) *cobra.Command {

	var opts = initOptions{}

	cmd := &cobra.Command{
		Use:     "init <name>",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Short:   "Initialize a new memex repo.",
		Long:    "Initialize a new memex repo.",
		RunE: func(cmd *cobra.Command, args []string) error {

			app.Logger.Debug("config: ", "ConfigDirectory", app.Config.ConfigDirectory())
			app.Logger.Debug("config: ", "LogDirectory", app.Config.LogDirectory())

			opts.name = args[0]

			app.Logger.Debug("flags: ", "directory", opts.directory)
			app.Logger.Debug("flags: ", "local", opts.local)
			app.Logger.Debug("flags: ", "name", opts.name)

			store, err := store.Init(app, opts.name)

			if err == nil {
				store.Close()
			}

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.directory, "directory", "d", app.Config.ConfigDirectory(), "The working directory.")
	cmd.Flags().BoolVarP(&opts.local, "local", "l", false, "Use a different working directory than MEMEX_CONFIG_DIR")

	return cmd

}
