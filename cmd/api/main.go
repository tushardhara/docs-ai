package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"cgap/api"
	"cgap/internal/llm"
	"cgap/internal/meilisearch"
	"cgap/internal/postgres"
	"cgap/internal/queue"
	"cgap/internal/service"
)

func main() {
	log.Println("cgap API starting...")

	// Load configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	meiliURL := os.Getenv("MEILISEARCH_URL")
	if meiliURL == "" {
		meiliURL = "http://localhost:7700"
	}

	meiliKey := os.Getenv("MEILISEARCH_KEY")
	if meiliKey == "" {
		meiliKey = "test_key"
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
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

	// Initialize Meilisearch client
	searchClient := meilisearch.New(meiliURL, meiliKey)

	// Initialize LLM client (OpenAI)
	llmClient := llm.New(openaiKey, "gpt-4o")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Parse from redisURL if needed
	})
	defer redisClient.Close()

	// Wire up service implementations
	chatService := service.NewChatService(store, llmClient, searchClient)
	searchService := service.NewSearchService(store, searchClient)
	deflectService := service.NewDeflectService(store, searchClient, llmClient)
	analyticsService := service.NewAnalyticsService(store)
	gapsService := service.NewGapsService(store, llmClient)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "cgap",
	})

	// Register handlers with injected services
	api.RegisterRoutesWithServices(app, &api.Services{
		Chat:      chatService,
		Search:    searchService,
		Deflect:   deflectService,
		Analytics: analyticsService,
		Gaps:      gapsService,
		Queue:     queue.NewProducer(redisClient),
	})

	// Start server
	go func() {
		log.Printf("Server listening on :%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down API server...")

	if err := app.ShutdownWithContext(context.Background()); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("API server stopped")
}
