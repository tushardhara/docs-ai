# CGAP - Browser Extension MVP (1-Month Sprint)

## 🎯 Mission
**Browser extension that captures page context (DOM + screenshot) and provides AI-guided setup steps with optional auto-click.**

Open any SaaS → click CGAP → ask question → get step-by-step guidance grounded in your docs + videos.

## Timeline: 4 Weeks to MVP Launch

### Phase 0: Browser Extension MVP ✓ (4-Week Sprint)
Week 1: Sources + Media DB | Week 2: Media Ingest | Week 3: Extension Endpoint | Week 4: Extension UI

#### Completed ✓
- [x] Ingest job status API
  - Persist job state in DB/Redis (id, project_id, status=queued|running|completed|failed, processed/total, started_at, finished_at, error)
  - GET `/v1/ingest/{job_id}` endpoint returns status and counters
  - Worker updates progress during ingest
- [x] Sitemap + robots crawler
  - Worker: parse robots.txt and sitemap.xml
  - Implement `single|sitemap|crawl` modes with scope (host/domain/prefix), allow/deny, concurrency, delay_ms, and URL dedup
  - Feed fetched pages into chunk→embed→store pipeline
- [x] OpenAPI documentation
  - Add `fail_fast` to ingest request
  - Document crawl options (mode, start_url, sitemap_url, scope, max_depth, max_pages, respect_robots, concurrency, delay_ms, allow, deny)
  - Define status response with processed/total, started_at, updated_at, finished_at, error
- [x] Structured logging (slog)
  - Migrate from log package to log/slog throughout codebase
  - Both API and worker use slog for structured logging with key-value pairs
- [x] Startup branding
  - Custom CGAP ASCII art banner with cyan colors
  - Print on API and worker startup
- [x] Health check endpoint
  - Worker /health endpoint on port 8081 (configurable via HEALTH_PORT)
  - Checks DB and Redis connectivity
  - Ready for ECS, Kubernetes, and orchestration platforms

#### Week 1: Sources + Media DB (Mon-Fri)
- [ ] Database migrations
  - [ ] Create `sources` table (id, project_id, type, url, crawl_mode, last_crawled_at, status)
  - [ ] Create `document_sources` mapping table (document_id, source_id)
  - [ ] Create `media_items` table (id, type: image|video|youtube, source_url)
  - [ ] Create `extracted_text` table (media_id, text, source_type: ocr|transcript, confidence_score)
  - [ ] Run migrations, verify tables exist
- [ ] Update ingest handler to record `source_id` for each document
- [ ] Test: POST /v1/ingest → documents tagged with source

#### Week 2: Media Ingest Handler (Mon-Fri)
- [ ] OCR Integration (Google Vision)
  - [ ] Detect images in ingest payload → call Google Vision API
  - [ ] Store extracted text in `extracted_text` table
  - [ ] Error handling: skip image, log
- [ ] YouTube Transcript Fetching
  - [ ] Detect YouTube URLs → fetch captions via YouTube API
  - [ ] Store transcripts with timestamps in `extracted_text`
- [ ] Worker: `handleMediaIngest()` function
  - [ ] Process media_items in parallel (semaphore)
  - [ ] Update job progress: processed/total
- [ ] Test: Ingest image + YouTube URL → extracted text appears

#### Week 3: Extension Endpoint + DOM Parsing (Mon-Fri)
- [ ] API Types
  - [ ] `ExtensionChatRequest` {project_id, dom, screenshot, url, question}
  - [ ] `ExtensionChatResponse` {guidance, steps, next_actions}
  - [ ] `DOMEntity` {selector, type, text}
- [ ] DOM Parser
  - [ ] Parse DOM JSON → extract buttons, inputs, links
  - [ ] Generate CSS selectors for each element
- [ ] Handler: `POST /v1/extension/chat`
  - [ ] Extract DOM entities + retrieve relevant docs
  - [ ] LLM: "User on [page], asking [q]. Elements: [entities]. Guidance:"
  - [ ] Return: {guidance, steps: [{description, selector, action}]}
- [ ] Test: Send Mixpanel DOM → get steps with selectors

#### Week 4: Browser Extension (Mon-Fri)
- [ ] Extension code (TypeScript + React)
  - [ ] manifest.json (Chrome v3)
  - [ ] popup/App.tsx (input + results)
  - [ ] content-script.ts (captureDOM + screenshot)
  - [ ] utils/api.ts, storage.ts
