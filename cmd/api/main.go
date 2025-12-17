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
	"cgap/internal/embedding"
	"cgap/internal/llm"
	"cgap/internal/meilisearch"
	"cgap/internal/postgres"
	"cgap/internal/queue"
	"cgap/internal/search"
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
	geminiKey := os.Getenv("GEMINI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	grokKey := os.Getenv("XAI_API_KEY")

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	// Initialize PostgreSQL storage with pgx
	store, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize postgres store: %v", err)
	}
	defer store.Close()

	// Initialize Meilisearch client (full-text provider)
	meiliClient := meilisearch.New(meiliURL, meiliKey)

	// Initialize LLM client with provider selection
	llmProvider := os.Getenv("LLM_PROVIDER")
	if llmProvider == "" {
		llmProvider = "openai"
	}

	llmModel := os.Getenv("LLM_MODEL")

	llmAPIKey := openaiKey
	switch llmProvider {
	case "openai":
		if llmAPIKey == "" {
			log.Fatal("OPENAI_API_KEY environment variable not set")
		}
	case "google":
		llmAPIKey = geminiKey
		if llmAPIKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable not set")
		}
	case "anthropic":
		llmAPIKey = anthropicKey
		if llmAPIKey == "" {
			log.Fatal("ANTHROPIC_API_KEY environment variable not set")
		}
	case "grok":
		llmAPIKey = grokKey
		if llmAPIKey == "" {
			log.Fatal("XAI_API_KEY environment variable not set")
		}
	case "mock":
		llmAPIKey = ""
	default:
		log.Fatalf("Unknown LLM_PROVIDER '%s'", llmProvider)
	}

	llmClient, err := llm.New(llm.ProviderConfig{
		Provider: llmProvider,
		APIKey:   llmAPIKey,
		Model:    llmModel,
	})
	if err != nil {
		log.Fatalf("Failed to initialize LLM client: %v", err)
	}

	// Initialize Search provider by strategy
	searchStrategy := os.Getenv("SEARCH_PROVIDER")
	if searchStrategy == "" {
		searchStrategy = "hybrid"
	}

	var searchClient service.Search

	// Initialize embedder for semantic search (used by pgvector provider)
	embProvider := os.Getenv("EMBEDDING_PROVIDER")
	if embProvider == "" {
		embProvider = "openai"
	}
	var embedder embedding.Embedder
	switch embProvider {
	case "openai":
		if openaiKey == "" {
			log.Fatal("OPENAI_API_KEY environment variable not set for embeddings")
		}
		embedder = embedding.NewOpenAIEmbedder(openaiKey, os.Getenv("EMBEDDING_MODEL"))
	case "google":
		if geminiKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable not set for embeddings")
		}
		embedder = embedding.NewGoogleEmbedder(geminiKey, os.Getenv("EMBEDDING_MODEL"))
	case "http":
		embedder = embedding.NewHTTPEmbedder(os.Getenv("EMBEDDING_ENDPOINT"), os.Getenv("EMBEDDING_MODEL"), os.Getenv("EMBEDDING_API_KEY"), os.Getenv("EMBEDDING_AUTH_HEADER"))
	case "mock":
		embedder = embedding.NewMockEmbedder(1536)
	default:
		log.Printf("Unknown EMBEDDING_PROVIDER '%s', defaulting to openai", embProvider)
		if openaiKey == "" {
			log.Fatal("OPENAI_API_KEY environment variable not set for embeddings")
		}
		embedder = embedding.NewOpenAIEmbedder(openaiKey, os.Getenv("EMBEDDING_MODEL"))
	}

	switch searchStrategy {
	case "pgvector":
		searchClient = search.NewPGVector(store, embedder)
	case "meilisearch":
		searchClient = meiliClient
	case "hybrid":
		searchClient = search.NewHybrid(search.NewPGVector(store, embedder), meiliClient)
	default:
		log.Printf("Unknown SEARCH_PROVIDER '%s', defaulting to hybrid", searchStrategy)
		searchClient = search.NewHybrid(search.NewPGVector(store, embedder), meiliClient)
	}

	// Initialize Redis client (URL or host:port)
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		redisOpts = &redis.Options{Addr: redisURL}
	}
	redisClient := redis.NewClient(redisOpts)
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

	// Register handlers with injected services and health dependencies
	api.RegisterRoutesWithServices(app, &api.Services{
		Chat:      chatService,
		Search:    searchService,
		Deflect:   deflectService,
		Analytics: analyticsService,
		Gaps:      gapsService,
		Queue:     queue.NewProducer(redisClient),
	}, &api.HealthDeps{
		DB:    store.Pool(),
		Redis: redisClient,
		Meili: meiliClient,
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
