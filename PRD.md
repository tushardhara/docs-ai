# cgap - Product Requirements Document

**Project**: cgap  
**Date**: December 17, 2025  
**Version**: 1.0

## Executive Summary

cgap is an open-source AI-powered documentation assistant platform designed to provide Kapa-like capabilities: Ask AI, Ticket Deflector, Internal Assistant, and Coverage Gap Detection. Built on Postgres + pgvector + Meilisearch, cgap enables organizations to deploy intelligent documentation solutions across their technical ecosystem.

### Core Capabilities
- **Ask AI**: Public-facing documentation assistant (widget, API, SDK)
- **Ticket Deflector**: Support form integration to reduce ticket volume
- **Internal Assistant**: Secure knowledge base for internal teams
- **Coverage Gap Detection**: Identify and prioritize documentation gaps

### Technology Stack
- **Backend**: Go (API & Worker services)
- **Database**: PostgreSQL with pgvector extension
- **Search**: Meilisearch (lexical search)
- **Cache/Queue**: Redis
- **Deployment**: Docker Compose

---

## 1. Competitive Analysis: Kapa Feature Matrix

### A) Ask AI (External)

* **Website widget** (+ customization) and **search mode inside widget** ([Kapa Docs][6])
* **Slack + Discord bots** ([Kapa Docs][7])
* **API + SDK builds** (chat endpoints power widget/slack/discord; streaming supported; thread-based follow-ups) ([Kapa Docs][3])
* Deploy across docs/product/community/support portal ([Kapa Docs][2])

### B) Ticket Deflector (External)

* “**one-line code**” on support form, deflect **20–40%** ([Kapa AI][8])

### C) Internal Technical Assistant (Internal)

* Internal chat, private KB, **Google sign-in**, project permissions ([Kapa Docs][9])
* Extra UX: **filter sources**, **share conversations**, **deep thinking mode**, **attach files** ([Kapa Docs][9])

### D) Analytics + Gap Detection

* Dashboard metrics: total questions, uncertain, reactions, unique users, support deflection, etc. ([Kapa Docs][4])
* Conversations list with filters (integration, labels, status, feedback, date…) ([Kapa Docs][10])
* **Coverage Gaps**: cluster *uncertain answers* → “Finding” + “Recommendation” ([Kapa Docs][5])
* They also market conversation labeling + question clustering + content gap analysis ([Kapa AI][11])

### E) MCP (nice-to-have parity)

* Kapa offers **MCP servers** so users can use the knowledge inside ChatGPT/Claude/Cursor etc. ([Kapa Docs][2])

---

## 2. Architecture Overview

### Services

1. **api** (Go): chat/orchestrator, streaming, threads, auth, analytics endpoints
2. **worker** (Go): crawl/connectors → chunk → embed → index
3. **postgres**: metadata + Q&A logs + **pgvector** embeddings
4. **meilisearch**: lexical index + search-mode
5. **redis** (optional): job queue (asynq) + rate limits

### Retrieval (hybrid)

* Meilisearch topK (lexical) + pgvector topK (dense) → merge + rerank(optional) → answer with citations.

---

## 3. Implementation Roadmap

### Phase 0: Foundation (Core Platform)

**Deliverables**

* `docker-compose.yml` (postgres+pgvector, meili, redis, api, worker)
* Project model: `projects`, `sources`, `documents`, `chunks`
* Base ingestion pipeline: docs crawl → extract → chunk → embed → index
* Base retrieval: hybrid merge (meili + pgvector)
* Base prompt contract: **citations required** (no citations => ask follow-up / “I don’t know”)

**Acceptance**

* `compose up` + add docs URL → queries return answer with citations.

---

### Phase 1: Ask AI - Website Widget + Search Mode

**Goal**: Deliver public-facing documentation assistant with embedded widget

**Deliverables**

* React widget SDK (`@cgap/widget`) + 1-line embed
* Two widget modes:

  * **Chat mode** (RAG answer)
  * **Search mode** (returns top doc sections) ([Kapa Docs][3])
* Streaming responses (SSE/WebSocket), because Kapa uses streamed endpoints for widget ([Kapa Docs][3])
* Threads: start conversation + continue in thread ([Kapa Docs][3])

