package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// GoogleEmbedder calls Gemini embedding API (gemini-embedding-001).
// Auth uses API key; set GOOGLE_API_KEY or GEMINI_API_KEY (fallback EMBEDDING_API_KEY).
// Supports configurable output dimensions via EMBEDDING_DIMENSION (768, 1536, or 3072).
type GoogleEmbedder struct {
	apiKey    string
	model     string
	dimension int
	client    *http.Client
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
		model = "gemini-embedding-001"
	}

	dimension := 768
	if dimStr := os.Getenv("EMBEDDING_DIMENSION"); dimStr != "" {
		if d, err := strconv.Atoi(dimStr); err == nil {
			dimension = d
		}
	}

	return &GoogleEmbedder{
		apiKey:    apiKey,
		model:     model,
		dimension: dimension,
		client:    &http.Client{},
	}
}

type googleEmbedRequest struct {
	Content              googleContent `json:"content"`
	TaskType             string        `json:"taskType,omitempty"`
	OutputDimensionality int           `json:"outputDimensionality,omitempty"`
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

	// Strip "models/" prefix if present
	modelName := strings.TrimPrefix(g.model, "models/")
	url := fmt.Sprintf(googleEmbedURL, modelName)

	reqBody := googleEmbedRequest{
		Content:  googleContent{Parts: []googlePart{{Text: text}}},
		TaskType: "RETRIEVAL_DOCUMENT",
	}

	// Only set output dimensionality if not using default 3072
	if g.dimension != 3072 {
		reqBody.OutputDimensionality = g.dimension
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
		var apiErr map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("GoogleEmbedder: status %d: %v", resp.StatusCode, apiErr)
	}

	var out googleEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	if len(out.Embedding.Values) == 0 {
		return nil, fmt.Errorf("GoogleEmbedder: empty embedding")
	}

	// Normalize embeddings for dimensions < 3072 (per Gemini docs)
	embedding := out.Embedding.Values
	if g.dimension < 3072 {
		embedding = normalize(embedding)
	}

	return embedding, nil
}

// normalize returns L2-normalized embedding (required for gemini-embedding-001 with dimensions < 3072)
func normalize(vec []float32) []float32 {
	var sumSquares float64
	for _, v := range vec {
		sumSquares += float64(v) * float64(v)
	}
	norm := math.Sqrt(sumSquares)
	if norm == 0 {
		return vec
	}

	normalized := make([]float32, len(vec))
	for i, v := range vec {
		normalized[i] = float32(float64(v) / norm)
	}
	return normalized
}