- [ ] Core features
  - [ ] Capture DOM → JSON
  - [ ] Take screenshot (html2canvas)
  - [ ] Send to /v1/extension/chat
  - [ ] Display guidance in popup
- [ ] Testing
  - [ ] Load extension on Chrome (dev mode)
  - [ ] Test on Mixpanel + Stripe dashboards
  - [ ] Demo: Ask question → get steps

#### MVP Done
- [ ] Extension loads in Chrome
- [ ] Captures DOM + screenshot
- [ ] API returns guidance
- [ ] Demo works end-to-end
- [ ] Media ingestion (OCR/transcripts)
  - **Image OCR Processing**
    - Optical Character Recognition (OCR) for images
    - Support multiple providers:
      - Google Cloud Vision API (recommended, highest accuracy)
      - AWS Textract (enterprise option)
      - Tesseract (open-source fallback)
    - Image format support: PNG, JPEG, WebP, GIF (first frame)
    - Text extraction with confidence scores
    - Handle multi-page documents (PDF → image conversion)
    - OCR caching to avoid re-processing same images
    - Metadata extraction (creation date, dimensions, EXIF)
  - **YouTube Transcript Fetching**
    - Auto-detect video captions/subtitles
    - Fetch captions via YouTube API or third-party service
    - Support multiple languages (with automatic fallback)
    - Chunk transcripts by speaker/timestamp
    - Link chunks back to specific video timestamps
    - Handle auto-generated vs manually created captions
    - Fallback to manual transcription service if unavailable
  - **Video Audio Extraction & ASR**
    - Extract audio tracks from video files (MP4, WebM, MOV, AVI)
    - Speech-to-Text (ASR) using:
      - Google Cloud Speech-to-Text (recommended)
      - AWS Transcribe
      - Deepgram API (real-time capable)
      - Whisper (open-source fallback)
    - Speaker diarization (identify who is speaking)
    - Timestamp-based chunking (chunk by paragraph/silence)
    - Punctuation and capitalization correction
    - Audio quality detection and normalization
  - **Storage & Linkage**
    - Add `media_items` table (type: image|video|youtube, source_url, media_type)
    - Add `extracted_text` table (media_id, text, source_type: ocr|transcript|asr, confidence_score, created_at)
    - Modify `documents` table to support media_id reference
    - Create `document_media` mapping for multi-modal documents
    - Store extraction metadata (duration, language, quality_score)
  - **Pipeline Integration**
    - New ingest source type: `media` with MediaSpec configuration
    - Worker extension: `handleMediaIngest()` function
    - Parallel processing of media files (semaphore-limited)
    - Error handling: fallback providers, partial failures
    - Polling for long-running ASR jobs (async transcription)
  - **Configuration & Env Vars**
    - `MEDIA_OCR_PROVIDER` (google|aws|tesseract)
    - `MEDIA_OCR_API_KEY`
    - `YOUTUBE_API_KEY`
    - `MEDIA_ASR_PROVIDER` (google|aws|deepgram|whisper)
    - `MEDIA_ASR_API_KEY`
    - `MEDIA_STORAGE_PATH` (local/s3/gcs for temp files)
    - `MEDIA_MAX_FILE_SIZE_MB` (default: 500MB)
    - `MEDIA_EXTRACTION_TIMEOUT_SECONDS` (default: 600s)
  - **API Endpoint**
    - `POST /v1/ingest` with `source.type: "media"`
    - Example payload:
      ```json
      {
        "project_id": "project-uuid",
        "source": {
          "type": "media",
          "media": {
            "urls": ["https://example.com/video.mp4", "https://example.com/image.png"],
            "extract_ocr": true,
            "extract_asr": true,
            "extract_youtube": false,
            "ocr_provider": "google",
            "asr_provider": "google"
          }
        }
      }
      ```
  - **Testing Strategy**
    - Unit tests for OCR/ASR provider abstraction
    - Integration tests with mock providers
    - E2E tests with real sample media files
    - Performance benchmarks for different media types
    - Fallback provider testing
  - **Observability**
    - Log extraction quality metrics (confidence, duration)
    - Track extraction time per media type
    - Monitor API quota usage for cloud providers
    - Alert on extraction failures/timeouts
    - Metrics: files_processed, extraction_success_rate, avg_extraction_time
