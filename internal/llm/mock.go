package llm

import (
	"context"
	"fmt"

	"cgap/internal/service"
)

// MockProvider implements Provider for testing without external API calls
type MockProvider struct {
	model string
}

// NewMockProvider creates a new mock provider
func NewMockProvider(model string) *MockProvider {
	if model == "" {
		model = "mock-model"
	}
	return &MockProvider{model: model}
}

// Chat returns a mock response
func (p *MockProvider) Chat(ctx context.Context, messages []service.Message) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}

	// Return mock response based on the last user message
	lastMsg := messages[len(messages)-1]
	return fmt.Sprintf("Mock response to: %s (model: %s)", lastMsg.Content, p.model), nil
}

// Stream returns a mock token stream
func (p *MockProvider) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	ch := make(chan string)

	go func() {
		defer close(ch)

		tokens := []string{"This ", "is ", "a ", "mock ", "streaming ", "response."}
		for _, token := range tokens {
			select {
			case <-ctx.Done():
				return
			case ch <- token:
			}
		}
	}()

	return ch, nil
}

// Name returns the provider name
func (p *MockProvider) Name() string {
	return "mock"
}
