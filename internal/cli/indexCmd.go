package cli

import (
	"fmt"
	"time"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/document/walker"
	"github.com/TylerDurham/memex/internal/embed"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/indexer"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type indexOptions struct {
	repoOptions
	app       string
	batchSize int
}

func newIndexCmd() *cobra.Command {
	var opts = indexOptions{}

	cmd := &cobra.Command{
		Use:   "index",
		Args:  cobra.NoArgs,
		Short: "Embed a repository's new and changed documents into its index.",
		Long: "Walk a repository, embed documents that are new or changed since the last run, " +
			"and remove documents that no longer exist. Unchanged documents are skipped.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.resolve(); err != nil {
				return err
			}
			logger.Debug("index", "directory", opts.directory, "name", opts.name,
				"app", opts.app, "ollamaURL", opts.ollamaURL, "model", opts.model)

			w, err := walker.NewWalker(opts.app, document.IncludeProperties|document.IncludeChunks)
			if err != nil {
				return err
			}

			st, err := store.Init(opts.name)
			if err != nil {
				return err
			}
			defer st.Close()

			docPrefix, _ := embed.Prefixes(opts.model)
			ix := &indexer.Indexer{
				Walker:         w,
				Store:          st,
				Embedder:       opts.embedder(),
				BatchSize:      opts.batchSize,
				DocumentPrefix: docPrefix,
			}

			start := time.Now()
			stats, err := ix.Index(cmd.Context(), opts.directory)
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

	opts.addFlags(cmd)
	cmd.Flags().StringVarP(&opts.app, "app", "a", "obsidian", "The application the repository belongs to.")
	cmd.Flags().IntVar(&opts.batchSize, "batch-size", indexer.DefaultBatchSize, "Approximate chunks per embedding request.")

	return cmd
}
