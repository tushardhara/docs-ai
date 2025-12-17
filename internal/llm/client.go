package llm

import (
	"context"
	"fmt"

	"cgap/internal/service"
)

// Client wraps LLM API (OpenAI/Anthropic).
type Client struct {
	provider string // "openai" or "anthropic"
	apiKey   string
	model    string
}

func New(provider, apiKey, model string) *Client {
	return &Client{
		provider: provider,
		apiKey:   apiKey,
		model:    model,
	}
}

// Chat sends messages and returns a single response.
func (c *Client) Chat(ctx context.Context, messages []service.Message) (string, error) {
	// TODO: Implement LLM API client (OpenAI or Anthropic)
	// Convert messages to API format, send request, parse response
	// Return assistant response text
	_ = ctx
	_ = messages
	return "", fmt.Errorf("not implemented")
}

// Stream sends messages and streams responses.
func (c *Client) Stream(ctx context.Context, messages []service.Message, onChunk func(string)) error {
	// TODO: Implement streaming LLM API client
	// Send request with stream=true, parse SSE chunks, call onChunk callback
	// Return error if stream fails
	_ = ctx
	_ = messages
	_ = onChunk
	return fmt.Errorf("not implemented")
}
