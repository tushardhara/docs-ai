# CGAP MVP - What's Built vs What's Needed

## ✅ Already Built (Reuse)

### Backend Infrastructure
- [x] Go API server (`cmd/api/main.go`)
- [x] Worker server (`cmd/worker/main.go`)
- [x] PostgreSQL storage layer (`internal/postgres/`)
- [x] Meilisearch integration (`internal/meilisearch/`)
- [x] Redis queue (`internal/queue/`)
- [x] slog structured logging
- [x] Health endpoints (`/health`)
- [x] Docker Compose setup

### Ingest & Processing
- [x] Ingest job status tracking (Redis)
- [x] Crawler: single/sitemap/crawl modes
- [x] robots.txt parser
- [x] URL deduplication
- [x] Chunking + tokenization
- [x] Embedding providers: OpenAI, Google, Anthropic, XAI
- [x] Job progress: queued → running → completed/failed

### Search & Retrieval
- [x] Hybrid search: pgvector + Meilisearch
- [x] Semantic search (embeddings)
- [x] Full-text search (Meilisearch)
- [x] Citation tracking
- [x] Relevance scoring

### LLM Integration
- [x] OpenAI client
- [x] Google Gemini client
- [x] Anthropic client
- [x] Streaming support
- [x] Context building from search results

### API Handlers
- [x] POST /v1/chat (Q&A)
- [x] POST /v1/search (hybrid search)
- [x] POST /v1/ingest (crawl + embed)
- [x] GET /v1/ingest/{job_id} (status)
- [x] POST /v1/deflect/suggest (deflection MVP)

---

## 🔨 Needs to Be Built (4-Week Sprint)

### Week 1: Database Schema

```
Missing Tables:
├─ sources (track where docs came from)
├─ document_sources (mapping)
├─ media_items (OCR images, YouTube, etc.)
├─ extracted_text (OCR text, transcripts)
└─ Migrations (goose format)

Files to Create:
├─ db/migrations/001_add_sources.sql
├─ db/migrations/002_add_media.sql
└─ db/migrations/003_add_document_sources.sql

Files to Update:
├─ cmd/worker/main.go (record source_id in ingest)
└─ api/handlers.go (update ingest handler)
```

### Week 2: Media Ingest

```
Missing Handlers:
├─ OCR processor (Google Vision)
├─ YouTube transcript fetcher
├─ ASR (Audio Speech Recognition) - optional Phase 2
└─ Worker handleMediaIngest() function

Files to Create:
├─ internal/media/ocr.go
├─ internal/media/youtube.go
└─ internal/media/handler.go

Files to Update:
├─ cmd/worker/main.go (call handleMediaIngest)
├─ api/types.go (add media ingest types)
└─ api/handlers.go (update ingest handler)

Environment Variables:
├─ GOOGLE_VISION_API_KEY
└─ YOUTUBE_API_KEY
```

### Week 3: Extension Chat Endpoint

```
Missing Endpoint:
├─ POST /v1/extension/chat (new)
└─ DOM parser

Files to Create:
├─ internal/extension/dom_parser.go
├─ internal/extension/handler.go
└─ api/extension_types.go

Files to Update:
├─ api/types.go (add ExtensionChatRequest/Response)
├─ api/handlers.go (add HandleExtensionChat)
├─ cmd/api/main.go (register route)
└─ cmd/api/main.go (inject services)

Key Features:
├─ Parse DOM JSON → extract buttons, inputs, links
├─ Generate CSS selectors
├─ Call search service (hybrid)
├─ LLM prompt engineering
└─ Return: {guidance, steps: [{description, selector, action}]}
```

### Week 4: Browser Extension

```
Missing Extension:
├─ manifest.json (Chrome v3)
├─ React popup UI
├─ Content script (DOM capture)
├─ Extension API client
└─ Storage (auth token)

Files to Create:
├─ extension/manifest.json
├─ extension/src/popup/App.tsx
├─ extension/src/popup/styles.css
├─ extension/src/content/script.ts
├─ extension/src/utils/api.ts
├─ extension/src/utils/storage.ts
├─ extension/package.json
├─ extension/tsconfig.json
├─ extension/webpack.config.js (or vite.config.js)
└─ extension/.env.example

Build Process:
├─ npm install
├─ npm run build → extension/dist/
└─ Load in Chrome: chrome://extensions/ → Load unpacked

Key Features:
├─ captureDOM() → serialize to JSON
├─ takeScreenshot() → html2canvas
├─ sendToAPI() → POST /v1/extension/chat
├─ displayResults() → popup
└─ Optional: autoClick() with user confirm
```

---

## 📋 Dependency Chain (Do in Order)

