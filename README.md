# cgap: Open-Source Documentation AI Assistant

cgap is an AI-powered documentation assistant that helps teams answer questions, deflect support tickets, and identify coverage gaps using hybrid semantic search (Meilisearch + pgvector), LLM-powered responses, and intelligent analytics.

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 16+ (optional, Docker includes it)

### Local Setup (Docker Compose)

1. Clone the repository:
```bash
git clone https://github.com/yourusername/cgap.git
cd cgap
```

2. Start the full stack:
```bash
docker-compose up
```

This starts:
- **PostgreSQL 16** (localhost:5432) - Metadata, conversations, analytics
- **Meilisearch 1.8** (localhost:7700) - Full-text search
- **Redis 7** (localhost:6379) - Job queue and caching
- **cgap API** (localhost:8080) - REST API server
- **cgap Worker** (background) - Async ingestion and gap detection

3. Check health:
```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Local Development Setup

1. Install Go dependencies:
```bash
go mod download
```

2. Set environment variables (.env):
```bash
# Database
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/cgap?sslmode=disable"
export POSTGRES_PASSWORD=postgres

# Meilisearch
export MEILI_URL=http://localhost:7700
export MEILI_API_KEY=masterKey

# Redis
export REDIS_URL=redis://localhost:6379

# LLM
export LLM_PROVIDER=openai  # or anthropic
export LLM_API_KEY=sk-...
export LLM_MODEL=gpt-4-turbo

# API
export PORT=8080
export WORKER_PORT=8081
```

3. Run migrations:
```bash
brew install goose  # macOS
goose -dir migrations postgres "$DATABASE_URL" up
```

4. Run the API locally:
```bash
go run cmd/api/main.go
```

5. Run the worker locally:
```bash
go run cmd/worker/main.go
```

## Architecture

```
┌─────────────────────────────────────┐
│     Frontend / Client Widget        │
└────────────┬────────────────────────┘
             │ HTTP/REST
┌────────────v────────────────────────┐
│         cgap API (Go)               │
│  ├─ /v1/chat (Q&A with context)    │
│  ├─ /v1/search (hybrid retrieval)  │
│  ├─ /v1/deflect (ticket deflector) │
│  ├─ /v1/analytics (metrics)        │
│  └─ /v1/gaps (coverage analysis)   │
└────┬─────────────────────────────────┘
     │
  ┌──┴──────────────────────────────┐
  │                                 │
  v                                 v
┌─────────────────┐      ┌──────────────────┐
│  PostgreSQL 16  │      │   Meilisearch 1.8│
│  + pgvector     │      │                  │
│  ├─ Chunks      │      │  Full-text      │
│  ├─ Threads     │      │  Search Index   │
│  ├─ Analytics   │      │                 │
│  └─ Gaps        │      │  + Ranking Rules│
└────────┬────────┘      └────────────────┘
         │
         v
    ┌────────────┐
    │   Redis    │
    │   Queue +  │
    │   Cache    │
    └────┬───────┘
         │
         v
    ┌─────────────────────┐
    │  cgap Worker (Go)   │
    │  ├─ Ingestion      │
    │  ├─ Embeddings     │
    │  ├─ Gap Detection  │
    │  └─ Analytics      │
    └─────────────────────┘
```

## API Examples

### 1. Ask a Question (Chat)

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_123",
    "thread_id": "thread_456",
    "message": "How do I deploy to production?",
    "context": {
      "user_id": "user_789",
      "conversation_id": "conv_abc"
    }
  }'
```

Response:
```json
{
  "answer": "To deploy to production, follow these steps: 1. Run `make build` to compile... [continued with context from docs]",
  "citations": [
    {"chunk_id": "chunk_xyz", "document_id": "doc_456", "title": "Deployment Guide", "url": "https://docs.example.com/deploy"}
  ],
  "thread_id": "thread_456",
  "message_id": "msg_789"
}
```

### 2. Search Documentation (Hybrid)

```bash
curl -X POST http://localhost:8080/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_123",
    "query": "API authentication methods",
    "limit": 10,
    "filters": {
      "document_type": "api_reference"
    }
  }'
```