**Acceptance**

* Widget embedded in docs, supports chat + search, streaming works, follow-ups maintain context.

---

### Phase 2: Public API + SDKs

**Goal**: Expose programmatic access to cgap capabilities

**Deliverables**

* API:

  * `POST /v1/chat` (new thread)
  * `POST /v1/threads/{id}/chat` (follow-up)
  * `POST /v1/chat/stream` + `/threads/{id}/chat/stream`
  * `POST /v1/search` (semantic/hybrid search)
* SDKs:

  * JS/TS SDK first (widget + custom builds)
  * Go SDK later

**Acceptance**

* Widget is implemented purely using public endpoints (like Kapa). ([Kapa Docs][3])

---

### Phase 3: Ticket Deflector

**Goal**: Reduce support ticket volume by 20-40% through intelligent deflection

**Deliverables**

* Deflector snippet: intercept subject/body typing → call suggestions endpoint
* API:

  * `POST /v1/deflect/suggest` → top 3 answers with citations
  * `POST /v1/deflect/event` (shown/clicked/solved/submitted)
* Admin: Support deflection metrics (aligns with dashboard section). ([Kapa Docs][4])

**Acceptance**

* Can integrate into any HTML form with one snippet; deflection events are measurable.

---

### Phase 4: Community Integrations (Slack + Discord)

**Goal**: Bring cgap into developer communities

**Deliverables**

* Slack bot:

  * DM + channel mode
  * "Global workspace" vs "single channel" config ([Kapa Docs][7])
* Discord bot:

  * respond to mentions + forum channels
  * "Global server" vs "single channel" config ([Kapa Docs][12])
* Same thread model: Slack threads / Discord threads map to cgap thread IDs

**Acceptance**

* Bots answer with citations and obey channel scoping.

---

### Phase 5: Analytics Dashboard

**Goal**: Provide visibility into usage patterns and effectiveness

**Deliverables**

* Conversations table (all Q&A across integrations)
* Filters: integration, date range, labels, status, feedback, text search ([Kapa Docs][10])
* Dashboard metrics:

  * total questions, uncertain, reactions, unique users, answered, deflection ([Kapa Docs][4])

**Acceptance**

* Admin can slice “what’s happening” by integration (widget vs Slack vs deflector).

---

### Phase 6: Coverage Gap Detection

**Goal**: Automatically identify and prioritize documentation gaps

**Deliverables**

1. **Uncertainty model**

   * Each response produces `is_uncertain` boolean (like Kapa’s API fields) ([Kapa Docs][13])
   * Heuristics (MVP): weak retrieval scores, low citation coverage, conflicting sources, LLM self-check
2. **Coverage Gaps pipeline**

   * For a time window (week/month/quarter), collect uncertain answers ([Kapa Docs][5])
   * Embed questions → cluster (k-means/HDBSCAN-style; can be Go impl)
   * For each cluster: generate:

     * **Finding** (what users asked + why it failed)
     * **Recommendation** (AI-suggested doc/product fixes) ([Kapa Docs][5])
3. **Coverage Gaps UI**

   * list clusters, volume, example questions, linked sources, export to GitHub issues

**Acceptance**

* Admin sees top recurring uncertain topics + actionable recommendations (human-reviewed). ([Kapa Docs][5])

---

### Phase 7: Internal Technical Assistant

**Goal**: Secure, permission-controlled internal knowledge base

**Deliverables**

* Internal chat UI + project selector
* Auth: Google OIDC first (like Kapa), later Okta/AzureAD ([Kapa Docs][9])
* RBAC + project permissions
* UX parity:

  * filter sources per question ([Kapa Docs][9])
  * share conversation links ([Kapa Docs][9])
  * deep thinking mode (more retrieval passes + broader context) ([Kapa Docs][9])
  * attach files (parse text and use as ephemeral context) ([Kapa Docs][9])

**Acceptance**

* Internal users can chat, restrict sources, share threads, and attach context files.

---

### Phase 8: MCP Server Integration (Optional)

**Goal**: Enable cgap knowledge access from AI coding assistants

**Deliverables**

