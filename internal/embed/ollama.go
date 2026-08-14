// Package embed provides an embeddings backend. Currently targets Ollama's
// /api/embed endpoint, since that's the path of least friction if there's
// already an Ollama instance running on the local network.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// maxErrBody caps how much of an error response we read back. Ollama's error
// payloads are a single short JSON object; anything larger is a proxy's HTML
// error page, which is noise in a CLI error message.
const maxErrBody = 1 << 10

// Client talks to an Ollama instance for embedding generation.
type Client struct {
	BaseURL string // e.g. "http://erebor.snork.co:11434"
	Model   string // e.g. "nomic-embed-text"
	HTTP    *http.Client
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL: NormalizeBaseURL(baseURL),
		Model:   model,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

// NormalizeBaseURL makes a loosely-written base URL safe to concatenate an API
// path onto. Ollama's own OLLAMA_HOST convention omits the scheme (e.g.
// "127.0.0.1:11434"), and a URL pasted from a browser usually carries a
// trailing slash; either would produce a malformed request URL.
func NormalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL != "" && !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	// Truncate clips inputs that exceed the model's context instead of
	// failing the request. The chunker already sizes chunks to fit, so this
	// is a backstop: without it an oversized chunk takes down the whole
	// index pass with an HTTP 500.
	Truncate bool `json:"truncate"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed returns the embedding vector for a single text input.
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	vecs, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return vecs[0], nil
}

// EmbedBatch embeds multiple texts in a single request. Ollama's /api/embed
// accepts an array input and returns one vector per element, so a whole note's
// chunks cost one round trip rather than one per chunk.
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(embedRequest{Model: c.Model, Input: texts, Truncate: true})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling ollama at %s: %w (is it running and reachable?)", c.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Ollama puts the real diagnosis in the body (e.g. "the input length
		// exceeds the context length"); a bare status code sends the reader
		// looking in the wrong place.
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrBody))
		if len(msg) == 0 {
			return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, bytes.TrimSpace(msg))
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	// Callers index the result positionally against their input slice, so a
	// short response would silently mis-pair vectors with chunks.
	if len(out.Embeddings) != len(texts) {
		return nil, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(out.Embeddings), len(texts))
	}
	for i, v := range out.Embeddings {
		if len(v) == 0 {
			return nil, fmt.Errorf("ollama returned an empty embedding for input %d — check the model name %q is pulled", i, c.Model)
		}
	}

	return out.Embeddings, nil
}
