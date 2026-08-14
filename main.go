// Command memex is a local-first semantic search CLI for Obsidian vaults (or
// any directory of markdown files). It embeds note chunks via Ollama, stores
// them in SQLite, and searches by cosine similarity — no cloud API, no
// vector-extension native dependency.
//
// The name is Vannevar Bush's: the memex of "As We May Think" (1945) was a
// personal store of documents traversed by associative trails rather than by
// index terms, which is what semantic search over your own notes amounts to.
//
// Usage — note that flags must precede the search query, since Go's flag
// package stops parsing at the first positional argument:
//
//	memex index  --vault ~/notes [--force] [--dry-run]
//	memex search --vault ~/notes [--limit 10] [--json] "runway pressure"
//	memex watch  --vault ~/notes
//	memex status --vault ~/notes
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tylersnork/memex/internal/chunker"
	"github.com/tylersnork/memex/internal/embed"
	"github.com/tylersnork/memex/internal/index"
	"github.com/tylersnork/memex/internal/store"
	"github.com/tylersnork/memex/internal/watch"
)

// defaultOllamaURL and defaultModel are overridable via flags or env vars so
// a single binary works whether Ollama lives on localhost or elsewhere on
// the Tailscale mesh (e.g. erebor).
const (
	defaultOllamaURL = "https://ollama.snork.co" //:11434"
	defaultModel     = "nomic-embed-text"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "index":
		err = runIndex(args)
	case "search":
		err = runSearch(args)
	case "watch":
		err = runWatch(args)
	case "status":
		err = runStatus(args)
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `memex — local-first semantic search for Obsidian vaults

Usage:
  memex index  --vault PATH [--force] [--dry-run]
  memex search --vault PATH [--limit N] [--min-score F] [--json] "query"
  memex watch  --vault PATH
  memex status --vault PATH

Flags common to all commands:
  --vault PATH       path to the vault root (required)
  --ollama-url URL   Ollama base URL (default: `+ollamaURLFromEnv()+`)
  --model NAME       embedding model name (default: `+envOr(defaultModel, "MEMEX_MODEL")+`)

Environment (flags win; first variable set wins):
  MEMEX_OLLAMA_URL, OLLAMA_URL, OLLAMA_HOST   Ollama base URL
  MEMEX_MODEL                                 embedding model name

Note: flags must precede the search query, e.g.
  memex search --vault PATH "runway pressure"
`)
}

// commonFlags registers the flags shared by every subcommand and returns
// the resolved db path and an embed.Client. Keeping this in one place is
// what makes each subcommand function short and consistent.
func commonFlags(fs *flag.FlagSet) (vaultRoot *string, ollamaURL *string, model *string) {
	vaultRoot = fs.String("vault", "", "path to the vault root (required)")
	ollamaURL = fs.String("ollama-url", ollamaURLFromEnv(), "Ollama base URL")
	model = fs.String("model", envOr(defaultModel, "MEMEX_MODEL"), "embedding model name")
	return
}

// ollamaURLEnvVars are the environment variables consulted for the Ollama base
// URL, in precedence order. MEMEX_OLLAMA_URL comes first so this tool can
// be aimed at a different instance without disturbing a machine-wide setting;
// OLLAMA_HOST is Ollama's own convention and is honoured last.
var ollamaURLEnvVars = []string{"MEMEX_OLLAMA_URL", "OLLAMA_URL", "OLLAMA_HOST"}

func ollamaURLFromEnv() string {
	return envOr(defaultOllamaURL, ollamaURLEnvVars...)
}

// envOr returns the first non-empty environment variable among keys, or
// fallback if none are set.
func envOr(fallback string, keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return fallback
}

// dbPath places the index inside the vault's .obsidian directory when it
// exists (so it travels with the vault but stays out of Obsidian's own
// file list), falling back to a sibling file for non-Obsidian markdown dirs.
func dbPath(vaultRoot string) string {
	obsidianDir := filepath.Join(vaultRoot, ".obsidian")
	if info, err := os.Stat(obsidianDir); err == nil && info.IsDir() {
		return filepath.Join(obsidianDir, "memex.db")
	}
	return filepath.Join(vaultRoot, ".memex.db")
}

