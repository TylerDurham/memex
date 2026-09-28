// Package cli builds the memex command tree: the whole user-facing surface of
// the binary lives here, and cmd/main.go does nothing but call Execute.
//
// The tree is rooted at "memex", with one subcommand per verb:
//
//	memex repo init <name> --directory ~/notes  save a repo's settings, create its index
//	memex repo list                             list repos
//	memex repo rm   <name>                      delete a repo's settings and index
//	memex index  <name>                         embed a repo's new and changed notes
//	memex search <name> <query...>              query that index by similarity
//
// Conventions worth keeping as commands are added:
//
//   - A command owns its flags and binds them to its own options struct;
//     rootCmd is the only shared package state.
//   - Prefer a newXxxCmd() constructor over a package-level var, so the
//     options struct stays scoped to the command that reads it.
//   - Use RunE rather than Run. Returning the error leaves reporting to
//     Execute, which prints to stderr and sets the exit status, so no command
//     calls os.Exit on its own.
//   - Keep commands thin. Parsing, validation and output formatting belong
//     here; chunking, embedding, storage and watching live under internal/
//     and should be callable without going through cobra.
package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/TylerDurham/memex/internal/globals"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	local     bool
	directory string
	verbose   bool
}

var global = globalOptions{}

var rootCmd = &cobra.Command{
	Use:   globals.App().Name(),
	Short: "Simple semantic search.",
	Long:  "Simple semantic search.",
	// Execute reports errors; don't let cobra also print them with usage.
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if global.verbose {
			logger.LogLevel.Set(slog.LevelDebug)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&global.verbose, "verbose", "v", false, "Show debug logging.")

	rootCmd.AddCommand(newRepoCmd())
	rootCmd.AddCommand(newIndexCmd())
	rootCmd.AddCommand(newSearchCmd())
}
