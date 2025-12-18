package service_test

import (
	"context"
	"testing"

	"cgap/api"
	"cgap/internal/model"
	"cgap/internal/service"
	"cgap/internal/storage"
	"cgap/internal/testutil"

	"github.com/google/uuid"
)

// MockProjectRepo implements storage.ProjectRepo for testing
type MockProjectRepo struct{}

func (m *MockProjectRepo) GetByID(ctx context.Context, id string) (*model.Project, error) {
	return nil, nil
}
func (m *MockProjectRepo) GetBySlug(ctx context.Context, slug string) (*model.Project, error) {
	return nil, nil
}
func (m *MockProjectRepo) Create(ctx context.Context, p *model.Project) error { return nil }
func (m *MockProjectRepo) Update(ctx context.Context, p *model.Project) error { return nil }

// MockDocumentRepo implements storage.DocumentRepo for testing
type MockDocumentRepo struct{}

func (m *MockDocumentRepo) GetByID(ctx context.Context, id string) (*model.Document, error) {
	return nil, nil
}
func (m *MockDocumentRepo) GetByURI(ctx context.Context, projectID, uri string) (*model.Document, error) {
	return nil, nil
}
func (m *MockDocumentRepo) Create(ctx context.Context, d *model.Document) error { return nil }
func (m *MockDocumentRepo) List(ctx context.Context, projectID string, limit, offset int) ([]*model.Document, error) {
	return nil, nil
}

// MockChunkRepo implements storage.ChunkRepo for testing
type MockChunkRepo struct{}

func (m *MockChunkRepo) GetByID(ctx context.Context, id string) (*model.Chunk, error) {
	return nil, nil
}
func (m *MockChunkRepo) CreateBatch(ctx context.Context, chunks []*model.Chunk) error { return nil }
func (m *MockChunkRepo) ListByDocument(ctx context.Context, documentID string) ([]*model.Chunk, error) {
	return nil, nil
}

// MockThreadRepo implements storage.ThreadRepo for testing
type MockThreadRepo struct{}

func (m *MockThreadRepo) GetByID(ctx context.Context, id string) (*model.Thread, error) {
	return nil, nil
}
func (m *MockThreadRepo) Create(ctx context.Context, t *model.Thread) error { return nil }
func (m *MockThreadRepo) Update(ctx context.Context, t *model.Thread) error { return nil }

// MockMessageRepo implements storage.MessageRepo for testing
type MockMessageRepo struct{}

func (m *MockMessageRepo) GetByID(ctx context.Context, id string) (*model.Message, error) {
	return nil, nil
}
func (m *MockMessageRepo) Create(ctx context.Context, msg *model.Message) error { return nil }
func (m *MockMessageRepo) ListByThread(ctx context.Context, threadID string, limit, offset int) ([]*model.Message, error) {
	return nil, nil
}

// MockAnswerRepo implements storage.AnswerRepo for testing
type MockAnswerRepo struct{}

func (m *MockAnswerRepo) Create(ctx context.Context, a *model.Answer) error { return nil }
func (m *MockAnswerRepo) GetByMessageID(ctx context.Context, messageID string) (*model.Answer, error) {
	return nil, nil
}

// MockCitationRepo implements storage.CitationRepo for testing
type MockCitationRepo struct{}

func (m *MockCitationRepo) CreateBatch(ctx context.Context, citations []*model.Citation) error {
	return nil
}
func (m *MockCitationRepo) ListByAnswer(ctx context.Context, answerID string) ([]*model.Citation, error) {
	return nil, nil
}

// MockAnalyticsRepo implements storage.AnalyticsRepo for testing
type MockAnalyticsRepo struct{}

func (m *MockAnalyticsRepo) RecordEvent(ctx context.Context, e *model.AnalyticsEvent) error {
	return nil
}
func (m *MockAnalyticsRepo) CountQuestions(ctx context.Context, projectID string, from, to string) (int, error) {
	return 0, nil
}
func (m *MockAnalyticsRepo) CountUncertain(ctx context.Context, projectID string, from, to string) (int, error) {
	return 0, nil
}

