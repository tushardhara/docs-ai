package testutil

import (
	"context"
	"errors"
	"sync"
	"time"

	"cgap/api"

	"github.com/redis/go-redis/v9"
)

// Test error for mocking failures
var ErrTestError = errors.New("test error")

// MockLLM provides a mock LLM implementation for testing
type MockLLM struct {
	ResponseText string
	Confidence   float32
	Error        error
}

func (m *MockLLM) Call(ctx context.Context, prompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.ResponseText, nil
}

// MockSearch provides a mock search implementation for testing
type MockSearch struct {
	Results []api.SearchHit
	Error   error
	mu      sync.Mutex
}

func (m *MockSearch) Search(ctx context.Context, projectID, query string, topK int, filters map[string]any) ([]api.SearchHit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Error != nil {
		return nil, m.Error
	}
	return m.Results, nil
}

// MockChatService provides a mock chat service for testing
type MockChatService struct {
	Response api.ChatResponse
	Error    error
}

func (m *MockChatService) Chat(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
	if m.Error != nil {
		return api.ChatResponse{}, m.Error
	}
	return m.Response, nil
}

func (m *MockChatService) ChatStream(ctx context.Context, req api.ChatRequest) (<-chan api.StreamFrame, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	ch := make(chan api.StreamFrame, 1)
	close(ch)
	return ch, nil
}

// MockSearchService provides a mock search service for testing
type MockSearchService struct {
	Hits  []api.SearchHit
	Error error
	mu    sync.Mutex
}

func (m *MockSearchService) Search(ctx context.Context, projectID, query string, topK int, filters map[string]any) ([]api.SearchHit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Error != nil {
		return nil, m.Error
	}
	return m.Hits, nil
}

// MockDeflectService provides a mock deflect service for testing
type MockDeflectService struct {
	Suggestions []api.DeflectSuggestion
	Deflected   bool
	Error       error
}

func (m *MockDeflectService) Suggest(ctx context.Context, projectID, subject, body string, topK int) (string, []api.DeflectSuggestion, error) {
	if m.Error != nil {
		return "", nil, m.Error
	}
	return "mock reason", m.Suggestions, nil
}

func (m *MockDeflectService) TrackEvent(ctx context.Context, projectID, suggestionID, action, threadID string, metadata map[string]any) error {
	if m.Error != nil {
		return m.Error
	}
	return nil
}

// MockAnalyticsService provides a mock analytics service for testing
type MockAnalyticsService struct {
	SummaryData api.AnalyticsSummary
	Error       error
}

func (m *MockAnalyticsService) Summary(ctx context.Context, projectID string, from, to *time.Time, integration string) (api.AnalyticsSummary, error) {
	if m.Error != nil {
		return api.AnalyticsSummary{}, m.Error
	}
	return m.SummaryData, nil
}

// MockGapsService provides a mock gaps service for testing
type MockGapsService struct {
	Gaps  []api.GapCluster
	Error error
}

func (m *MockGapsService) FindGaps(ctx context.Context, projectID string, topK int) ([]api.GapCluster, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Gaps, nil
}

// MockDBPinger provides a mock database pinger
type MockDBPinger struct {
	PingError error
}

func (m *MockDBPinger) Ping(ctx context.Context) error {
	return m.PingError
}

// MockRedisPinger provides a mock redis pinger
type MockRedisPinger struct {
	StatusCmd *redis.StatusCmd
}

func (m *MockRedisPinger) Ping(ctx context.Context) *redis.StatusCmd {
	return m.StatusCmd
}

// MockMeiliChecker provides a mock Meilisearch health checker
type MockMeiliChecker struct {
	HealthError error
}

func (m *MockMeiliChecker) Health(ctx context.Context) error {
	return m.HealthError
}
