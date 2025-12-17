package llm

import (
	"context"
	"fmt"

	"cgap/internal/service"
)

// AnthropicProvider implements Provider for Anthropic Claude API
type AnthropicProvider struct {
	apiKey string
	model  string
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// Chat sends messages and returns a single response
func (p *AnthropicProvider) Chat(ctx context.Context, messages []service.Message) (string, error) {
	// TODO: Implement Anthropic Claude API integration
	// Use: github.com/anthropics/anthropic-sdk-go
	// POST https://api.anthropic.com/v1/messages
	return "", fmt.Errorf("anthropic provider not yet implemented")
}

// Stream sends messages and streams responses token by token
func (p *AnthropicProvider) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	// TODO: Implement Anthropic Claude streaming with Server-Sent Events
	return nil, fmt.Errorf("anthropic streaming not yet implemented")
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}
