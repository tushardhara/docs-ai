# CGAP: AI Browser Extension for SaaS Setup

**CGAP** provides AI-guided setup assistance inside any SaaS dashboard via a browser extension.

## How It Works

1. **Install** CGAP extension on Chrome (dev mode)
2. **Open** any SaaS dashboard (Mixpanel, Stripe, HubSpot, etc.)
3. **Click** CGAP icon → Ask "How do I create a campaign?"
4. **Get** step-by-step guidance with UI element selectors
5. **Optional** Auto-click each step (user confirms)

## Tech Stack

- **Backend**: Go + PostgreSQL + pgvector + Meilisearch + Redis
- **Extension**: TypeScript + React + Manifest v3
- **LLM**: Gemini / OpenAI
- **Infrastructure**: Docker Compose (local) + Chrome Web Store (extension)

---

## Quick Start (4-Week MVP)

### Prerequisites
```bash
# Install
docker-compose up  # Postgres, Meilisearch, Redis
go mod download     # Go dependencies
npm install         # Node (for extension)

# Env
cp .env.example .env  # Edit with your API keys
```

### Backend (API + Worker)
```bash
# Terminal 1: API server
go run cmd/api/main.go      # Runs on :8080

# Terminal 2: Worker
go run cmd/worker/main.go   # Runs on :8081

# Verify
curl http://localhost:8080/health  # {"status":"ok"}
```

### Extension (Chrome)
```bash
cd extension
npm run build        # Compile to dist/

# In Chrome:
# 1. chrome://extensions/
# 2. Enable "Developer mode"
# 3. Load unpacked → select extension/dist/
# 4. Click CGAP icon on any SaaS website
```

---

## 4-Week Sprint

| Week | Deliverable | Status |
|------|-------------|--------|
| **1** | Sources + Media DB schema | `TODO` |
| **2** | OCR + YouTube ingest handler | `TODO` |
| **3** | Extension chat endpoint API | `TODO` |
| **4** | Browser extension UI | `TODO` |

See [TODO_MVP.md](./TODO_MVP.md) for detailed checklist.

---

## API Endpoints (MVP)

### Extension Chat (Week 3)
```bash
POST /v1/extension/chat
{
  "project_id": "demo",
  "question": "How do I create a dashboard?",
  "dom": "{...DOM JSON...}",
  "screenshot": "data:image/png;base64,...",
  "url": "https://mixpanel.com/dashboards",
  "page_title": "Dashboards"
}

# Response
{
  "guidance": "You're on the Dashboards page...",
  "steps": [
    {"description": "Click New Dashboard", "selector": "button.new-dashboard", "action": "click"},
    {"description": "Enter name", "selector": "input.dashboard-name", "action": "fill"}
  ],
  "confidence": 0.92
}
```

### Ingest (Week 2)
```bash
POST /v1/ingest
{
  "project_id": "demo",
  "source": {
    "type": "crawl",
    "start_url": "https://docs.mixpanel.com",
    "crawl": {"scope": "prefix", "max_pages": 100}
  }
}

# Returns
{"job_id": "ingest_xyz", "status": "queued"}
```

### Search
```bash
POST /v1/search
{
  "project_id": "demo",
  "query": "create dashboard",
  "limit": 5
}

# Returns
{
  "hits": [
    {"text": "To create a dashboard...", "confidence": 0.95, "document_id": "doc_1"}
  ],
  "query_time_ms": 125
}
```

---

## Database Schema (Week 1)

