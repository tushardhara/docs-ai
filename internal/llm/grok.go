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

// GrokProvider implements Provider for xAI Grok API (OpenAI-compatible HTTP)
type GrokProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGrokProvider creates a new Grok provider
func NewGrokProvider(apiKey, model string) (*GrokProvider, error) {
	// Prefer provider-specific env if set
	if v := os.Getenv("XAI_API_KEY"); v != "" {
		apiKey = v
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Grok API key is required")
	}

	if model == "" {
		model = "grok-3"
	}

	return &GrokProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

type grokChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type grokChatRequest struct {
	Model    string            `json:"model"`
	Messages []grokChatMessage `json:"messages"`
	Stream   bool              `json:"stream,omitempty"`
}

type grokChatChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type grokChatResponse struct {
	Choices []grokChatChoice `json:"choices"`
}

// Chat sends messages and returns a single response
func (p *GrokProvider) Chat(ctx context.Context, messages []service.Message) (string, error) {
	msgs := make([]grokChatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = grokChatMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := grokChatRequest{Model: p.model, Messages: msgs}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.x.ai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("grok chat: status %d", resp.StatusCode)
	}

	var out grokChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("grok chat: empty choices")
	}
	return out.Choices[0].Message.Content, nil
}

// Stream sends messages and streams responses token by token
func (p *GrokProvider) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	msgs := make([]grokChatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = grokChatMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := grokChatRequest{Model: p.model, Messages: msgs, Stream: true}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.x.ai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
			var chunk grokChatResponse
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				ch <- chunk.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}

// Name returns the provider name
func (p *GrokProvider) Name() string {
	return "grok"
}
