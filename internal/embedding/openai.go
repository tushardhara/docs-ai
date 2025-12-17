package embedding

import (
	"context"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIEmbedder calls OpenAI's embeddings API.
type OpenAIEmbedder struct {
	client *openai.Client
	model  string
}

func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		// Default to a lightweight model; adjust via EMBEDDING_MODEL if desired
		model = os.Getenv("EMBEDDING_MODEL")
		if model == "" {
			model = "text-embedding-3-small"
		}
	}
	return &OpenAIEmbedder{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(e.model),
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	// The SDK returns []float32 already
	return resp.Data[0].Embedding, nil
}
