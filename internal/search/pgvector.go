package search

import (
	"context"
	"fmt"
	"regexp"

	"cgap/internal/embedding"
	"cgap/internal/service"
	"cgap/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// PGVector implements service.Search using PostgreSQL (pgvector).
type PGVector struct {
	store    storage.Store
	embedder embedding.Embedder
}

func NewPGVector(store storage.Store, embedder embedding.Embedder) *PGVector {
	return &PGVector{store: store, embedder: embedder}
}

// Search performs semantic search using pgvector over precomputed chunk embeddings.
func (p *PGVector) Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]service.SearchResult, error) {
	if topK <= 0 {
		topK = 10
	}

	// Require project filter
	projectIDVal, ok := filters["project_id"]
	if !ok {
		return nil, fmt.Errorf("pgvector search requires project_id filter")
	}
	projectID, ok := projectIDVal.(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("invalid project_id filter")
	}

	if p.embedder == nil {
		return nil, fmt.Errorf("pgvector: no embedder configured")
	}

	// Create query embedding
	qvec, err := p.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query failed: %w", err)
	}

	// Convert to pgvector.Vector for proper SQL encoding
	vec := pgvector.NewVector(qvec)

	// Get underlying pgx pool from store (non-invasive via ad-hoc interface)
	var pool *pgxpool.Pool
	if ps, ok := p.store.(interface{ Pool() *pgxpool.Pool }); ok {
		pool = ps.Pool()
	} else {
		return nil, fmt.Errorf("store does not expose connection pool")
	}

	// Support passing project slug by resolving to UUID if needed
	pid := projectID
	if !looksLikeUUID(projectID) {
		if err := pool.QueryRow(ctx, `SELECT id FROM projects WHERE slug = $1`, projectID).Scan(&pid); err != nil {
			// Return a clean, user-friendly error without internal DB details
			return nil, fmt.Errorf("project '%s' not found", projectID)
		}
	}

	const sql = `
		SELECT
			c.id,
			c.text,
			d.id AS document_id,
			1.0 - (ce.embedding <=> $1) AS score
		FROM chunk_embeddings ce
		JOIN chunks c ON c.id = ce.chunk_id
		JOIN documents d ON d.id = c.document_id
		WHERE d.project_id = $2
		ORDER BY ce.embedding <=> $1
		LIMIT $3
	`

	rows, err := pool.Query(ctx, sql, vec, pid, topK)
	if err != nil {
		return nil, fmt.Errorf("pgvector query failed: %w", err)
	}
	defer rows.Close()

	var out []service.SearchResult
	for rows.Next() {
		var (
			chunkID    string
			text       string
			documentID string
			score      float32
		)
		if err := rows.Scan(&chunkID, &text, &documentID, &score); err != nil {
			return nil, fmt.Errorf("scan row failed: %w", err)
		}
		out = append(out, service.SearchResult{
			ID:   chunkID,
			Text: text,
			Metadata: map[string]any{
				"document_id": documentID,
			},
			Score: score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return out, nil
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func looksLikeUUID(s string) bool {
	return uuidRe.MatchString(s)
}
