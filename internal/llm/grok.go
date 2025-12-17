package llm

import (
	"context"
	"fmt"

	"cgap/internal/service"
)

// GrokProvider implements Provider for xAI Grok API
type GrokProvider struct {
	apiKey string
	model  string
}

// NewGrokProvider creates a new Grok provider
func NewGrokProvider(apiKey, model string) (*GrokProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Grok API key is required")
	}

	if model == "" {
		model = "grok-3"
	}

	return &GrokProvider{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// Chat sends messages and returns a single response
func (p *GrokProvider) Chat(ctx context.Context, messages []service.Message) (string, error) {
	// TODO: Implement xAI Grok API integration
	// Grok API follows OpenAI-compatible format
	// POST https://api.x.ai/v1/chat/completions
	return "", fmt.Errorf("grok provider not yet implemented")
}

// Stream sends messages and streams responses token by token
func (p *GrokProvider) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	// TODO: Implement Grok streaming
	return nil, fmt.Errorf("grok streaming not yet implemented")
}

// Name returns the provider name
func (p *GrokProvider) Name() string {
	return "grok"
}
