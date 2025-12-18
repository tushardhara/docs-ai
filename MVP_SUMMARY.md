# CGAP MVP - 1-Month Sprint Summary

## 🎯 Mission
**Build a browser extension that captures page DOM + screenshot and provides AI-guided setup steps grounded in customer docs + videos.**

## 📅 4-Week Timeline

### Week 1: Database Foundation
**Deliverable**: Schema ready for ingest + media

```
Tasks:
├─ Create sources table (id, project_id, type, url, crawl_mode, last_crawled_at, status)
├─ Create document_sources mapping table
├─ Create media_items table (id, type: image|video|youtube, source_url)
├─ Create extracted_text table (media_id, text, source_type, confidence_score)
├─ Run migrations with goose
└─ Update ingest handler to record source_id

Acceptance: 
  - go build ./... passes
  - SELECT * FROM sources returns empty table
  - POST /v1/ingest → documents tagged with source_id
```

### Week 2: Media Ingest
**Deliverable**: OCR + YouTube transcripts working

```
Tasks:
├─ Implement OCR handler (Google Vision API)
│  ├─ Detect images in ingest payload
│  ├─ Call API → extract text
│  └─ Store in extracted_text table
├─ Implement YouTube transcript fetcher
│  ├─ Detect YouTube URLs
│  ├─ Fetch captions
│  └─ Store with timestamps
├─ Update worker: handleMediaIngest()
│  ├─ Process media in parallel (semaphore)
│  └─ Update job progress
└─ Test: POST /v1/ingest with image + YouTube URL

Acceptance:
  - Ingest image → extracted_text table has text
  - Ingest YouTube URL → extracted_text has transcript
  - Job status shows media_items_processed
```

### Week 3: Extension Endpoint
**Deliverable**: API that takes DOM + screenshot + question, returns guidance

```
Tasks:
├─ Add ExtensionChatRequest type
├─ Add ExtensionChatResponse type
├─ DOM parser: JSON → entities (buttons, inputs, links)
├─ Handler: POST /v1/extension/chat
│  ├─ Extract DOM entities
│  ├─ Hybrid search: retrieve relevant docs
│  ├─ LLM prompt: "User on [page], asking [q]. Elements: [entities]. Steps:"
│  └─ Return: {guidance, steps: [{description, selector, action}]}
└─ Test: curl with Mixpanel DOM → get steps

Acceptance:
  - Endpoint returns steps with selectors
  - Steps are numbered + actionable
  - Confidence score included
  - Latency <1s
```

### Week 4: Browser Extension UI
**Deliverable**: Working Chrome extension, ready for demo

```
Tasks:
├─ Create extension/ directory structure
├─ manifest.json (Chrome v3)
├─ popup/App.tsx
│  ├─ Input box for questions
│  ├─ Loading state
│  └─ Display results
├─ content/script.ts
│  ├─ captureDOM() → JSON
│  ├─ Take screenshot (html2canvas)
│  └─ Extract URL + page_title
├─ utils/api.ts
│  ├─ Call /v1/extension/chat
│  └─ Handle auth
├─ Build: npm run build
├─ Load in Chrome (dev mode)
└─ Test on: Mixpanel, Stripe, HubSpot

Acceptance:
  - Extension loads in Chrome
  - Can capture any page's DOM + screenshot
  - Sends to API successfully
  - Displays guidance in popup
  - Demo: Mixpanel → "create dashboard" → see steps
```

---

## 📊 Success Criteria (Week 4 Friday)

- [ ] Extension installs and loads (no errors)
- [ ] Can ask question on Mixpanel dashboard
- [ ] Gets back guidance with 3+ steps
- [ ] Steps have CSS selectors
- [ ] Demo takes <5 minutes (live walkthrough)
- [ ] Code: `go build ./...` passes
- [ ] All tests passing
- [ ] README has "How to install extension" section

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend API** | Go + Fiber |
| **Database** | PostgreSQL 16 + pgvector |
| **Search** | Meilisearch + pgvector hybrid |
| **Cache/Queue** | Redis |
| **LLM** | Gemini or OpenAI |
| **Embeddings** | Gemini or OpenAI (768 dims) |
| **Extension** | TypeScript + React + Manifest v3 |
| **Media APIs** | Google Vision (OCR) + YouTube API (transcripts) |
| **Deployment** | Docker Compose (local) |

---

## 🚀 What NOT to Build (Out of Scope)