* MCP server that exposes:

  * `search(project, query)`
  * `ask(project, query, thread_id?)`
  * `get_sources(project)`
* Auth: API key / OAuth for MCP clients

**Acceptance**

* Works with MCP-compatible clients and returns cited answers.

---

## 4. Data Source Integration Roadmap

**cgap will support the following sources across phased releases:**

* v0.1: docs crawl
* v0.2: GitHub Issues/Discussions + OpenAPI ingestion
* v0.4: Slack/Discord
* v0.7: Confluence/Notion/Drive/S3 + file uploads

---

## 5. Success Metrics

### Phase 0-2 (Foundation + Widget)
- System uptime: >99.5%
- Query response time: <2s (p95)
- Widget load time: <500ms
- Answer accuracy: >85% (human eval)

### Phase 3 (Deflector)
- Ticket deflection rate: 20-40%
- False positive rate: <5%
- User satisfaction: >4/5

### Phase 5-6 (Analytics + Gaps)
- Gap detection precision: >70%
- Actionable recommendations: >50% of detected gaps
- Time to identify coverage gap: <7 days

### Phase 7 (Internal)
- Internal adoption rate: >60% of technical staff
- Average queries per active user: >10/week
- Knowledge retrieval accuracy: >90%

---

## 6. Technical Specifications

### 6.1 Database (Postgres + pgvector)

**Conventions**: all tables include `id (uuid)`, `created_at`, `updated_at`, `project_id` (except `projects`, `users`). Embedding dimension: 1536 (IVFFLAT, lists=100, probes=10).

**Core data**
- `projects`: name, slug, `default_model`, `settings` (JSONB), `usage_plan`.
- `users`: email, name, `auth_provider`, `picture_url`.
- `project_members`: `project_id`, `user_id`, role (`owner|admin|editor|viewer`).
- `api_keys`: `project_id`, name, hashed key, `expires_at`, `scopes`.

**Content & retrieval**
- `sources`: type (`crawl|github|openapi|slack|discord|upload`), config JSONB, status.
- `documents`: `source_id`, `uri`, `title`, `lang`, `version`, `hash`, `published_at`.
- `chunks`: `document_id`, `ord`, `text`, `token_count`, `section_path`, `score_raw`.
- `chunk_embeddings`: `chunk_id`, `embedding vector(1536)`. Index: `ivfflat` on embedding, `btree` on `document_id`.
- `meili_records` (optional cache): `chunk_id`, flattened fields synced to Meilisearch.

**Conversations & answers**
- `threads`: `project_id`, `integration` (`widget|api|slack|discord|deflector|internal`), `external_ref` (channel/thread ids), `status`.
- `messages`: `thread_id`, `role` (`user|assistant|system`), `content`, `meta` (sources filter, mode), `latency_ms`.
- `answers`: `message_id`, `model`, `is_uncertain`, `reasoning_trace`, `prompt_version`.
- `citations`: `answer_id`, `chunk_id`, `score`, `quote`, `start_char`, `end_char`.
- `feedback`: `answer_id`, `type` (`thumbs_up|thumbs_down`), `comment`, `user_id`.

**Deflection & analytics**
- `deflect_events`: `project_id`, `session_id`, `subject`, `body`, `suggestion_ids`, `action` (`shown|clicked|solved|submitted`).
- `analytics_events`: `project_id`, `thread_id`, `message_id`, `type` (`question|answer|uncertain|reaction`), `properties` JSONB, `occurred_at` (ts with tz).

**Coverage gaps**
- `gap_candidates`: `answer_id`, `question_embedding`, `uncertainty_reason`.
- `gap_clusters`: `project_id`, `window` (`7d|30d|90d`), `label`, `summary`, `recommendation`, `size`, `status` (`open|in_review|done`).
- `gap_cluster_examples`: `cluster_id`, `answer_id`, `question`, `citations` (array), `representative_score`.

### 6.2 API Contracts (REST + SSE)

**Auth**: `Authorization: Bearer <api_key>` for external; Google OIDC for internal dashboard. Rate limits per project.

