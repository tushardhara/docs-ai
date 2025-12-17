-- +goose Up
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
  time_window text CHECK (time_window IN ('7d','30d','90d')),
  label text,
  summary text,
  recommendation text,
  size int,
  status text CHECK (status IN ('open','in_review','done')) DEFAULT 'open',
  created_at timestamptz DEFAULT now()
);
CREATE INDEX gap_clusters_project_time_window ON gap_clusters(project_id, time_window);

CREATE TABLE gap_cluster_examples (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  cluster_id uuid REFERENCES gap_clusters(id) ON DELETE CASCADE,
  answer_id uuid REFERENCES answers(message_id) ON DELETE CASCADE,
  question text,
  citations uuid[],
  representative_score real
);
CREATE INDEX gap_examples_cluster ON gap_cluster_examples(cluster_id);

-- +goose Down
DROP TABLE IF EXISTS gap_cluster_examples;
DROP TABLE IF EXISTS gap_clusters;
DROP TABLE IF EXISTS gap_candidates;
DROP TABLE IF EXISTS analytics_events;
DROP TABLE IF EXISTS deflect_events;
DROP TABLE IF EXISTS feedback;
DROP TABLE IF EXISTS citations;
DROP TABLE IF EXISTS answers;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS chunk_embeddings;
DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS sources;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS projects;
