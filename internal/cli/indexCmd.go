// Package cli
package cli

import (
	"fmt"
	"log/slog"

	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/indexers"
	"github.com/TylerDurham/memex/internal/repo"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index <repo>",
	Short: "Index a memex repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfgDir, err := configDir()
		if err != nil {
			return err
		}
		if verbose {
			logger.LogLevel.Set(slog.LevelDebug)
			logger.Debug("cmd: index", "config dir", cfgDir)
		}

		repo, err := repo.Load(cfgDir, name)
		if err != nil {
			return err
		}
		if verbose {
			logger.Debug("cmd: index", "repo dir", repo.Directory)
		}

		strat, err := indexers.GetStrategy(repo.Application)
		if err != nil {
			return err
		}
		err = indexers.Index(*repo, strat)
		if err != nil {
			return fmt.Errorf("could not index: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