func requireVault(vaultRoot string) error {
	if vaultRoot == "" {
		return fmt.Errorf("--vault is required")
	}
	info, err := os.Stat(vaultRoot)
	if err != nil {
		return fmt.Errorf("vault path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("--vault must be a directory")
	}
	return nil
}

func logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// --- index ---

func runIndex(args []string) error {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	vaultRoot, ollamaURL, model := commonFlags(fs)
	force := fs.Bool("force", false, "re-embed all chunks, ignoring content hashes")
	dryRun := fs.Bool("dry-run", false, "report what would change without writing or embedding")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireVault(*vaultRoot); err != nil {
		return err
	}

	st, err := store.Open(dbPath(*vaultRoot))
	if err != nil {
		return err
	}
	defer st.Close()

	emb := embed.NewClient(*ollamaURL, *model)
	log := logger()

	stats, err := index.Run(context.Background(), st, emb, index.Options{
		VaultRoot: *vaultRoot,
		Force:     *force,
		DryRun:    *dryRun,
		Chunker:   chunker.DefaultOptions(),
	}, log)
	if err != nil {
		return err
	}

	fmt.Printf("scanned %d files: %d changed, %d skipped (unchanged), %d deleted, %d chunks written\n",
		stats.FilesScanned, stats.FilesChanged, stats.FilesSkipped, stats.FilesDeleted, stats.ChunksWritten)
	if *dryRun {
		fmt.Println("(dry run — nothing was written)")
	}
	return nil
}

// --- search ---

type jsonResult struct {
	File        string  `json:"file"`
	Heading     string  `json:"heading,omitempty"`
	HeadingPath string  `json:"heading_path,omitempty"`
	Score       float32 `json:"score"`
	StartLine   int     `json:"start_line"`
	Snippet     string  `json:"snippet"`
}

func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	vaultRoot, ollamaURL, model := commonFlags(fs)
	limit := fs.Int("limit", 10, "max results to return")
	minScore := fs.Float64("min-score", 0.5, "minimum cosine similarity to include")
	asJSON := fs.Bool("json", false, "output results as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf(`search requires a query, e.g. memex search --vault PATH "runway pressure"`)
	}
	query := fs.Arg(0)

	if err := requireVault(*vaultRoot); err != nil {
		return err
	}

	st, err := store.Open(dbPath(*vaultRoot))
	if err != nil {
		return err
	}
	defer st.Close()

	emb := embed.NewClient(*ollamaURL, *model)
	ctx := context.Background()

	qvec, err := emb.Embed(ctx, query)
	if err != nil {
		return fmt.Errorf("embed query: %w", err)
	}

	results, err := st.Search(ctx, qvec, *limit, float32(*minScore))
	if err != nil {
		return err
	}

	if *asJSON {
		out := make([]jsonResult, len(results))
		for i, r := range results {
			out[i] = jsonResult{
				File:        r.FilePath,
				Heading:     r.Heading,
				HeadingPath: r.HeadingPath,
				Score:       r.Score,
				StartLine:   r.StartLine,
				Snippet:     snippet(r.Content, 200),
			}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	if len(results) == 0 {
		fmt.Println("no results at or above the min-score threshold")
		return nil
	}
	for _, r := range results {
		fmt.Printf("%.3f  %s", r.Score, r.FilePath)
		if r.HeadingPath != "" {
			fmt.Printf("  [%s]", r.HeadingPath)
		}
		fmt.Printf("\n      %s\n\n", snippet(r.Content, 160))
	}
	return nil
}

func snippet(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + "…"
}

// --- watch ---

func runWatch(args []string) error {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	vaultRoot, ollamaURL, model := commonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireVault(*vaultRoot); err != nil {
		return err
	}

	st, err := store.Open(dbPath(*vaultRoot))
	if err != nil {
		return err
	}
	defer st.Close()

	emb := embed.NewClient(*ollamaURL, *model)
	log := logger()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run one full pass up front so watch mode is useful immediately,
	// rather than waiting for the first file change to build any index.
	log.Info("running initial index pass before watching")
	if _, err := index.Run(ctx, st, emb, index.Options{
		VaultRoot: *vaultRoot,
		Chunker:   chunker.DefaultOptions(),
	}, log); err != nil {
		return fmt.Errorf("initial index pass: %w", err)
	}

	return watch.Run(ctx, st, emb, index.Options{
		VaultRoot: *vaultRoot,
		Chunker:   chunker.DefaultOptions(),
	}, log)
}

// --- status ---

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	vaultRoot, _, _ := commonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireVault(*vaultRoot); err != nil {
		return err
	}

	path := dbPath(*vaultRoot)
	if _, err := os.Stat(path); err != nil {
		fmt.Println("no index found — run `memex index --vault", *vaultRoot, "` first")
		return nil
	}

	st, err := store.Open(path)
	if err != nil {
		return err
	}
	defer st.Close()

	chunks, files, err := st.Count(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("index: %s\nfiles indexed: %d\nchunks indexed: %d\n", path, files, chunks)
	return nil
}
