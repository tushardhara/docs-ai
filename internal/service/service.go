package service

import (
	"context"

	"cgap/api"
	"cgap/internal/storage"
)

// ChatService implementation (skeleton).
type ChatServiceImpl struct {
	store  storage.Store
	llm    LLM    // TODO: OpenAI/Anthropic client
	search Search // TODO: Meilisearch client
}

func NewChatService(store storage.Store) *ChatServiceImpl {
	return &ChatServiceImpl{
		store: store,
	}
}

func (s *ChatServiceImpl) Chat(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
	// TODO: Implement
	// 1. Search hybrid (meili + pgvector)
	// 2. Call LLM with context
	// 3. Parse citations
	// 4. Store thread + message + answer + citations
	// 5. Return response
	return api.ChatResponse{}, nil
}

func (s *ChatServiceImpl) ChatStream(ctx context.Context, req api.ChatRequest) (<-chan api.StreamFrame, error) {
	// TODO: Implement SSE streaming
	// Similar to Chat but yields frames as LLM streams tokens
	ch := make(chan api.StreamFrame)
	return ch, nil
}

// SearchService implementation (skeleton).
type SearchServiceImpl struct {
	store  storage.Store
	search Search
}

func NewSearchService(store storage.Store) *SearchServiceImpl {
	return &SearchServiceImpl{
		store: store,
	}
}

func (s *SearchServiceImpl) Search(ctx context.Context, projectID, query string, topK int, filters map[string]any) ([]api.SearchHit, error) {
	// TODO: Implement
	// 1. Query Meilisearch
	// 2. Optionally re-rank with pgvector
	// 3. Return hits
	return nil, nil
}

// DeflectService implementation (skeleton).
type DeflectServiceImpl struct {
	store  storage.Store
	search Search
	llm    LLM
}

func NewDeflectService(store storage.Store) *DeflectServiceImpl {
	return &DeflectServiceImpl{
		store: store,
	}
}

func (s *DeflectServiceImpl) Suggest(ctx context.Context, projectID, subject, body string, topK int) (string, []api.DeflectSuggestion, error) {
	// TODO: Implement
	// 1. Combine subject + body -> query
	// 2. Call Search
	// 3. Return top suggestions with citations
	return "", nil, nil
}

func (s *DeflectServiceImpl) TrackEvent(ctx context.Context, projectID, suggestionID, action, threadID string, metadata map[string]any) error {
	// TODO: Log deflection event to analytics
	return nil
}

// AnalyticsService implementation (skeleton).
type AnalyticsServiceImpl struct {
	store storage.Store
}

func NewAnalyticsService(store storage.Store) *AnalyticsServiceImpl {
	return &AnalyticsServiceImpl{
		store: store,
	}
}

func (s *AnalyticsServiceImpl) Summary(ctx context.Context, projectID string, from, to *string, integration string) (api.AnalyticsSummary, error) {
	// TODO: Implement
	// Query analytics_events for summary stats
	return api.AnalyticsSummary{}, nil
}

// GapsService implementation (skeleton).
type GapsServiceImpl struct {
	store storage.Store
	llm   LLM
}

func NewGapsService(store storage.Store) *GapsServiceImpl {
	return &GapsServiceImpl{
		store: store,
	}
}

func (s *GapsServiceImpl) Run(ctx context.Context, projectID, window string) (string, error) {
	// TODO: Implement gap clustering job
	return "", nil
}

func (s *GapsServiceImpl) List(ctx context.Context, projectID string) ([]api.GapCluster, error) {
	// TODO: Implement
	return nil, nil
}

func (s *GapsServiceImpl) Get(ctx context.Context, projectID, clusterID string) (api.GapCluster, []api.GapClusterExample, error) {
	// TODO: Implement
	return api.GapCluster{}, nil, nil
}

// LLM interface for pluggable LLM clients.
type LLM interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	Stream(ctx context.Context, messages []Message) (<-chan string, error)
}

// Message represents a chat message.
type Message struct {
	Role    string
	Content string
}

// Search interface for pluggable search clients.
type Search interface {
	Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]SearchResult, error)
}

// SearchResult represents a search hit.
type SearchResult struct {
	ID       string
	Text     string
	Metadata map[string]any
	Score    float32
}
