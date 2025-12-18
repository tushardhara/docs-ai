package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Task represents a job to be processed.
type Task struct {
	Type    string `json:"type"` // "ingest", "gap_cluster", etc.
	Payload any    `json:"payload"`
	ID      string `json:"id,omitempty"`
	Retries int    `json:"retries,omitempty"`
}

// Producer enqueues tasks.
type Producer struct {
	client *redis.Client
	key    string
}

// NewProducer creates a new task queue producer.
func NewProducer(redisClient *redis.Client) *Producer {
	return &Producer{
		client: redisClient,
		key:    "cgap:tasks",
	}
}

// Enqueue adds a task to the queue.
func (p *Producer) Enqueue(ctx context.Context, task Task) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := p.client.LPush(ctx, p.key, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// Consumer dequeues and processes tasks.
type Consumer struct {
	client  *redis.Client
	key     string
	timeout time.Duration
}

// NewConsumer creates a new task queue consumer.
func NewConsumer(redisClient *redis.Client) *Consumer {
	return &Consumer{
		client:  redisClient,
		key:     "cgap:tasks",
		timeout: 30 * time.Second,
	}
}

// Process retrieves and processes the next task from the queue.
// Returns nil if timeout occurs without a task.
func (c *Consumer) Process(ctx context.Context) (*Task, error) {
	// Use BLPOP with timeout to get next task
	result, err := c.client.BLPop(ctx, c.timeout, c.key).Result()
	if err != nil {
		if err == redis.Nil {
			// Timeout - no task available
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pop task: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected BLPOP result format")
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}
