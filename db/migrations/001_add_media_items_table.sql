-- +goose Up
-- +goose StatementBegin

-- media_items table: Store references to media (images, videos, PDFs)
-- Used for OCR extraction and YouTube transcript fetching
CREATE TABLE IF NOT EXISTS media_items (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  source_id uuid NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
  type text NOT NULL CHECK (type IN ('image', 'video', 'pdf', 'audio')),
  -- URL or file path (cloud storage)
  url text NOT NULL,
  -- For YouTube videos: video_id from URL
  external_id text,
  -- OCR or transcript status: pending, processing, completed, failed
  processing_status text DEFAULT 'pending' CHECK (processing_status IN ('pending', 'processing', 'completed', 'failed')),
  -- File metadata
  file_size_bytes int,
  duration_seconds int,
  -- Processing metadata
  error_message text,
  processed_at timestamptz,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

CREATE INDEX media_items_project ON media_items(project_id);
CREATE INDEX media_items_source ON media_items(source_id);
CREATE INDEX media_items_status ON media_items(processing_status);
CREATE INDEX media_items_external_id ON media_items(external_id);
CREATE INDEX media_items_created ON media_items(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS media_items CASCADE;

-- +goose StatementEnd