### Core Tables
```sql
-- Sources (new)
CREATE TABLE sources (
  id UUID PRIMARY KEY,
  project_id UUID,
  type TEXT,           -- crawl, github, youtube, etc.
  url TEXT,
  crawl_mode TEXT,
  last_crawled_at TIMESTAMP,
  status TEXT          -- active, paused, failed
);

-- Media Items (new)
CREATE TABLE media_items (
  id UUID PRIMARY KEY,
  project_id UUID,
  type TEXT,           -- image, video, youtube
  source_url TEXT,
  created_at TIMESTAMP
);

-- Extracted Text (new)
CREATE TABLE extracted_text (
  id UUID PRIMARY KEY,
  media_id UUID,
  text TEXT,
  source_type TEXT,    -- ocr, transcript
  confidence_score FLOAT,
  created_at TIMESTAMP
);

-- Document Sources (mapping)
CREATE TABLE document_sources (
  document_id UUID,
  source_id UUID,
  PRIMARY KEY (document_id, source_id)
);

-- Existing (unchanged)
documents, chunks, chunk_embeddings, threads, messages, etc.
```

---

## Environment Variables

```bash
# Database
DATABASE_URL=postgres://cgap:cgap@localhost:5432/cgap

# Search + Cache
MEILISEARCH_URL=http://localhost:7700
REDIS_URL=redis://localhost:6379

# LLM & Embeddings
LLM_PROVIDER=google        # or openai
GEMINI_API_KEY=your-key
EMBEDDING_PROVIDER=google
EMBEDDING_MODEL=gemini-embedding-001

# Media APIs
GOOGLE_VISION_API_KEY=your-key
YOUTUBE_API_KEY=your-key

# Servers
PORT=8080
HEALTH_PORT=8081
LOG_LEVEL=info
```

---

## Demo (End of Week 4)

```
Live Demo on Mixpanel:
1. Open: https://mixpanel.com/dashboards (logged in)
2. Click CGAP extension icon
3. Ask: "How do I create a dashboard for user events?"
4. See guidance:
   - "Step 1: Click 'Dashboards' in left sidebar"
   - "Step 2: Click 'New Dashboard' button"
   - "Step 3: Enter name 'User Events Dashboard'"
   - "Step 4: Click 'Create'"
5. Optionally: Click "Auto-execute" → extension auto-clicks each step
```

---

## Project Structure

```
cgap/
├── cmd/
│   ├── api/main.go              # API server
│   └── worker/main.go           # Worker server
├── api/
│   ├── types.go                 # DTOs
│   └── handlers.go              # HTTP handlers
├── internal/
│   ├── service/service.go       # Business logic
│   ├── storage/                 # DB interfaces
│   ├── postgres/                # PostgreSQL impl
│   ├── meilisearch/             # Search impl
│   ├── embedding/               # Embeddings
│   └── llm/                     # LLM clients
├── extension/
│   ├── manifest.json            # Chrome extension manifest
│   ├── src/
│   │   ├── popup/App.tsx        # Extension popup UI
│   │   ├── content/script.ts    # DOM capture
│   │   └── utils/               # API client, storage
│   └── dist/                    # Built extension
├── db/migrations/               # Database migrations
├── docker-compose.yml           # Local dev stack
└── README.md
```

---

## Next Steps

1. **Week 1**: Create DB migrations, run `go build ./...`
2. **Week 2**: Add OCR + YouTube handlers to worker, test ingest
3. **Week 3**: Add `/v1/extension/chat` endpoint, test with curl
4. **Week 4**: Build extension UI, demo on real SaaS dashboard

See [TODO_MVP.md](./TODO_MVP.md) for detailed checklist.

---

## Contributing

```bash
# Development
go mod tidy
go build ./...
go test ./...

# Linting
golangci-lint run

# Docker
docker-compose up      # Full stack
docker-compose logs -f api worker

# Extension development
cd extension
npm run dev            # Watch mode
npm run build          # Production build
```

---

## License

MIT License - see LICENSE file

## Support

- **Docs**: See [PITCH_MVP.md](./PITCH_MVP.md) for strategy
- **Roadmap**: See [TODO_MVP.md](./TODO_MVP.md) for 4-week sprint
- **Architecture**: See [PRD.md](./PRD.md) for Phase 0 scope