- [ ] Phase 0 completion tasks
  - Deflection endpoint refinement
  - Analytics endpoint implementation
  - Integration tests for end-to-end pipeline
  - README with deployment examples

---

### Phase 1: Advanced Deflection
Auto-suggest responses, webhook callbacks, deflection scoring, feedback loops, and ticketing integrations.

#### Tasks
- [ ] Auto-suggest responses
  - Implement response suggestion based on similar resolved questions
  - Ranking by relevance and feedback score
- [ ] Webhook callbacks
  - Enable webhooks for deflection events (resolved, escalated, user feedback)
  - Webhook retry logic and delivery tracking
- [ ] Deflection scoring
  - Confidence score for deflection suggestions
  - User satisfaction tracking
- [ ] Feedback loops
  - Capture user feedback (helpful/not helpful)
  - Update rankings and suggestions based on feedback
- [ ] Ticketing integrations
  - Jira integration for escalations
  - Zendesk integration for ticket creation
  - Custom webhook for internal systems

---

### Phase 2: Internal Assistant
Slack/Teams bots, auth, context routing, thread continuity, and slash commands.

#### Tasks
- [ ] Slack bot integration
  - OAuth setup and token management
  - Message handling and threading
  - Slash command handlers (/search, /ask, /escalate, etc.)
- [ ] Microsoft Teams bot integration
  - Bot registration and authentication
  - Adaptive card responses
  - Teams-specific formatting
- [ ] Authentication & authorization
  - User identity mapping (Slack user → internal user)
  - Role-based access control (RBAC) for bot commands
- [ ] Thread continuity
  - Maintain conversation context across messages
  - User session tracking within channels
- [ ] Conversation state management
  - Store conversation history in DB
  - Enable context-aware follow-up questions

---

### Phase 3: Coverage Analytics
Gap detection pipeline, trending topics, unresolved clusters, dashboards and export endpoints.

#### Tasks
- [ ] Gap detection pipeline
  - Identify questions not covered by knowledge base
  - ML-based clustering of unresolved queries
  - Automated gap reporting
- [ ] Trending topics
  - Track most common questions
  - Identify emerging support needs
  - Trend visualization
- [ ] Unresolved clusters
  - Group similar unresolved questions
  - Priority scoring by frequency
  - Recommendation for knowledge base expansion
- [ ] Analytics dashboards
  - Real-time metrics (questions/hour, deflection rate, etc.)
  - Historical trends (daily/weekly/monthly)
  - Export endpoints (CSV, JSON)
- [ ] Reporting endpoints
  - GET `/v1/analytics/gaps` - gap analysis
  - GET `/v1/analytics/trending` - trending topics
  - GET `/v1/analytics/report` - comprehensive report with export options

---

### Phase 4: Fine-tuning + Reranking
Custom domain embeddings, cross-encoder reranker, A/B testing and offline eval harness.

#### Tasks
- [ ] Custom domain embeddings
  - Fine-tune embedding model on company-specific data
  - Improve semantic search relevance for domain terms
  - Embedding model versioning and rollback
- [ ] Cross-encoder reranker
  - Implement cross-encoder for reranking search results
  - Fine-tune on company QA pairs
  - A/B test reranker vs base ranking
- [ ] A/B testing framework
  - Traffic splitting for experiments
  - Metrics collection and analysis
  - Experiment lifecycle management
- [ ] Offline evaluation harness
  - Benchmark suite for search quality
  - Automated testing on new model versions
  - Performance regression detection

---

### Phase 5: Scalability & Performance
Multi-region ready infra, async caching, vector index tuning (IVFFLAT/HNSW), background maintenance jobs.

#### Tasks
- [ ] Multi-region architecture
  - Database replication strategy
  - Redis cluster setup
  - Meilisearch index synchronization
- [ ] Async task queue
  - Long-running tasks (bulk ingestion, fine-tuning)
  - Task prioritization
  - Retry and dead-letter handling
- [ ] Caching layer
  - Redis-backed caching for search results
  - Cache invalidation strategies
  - TTL management
- [ ] Vector index optimization
  - Switch from SEQUENTIAL to IVFFLAT or HNSW for pgvector
  - Index tuning and benchmarking
  - Dimension optimization
