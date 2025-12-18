package llm

import (
	"context"
	"fmt"

	"cgap/internal/service"
)

// GoogleProvider implements Provider for Google Gemini API
type GoogleProvider struct {
	apiKey string
	model  string
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(apiKey, model string) (*GoogleProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("google API key is required")
	}

	if model == "" {
		model = "gemini-2.0-flash"
	}

	return &GoogleProvider{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// Chat sends messages and returns a single response
func (p *GoogleProvider) Chat(ctx context.Context, messages []service.Message) (string, error) {
	// TODO: Implement Google Gemini API integration
	// Use: github.com/google/generative-ai-go
	// POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
	return "", fmt.Errorf("google provider not yet implemented")
}

// Stream sends messages and streams responses token by token
func (p *GoogleProvider) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	// TODO: Implement Google Gemini streaming
	return nil, fmt.Errorf("google streaming not yet implemented")
}

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}
