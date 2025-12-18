// Package worker defines types and structures for background job processing.
package worker

import "time"

// IngestJob represents an ingestion job for processing documents.
type IngestJob struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	SourceID  string    `json:"source_id"`
	Type      string    `json:"type"` // crawl, github, openapi, slack, discord, upload
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RawDocument represents a document before processing.
type RawDocument struct {
	URI         string
	Title       string
	Language    string
	PublishedAt *time.Time
	Content     string
}

// Chunk represents a text chunk from a document.
type Chunk struct {
	Ord         int
	Text        string
	TokenCount  int
	SectionPath string
	ScoreRaw    float32
}

// EmbeddingTask represents a batch of chunks to be embedded.
type EmbeddingTask struct {
	ProjectID string
	Chunks    []Chunk
}

// MeiliRecord represents a document record in Meilisearch.
type MeiliRecord struct {
	ID          string            `json:"id"`
	ProjectID   string            `json:"project_id"`
	DocumentURI string            `json:"document_uri"`
	SourceType  string            `json:"source_type"`
	Title       string            `json:"title"`
	Text        string            `json:"text"`
	SectionPath string            `json:"section_path"`
	Ord         int               `json:"ord"`
	ScoreRaw    float32           `json:"score_raw"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// GapClusteringJob represents a job for clustering gap candidates.
type GapClusteringJob struct {
	ProjectID string
	Window    string // 7d, 30d, 90d
}