- ❌ Deflection scoring refinement
- ❌ Analytics dashboard
- ❌ Internal assistant / Slack bots
- ❌ Fine-tuning / reranking
- ❌ Multi-region / scalability
- ❌ Enterprise features (SSO, RBAC, audit logs)
- ❌ MCP integration
- ❌ Auto-click without user confirmation (Phase 2)
- ❌ Phases 1-8 (save for later)

---

## 📁 New Files to Create/Update

| File | Purpose | Status |
|------|---------|--------|
| `extension/manifest.json` | Chrome extension config | Create |
| `extension/src/popup/App.tsx` | Extension UI | Create |
| `extension/src/content/script.ts` | DOM capture + screenshot | Create |
| `api/types.go` | Add ExtensionChatRequest | Update |
| `api/handlers.go` | Add HandleExtensionChat | Update |
| `cmd/worker/main.go` | Add handleMediaIngest | Update |
| `internal/*/` | Add OCR + YouTube handlers | Create |
| `db/migrations/` | Create tables (sources, media_items, etc) | Create |
| `TODO_MVP.md` | 4-week sprint checklist | Create |
| `PITCH_MVP.md` | Updated pitch (extension-first) | Create |
| `README_MVP.md` | Simplified README for MVP | Create |

---

## 🎬 Demo Script (End of Week 4)

```
Setup (5 min before):
- Have Mixpanel dashboard open in Chrome
- Extension installed and loaded
- API running on :8080
- Docs ingested into system

Live Demo (3 min):
1. Open Mixpanel Dashboards page
2. Click CGAP extension icon
3. Type: "How do I create a dashboard for tracking user behavior?"
4. Watch as extension:
   - Captures DOM (showing in console)
   - Takes screenshot
   - Sends to API
   - Gets guidance back
5. Display result:
   "Step 1: Click 'New Dashboard' (selector: button.create-dash)
    Step 2: Enter name 'User Behavior'
    Step 3: Click 'Create'"
6. Show: "Optional auto-click" (but don't click, keep manual)
7. Q&A: "This works for any SaaS dashboard"
```

---

## 📈 Metrics to Track

| Metric | Target | How to Measure |
|--------|--------|---|
| **Build Pass Rate** | 100% | `go build ./...` |
| **API Latency** | <1s | curl + timer |
| **Guidance Accuracy** | >85% | Manual testing on 5 SaaS dashboards |
| **Extension Load Time** | <500ms | Chrome dev tools |
| **Demo Success** | 5/5 tasks | Complete demo without errors |

---

## 🎓 Knowledge Required

### Backend Developer
- Go fundamentals
- PostgreSQL + pgvector
- Meilisearch basics
- LLM API integration (OpenAI/Gemini)
- REST API design

### Extension Developer
- TypeScript + React
- Chrome Manifest v3
- DOM manipulation
- HTML2Canvas
- REST API calls

### DevOps
- Docker Compose
- PostgreSQL migrations
- Go build process

---

## 📞 Quick Reference

### Build & Run
```bash
# Full stack
docker-compose up

# API
go run cmd/api/main.go

# Worker
go run cmd/worker/main.go

# Extension
cd extension && npm run build

# Load in Chrome
chrome://extensions/ → Load unpacked → extension/dist/
```

### Test Endpoints
```bash
# Health
curl http://localhost:8080/health

# Ingest
curl -X POST http://localhost:8080/v1/ingest \
  -H "Content-Type: application/json" \
  -d '{"project_id":"demo","source":{"type":"crawl","start_url":"https://docs.mixpanel.com"}}'

# Extension Chat (Week 3+)
curl -X POST http://localhost:8080/v1/extension/chat \
  -H "Content-Type: application/json" \
  -d '{"project_id":"demo","dom":"{...}","screenshot":"data:...","url":"...","question":"..."}'
```

### Database
```bash
# Migrations
goose -dir db/migrations postgres "$DATABASE_URL" up

# Check tables
psql -c "\dt" $DATABASE_URL
```

---

## 🏁 Done = Ready for Customer Pilot

By Friday of Week 4, you'll have:
1. ✅ Working backend API (ingest + search + LLM)
2. ✅ Browser extension that captures context
3. ✅ API endpoint that returns guided steps
4. ✅ Demo that works end-to-end
5. ✅ Ready to share with first customer (Mixpanel user, Stripe merchant, HubSpot admin)

**Next phase**: Get customer feedback → iterate → launch → acquire more customers.
