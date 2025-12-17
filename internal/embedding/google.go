package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// GoogleEmbedder calls Vertex AI (Gemini) embedding endpoint (e.g., text-embedding-004).
// Auth uses API key; set GOOGLE_API_KEY or GEMINI_API_KEY (fallback EMBEDDING_API_KEY).
type GoogleEmbedder struct {
	apiKey string
	model  string
	client *http.Client
}

const googleEmbedURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent"

func NewGoogleEmbedder(apiKey, model string) *GoogleEmbedder {
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("EMBEDDING_API_KEY")
	}

	if model == "" {
		model = os.Getenv("EMBEDDING_MODEL")
	}
	if model == "" {
		model = "text-embedding-004"
	}

	return &GoogleEmbedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type googleEmbedRequest struct {
	Model   string        `json:"model,omitempty"`
	Content googleContent `json:"content"`
}

type googleContent struct {
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text"`
}

type googleEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func (g *GoogleEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("GoogleEmbedder: missing API key (GOOGLE_API_KEY/GEMINI_API_KEY)")
	}

	modelName := g.model
	if !strings.HasPrefix(modelName, "models/") {
		modelName = "models/" + modelName
	}

	url := fmt.Sprintf(googleEmbedURL, modelName)

	reqBody := googleEmbedRequest{
		Model:   modelName,
		Content: googleContent{Parts: []googlePart{{Text: text}}},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"?key="+g.apiKey, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GoogleEmbedder: status %d", resp.StatusCode)
	}

	var out googleEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	if len(out.Embedding.Values) == 0 {
		return nil, fmt.Errorf("GoogleEmbedder: empty embedding")
	}

	return out.Embedding.Values, nil
}
