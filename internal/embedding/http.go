package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// HTTPEmbedder calls a generic HTTP endpoint to get embeddings.
// Expected response formats supported:
// - {"embedding": [..]}
// - {"data": [{"embedding": [..]}]}
type HTTPEmbedder struct {
	url        string
	model      string
	apiKey     string
	authHeader string
	client     *http.Client
}

func NewHTTPEmbedder(url, model, apiKey, authHeader string) *HTTPEmbedder {
	if url == "" {
		url = os.Getenv("EMBEDDING_ENDPOINT")
	}
	if model == "" {
		model = os.Getenv("EMBEDDING_MODEL")
	}
	if apiKey == "" {
		apiKey = os.Getenv("EMBEDDING_API_KEY")
	}
	if authHeader == "" {
		if v := os.Getenv("EMBEDDING_AUTH_HEADER"); v != "" {
			authHeader = v
		} else {
			authHeader = "Authorization"
		}
	}
	return &HTTPEmbedder{
		url:        url,
		model:      model,
		apiKey:     apiKey,
		authHeader: authHeader,
		client:     &http.Client{},
	}
}

type httpEmbedRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"`
}

type httpEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
	Data      []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (h *HTTPEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if h.url == "" {
		return nil, fmt.Errorf("HTTPEmbedder: EMBEDDING_ENDPOINT not set")
	}
	payload := httpEmbedRequest{Input: text, Model: h.model}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.apiKey != "" {
		req.Header.Set(h.authHeader, "Bearer "+h.apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTPEmbedder: status %d", resp.StatusCode)
	}

	var out httpEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embedding) > 0 {
		return out.Embedding, nil
	}
	if len(out.Data) > 0 {
		return out.Data[0].Embedding, nil
	}
	return nil, fmt.Errorf("HTTPEmbedder: no embedding in response")
}