- [ ] Background maintenance jobs
  - Index maintenance and rebuilding
  - Stale data cleanup
  - Analytics aggregation jobs
- [ ] Query performance monitoring
  - Slow query logging
  - Query plan analysis
  - Performance SLO tracking

---

### Phase 6: Observability & Tracing
OpenTelemetry tracing, structured logs, metrics, SLOs, and tracing across API/Worker.

#### Tasks
- [ ] OpenTelemetry integration
  - Tracing instrumentation for API endpoints
  - Worker task tracing
  - External service call tracing (DB, Redis, LLM APIs)
- [ ] Distributed tracing
  - Trace correlation across API and worker processes
  - Trace export to Jaeger/Tempo/DataDog
- [ ] Metrics collection
  - Request latency, error rates, throughput
  - Business metrics (deflection rate, questions/hour, etc.)
  - Resource metrics (CPU, memory, DB connections)
- [ ] Logging infrastructure
  - Centralized log aggregation (ELK, Loki, etc.)
  - Log levels and sampling
  - Structured log parsing
- [ ] SLO tracking
  - Define service SLOs (availability, latency, error rate)
  - SLO dashboards
  - SLO-based alerting
- [ ] Alerting
  - Alert rules for SLO breaches
  - Alert routing and escalation
  - Alert fatigue management

---

### Phase 7: Enterprise Features
SSO (SAML/OIDC), RBAC, audit logs, data retention controls, org/project boundaries.

#### Tasks
- [ ] SSO integration
  - SAML 2.0 support
  - OIDC support (Okta, Azure AD, etc.)
  - JIT user provisioning
- [ ] Role-based access control (RBAC)
  - Role definitions (admin, editor, viewer)
  - Permission matrix
  - Fine-grained resource permissions
- [ ] Audit logging
  - Log all user actions (create, read, update, delete)
  - Audit log retention and archival
  - Audit log export and compliance reporting
- [ ] Data retention policies
  - Configurable retention periods by data type
  - GDPR right-to-be-forgotten support
  - Data anonymization options
- [ ] Organization & project management
  - Multi-tenant support
  - Organization hierarchy
  - Project-level access controls
  - Service accounts for integrations
- [ ] Compliance
  - SOC 2 readiness
  - Data encryption at rest and in transit
  - Compliance reporting

---

### Phase 8: MCP Integration
Model Context Protocol server and IDE/desktop integrations; commands for search/ingest from clients.

#### Tasks
- [ ] MCP server implementation
  - Implement Model Context Protocol server
  - Resources (knowledge base, documents)
  - Tools (search, ask, escalate, ingest)
- [ ] MCP client integrations
  - VS Code extension for IDE integration
  - Desktop CLI tool
  - API client library
- [ ] IDE features
  - Code documentation search
  - Inline help and suggestions
  - Documentation panel
- [ ] Desktop CLI
  - Search command: `cgap search <query>`
  - Ask command: `cgap ask <question>`
  - Ingest command: `cgap ingest <url>`
  - Config management
- [ ] Client library
  - Go/Python/JavaScript SDK
  - Authentication and session management
  - Type-safe API bindings

---

## Technical Debt & Maintenance

- [ ] Dependency updates
  - Regular Go package updates
  - Security vulnerability scanning
  - Deprecation management
- [ ] Code refactoring
  - Reduce code duplication in crawler logic
  - Improve error handling consistency
  - Package organization review
- [ ] Test coverage
  - Increase unit test coverage to >80%
  - Add integration tests for critical paths
  - E2E test scenarios
- [ ] Documentation
  - Architecture decision records (ADRs)
  - Deployment runbooks
  - API documentation improvements
  - Developer guide

---

## Known Issues & Improvements

- [ ] Search result pagination
  - Implement cursor-based pagination for large result sets
  - Consistency across pgvector and Meilisearch backends
- [ ] Error messages
  - Improve user-facing error messages
  - Better error categorization and codes
- [ ] Configuration management
  - Centralized config service
  - Feature flags for gradual rollouts
  - Config versioning and rollback
- [ ] Resource cleanup
  - Implement document/chunk deletion with cascading
  - Periodic cleanup of abandoned jobs
  - Redis key expiration policies

---

## Priority Matrix

