package main

import (
	"context"
	"fmt"
	"log/slog"
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
	// Print custom startup banner first
	printCGAPBanner("8080")

	slog.Info("cgap API starting")

	// Load configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL environment variable not set")
		os.Exit(1)
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
		slog.Error("Failed to initialize postgres store", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize Meilisearch client (full-text provider)
	meiliClient := meilisearch.New(meiliURL, meiliKey)

	// Initialize LLM client with provider selection
	llmProvider := os.Getenv("LLM_PROVIDER")
	if llmProvider == "" {
		llmProvider = llm.ProviderOpenAI
	}

	llmModel := os.Getenv("LLM_MODEL")

	llmAPIKey := openaiKey
	switch llmProvider {
	case llm.ProviderOpenAI:
		if llmAPIKey == "" {
			slog.Error("OPENAI_API_KEY environment variable not set")
			os.Exit(1)
		}
	case llm.ProviderGoogle:
		llmAPIKey = geminiKey
		if llmAPIKey == "" {
			slog.Error("GEMINI_API_KEY environment variable not set")
			os.Exit(1)
		}
	case llm.ProviderAnthropic:
		llmAPIKey = anthropicKey
		if llmAPIKey == "" {
			slog.Error("ANTHROPIC_API_KEY environment variable not set")
			os.Exit(1)
		}
	case llm.ProviderGrok:
		llmAPIKey = grokKey
		if llmAPIKey == "" {
			slog.Error("XAI_API_KEY environment variable not set")
			os.Exit(1)
		}
	case llm.ProviderMock:
		llmAPIKey = ""
	default:
		slog.Error("Unknown LLM_PROVIDER", "provider", llmProvider)
		os.Exit(1)
	}

	llmClient, err := llm.New(llm.ProviderConfig{
		Provider: llmProvider,
		APIKey:   llmAPIKey,
		Model:    llmModel,
	})
	if err != nil {
		slog.Error("Failed to initialize LLM client", "error", err)
		os.Exit(1)
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
			slog.Error("OPENAI_API_KEY environment variable not set for embeddings")
			os.Exit(1)
		}
		embedder = embedding.NewOpenAIEmbedder(openaiKey, os.Getenv("EMBEDDING_MODEL"))
	case "google":
		if geminiKey == "" {
			slog.Error("GEMINI_API_KEY environment variable not set for embeddings")
			os.Exit(1)
		}
		embedder = embedding.NewGoogleEmbedder(geminiKey, os.Getenv("EMBEDDING_MODEL"))
	case "http":
		embedder = embedding.NewHTTPEmbedder(os.Getenv("EMBEDDING_ENDPOINT"), os.Getenv("EMBEDDING_MODEL"), os.Getenv("EMBEDDING_API_KEY"), os.Getenv("EMBEDDING_AUTH_HEADER"))
	case "mock":
		embedder = embedding.NewMockEmbedder(1536)
	default:
		slog.Warn("Unknown EMBEDDING_PROVIDER, defaulting to openai", "provider", embProvider)
		if openaiKey == "" {
			slog.Error("OPENAI_API_KEY environment variable not set for embeddings")
			os.Exit(1)
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
		slog.Warn("Unknown SEARCH_PROVIDER, defaulting to hybrid", "provider", searchStrategy)
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
		DB:        store.Pool(),
	}, &api.HealthDeps{
		DB:    store.Pool(),
		Redis: redisClient,
		Meili: meiliClient,
	})

	// Start server
	go func() {
		if err := app.Listen(":"+port, fiber.ListenConfig{
			DisableStartupMessage: true,
		}); err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down API server...")

	if err := app.ShutdownWithContext(context.Background()); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	slog.Info("API server stopped")
}

// printCGAPBanner prints the cgap startup banner with colors.
func printCGAPBanner(port string) {
	const (
		colorCyan  = "\033[36m"
		colorReset = "\033[0m"
		colorBold  = "\033[1m"
	)
	fmt.Printf("%s%s", colorCyan, colorBold)
	fmt.Println(`
    ░▒▓██████▓▒░ ░▒▓██████▓▒░ ░▒▓██████▓▒░░▒▓███████▓▒░  
   ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
   ░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
   ░▒▓█▓▒░      ░▒▓█▓▒▒▓███▓▒░▒▓████████▓▒░▒▓███████▓▒░  
   ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░        
   ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░        
    ░▒▓██████▓▒░ ░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░        
                                                `)
	fmt.Printf("%s%s", colorReset, colorCyan)
	fmt.Printf("  Documentation AI Assistant | v0.1.0\n")
	fmt.Printf("  Running on http://localhost:%s\n", port)
	fmt.Printf("%s\n", colorReset)
	_ = os.Stdout.Sync()
}
