// Package index orchestrates the full indexing pipeline: walk the vault,
// skip unchanged files via content hash, chunk changed files, embed the
// chunks, and persist them.
package index

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tylersnork/memex/internal/chunker"
	"github.com/tylersnork/memex/internal/embed"
	"github.com/tylersnork/memex/internal/store"
	"github.com/tylersnork/memex/internal/vault"
)

type Options struct {
	VaultRoot string
	Force     bool // re-embed everything, ignoring content hashes
	DryRun    bool // report what would change without writing or embedding
	Chunker   chunker.Options
}

type Stats struct {
	FilesScanned  int
	FilesChanged  int
	FilesSkipped  int
	FilesDeleted  int
	ChunksWritten int
}

// Run executes one full (or incremental) index pass.
func Run(ctx context.Context, st *store.Store, emb *embed.Client, opts Options, log *slog.Logger) (Stats, error) {
	var stats Stats

	files, err := vault.Walk(opts.VaultRoot)
	if err != nil {
		return stats, fmt.Errorf("walk vault: %w", err)
	}
	stats.FilesScanned = len(files)

	seen := make(map[string]bool, len(files))

	for _, f := range files {
		seen[f.RelPath] = true

		raw, err := vault.ReadFile(f.AbsPath)
		if err != nil {
			log.Warn("skipping unreadable file", "path", f.RelPath, "err", err)
			continue
		}
		hash := vault.HashContent(raw)

		if !opts.Force {
			prevHash, err := st.FileHash(ctx, f.RelPath)
			if err != nil {
				return stats, fmt.Errorf("lookup prev hash for %s: %w", f.RelPath, err)
			}
			if prevHash == hash {
				stats.FilesSkipped++
				continue
			}
		}

		stats.FilesChanged++
		log.Info("indexing", "path", f.RelPath)

		if opts.DryRun {
			continue
		}

		chunks := chunker.Split(string(raw), opts.Chunker)
		if len(chunks) == 0 {
			// Empty or whitespace-only note — still record the file so we
			// don't re-scan it every pass, but with no chunks.
			if err := st.ReplaceFile(ctx, f.RelPath, hash, f.ModTime, nil); err != nil {
				return stats, fmt.Errorf("replace file %s: %w", f.RelPath, err)
			}
			continue
		}

		// Build the embedding inputs alongside the chunks they belong to,
		// dropping any chunk that is pure markup — embedding an empty
		// string wastes a slot and yields a meaningless vector.
		texts := make([]string, 0, len(chunks))
		kept := make([]chunker.Chunk, 0, len(chunks))
		for _, c := range chunks {
			body := chunker.StripForEmbedding(c.Content)
			if body == "" {
				continue
			}
			// Prepend the heading path so the embedding captures context
			// that the chunk body alone wouldn't have (e.g. a section
			// titled "Assumptions" is meaningless without knowing it's
			// under "Q3 Cash Plan > Runway").
			if c.HeadingPath != "" {
				texts = append(texts, c.HeadingPath+"\n\n"+body)
			} else {
				texts = append(texts, body)
			}
			kept = append(kept, c)
		}

		if len(kept) == 0 {
			// Every chunk was markup-only (an image-dump note). Record the
			// file so we don't re-scan it each pass.
			if err := st.ReplaceFile(ctx, f.RelPath, hash, f.ModTime, nil); err != nil {
				return stats, fmt.Errorf("replace file %s: %w", f.RelPath, err)
			}
			continue
		}

		vectors, err := emb.EmbedBatch(ctx, texts)
		if err != nil {
			return stats, fmt.Errorf("embed chunks for %s: %w", f.RelPath, err)
		}

		storeChunks := make([]store.Chunk, len(kept))
		for i, c := range kept {
			// Content stays as written so search snippets and jump-to-line
			// still mirror the file; only the vector saw the stripped text.
			storeChunks[i] = store.Chunk{
				FilePath:    f.RelPath,
				Heading:     c.Heading,
				HeadingPath: c.HeadingPath,
				Content:     c.Content,
				StartLine:   c.StartLine,
				Embedding:   vectors[i],
			}
		}

		if err := st.ReplaceFile(ctx, f.RelPath, hash, f.ModTime, storeChunks); err != nil {
			return stats, fmt.Errorf("replace file %s: %w", f.RelPath, err)
		}
		stats.ChunksWritten += len(storeChunks)
	}

	// Anything indexed previously but no longer on disk gets removed.
	if !opts.DryRun {
		known, err := st.KnownFiles(ctx)
		if err != nil {
			return stats, fmt.Errorf("list known files: %w", err)
		}
		for path := range known {
			if !seen[path] {
				log.Info("removing deleted file from index", "path", path)
				if err := st.RemoveFile(ctx, path); err != nil {
					return stats, fmt.Errorf("remove %s: %w", path, err)
				}
				stats.FilesDeleted++
			}
		}
	}

	return stats, nil
}
