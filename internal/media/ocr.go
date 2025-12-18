package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// GoogleVisionOCR implements OCRHandler using Google Cloud Vision API
type GoogleVisionOCR struct {
	apiKey string
	logger *slog.Logger
}

// NewGoogleVisionOCR creates a new Google Vision OCR handler
func NewGoogleVisionOCR(_ context.Context, logger *slog.Logger) (*GoogleVisionOCR, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_CLOUD_VISION_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}

	if apiKey == "" {
		logger.Warn("GOOGLE_CLOUD_VISION_API_KEY or GOOGLE_API_KEY not set, OCR will use mock data for testing")
	}

	return &GoogleVisionOCR{
		apiKey: apiKey,
		logger: logger,
	}, nil
}

// ExtractFromURL extracts text from an image URL using Google Vision
func (g *GoogleVisionOCR) ExtractFromURL(ctx context.Context, imageURL string) (*OCRResult, error) {
	g.logger.Debug("Extracting text from URL", "url", imageURL)

	// Validate URL
	u, err := url.Parse(imageURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("%w: unsupported scheme %s", ErrInvalidURL, u.Scheme)
	}

	// If we have API key, use real Google Vision API
	if g.apiKey != "" {
		return g.extractWithGoogleVisionAPI(ctx, imageURL)
	}

	// Otherwise, download and process locally
	return g.extractFromURLLocal(ctx, imageURL)
}

// extractFromURLLocal downloads image and processes locally
func (g *GoogleVisionOCR) extractFromURLLocal(ctx context.Context, imageURL string) (*OCRResult, error) {
	// Download the image (validated URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrExtractionFailed, resp.StatusCode)
	}

	// Save to temp file
	tempFile, err := os.CreateTemp("", "ocr_*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		_ = tempFile.Close()
		return nil, fmt.Errorf("failed to write image: %w", err)
	}
	_ = tempFile.Close()

	// Extract from file
	return g.ExtractFromFile(ctx, tempFile.Name())
}

// ExtractFromFile extracts text from a local image file
func (g *GoogleVisionOCR) ExtractFromFile(ctx context.Context, filePath string) (*OCRResult, error) {
	g.logger.Debug("Extracting text from file", "path", filePath)

	// Sanitize and check file path
	cleanPath := filepath.Clean(filePath)
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// If we have API key, use Google Vision API
	if g.apiKey != "" {
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return g.extractWithGoogleVisionAPIBytes(ctx, data)
	}

	// For now, return mock data for local processing (placeholder for future OCR engine)
	g.logger.Debug("Processing file locally with mock OCR", "path", filePath)

	return &OCRResult{
		Text:            fmt.Sprintf("[Mock OCR] Text extracted from: %s\n\nNote: For production use, set GOOGLE_CLOUD_VISION_API_KEY environment variable.", filepath.Base(filePath)),
		ConfidenceScore: 0.75,
		Language:        "en",
		BoundingBoxes:   []TextBoundingBox{},
		RawResponse:     nil,
	}, nil
}

// extractWithGoogleVisionAPI calls Google Cloud Vision API with image URL
func (g *GoogleVisionOCR) extractWithGoogleVisionAPI(_ context.Context, imageURL string) (*OCRResult, error) {
	g.logger.Debug("Google Vision placeholder: using image URL", "url", imageURL)
	// Build request to Google Vision API
	// This would use the actual Google Vision client library
	// For now, placeholder for future implementation
	g.logger.Warn("Google Vision API integration not yet implemented")

	// Return mock response
	return &OCRResult{
		Text:            "[Google Vision API] Text extraction not yet implemented. Please configure and implement.",
		ConfidenceScore: 0.0,
		Language:        "unknown",
		BoundingBoxes:   []TextBoundingBox{},
		RawResponse:     nil,
	}, nil
}

// extractWithGoogleVisionAPIBytes calls Google Cloud Vision API with image bytes
func (g *GoogleVisionOCR) extractWithGoogleVisionAPIBytes(_ context.Context, imageData []byte) (*OCRResult, error) {
	g.logger.Debug("Google Vision placeholder: using image bytes", "size", len(imageData))
	// Build request to Google Vision API
	// This would use the actual Google Vision client library
	// For now, placeholder for future implementation
	g.logger.Warn("Google Vision API integration not yet implemented")

	// Return mock response
	return &OCRResult{
		Text:            "[Google Vision API] Text extraction not yet implemented. Please configure and implement.",
		ConfidenceScore: 0.0,
		Language:        "unknown",
		BoundingBoxes:   []TextBoundingBox{},
		RawResponse:     nil,
	}, nil
}

// Close closes any resources
func (g *GoogleVisionOCR) Close() error {
	return nil
}
