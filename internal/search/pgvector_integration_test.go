//go:build integration

package search

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"cgap/internal/embedding"
	"cgap/internal/postgres"
)

// Integration test requires real Postgres with pgvector and a valid OPENAI_API_KEY.
// Run with: DATABASE_URL=... OPENAI_API_KEY=... go test -tags=integration ./internal/search -run TestPGVectorIntegration
func TestPGVectorIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	apiKey := os.Getenv("OPENAI_API_KEY")
	if dbURL == "" || apiKey == "" {
		t.Skip("DATABASE_URL and OPENAI_API_KEY must be set for integration test")
	}

	ctx := context.Background()

	store, err := postgres.New(dbURL)
	if err != nil {
		t.Fatalf("failed to init postgres: %v", err)
	}
	defer store.Close()

	embedder := embedding.NewOpenAIEmbedder(apiKey, os.Getenv("EMBEDDING_MODEL"))

	// Create a small fixture: project -> document -> chunk -> chunk_embedding
	projectID := uuid.New().String()
	docID := uuid.New().String()
	chunkID := uuid.New().String()

	pool := store.Pool()

	_, err = pool.Exec(ctx, `
        INSERT INTO projects (id, name, slug)
        VALUES ($1, 'testproj', 'testproj')
    `, projectID)
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
    `, chunkID, vec)
	if err != nil {
		t.Fatalf("insert chunk embedding: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM chunk_embeddings WHERE chunk_id = $1`, chunkID)

	pg := NewPGVector(store, embedder)
	results, err := pg.Search(ctx, "chunks", "hello world", 3, map[string]any{"project_id": projectID})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one result")
	}
	if results[0].ID != chunkID {
		t.Fatalf("expected chunk %s, got %s", chunkID, results[0].ID)
	}
}
