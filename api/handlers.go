package api

import (
	"cgap/internal/embedding"
	"cgap/internal/media"
	"cgap/internal/queue"
	"context"
	"fmt"
	"time"

	"os"
	"regexp"

	"strconv"

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

	start := time.Now()

	// Call search service
	hits, err := services.Search.Search(context.Background(), req.ProjectID, req.Query, req.Limit, req.Filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	queryTimeMS := int(time.Since(start).Milliseconds())

	// Ensure empty array instead of null in JSON
	if hits == nil {
		hits = []SearchHit{}
	}

	return c.Status(fiber.StatusOK).JSON(SearchResponse{
		Hits:        hits,
		Total:       len(hits),
		QueryTimeMS: queryTimeMS,
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

// OCRHandler handles POST /v1/media/ocr for optical character recognition
func OCRHandler(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req OCRRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.ProjectID == "" || req.ImageURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and image_url required"})
	}

	// Initialize OCR handler
	ocrHandler, err := media.NewGoogleVisionOCR(ctx, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to initialize OCR handler: %v", err),
		})
	}
	defer ocrHandler.Close()

	// Extract text from image URL
	result, err := ocrHandler.ExtractFromURL(ctx, req.ImageURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("OCR extraction failed: %v", err),
		})
	}

	// Convert bounding boxes to API response format
	textRegions := make([]TextRegion, 0)
	for _, box := range result.BoundingBoxes {
		textRegions = append(textRegions, TextRegion{
			Text:       box.Text,
			Confidence: box.Confidence,
			X1:         box.X1,
			Y1:         box.Y1,
			X2:         box.X2,
			Y2:         box.Y2,
		})
	}

	// Determine extraction status
	extractionStatus := "success"
	if result.ConfidenceScore < 0.5 {
		extractionStatus = "partial"
	}
	if result.Text == "" {
		extractionStatus = "failed"
	}

	// Generate a temporary media_item_id for this response
	mediaItemID := "media_" + fmt.Sprintf("%d", time.Now().Unix())

	response := OCRResponse{
		MediaItemID:    mediaItemID,
		Text:           result.Text,
		Confidence:     result.ConfidenceScore,
		Language:       result.Language,
		TextRegions:    textRegions,
		ProcessedAt:    time.Now().UTC().Format(time.RFC3339),
		ExtractionStat: extractionStatus,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// IngestHandler handles POST /v1/ingest
func IngestHandler(c fiber.Ctx) error {
	var req IngestRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.ProjectID == "" || req.Source.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and source.type required"})
	}

	// Basic source validation by type
	switch req.Source.Type {
	case "url":
		if req.Source.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.url required for type=url"})
		}
	case "crawl":
		if req.Source.Crawl == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.crawl required for type=crawl"})
		}
		mode := req.Source.Crawl.Mode
		if mode == "" {
			mode = "crawl"
			req.Source.Crawl.Mode = mode
		}
		switch mode {
		case "single":
			if req.Source.Crawl.StartURL == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "crawl.start_url required for mode=single"})
			}
		case "sitemap":
			if req.Source.Crawl.SitemapURL == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "crawl.sitemap_url required for mode=sitemap"})
			}
		case "crawl":
			if req.Source.Crawl.StartURL == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "crawl.start_url required for mode=crawl"})
			}
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported crawl.mode"})
		}
	case "openapi":
		if req.Source.OpenAPIURL == "" && req.Source.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.openapi_url or source.url required for openapi"})
		}
	case "github":
		if req.Source.Owner == "" || req.Source.Repo == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.owner and source.repo required for github"})
		}
	case "document", "documents", "file", "files", "pdf", "markdown", "md", "txt":
		// Accept either single source.url or files.urls
		if req.Source.URL == "" && (req.Source.Files == nil || len(req.Source.Files.URLs) == 0) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.url or files.urls required for document ingestion"})
		}
	case "image", "images":
		// Allow either single URL or media.urls
		if req.Source.URL == "" && (req.Source.Media == nil || len(req.Source.Media.URLs) == 0) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.url or media.urls required for image ingestion"})
		}
	case "video", "videos":
		if (req.Source.Media == nil || (len(req.Source.Media.URLs) == 0 && len(req.Source.Media.YouTubeIDs) == 0)) && req.Source.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provide media.urls or media.youtube_ids or source.url for video ingestion"})
		}
		// If youtube IDs are provided, prefer transcript flow.
		// Worker will decide provider based on transcript_provider.
	case "youtube":
		if req.Source.Media == nil || len(req.Source.Media.YouTubeIDs) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "media.youtube_ids required for type=youtube"})
		}
	case "upload":
		if req.Source.UploadID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.upload_id required for upload"})
		}
	case "slack", "discord":
		// allow minimal config; worker can validate tokens/channels later
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported source.type"})
	}

	// Build task payload
	payload := IngestTaskPayload{
		ProjectID:      req.ProjectID,
		Source:         req.Source,
		ChunkStrategy:  req.ChunkStrategy,
		ChunkSizeToken: req.ChunkSizeToken,
		FailFast:       req.FailFast,
	}

	jobID := fmt.Sprintf("job_%s_%d", req.ProjectID, time.Now().UnixNano())

	// Try to enqueue via Redis producer if wired
	if services != nil && services.Queue != nil {
		if prod, ok := services.Queue.(*queue.Producer); ok && prod != nil {
			t := queue.Task{Type: "ingest", Payload: payload, ID: jobID}
			if err := prod.Enqueue(context.Background(), t); err != nil {
				// If enqueue fails, return 500 with error
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "enqueue failed", "details": err.Error()})
			}
		}
	}

	// Initialize job status in Redis (best-effort)
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		if opts, err := redis.ParseURL(redisURL); err == nil {
			rdb := redis.NewClient(opts)
			defer rdb.Close()
			key := "cgap:job:" + jobID
			now := time.Now().UTC().Format(time.RFC3339)
			_ = rdb.HSet(context.Background(), key, map[string]any{
				"job_id":     jobID,
				"project_id": req.ProjectID,
				"status":     "queued",
				"processed":  0,
				"total":      0,
				"started_at": now,
				"error":      "",
				"updated_at": now,
			}).Err()
			// TTL to avoid leaking forever (24h)
			_ = rdb.Expire(context.Background(), key, 24*time.Hour).Err()
		}
	}

	// Return accepted response
	return c.Status(fiber.StatusAccepted).JSON(IngestResponse{
		JobID:     jobID,
		Status:    "queued",
		ProjectID: req.ProjectID,
	})
}