// MockGapRepo implements storage.GapRepo for testing
type MockGapRepo struct{}

func (m *MockGapRepo) CreateCandidate(ctx context.Context, gc *model.GapCandidate) error { return nil }
func (m *MockGapRepo) CreateCluster(ctx context.Context, gc *model.GapCluster) error     { return nil }
func (m *MockGapRepo) CreateExample(ctx context.Context, gce *model.GapClusterExample) error {
	return nil
}
func (m *MockGapRepo) ListClusters(ctx context.Context, projectID, window string) ([]*model.GapCluster, error) {
	return nil, nil
}
func (m *MockGapRepo) GetClusterDetail(ctx context.Context, clusterID string) (*model.GapCluster, []*model.GapClusterExample, error) {
	return nil, nil, nil
}

// MockStore implements storage.Store interface for testing
type MockStore struct {
	StoreError error
}

func (m *MockStore) Projects() storage.ProjectRepo    { return &MockProjectRepo{} }
func (m *MockStore) Documents() storage.DocumentRepo  { return &MockDocumentRepo{} }
func (m *MockStore) Chunks() storage.ChunkRepo        { return &MockChunkRepo{} }
func (m *MockStore) Threads() storage.ThreadRepo      { return &MockThreadRepo{} }
func (m *MockStore) Messages() storage.MessageRepo    { return &MockMessageRepo{} }
func (m *MockStore) Answers() storage.AnswerRepo      { return &MockAnswerRepo{} }
func (m *MockStore) Citations() storage.CitationRepo  { return &MockCitationRepo{} }
func (m *MockStore) Analytics() storage.AnalyticsRepo { return &MockAnalyticsRepo{} }
func (m *MockStore) Gaps() storage.GapRepo            { return &MockGapRepo{} }
func (m *MockStore) Close() error {
	return m.StoreError
}

// MockLLM implements service.LLM interface for testing
type MockLLM struct {
	ChatResponse string
	ChatError    error
	StreamError  error
	StreamTokens []string
}

func (m *MockLLM) Chat(ctx context.Context, messages []service.Message) (string, error) {
	if m.ChatError != nil {
		return "", m.ChatError
	}
	return m.ChatResponse, nil
}

func (m *MockLLM) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	if m.StreamError != nil {
		return nil, m.StreamError
	}

	ch := make(chan string, len(m.StreamTokens))
	for _, token := range m.StreamTokens {
		ch <- token
	}
	close(ch)
	return ch, nil
}

// MockSearch implements service.Search interface for testing
type MockSearch struct {
	Results     []service.SearchResult
	SearchError error
}

func (m *MockSearch) Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]service.SearchResult, error) {
	if m.SearchError != nil {
		return nil, m.SearchError
	}
	return m.Results, nil
}

// ============ Chat Service Tests ============

func TestChatService_Chat_Success(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	query := "How do I create a dashboard?"

	mockSearch := &MockSearch{
		Results: []service.SearchResult{
			{
				ID:   uuid.New().String(),
				Text: "To create a dashboard, go to the dashboard page and click add.",
				Metadata: map[string]any{
					"document_id": "doc-123",
				},
				Score: 0.95,
			},
		},
	}

	mockLLM := &MockLLM{
		ChatResponse: "Follow these steps to create a dashboard...",
	}

	mockStore := &MockStore{}

	chatSvc := service.NewChatService(mockStore, mockLLM, mockSearch)

	req := api.ChatRequest{
		ProjectID: projectID,
		Query:     query,
		TopK:      5,
	}

	resp, err := chatSvc.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Answer == "" {
		t.Error("Expected non-empty answer")
	}
	if len(resp.Citations) == 0 {
		t.Error("Expected at least one citation")
	}
	if resp.Confidence == 0 {
		t.Error("Expected non-zero confidence")
	}
}

func TestChatService_Chat_SearchError(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockSearch := &MockSearch{
		SearchError: testutil.ErrTestError,
	}

	mockLLM := &MockLLM{}
	mockStore := &MockStore{}

	chatSvc := service.NewChatService(mockStore, mockLLM, mockSearch)

	req := api.ChatRequest{
		ProjectID: projectID,
		Query:     "test",
	}

	_, err := chatSvc.Chat(ctx, req)
	if err == nil {
		t.Error("Expected error from search failure")
	}
}

