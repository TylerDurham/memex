package embed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Ollama_Embed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/embed", r.URL.Path)
		var req struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "test-model", req.Model)

		vecs := make([][]float32, len(req.Input))
		for i := range vecs {
			vecs[i] = []float32{float32(i), 1}
		}
		json.NewEncoder(w).Encode(map[string]any{"embeddings": vecs})
	}))
	defer srv.Close()

	vecs, err := NewOllama(srv.URL, "test-model").Embed(context.Background(), []string{"a", "b"})
	require.NoError(t, err)
	assert.Equal(t, [][]float32{{0, 1}, {1, 1}}, vecs)
}

func Test_OpenAICompatible_Embed_OrdersByIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/embeddings", r.URL.Path)
		assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
		w.Write([]byte(`{"data":[{"index":1,"embedding":[2]},{"index":0,"embedding":[1]}]}`))
	}))
	defer srv.Close()

	vecs, err := NewOpenAICompatible(srv.URL+"/v1", "m", "secret").Embed(context.Background(), []string{"a", "b"})
	require.NoError(t, err)
	assert.Equal(t, [][]float32{{1}, {2}}, vecs)
}

func Test_Embed_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `model "nope" not found`, http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := NewOllama(srv.URL, "nope").Embed(context.Background(), []string{"a"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
