package worker

import "time"

// Ingestion jobs and index payloads used by the worker.
type IngestJob struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	SourceID  string    `json:"source_id"`
	Type      string    `json:"type"` // crawl, github, openapi, slack, discord, upload
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RawDocument struct {
	URI         string
	Title       string
	Language    string
	PublishedAt *time.Time
	Content     string
}

type Chunk struct {
	Ord         int
	Text        string
	TokenCount  int
	SectionPath string
	ScoreRaw    float32
}

type EmbeddingTask struct {
	ProjectID string
	Chunks    []Chunk
}

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

type GapClusteringJob struct {
	ProjectID string
	Window    string // 7d, 30d, 90d
}
