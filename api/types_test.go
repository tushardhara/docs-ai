package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cgap/api"

	"github.com/google/uuid"
)

func TestSearchRequestMarshaling(t *testing.T) {
	req := api.SearchRequest{
		ProjectID: "test-proj",
		Query:     "test query",
		Limit:     10,
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded api.SearchRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.ProjectID != req.ProjectID {
		t.Errorf("ProjectID mismatch: %s != %s", decoded.ProjectID, req.ProjectID)
	}
}

func TestChatRequestMarshaling(t *testing.T) {
	req := api.ChatRequest{
		ProjectID: "test-proj",
		Query:     "test question",
		TopK:      5,
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded api.ChatRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.Query != req.Query {
		t.Errorf("Query mismatch: %s != %s", decoded.Query, req.Query)
	}
}

func TestSearchHitMarshaling(t *testing.T) {
	hit := api.SearchHit{
		ChunkID:    uuid.New().String(),
		Text:       "Sample text",
		DocumentID: uuid.New().String(),
		Confidence: 0.95,
	}

	body, err := json.Marshal(hit)
	if err != nil {
		t.Fatalf("Failed to marshal hit: %v", err)
	}

	var decoded api.SearchHit
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal hit: %v", err)
	}

	if decoded.ChunkID != hit.ChunkID {
		t.Errorf("ChunkID mismatch")
	}
	if decoded.Confidence != hit.Confidence {
		t.Errorf("Confidence mismatch: %f != %f", decoded.Confidence, hit.Confidence)
	}
}

func TestExtensionChatRequestMarshaling(t *testing.T) {
	req := api.ExtensionChatRequest{
		ProjectID: "test-proj",
		URL:       "https://example.com",
		Question:  "How to use this?",
		DOM: []api.DOMEntity{
			{
				Selector: ".btn-test",
				Type:     "button",
				Text:     "Test Button",
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded api.ExtensionChatRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.URL != req.URL {
		t.Errorf("URL mismatch")
	}
	if len(decoded.DOM) != 1 {
		t.Errorf("DOM mismatch")
	}
}

func TestHealthEndpoint(t *testing.T) {
	// Simple test to verify types compile and work
	req := httptest.NewRequest("GET", "/health", nil)
	if req == nil {
		t.Error("Failed to create test request")
	}

	if req.Method != http.MethodGet {
		t.Errorf("Expected GET, got %s", req.Method)
	}
}
