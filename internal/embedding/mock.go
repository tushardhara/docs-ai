package embedding

import (
	"context"
	"hash/fnv"
)

// MockEmbedder produces deterministic pseudo-embeddings for testing.
type MockEmbedder struct {
	Dim int
}

func NewMockEmbedder(dim int) *MockEmbedder {
	if dim <= 0 {
		dim = 1536
	}
	return &MockEmbedder{Dim: dim}
}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(text))
	seed := h.Sum64()
	// simple LCG
	var x uint64 = seed | 1
	out := make([]float32, m.Dim)
	for i := 0; i < m.Dim; i++ {
		x = (6364136223846793005*x + 1)
		// map to (0,1)
		out[i] = float32((x>>32)&0xffffffff) / float32(^uint32(0))
	}
	return out, nil
}
