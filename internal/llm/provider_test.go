package llm_test

import (
	"context"
	"testing"

	"cgap/internal/llm"
	"cgap/internal/service"
)

func TestOpenAIProvider_MissingAPIKey(t *testing.T) {
	_, err := llm.NewOpenAIProvider("", "gpt-4")
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestOpenAIProvider_DefaultModel(t *testing.T) {
	provider, err := llm.NewOpenAIProvider("test-key", "")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider == nil {
		t.Error("Expected non-nil provider with default model")
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	provider, err := llm.NewOpenAIProvider("test-key", "gpt-4")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	name := provider.Name()
	if name != "openai" {
		t.Errorf("Expected name 'openai', got '%s'", name)
	}
}

func TestAnthropicProvider_MissingAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	_, err := llm.NewAnthropicProvider("", "claude-3")
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestAnthropicProvider_DefaultModel(t *testing.T) {
	provider, err := llm.NewAnthropicProvider("test-key", "")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider == nil {
		t.Error("Expected non-nil provider with default model")
	}
}

func TestAnthropicProvider_Name(t *testing.T) {
	provider, err := llm.NewAnthropicProvider("test-key", "claude-3")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	name := provider.Name()
	if name != "anthropic" {
		t.Errorf("Expected name 'anthropic', got '%s'", name)
	}
}

func TestGoogleProvider_MissingAPIKey(t *testing.T) {
	_, err := llm.NewGoogleProvider("", "gemini-pro")
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestGoogleProvider_DefaultModel(t *testing.T) {
	provider, err := llm.NewGoogleProvider("test-key", "")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider == nil {
		t.Error("Expected non-nil provider with default model")
	}
}

func TestGoogleProvider_Name(t *testing.T) {
	provider, err := llm.NewGoogleProvider("test-key", "gemini-pro")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	name := provider.Name()
	if name != "google" {
		t.Errorf("Expected name 'google', got '%s'", name)
	}
}

func TestGrokProvider_MissingAPIKey(t *testing.T) {
	_, err := llm.NewGrokProvider("", "grok-beta")
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestGrokProvider_DefaultModel(t *testing.T) {
	provider, err := llm.NewGrokProvider("test-key", "")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider == nil {
		t.Error("Expected non-nil provider with default model")
	}
}

func TestGrokProvider_Name(t *testing.T) {
	provider, err := llm.NewGrokProvider("test-key", "grok-beta")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	name := provider.Name()
	if name != "grok" {
		t.Errorf("Expected name 'grok', got '%s'", name)
	}
}

func TestMockProvider_Chat(t *testing.T) {
	provider := llm.NewMockProvider("mock-model")

	ctx := context.Background()
	messages := []service.Message{
		{Role: "user", Content: "Test message"},
	}

	response, err := provider.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response from mock provider")
	}
}

func TestMockProvider_Stream(t *testing.T) {
	provider := llm.NewMockProvider("mock-model")

	ctx := context.Background()
	messages := []service.Message{
		{Role: "user", Content: "Test message"},
	}

	ch, err := provider.Stream(ctx, messages)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	tokenCount := 0
	for token := range ch {
		if token != "" {
			tokenCount++
		}
	}

	if tokenCount == 0 {
		t.Error("Expected at least one token from mock provider stream")
	}
}

func TestMockProvider_Name(t *testing.T) {
	provider := llm.NewMockProvider("mock-model")

	name := provider.Name()
	if name != "mock" {
		t.Errorf("Expected name 'mock', got '%s'", name)
	}
}