**Chat & Search**
- `POST /v1/chat`: `{ project_id, query, context_filters?, mode? (chat|search), top_k? }` → `{ thread_id, answer, citations[], is_uncertain, sources[] }`.
- `POST /v1/chat/stream`: same body, SSE frames `{ delta, citations?, is_uncertain? }`.
- `POST /v1/threads/{id}/chat`: follow-up in thread.
- `POST /v1/threads/{id}/chat/stream`: SSE follow-up.
- `POST /v1/search`: `{ project_id, query, top_k?, filters? }` → `{ hits: [{chunk_id, text, document_uri, score, source_type}], fusion_scores? }`.

**Deflector**
- `POST /v1/deflect/suggest`: `{ project_id, subject, body, top_k?=3 }` → `{ suggestions: [{answer, citations, score}], suggestion_id }`.
- `POST /v1/deflect/event`: `{ project_id, suggestion_id, action (shown|clicked|solved|submitted), thread_id?, metadata? }`.

**Sources & ingestion**
- `POST /v1/sources`: create source (crawl URL, GitHub repo, OpenAPI URL, etc.).
- `POST /v1/ingest`: upload file or trigger crawl; returns job id.
- `GET /v1/ingest/{job_id}`: job status and errors.

**Analytics**
- `GET /v1/analytics/summary`: query params `project_id`, `from`, `to`, `integration?` → totals (questions, uncertain, users, deflection, reactions).
- `GET /v1/analytics/conversations`: filters `integration, label, status, feedback, date_range, text` → paged list.

**Coverage gaps**
- `POST /v1/gaps/run`: trigger clustering for a window (`7d|30d|90d`).
- `GET /v1/gaps`: list clusters with counts and recommendations.
- `GET /v1/gaps/{id}`: details, example questions, linked sources; optional `export=github` to open an issue payload.

### 6.3 Meilisearch Configuration

- **Index name**: `cgap_chunks_{project_id}`.
- **Fields**: `id, project_id, document_uri, source_type, title, text, section_path, ord, score_raw`.
- **Searchable attributes**: `title`, `text`, `section_path`.
- **Filterable attributes**: `project_id`, `source_type`, `document_uri`.
- **Sortable attributes**: `score_raw`, `ord`.
- **Ranking rules** (override defaults):
  1. `typo`
  2. `words`
  3. `proximity`
  4. `attribute`
  5. `sort` (uses `score_raw` when provided)
  6. `exactness`
- **Synonyms/stop-words**: per-project configurable; store in `projects.settings` and sync at index creation.
- **Highlighting**: return `text` with match positions for UI snippets.

### 6.4 OpenAPI Skeleton (v0)

```yaml
openapi: 3.1.0
info:
  title: cgap API
  version: 0.1.0
servers:
  - url: https://api.cgap.dev
security:
  - apiKeyAuth: []
components:
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: Authorization
  schemas:
    Citation:
      type: object
      properties:
        chunk_id: { type: string }
        quote: { type: string }
        score: { type: number }
    ChatRequest:
      type: object
      required: [project_id, query]
      properties:
        project_id: { type: string }
        query: { type: string }
        mode: { type: string, enum: [chat, search] }
        context_filters: { type: object, additionalProperties: true }
    ChatResponse:
      type: object
      properties:
        thread_id: { type: string }
        answer: { type: string }
        is_uncertain: { type: boolean }
        citations:
          type: array
          items: { $ref: '#/components/schemas/Citation' }
    SearchHit:
      type: object
      properties:
        chunk_id: { type: string }
        text: { type: string }
        document_uri: { type: string }
        score: { type: number }
    DeflectSuggestion:
      type: object
      properties:
        answer: { type: string }
        citations:
          type: array
          items: { $ref: '#/components/schemas/Citation' }
        score: { type: number }
    AnalyticsSummary:
      type: object
      properties:
        total_questions: { type: integer }
        uncertain: { type: integer }
        unique_users: { type: integer }
        deflection_rate: { type: number }
    GapCluster:
      type: object
      properties:
        id: { type: string }
        window: { type: string }
        label: { type: string }
        recommendation: { type: string }
        size: { type: integer }

paths:
  /v1/chat:
    post:
      summary: Single-turn chat (new thread)
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/ChatRequest' }
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ChatResponse' }

  /v1/chat/stream:
    post:
      summary: Streaming chat (SSE)
      responses:
        '200':
          description: text/event-stream

  /v1/threads/{id}/chat:
    post:
      summary: Follow-up in thread

  /v1/threads/{id}/chat/stream:
    post:
      summary: Streaming follow-up (SSE)

  /v1/search:
    post:
      summary: Hybrid search
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  hits:
                    type: array
                    items: { $ref: '#/components/schemas/SearchHit' }

  /v1/deflect/suggest:
    post:
      summary: Ticket deflection suggestions
  /v1/deflect/event:
    post:
      summary: Track deflection outcomes

  /v1/sources:
    post:
      summary: Create a source (crawl, GitHub, OpenAPI, etc.)

  /v1/ingest:
    post:
      summary: Trigger ingest or file upload
  /v1/ingest/{job_id}:
    get:
      summary: Ingest job status

  /v1/analytics/summary:
    get:
      summary: Aggregate metrics
      parameters:
        - in: query
          name: project_id
          required: true
        - in: query
          name: from
        - in: query
          name: to
  /v1/analytics/conversations:
    get:
      summary: Paged conversation list with filters

  /v1/gaps/run:
    post:
      summary: Trigger gap clustering for a window
  /v1/gaps:
    get:
      summary: List clusters
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  clusters:
                    type: array
                    items: { $ref: '#/components/schemas/GapCluster' }
  /v1/gaps/{id}:
    get:
      summary: Cluster detail
```

