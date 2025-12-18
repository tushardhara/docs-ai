package llm_test

import (
	"context"
	"testing"

	"cgap/internal/llm"
	"cgap/internal/service"
)

func TestNew_OpenAI(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestNew_Google(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "google",
		APIKey:   "test-key",
		Model:    "gemini-pro",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Google client: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestNew_Anthropic(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "anthropic",
		APIKey:   "test-key",
		Model:    "claude-3",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Anthropic client: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestNew_Grok(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "grok",
		APIKey:   "test-key",
		Model:    "grok-beta",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Grok client: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestNew_Mock(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "mock",
		Model:    "mock-model",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Mock client: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "unknown",
		APIKey:   "test-key",
		Model:    "test",
	}

	_, err := llm.New(cfg)
	if err == nil {
		t.Error("Expected error for unknown provider")
	}
}

func TestClient_Chat(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "mock",
		Model:    "mock-model",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	messages := []service.Message{
		{Role: "user", Content: "Hello"},
	}

	response, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestClient_Stream(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "mock",
		Model:    "mock-model",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	messages := []service.Message{
		{Role: "user", Content: "Hello"},
	}

	ch, err := client.Stream(ctx, messages)
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
		t.Error("Expected at least one token from stream")
	}
}

func TestClient_Name(t *testing.T) {
	cfg := llm.ProviderConfig{
		Provider: "mock",
		Model:    "test-model",
	}

	client, err := llm.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	name := client.Name()
	if name == "" {
		t.Error("Expected non-empty provider name")
	}
}
