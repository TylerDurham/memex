// Package indexer walks a repository, embeds new or changed documents, and
// writes them to the store.
package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/document/walker"
	"github.com/TylerDurham/memex/internal/embed"
	"github.com/TylerDurham/memex/internal/store"
)

const DefaultBatchSize = 64

// Indexer brings the store in line with a repository on disk.
type Indexer struct {
	Walker   walker.Walker // must include document.IncludeChunks
	Store    *store.Store
	Embedder embed.Embedder

	// BatchSize is roughly how many chunks go to the embedder per call.
	// Chunks of one document are never split across calls.
	BatchSize int

	// DocumentPrefix is prepended to every embedded chunk. Some models need
	// a task prefix, e.g. "search_document: " for nomic-embed-text (queries
	// then use "search_query: ").
	DocumentPrefix string
}

// Stats summarizes an Index run.
type Stats struct {
	Indexed   int // documents embedded and written
	Unchanged int // documents skipped because their hash matched
	Removed   int // documents in the store no longer on disk
	Chunks    int // chunks embedded
}

// pending is a changed document waiting to be embedded.
type pending struct {
	doc  document.IndexDocument
	hash string
}

// Index walks root and updates the store: new or changed documents are
// re-embedded, unchanged ones skipped, and deleted ones removed.
func (ix *Indexer) Index(ctx context.Context, root string) (Stats, error) {
	var stats Stats

	docs, err := ix.Walker.Walk(root)
	if err != nil {
		return stats, fmt.Errorf("walk %q: %w", root, err)
	}

	known, err := ix.Store.KnownFiles(ctx)
	if err != nil {
		return stats, fmt.Errorf("load known files: %w", err)
	}

	batchSize := ix.BatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	var batch []pending
	batchChunks := 0
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		n, err := ix.embedAndStore(ctx, batch)
		if err != nil {
			return err
		}
		stats.Indexed += len(batch)
		stats.Chunks += n
		batch, batchChunks = batch[:0], 0
		return nil
	}

	for _, doc := range docs {
		delete(known, doc.RelPath)

		hash := ix.hash(doc)
		stored, err := ix.Store.FileHash(ctx, doc.RelPath)
		if err != nil {
			return stats, err
		}
		if stored == hash {
			// Content is unchanged, but the URI can still change (e.g. the
			// vault was renamed), and older indexes lack these fields.
			if err := ix.Store.UpdateFileInfo(ctx, fileRecord(doc, hash)); err != nil {
				return stats, err
			}
			stats.Unchanged++
			continue
		}

		if batchChunks > 0 && batchChunks+len(doc.Chunks) > batchSize {
			if err := flush(); err != nil {
				return stats, err
			}
		}
		batch = append(batch, pending{doc, hash})
		batchChunks += len(doc.Chunks)
	}
	if err := flush(); err != nil {
		return stats, err
	}

	// Anything left in known wasn't seen on disk.
	for relPath := range known {
		if err := ix.Store.RemoveFile(ctx, relPath); err != nil {
			return stats, fmt.Errorf("remove %q: %w", relPath, err)
		}
		stats.Removed++
	}

	return stats, nil
}

// embedAndStore embeds every chunk in batch in one call, then writes each
// document. Each ReplaceFile is its own transaction, so an interrupted run
// keeps the documents already written and resumes from there.
func (ix *Indexer) embedAndStore(ctx context.Context, batch []pending) (int, error) {
	var texts []string
	for _, p := range batch {
		for _, c := range p.doc.Chunks {
			texts = append(texts, ix.embedText(c))
		}
	}

	var vecs [][]float32
	if len(texts) > 0 {
		var err error
		if vecs, err = ix.Embedder.Embed(ctx, texts); err != nil {
			return 0, err
		}
	}

	i := 0
	for _, p := range batch {
		chunks := make([]store.Chunk, len(p.doc.Chunks))
		for j, c := range p.doc.Chunks {
			chunks[j] = store.Chunk{
				FilePath:    p.doc.RelPath,
				Heading:     lastOrEmpty(c.HeadingPath),
				HeadingPath: strings.Join(c.HeadingPath, " > "),
				Content:     c.Text,
				StartLine:   c.StartLine,
				Embedding:   vecs[i],
			}
			i++
		}
		if err := ix.Store.ReplaceFile(ctx, fileRecord(p.doc, p.hash), chunks); err != nil {
			return 0, fmt.Errorf("store %q: %w", p.doc.RelPath, err)
		}
	}
	return len(texts), nil
}

// embedText is what actually gets embedded for a chunk. The heading path
// gives short chunks the context of where they sit in the note.
func (ix *Indexer) embedText(c document.Chunk) string {
	if len(c.HeadingPath) == 0 {
		return ix.DocumentPrefix + c.Text
	}
	return ix.DocumentPrefix + strings.Join(c.HeadingPath, " > ") + "\n\n" + c.Text
}

// hash covers everything that affects the stored vectors: the model, the
// prefix, and each chunk's embedded text and location. A change to any of
// them re-embeds the document.
func (ix *Indexer) hash(doc document.IndexDocument) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00", ix.Embedder.Model(), ix.DocumentPrefix)
	for _, c := range doc.Chunks {
		fmt.Fprintf(h, "%d\x00%s\x00", c.StartLine, ix.embedText(c))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func fileRecord(doc document.IndexDocument, hash string) store.File {
	return store.File{
		Path:        doc.RelPath,
		Application: doc.Application,
		URI:         doc.URI,
		ContentHash: hash,
		ModTime:     doc.ModTime.Unix(),
	}
}

func lastOrEmpty(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}
