package media_test

import (
	"context"
	"log/slog"
	"testing"

	"cgap/internal/media"
)

func TestVideoTranscriber_GetSupportedFormats(t *testing.T) {
	logger := slog.Default()

	transcriber := media.NewVideoTranscriber(logger)

	formats := transcriber.GetSupportedFormats()
	if len(formats) == 0 {
		t.Error("Expected at least one supported format")
	}

	// Check for common video formats
	expectedFormats := map[string]bool{
		"mp4": false, "avi": false, "mov": false,
	}

	for _, format := range formats {
		expectedFormats[format] = true
	}

	// Verify we have at least some expected formats
	if !expectedFormats["mp4"] && !expectedFormats["avi"] && !expectedFormats["mov"] {
		t.Error("Expected common video formats (mp4, avi, or mov)")
	}
}

func TestVideoTranscriber_EstimateProcessingTime(t *testing.T) {
	logger := slog.Default()
	transcriber := media.NewVideoTranscriber(logger)

	testCases := []struct {
		durationSeconds int
		expectedMin     int
		expectedMax     int
	}{
		{durationSeconds: 60, expectedMin: 6, expectedMax: 20},
		{durationSeconds: 300, expectedMin: 30, expectedMax: 100},
		{durationSeconds: 0, expectedMin: 0, expectedMax: 10},
	}

	for _, tc := range testCases {
		estimate := transcriber.EstimateProcessingTime(tc.durationSeconds)
		if estimate < tc.expectedMin || estimate > tc.expectedMax {
			t.Errorf("For duration %d seconds, expected estimate between %d-%d, got %d",
				tc.durationSeconds, tc.expectedMin, tc.expectedMax, estimate)
		}
	}
}

func TestVideoTranscriber_TranscribeFromURL(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	transcriber := media.NewVideoTranscriber(logger)

	testCases := []struct {
		name        string
		videoURL    string
		expectError bool
		validate    func(*media.TranscriptResult) error
	}{
		{
			name:        "valid video URL",
			videoURL:    "https://example.com/video.mp4",
			expectError: false,
			validate: func(r *media.TranscriptResult) error {
				if r.Transcript == "" {
					t.Error("Expected non-empty transcript")
				}
				if r.Language == "" {
					t.Error("Expected language to be detected")
				}
				if len(r.Segments) == 0 {
					t.Error("Expected at least one segment")
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := transcriber.TranscribeFromURL(ctx, tc.videoURL)
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

func TestNewVideoTranscriber(t *testing.T) {
	logger := slog.Default()

	transcriber := media.NewVideoTranscriber(logger)
	if transcriber == nil {
		t.Error("Expected non-nil transcriber")
	}
}

func TestNewVideoTranscriber_NilLogger(t *testing.T) {
	// Should not error with nil logger
	transcriber := media.NewVideoTranscriber(nil)
	if transcriber == nil {
		t.Error("Expected non-nil transcriber with nil logger")
	}
}
