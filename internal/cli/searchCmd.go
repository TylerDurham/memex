package cli

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/TylerDurham/memex/internal/embed"
	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	top       int
	minScore  float32
	allChunks bool
}

func newSearchCmd() *cobra.Command {
	var opts = searchOptions{}

	cmd := &cobra.Command{
		Use:     "search <name> <query...>",
		Args:    cobra.MinimumNArgs(2),
		Example: "  memex search tech-kasten go printf verbs",
		Short:   "Search a repository's index by meaning.",
		Long: "Search a repository's index by meaning. By default each file appears once, " +
			"represented by its best-matching section; --all-chunks lists every matching section.",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := config.LoadRepo(args[0])
			if err != nil {
				return err
			}

			st, err := store.Init(repo.Name)
			if err != nil {
				return err
			}
			defer st.Close()

			_, chunks, err := st.Count(cmd.Context())
			if err != nil {
				return err
			}
			if chunks == 0 {
				return errors.New("index is empty; run 'memex index' first")
			}

			// Queries must be embedded with the model the repo was indexed with.
			_, queryPrefix := embed.Prefixes(repo.Model)
			embedder := embed.NewOllama(repo.OllamaURL, repo.Model)
			vecs, err := embedder.Embed(cmd.Context(), []string{queryPrefix + strings.Join(args[1:], " ")})
			if err != nil {
				return err
			}

			search := st.SearchFiles
			if opts.allChunks {
				search = st.Search
			}
			results, err := search(cmd.Context(), vecs[0], opts.top, opts.minScore)
			if err != nil {
				return err
			}

			out := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, r := range results {
				fmt.Fprintf(out, "%.3f\t%s:%d\t%s\n", r.Score, r.FilePath, r.StartLine, r.HeadingPath)
			}
			return out.Flush()
		},
	}

	cmd.Flags().IntVarP(&opts.top, "top", "k", 5, "Maximum number of results.")
	cmd.Flags().Float32Var(&opts.minScore, "min-score", 0.5, "Minimum similarity score (0-1).")
	cmd.Flags().BoolVar(&opts.allChunks, "all-chunks", false, "List every matching section instead of one per file.")

	return cmd
}
