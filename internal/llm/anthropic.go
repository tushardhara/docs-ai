package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cgap/internal/service"
)

// AnthropicProvider implements Provider for Anthropic Claude API (HTTP).
type AnthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string) (*AnthropicProvider, error) {
	// Prefer provider-specific env if set
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		apiKey = v
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicTextBlock `json:"content"`
}

// Chat sends messages and returns a single response
func (p *AnthropicProvider) Chat(ctx context.Context, messages []service.Message) (string, error) {
	msgs := make([]anthropicMessage, len(messages))
	for i, m := range messages {
		msgs[i] = anthropicMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := anthropicRequest{Model: p.model, Messages: msgs, MaxTokens: 1024}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("anthropic chat: status %d", resp.StatusCode)
	}

	var out anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, c := range out.Content {
		if c.Type == "text" {
			b.WriteString(c.Text)
		}
	}
	return b.String(), nil
}

// Stream sends messages and streams responses token by token
func (p *AnthropicProvider) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	msgs := make([]anthropicMessage, len(messages))
	for i, m := range messages {
		msgs[i] = anthropicMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := anthropicRequest{Model: p.model, Messages: msgs, MaxTokens: 1024, Stream: true}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 4)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" || payload == "" {
				continue
			}
			// Anthropic streaming events: look for text deltas
			var m map[string]any
			if err := json.Unmarshal([]byte(payload), &m); err != nil {
				continue
			}
			if m["type"] == "content_block_delta" {
				if delta, ok := m["delta"].(map[string]any); ok {
					if delta["type"] == "text_delta" {
						if txt, ok := delta["text"].(string); ok && txt != "" {
							ch <- txt
						}
					}
				}
			}
		}
	}()

	return ch, nil
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}
