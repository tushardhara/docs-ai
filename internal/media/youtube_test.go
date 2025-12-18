package media_test

import (
	"context"
	"log/slog"
	"testing"

	"cgap/internal/media"
)

func TestYouTubeTranscriptFetcher_GetTranscript(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	fetcher := media.NewYouTubeTranscriptFetcher(logger)

	testCases := []struct {
		name        string
		videoID     string
		expectError bool
		validate    func(*media.TranscriptResult) error
	}{
		{
			name:        "valid video ID",
			videoID:     "dQw4w9WgXcQ",
			expectError: false,
			validate: func(r *media.TranscriptResult) error {
				if r.Transcript == "" {
					t.Error("Expected non-empty transcript")
				}
				if r.Language == "" {
					t.Error("Expected language to be set")
				}
				if len(r.Segments) == 0 {
					t.Error("Expected at least one segment")
				}
				return nil
			},
		},
		{
			name:        "empty video ID",
			videoID:     "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fetcher.GetTranscript(ctx, tc.videoID)
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}
			if !tc.expectError && result != nil {
				if err := tc.validate(result); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestYouTubeTranscriptFetcher_ExtractVideoIDFromURL(t *testing.T) {
	logger := slog.Default()
	fetcher := media.NewYouTubeTranscriptFetcher(logger)

	testCases := []struct {
		name        string
		url         string
		expectedID  string
		expectError bool
	}{
		{
			name:       "youtube.com URL",
			url:        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expectedID: "dQw4w9WgXcQ",
		},
		{
			name:       "youtu.be short URL",
			url:        "https://youtu.be/dQw4w9WgXcQ",
			expectedID: "dQw4w9WgXcQ",
		},
		{
			name:       "youtu.be with parameters",
			url:        "https://youtu.be/dQw4w9WgXcQ?t=10",
			expectedID: "dQw4w9WgXcQ",
		},
		{
			name:        "invalid URL",
			url:         "https://example.com/not-youtube",
			expectError: true,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := fetcher.ExtractVideoIDFromURL(tc.url)
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}
			if !tc.expectError && id != tc.expectedID {
				t.Errorf("Expected ID %s, got %s", tc.expectedID, id)
			}
		})
	}
}

func TestNewYouTubeTranscriptFetcher(t *testing.T) {
	logger := slog.Default()

	fetcher := media.NewYouTubeTranscriptFetcher(logger)
	if fetcher == nil {
		t.Error("Expected non-nil fetcher")
	}
}

func TestNewYouTubeTranscriptFetcher_NilLogger(t *testing.T) {
	// Should not error with nil logger
	fetcher := media.NewYouTubeTranscriptFetcher(nil)
	if fetcher == nil {
		t.Error("Expected non-nil fetcher with nil logger")
	}
}