### High Priority (Next Quarter)
1. Persist sources and linkage (Phase 0)
2. Media ingestion (Phase 0)
3. Advanced deflection (Phase 1)
4. Analytics & gaps (Phase 3)

### Medium Priority
1. Internal assistant integrations (Phase 2)
2. Fine-tuning & reranking (Phase 4)
3. Observability & tracing (Phase 6)

### Low Priority (Later)
1. Enterprise features (Phase 7)
2. MCP integration (Phase 8)
3. Multi-region scalability (Phase 5)

---

## Definition of Done Checklist

For each feature:
- [ ] Code review approved
- [ ] Unit tests written and passing (>80% coverage)
- [ ] Integration tests added
- [ ] Documentation updated (README, OpenAPI, code comments)
- [ ] Manual testing completed
- [ ] Performance benchmarks run
- [ ] Security review completed
- [ ] Merged to main branch
- [ ] Deployed to staging
- [ ] Production deployment planned

---

## Contact & Questions

For questions about this roadmap or to request prioritization changes, please contact the CGAP team.

Last Updated: December 18, 2025

---

## Epics Overview

- **Ask AI (Widget + API)**: Public chat/search, streaming, citations, threads.
- **Ticket Deflector**: Form interception, suggestions, deflection metrics.
- **Internal Assistant**: Auth, RBAC, private KB, context routing, file attach.
- **Coverage Gaps**: Uncertainty pipeline, clustering, recommendations, export.
- **Media Ingestion**: OCR, transcripts, ASR end-to-end with provenance.
- **Observability**: Logs, metrics, tracing, SLOs, alerts.
- **Scalability**: Multi-region, caching, index tuning, background jobs.
- **Enterprise**: SSO/OIDC, audit logs, retention, org/project boundaries.
- **MCP Integration**: Server + clients for IDE/desktop integrations.

---

## Backlog: Small Tasks (Actionable)

### API
- [ ] Add `X-Request-ID` propagation across handlers.
- [ ] Centralize error responses: `{code,message,details}?` helper.
- [ ] Input validation layer (DTO validation, limits, required fields).
- [ ] Pagination: cursor-based for search/results with `next_cursor`.
- [ ] Rate limiting middleware (per IP and per API key).
- [ ] SSE streaming endpoints (`/v1/chat/stream`, follow-ups).
- [ ] OpenAPI coverage: ensure all endpoints documented with examples.
- [ ] Consistent JSON field casing and null handling.
- [ ] CORS configuration (allowlist + credentials toggle).
- [ ] Health endpoints: split `readiness` vs `liveness`.

### Worker
- [ ] Retry policy with exponential backoff for transient errors.
- [ ] Dead-letter queue for permanently failed jobs.
- [ ] Task timeout + cancellation propagation to HTTP client.
- [ ] Concurrency limits per source with semaphore + watchdog.
- [ ] Memory/CPU safeguard (large pages, oversized media).
- [ ] Structured job logs with job_id, project_id, source_id.
- [ ] Prometheus metrics: jobs_running, processed, failed, duration.
- [ ] Graceful drain on SIGTERM (finish current, stop accepting new).

### Ingestion & Crawler
- [ ] Respect `User-Agent` config for robots and fetches.
- [ ] 429/backoff handling; jittered retries on rate limits.
- [ ] Content-type aware parsing (HTML/Markdown/Text/PDF stub).
- [ ] Canonical URL dedup (normalize trailing slash, query params policy).
- [ ] Sitemapindex limits configurable; log truncation reason.
- [ ] Chunking strategy: sentence/paragraph boundaries, token-aware.
- [ ] Max chunks per doc configurable; default sane caps.
- [ ] Allow/Deny patterns precompile; invalid regex warning path.
- [ ] Per-source politeness (`delay_ms`, `concurrency`) validation.
- [ ] Optional `noindex` meta detection to skip pages.

### Storage / Database
- [ ] Create `sources` and `document_sources` tables + DDL.
- [ ] Add indexes: `documents(project_id, uri)` (exists), `chunks(document_id, ord)` (exists) review.
- [ ] Add foreign keys and ON DELETE CASCADE where appropriate.
- [ ] Vector index config via env: `IVFFLAT lists/probes` defaults.
- [ ] Migration files with `goose`; add seed data for dev.
- [ ] Query performance: add EXPLAIN docs for heavy paths.

