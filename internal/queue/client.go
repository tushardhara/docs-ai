package queue

import (
	"context"
	"fmt"
)

// Task represents a job to be processed.
type Task struct {
	Type    string `json:"type"` // "ingest", "gap_cluster", etc.
	Payload any    `json:"payload"`
}

// Producer enqueues tasks.
type Producer struct {
	redisURL string
}

func NewProducer(redisURL string) *Producer {
	return &Producer{
		redisURL: redisURL,
	}
}

// Enqueue adds a task to the queue.
func (p *Producer) Enqueue(ctx context.Context, task Task) error {
	// TODO: Implement Redis queue enqueue
	// Connect to Redis, push task JSON to queue list (e.g., "cgap:tasks")
	// Return error if enqueue fails
	_ = ctx
	_ = task
	return fmt.Errorf("not implemented")
}

// Consumer dequeues and processes tasks.
type Consumer struct {
	redisURL string
}

func NewConsumer(redisURL string) *Consumer {
	return &Consumer{
		redisURL: redisURL,
	}
}

// Process starts consuming tasks and calls handler for each.
func (c *Consumer) Process(ctx context.Context, handler func(Task) error) error {
	// TODO: Implement Redis queue consumer
	// Connect to Redis, BLPOP from queue list with timeout
	// Unmarshal task JSON, call handler, return error if handler fails
	// Loop until context is cancelled
	_ = ctx
	_ = handler
	return fmt.Errorf("not implemented")
}
