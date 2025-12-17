package search

import (
	"context"
	"fmt"

	"cgap/internal/service"
	"cgap/internal/storage"
)

// PGVector implements service.Search using PostgreSQL (pgvector).
type PGVector struct {
	store storage.Store
}

func NewPGVector(store storage.Store) *PGVector {
	return &PGVector{store: store}
}

// Search performs semantic search using pgvector (TODO).
// Currently returns a not implemented error so Hybrid provider can fallback.
func (p *PGVector) Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]service.SearchResult, error) {
	// TODO: Implement semantic search using embeddings once ingestion pipeline writes vectors
	// Suggested SQL sketch (when embeddings exist):
	// SELECT chunk_id, document_id, text, 1.0 - (embedding <=> $1) AS score
	// FROM chunks WHERE project_id = $2 ORDER BY embedding <=> $1 ASC LIMIT $3
	return nil, fmt.Errorf("pgvector provider not yet implemented")
}
