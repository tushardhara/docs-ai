package api

import (
	"github.com/gofiber/fiber/v3"
)

// ChatHandler handles POST /v1/chat
func ChatHandler(c fiber.Ctx) error {
	var req ChatRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// TODO: Implement chat request handling
	// 1. Validate project/thread IDs
	// 2. Search for relevant chunks (hybrid retrieval)
	// 3. Call LLM service with context
	// 4. Save message and response to DB
	// 5. Return ChatResponse

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not implemented"})
}

// SearchHandler handles POST /v1/search
func SearchHandler(c fiber.Ctx) error {
	var req SearchRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// TODO: Implement search request handling
	// 1. Validate project ID
	// 2. Search chunks (Meilisearch + pgvector)
	// 3. Return SearchResponse with ranked hits

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not implemented"})
}

// DeflectSuggestHandler handles POST /v1/deflect/suggest
func DeflectSuggestHandler(c fiber.Ctx) error {
	var req DeflectRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// TODO: Implement deflect suggestion handling
	// 1. Validate project ID
	// 2. Search FAQ/docs for similar questions
	// 3. Return DeflectResponse with suggestions

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not implemented"})
}

// DeflectEventHandler handles POST /v1/deflect/event
func DeflectEventHandler(c fiber.Ctx) error {
	var req DeflectEventRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// TODO: Implement deflect event logging
	// 1. Validate project ID
	// 2. Store deflect event (deflected/escalated/etc.)
	// 3. Update analytics

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not implemented"})
}

// IngestHandler handles POST /v1/ingest
func IngestHandler(c fiber.Ctx) error {
	var req IngestRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// TODO: Implement ingest request handling
	// 1. Validate project ID
	// 2. Queue IngestJob for worker
	// 3. Return IngestResponse with job ID

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not implemented"})
}

// AnalyticsHandler handles GET /v1/analytics/:project_id
func AnalyticsHandler(c fiber.Ctx) error {
	// TODO: Implement analytics retrieval
	// 1. Extract project_id from URL
	// 2. Query analytics_events for aggregations
	// 3. Return AnalyticsResponse with metrics

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not implemented"})
}

// GapsHandler handles GET /v1/gaps/:project_id
func GapsHandler(c fiber.Ctx) error {
	// TODO: Implement gaps retrieval
	// 1. Extract project_id from URL
	// 2. Query gap_clusters for recent clusters
	// 3. Return GapsResponse with ranked gaps

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not implemented"})
}

// HealthHandler handles GET /health
func HealthHandler(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

// RegisterRoutes registers all HTTP handlers with Fiber.
func RegisterRoutes(app *fiber.App) {
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

	// Analytics
	app.Get("/v1/analytics/:project_id", AnalyticsHandler)

	// Gaps
	app.Get("/v1/gaps/:project_id", GapsHandler)
}
