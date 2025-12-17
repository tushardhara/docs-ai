package llm

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"

	"cgap/internal/service"
)

// Client wraps OpenAI API for LLM operations.
type Client struct {
	openai *openai.Client
	model  string
}

func New(apiKey, model string) *Client {
	if model == "" {
		model = openai.GPT4o
	}
	return &Client{
		openai: openai.NewClient(apiKey),
		model:  model,
	}
}

// Chat sends messages and returns a single response.
func (c *Client) Chat(ctx context.Context, messages []service.Message) (string, error) {
	// Convert service.Message to openai.ChatCompletionMessage
	msgs := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		msgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	resp, err := c.openai.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: msgs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Stream sends messages and streams responses.
func (c *Client) Stream(ctx context.Context, messages []service.Message) (<-chan string, error) {
	// Convert service.Message to openai.ChatCompletionMessage
	msgs := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		msgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	stream, err := c.openai.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: msgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming chat completion: %w", err)
	}

	// Create channel to stream tokens
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
				// Error in stream - close and return
				return
			}

			if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
				ch <- response.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}
