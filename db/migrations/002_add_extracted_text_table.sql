-- +goose Up
-- +goose StatementBegin

-- extracted_text table: Store extracted text from media items
-- OCR results from images, YouTube transcripts, PDF text extraction
CREATE TABLE IF NOT EXISTS extracted_text (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  media_item_id uuid NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
  -- Source of extraction: ocr, youtube_transcript, pdf_text, etc.
  source_type text NOT NULL CHECK (source_type IN ('ocr', 'youtube_transcript', 'pdf_text', 'audio_transcript')),
  -- The extracted text content
  text text NOT NULL,
  -- Confidence score for OCR results (0-1)
  confidence_score real,
  -- For videos: timestamp when this text appears
  timestamp_seconds int,
  -- Language detected
  language text,
  -- Processing metadata
  extracted_at timestamptz DEFAULT now(),
  created_at timestamptz DEFAULT now()
);

CREATE INDEX extracted_text_media_item ON extracted_text(media_item_id);
CREATE INDEX extracted_text_source_type ON extracted_text(source_type);
CREATE INDEX extracted_text_created ON extracted_text(created_at DESC);

-- Create a full-text search index on extracted text for fast queries
CREATE INDEX extracted_text_search ON extracted_text USING gin(to_tsvector('english', text));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS extracted_text CASCADE;

-- +goose StatementEnd
