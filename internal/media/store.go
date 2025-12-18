package media

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sourceTypeOCR               = "ocr"
	sourceTypeYouTubeTranscript = "youtube_transcript"
	sourceTypeAudioTranscript   = "audio_transcript"
	sourceTypePDFText           = "pdf_text"

	contentTypeTranscript = "transcript"
)

// MediaStore handles database operations for media items and extracted text
type MediaStore struct {
	db *pgxpool.Pool
}

// NewMediaStore creates a new media store
func NewMediaStore(db *pgxpool.Pool) *MediaStore {
	return &MediaStore{db: db}
}

// CreateMediaItem inserts a new media item record
func (s *MediaStore) CreateMediaItem(ctx context.Context, item *MediaItem) error {
	query := `
		INSERT INTO media_items (
			id, project_id, source_id, type, url, external_id,
			processing_status, file_size_bytes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	id, err := uuid.Parse(item.ID)
	if err != nil {
		id = uuid.New()
		item.ID = id.String()
	}

	projectID, err := uuid.Parse(item.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project_id: %w", err)
	}

	sourceID, err := uuid.Parse(item.SourceID)
	if err != nil {
		return fmt.Errorf("invalid source_id: %w", err)
	}

	now := time.Now()
	_, err = s.db.Exec(ctx, query,
		id,
		projectID,
		sourceID,
		item.Type,
		item.URL,
		item.ExternalID,
		"pending",
		item.FileSizeBytes,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create media item: %w", err)
	}

	return nil
}

// UpdateMediaItemStatus updates the processing status of a media item
func (s *MediaStore) UpdateMediaItemStatus(ctx context.Context, mediaItemID, status string, errorMsg *string) error {
	query := `
		UPDATE media_items
		SET processing_status = $1,
		    error_message = $2,
		    processed_at = CASE WHEN $1 IN ('completed', 'failed') THEN NOW() ELSE processed_at END,
		    updated_at = NOW()
		WHERE id = $3
	`

	id, err := uuid.Parse(mediaItemID)
	if err != nil {
		return fmt.Errorf("invalid media_item_id: %w", err)
	}

	_, err = s.db.Exec(ctx, query, status, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update media item status: %w", err)
	}

	return nil
}

// SaveExtractedText saves extracted text content to the database
func (s *MediaStore) SaveExtractedText(ctx context.Context, content *ExtractedContent) error {
	query := `
		INSERT INTO extracted_text (
			id, media_item_id, source_type, text, confidence_score,
			language, extracted_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	id := uuid.New()
	mediaItemID, err := uuid.Parse(content.MediaItemID)
	if err != nil {
		return fmt.Errorf("invalid media_item_id: %w", err)
	}

	// Map content type to source_type
	sourceType := mapContentTypeToSourceType(content.ContentType, content.Metadata)

	var confidenceScore *float64
	if content.Confidence > 0 {
		confidenceScore = &content.Confidence
	}

	extractedAt, err := time.Parse(time.RFC3339, content.ExtractedAt)
	if err != nil {
		extractedAt = time.Now()
	}

	now := time.Now()
	_, err = s.db.Exec(ctx, query,
		id,
		mediaItemID,
		sourceType,
		content.Text,
		confidenceScore,
		content.Language,
		extractedAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save extracted text: %w", err)
	}

	return nil
}

// GetMediaItem retrieves a media item by ID
func (s *MediaStore) GetMediaItem(ctx context.Context, mediaItemID string) (*MediaItem, error) {
	query := `
		SELECT id, project_id, source_id, type, url, external_id,
		       processing_status, file_size_bytes, created_at
		FROM media_items
		WHERE id = $1
	`

	id, err := uuid.Parse(mediaItemID)
	if err != nil {
		return nil, fmt.Errorf("invalid media_item_id: %w", err)
	}

	var item MediaItem
	var projectID, sourceID uuid.UUID
	var createdAt time.Time

	err = s.db.QueryRow(ctx, query, id).Scan(
		&id,
		&projectID,
		&sourceID,
		&item.Type,
		&item.URL,
		&item.ExternalID,
		&item.FileSizeBytes,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get media item: %w", err)
	}

	item.ID = id.String()
	item.ProjectID = projectID.String()
	item.SourceID = sourceID.String()
	item.CreatedAt = createdAt.Format(time.RFC3339)

	return &item, nil
}

// GetExtractedText retrieves all extracted text for a media item
func (s *MediaStore) GetExtractedText(ctx context.Context, mediaItemID string) ([]*ExtractedContent, error) {
	query := `
		SELECT id, media_item_id, source_type, text, confidence_score,
		       language, extracted_at
		FROM extracted_text
		WHERE media_item_id = $1
		ORDER BY created_at ASC
	`

	id, err := uuid.Parse(mediaItemID)
	if err != nil {
		return nil, fmt.Errorf("invalid media_item_id: %w", err)
	}

	rows, err := s.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query extracted text: %w", err)
	}
	defer rows.Close()

	var results []*ExtractedContent
	for rows.Next() {
		var content ExtractedContent
		var textID uuid.UUID
		var mediaID uuid.UUID
		var sourceType string
		var confidenceScore *float64
		var extractedAt time.Time

		err := rows.Scan(
			&textID,
			&mediaID,
			&sourceType,
			&content.Text,
			&confidenceScore,
			&content.Language,
			&extractedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan extracted text: %w", err)
		}

		content.MediaItemID = mediaID.String()
		content.ContentType = mapSourceTypeToContentType(sourceType)
		if confidenceScore != nil {
			content.Confidence = *confidenceScore
		}
		content.ExtractedAt = extractedAt.Format(time.RFC3339)

		results = append(results, &content)
	}

	return results, nil
}

// Helper functions

func mapContentTypeToSourceType(contentType string, metadata map[string]interface{}) string {
	if metadata != nil {
		if _, ok := metadata["youtube"]; ok {
			return sourceTypeYouTubeTranscript
		}
		if _, ok := metadata["video"]; ok {
			return sourceTypeAudioTranscript
		}
		if _, ok := metadata["ocr"]; ok {
			return sourceTypeOCR
		}
	}

	switch contentType {
	case contentTypeText:
		return sourceTypeOCR
	case contentTypeTranscript:
		return sourceTypeAudioTranscript
	default:
		return sourceTypeOCR
	}
}

func mapSourceTypeToContentType(sourceType string) string {
	switch sourceType {
	case sourceTypeOCR:
		return contentTypeText
	case sourceTypeYouTubeTranscript, sourceTypeAudioTranscript:
		return contentTypeTranscript
	case sourceTypePDFText:
		return contentTypeText
	default:
		return contentTypeText
	}
}
