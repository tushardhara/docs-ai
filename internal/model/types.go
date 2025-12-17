package model

import "time"

// Core domain models aligned with schema.
type Project struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Slug         string         `json:"slug"`
	DefaultModel string         `json:"default_model,omitempty"`
	Settings     map[string]any `json:"settings,omitempty"`
	UsagePlan    string         `json:"usage_plan,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name,omitempty"`
	AuthProvider string    `json:"auth_provider,omitempty"`
	PictureURL   string    `json:"picture_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProjectMember struct {
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type APIKey struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Source struct {
	ID        string         `json:"id"`
	ProjectID string         `json:"project_id"`
	Type      string         `json:"type"`
	Config    map[string]any `json:"config"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Document struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	SourceID    string    `json:"source_id"`
	URI         string    `json:"uri"`
	Title       string    `json:"title"`
	Lang        string    `json:"lang"`
	Version     string    `json:"version"`
	Hash        string    `json:"hash"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Chunk struct {
	ID          string    `json:"id"`
	DocumentID  string    `json:"document_id"`
	Ord         int       `json:"ord"`
	Text        string    `json:"text"`
	TokenCount  int       `json:"token_count"`
	SectionPath string    `json:"section_path"`
	ScoreRaw    float32   `json:"score_raw"`
	CreatedAt   time.Time `json:"created_at"`
}

type Thread struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Integration string    `json:"integration"`
	ExternalRef string    `json:"external_ref"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Message struct {
	ID        string         `json:"id"`
	ThreadID  string         `json:"thread_id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Meta      map[string]any `json:"meta"`
	LatencyMS int            `json:"latency_ms"`
	CreatedAt time.Time      `json:"created_at"`
}

type Answer struct {
	MessageID      string         `json:"message_id"`
	Model          string         `json:"model"`
	IsUncertain    bool           `json:"is_uncertain"`
	ReasoningTrace map[string]any `json:"reasoning_trace"`
	PromptVersion  string         `json:"prompt_version"`
}

type Citation struct {
	ID        string  `json:"id"`
	AnswerID  string  `json:"answer_id"`
	ChunkID   string  `json:"chunk_id"`
	Score     float32 `json:"score"`
	Quote     string  `json:"quote"`
	StartChar int     `json:"start_char"`
	EndChar   int     `json:"end_char"`
}

type Feedback struct {
	ID        string    `json:"id"`
	AnswerID  string    `json:"answer_id"`
	Type      string    `json:"type"`
	Comment   string    `json:"comment"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type DeflectEvent struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	SessionID     string    `json:"session_id"`
	Subject       string    `json:"subject"`
	Body          string    `json:"body"`
	SuggestionIDs []string  `json:"suggestion_ids"`
	Action        string    `json:"action"`
	ThreadID      string    `json:"thread_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type AnalyticsEvent struct {
	ID         string         `json:"id"`
	ProjectID  string         `json:"project_id"`
	ThreadID   string         `json:"thread_id"`
	MessageID  string         `json:"message_id"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	OccurredAt time.Time      `json:"occurred_at"`
}

type GapCandidate struct {
	AnswerID          string    `json:"answer_id"`
	QuestionEmbedding []float32 `json:"question_embedding"`
	UncertaintyReason string    `json:"uncertainty_reason"`
}

type GapCluster struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	Window         string    `json:"window"`
	Label          string    `json:"label"`
	Summary        string    `json:"summary"`
	Recommendation string    `json:"recommendation"`
	Size           int       `json:"size"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type GapClusterExample struct {
	ID                  string   `json:"id"`
	ClusterID           string   `json:"cluster_id"`
	AnswerID            string   `json:"answer_id"`
	Question            string   `json:"question"`
	Citations           []string `json:"citations"`
	RepresentativeScore float32  `json:"representative_score"`
}
