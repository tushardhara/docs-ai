# CGAP AI Coding Agent Instructions

## Project Overview
**CGAP** is a browser extension providing AI-guided SaaS setup assistance. Built with Go (API + Worker), PostgreSQL + pgvector, Meilisearch, Redis, and supports multiple LLM/embedding providers (OpenAI, Google Gemini, Anthropic, xAI Grok).

**Architecture**: Hybrid RAG system with dual search (pgvector semantic + Meilisearch full-text) feeding LLM chat endpoints. Media processing pipeline handles OCR (Google Vision), YouTube transcripts, and video transcription (Whisper/AssemblyAI).

## Core Development Patterns

### 1. Repository Pattern with Interface Segregation
All data access goes through `internal/storage/repo.go` interfaces implemented by `internal/postgres/store.go`:
```go
// Define interface in storage/repo.go
type DocumentRepo interface {
    GetByID(ctx context.Context, id string) (*model.Document, error)
    Create(ctx context.Context, d *model.Document) error
}

// Implement in postgres/store.go
type DocumentRepo struct { pool *pgxpool.Pool }
```
**Never bypass the repository layer** - always add methods to the interface first.

### 2. Provider Pattern for Pluggable Components
LLM, embedding, and search clients use interfaces for multi-provider support:
- **LLMs**: `internal/llm/` - Select via `LLM_PROVIDER` env (openai|google|anthropic|grok)
- **Embeddings**: `internal/embedding/` - Select via `EMBEDDING_PROVIDER` env (openai|google|http|mock)
- **Search**: `internal/search/hybrid_impl.go` - Dual search (pgvector + Meilisearch)

When adding providers: implement the interface (`Embedder`, `LLM`, `Search`) and register in main.go initialization.

### 3. Service Layer Business Logic
`internal/service/service.go` orchestrates between storage, search, and LLM:
```go
type ChatServiceImpl struct {
    store  storage.Store
    llm    LLM
    search Search
}
```
API handlers in `api/handlers.go` are **thin routing layers** - all business logic belongs in service layer.

### 4. Media Processing Pipeline
`internal/media/orchestrator.go` routes by type:
- **Images** → `ocr.go` (Google Vision API)
- **YouTube** → `youtube.go` (transcript fetcher)
- **Video** → `video.go` (Whisper/AssemblyAI)

Media items stored in `media_items` table; extracted text in `extracted_text` table (Week 1 architecture).

### 5. Database Migrations
Use **goose** for migrations in `db/migrations/*.sql`:
```bash
goose -dir db/migrations postgres "$DATABASE_URL" up
```
Migration files follow `NNN_description.sql` pattern with `-- +goose Up` and `-- +goose Down` markers. Schema baseline in `db/schema.sql`.

## Build & Test Commands

### Local Development
```bash
# Start dependencies
docker-compose up postgres meilisearch redis

# Run API server (port 8080)
go run cmd/api/main.go

# Run worker (background jobs)
go run cmd/worker/main.go

# Build everything
go build ./...

# Run all tests
go test ./...

# Run with race detector
go test -race ./...
```

### Environment Setup
Copy `.env.example` to `.env` and set keys:
- `DATABASE_URL` - Required (postgres connection)
- `LLM_PROVIDER` + corresponding API key (OPENAI_API_KEY, GEMINI_API_KEY, etc.)
- `EMBEDDING_PROVIDER` + corresponding API key
- `MEILISEARCH_URL` and `MEILISEARCH_KEY`

### Testing Patterns
- Tests use `*_test.go` files with `package <name>_test` convention
- Mocks in `internal/testutil/mocks.go` and inline service tests
- Integration tests skip by default: `t.Skip("requires live API key")`
- Use `MockLLM`, `MockSearch`, `MockStore` for service layer tests

## Key Files & Their Roles

| File | Purpose |
|------|---------|
| `cmd/api/main.go` | API server entrypoint - wires dependencies |
| `api/handlers.go` | HTTP handlers (Fiber v3) - thin routing |
| `api/types.go` | Request/response DTOs + service interfaces |
| `internal/service/service.go` | Business logic orchestration |
| `internal/storage/repo.go` | Repository interface definitions |
| `internal/postgres/store.go` | PostgreSQL implementations (pgx) |
| `internal/search/hybrid_impl.go` | Dual search (pgvector + Meilisearch) |
| `internal/media/orchestrator.go` | Media processing dispatcher |
| `db/migrations/*.sql` | Database schema changes (goose) |

