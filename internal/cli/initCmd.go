// Package cli

package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command {
		Use: "init",
		Aliases: []string{ "i"},
		Short: "Initialize a new memex repo.",
		Long: "Initialize a new memex repo.",
		RunE: func(cmd *cobra.Command, args []string) error {

			slog.Info("directory: ", "directory", global.directory)
			slog.Info("local: ", "local", global.local)

			return nil
		},

	}

	return cmd

}
