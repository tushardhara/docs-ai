package search_test

import (
	"context"
	"testing"

	"cgap/internal/search"
	"cgap/internal/service"
	"cgap/internal/testutil"
)

// MockSearch implements service.Search for testing
type MockSearch struct {
	Results     []service.SearchResult
	SearchError error
}

func (m *MockSearch) Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]service.SearchResult, error) {
	if m.SearchError != nil {
		return nil, m.SearchError
	}
	if len(m.Results) > topK {
		return m.Results[:topK], nil
	}
	return m.Results, nil
}

func TestHybrid_Search_PrimarySuccess(t *testing.T) {
	primary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "1", Text: "Primary result 1", Score: 0.95},
			{ID: "2", Text: "Primary result 2", Score: 0.85},
		},
	}

	secondary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "3", Text: "Secondary result 1", Score: 0.75},
		},
	}

	hybrid := search.NewHybrid(primary, secondary)

	ctx := context.Background()
	results, err := hybrid.Search(ctx, "test-index", "test query", 10, map[string]any{})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}

	// First result should be from primary
	if results[0].ID != "1" {
		t.Errorf("Expected first result from primary, got ID %s", results[0].ID)
	}
}

func TestHybrid_Search_PrimaryFailsSecondaryWorks(t *testing.T) {
	primary := &MockSearch{
		SearchError: testutil.ErrTestError,
	}

	secondary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "1", Text: "Secondary result", Score: 0.75},
		},
	}

	hybrid := search.NewHybrid(primary, secondary)

	ctx := context.Background()
	results, err := hybrid.Search(ctx, "test-index", "test query", 10, map[string]any{})

	if err != nil {
		t.Fatalf("Search should succeed with secondary: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result from secondary, got %d", len(results))
	}
}

func TestHybrid_Search_BothFail(t *testing.T) {
	primary := &MockSearch{
		SearchError: testutil.ErrTestError,
	}

	secondary := &MockSearch{
		SearchError: testutil.ErrTestError,
	}

	hybrid := search.NewHybrid(primary, secondary)

	ctx := context.Background()
	_, err := hybrid.Search(ctx, "test-index", "test query", 10, map[string]any{})

	if err == nil {
		t.Error("Expected error when both searches fail")
	}
}

func TestHybrid_Search_Deduplication(t *testing.T) {
	primary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "1", Text: "Result 1", Score: 0.95},
			{ID: "2", Text: "Result 2", Score: 0.85},
		},
	}

	secondary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "1", Text: "Result 1 duplicate", Score: 0.75}, // Duplicate
			{ID: "3", Text: "Result 3", Score: 0.65},
		},
	}

	hybrid := search.NewHybrid(primary, secondary)

	ctx := context.Background()
	results, err := hybrid.Search(ctx, "test-index", "test query", 10, map[string]any{})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should have 3 unique results (1, 2, 3), not 4
	if len(results) != 3 {
		t.Errorf("Expected 3 unique results, got %d", len(results))
	}

	// Verify no duplicate IDs
	seen := make(map[string]bool)
	for _, r := range results {
		if seen[r.ID] {
			t.Errorf("Found duplicate result ID: %s", r.ID)
		}
		seen[r.ID] = true
	}
}

func TestHybrid_Search_TopKLimit(t *testing.T) {
	primary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "1", Score: 0.95},
			{ID: "2", Score: 0.85},
			{ID: "3", Score: 0.75},
		},
	}

	secondary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "4", Score: 0.65},
			{ID: "5", Score: 0.55},
		},
	}

	hybrid := search.NewHybrid(primary, secondary)

	ctx := context.Background()
	results, err := hybrid.Search(ctx, "test-index", "test query", 3, map[string]any{})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) > 3 {
		t.Errorf("Expected at most 3 results (topK=3), got %d", len(results))
	}
}

func TestHybrid_Search_DefaultTopK(t *testing.T) {
	primary := &MockSearch{
		Results: []service.SearchResult{
			{ID: "1", Score: 0.95},
		},
	}

	secondary := &MockSearch{
		Results: []service.SearchResult{},
	}

	hybrid := search.NewHybrid(primary, secondary)

	ctx := context.Background()
	results, err := hybrid.Search(ctx, "test-index", "test query", 0, map[string]any{})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should default to 10 and return what's available
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestNewHybrid(t *testing.T) {
	primary := &MockSearch{}
	secondary := &MockSearch{}

	hybrid := search.NewHybrid(primary, secondary)

	if hybrid == nil {
		t.Error("Expected non-nil hybrid search instance")
	}
}
