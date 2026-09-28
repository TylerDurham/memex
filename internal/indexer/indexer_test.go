package indexer

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/document/walker"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeEmbedder returns a fixed-size vector per text and records call sizes.
type fakeEmbedder struct {
	model string
	calls []int
}

func (f *fakeEmbedder) Model() string { return f.model }

func (f *fakeEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	f.calls = append(f.calls, len(texts))
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = []float32{float32(len(t)), 1}
	}
	return out, nil
}

// copyVault copies the walker's obsidian testdata into a temp dir so the test
// can add and delete notes.
func copyVault(t *testing.T) string {
	src, err := filepath.Abs("../document/walker/testdata/obsidian")
	require.NoError(t, err)
	dst := t.TempDir()
	require.NoError(t, os.CopyFS(dst, os.DirFS(src)))
	return dst
}

func newIndexer(t *testing.T, emb *fakeEmbedder) *Indexer {
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })

	w, err := walker.NewWalker("obsidian", document.IncludeProperties|document.IncludeChunks)
	require.NoError(t, err)

	return &Indexer{Walker: w, Store: st, Embedder: emb, BatchSize: 16}
}

func Test_Index_Incremental(t *testing.T) {
	ctx := context.Background()
	vault := copyVault(t)
	emb := &fakeEmbedder{model: "m1"}
	ix := newIndexer(t, emb)

	// First run embeds everything.
	stats, err := ix.Index(ctx, vault)
	require.NoError(t, err)
	assert.Positive(t, stats.Indexed)
	assert.Positive(t, stats.Chunks)
	assert.Zero(t, stats.Unchanged)
	total := stats.Indexed

	chunks, files, err := ix.Store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, total, files)
	assert.Equal(t, stats.Chunks, chunks)

	// File records carry the provider's application and URI.
	results, err := ix.Store.SearchFiles(ctx, []float32{1, 1}, 0, -1)
	require.NoError(t, err)
	require.Len(t, results, total)
	for _, r := range results {
		assert.Equal(t, "obsidian", r.Application, r.FilePath)
		assert.Contains(t, r.URI, "obsidian://open?vault=", r.FilePath)
		// Every gear spec note has a frontmatter title and description.
		if strings.HasSuffix(r.FilePath, "-specs.md") {
			assert.NotEmpty(t, r.Title, r.FilePath)
			assert.NotEmpty(t, r.Description, r.FilePath)
		}
	}

	// Batches stay near BatchSize, except a lone document bigger than it.
	for _, n := range emb.calls[:len(emb.calls)-1] {
		assert.LessOrEqual(t, n, 16*2)
	}

	// Second run with no changes embeds nothing.
	emb.calls = nil
	stats, err = ix.Index(ctx, vault)
	require.NoError(t, err)
	assert.Equal(t, Stats{Unchanged: total}, stats)
	assert.Empty(t, emb.calls)

	// Edit one note, delete another.
	notes, _ := filepath.Glob(filepath.Join(vault, "*-specs.md"))
	require.GreaterOrEqual(t, len(notes), 2)
	f, err := os.OpenFile(notes[0], os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteString("\n## Added\n\nNew section.\n")
	require.NoError(t, err)
	f.Close()
	require.NoError(t, os.Remove(notes[1]))

	stats, err = ix.Index(ctx, vault)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Indexed)
	assert.Equal(t, 1, stats.Removed)
	assert.Equal(t, total-2, stats.Unchanged)
}

func Test_Index_FrontmatterEditUpdatesTitleWithoutReembedding(t *testing.T) {
	ctx := context.Background()
	vault := copyVault(t)
	emb := &fakeEmbedder{model: "m1"}
	ix := newIndexer(t, emb)

	_, err := ix.Index(ctx, vault)
	require.NoError(t, err)

	note := filepath.Join(vault, "apple-mac-studio-specs.md")
	data, err := os.ReadFile(note)
	require.NoError(t, err)
	edited := regexp.MustCompile(`(?m)^title: .*$`).ReplaceAll(data, []byte("title: Renamed Mac Studio"))
	require.NotEqual(t, data, edited, "fixture has no title line")
	require.NoError(t, os.WriteFile(note, edited, 0o644))

	emb.calls = nil
	stats, err := ix.Index(ctx, vault)
	require.NoError(t, err)
	assert.Zero(t, stats.Indexed)
	assert.Empty(t, emb.calls, "frontmatter-only edit must not re-embed")

	results, err := ix.Store.SearchFiles(ctx, []float32{1, 1}, 0, -1)
	require.NoError(t, err)
	for _, r := range results {
		if r.FilePath == "apple-mac-studio-specs.md" {
			assert.Equal(t, "Renamed Mac Studio", r.Title)
			return
		}
	}
	t.Fatal("apple-mac-studio-specs.md not in results")
}

func Test_Index_ModelChangeReembeds(t *testing.T) {
	ctx := context.Background()
	vault := copyVault(t)
	emb := &fakeEmbedder{model: "m1"}
	ix := newIndexer(t, emb)

	first, err := ix.Index(ctx, vault)
	require.NoError(t, err)

	emb.model = "m2"
	second, err := ix.Index(ctx, vault)
	require.NoError(t, err)
	assert.Equal(t, first.Indexed, second.Indexed)
	assert.Zero(t, second.Unchanged)
}