```
Week 1 (DB) → Week 2 (Media Ingest) → Week 3 (Extension Endpoint) → Week 4 (Extension UI)

Week 1 is BLOCKING:
  Can't ingest media without media_items + extracted_text tables

Week 2 is BLOCKING Week 3:
  Extension needs source knowledge (OCR text, transcripts) to give better guidance

Week 3 is BLOCKING Week 4:
  Extension needs API endpoint to send DOM + get guidance

Week 4 is NOT BLOCKING others:
  But it's the demo / customer-facing piece
```

---

## 🎯 Minimum Viable Features (Keep MVP Small)

### Week 1-2 (Keep It Simple)
- ✅ Store sources + media metadata
- ✅ OCR: Google Vision only (not AWS/Tesseract)
- ✅ YouTube: Auto-detect URLs + fetch transcripts (no speaker diarization)
- ❌ NO: ASR, PDF parsing, multi-page documents, confidence scoring improvements

### Week 3 (Keep It Simple)
- ✅ Extract buttons, inputs, links from DOM
- ✅ Generate CSS selectors
- ✅ Simple LLM prompt (no few-shot examples)
- ✅ Return steps with selectors
- ❌ NO: Auto-correct selectors, image annotation, confidence tuning

### Week 4 (Keep It Simple)
- ✅ Popup UI with input box
- ✅ Capture DOM + screenshot
- ✅ Send to API + display results
- ✅ Show steps as numbered list
- ❌ NO: Auto-click (Phase 2), animation, offline mode, caching

---

## 📊 Comparison: Old vs New MVP

| Aspect | Old Plan | New MVP | Change |
|--------|----------|---------|--------|
| **Timeline** | 8 weeks | 4 weeks | **2x faster** |
| **Scope** | Sources + Media + Deflection + Analytics | Sources + Media + Extension | **Focused** |
| **Phases** | 1-8 planned | Phase 0 only | **MVP-first** |
| **Entry point** | Q&A text | Browser extension | **More viral** |
| **Demo** | Curl requests | Live on SaaS dashboard | **Wow factor** |
| **First customer** | Unknown | Mixpanel/Stripe/HubSpot user | **Clear** |

---

## 🚀 After Week 4 (Phase 1+)

Once MVP is working:

### Phase 1 (Weeks 5-6)
- [ ] Auto-click execution (with user confirm)
- [ ] Deflection scoring refinement
- [ ] Analytics: gap detection
- [ ] Error handling hardening

### Phase 2 (Weeks 7-10)
- [ ] Slack bot integration
- [ ] Dashboard UI (admin view)
- [ ] Customer onboarding
- [ ] Launch: Chrome Web Store

### Phase 3 (Weeks 11+)
- [ ] Custom fine-tuning
- [ ] Multi-language support
- [ ] Enterprise customers
- [ ] Series A pitch

---

## ✨ Quick Wins (Low-effort, High-impact)

Do these first to build momentum:

1. **Week 1**: Database migrations (straightforward SQL)
2. **Week 2**: OCR handler (Google Vision API is simple)
3. **Week 3**: DOM parser (regex + CSS selector generation)
4. **Week 4**: Extension popup (React boilerplate)

Each one delivers value immediately (can test as you go).

---

## 🎓 Skills Gap (What You Might Need Help On)

| Area | Difficulty | How to Approach |
|------|-----------|---|
| **Browser extension** | Medium | Start with Chrome examples, use Manifest v3 template |
| **OCR integration** | Easy | Google Vision API has simple HTTP API |
| **YouTube API** | Easy | youtube-transcript-api npm package (no auth needed) |
| **DOM parsing** | Medium | JavaScript querySelector + regex for selectors |
| **TypeScript/React** | Medium | Use Create React App or Vite template |

All of these have good docs + examples. **Doable in 4 weeks.**

---

## 📞 Code Quality Goals (MVP)

- `go build ./...` → **PASS** (no errors)
- `go test ./...` → **PASS** (>70% coverage on critical paths)
- `npm run build` → **PASS** (no warnings)
- Manual testing: Mixpanel + Stripe → **PASS** (both work)

**Don't optimize.** Just ship working code.

---

## 🏁 Week 4 Friday: Launch Checklist

- [ ] `go build ./...` passes
- [ ] `npm run build` completes
- [ ] Extension loads in Chrome (dev mode)
- [ ] Can ask question on Mixpanel dashboard
- [ ] Gets guidance back (3+ steps)
- [ ] Steps have selectors
- [ ] Demo runs <5 minutes
- [ ] README updated with install instructions
- [ ] Code committed to `main` branch
- [ ] Ready to show investor / first customer

**If all checked: 🎉 MVP Complete. Ready for customer pilot.**
