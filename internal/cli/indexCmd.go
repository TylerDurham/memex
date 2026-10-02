// Package cli
package cli

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/indexer"
	"github.com/TylerDurham/memex/internal/repo"
	"github.com/spf13/cobra"
)

func handleIdxEvent(e indexer.Event) {
	if e.Kind == indexer.EventDocIndexed {
		fmt.Printf(" - %s %s\n", e.Kind, e.Path)
		if verbose {
			json, err := e.Doc.ToJSONString()
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not convert doc to JSON: %v\n", err)
				return
			}
			fmt.Printf("%s\n", json)
		}
	}
}

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
			logger.ConsoleLevel.Set(slog.LevelDebug)
			logger.Debug("cmd: index", "config dir", cfgDir)
		}

		repo, err := repo.Load(cfgDir, name)
		if err != nil {
			return err
		}
		if verbose {
			logger.Debug("cmd: index", "repo dir", repo.Directory)
		}

		strat, err := indexer.For(repo.AppType)
		if err != nil {
			return err
		}

		req := indexer.DirIndexRequest{
			Repo:        *repo,
			IdxStrategy: strat,
			OnEvent: func(e indexer.Event) {
				handleIdxEvent(e)
			},
		}
		var res indexer.IndexResult
		res, err = indexer.IndexDir(req)

		if err != nil {
			return fmt.Errorf("could not index: %v", err)
		}

		log.Printf("%+v", res.Stats)

		if verbose {
			for _, doc := range res.Documents {
				log.Printf("%+v", doc)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
