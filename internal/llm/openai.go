package llm

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"

	"cgap/internal/service"
)

// OpenAIProvider implements Provider for OpenAI API
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if model == "" {
		model = openai.GPT4o
	}

	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
		model:  model,
	}, nil
}

// Chat sends messages and returns a single response
func (p *OpenAIProvider) Chat(ctx context.Context, messages []service.Message) (string, error) {
	msgs := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		msgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: msgs,
	})
	if err != nil {
		return "", fmt.Errorf("openai chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Stream sends messages and streams responses token by token
func (p *OpenAIProvider) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	msgs := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		msgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: msgs,
	})
	if err != nil {
		return nil, fmt.Errorf("openai streaming failed: %w", err)
	}

	ch := make(chan string, 1)

	go func() {
		defer close(ch)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return
			}

			if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
				ch <- response.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}
