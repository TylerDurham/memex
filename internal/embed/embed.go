// Package embed turns text into embedding vectors.
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

// Embedder turns texts into vectors. Implementations must return exactly one
// vector per input, in input order.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	// Model identifies the embedding model. Vectors from different models
	// aren't comparable, so the indexer re-embeds when this changes.
	Model() string
}

const DefaultOllamaURL = "http://localhost:11434"

var defaultHTTPClient = &http.Client{Timeout: 2 * time.Minute}

// Ollama calls Ollama's native /api/embed endpoint.
type Ollama struct {
	BaseURL string // defaults to DefaultOllamaURL
	Name    string // model, e.g. "nomic-embed-text"
	Client  *http.Client
}

func NewOllama(baseURL, model string) *Ollama {
	return &Ollama{BaseURL: baseURL, Name: model}
}

func (o *Ollama) Model() string { return o.Name }

func (o *Ollama) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	req := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{o.Name, texts}

	var resp struct {
		Embeddings [][]float32 `json:"embeddings"`
	}

	baseURL := o.BaseURL
	if baseURL == "" {
		baseURL = DefaultOllamaURL
	}
	if err := postJSON(ctx, o.Client, baseURL+"/api/embed", "", req, &resp); err != nil {
		return nil, err
	}
	return checkCount(resp.Embeddings, len(texts))
}

// OpenAICompatible calls a /v1/embeddings endpoint: OpenAI itself, or local
// servers that mimic it (Ollama, llama.cpp server, LM Studio, vLLM).
type OpenAICompatible struct {
	BaseURL string // e.g. "http://localhost:11434/v1" or "https://api.openai.com/v1"
	Name    string
	APIKey  string // optional; sent as a bearer token
	Client  *http.Client
}

func NewOpenAICompatible(baseURL, model, apiKey string) *OpenAICompatible {
	return &OpenAICompatible{BaseURL: baseURL, Name: model, APIKey: apiKey}
}

func (o *OpenAICompatible) Model() string { return o.Name }

func (o *OpenAICompatible) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	req := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{o.Name, texts}

	var resp struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := postJSON(ctx, o.Client, o.BaseURL+"/embeddings", o.APIKey, req, &resp); err != nil {
		return nil, err
	}

	// The spec doesn't guarantee order; place each vector by its index.
	out := make([][]float32, len(texts))
	for _, d := range resp.Data {
		if d.Index < 0 || d.Index >= len(out) {
			return nil, fmt.Errorf("embed: response index %d out of range", d.Index)
		}
		out[d.Index] = d.Embedding
	}
	return checkCount(out, len(texts))
}

func checkCount(vecs [][]float32, want int) ([][]float32, error) {
	if len(vecs) != want {
		return nil, fmt.Errorf("embed: got %d vectors for %d inputs", len(vecs), want)
	}
	for i, v := range vecs {
		if len(v) == 0 {
			return nil, fmt.Errorf("embed: empty vector for input %d", i)
		}
	}
	return vecs, nil
}

func postJSON(ctx context.Context, client *http.Client, url, apiKey string, body, out any) error {
	if client == nil {
		client = defaultHTTPClient
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("embed: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("embed: %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("embed: %s: %s: %s", url, resp.Status, bytes.TrimSpace(msg))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("embed: decode response: %w", err)
	}
	return nil
}

// Prefixes returns the task prefixes a model expects on documents and on
// search queries. Models not listed need none. Documents and queries must be
// embedded with matching prefixes or similarity scores degrade.
func Prefixes(model string) (document, query string) {
	name, _, _ := strings.Cut(model, ":") // drop an Ollama tag like ":latest"
	switch name {
	case "nomic-embed-text":
		return "search_document: ", "search_query: "
	case "mxbai-embed-large":
		return "", "Represent this sentence for searching relevant passages: "
	}
	return "", ""
}
