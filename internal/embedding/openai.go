package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// OpenAIEmbedder calls OpenAI's embeddings API via HTTP.
type OpenAIEmbedder struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = os.Getenv("EMBEDDING_MODEL")
		if model == "" {
			model = "text-embedding-3-small"
		}
	}
	return &OpenAIEmbedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type openaiEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openaiEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("openai embedder: missing OPENAI_API_KEY")
	}

	reqBody := openaiEmbedRequest{Model: e.model, Input: []string{text}}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai embedder: status %d", resp.StatusCode)
	}

	var out openaiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("openai embedder: empty data")
	}

	return out.Data[0].Embedding, nil
}
