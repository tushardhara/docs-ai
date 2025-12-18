package api

import (
	"cgap/internal/embedding"
	"cgap/internal/media"
	"cgap/internal/model"
	"cgap/internal/queue"
	"context"
	"fmt"
	"log/slog"
	"strings"
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
	extractionStatus := model.ExtractionSuccess
	if result.ConfidenceScore < 0.5 {
		extractionStatus = model.ExtractionPartial
	}
	if result.Text == "" {
		extractionStatus = model.ExtractionFailed
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

// YouTubeHandler handles POST /v1/media/youtube for transcript extraction
func YouTubeHandler(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req YouTubeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.ProjectID == "" || req.VideoURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id and video_url required"})
	}

	// Initialize YouTube handler
	youtubeHandler := media.NewYouTubeTranscriptFetcher(nil)

	// Extract video ID from URL
	videoID, err := youtubeHandler.ExtractVideoIDFromURL(req.VideoURL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to extract video ID: %v", err),
		})
	}

	// Get transcript
	result, err := youtubeHandler.GetTranscript(ctx, videoID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Transcript extraction failed: %v", err),
		})
	}

	// Convert segments to response format
	segments := make([]TranscriptSegmentResponse, 0)
	for _, seg := range result.Segments {
		segments = append(segments, TranscriptSegmentResponse{
			Text:         seg.Text,
			StartSeconds: seg.StartSeconds,
			EndSeconds:   seg.EndSeconds,
		})
	}

	// Determine extraction status
	extractionStatus := model.ExtractionSuccess
	if result.Transcript == "" {
		extractionStatus = model.ExtractionFailed
	} else if len(result.Segments) == 0 {
		extractionStatus = model.ExtractionPartial
	}

	// Generate a temporary media_item_id for this response
	mediaItemID := "media_" + fmt.Sprintf("%d", time.Now().Unix())

	response := YouTubeResponse{
		MediaItemID:      mediaItemID,
		Transcript:       result.Transcript,
		Language:         result.Language,
		Segments:         segments,
		IsAutoGenerated:  result.IsAutoGenerated,
		ProcessedAt:      time.Now().UTC().Format(time.RFC3339),
		ExtractionStatus: extractionStatus,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// VideoHandler handles POST /v1/media/video for direct video file transcription
func VideoHandler(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second) // Longer timeout for video processing
	defer cancel()

	var req VideoRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.ProjectID == "" || req.SourceID == "" || req.VideoURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_id, source_id, and video_url are required",
		})
	}

	// Initialize video transcriber
	videoHandler := media.NewVideoTranscriber(slog.Default())

	// Transcribe video
	result, err := videoHandler.TranscribeFromURL(ctx, req.VideoURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Video transcription failed: %v", err),
		})
	}

	// Convert segments to response format
	segments := make([]TranscriptSegmentResponse, 0)
	for _, seg := range result.Segments {
		segments = append(segments, TranscriptSegmentResponse{
			Text:         seg.Text,
			StartSeconds: seg.StartSeconds,
			EndSeconds:   seg.EndSeconds,
		})
	}

	// Determine extraction status
	extractionStatus := model.ExtractionSuccess
	if result.Transcript == "" {
		extractionStatus = model.ExtractionFailed
	} else if len(result.Segments) == 0 {
		extractionStatus = model.ExtractionPartial
	}

	// Generate a temporary media_item_id for this response
	mediaItemID := "media_" + fmt.Sprintf("%d", time.Now().Unix())

	response := VideoResponse{
		MediaItemID:      mediaItemID,
		Transcript:       result.Transcript,
		Language:         result.Language,
		Segments:         segments,
		IsAutoGenerated:  result.IsAutoGenerated,
		Duration:         result.Duration,
		ProcessedAt:      time.Now().UTC().Format(time.RFC3339),
		ExtractionStatus: extractionStatus,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// MediaProcessHandler handles POST /v1/media/process - unified media processing endpoint
func MediaProcessHandler(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var req MediaProcessRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.ProjectID == "" || req.SourceID == "" || req.MediaURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_id, source_id, and media_url are required",
		})
	}

	// Initialize media orchestrator
	orchestrator, err := media.NewMediaOrchestrator(slog.Default())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to initialize media processor: %v", err),
		})
	}
	defer orchestrator.Close()

	// Detect media type if not provided
	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = orchestrator.DetectMediaType(req.MediaURL)
		slog.Info("Auto-detected media type", "url", req.MediaURL, "type", mediaType)
	}

	// Validate media type
	supportedTypes := orchestrator.GetSupportedTypes()
	isSupported := false
	for _, t := range supportedTypes {
		if t == mediaType {
			isSupported = true
			break
		}
	}
	if !isSupported {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":           fmt.Sprintf("Unsupported media type: %s", mediaType),
			"supported_types": supportedTypes,
		})
	}

	// Create media item
	mediaItem := &media.MediaItem{
		ID:        fmt.Sprintf("media_%d", time.Now().Unix()),
		ProjectID: req.ProjectID,
		SourceID:  req.SourceID,
		Type:      mediaType,
		URL:       req.MediaURL,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Save media item to database if DB is available
	if services != nil && services.DB != nil {
		store := media.NewMediaStore(services.DB)

		// Create media item record
		if err := store.CreateMediaItem(ctx, mediaItem); err != nil {
			slog.Warn("Failed to save media item to database", "error", err)
			// Continue processing even if DB save fails
		}

		// Update status to processing
		if err := store.UpdateMediaItemStatus(ctx, mediaItem.ID, "processing", nil); err != nil {
			slog.Warn("Failed to update media item status", "error", err)
		}
	}

	// Process media
	result, err := orchestrator.ProcessMediaItem(ctx, mediaItem)
	if err != nil {
		// Update status to failed if DB is available
		if services != nil && services.DB != nil {
			store := media.NewMediaStore(services.DB)
			errMsg := err.Error()
			_ = store.UpdateMediaItemStatus(ctx, mediaItem.ID, "failed", &errMsg)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Media processing failed: %v", err),
		})
	}

	// Save extracted text to database if DB is available
	if services != nil && services.DB != nil {
		store := media.NewMediaStore(services.DB)

		if err := store.SaveExtractedText(ctx, result); err != nil {
			slog.Warn("Failed to save extracted text to database", "error", err)
			// Continue even if DB save fails
		}

		// Update status to completed
		if err := store.UpdateMediaItemStatus(ctx, mediaItem.ID, "completed", nil); err != nil {
			slog.Warn("Failed to update media item status to completed", "error", err)
		}
	}

	// Build response
	response := MediaProcessResponse{
		MediaItemID:      result.MediaItemID,
		MediaType:        mediaType,
		Text:             result.Text,
		Language:         result.Language,
		Confidence:       result.Confidence,
		ContentType:      result.ContentType,
		Metadata:         result.Metadata,
		ProcessedAt:      result.ExtractedAt,
		ExtractionStatus: result.Status,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ExtensionChatHandler handles POST /v1/extension/chat - browser extension endpoint
func ExtensionChatHandler(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req ExtensionChatRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.ProjectID == "" || req.Question == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_id and question are required",
		})
	}

	// Build DOM context for LLM
	domContext := buildDOMContextString(req.DOM, 20)

	// Perform hybrid search to find relevant docs
	var searchResults []SearchHit
	if services != nil && services.Search != nil {
		results, err := services.Search.Search(ctx, req.ProjectID, req.Question, 5, nil)
		if err != nil {
			slog.Warn("Search failed in extension chat", "error", err)
		} else {
			searchResults = results
		}
	}

	// Build context from search results
	docsContext := buildDocsContext(searchResults)

	// Generate LLM prompt
	prompt := buildExtensionPrompt(req.URL, req.Question, domContext, docsContext)

	// Call LLM
	var guidance string
	var steps []GuidanceStep
	var confidence float32 = 0.8

	if services != nil && services.Chat != nil {
		chatReq := ChatRequest{
			ProjectID: req.ProjectID,
			Query:     prompt,
		}

		chatResp, err := services.Chat.Chat(ctx, chatReq)
		if err != nil {
			slog.Error("LLM call failed in extension chat", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate guidance",
			})
		}

		guidance = chatResp.Answer
		confidence = chatResp.Confidence

		// Parse steps from LLM response
		steps = parseStepsFromGuidance(guidance, req.DOM)
	} else {
		// Fallback mock response
		guidance = fmt.Sprintf("To %s on %s:\n\n1. Look for the main navigation menu\n2. Find the relevant section or button\n3. Click to proceed\n\nNote: This is a mock response. Configure LLM service for real guidance.",
			req.Question, req.URL)
		steps = []GuidanceStep{
			{StepNumber: 1, Description: "Locate main navigation", Confidence: 0.7},
			{StepNumber: 2, Description: "Find relevant section", Confidence: 0.6},
			{StepNumber: 3, Description: "Click to proceed", Confidence: 0.5},
		}
	}

	// Extract citations from search results
	var citations []Citation
	for _, hit := range searchResults {
		citations = append(citations, Citation{
			ChunkID: hit.ChunkID,
			Quote:   hit.Text,
			Score:   hit.Confidence,
		})
	}

	response := ExtensionChatResponse{
		Guidance:    guidance,
		Steps:       steps,
		Confidence:  confidence,
		Sources:     citations,
		NextActions: generateNextActions(req.Question),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// Helper functions for ExtensionChatHandler

func buildDOMContextString(entities []DOMEntity, maxElements int) string {
	if len(entities) == 0 {
		return "No interactive elements detected on the page."
	}

	interactive := filterInteractiveElements(entities)
	if len(interactive) > maxElements {
		interactive = interactive[:maxElements]
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Page has %d interactive elements:", len(interactive)))

	for i, entity := range interactive {
		desc := fmt.Sprintf("%d. %s", i+1, entity.Type)
		if entity.Text != "" {
			text := entity.Text
			if len(text) > 50 {
				text = text[:50] + "..."
			}
			desc += fmt.Sprintf(" \"%s\"", text)
		}
		if entity.Selector != "" {
			desc += fmt.Sprintf(" (%s)", entity.Selector)
		}
		parts = append(parts, desc)
	}

	return strings.Join(parts, "\n")
}

func filterInteractiveElements(entities []DOMEntity) []DOMEntity {
	interactive := map[string]bool{
		"button": true, "input": true, "select": true, "textarea": true,
		"a": true, "link": true,
	}

	var result []DOMEntity
	for _, entity := range entities {
		if interactive[strings.ToLower(entity.Type)] {
			result = append(result, entity)
		}
	}
	return result
}

func buildDocsContext(results []SearchHit) string {
	if len(results) == 0 {
		return "No relevant documentation found."
	}

	var parts []string
	parts = append(parts, "Relevant documentation:")

	for i, hit := range results {
		if i >= 3 {
			break
		}
		content := hit.Text
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		parts = append(parts, fmt.Sprintf("\n%d. Document %s\n   %s", i+1, hit.DocumentID, content))
	}

	return strings.Join(parts, "\n")
}

func buildExtensionPrompt(url, question, domContext, docsContext string) string {
	return fmt.Sprintf(`You are helping a user navigate a web application.

Current Page: %s
User Question: %s

Available Elements on Page:
%s

Relevant Documentation:
%s

Provide clear, step-by-step guidance to answer the user's question. 
For each step, specify which element to interact with using CSS selectors when possible.
Format your response as numbered steps.`, url, question, domContext, docsContext)
}

func parseStepsFromGuidance(guidance string, domEntities []DOMEntity) []GuidanceStep {
	var steps []GuidanceStep

	// Simple parsing: look for numbered lines
	lines := strings.Split(guidance, "\n")
	stepNum := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Check if line starts with a number
		if len(line) > 2 && (line[0] >= '1' && line[0] <= '9') && (line[1] == '.' || line[1] == ')') {
			description := strings.TrimSpace(line[2:])

			// Try to find matching selector from DOM
			selector := findSelectorForStep(description, domEntities)
			action := inferAction(description)

			steps = append(steps, GuidanceStep{
				StepNumber:  stepNum,
				Description: description,
				Selector:    selector,
				Action:      action,
				Confidence:  0.75,
			})
			stepNum++
		}
	}

	return steps
}

func findSelectorForStep(description string, entities []DOMEntity) string {
	descLower := strings.ToLower(description)

	// Look for text matches in description
	for _, entity := range entities {
		if entity.Text != "" && strings.Contains(descLower, strings.ToLower(entity.Text)) {
			if entity.Selector != "" {
				return entity.Selector
			}
		}
	}

	return ""
}

func inferAction(description string) string {
	descLower := strings.ToLower(description)

	if strings.Contains(descLower, "click") {
		return "click"
	}
	if strings.Contains(descLower, "type") || strings.Contains(descLower, "enter") {
		return "type"
	}
	if strings.Contains(descLower, "select") || strings.Contains(descLower, "choose") {
		return "select"
	}

	return "navigate"
}

func generateNextActions(question string) []string {
	// Simple heuristics for suggesting next actions
	questionLower := strings.ToLower(question)

	if strings.Contains(questionLower, "create") || strings.Contains(questionLower, "add") {
		return []string{
			"How do I edit this after creating it?",
			"Can I duplicate this?",
			"How do I delete this if I make a mistake?",
		}
	}

	if strings.Contains(questionLower, "dashboard") {
		return []string{
			"How do I customize the dashboard?",
			"Can I share this dashboard?",
			"How do I export dashboard data?",
		}
	}

	return []string{
		"What else can I do here?",
		"How do I save my changes?",
		"Where can I find more help?",
	}
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
	case model.SourceTypeCrawl:
		if req.Source.Crawl == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "source.crawl required for type=crawl"})
		}
		mode := req.Source.Crawl.Mode
		if mode == "" {
			mode = model.SourceTypeCrawl
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
		case model.SourceTypeCrawl:
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
	payload := IngestTaskPayload(req)

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
	app.Post("/v1/media/process", MediaProcessHandler) // Unified endpoint (auto-detects type)
	app.Post("/v1/media/ocr", OCRHandler)
	app.Post("/v1/media/youtube", YouTubeHandler)
	app.Post("/v1/media/video", VideoHandler)

	// Browser Extension
	app.Post("/v1/extension/chat", ExtensionChatHandler)

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
