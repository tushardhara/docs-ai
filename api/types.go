package api

import (
	"context"
	"time"

	"cgap/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
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
	ProjectID      string     `json:"project_id"`
	Source         SourceSpec `json:"source"`                   // {"type": "url"|"crawl"|"github"|"openapi"|"slack"|"discord"|"upload", ...}
	ChunkStrategy  string     `json:"chunk_strategy,omitempty"` // "semantic", "fixed", etc.
	ChunkSizeToken int        `json:"chunk_size_token,omitempty"`
	FailFast       bool       `json:"fail_fast,omitempty"` // If true, stop on first URL error
}

type IngestResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"` // "queued", "processing", "completed"
	ProjectID string `json:"project_id"`
}

// IngestStatusResponse represents the current status of an ingest job
type IngestStatusResponse struct {
	JobID      string `json:"job_id"`
	ProjectID  string `json:"project_id"`
	Status     string `json:"status"`      // queued|running|completed|failed
	Processed  int    `json:"processed"`   // processed units (pages or chunks)
	Total      int    `json:"total"`       // total units if known
	StartedAt  string `json:"started_at"`  // RFC3339
	FinishedAt string `json:"finished_at"` // RFC3339
	Error      string `json:"error,omitempty"`
}

// SourceSpec describes an ingestion source.
type SourceSpec struct {
	Type       string         `json:"type"`             // url|crawl|github|openapi|slack|discord|upload
	Config     map[string]any `json:"config,omitempty"` // arbitrary provider-specific config
	URL        string         `json:"url,omitempty"`    // for url/crawl/openapi
	OpenAPIURL string         `json:"openapi_url,omitempty"`
	Repo       string         `json:"repo,omitempty"` // for github
	Owner      string         `json:"owner,omitempty"`
	Token      string         `json:"token,omitempty"`     // optional access tokens for providers
	UploadID   string         `json:"upload_id,omitempty"` // for upload
	Crawl      *CrawlSpec     `json:"crawl,omitempty"`     // crawl configuration for type=crawl
	Media      *MediaSpec     `json:"media,omitempty"`     // media ingestion for type=image|video|youtube
	Files      *FileSpec      `json:"files,omitempty"`     // document ingestion for type=document|pdf|markdown|txt
}

// IngestTaskPayload is the message body enqueued for ingestion.
type IngestTaskPayload struct {
	ProjectID      string     `json:"project_id"`
	Source         SourceSpec `json:"source"`
	ChunkStrategy  string     `json:"chunk_strategy,omitempty"`
	ChunkSizeToken int        `json:"chunk_size_token,omitempty"`
	FailFast       bool       `json:"fail_fast,omitempty"`
}

// CrawlSpec describes how to fetch web content for web sources.
type CrawlSpec struct {
	// Mode selects strategy: "single" (just one page), "sitemap" (discover from sitemap), "crawl" (follow links).
	Mode string `json:"mode"`
	// StartURL is required for mode=single and crawl.
	StartURL string `json:"start_url,omitempty"`
	// SitemapURL is required for mode=sitemap.
	SitemapURL string `json:"sitemap_url,omitempty"`
	// Scope controls allowed URLs when crawling: "host" (default), "domain", or "prefix".
	Scope string `json:"scope,omitempty"`
	// Allow/Deny are regex or glob patterns to include/exclude paths.
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
	// Limits and politeness.
	MaxDepth      int  `json:"max_depth,omitempty"`
	MaxPages      int  `json:"max_pages,omitempty"`
	RespectRobots bool `json:"respect_robots,omitempty"`
	Concurrency   int  `json:"concurrency,omitempty"`
	DelayMS       int  `json:"delay_ms,omitempty"`
}

// MediaSpec describes image/video ingestion parameters.
type MediaSpec struct {
	// Common
	URLs []string `json:"urls,omitempty"` // direct media URLs

	// Images
	OCR     bool   `json:"ocr,omitempty"`
	OCRLang string `json:"ocr_lang,omitempty"`

	// Video/Audio transcripts
	Transcript         bool     `json:"transcript,omitempty"`
	TranscriptProvider string   `json:"transcript_provider,omitempty"` // youtube|whisper|assemblyai|none
	YouTubeIDs         []string `json:"youtube_ids,omitempty"`
	MaxDurationSec     int      `json:"max_duration_sec,omitempty"`
}

// FileSpec describes document/file ingestion parameters.
type FileSpec struct {
	URLs    []string `json:"urls,omitempty"`    // one or more document URLs
	Format  string   `json:"format,omitempty"`  // pdf|txt|markdown|md|auto (default auto by extension/content-type)
	Extract string   `json:"extract,omitempty"` // reader to use; e.g., pdfium, tika, md, plain
}

// Dev-only seeding endpoint to insert a document, chunk and embedding
type SeedRequest struct {
	ProjectID string `json:"project_id"` // slug or UUID
	URI       string `json:"uri"`
	Title     string `json:"title"`
	Text      string `json:"text"`
}

type SeedResponse struct {
	ProjectID  string `json:"project_id"`
	DocumentID string `json:"document_id"`
	ChunkID    string `json:"chunk_id"`
	Status     string `json:"status"`
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

// OCR Request/Response types
type OCRRequest struct {
	ProjectID string `json:"project_id"`
	SourceID  string `json:"source_id"`
	ImageURL  string `json:"image_url"`
}

type TextRegion struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	X1         float32 `json:"x1"`
	Y1         float32 `json:"y1"`
	X2         float32 `json:"x2"`
	Y2         float32 `json:"y2"`
}

