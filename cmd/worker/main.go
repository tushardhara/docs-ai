package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
		redisURL = "redis://localhost:6379"
	}

	// Initialize PostgreSQL storage with pgx
	store, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize postgres store: %v", err)
	}
	defer store.Close()

	// TODO: Initialize Meilisearch client
	// TODO: Initialize LLM client

	// Initialize Redis queue consumer
	consumer := queue.NewConsumer(redisURL)

	// Start job processor
	go func() {
		log.Println("Starting job consumer...")
		err := consumer.Process(context.Background(), func(task queue.Task) error {
			// TODO: Route task to appropriate handler based on task.Type
			// - "ingest": handle document ingestion (crawl, chunk, embed, index)
			// - "gap_cluster": handle gap detection and clustering
			// - etc.
			log.Printf("Processing task: %s", task.Type)
			return nil
		})
		if err != nil {
			log.Printf("Consumer error: %v", err)
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
