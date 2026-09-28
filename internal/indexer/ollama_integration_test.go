package indexer

import (
	"context"
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Index_Ollama indexes the test vault with a real Ollama server and runs
// a few searches. Skipped unless MEMEX_OLLAMA_URL is set, e.g.
//
//	MEMEX_OLLAMA_URL=http://localhost:11434 go test ./internal/indexer -run Ollama -v
func Test_Index_Ollama(t *testing.T) {
	baseURL := os.Getenv("MEMEX_OLLAMA_URL")
	if baseURL == "" {
		t.Skip("MEMEX_OLLAMA_URL not set")
	}
	model := os.Getenv("MEMEX_OLLAMA_MODEL")
	if model == "" {
		model = "nomic-embed-text"
	}

	ctx := context.Background()
	emb := embed.NewOllama(baseURL, model)
	ix := newIndexer(t, nil)
	ix.Embedder = emb
	ix.DocumentPrefix = "search_document: "

	stats, err := ix.Index(ctx, copyVault(t))
	require.NoError(t, err)
	t.Logf("%+v", stats)
	require.Positive(t, stats.Chunks)

	queries := map[string]string{
		"which dock has thunderbolt 5":      "dock",
		"lawn mower battery runtime":        "mower",
		"NAS with drive bays for backups":   "synology",
		"go printf verbs for struct fields": "Formatting Verbs",
	}
	for q, wantInPath := range queries {
		vecs, err := emb.Embed(ctx, []string{"search_query: " + q})
		require.NoError(t, err)

		results, err := ix.Store.Search(ctx, vecs[0], 3, 0)
		require.NoError(t, err)
		require.NotEmpty(t, results)

		t.Logf("%q", q)
		for _, r := range results {
			t.Logf("  %.3f  %s  [%s]", r.Score, r.FilePath, r.HeadingPath)
		}
		assert.Contains(t, results[0].FilePath, wantInPath, q)
	}
}
