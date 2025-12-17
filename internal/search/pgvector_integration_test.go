//go:build integration

package search

import (
	"context"
	"os"
	"testing"

	"cgap/internal/embedding"
	"cgap/internal/postgres"
	"cgap/internal/service"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// Integration test requires real Postgres with pgvector. Run with:
// DATABASE_URL=... go test -tags=integration ./internal/search -run TestPGVectorIntegration
func TestPGVectorIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL must be set for integration test")
	}

	ctx := context.Background()

	store, err := postgres.New(dbURL)
	if err != nil {
		t.Fatalf("failed to init postgres: %v", err)
	}
	defer store.Close()

	embedder := embedding.NewMockEmbedder(768)

	// Create a small fixture: project -> document -> chunk -> chunk_embedding
	projectID := uuid.New().String()
	slug := "itest-" + uuid.New().String()
	docID := uuid.New().String()
	chunkID := uuid.New().String()

	pool := store.Pool()

	_, err = pool.Exec(ctx, `
        INSERT INTO projects (id, name, slug)
        VALUES ($1, 'Integration Test', $2)
    `, projectID, slug)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, projectID)

	_, err = pool.Exec(ctx, `
        INSERT INTO documents (id, project_id, uri, title)
        VALUES ($1, $2, 'test://doc', 'Test Doc')
    `, docID, projectID)
	if err != nil {
		t.Fatalf("insert document: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM documents WHERE id = $1`, docID)

	chunkText := "hello world from pgvector"
	_, err = pool.Exec(ctx, `
        INSERT INTO chunks (id, document_id, ord, text)
        VALUES ($1, $2, 0, $3)
    `, chunkID, docID, chunkText)
	if err != nil {
		t.Fatalf("insert chunk: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM chunks WHERE id = $1`, chunkID)

	vec, err := embedder.Embed(ctx, chunkText)
	if err != nil {
		t.Fatalf("embed text: %v", err)
	}

	_, err = pool.Exec(ctx, `
        INSERT INTO chunk_embeddings (chunk_id, embedding)
        VALUES ($1, $2)
    `, chunkID, pgvector.NewVector(vec))
	if err != nil {
		t.Fatalf("insert chunk embedding: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM chunk_embeddings WHERE chunk_id = $1`, chunkID)

	pg := NewPGVector(store, embedder)
	svc := service.NewSearchService(store, pg)

	hits, err := svc.Search(ctx, slug, "hello world", 3, nil)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].ChunkID != chunkID {
		t.Fatalf("expected chunk %s, got %s", chunkID, hits[0].ChunkID)
	}
	if hits[0].DocumentID != docID {
		t.Fatalf("expected document %s, got %s", docID, hits[0].DocumentID)
	}
	if hits[0].Confidence <= 0 {
		t.Fatalf("expected positive confidence, got %f", hits[0].Confidence)
	}
}
