package api

import (
	"context"
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
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamFrame, error)
}

type SearchService interface {
	Search(ctx context.Context, projectID, query string, topK int, filters map[string]any) ([]SearchHit, error)
}

type DeflectService interface {
	Suggest(ctx context.Context, projectID, subject, body string, topK int) (string, []DeflectSuggestion, error)
	TrackEvent(ctx context.Context, projectID, suggestionID, action, threadID string, metadata map[string]any) error
}

type AnalyticsService interface {
	Summary(ctx context.Context, projectID string, from, to *time.Time, integration string) (AnalyticsSummary, error)
}

type GapsService interface {
	Run(ctx context.Context, projectID, window string) (string, error)
	List(ctx context.Context, projectID string) ([]GapCluster, error)
	Get(ctx context.Context, projectID, clusterID string) (GapCluster, []GapClusterExample, error)
}

// Request/response DTOs align with OpenAPI.
type ChatRequest struct {
	ProjectID      string         `json:"project_id"`
	Query          string         `json:"query"`
	UserID         string         `json:"user_id,omitempty"`
	Mode           string         `json:"mode,omitempty"`
	ContextFilters map[string]any `json:"context_filters,omitempty"`
	TopK           int            `json:"top_k,omitempty"`
	ThreadID       string         `json:"thread_id,omitempty"`
}

type ThreadCreateRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id,omitempty"`
}

type ChatResponse struct {
	ThreadID    string   `json:"thread_id"`
	Answer      string   `json:"answer"`
	IsUncertain bool     `json:"is_uncertain"`
	Citations   []string `json:"citations"`
	Confidence  float32  `json:"confidence"`
}

type StreamFrame struct {
	Type string         `json:"type"` // "token", "done", "error"
	Data map[string]any `json:"data"`
}

type SearchHit struct {
	ChunkID    string  `json:"chunk_id"`
	Text       string  `json:"text"`
	DocumentID string  `json:"document_id"`
	Confidence float32 `json:"confidence"`
}

type DeflectSuggestion struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Relevance float32 `json:"relevance"`
	Rank      int     `json:"rank"`
}

type AnalyticsSummary struct {
	ProjectID      string  `json:"project_id"`
	TotalChats     int     `json:"total_chats"`
	TotalSearches  int     `json:"total_searches"`
	TotalDeflected int     `json:"total_deflected"`
	AvgConfidence  float32 `json:"avg_confidence"`
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

// Services container holds all service implementations for dependency injection
type Services struct {
	Chat      ChatService
	Search    SearchService
	Deflect   DeflectService
	Analytics AnalyticsService
	Gaps      GapsService
	Queue     interface{} // queue.Producer
}
