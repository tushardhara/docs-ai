# CGAP Week 1: Database Schema Foundation

## Overview

Week 1 focuses on creating the database schema needed for media ingestion and extraction. By Friday EOD, we'll have:

✅ `media_items` table (for images, videos, PDFs, audio)
✅ `extracted_text` table (for OCR results, transcripts)
✅ Proper indices for performance
✅ Migration testing

## What We're Building

### New Tables

#### 1. `media_items` Table
Stores references to media files that will be processed for content extraction.

```sql
CREATE TABLE media_items (
  id uuid PRIMARY KEY,
  project_id uuid REFERENCES projects(id),
  source_id uuid REFERENCES sources(id),
  type text CHECK (type IN ('image', 'video', 'pdf', 'audio')),
  url text,
  external_id text,                    -- YouTube video ID, etc.
  processing_status text,              -- pending|processing|completed|failed
  file_size_bytes int,
  duration_seconds int,
  error_message text,
  processed_at timestamptz,
  created_at timestamptz,
  updated_at timestamptz
);
```

**Use Cases:**
- Store references to OCR-able images
- Store YouTube video IDs for transcript fetching
- Track PDF files for text extraction
- Store audio files for transcription (Phase 2)

#### 2. `extracted_text` Table
Stores text extracted from media items.

```sql
CREATE TABLE extracted_text (
  id uuid PRIMARY KEY,
  media_item_id uuid REFERENCES media_items(id),
  source_type text CHECK (source_type IN ('ocr', 'youtube_transcript', 'pdf_text', 'audio_transcript')),
  text text,
  confidence_score real,               -- 0-1, for OCR
  timestamp_seconds int,               -- For video timestamps
  language text,
  extracted_at timestamptz,
  created_at timestamptz
);
```

**Use Cases:**
- Store OCR output from images (with confidence scores)
- Store YouTube transcripts with timestamps
- Store PDF extracted text
- Enable full-text search on extracted content

## Quick Start

### Step 1: Start Docker Services
```bash
docker-compose up -d
```

Verify services are running:
```bash
# Check PostgreSQL
docker-compose exec db psql -U cgap -d cgap -c "\dt"

# Check Redis
docker-compose exec redis redis-cli PING

# Check Meilisearch
curl http://localhost:7700/health
```

### Step 2: Run Migrations
```bash
cd /Users/tushar.dhara/docs-ai

# View migration status
DATABASE_URL="postgres://cgap:cgap_dev_password@localhost:5432/cgap?sslmode=disable" \
  goose -dir db/migrations postgres "$DATABASE_URL" status

# Apply migrations
DATABASE_URL="postgres://cgap:cgap_dev_password@localhost:5432/cgap?sslmode=disable" \
  goose -dir db/migrations postgres "$DATABASE_URL" up
```

### Step 3: Verify Tables Exist
```bash
# Connect to database
docker-compose exec db psql -U cgap -d cgap

# Inside psql:
\dt media_items
\dt extracted_text
\d media_items
\d extracted_text
```

### Step 4: Build Project
```bash
go build ./...
```

## Migration Files

Located in `db/migrations/`:

- `001_add_media_items_table.sql` - Creates media_items table + indices
- `002_add_extracted_text_table.sql` - Creates extracted_text table + indices

Each migration includes:
- `+goose Up` block (creation)
- `+goose Down` block (rollback)

## Schema Diagram

```
projects
    ↓
    └── sources
        ├── documents (already exists)
        ├── chunks (already exists)
        └── media_items ← NEW
            └── extracted_text ← NEW
```

## Week 1 Acceptance Criteria

**By Friday EOD:**

- [ ] `docker-compose up` succeeds
- [ ] `goose status` shows all migrations applied
- [ ] `psql` shows both new tables with correct structure
- [ ] `go build ./...` passes without errors
- [ ] Indices created and queryable
- [ ] `source_id` foreign key in `media_items` is properly set

**Test Commands:**
```bash
# Check tables exist
psql -U cgap -d cgap -c "\dt media_items"
psql -U cgap -d cgap -c "\dt extracted_text"

# Check indices
psql -U cgap -d cgap -c "\di media_items*"
psql -U cgap -d cgap -c "\di extracted_text*"

# Verify build
go build ./...
echo $?  # Should return 0
```

## Next: Week 2 Preview

Week 2 will build on this foundation:
- **OCR Handler**: Google Vision API → extracted_text
- **YouTube Handler**: youtube-api → extracted_text with transcripts
- **Worker Job**: Process media_items asynchronously

```go
// Week 2 pseudo-code
for mediaItem := range pendingMediaItems {
    switch mediaItem.Type {
    case "image":
        text = ocrHandler.Extract(mediaItem.URL)
    case "video":
        text = youtubeHandler.GetTranscript(mediaItem.ExternalID)
    }
    
    insertExtractedText(mediaItem.ID, text)
}
```

## Troubleshooting

### Migration fails: "column does not exist"
**Cause**: Running migrations out of order
**Fix**: 
```bash
# Rollback all
goose down-to 0

# Re-apply
goose up
```

### Cannot connect to database
**Cause**: PostgreSQL not running
**Fix**:
```bash
docker-compose up -d db
docker-compose logs db
```

### Table already exists error
**Cause**: Migration already applied
**Fix**: Check status first
```bash
goose status
# If applied, migration is idempotent (CREATE TABLE IF NOT EXISTS)
```

## Files Created This Week

- `db/migrations/001_add_media_items_table.sql`
- `db/migrations/002_add_extracted_text_table.sql`
- `db/migrations/migrations.go` (Go helper)
- `db/migrations/run.sh` (Shell script)
- `WEEK_1_CHECKLIST.md` (this file)

## Success Looks Like

```bash
$ psql -U cgap -d cgap
cgap=# \dt
                 List of relations
 Schema |       Name       | Type  |  Owner
--------+------------------+-------+--------
 public | analytics_events | table | cgap
 public | api_keys         | table | cgap
 public | answers          | table | cgap
 public | chunk_embeddings | table | cgap
 public | chunks           | table | cgap
 public | citations        | table | cgap
 public | deflect_events   | table | cgap
 public | documents        | table | cgap
 public | extracted_text   | table | cgap     ← NEW
 public | feedback         | table | cgap
 public | gap_candidates   | table | cgap
 public | gap_cluster_examples | table | cgap
 public | gap_clusters     | table | cgap
 public | media_items      | table | cgap     ← NEW
 public | messages         | table | cgap
 public | project_members  | table | cgap
 public | projects         | table | cgap
 public | sources          | table | cgap
 public | threads          | table | cgap
 public | users            | table | cgap
(21 rows)
```

Done! Both new tables visible. Ready for Week 2.