// IngestStatusHandler handles GET /v1/ingest/:job_id
func IngestStatusHandler(c fiber.Ctx) error {
	jobID := c.Params("job_id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "job_id required"})
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "REDIS_URL not set"})
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// fallback: treat as host:port
		opts = &redis.Options{Addr: redisURL}
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()

	key := "cgap:job:" + jobID
	m, err := rdb.HGetAll(context.Background(), key).Result()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "redis error"})
	}
	if len(m) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}

	// Map fields into response
	resp := IngestStatusResponse{
		JobID:      m["job_id"],
		ProjectID:  m["project_id"],
		Status:     m["status"],
		StartedAt:  m["started_at"],
		FinishedAt: m["finished_at"],
		Error:      m["error"],
	}
	if v, ok := m["processed"]; ok {
		if n, perr := strconv.Atoi(v); perr == nil {
			resp.Processed = n
		}
	}
	if v, ok := m["total"]; ok {
		if n, perr := strconv.Atoi(v); perr == nil {
			resp.Total = n
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)
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

	// Media handlers
	app.Post("/v1/media/ocr", OCRHandler)

	// Ingest
	app.Post("/v1/ingest", IngestHandler)
	app.Get("/v1/ingest/:job_id", IngestStatusHandler)
	// Dev seed
	app.Post("/v1/dev/seed", DevSeedHandler)

	// Analytics
	app.Get("/v1/analytics/:project_id", AnalyticsHandler)

	// Gaps
	app.Get("/v1/gaps/:project_id", GapsHandler)
}

var uuidReHandlers = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func looksLikeUUID(s string) bool { return uuidReHandlers.MatchString(s) }