func TestChatService_Chat_LLMError(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockSearch := &MockSearch{
		Results: []service.SearchResult{
			{
				ID:   uuid.New().String(),
				Text: "Test content",
				Metadata: map[string]any{
					"document_id": "doc-1",
				},
				Score: 0.9,
			},
		},
	}

	mockLLM := &MockLLM{
		ChatError: testutil.ErrTestError,
	}

	mockStore := &MockStore{}

	chatSvc := service.NewChatService(mockStore, mockLLM, mockSearch)

	req := api.ChatRequest{
		ProjectID: projectID,
		Query:     "test",
	}

	_, err := chatSvc.Chat(ctx, req)
	if err == nil {
		t.Error("Expected error from LLM failure")
	}
}

func TestChatService_Chat_NoResults(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockSearch := &MockSearch{
		Results: []service.SearchResult{},
	}

	mockLLM := &MockLLM{
		ChatResponse: "I don't have relevant information to answer this question.",
	}

	mockStore := &MockStore{}

	chatSvc := service.NewChatService(mockStore, mockLLM, mockSearch)

	req := api.ChatRequest{
		ProjectID: projectID,
		Query:     "test",
	}

	resp, err := chatSvc.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat should not fail on empty search: %v", err)
	}

	if resp.Answer == "" {
		t.Error("Expected answer even with no search results")
	}
}

func TestChatService_ChatStream_Success(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockSearch := &MockSearch{
		Results: []service.SearchResult{
			{
				ID:   uuid.New().String(),
				Text: "Stream test content",
				Metadata: map[string]any{
					"document_id": "doc-stream",
				},
				Score: 0.85,
			},
		},
	}

	mockLLM := &MockLLM{
		StreamTokens: []string{"This", " is", " a", " streaming", " response"},
	}

	mockStore := &MockStore{}

	chatSvc := service.NewChatService(mockStore, mockLLM, mockSearch)

	req := api.ChatRequest{
		ProjectID: projectID,
		Query:     "stream test",
	}

	ch, err := chatSvc.ChatStream(ctx, req)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	tokenCount := 0
	doneReceived := false

	for frame := range ch {
		if frame.Type == "token" {
			tokenCount++
		} else if frame.Type == "done" {
			doneReceived = true
		}
	}

	if !doneReceived {
		t.Error("Expected done frame")
	}
	if tokenCount == 0 {
		t.Error("Expected at least one token")
	}
}

// ============ Search Service Tests ============

func TestSearchService_Search_Success(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	query := "dashboard"

	mockSearch := &MockSearch{
		Results: []service.SearchResult{
			{
				ID:   "chunk-1",
				Text: "Dashboard overview text",
				Metadata: map[string]any{
					"document_id": "doc-1",
				},
				Score: 0.95,
			},
			{
				ID:   "chunk-2",
				Text: "Dashboard creation guide",
				Metadata: map[string]any{
					"document_id": "doc-2",
				},
				Score: 0.87,
			},
		},
	}

	mockStore := &MockStore{}

	searchSvc := service.NewSearchService(mockStore, mockSearch)

	hits, err := searchSvc.Search(ctx, projectID, query, 10, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(hits) != 2 {
		t.Errorf("Expected 2 hits, got %d", len(hits))
	}

	if hits[0].ChunkID != "chunk-1" {
		t.Errorf("Expected chunk-1, got %s", hits[0].ChunkID)
	}

	if hits[0].DocumentID != "doc-1" {
		t.Errorf("Expected doc-1, got %s", hits[0].DocumentID)
	}

	if hits[0].Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", hits[0].Confidence)
	}
}

func TestSearchService_Search_Error(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockSearch := &MockSearch{
		SearchError: testutil.ErrTestError,
	}

	mockStore := &MockStore{}

	searchSvc := service.NewSearchService(mockStore, mockSearch)

	_, err := searchSvc.Search(ctx, projectID, "test", 10, nil)
	if err == nil {
		t.Error("Expected search error")
	}
}

