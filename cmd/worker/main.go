package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"

	"cgap/api"
	"cgap/internal/embedding"
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

	// Build embedder for ingestion
	embedder := buildEmbedder()

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

			switch task.Type {
			case "ingest":
				if err := handleIngest(ctx, store, embedder, task.Payload); err != nil {
					log.Printf("ingest error (task=%s): %v", task.ID, err)
				} else {
					log.Printf("ingest completed (task=%s)", task.ID)
				}
			default:
				// Not handled yet
			}
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

// buildEmbedder constructs an embedder from environment configuration.
func buildEmbedder() embedding.Embedder {
	provider := os.Getenv("EMBEDDING_PROVIDER")
	model := os.Getenv("EMBEDDING_MODEL")
	if provider == "" {
		provider = "google"
	}
	switch provider {
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			log.Println("OPENAI_API_KEY not set; embeddings may fail")
		}
		return embedding.NewOpenAIEmbedder(key, model)
	case "http":
		return embedding.NewHTTPEmbedder(os.Getenv("EMBEDDING_ENDPOINT"), model, os.Getenv("EMBEDDING_API_KEY"), os.Getenv("EMBEDDING_AUTH_HEADER"))
	case "mock":
		return embedding.NewMockEmbedder(768)
	default: // google
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			log.Println("GEMINI_API_KEY not set; embeddings may fail")
		}
		return embedding.NewGoogleEmbedder(key, model)
	}
}

// handleIngest performs a minimal ingestion: fetch content from URL(s),
// create document and one-or-more chunks, embed and store in Postgres.
func handleIngest(ctx context.Context, store *postgres.Store, emb embedding.Embedder, payload any) error {
	// Decode payload into API DTO
	mp, ok := payload.(map[string]any)
	if !ok {
		return nil // ignore unknown payload format
	}

	// Quick, robust decode: only fields we need
	p := api.IngestTaskPayload{}
	if v, ok := mp["project_id"].(string); ok {
		p.ProjectID = v
	}
	if src, ok := mp["source"].(map[string]any); ok {
		p.Source.Type, _ = src["type"].(string)
		p.Source.URL, _ = src["url"].(string)
		if files, ok := src["files"].(map[string]any); ok {
			if urls, ok := files["urls"].([]any); ok {
				for _, u := range urls {
					if s, ok := u.(string); ok {
						if p.Source.Files == nil {
							p.Source.Files = &api.FileSpec{}
						}
						p.Source.Files.URLs = append(p.Source.Files.URLs, s)
					}
				}
			}
			p.Source.Files.Format, _ = files["format"].(string)
		}
	}

	if p.ProjectID == "" {
		return nil
	}

	urls := make([]string, 0, 4)
	if p.Source.URL != "" {
		urls = append(urls, p.Source.URL)
	}
	if p.Source.Files != nil && len(p.Source.Files.URLs) > 0 {
		urls = append(urls, p.Source.Files.URLs...)
	}
	if len(urls) == 0 {
		return nil
	}

	pool := store.Pool()
	httpClient := &http.Client{Timeout: 30 * time.Second}

	for _, u := range urls {
		// Fetch content
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return err
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("fetch failed: %s status=%d", u, resp.StatusCode)
			continue
		}

		text := string(body)
		// naive text cleanup for markdown/plain; leave HTML as-is for now
		if strings.HasSuffix(strings.ToLower(u), ".md") || p.Source.Files != nil && (p.Source.Files.Format == "markdown" || p.Source.Files.Format == "md") {
			// keep as-is
		} else if p.Source.Files != nil && (p.Source.Files.Format == "txt" || p.Source.Files.Format == "text") {
			// keep as-is
		} else {
			// basic fallback: strip tags lightly
			text = stripBasicHTML(text)
		}

		// Upsert document (by project_id + uri)
		var docID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO documents (project_id, uri, title)
			VALUES ($1, $2, COALESCE($3, 'Untitled'))
			ON CONFLICT (project_id, uri) DO UPDATE SET title = COALESCE(EXCLUDED.title, documents.title)
			RETURNING id
		`, p.ProjectID, u, "").Scan(&docID); err != nil {
			return err
		}

		// Very simple chunking: split by blank lines, cap to first N chunks
		parts := splitIntoParagraphs(text)
		if len(parts) == 0 {
			continue
		}
		const maxChunks = 20
		if len(parts) > maxChunks {
			parts = parts[:maxChunks]
		}

		// Insert chunks and embeddings
		for i, t := range parts {
			var chunkID string
			if err := pool.QueryRow(ctx, `
				INSERT INTO chunks (document_id, ord, text)
				VALUES ($1, $2, $3)
				RETURNING id
			`, docID, i, t).Scan(&chunkID); err != nil {
				return err
			}

			vec, err := emb.Embed(ctx, t)
			if err != nil {
				return err
			}
			if _, err := pool.Exec(ctx, `
				INSERT INTO chunk_embeddings (chunk_id, embedding)
				VALUES ($1, $2)
				ON CONFLICT (chunk_id) DO UPDATE SET embedding = EXCLUDED.embedding
			`, chunkID, pgvector.NewVector(vec)); err != nil {
				return err
			}
		}
	}

	return nil
}

func splitIntoParagraphs(s string) []string {
	// split on blank lines and trim
	blocks := strings.Split(s, "\n\n")
	out := make([]string, 0, len(blocks))
	for _, b := range blocks {
		t := strings.TrimSpace(b)
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 && strings.TrimSpace(s) != "" {
		// fallback to single chunk
		return []string{strings.TrimSpace(s)}
	}
	return out
}

func stripBasicHTML(s string) string {
	// very naive: remove <...> tags and collapse whitespace
	b := make([]rune, 0, len(s))
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
			continue
		case '>':
			inTag = false
			continue
		}
		if !inTag {
			b = append(b, r)
		}
	}
	t := strings.ReplaceAll(string(b), "\r", "")
	t = strings.ReplaceAll(t, "\t", " ")
	// collapse multiple newlines
	for strings.Contains(t, "\n\n\n") {
		t = strings.ReplaceAll(t, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(t)
}
