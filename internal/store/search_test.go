package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SearchFiles_BestChunkPerFile(t *testing.T) {
	ctx := context.Background()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	defer st.Close()

	// a.md has three chunks that all beat b.md's only chunk.
	require.NoError(t, st.ReplaceFile(ctx, File{Path: "a.md", Application: "obsidian", URI: "obsidian://open?vault=v&file=a.md", ContentHash: "h"}, []Chunk{
		{Content: "a1", Embedding: []float32{1, 0}},
		{Content: "a2", Embedding: []float32{1, 0.1}},
		{Content: "a3", Embedding: []float32{1, 0.2}},
	}))
	require.NoError(t, st.ReplaceFile(ctx, File{Path: "b.md", ContentHash: "h"}, []Chunk{
		{Content: "b1", Embedding: []float32{1, 0.5}},
	}))
	require.NoError(t, st.ReplaceFile(ctx, File{Path: "c.md", ContentHash: "h"}, []Chunk{
		{Content: "c1", Embedding: []float32{0, 1}}, // orthogonal: score 0
	}))

	query := []float32{1, 0}

	chunks, err := st.Search(ctx, query, 2, 0.1)
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "a2"}, contents(chunks))

	files, err := st.SearchFiles(ctx, query, 2, 0.1)
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "b1"}, contents(files))

	// Results carry their file's application and URI.
	assert.Equal(t, "obsidian", files[0].Application)
	assert.Equal(t, "obsidian://open?vault=v&file=a.md", files[0].URI)
	assert.Empty(t, files[1].URI)

	// minScore still applies, and topK <= 0 means no limit.
	files, err = st.SearchFiles(ctx, query, 0, 0.1)
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "b1"}, contents(files))
}

func contents(rs []Result) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Content
	}
	return out
}

func Test_UpdateFileInfo(t *testing.T) {
	ctx := context.Background()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	defer st.Close()

	// A record written before application/uri were stored.
	require.NoError(t, st.ReplaceFile(ctx, File{Path: "a.md", ContentHash: "h"}, []Chunk{
		{Content: "a1", Embedding: []float32{1, 0}},
	}))

	require.NoError(t, st.UpdateFileInfo(ctx, File{Path: "a.md", Application: "obsidian", URI: "obsidian://a"}))
	require.NoError(t, st.UpdateFileInfo(ctx, File{Path: "missing.md", URI: "x"})) // no-op

	results, err := st.SearchFiles(ctx, []float32{1, 0}, 0, 0)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "obsidian", results[0].Application)
	assert.Equal(t, "obsidian://a", results[0].URI)
	assert.Equal(t, "a1", results[0].Content, "chunks untouched")

	hash, err := st.FileHash(ctx, "a.md")
	require.NoError(t, err)
	assert.Equal(t, "h", hash, "hash untouched")

	_, files, err := st.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, files, "unknown path not inserted")
}
