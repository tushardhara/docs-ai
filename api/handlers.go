package api

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

var services *Services

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
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

// RegisterRoutes registers all HTTP handlers with Fiber (deprecated).
func RegisterRoutes(app *fiber.App) {
	RegisterRoutesWithServices(app, &Services{})
}

// RegisterRoutesWithServices registers all HTTP handlers with Fiber and injects services.
func RegisterRoutesWithServices(app *fiber.App, svc *Services) {
	services = svc

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
