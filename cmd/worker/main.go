package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"cgap/internal/postgres"
	"cgap/internal/queue"
)

func main() {
	log.Println("cgap worker starting...")

	// Load configuration from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	// Initialize PostgreSQL storage with pgx
	store, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize postgres store: %v", err)
	}
	defer store.Close()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	defer redisClient.Close()

	// Initialize Redis queue consumer
	consumer := queue.NewConsumer(redisClient)

	// Start job processor
	go func() {
		log.Println("Starting job consumer...")
		ctx := context.Background()
		for {
			// Get next task
			task, err := consumer.Process(ctx)
			if err != nil {
				log.Printf("Consumer error: %v", err)
				continue
			}

			// Nil task means timeout - no task available
			if task == nil {
				continue
			}

			// Route task to appropriate handler based on task.Type
			log.Printf("Processing task: type=%s, id=%s", task.Type, task.ID)

			// TODO: Route task to appropriate handler based on task.Type
			// - "ingest": handle document ingestion (crawl, chunk, embed, index)
			// - "gap_cluster": handle gap detection and clustering
			// - etc.
		}
	}()

	log.Println("Worker ready")

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
	log.Println("Worker stopped")
}
