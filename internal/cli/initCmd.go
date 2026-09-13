// Package cli

package cli

import (
	"path/filepath"

	g "github.com/TylerDurham/memex/internal/globals"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type initOptions struct {
	local     bool
	directory string
	name      string
}

func newInitCmd(ctx g.App) *cobra.Command {

	var opts = initOptions{}

	cmd := &cobra.Command{
		Use:     "init <name>",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Short:   "Initialize a new memex repo.",
		Long:    "Initialize a new memex repo.",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx.Logger.Debug("config: ", "ConfigDirectory", ctx.Config.ConfigDirectory())
			ctx.Logger.Debug("config: ", "LogDirectory", ctx.Config.LogDirectory())

			opts.name = args[0]

			ctx.Logger.Debug("flags: ", "directory", opts.directory)
			ctx.Logger.Debug("flags: ", "local", opts.local)
			ctx.Logger.Debug("flags: ", "name", opts.name)

			storePath := filepath.Join(ctx.Config.ConfigDirectory(), opts.name, g.DBName)
			ctx.Logger.Debug("creating directory", "directory", storePath)

			g.EnsureDirectory(filepath.Dir(storePath))

			ctx.Logger.Debug("opening database", "database", storePath)
			store, err := store.Open(storePath)

			if err == nil {
				store.Close()
			}

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.directory, "directory", "d", ctx.Config.ConfigDirectory(), "The working directory.")
	cmd.Flags().BoolVarP(&opts.local, "local", "l", false, "Use a different working directory than MEMEX_CONFIG_DIR")

	return cmd

}
