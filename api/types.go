package api

import (
	"time"

	"cgap/internal/model"
)

// Core domain models aliased to shared internal/model definitions to avoid drift.
type (
	Project           = model.Project
	User              = model.User
	ProjectMember     = model.ProjectMember
	APIKey            = model.APIKey
	Source            = model.Source
	Document          = model.Document
	Chunk             = model.Chunk
	Thread            = model.Thread
	Message           = model.Message
	Answer            = model.Answer
	Citation          = model.Citation
	Feedback          = model.Feedback
	DeflectEvent      = model.DeflectEvent
	AnalyticsEvent    = model.AnalyticsEvent
	GapCandidate      = model.GapCandidate
	GapCluster        = model.GapCluster
	GapClusterExample = model.GapClusterExample
)

// Interfaces keep transport decoupled from data stores.
type ChatService interface {
	Chat(req ChatRequest) (ChatResponse, error)
	ChatStream(req ChatRequest) (<-chan StreamFrame, error)
}

type SearchService interface {
	Search(projectID, query string, topK int, filters map[string]any) ([]SearchHit, error)
}

type DeflectService interface {
	Suggest(projectID, subject, body string, topK int) (string, []DeflectSuggestion, error)
	TrackEvent(projectID, suggestionID, action, threadID string, metadata map[string]any) error
}

type AnalyticsService interface {
	Summary(projectID string, from, to *time.Time, integration string) (AnalyticsSummary, error)
}

type GapsService interface {
	Run(projectID, window string) (string, error)
	List(projectID string) ([]GapCluster, error)
	Get(projectID, clusterID string) (GapCluster, []GapClusterExample, error)
}

// Request/response DTOs align with OpenAPI.
type ChatRequest struct {
	ProjectID      string         `json:"project_id"`
	Query          string         `json:"query"`
	Mode           string         `json:"mode,omitempty"`
	ContextFilters map[string]any `json:"context_filters,omitempty"`
	TopK           int            `json:"top_k,omitempty"`
	ThreadID       string         `json:"thread_id,omitempty"`
}

type ChatResponse struct {
	ThreadID    string     `json:"thread_id"`
	Answer      string     `json:"answer"`
	IsUncertain bool       `json:"is_uncertain"`
	Citations   []Citation `json:"citations"`
}

type StreamFrame struct {
	Delta       string     `json:"delta"`
	Citations   []Citation `json:"citations,omitempty"`
	IsUncertain bool       `json:"is_uncertain,omitempty"`
}

type SearchHit struct {
	ChunkID     string  `json:"chunk_id"`
	Text        string  `json:"text"`
	DocumentURI string  `json:"document_uri"`
	SourceType  string  `json:"source_type"`
	Score       float32 `json:"score"`
}

type DeflectSuggestion struct {
	Answer    string     `json:"answer"`
	Citations []Citation `json:"citations"`
	Score     float32    `json:"score"`
}

type AnalyticsSummary struct {
	TotalQuestions int     `json:"total_questions"`
	Uncertain      int     `json:"uncertain"`
	UniqueUsers    int     `json:"unique_users"`
	DeflectionRate float32 `json:"deflection_rate"`
}

type SearchRequest struct {
	ProjectID string         `json:"project_id"`
	Query     string         `json:"query"`
	Limit     int            `json:"limit,omitempty"`
	Filters   map[string]any `json:"filters,omitempty"`
}

type SearchResponse struct {
	Hits        []SearchHit `json:"hits"`
	Total       int         `json:"total"`
	QueryTimeMS int         `json:"query_time_ms"`
}

type DeflectRequest struct {
	ProjectID  string `json:"project_id"`
	TicketText string `json:"ticket_text"`
	TopK       int    `json:"top_k,omitempty"`
}

type DeflectResponse struct {
	Suggestions []DeflectSuggestion `json:"suggestions"`
	Deflected   bool                `json:"deflected"`
}

type DeflectEventRequest struct {
	ProjectID    string         `json:"project_id"`
	EventType    string         `json:"event_type"` // "deflected", "escalated", etc.
	SuggestionID string         `json:"suggestion_id,omitempty"`
	ThreadID     string         `json:"thread_id,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type IngestRequest struct {
	ProjectID      string         `json:"project_id"`
	Source         map[string]any `json:"source"`                   // {"type": "url", "url": "..."}
	ChunkStrategy  string         `json:"chunk_strategy,omitempty"` // "semantic", "fixed", etc.
	ChunkSizeToken int            `json:"chunk_size_token,omitempty"`
}

type IngestResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"` // "queued", "processing", "completed"
	ProjectID string `json:"project_id"`
}

type AnalyticsResponse struct {
	ProjectID   string           `json:"project_id"`
	Summary     AnalyticsSummary `json:"summary"`
	DateRange   map[string]any   `json:"date_range"`
	Integration string           `json:"integration"`
}

type GapsResponse struct {
	Gaps  []GapCluster `json:"gaps"`
	Total int          `json:"total"`
}

type GapDetailResponse struct {
	Cluster  GapCluster          `json:"cluster"`
	Examples []GapClusterExample `json:"examples"`
}
