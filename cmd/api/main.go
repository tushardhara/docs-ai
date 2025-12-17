package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"

	"cgap/api"
	"cgap/internal/postgres"
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

	// Initialize PostgreSQL storage with pgx
	store, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize postgres store: %v", err)
	}
	defer store.Close()

	// TODO: Initialize Meilisearch client
	// TODO: Initialize LLM client
	// TODO: Wire up service implementations with store and clients

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "cgap",
	})

	// Register handlers
	api.RegisterRoutes(app)

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