func TestSearchService_Search_EmptyResults(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockSearch := &MockSearch{
		Results: []service.SearchResult{},
	}

	mockStore := &MockStore{}

	searchSvc := service.NewSearchService(mockStore, mockSearch)

	hits, err := searchSvc.Search(ctx, projectID, "nonexistent", 10, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(hits) != 0 {
		t.Errorf("Expected 0 hits, got %d", len(hits))
	}
}

// ============ Deflect Service Tests ============

func TestDeflectService_Suggest_Success(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	subject := "Create Dashboard"
	body := "How do I create a new dashboard?"

	mockSearch := &MockSearch{
		Results: []service.SearchResult{
			{
				ID:   uuid.New().String(),
				Text: "How to create a dashboard - step by step guide with detailed instructions and more info here",
				Metadata: map[string]any{
					"document_id": "doc-guide",
				},
				Score: 0.98,
			},
			{
				ID:   uuid.New().String(),
				Text: "Dashboard templates and presets available for your organization and team members here",
				Metadata: map[string]any{
					"document_id": "doc-templates",
				},
				Score: 0.85,
			},
		},
	}

	mockLLM := &MockLLM{}
	mockStore := &MockStore{}

	deflectSvc := service.NewDeflectService(mockStore, mockSearch, mockLLM)

	query, suggestions, err := deflectSvc.Suggest(ctx, projectID, subject, body, 5)
	if err != nil {
		t.Fatalf("Suggest failed: %v", err)
	}

	if query == "" {
		t.Error("Expected non-empty query")
	}

	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(suggestions))
	}

	if suggestions[0].Relevance != 0.98 {
		t.Errorf("Expected relevance 0.98, got %f", suggestions[0].Relevance)
	}

	if suggestions[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", suggestions[0].Rank)
	}
}

func TestDeflectService_Suggest_SearchError(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockSearch := &MockSearch{
		SearchError: testutil.ErrTestError,
	}

	mockLLM := &MockLLM{}
	mockStore := &MockStore{}

	deflectSvc := service.NewDeflectService(mockStore, mockSearch, mockLLM)

	_, _, err := deflectSvc.Suggest(ctx, projectID, "test", "test body", 5)
	if err == nil {
		t.Error("Expected error from search failure")
	}
}

func TestDeflectService_TrackEvent(t *testing.T) {
	ctx := context.Background()

	mockSearch := &MockSearch{}
	mockLLM := &MockLLM{}
	mockStore := &MockStore{}

	deflectSvc := service.NewDeflectService(mockStore, mockSearch, mockLLM)

	err := deflectSvc.TrackEvent(ctx, "proj-1", "sugg-1", "clicked", "thread-1", map[string]any{})
	if err != nil {
		t.Fatalf("TrackEvent failed: %v", err)
	}
}

// ============ Analytics Service Tests ============

func TestAnalyticsService_Summary(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockStore := &MockStore{}
	analyticsSvc := service.NewAnalyticsService(mockStore)

	summary, err := analyticsSvc.Summary(ctx, projectID, nil, nil, "")
	if err != nil {
		t.Fatalf("Summary failed: %v", err)
	}

	if summary.ProjectID != projectID {
		t.Errorf("Expected project_id %s, got %s", projectID, summary.ProjectID)
	}
}

// ============ Gaps Service Tests ============

func TestGapsService_Run(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockStore := &MockStore{}
	mockLLM := &MockLLM{}

	gapsSvc := service.NewGapsService(mockStore, mockLLM)

	jobID, err := gapsSvc.Run(ctx, projectID, "daily")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if jobID == "" {
		t.Error("Expected non-empty job ID")
	}
}

func TestGapsService_List(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	mockStore := &MockStore{}
	mockLLM := &MockLLM{}

	gapsSvc := service.NewGapsService(mockStore, mockLLM)

	clusters, err := gapsSvc.List(ctx, projectID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if clusters == nil {
		clusters = []api.GapCluster{}
	}
}