## Common Tasks

### Adding a New API Endpoint
1. Define request/response types in `api/types.go`
2. Add service interface method (e.g., `ChatService`)
3. Implement in `internal/service/service.go`
4. Add handler in `api/handlers.go` (register route in `cmd/api/main.go`)
5. Update `openapi.yaml` spec

### Adding a New LLM Provider
1. Create `internal/llm/<provider>.go` implementing `LLM` interface
2. Add provider constant in `internal/llm/client.go`
3. Register in `cmd/api/main.go` switch statement
4. Add env vars to `.env.example`

### Adding Database Tables
1. Create migration: `db/migrations/NNN_add_<name>_table.sql`
2. Define model in `internal/model/types.go`
3. Add repository interface in `internal/storage/repo.go`
4. Implement in `internal/postgres/store.go`
5. Update `Store` interface aggregator

### Working with Media
Media flows: Ingest → `media_items` table → Orchestrator → Handler → `extracted_text` table
- Check `internal/media/types.go` for shared types
- Handlers expect `context.Context` and `*MediaItem` input
- Return `*ExtractedContent` with status (pending|completed|failed)

## Project Conventions

- **Error handling**: Wrap with context using `fmt.Errorf("context: %w", err)`
- **Logging**: Use `slog` (structured logging), not `fmt.Println`
- **UUIDs**: Use `github.com/google/uuid` for IDs
- **JSON tags**: Always include on API types
- **Context**: Pass `context.Context` as first param to all service/repo methods
- **Nil slices**: Return empty slices `[]T{}` not `nil` for JSON arrays

## Sprint Context (Week 1-4 MVP)
Currently in **4-week sprint** focused on browser extension MVP:
- **Week 1** ✅ - Database foundation (media_items, extracted_text tables)
- **Week 2** ✅ - Media handlers (OCR, YouTube, video)
- **Week 3** ⏳ - Extension chat endpoint (95% complete)
- **Week 4** - Chrome extension UI

See `docs/WEEK_*_*.md` for detailed sprint plans and status.

## GitHub Workflow (MANDATORY)

### Branch Protection Rules
**ALL changes MUST go through Pull Requests - NO direct pushes to `main`**

### Development Workflow
1. **Pick a GitHub Issue**: Check https://github.com/tushardhara/docs-ai/issues
2. **Create Feature Branch**: `git checkout -b feature/issue-N-description`
3. **Make Changes**: Implement the feature/fix
4. **Test Locally**: Ensure `go build ./...` and `go test ./...` pass
5. **Commit with Convention**: `git commit -m "feat: description (#N)"` or `fix:`, `docs:`, `test:`
6. **Push Branch**: `git push origin feature/issue-N-description`
7. **Create PR**: Link to issue, add description, request review
8. **Get Approval**: Wait for review and approval
9. **Merge**: Squash and merge to main
10. **Close Issue**: Automatically closed via PR merge

### Commit Message Convention
```
feat: Add DOM capture content script (#5)
fix: Handle nil pointer in OCR handler (#7)
docs: Update API endpoint documentation (#8)
test: Add integration tests for video handler (#9)
chore: Update dependencies (#10)
```

### PR Template
When creating PRs, include:
- **Related Issue**: Fixes #N
- **Changes**: Brief description
- **Testing**: How to test the changes
- **Screenshots**: For UI changes

## External Dependencies
- **Database**: PostgreSQL 16+ with pgvector extension
- **Search**: Meilisearch 1.8 for full-text search
- **Queue**: Redis for background jobs
- **APIs**: Google Cloud Vision (OCR), OpenAI/Gemini/Anthropic (LLM), Whisper/AssemblyAI (transcription)

## Debugging Tips
- Health check: `curl http://localhost:8080/health`
- Check logs: API uses custom CGAP banner + structured logging
- DB connection: Test with `docker-compose exec postgres psql -U cgap -d cgap`
- Migration status: `goose -dir db/migrations postgres "$DATABASE_URL" status`
- Redis queue: `docker-compose exec redis redis-cli MONITOR`

## Documentation
- `docs/START_HERE.md` - Quickstart guide
- `docs/INDEX.md` - Documentation index
- `docs/COMMANDS.md` - Common command reference
- `README.md` - Project overview + API examples
- OpenAPI spec: `openapi.yaml`