### Search
- [ ] Hybrid fusion: tune weights, tie-breaks, and normalization.
- [ ] Meilisearch synonyms/stop-words per project; bootstrap script.
- [ ] Snippet highlighting and positions for UI.
- [ ] Filters: support `source_type`, `document_uri` consistently.
- [ ] Fallback if Meilisearch unavailable → pgvector-only mode.
- [ ] Consistent response schema across providers.

### Embeddings & LLM
- [ ] Alignment: enforce `EMBEDDING_DIMENSION` across providers.
- [ ] Timeouts and retries for embedding/LLM calls.
- [ ] Simple caching for recent embeddings (Redis TTL).
- [ ] Rate limit & cost logging per provider.
- [ ] HTTP embedder: auth header + endpoint health check.
- [ ] Safe text preprocessing (unicode normalization, trimming).

### Observability
- [ ] Configure `slog` JSON/text via `LOG_FORMAT` env.
- [ ] Log level control via `LOG_LEVEL` (debug/info/warn/error).
- [ ] Correlation IDs across API ↔ worker ↔ DB ↔ Redis.
- [ ] Prometheus metrics endpoint (`/metrics`) for API + worker.
- [ ] OpenTelemetry stubs; trace IDs in logs.
- [ ] pprof endpoints for performance diagnostics (dev-only).

### Configuration
- [ ] `.env.example` aligned with README variable names.
- [ ] Central config loader with validation and defaults.
- [ ] Secret handling guidance (env, Docker secrets).
- [ ] Document feature flags for gradual rollouts.

### CI/CD
- [ ] GitHub Actions: build, test, lint, vet workflows.
- [ ] Docker image build/push with tags (`api`, `worker`).
- [ ] Vulnerability scan (Trivy/GH CodeQL) gates.
- [ ] Caching for `go mod` and test runs.
- [ ] Release notes automation (CHANGELOG generation).

### Documentation
- [ ] README: exact env var names and sample `.env`.
- [ ] API examples: full request/response bodies with errors.
- [ ] Ingest scenarios: add media examples (OCR/ASR/YouTube).
- [ ] Deployment runbooks (Docker Compose, ECS, K8s).
- [ ] Architecture diagrams kept in `docs/` with updates.

### Security
- [ ] API keys: creation, hashing, expiration policies.
- [ ] Input sanitization; HTML stripping safety for ingest.
- [ ] CORS + TLS guidance; reverse proxy notes.
- [ ] Rate limiting defaults + overrides per project.
- [ ] Audit log scaffolding (who/what/when).

### Testing
- [ ] Unit tests for API handlers and services.
- [ ] Integration tests: Postgres + pgvector + Meilisearch.
- [ ] Worker E2E ingest tests with mock HTTP server.
- [ ] Mock providers for embeddings/LLM/ASR/OCR.
- [ ] Benchmark: chunking and search performance.

### Deployment
- [ ] Docker Compose: health checks for API/worker (readiness/liveness).
- [ ] ECS/Kubernetes manifests (Deployment, Service, HPA).
- [ ] Resource limits/requests; autoscaling guidance.
- [ ] Logs shipping (Loki/ELK) with labels/fields.
- [ ] Secrets management (AWS/GCP secret managers).

---

## Acceptance Criteria Notes (per area)
- **API**: All endpoints documented in OpenAPI with examples; errors standardized.
- **Worker**: Jobs retry on transient failures; DLQ captures unrecoverables.
- **Ingestion**: Crawl respects scope/robots; chunking token-aware; dedup canonicalized.
- **DB**: Migrations reproducible; indexes present; foreign keys enforced.
- **Search**: Hybrid returns consistent schema; fallback operational.
- **Embeddings**: Dimension mismatch prevented; caching reduces redundant calls.
- **Observability**: Metrics and logs present; health endpoints separated.
- **Config**: `.env.example` and validation prevent misconfig.
- **CI/CD**: Pipeline green for build, test, lint; images published.
- **Docs**: Quick start reproducible; runbooks actionable.
- **Security**: Keys managed; rate limits enforced; sanitized inputs.
- **Testing**: Unit + integration + E2E coverage; mocks available.
- **Deployment**: Health checks, autoscaling, secrets wired; logs ship correctly.
