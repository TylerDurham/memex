package cli

import (
	"fmt"
	"time"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/document/walker"
	"github.com/TylerDurham/memex/internal/embed"
	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/indexer"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type indexOptions struct {
	batchSize int
}

func newIndexCmd() *cobra.Command {
	var opts = indexOptions{}

	cmd := &cobra.Command{
		Use:   "index <name>",
		Args:  cobra.ExactArgs(1),
		Short: "Embed a repository's new and changed documents into its index.",
		Long: "Walk a repository, embed documents that are new or changed since the last run, " +
			"and remove documents that no longer exist. Unchanged documents are skipped.",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := config.LoadRepo(args[0])
			if err != nil {
				return err
			}
			logger.Debug("index", "name", repo.Name, "directory", repo.Directory,
				"app", repo.App, "ollamaURL", repo.OllamaURL, "model", repo.Model)

			w, err := walker.NewWalker(repo.App, document.IncludeProperties|document.IncludeChunks)
			if err != nil {
				return err
			}

			st, err := store.Init(repo.Name)
			if err != nil {
				return err
			}
			defer st.Close()

			docPrefix, _ := embed.Prefixes(repo.Model)
			ix := &indexer.Indexer{
				Walker:         w,
				Store:          st,
				Embedder:       embed.NewOllama(repo.OllamaURL, repo.Model),
				BatchSize:      opts.batchSize,
				DocumentPrefix: docPrefix,
			}

			start := time.Now()
			stats, err := ix.Index(cmd.Context(), repo.Directory)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(),
				"indexed %d, unchanged %d, removed %d (%d chunks embedded) in %s\n",
				stats.Indexed, stats.Unchanged, stats.Removed, stats.Chunks,
				time.Since(start).Round(time.Millisecond))
			return nil
		},
	}

	cmd.Flags().IntVar(&opts.batchSize, "batch-size", indexer.DefaultBatchSize, "Approximate chunks per embedding request.")

	return cmd
}
