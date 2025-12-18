package testutil

import (
	"time"

	"github.com/google/uuid"
)

// TestProject creates a test project fixture
func TestProject() map[string]any {
	return map[string]any{
		"id":         uuid.New().String(),
		"slug":       "test-project-" + uuid.New().String()[:8],
		"name":       "Test Project",
		"created_at": time.Now().UTC().String(),
	}
}

// TestDocument creates a test document fixture
func TestDocument(projectID string) map[string]any {
	return map[string]any{
		"id":         uuid.New().String(),
		"project_id": projectID,
		"source_id":  uuid.New().String(),
		"title":      "Test Document",
		"uri":        "https://example.com/doc",
		"created_at": time.Now().UTC().String(),
	}
}

// TestChunk creates a test chunk fixture
func TestChunk(documentID string) map[string]any {
	return map[string]any{
		"id":          uuid.New().String(),
		"document_id": documentID,
		"text":        "This is a test chunk of text content",
		"token_count": 7,
		"created_at":  time.Now().UTC().String(),
	}
}

// TestMediaItem creates a test media item fixture
func TestMediaItem(projectID, sourceID string) map[string]any {
	return map[string]any{
		"id":                uuid.New().String(),
		"project_id":        projectID,
		"source_id":         sourceID,
		"type":              "image",
		"url":               "https://example.com/image.png",
		"processing_status": "pending",
		"created_at":        time.Now().UTC().String(),
	}
}

// TestExtractedText creates a test extracted text fixture
func TestExtractedText(mediaItemID string) map[string]any {
	return map[string]any{
		"id":               uuid.New().String(),
		"media_item_id":    mediaItemID,
		"source_type":      "ocr",
		"text":             "Extracted text from media",
		"confidence_score": 0.95,
		"language":         "en",
		"created_at":       time.Now().UTC().String(),
	}
}

// TestSearchHit creates a test search hit fixture
func TestSearchHit() map[string]any {
	return map[string]any{
		"chunk_id":    uuid.New().String(),
		"text":        "Relevant search result text",
		"document_id": uuid.New().String(),
		"confidence":  0.87,
	}
}

// TestChatRequest creates a test chat request
func TestChatRequest(projectID string) map[string]any {
	return map[string]any{
		"project_id": projectID,
		"query":      "How do I create a dashboard?",
		"top_k":      5,
	}
}

// TestChatResponse creates a test chat response
func TestChatResponse(threadID string) map[string]any {
	return map[string]any{
		"thread_id":    threadID,
		"answer":       "To create a dashboard, follow these steps...",
		"is_uncertain": false,
		"citations":    []string{"doc1", "doc2"},
		"confidence":   0.92,
	}
}

// TestExtensionChatRequest creates a test extension chat request
func TestExtensionChatRequest(projectID string) map[string]any {
	return map[string]any{
		"project_id": projectID,
		"url":        "https://example.com/dashboard",
		"question":   "How do I add a widget?",
		"dom": []map[string]any{
			{
				"selector": ".btn-add-widget",
				"type":     "button",
				"text":     "Add Widget",
			},
		},
	}
}

// TestSearchRequest creates a test search request
func TestSearchRequest(projectID string) map[string]any {
	return map[string]any{
		"project_id": projectID,
		"query":      "dashboard creation",
		"limit":      10,
	}
}
