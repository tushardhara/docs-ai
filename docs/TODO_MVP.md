# CGAP - Browser Extension MVP (1-Month Sprint)

## 🎯 Mission
**Browser extension that captures page context (DOM + screenshot) and provides AI-guided setup steps.**

## Timeline: 4 Weeks to MVP

### Week 1: Sources + Media DB (Mon-Fri)
```
Deliverable: Database schema ready
```
- [ ] Database migrations
  - [ ] Create `sources` table (id, project_id, type, url, crawl_mode, last_crawled_at, status)
  - [ ] Create `document_sources` table (document_id, source_id)
  - [ ] Create `media_items` table (id, type: image|video|youtube, source_url)
  - [ ] Create `extracted_text` table (media_id, text, source_type: ocr|transcript, confidence_score)
- [ ] Update ingest handler: record source_id for each document
- [ ] Test: `go build ./...` passes

### Week 2: Media Ingest Handler (Mon-Fri)
```
Deliverable: Ingest images + YouTube transcripts
```
- [ ] OCR Integration (Google Vision)
  - [ ] Detect images in ingest payload
  - [ ] Call Google Vision API → extract text
  - [ ] Store in extracted_text table
- [ ] YouTube Transcript Fetcher
  - [ ] Detect YouTube URLs
  - [ ] Fetch captions via youtube-transcript-api
  - [ ] Store with timestamps
- [ ] Worker: `handleMediaIngest()` 
  - [ ] Process media in parallel
  - [ ] Update job progress
- [ ] Test: `POST /v1/ingest` with image + YouTube URL → extracted text stored

### Week 3: Extension Endpoint (Mon-Fri) ✅ COMPLETE
```
Deliverable: API endpoint for browser extension
```
- [x] API Types
  - [x] `ExtensionChatRequest` {project_id, dom, screenshot, url, question}
  - [x] `ExtensionChatResponse` {guidance, steps: [], next_actions: []}
  - [x] `DOMEntity` {selector, type, text, id, class, aria_label}
- [x] DOM Parser
  - [x] Parse DOM JSON → extract buttons, inputs, links
  - [x] Generate CSS selectors
  - [x] Filter interactive elements
- [x] Handler: `POST /v1/extension/chat`
  - [x] Extract DOM entities
  - [x] Hybrid search: retrieve relevant docs (pgvector + Meilisearch)
  - [x] LLM integration: "User on [page], asking [q]. Elements: [entities]. Provide steps:"
  - [x] Return: {guidance, steps: [{description, selector, confidence}], next_actions: [], sources: []}
- [x] Test: `curl POST /v1/extension/chat` with mock DOM → get steps
- [x] Code quality: `go build ./...` passes
- [x] API registered and working

### Week 4: Browser Extension UI (Mon-Fri) ⏳ IN PROGRESS
```
Deliverable: Working Chrome extension MVP
```
- [ ] Create extension/ directory structure (Issue #3)
  ```
  extension/
  ├─ manifest.json
  ├─ package.json
  ├─ tsconfig.json
  ├─ webpack.config.js
  ├─ src/
  │  ├─ popup/
  │  │  ├─ App.tsx
  │  │  ├─ index.tsx
  │  │  └─ index.html
  │  ├─ content/
  │  │  ├─ capture.ts
  │  │  ├─ highlighter.ts
  │  │  └─ types.ts
  │  └─ utils/
  │     ├─ api.ts
  │     ├─ storage.ts
  │     ├─ types.ts
  │     └─ config.ts
  └─ dist/ (built)
  ```
## MVP Definition of Done
- [x] Backend API complete and tested
- [x] Media handlers (OCR, YouTube, Video) working
- [x] Extension chat endpoint fully implemented
- [ ] Extension loads in Chrome
- [ ] Can capture DOM + screenshot from any SaaS
- [ ] Extension sends data to API endpoint
- [ ] Extension displays guidance steps
- [ ] Demo script works end-to-end: Mixpanel → ask → get steps
- [x] Code: `go build ./...` passes (backend)
## Demo Script (Week 4 Friday Target)
```
1. Open Mixpanel in Chrome, log in
2. Click CGAP extension icon
3. Type: "How do I create a dashboard?"
4. Click "Analyze Page" button
5. See guidance:
   - "Step 1: Click Dashboards in sidebar" (with selector: .nav-dashboards)
   - "Step 2: Click New Dashboard" (with selector: .btn-new-dashboard)
   - "Step 3: Enter dashboard name" (with selector: input#name)
   - "Step 4: Click Save" (with selector: .btn-save)
6. Hover over steps to highlight elements on page
7. Click step to scroll to and highlight the element
```
---

## MVP Definition of Done
- [ ] Extension loads in Chrome
- [ ] Can capture DOM + screenshot from any SaaS
- [ ] API endpoint returns guidance steps
- [ ] Demo script works end-to-end: Mixpanel → ask → get steps
- [ ] Code: `go build ./...` passes
- [ ] README: "How to install extension" section added

---

## Demo Script (Week 4 Friday)
```
1. Open Mixpanel in Chrome, log in
2. Click CGAP extension icon
3. Type: "How do I create a dashboard?"
4. See guidance:
   - "1. Click Dashboards in sidebar"
   - "2. Click New Dashboard"
   - "3. Enter name"
   - "4. Click Save"
5. Optional: Auto-click each step (with confirm)
```

---

## Out of Scope (Post-MVP)
- [ ] Deflection scoring refinement
- [ ] Analytics dashboard
- [ ] Internal assistant / Slack bots
- [ ] Fine-tuning / reranking
- [ ] Multi-region / scalability
- [ ] Enterprise features (SSO, RBAC)
- [ ] MCP integration
- [ ] Auto-click without user confirm

---

## Tech Stack (MVP)
- Backend: Go, PostgreSQL + pgvector, Meilisearch, Redis
- Extension: TypeScript, React, Manifest v3
- APIs: Google Vision, YouTube, Gemini/OpenAI
- Deployment: Docker Compose (local), Chrome Web Store (extension)

---

## Success Metrics (Week 4)
- [ ] Extension installs cleanly
- [ ] No build errors
- [ ] API response <2s
- [ ] Guidance accurate for demo dashboard
- [ ] Can be used by non-technical person
