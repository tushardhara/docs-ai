package api

import (
	"cgap/internal/embedding"
	"context"
	"time"

	"os"
	"regexp"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
)

var services *Services
var healthDeps *HealthDeps

// HealthDeps holds upstream dependencies for health checks.
type HealthDeps struct {
	DB    DBPinger
	Redis RedisPinger
	Meili MeiliChecker
}

// Interfaces for health checks to keep API decoupled from concrete types.
type DBPinger interface {
	Ping(ctx context.Context) error
}

type RedisPinger interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

type MeiliChecker interface {
	Health(ctx context.Context) error
}

// ChatHandler handles POST /v1/chat
func ChatHandler(c fiber.Ctx) error {
	var req ChatRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.ProjectID == "" || req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and query required"})
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	// Call chat service
	resp, err := services.Chat.Chat(context.Background(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// SearchHandler handles POST /v1/search
func SearchHandler(c fiber.Ctx) error {
	var req SearchRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.ProjectID == "" || req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and query required"})
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	// Call search service
	hits, err := services.Search.Search(context.Background(), req.ProjectID, req.Query, req.Limit, req.Filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Ensure empty array instead of null in JSON
	if hits == nil {
		hits = []SearchHit{}
	}

	return c.Status(fiber.StatusOK).JSON(SearchResponse{
		Hits:  hits,
		Total: len(hits),
	})
}

// DeflectSuggestHandler handles POST /v1/deflect/suggest
func DeflectSuggestHandler(c fiber.Ctx) error {
	var req DeflectRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.ProjectID == "" || req.TicketText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and ticket_text required"})
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	// Call deflect service
	_, suggestions, err := services.Deflect.Suggest(context.Background(), req.ProjectID, "", req.TicketText, req.TopK)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(DeflectResponse{
		Suggestions: suggestions,
		Deflected:   len(suggestions) > 0,
	})
}

// DeflectEventHandler handles POST /v1/deflect/event
func DeflectEventHandler(c fiber.Ctx) error {
	var req DeflectEventRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.ProjectID == "" || req.EventType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and event_type required"})
	}

	// Call deflect service to track event
	err := services.Deflect.TrackEvent(context.Background(), req.ProjectID, req.SuggestionID, req.EventType, req.ThreadID, req.Metadata)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "logged"})
}

// IngestHandler handles POST /v1/ingest
func IngestHandler(c fiber.Ctx) error {
	var req IngestRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.ProjectID == "" || len(req.Source) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and source required"})
	}

	// Queue ingest job
	// For now, return placeholder job ID
	return c.Status(fiber.StatusAccepted).JSON(IngestResponse{
		JobID:     "job_" + req.ProjectID,
		Status:    "queued",
		ProjectID: req.ProjectID,
	})
}

