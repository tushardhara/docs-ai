package embedding

import "context"

// Embedder generates an embedding vector for input text.
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}
