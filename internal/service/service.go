package service

import (
	"context"
	"time"

	"cgap/api"
	"cgap/internal/storage"
)

// ChatService implementation.
type ChatServiceImpl struct {
	store  storage.Store
	llm    LLM
	search Search
}

func NewChatService(store storage.Store, llm LLM, search Search) *ChatServiceImpl {
	return &ChatServiceImpl{
		store:  store,
		llm:    llm,
		search: search,
	}
}

func (s *ChatServiceImpl) Chat(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
	// 1. Search hybrid (meili + pgvector)
	searchResults, err := s.search.Search(ctx, "chunks", req.Query, 5, map[string]any{
		"project_id": req.ProjectID,
	})
	if err != nil {
		return api.ChatResponse{}, err
	}

	// 2. Build context from search results
	var context string
	var citations []string
	for _, result := range searchResults {
		context += result.Text + "\n"
		if docID, ok := result.Metadata["document_id"].(string); ok {
			citations = append(citations, docID)
		}
	}

	// 3. Call LLM with context
	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant. Use the provided context to answer the question."},
		{Role: "user", Content: "Context:\n" + context + "\n\nQuestion: " + req.Query},
	}

	llmResponse, err := s.llm.Chat(ctx, messages)
	if err != nil {
		return api.ChatResponse{}, err
	}

	// 4. Return response (TODO: Store thread + message + answer)
	return api.ChatResponse{
		Answer:     llmResponse,
		Citations:  citations,
		Confidence: 0.8,
	}, nil
}

func (s *ChatServiceImpl) ChatStream(ctx context.Context, req api.ChatRequest) (<-chan api.StreamFrame, error) {
	ch := make(chan api.StreamFrame)

	go func() {
		defer close(ch)

		// Search for context
		searchResults, err := s.search.Search(ctx, "chunks", req.Query, 5, map[string]any{
			"project_id": req.ProjectID,
		})
		if err != nil {
			ch <- api.StreamFrame{Type: "error", Data: map[string]any{"error": err.Error()}}
			return
		}

		// Build context
		var context string
		var citations []string
		for _, result := range searchResults {
			context += result.Text + "\n"
			if docID, ok := result.Metadata["document_id"].(string); ok {
				citations = append(citations, docID)
			}
		}

		// Stream from LLM
		messages := []Message{
			{Role: "system", Content: "You are a helpful assistant. Use the provided context to answer the question."},
			{Role: "user", Content: "Context:\n" + context + "\n\nQuestion: " + req.Query},
		}

		tokenChan, err := s.llm.Stream(ctx, messages)
		if err != nil {
			ch <- api.StreamFrame{Type: "error", Data: map[string]any{"error": err.Error()}}
			return
		}

		for token := range tokenChan {
			ch <- api.StreamFrame{
				Type: "token",
				Data: map[string]any{"token": token},
			}
		}

		ch <- api.StreamFrame{
			Type: "done",
			Data: map[string]any{"citations": citations},
		}
	}()

	return ch, nil
}

// SearchService implementation.
type SearchServiceImpl struct {
	store  storage.Store
	search Search
}

func NewSearchService(store storage.Store, search Search) *SearchServiceImpl {
	return &SearchServiceImpl{
		store:  store,
		search: search,
	}
}

func (s *SearchServiceImpl) Search(ctx context.Context, projectID, query string, topK int, filters map[string]any) ([]api.SearchHit, error) {
	// 1. Query Meilisearch
	results, err := s.search.Search(ctx, "chunks", query, topK, map[string]any{
		"project_id": projectID,
	})
	if err != nil {
		return nil, err
	}

	// 2. Convert to API hits
	var hits []api.SearchHit
	for _, result := range results {
		docID, _ := result.Metadata["document_id"].(string)
		hits = append(hits, api.SearchHit{
			ChunkID:    result.ID,
			Text:       result.Text,
			DocumentID: docID,
			Confidence: result.Score,
		})
	}

	return hits, nil
}

// DeflectService implementation.
type DeflectServiceImpl struct {
	store  storage.Store
	search Search
	llm    LLM
}

func NewDeflectService(store storage.Store, search Search, llm LLM) *DeflectServiceImpl {
	return &DeflectServiceImpl{
		store:  store,
		search: search,
		llm:    llm,
	}
}

func (s *DeflectServiceImpl) Suggest(ctx context.Context, projectID, subject, body string, topK int) (string, []api.DeflectSuggestion, error) {
	// 1. Combine subject + body -> query
	query := subject + " " + body

	// 2. Call Search
	results, err := s.search.Search(ctx, "chunks", query, topK, map[string]any{
		"project_id": projectID,
	})
	if err != nil {
		return "", nil, err
	}

	// 3. Build suggestions
	var suggestions []api.DeflectSuggestion
	for i, result := range results {
		suggestions = append(suggestions, api.DeflectSuggestion{
			ID:        result.ID,
			Title:     result.Text[:50], // Truncate for title
			Relevance: result.Score,
			Rank:      i + 1,
		})
	}

	return query, suggestions, nil
}

func (s *DeflectServiceImpl) TrackEvent(ctx context.Context, projectID, suggestionID, action, threadID string, metadata map[string]any) error {
	// Log deflection event to analytics
	// This would be implemented with the analytics store
	return nil
}

// AnalyticsService implementation.
type AnalyticsServiceImpl struct {
	store storage.Store
}

func NewAnalyticsService(store storage.Store) *AnalyticsServiceImpl {
	return &AnalyticsServiceImpl{
		store: store,
	}
}

func (s *AnalyticsServiceImpl) Summary(ctx context.Context, projectID string, from, to *time.Time, integration string) (api.AnalyticsSummary, error) {
	// Query analytics_events for summary stats
	// This would retrieve data from the store
	return api.AnalyticsSummary{
		ProjectID:      projectID,
		TotalChats:     0,
		TotalSearches:  0,
		TotalDeflected: 0,
		AvgConfidence:  0,
	}, nil
}

// GapsService implementation.
type GapsServiceImpl struct {
	store storage.Store
	llm   LLM
}

func NewGapsService(store storage.Store, llm LLM) *GapsServiceImpl {
	return &GapsServiceImpl{
		store: store,
		llm:   llm,
	}
}

func (s *GapsServiceImpl) Run(ctx context.Context, projectID, window string) (string, error) {
	// Implement gap clustering job - typically async
	// For now, return placeholder
	return "gap_job_" + projectID, nil
}

func (s *GapsServiceImpl) List(ctx context.Context, projectID string) ([]api.GapCluster, error) {
	// Retrieve gap clusters from store
	return nil, nil
}

func (s *GapsServiceImpl) Get(ctx context.Context, projectID, clusterID string) (api.GapCluster, []api.GapClusterExample, error) {
	// Retrieve specific gap cluster and examples
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
