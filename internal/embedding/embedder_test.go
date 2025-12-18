package embedding_test

import (
	"context"
	"testing"

	"cgap/internal/embedding"
)

func TestMockEmbedder_Embed(t *testing.T) {
	ctx := context.Background()
	embedder := embedding.NewMockEmbedder(768)

	testCases := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "simple text",
			text:    "Hello world",
			wantErr: false,
		},
		{
			name:    "long text",
			text:    "This is a longer text that should still be embedded",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			embedding, err := embedder.Embed(ctx, tc.text)
			if (err != nil) != tc.wantErr {
				t.Errorf("Expected error: %v, got: %v", tc.wantErr, err)
			}
			if !tc.wantErr && embedding != nil {
				if len(embedding) != 768 {
					t.Errorf("Expected 768-dim embedding, got %d", len(embedding))
				}
			}
		})
	}
}