Response:
```json
{
  "hits": [
    {
      "chunk_id": "chunk_001",
      "document_id": "doc_abc",
      "title": "Authentication Overview",
      "url": "https://docs.example.com/auth",
      "content": "Our API supports three authentication methods: API keys, OAuth 2.0, and JWT tokens...",
      "score": 0.95
    }
  ],
  "total": 42,
  "query_time_ms": 145
}
```

### 3. Deflect a Support Ticket

```bash
curl -X POST http://localhost:8080/v1/deflect/suggest \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_123",
    "ticket_text": "I cant log in to my account"
  }'
```

Response:
```json
{
  "suggestions": [
    {
      "title": "Troubleshooting Login Issues",
      "url": "https://docs.example.com/login-help",
      "snippet": "If you cannot log in, try the following steps: 1. Clear your browser cache...",
      "confidence": 0.92
    }
  ],
  "deflected": true
}
```

### 4. Report Analytics Events

```bash
curl -X POST http://localhost:8080/v1/analytics/events \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_123",
    "event_type": "chat_completed",
    "metadata": {
      "response_time_ms": 2340,
      "answer_helpful": true,
      "deflected_ticket": false
    }
  }'
```

### 5. Get Coverage Gaps

```bash
curl http://localhost:8080/v1/gaps/proj_123?limit=5
```

Response:
```json
{
  "gaps": [
    {
      "topic": "GraphQL API",
      "unanswered_questions": 45,
      "last_seen": "2024-01-15T10:30:00Z",
      "suggestion": "Create GraphQL quickstart guide"
    }
  ],
  "total": 8
}
```

## Ingest Documentation

Queue a documentation crawl/ingest job:

```bash
curl -X POST http://localhost:8080/v1/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_123",
    "source": {
      "type": "url",
      "url": "https://docs.example.com"
    },
    "chunk_strategy": "semantic"
  }'
```

Response:
```json
{
  "job_id": "ingest_xyz123",
  "status": "queued",
  "project_id": "proj_123"
}
```

The worker will:
1. Crawl the documentation tree
2. Extract text from HTML/Markdown
3. Generate embeddings (OpenAI API)
4. Index in Meilisearch
5. Store chunks in PostgreSQL
6. Report completion via webhook

## Project Structure

```
cgap/
├── cmd/
│   ├── api/
│   │   └── main.go              # API server entrypoint
│   └── worker/
│       └── main.go              # Worker server entrypoint
├── api/
│   ├── types.go                 # Request/response DTOs and service interfaces
│   └── handlers.go              # HTTP request handlers
├── internal/
│   ├── model/
│   │   └── types.go             # Shared domain models
│   ├── service/
│   │   └── service.go           # Business logic (ChatService, SearchService, etc.)
│   ├── storage/
│   │   └── repo.go              # Repository interfaces
│   ├── postgres/
│   │   └── store.go             # PostgreSQL implementation of storage.Store
│   ├── meilisearch/
│   │   └── client.go            # Meilisearch HTTP client
│   ├── llm/
│   │   └── client.go            # LLM API wrapper (OpenAI/Anthropic)
│   ├── queue/
│   │   └── client.go            # Redis queue producer/consumer
│   └── ingestion/
│       └── pipeline.go          # Document ingestion pipeline
├── db/
│   └── schema.sql               # PostgreSQL DDL with pgvector
├── migrations/
│   └── 0001_init.sql            # Goose migration format
├── build/
│   ├── Dockerfile.api           # Multi-stage API build
│   └── Dockerfile.worker        # Multi-stage worker build
├── scripts/
│   └── meili_bootstrap.sh       # Meilisearch index creation
├── docker-compose.yml           # Local dev environment
├── go.mod                        # Go module definition
├── go.sum                        # Go dependency checksums
├── .golangci.yml                # Lint configuration
├── openapi.yaml                 # OpenAPI 3.0 specification
├── PRD.md                        # Product requirements document
└── README.md                     # This file
```

## Development Workflow

### Running Tests
```bash
go test ./...
```

- Build binaries:
```bash
go build -o bin/api cmd/api/main.go
go build -o bin/worker cmd/worker/main.go
```

