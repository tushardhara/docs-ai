package llm

import (
	"context"
	"fmt"

	"cgap/internal/service"
)

// ProviderConfig holds configuration for any LLM provider
type ProviderConfig struct {
	Provider string // "openai", "google", "anthropic", "grok", etc.
	APIKey   string
	Model    string
	Config   map[string]interface{} // Custom provider-specific config
}

// Client is the LLM client abstraction (provider-agnostic)
type Client struct {
	provider Provider
}

// Provider is the interface all LLM implementations must satisfy
type Provider interface {
	Chat(ctx context.Context, messages []service.Message) (string, error)
	Stream(ctx context.Context, messages []service.Message) (<-chan string, error)
	Name() string
}

// New creates a new LLM client with the specified provider
func New(cfg ProviderConfig) (*Client, error) {
	var provider Provider
	var err error

	switch cfg.Provider {
	case "openai":
		provider, err = NewOpenAIProvider(cfg.APIKey, cfg.Model)
	case "google":
		provider, err = NewGoogleProvider(cfg.APIKey, cfg.Model)
	case "anthropic":
		provider, err = NewAnthropicProvider(cfg.APIKey, cfg.Model)
	case "grok":
		provider, err = NewGrokProvider(cfg.APIKey, cfg.Model)
	case "mock":
		provider = NewMockProvider(cfg.Model)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize %s provider: %w", cfg.Provider, err)
	}

	return &Client{provider: provider}, nil
}

// Chat delegates to the underlying provider
func (c *Client) Chat(ctx context.Context, messages []service.Message) (string, error) {
	return c.provider.Chat(ctx, messages)
}

// Stream delegates to the underlying provider
func (c *Client) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	return c.provider.Stream(ctx, messages)
}

// Name returns the provider name
func (c *Client) Name() string {
	return c.provider.Name()
}