type OCRResponse struct {
	MediaItemID    string       `json:"media_item_id"`
	Text           string       `json:"text"`
	Confidence     float64      `json:"confidence"`
	Language       string       `json:"language"`
	TextRegions    []TextRegion `json:"text_regions,omitempty"`
	ProcessedAt    string       `json:"processed_at"`
	ExtractionStat string       `json:"extraction_status"` // "success", "partial", "failed"
}

// YouTube Request/Response types
type YouTubeRequest struct {
	ProjectID string `json:"project_id"`
	SourceID  string `json:"source_id"`
	VideoURL  string `json:"video_url"`
}

type TranscriptSegmentResponse struct {
	Text         string `json:"text"`
	StartSeconds int    `json:"start_seconds"`
	EndSeconds   int    `json:"end_seconds"`
}

type YouTubeResponse struct {
	MediaItemID      string                      `json:"media_item_id"`
	Transcript       string                      `json:"transcript"`
	Language         string                      `json:"language"`
	Segments         []TranscriptSegmentResponse `json:"segments,omitempty"`
	IsAutoGenerated  bool                        `json:"is_auto_generated"`
	ProcessedAt      string                      `json:"processed_at"`
	ExtractionStatus string                      `json:"extraction_status"` // "success", "partial", "failed"
}

// Video File Request/Response types (for direct video files: MP4, AVI, MOV, etc.)
type VideoRequest struct {
	ProjectID string `json:"project_id"`
	SourceID  string `json:"source_id"`
	VideoURL  string `json:"video_url"` // URL to video file or file upload path
}

type VideoResponse struct {
	MediaItemID      string                      `json:"media_item_id"`
	Transcript       string                      `json:"transcript"`
	Language         string                      `json:"language"`
	Segments         []TranscriptSegmentResponse `json:"segments,omitempty"`
	IsAutoGenerated  bool                        `json:"is_auto_generated"`
	Duration         int                         `json:"duration_seconds"`
	ProcessedAt      string                      `json:"processed_at"`
	ExtractionStatus string                      `json:"extraction_status"` // "success", "partial", "failed"
}

// Unified Media Processing Request/Response types
type MediaProcessRequest struct {
	ProjectID string `json:"project_id"`
	SourceID  string `json:"source_id"`
	MediaURL  string `json:"media_url"`
	MediaType string `json:"media_type,omitempty"` // Optional: "image", "youtube", "video" - will auto-detect if empty
}

type MediaProcessResponse struct {
	MediaItemID      string                 `json:"media_item_id"`
	MediaType        string                 `json:"media_type"` // Detected or provided type
	Text             string                 `json:"text"`
	Language         string                 `json:"language"`
	Confidence       float64                `json:"confidence"`
	ContentType      string                 `json:"content_type"` // "text", "transcript"
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt      string                 `json:"processed_at"`
	ExtractionStatus string                 `json:"extraction_status"` // "success", "partial", "failed"
}

// ===== Browser Extension API Types =====

// DOMEntity represents an interactive element on the page
type DOMEntity struct {
	Selector string `json:"selector"` // CSS selector (e.g., ".btn-dashboard")
	Type     string `json:"type"`     // "button", "input", "link", "select", etc.
	Text     string `json:"text"`     // Visible text or label
	ID       string `json:"id,omitempty"`
	Class    string `json:"class,omitempty"`
}

// ExtensionChatRequest is the payload from browser extension
type ExtensionChatRequest struct {
	ProjectID  string      `json:"project_id"`
	URL        string      `json:"url"`                  // Current page URL
	Question   string      `json:"question"`             // User's question
	DOM        []DOMEntity `json:"dom"`                  // Parsed DOM entities
	Screenshot string      `json:"screenshot,omitempty"` // Base64 image (optional)
}

// GuidanceStep represents a single action step
type GuidanceStep struct {
	StepNumber  int     `json:"step_number"`
	Description string  `json:"description"`
	Selector    string  `json:"selector,omitempty"`   // CSS selector for this step
	Action      string  `json:"action,omitempty"`     // "click", "type", "select", etc.
	Value       string  `json:"value,omitempty"`      // Value to type/select
	Confidence  float32 `json:"confidence,omitempty"` // 0-1 confidence in this step
}

// ExtensionChatResponse contains guidance for the user
type ExtensionChatResponse struct {
	Guidance    string         `json:"guidance"`               // Natural language explanation
	Steps       []GuidanceStep `json:"steps"`                  // Actionable steps with selectors
	Confidence  float32        `json:"confidence"`             // Overall confidence (0-1)
	Sources     []Citation     `json:"sources"`                // Supporting documentation
	NextActions []string       `json:"next_actions,omitempty"` // Suggested follow-ups
}

// Services container holds all service implementations for dependency injection
type Services struct {
	Chat      ChatService
	Search    SearchService
	Deflect   DeflectService
	Analytics AnalyticsService
	Gaps      GapsService
	Queue     interface{}   // queue.Producer
	DB        *pgxpool.Pool // Database connection pool for media storage
}