### 6.5 Database DDL (initial cut)

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE projects (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  slug text UNIQUE NOT NULL,
  default_model text,
  settings jsonb DEFAULT '{}',
  usage_plan text,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email text UNIQUE NOT NULL,
  name text,
  auth_provider text,
  picture_url text,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

CREATE TABLE project_members (
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  user_id uuid REFERENCES users(id) ON DELETE CASCADE,
  role text CHECK (role IN ('owner','admin','editor','viewer')),
  created_at timestamptz DEFAULT now(),
  PRIMARY KEY (project_id, user_id)
);

CREATE TABLE api_keys (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  name text,
  key_hash text NOT NULL,
  scopes text[] DEFAULT '{chat,search}',
  expires_at timestamptz,
  created_at timestamptz DEFAULT now()
);

CREATE TABLE sources (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  type text CHECK (type IN ('crawl','github','openapi','slack','discord','upload')),
  config jsonb NOT NULL,
  status text DEFAULT 'pending',
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

CREATE TABLE documents (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  source_id uuid REFERENCES sources(id) ON DELETE SET NULL,
  uri text NOT NULL,
  title text,
  lang text,
  version text,
  hash text,
  published_at timestamptz,
  created_at timestamptz DEFAULT now()
);
CREATE UNIQUE INDEX documents_project_uri ON documents(project_id, uri);

CREATE TABLE chunks (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  document_id uuid REFERENCES documents(id) ON DELETE CASCADE,
  ord int NOT NULL,
  text text NOT NULL,
  token_count int,
  section_path text,
  score_raw real,
  created_at timestamptz DEFAULT now()
);
CREATE INDEX chunks_document_ord ON chunks(document_id, ord);

CREATE TABLE chunk_embeddings (
  chunk_id uuid PRIMARY KEY REFERENCES chunks(id) ON DELETE CASCADE,
  embedding vector(1536)
);
CREATE INDEX ON chunk_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE TABLE threads (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  integration text CHECK (integration IN ('widget','api','slack','discord','deflector','internal')),
  external_ref text,
  status text DEFAULT 'open',
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

CREATE TABLE messages (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  thread_id uuid REFERENCES threads(id) ON DELETE CASCADE,
  role text CHECK (role IN ('user','assistant','system')),
  content text NOT NULL,
  meta jsonb DEFAULT '{}',
  latency_ms int,
  created_at timestamptz DEFAULT now()
);
CREATE INDEX messages_thread_created ON messages(thread_id, created_at);

CREATE TABLE answers (
  message_id uuid PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
  model text,
  is_uncertain boolean,
  reasoning_trace jsonb,
  prompt_version text
);

CREATE TABLE citations (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  answer_id uuid REFERENCES answers(message_id) ON DELETE CASCADE,
  chunk_id uuid REFERENCES chunks(id) ON DELETE CASCADE,
  score real,
  quote text,
  start_char int,
  end_char int
);
CREATE INDEX citations_answer ON citations(answer_id);

CREATE TABLE feedback (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  answer_id uuid REFERENCES answers(message_id) ON DELETE CASCADE,
  type text CHECK (type IN ('thumbs_up','thumbs_down')),
  comment text,
  user_id uuid REFERENCES users(id),
  created_at timestamptz DEFAULT now()
);

CREATE TABLE deflect_events (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  session_id text,
  subject text,
  body text,
  suggestion_ids uuid[],
  action text CHECK (action IN ('shown','clicked','solved','submitted')),
  thread_id uuid REFERENCES threads(id),
  created_at timestamptz DEFAULT now()
);

CREATE TABLE analytics_events (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  thread_id uuid REFERENCES threads(id),
  message_id uuid REFERENCES messages(id),
  type text CHECK (type IN ('question','answer','uncertain','reaction')),
  properties jsonb DEFAULT '{}',
  occurred_at timestamptz DEFAULT now()
);
CREATE INDEX analytics_project_time ON analytics_events(project_id, occurred_at);

CREATE TABLE gap_candidates (
  answer_id uuid PRIMARY KEY REFERENCES answers(message_id) ON DELETE CASCADE,
  question_embedding vector(1536),
  uncertainty_reason text
);
CREATE INDEX ON gap_candidates USING ivfflat (question_embedding vector_cosine_ops) WITH (lists = 100);

CREATE TABLE gap_clusters (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
  window text CHECK (window IN ('7d','30d','90d')),
  label text,
  summary text,
  recommendation text,
  size int,
  status text CHECK (status IN ('open','in_review','done')) DEFAULT 'open',
  created_at timestamptz DEFAULT now()
);
CREATE INDEX gap_clusters_project_window ON gap_clusters(project_id, window);

CREATE TABLE gap_cluster_examples (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  cluster_id uuid REFERENCES gap_clusters(id) ON DELETE CASCADE,
  answer_id uuid REFERENCES answers(message_id) ON DELETE CASCADE,
  question text,
  citations uuid[],
  representative_score real
);
CREATE INDEX gap_examples_cluster ON gap_cluster_examples(cluster_id);
```

---

## References

This PRD is informed by analysis of Kapa.ai's public documentation and feature set.

[1]: https://www.kapa.ai/use-cases "kapa.ai - AI Assistant for Technical Documentation"
[2]: https://docs.kapa.ai/overview/showcase "Showcase | kapa.ai docs"
[3]: https://docs.kapa.ai/api/overview "Overview | kapa.ai docs"
[4]: https://docs.kapa.ai/analytics/dashboard "Dashboard | kapa.ai docs"
[5]: https://docs.kapa.ai/analytics/coverage-gaps "Coverage Gaps | kapa.ai docs"
[6]: https://docs.kapa.ai/integrations/website-widget "Website Widget | kapa.ai docs"
[7]: https://docs.kapa.ai/integrations/slack-bot "Slack Bot | kapa.ai docs"
[8]: https://www.kapa.ai/use-cases/support-ticket-deflector "AI Support Ticket Deflector - 20-40% Reduction | kapa.ai"
[9]: https://docs.kapa.ai/integrations/internal-technical-assistant "Internal Technical Assistant | kapa.ai docs"
[10]: https://docs.kapa.ai/analytics/conversations?utm_source=chatgpt.com "Conversations | kapa.ai docs"
[11]: https://www.kapa.ai/meeting-booked?utm_source=chatgpt.com "AI Assistant for Technical Documentation"
[12]: https://docs.kapa.ai/integrations/discord-bot "Discord Bot | kapa.ai docs"
[13]: https://docs.kapa.ai/api/reference/query-v-1-threads-chat?utm_source=chatgpt.com "Chat in thread | kapa.ai docs"
[14]: https://www.kapa.ai/meeting-requested?utm_source=chatgpt.com "kapa.ai - AI Assistant for Technical Documentation"
