// Package cli builds the memex command tree: the whole user-facing surface of
// the binary lives here, and cmd/main.go does nothing but call Execute.
//
// The tree is rooted at "memex", with one subcommand per verb:
//
//	memex index  --directory ~/notes    embed a vault's notes into the index
//	memex search --directory ~/notes …  query that index by similarity
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
	"os"

	g "github.com/TylerDurham/memex/internal/globals"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	local     bool
	directory string
}

var global = globalOptions{}

var rootCmd = &cobra.Command{
	Use:   "memex",
	Short: "Simple semantic search.",
	Long:  "Simple semantic search.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return g.EnsureMemexConfigDir()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&global.directory, "directory", "d", g.GetConfigDir(), "The working directory.")
	rootCmd.PersistentFlags().BoolVarP(&global.local, "local", "l", false, "Use a different working directory than MEMEX_CONFIG_DIR")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newSearchCmd())
	rootCmd.AddCommand(newIndexCmd())
}