- Integration test (requires real Postgres with pgvector + OpenAI key):
```bash
DATABASE_URL=postgres://user:pass@localhost:5432/dbname \
OPENAI_API_KEY=sk-... \
go test -tags=integration ./internal/search -run TestPGVectorIntegration
```


### Running Linter
```bash
golangci-lint run
```

### Building Binaries
```bash
go build -o api cmd/api/main.go
go build -o worker cmd/worker/main.go
```

### Building Docker Images
```bash
docker build -f build/Dockerfile.api -t cgap:api .
docker build -f build/Dockerfile.worker -t cgap:worker .
```

### Database Migrations
```bash
# Create new migration
goose -dir migrations create add_new_table sql

# Apply migrations
goose -dir migrations postgres "$DATABASE_URL" up

# Rollback
goose -dir migrations postgres "$DATABASE_URL" down
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | - | PostgreSQL connection string |
| `MEILI_URL` | http://localhost:7700 | Meilisearch base URL |
| `MEILI_API_KEY` | masterKey | Meilisearch API key |
| `REDIS_URL` | redis://localhost:6379 | Redis connection URL |
| `LLM_PROVIDER` | openai | LLM provider (openai or anthropic) |
| `LLM_API_KEY` | - | LLM API key |
| `LLM_MODEL` | gpt-4-turbo | LLM model identifier |
| `SEARCH_PROVIDER` | hybrid | Search provider: `pgvector`, `meilisearch`, or `hybrid` |
| `PORT` | 8080 | API server port |
| `WORKER_PORT` | 8081 | Worker server port |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |

## Troubleshooting

### PostgreSQL Connection Failed
```bash
# Check if Postgres is running
docker-compose logs postgres

# Verify credentials in DATABASE_URL
# Default: postgres://postgres:postgres@localhost:5432/cgap?sslmode=disable
```

### Meilisearch Index Not Found
```bash
# Bootstrap Meilisearch with initial indexes
bash scripts/meili_bootstrap.sh

# Or manually create index:
curl -X POST http://localhost:7700/indexes \
  -H "Content-Type: application/json" \
  -d '{"uid": "chunks", "primaryKey": "chunk_id"}'
```

### Worker Not Processing Jobs
```bash
# Check Redis connection
docker-compose logs redis

# Check worker logs
docker-compose logs worker

# Verify job queue has items
redis-cli LLEN cgap:tasks
```

### LLM API Errors
```bash
# Verify API key is set
echo $LLM_API_KEY

# Check LLM provider is valid
# OpenAI: api.openai.com
# Anthropic: api.anthropic.com
```

## Performance Tips

1. **Chunking Strategy**: Semantic chunking (vs. fixed-size) improves retrieval quality. Tune chunk size (512-1024 tokens) based on your docs.

2. **Embedding Model**: Use pgvector with IVFFLAT indexing for scalable semantic search. Rebuild indexes periodically for optimal performance.

3. **Meilisearch Ranking**: Adjust ranking rules in `/scripts/meili_bootstrap.sh` to prioritize relevance for your docs.

4. **Caching**: Redis caches frequently searched chunks and LLM responses. Tune TTL based on content update frequency.

5. **Worker Parallelism**: Scale worker replicas in docker-compose for faster ingestion.

## Roadmap

See [PRD.md](./PRD.md) for the full 8-phase roadmap:
- **Phase 0**: Foundation (core Q&A, search, deflection)
- **Phase 1**: Advanced deflection (automation, webhooks)
- **Phase 2**: Internal assistant (Slack, Teams integration)
- **Phase 3**: Coverage analytics (gap detection, reporting)
- **Phase 4**: Fine-tuning (custom embeddings, reranking)
- **Phase 5**: Scalability (multi-region, caching)
- **Phase 6**: Observability (OpenTelemetry, analytics)
- **Phase 7**: Enterprise (SSO, audit, RBAC)
- **Phase 8**: MCP integration (Claude Desktop, IDE plugins)

## Contributing

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit changes (`git commit -am 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Docs**: See [PRD.md](./PRD.md) for architecture and design docs
