package media_test

import (
	"cgap/internal/media"
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestMediaOrchestrator_ProcessImage(t *testing.T) {
	t.Skip("Skipping image test - requires valid image URL or local file path")

	// Note: This test would require either:
	// 1. A valid image URL that's accessible
	// 2. Google Cloud Vision API credentials
	// 3. Mock HTTP server for testing
	// For now, we rely on OCR handler's own unit tests
}

func TestMediaOrchestrator_ProcessYouTube(t *testing.T) {
	orchestrator, err := media.NewMediaOrchestrator(slog.Default())
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orchestrator.Close()

	mediaItem := &media.MediaItem{
		ID:        "test_yt_1",
		ProjectID: "proj_123",
		SourceID:  "src_456",
		Type:      "youtube",
		URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	ctx := context.Background()
	result, err := orchestrator.ProcessMediaItem(ctx, mediaItem)
	if err != nil {
		t.Fatalf("ProcessMediaItem failed: %v", err)
	}

	// Verify result structure
	if result.MediaItemID != mediaItem.ID {
		t.Errorf("Expected MediaItemID=%s, got %s", mediaItem.ID, result.MediaItemID)
	}
	if result.Text == "" {
		t.Error("Expected non-empty transcript")
	}
	if result.ContentType != "transcript" {
		t.Errorf("Expected ContentType=transcript, got %s", result.ContentType)
	}

	// Verify metadata contains YouTube-specific fields
	if result.Metadata == nil {
		t.Fatal("Expected metadata to be present")
	}
	if _, ok := result.Metadata["youtube"]; !ok {
		t.Error("Expected 'youtube' field in metadata")
	}
	if _, ok := result.Metadata["video_id"]; !ok {
		t.Error("Expected 'video_id' field in metadata")
	}
}

func TestMediaOrchestrator_ProcessVideo(t *testing.T) {
	orchestrator, err := media.NewMediaOrchestrator(slog.Default())
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orchestrator.Close()

	mediaItem := &media.MediaItem{
		ID:        "test_vid_1",
		ProjectID: "proj_123",
		SourceID:  "src_456",
		Type:      "video",
		URL:       "https://example.com/test.mp4",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	ctx := context.Background()
	result, err := orchestrator.ProcessMediaItem(ctx, mediaItem)
	if err != nil {
		t.Fatalf("ProcessMediaItem failed: %v", err)
	}

	// Verify result structure
	if result.MediaItemID != mediaItem.ID {
		t.Errorf("Expected MediaItemID=%s, got %s", mediaItem.ID, result.MediaItemID)
	}
	if result.Text == "" {
		t.Error("Expected non-empty transcript")
	}
	if result.ContentType != "transcript" {
		t.Errorf("Expected ContentType=transcript, got %s", result.ContentType)
	}

	// Verify metadata contains video-specific fields
	if result.Metadata == nil {
		t.Fatal("Expected metadata to be present")
	}
	if _, ok := result.Metadata["video"]; !ok {
		t.Error("Expected 'video' field in metadata")
	}
}

func TestMediaOrchestrator_DetectMediaType(t *testing.T) {
	orchestrator, err := media.NewMediaOrchestrator(slog.Default())
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orchestrator.Close()

	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/image.jpg", "image"},
		{"https://example.com/image.png", "image"},
		{"https://www.youtube.com/watch?v=abc123", "youtube"},
		{"https://youtu.be/abc123", "youtube"},
		{"https://example.com/video.mp4", "video"},
		{"https://example.com/video.mov", "video"},
		{"https://example.com/video.avi", "video"},
		{"https://example.com/unknown.xyz", "image"}, // defaults to image
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := orchestrator.DetectMediaType(tt.url)
			if result != tt.expected {
				t.Errorf("DetectMediaType(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestMediaOrchestrator_GetSupportedTypes(t *testing.T) {
	orchestrator, err := media.NewMediaOrchestrator(slog.Default())
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orchestrator.Close()

	types := orchestrator.GetSupportedTypes()

	expectedTypes := map[string]bool{
		"image":   true,
		"youtube": true,
		"video":   true,
	}

	for _, typ := range types {
		if !expectedTypes[typ] {
			t.Errorf("Unexpected type in supported list: %s", typ)
		}
		delete(expectedTypes, typ)
	}

	if len(expectedTypes) > 0 {
		t.Errorf("Missing expected types: %v", expectedTypes)
	}
}

func TestMediaOrchestrator_UnsupportedType(t *testing.T) {
	orchestrator, err := media.NewMediaOrchestrator(slog.Default())
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orchestrator.Close()

	mediaItem := &media.MediaItem{
		ID:        "test_unsupported",
		ProjectID: "proj_123",
		SourceID:  "src_456",
		Type:      "unsupported_type",
		URL:       "https://example.com/test.xyz",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	ctx := context.Background()
	_, err = orchestrator.ProcessMediaItem(ctx, mediaItem)
	if err == nil {
		t.Error("Expected error for unsupported media type, got nil")
	}
	if err != nil && err.Error() != "unsupported media type: unsupported_type" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}
