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

### Week 3: Extension Endpoint (Mon-Fri)
```
Deliverable: API endpoint for browser extension
```
- [ ] API Types
  - [ ] `ExtensionChatRequest` {project_id, dom, screenshot, url, question}
  - [ ] `ExtensionChatResponse` {guidance, steps: [], next_actions: []}
  - [ ] `DOMEntity` {selector, type, text}
- [ ] DOM Parser
  - [ ] Parse DOM JSON → extract buttons, inputs, links
  - [ ] Generate CSS selectors
- [ ] Handler: `POST /v1/extension/chat`
  - [ ] Extract DOM entities
  - [ ] Hybrid search: retrieve relevant docs
  - [ ] LLM prompt: "User on [page], asking [q]. Elements: [entities]. Provide steps:"
  - [ ] Return: {guidance, steps: [{description, selector}]}
- [ ] Test: `curl POST /v1/extension/chat` with mock DOM → get steps

### Week 4: Browser Extension UI (Mon-Fri)
```
Deliverable: Working Chrome extension MVP
```
- [ ] Create extension/ directory structure
  ```
  extension/
  ├─ manifest.json
  ├─ src/
  │  ├─ popup/App.tsx
  │  ├─ content/script.ts
  │  └─ utils/api.ts
  └─ dist/ (built)
  ```
- [ ] manifest.json (Chrome v3)
- [ ] Popup UI: input box + loading state + results display
- [ ] Content script: captureDOM() + screenshot
- [ ] Core flow:
  1. Capture page DOM → JSON
  2. Take screenshot (html2canvas)
  3. Send to /v1/extension/chat
  4. Display guidance in popup
- [ ] Test on Mixpanel + Stripe dashboards
- [ ] Load in Chrome (dev mode)

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
