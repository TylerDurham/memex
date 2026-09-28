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
	require.NoError(t, st.ReplaceFile(ctx, "a.md", "h", 0, []Chunk{
		{Content: "a1", Embedding: []float32{1, 0}},
		{Content: "a2", Embedding: []float32{1, 0.1}},
		{Content: "a3", Embedding: []float32{1, 0.2}},
	}))
	require.NoError(t, st.ReplaceFile(ctx, "b.md", "h", 0, []Chunk{
		{Content: "b1", Embedding: []float32{1, 0.5}},
	}))
	require.NoError(t, st.ReplaceFile(ctx, "c.md", "h", 0, []Chunk{
		{Content: "c1", Embedding: []float32{0, 1}}, // orthogonal: score 0
	}))

	query := []float32{1, 0}

	chunks, err := st.Search(ctx, query, 2, 0.1)
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "a2"}, contents(chunks))

	files, err := st.SearchFiles(ctx, query, 2, 0.1)
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "b1"}, contents(files))

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