// DevSeedHandler handles POST /v1/dev/seed to insert a document, chunk, and embedding
func DevSeedHandler(c fiber.Ctx) error {
	var req SeedRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.ProjectID == "" || req.URI == "" || req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id, uri and text required"})
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DATABASE_URL not set"})
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db connect failed"})
	}
	defer pool.Close()

	// Resolve project slug -> UUID if needed
	pid := req.ProjectID
	if !looksLikeUUID(req.ProjectID) {
		if err := pool.QueryRow(context.Background(), `SELECT id FROM projects WHERE slug = $1`, req.ProjectID).Scan(&pid); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
		}
	}

	// Upsert document
	var docID string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO documents (project_id, uri, title)
		VALUES ($1, $2, COALESCE($3, 'Untitled'))
		ON CONFLICT (project_id, uri) DO UPDATE SET title = EXCLUDED.title
		RETURNING id
	`, pid, req.URI, req.Title).Scan(&docID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "insert document failed"})
	}

	// Insert chunk
	var chunkID string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO chunks (document_id, ord, text)
		VALUES ($1, 0, $2)
		RETURNING id
	`, docID, req.Text).Scan(&chunkID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "insert chunk failed"})
	}

	// Build embedder from env similar to main
	embProvider := os.Getenv("EMBEDDING_PROVIDER")
	if embProvider == "" {
		embProvider = "openai"
	}
	var embVec []float32
	{
		// Create embedder per provider
		switch embProvider {
		case "openai":
			apiKey := os.Getenv("OPENAI_API_KEY")
			if apiKey == "" {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "OPENAI_API_KEY not set"})
			}
			emb := embedding.NewOpenAIEmbedder(apiKey, os.Getenv("EMBEDDING_MODEL"))
			v, err := emb.Embed(context.Background(), req.Text)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "embedding failed"})
			}
			embVec = v
		case "google":
			apiKey := os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "GEMINI_API_KEY not set"})
			}
			emb := embedding.NewGoogleEmbedder(apiKey, os.Getenv("EMBEDDING_MODEL"))
			v, err := emb.Embed(context.Background(), req.Text)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "embedding failed"})
			}
			embVec = v
		case "http":
			emb := embedding.NewHTTPEmbedder(os.Getenv("EMBEDDING_ENDPOINT"), os.Getenv("EMBEDDING_MODEL"), os.Getenv("EMBEDDING_API_KEY"), os.Getenv("EMBEDDING_AUTH_HEADER"))
			v, err := emb.Embed(context.Background(), req.Text)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "embedding failed"})
			}
			embVec = v
		case "mock":
			emb := embedding.NewMockEmbedder(768)
			v, err := emb.Embed(context.Background(), req.Text)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "embedding failed"})
			}
			embVec = v
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unknown EMBEDDING_PROVIDER"})
		}
	}

	// Insert embedding using pgvector-go for correct type
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO chunk_embeddings (chunk_id, embedding) VALUES ($1, $2)
		ON CONFLICT (chunk_id) DO UPDATE SET embedding = EXCLUDED.embedding
	`, chunkID, pgvector.NewVector(embVec)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "insert embedding failed"})
	}

	return c.Status(fiber.StatusOK).JSON(SeedResponse{
		ProjectID:  pid,
		DocumentID: docID,
		ChunkID:    chunkID,
		Status:     "seeded",
	})
}

// AnalyticsHandler handles GET /v1/analytics/:project_id
func AnalyticsHandler(c fiber.Ctx) error {
	projectID := c.Params("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id required"})
	}

	// Call analytics service
	summary, err := services.Analytics.Summary(context.Background(), projectID, nil, nil, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(AnalyticsResponse{
		ProjectID: projectID,
		Summary:   summary,
	})
}

// GapsHandler handles GET /v1/gaps/:project_id
func GapsHandler(c fiber.Ctx) error {
	projectID := c.Params("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id required"})
	}

	// Call gaps service
	gaps, err := services.Gaps.List(context.Background(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(GapsResponse{
		Gaps:  gaps,
		Total: len(gaps),
	})
}

// HealthHandler handles GET /health
func HealthHandler(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := fiber.Map{
		"status": "ok",
	}
	unhealthy := false

	if healthDeps != nil && healthDeps.DB != nil {
		if err := healthDeps.DB.Ping(ctx); err != nil {
			status["db"] = err.Error()
			unhealthy = true
		} else {
			status["db"] = "ok"
		}
	}

	if healthDeps != nil && healthDeps.Redis != nil {
		if err := healthDeps.Redis.Ping(ctx).Err(); err != nil {
			status["redis"] = err.Error()
			unhealthy = true
		} else {
			status["redis"] = "ok"
		}
	}

	if healthDeps != nil && healthDeps.Meili != nil {
		if err := healthDeps.Meili.Health(ctx); err != nil {
			status["meilisearch"] = err.Error()
			unhealthy = true
		} else {
			status["meilisearch"] = "ok"
		}
	}

	if unhealthy {
		status["status"] = "degraded"
		return c.Status(fiber.StatusServiceUnavailable).JSON(status)
	}

	return c.Status(fiber.StatusOK).JSON(status)
}

// RegisterRoutes registers all HTTP handlers with Fiber (deprecated).
func RegisterRoutes(app *fiber.App) {
	RegisterRoutesWithServices(app, &Services{}, nil)
}

// RegisterRoutesWithServices registers all HTTP handlers with Fiber and injects services.
func RegisterRoutesWithServices(app *fiber.App, svc *Services, deps *HealthDeps) {
	services = svc
	healthDeps = deps

	// Health
	app.Get("/health", HealthHandler)

	// Chat
	app.Post("/v1/chat", ChatHandler)

	// Search
	app.Post("/v1/search", SearchHandler)

	// Deflect
	app.Post("/v1/deflect/suggest", DeflectSuggestHandler)
	app.Post("/v1/deflect/event", DeflectEventHandler)

	// Ingest
	app.Post("/v1/ingest", IngestHandler)
	// Dev seed
	app.Post("/v1/dev/seed", DevSeedHandler)

	// Analytics
	app.Get("/v1/analytics/:project_id", AnalyticsHandler)

	// Gaps
	app.Get("/v1/gaps/:project_id", GapsHandler)
}

var uuidReHandlers = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func looksLikeUUID(s string) bool { return uuidReHandlers.MatchString(s) }
