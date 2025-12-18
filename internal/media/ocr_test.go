package media_test

import (
	"context"
	"log/slog"
	"testing"

	"cgap/internal/media"
)

func TestGoogleVisionOCR_ExtractFromURL(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	ocr, err := media.NewGoogleVisionOCR(ctx, logger)
	if err != nil {
		t.Fatalf("Failed to create GoogleVisionOCR: %v", err)
	}

	testCases := []struct {
		name        string
		imageURL    string
		expectError bool
	}{
		{
			name:        "valid image URL",
			imageURL:    "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png",
			expectError: false,
		},
		{
			name:        "invalid URL",
			imageURL:    "not a valid url",
			expectError: true,
		},
		{
			name:        "empty URL",
			imageURL:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ocr.ExtractFromURL(ctx, tc.imageURL)
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}
			if !tc.expectError && result != nil {
				if result.Text == "" {
					t.Error("Expected non-empty text")
				}
				if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
					t.Errorf("Invalid confidence: %f", result.ConfidenceScore)
				}
			}
		})
	}
}

func TestNewGoogleVisionOCR(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	ocr, err := media.NewGoogleVisionOCR(ctx, logger)
	if err != nil {
		t.Fatalf("Failed to create GoogleVisionOCR: %v", err)
	}

	if ocr == nil {
		t.Error("Expected non-nil OCR handler")
	}
}
