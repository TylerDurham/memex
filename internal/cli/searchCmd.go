package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/embed"
	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	top       int
	minScore  float32
	allChunks bool
	json      bool
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

			if opts.json {
				return writeResultsJSON(cmd.OutOrStdout(), results)
			}
			writeResultsText(cmd.OutOrStdout(), results)
			return nil
		},
	}

	cmd.Flags().IntVarP(&opts.top, "top", "k", 5, "Maximum number of results.")
	cmd.Flags().Float32Var(&opts.minScore, "min-score", 0.5, "Minimum similarity score (0-1).")
	cmd.Flags().BoolVar(&opts.allChunks, "all-chunks", false, "List every matching section instead of one per file.")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print results as a JSON array, including each section's text.")

	return cmd
}

// displayTitle is the result's frontmatter title, or its file name without
// the extension when it has none, as Obsidian shows it.
func displayTitle(r store.Result) string {
	if r.Title != "" {
		return r.Title
	}
	base := filepath.Base(r.FilePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// writeResultsText prints two lines per result: score and title, then where
// it matched.
func writeResultsText(out io.Writer, results []store.Result) {
	for _, r := range results {
		fmt.Fprintf(out, "%.3f  %s\n", r.Score, displayTitle(r))
		fmt.Fprintf(out, "       %s:%d", r.FilePath, r.StartLine)
		if r.Heading != "" {
			fmt.Fprintf(out, " · %s", r.Heading)
		}
		fmt.Fprintln(out)
	}
}

// jsonResult is one search result in --json output. Field names are part of
// the CLI's interface; add fields rather than renaming them.
type jsonResult struct {
	Score       float32 `json:"score"`
	Title       string  `json:"title"` // frontmatter title, or file name if none
	Description string  `json:"description,omitempty"`
	Path        string  `json:"path"` // repo-relative
	Line        int     `json:"line"` // 1-based line where the section starts
	Heading     string  `json:"heading,omitempty"`
	HeadingPath string  `json:"heading_path,omitempty"` // "Top > Sub > Heading"
	Application string  `json:"application,omitempty"`
	URI         string  `json:"uri,omitempty"`
	Content     string  `json:"content"` // the matched section's text
}

// writeResultsJSON prints results as an indented JSON array; no results
// print [] rather than null.
func writeResultsJSON(out io.Writer, results []store.Result) error {
	items := make([]jsonResult, len(results))
	for i, r := range results {
		items[i] = jsonResult{
			Score:       r.Score,
			Title:       displayTitle(r),
			Description: r.Description,
			Path:        r.FilePath,
			Line:        r.StartLine,
			Heading:     r.Heading,
			HeadingPath: r.HeadingPath,
			Application: r.Application,
			URI:         r.URI,
			Content:     r.Content,
		}
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false) // keep "&", "<", ">" readable in note text
	return enc.Encode(items)
}
