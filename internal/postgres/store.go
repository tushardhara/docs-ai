package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cgap/internal/model"
	"cgap/internal/storage"
)

// Store implements storage.Store using PostgreSQL with pgx.
type Store struct {
	pool *pgxpool.Pool
}

func New(dsn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Projects() storage.ProjectRepo {
	return &ProjectRepo{pool: s.pool}
}

func (s *Store) Documents() storage.DocumentRepo {
	return &DocumentRepo{pool: s.pool}
}

func (s *Store) Chunks() storage.ChunkRepo {
	return &ChunkRepo{pool: s.pool}
}

func (s *Store) Threads() storage.ThreadRepo {
	return &ThreadRepo{pool: s.pool}
}

func (s *Store) Messages() storage.MessageRepo {
	return &MessageRepo{pool: s.pool}
}

func (s *Store) Answers() storage.AnswerRepo {
	return &AnswerRepo{pool: s.pool}
}

func (s *Store) Citations() storage.CitationRepo {
	return &CitationRepo{pool: s.pool}
}

func (s *Store) Analytics() storage.AnalyticsRepo {
	return &AnalyticsRepo{pool: s.pool}
}

func (s *Store) Gaps() storage.GapRepo {
	return &GapRepo{pool: s.pool}
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

// Pool exposes the underlying pgx pool for advanced queries (e.g., pgvector search).
// This method allows internal packages to perform specialized SQL without expanding
// the storage interfaces.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// ProjectRepo implementation.
type ProjectRepo struct {
	pool *pgxpool.Pool
}

func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*model.Project, error) {
	const query = `
		SELECT id, name, slug, default_model, settings, usage_plan, created_at, updated_at
		FROM projects WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	p := &model.Project{}
	err := row.Scan(
		&p.ID, &p.Name, &p.Slug, &p.DefaultModel, &p.Settings, &p.UsagePlan, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepo) GetBySlug(ctx context.Context, slug string) (*model.Project, error) {
	const query = `
		SELECT id, name, slug, default_model, settings, usage_plan, created_at, updated_at
		FROM projects WHERE slug = $1
	`
	row := r.pool.QueryRow(ctx, query, slug)
	p := &model.Project{}
	err := row.Scan(
		&p.ID, &p.Name, &p.Slug, &p.DefaultModel, &p.Settings, &p.UsagePlan, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project by slug: %w", err)
	}
	return p, nil
}

func (r *ProjectRepo) Create(ctx context.Context, p *model.Project) error {
	const query = `
		INSERT INTO projects (id, name, slug, default_model, settings, usage_plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query, p.ID, p.Name, p.Slug, p.DefaultModel, p.Settings, p.UsagePlan, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func (r *ProjectRepo) Update(ctx context.Context, p *model.Project) error {
	const query = `
		UPDATE projects 
		SET name = $1, slug = $2, default_model = $3, settings = $4, usage_plan = $5, updated_at = $6
		WHERE id = $7
	`
	_, err := r.pool.Exec(ctx, query, p.Name, p.Slug, p.DefaultModel, p.Settings, p.UsagePlan, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// DocumentRepo implementation.
type DocumentRepo struct {
	pool *pgxpool.Pool
}

func (r *DocumentRepo) GetByID(ctx context.Context, id string) (*model.Document, error) {
	const query = `
		SELECT id, project_id, source_id, uri, title, lang, version, hash, published_at, created_at
		FROM documents WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	d := &model.Document{}
	err := row.Scan(
		&d.ID, &d.ProjectID, &d.SourceID, &d.URI, &d.Title, &d.Lang, &d.Version, &d.Hash, &d.PublishedAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return d, nil
}

func (r *DocumentRepo) GetByURI(ctx context.Context, projectID, uri string) (*model.Document, error) {
	const query = `
		SELECT id, project_id, source_id, uri, title, lang, version, hash, published_at, created_at
		FROM documents WHERE project_id = $1 AND uri = $2
	`
	row := r.pool.QueryRow(ctx, query, projectID, uri)
	d := &model.Document{}
	err := row.Scan(
		&d.ID, &d.ProjectID, &d.SourceID, &d.URI, &d.Title, &d.Lang, &d.Version, &d.Hash, &d.PublishedAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document by uri: %w", err)
	}
	return d, nil
}

func (r *DocumentRepo) Create(ctx context.Context, d *model.Document) error {
	const query = `
		INSERT INTO documents (id, project_id, source_id, uri, title, lang, version, hash, published_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query, d.ID, d.ProjectID, d.SourceID, d.URI, d.Title, d.Lang, d.Version, d.Hash, d.PublishedAt, d.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	return nil
}

func (r *DocumentRepo) List(ctx context.Context, projectID string, limit, offset int) ([]*model.Document, error) {
	const query = `
		SELECT id, project_id, source_id, uri, title, lang, version, hash, published_at, created_at
		FROM documents WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		d := &model.Document{}
		err := rows.Scan(
			&d.ID, &d.ProjectID, &d.SourceID, &d.URI, &d.Title, &d.Lang, &d.Version, &d.Hash, &d.PublishedAt, &d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return docs, nil
}

// ChunkRepo implementation.
type ChunkRepo struct {
	pool *pgxpool.Pool
}

func (r *ChunkRepo) GetByID(ctx context.Context, id string) (*model.Chunk, error) {
	const query = `
		SELECT id, document_id, ord, text, token_count, section_path, score_raw, created_at
		FROM chunks WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	c := &model.Chunk{}
	err := row.Scan(
		&c.ID, &c.DocumentID, &c.Ord, &c.Text, &c.TokenCount, &c.SectionPath, &c.ScoreRaw, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk: %w", err)
	}
	return c, nil
}

func (r *ChunkRepo) CreateBatch(ctx context.Context, chunks []*model.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, c := range chunks {
		batch.Queue(
			"INSERT INTO chunks (id, document_id, ord, text, token_count, section_path, score_raw, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			c.ID, c.DocumentID, c.Ord, c.Text, c.TokenCount, c.SectionPath, c.ScoreRaw, c.CreatedAt,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(chunks); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to create chunk batch: %w", err)
		}
	}

	return nil
}

func (r *ChunkRepo) ListByDocument(ctx context.Context, documentID string) ([]*model.Chunk, error) {
	const query = `
		SELECT id, document_id, ord, text, token_count, section_path, score_raw, created_at
		FROM chunks WHERE document_id = $1
		ORDER BY ord ASC
	`
	rows, err := r.pool.Query(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chunks: %w", err)
	}
	defer rows.Close()

	var chunks []*model.Chunk
	for rows.Next() {
		c := &model.Chunk{}
		err := rows.Scan(
			&c.ID, &c.DocumentID, &c.Ord, &c.Text, &c.TokenCount, &c.SectionPath, &c.ScoreRaw, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}
		chunks = append(chunks, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return chunks, nil
}

// ThreadRepo implementation.
type ThreadRepo struct {
	pool *pgxpool.Pool
}

func (r *ThreadRepo) GetByID(ctx context.Context, id string) (*model.Thread, error) {
	const query = `
		SELECT id, project_id, integration, external_ref, status, created_at, updated_at
		FROM threads WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	t := &model.Thread{}
	err := row.Scan(&t.ID, &t.ProjectID, &t.Integration, &t.ExternalRef, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	return t, nil
}

func (r *ThreadRepo) Create(ctx context.Context, t *model.Thread) error {
	const query = `
		INSERT INTO threads (id, project_id, integration, external_ref, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, t.ID, t.ProjectID, t.Integration, t.ExternalRef, t.Status, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create thread: %w", err)
	}
	return nil
}

func (r *ThreadRepo) Update(ctx context.Context, t *model.Thread) error {
	const query = `
		UPDATE threads
		SET integration = $1, external_ref = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, t.Integration, t.ExternalRef, t.Status, t.UpdatedAt, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update thread: %w", err)
	}
	return nil
}

// MessageRepo implementation.
type MessageRepo struct {
	pool *pgxpool.Pool
}

func (r *MessageRepo) GetByID(ctx context.Context, id string) (*model.Message, error) {
	const query = `
		SELECT id, thread_id, role, content, meta, latency_ms, created_at
		FROM messages WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	m := &model.Message{}
	err := row.Scan(&m.ID, &m.ThreadID, &m.Role, &m.Content, &m.Meta, &m.LatencyMS, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return m, nil
}

func (r *MessageRepo) Create(ctx context.Context, m *model.Message) error {
	const query = `
		INSERT INTO messages (id, thread_id, role, content, meta, latency_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, m.ID, m.ThreadID, m.Role, m.Content, m.Meta, m.LatencyMS, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (r *MessageRepo) ListByThread(ctx context.Context, threadID string, limit, offset int) ([]*model.Message, error) {
	const query = `
		SELECT id, thread_id, role, content, meta, latency_ms, created_at
		FROM messages WHERE thread_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, threadID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		m := &model.Message{}
		err := rows.Scan(&m.ID, &m.ThreadID, &m.Role, &m.Content, &m.Meta, &m.LatencyMS, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return messages, nil
}

// AnswerRepo implementation.
type AnswerRepo struct {
	pool *pgxpool.Pool
}

func (r *AnswerRepo) Create(ctx context.Context, a *model.Answer) error {
	const query = `
		INSERT INTO answers (message_id, model, is_uncertain, reasoning_trace, prompt_version)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, a.MessageID, a.Model, a.IsUncertain, a.ReasoningTrace, a.PromptVersion)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}
	return nil
}

func (r *AnswerRepo) GetByMessageID(ctx context.Context, messageID string) (*model.Answer, error) {
	const query = `
		SELECT message_id, model, is_uncertain, reasoning_trace, prompt_version
		FROM answers WHERE message_id = $1
	`
	row := r.pool.QueryRow(ctx, query, messageID)
	a := &model.Answer{}
	err := row.Scan(&a.MessageID, &a.Model, &a.IsUncertain, &a.ReasoningTrace, &a.PromptVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get answer: %w", err)
	}
	return a, nil
}

// CitationRepo implementation.
type CitationRepo struct {
	pool *pgxpool.Pool
}

func (r *CitationRepo) CreateBatch(ctx context.Context, citations []*model.Citation) error {
	if len(citations) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, c := range citations {
		batch.Queue(
			"INSERT INTO citations (id, answer_id, chunk_id, score, quote, start_char, end_char) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			c.ID, c.AnswerID, c.ChunkID, c.Score, c.Quote, c.StartChar, c.EndChar,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(citations); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to create citation batch: %w", err)
		}
	}

	return nil
}

func (r *CitationRepo) ListByAnswer(ctx context.Context, answerID string) ([]*model.Citation, error) {
	const query = `
		SELECT id, answer_id, chunk_id, score, quote, start_char, end_char
		FROM citations WHERE answer_id = $1
	`
	rows, err := r.pool.Query(ctx, query, answerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list citations: %w", err)
	}
	defer rows.Close()

	var citations []*model.Citation
	for rows.Next() {
		c := &model.Citation{}
		err := rows.Scan(&c.ID, &c.AnswerID, &c.ChunkID, &c.Score, &c.Quote, &c.StartChar, &c.EndChar)
		if err != nil {
			return nil, fmt.Errorf("failed to scan citation: %w", err)
		}
		citations = append(citations, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return citations, nil
}

// AnalyticsRepo implementation.
type AnalyticsRepo struct {
	pool *pgxpool.Pool
}

func (r *AnalyticsRepo) RecordEvent(ctx context.Context, e *model.AnalyticsEvent) error {
	const query = `
		INSERT INTO analytics_events (id, project_id, thread_id, message_id, type, properties, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, e.ID, e.ProjectID, e.ThreadID, e.MessageID, e.Type, e.Properties, e.OccurredAt)
	if err != nil {
		return fmt.Errorf("failed to record event: %w", err)
	}
	return nil
}

func (r *AnalyticsRepo) CountQuestions(ctx context.Context, projectID string, from, to string) (int, error) {
	const query = `
		SELECT COUNT(DISTINCT message_id) FROM messages
		WHERE thread_id IN (
			SELECT id FROM threads WHERE project_id = $1
		)
		AND created_at >= $2::timestamp AND created_at <= $3::timestamp
	`
	var count int
	err := r.pool.QueryRow(ctx, query, projectID, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count questions: %w", err)
	}
	return count, nil
}

func (r *AnalyticsRepo) CountUncertain(ctx context.Context, projectID string, from, to string) (int, error) {
	const query = `
		SELECT COUNT(*) FROM answers
		WHERE is_uncertain = true
		AND message_id IN (
			SELECT m.id FROM messages m
			WHERE m.thread_id IN (
				SELECT id FROM threads WHERE project_id = $1
			)
			AND m.created_at >= $2::timestamp AND m.created_at <= $3::timestamp
		)
	`
	var count int
	err := r.pool.QueryRow(ctx, query, projectID, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count uncertain answers: %w", err)
	}
	return count, nil
}

// GapRepo implementation.
type GapRepo struct {
	pool *pgxpool.Pool
}

func (r *GapRepo) CreateCandidate(ctx context.Context, gc *model.GapCandidate) error {
	const query = `
		INSERT INTO gap_candidates (answer_id, question_embedding, uncertainty_reason)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, gc.AnswerID, gc.QuestionEmbedding, gc.UncertaintyReason)
	if err != nil {
		return fmt.Errorf("failed to create gap candidate: %w", err)
	}
	return nil
}

func (r *GapRepo) CreateCluster(ctx context.Context, gc *model.GapCluster) error {
	const query = `
		INSERT INTO gap_clusters (id, project_id, window, label, summary, recommendation, size, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query, gc.ID, gc.ProjectID, gc.Window, gc.Label, gc.Summary, gc.Recommendation, gc.Size, gc.Status, gc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create gap cluster: %w", err)
	}
	return nil
}

func (r *GapRepo) CreateExample(ctx context.Context, gce *model.GapClusterExample) error {
	const query = `
		INSERT INTO gap_cluster_examples (id, cluster_id, answer_id, question, citations, representative_score)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, gce.ID, gce.ClusterID, gce.AnswerID, gce.Question, gce.Citations, gce.RepresentativeScore)
	if err != nil {
		return fmt.Errorf("failed to create gap cluster example: %w", err)
	}
	return nil
}

func (r *GapRepo) ListClusters(ctx context.Context, projectID, window string) ([]*model.GapCluster, error) {
	const query = `
		SELECT id, project_id, window, label, summary, recommendation, size, status, created_at
		FROM gap_clusters
		WHERE project_id = $1 AND window = $2
		ORDER BY size DESC
	`
	rows, err := r.pool.Query(ctx, query, projectID, window)
	if err != nil {
		return nil, fmt.Errorf("failed to list gap clusters: %w", err)
	}
	defer rows.Close()

	var clusters []*model.GapCluster
	for rows.Next() {
		gc := &model.GapCluster{}
		err := rows.Scan(&gc.ID, &gc.ProjectID, &gc.Window, &gc.Label, &gc.Summary, &gc.Recommendation, &gc.Size, &gc.Status, &gc.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gap cluster: %w", err)
		}
		clusters = append(clusters, gc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return clusters, nil
}

func (r *GapRepo) GetClusterDetail(ctx context.Context, clusterID string) (*model.GapCluster, []*model.GapClusterExample, error) {
	const queryCluster = `
		SELECT id, project_id, window, label, summary, recommendation, size, status, created_at
		FROM gap_clusters WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, queryCluster, clusterID)
	gc := &model.GapCluster{}
	err := row.Scan(&gc.ID, &gc.ProjectID, &gc.Window, &gc.Label, &gc.Summary, &gc.Recommendation, &gc.Size, &gc.Status, &gc.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get gap cluster: %w", err)
	}

	const queryExamples = `
		SELECT id, cluster_id, answer_id, question, citations, representative_score
		FROM gap_cluster_examples WHERE cluster_id = $1
		ORDER BY representative_score DESC
	`
	rows, err := r.pool.Query(ctx, queryExamples, clusterID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list gap cluster examples: %w", err)
	}
	defer rows.Close()

	var examples []*model.GapClusterExample
	for rows.Next() {
		gce := &model.GapClusterExample{}
		err := rows.Scan(&gce.ID, &gce.ClusterID, &gce.AnswerID, &gce.Question, &gce.Citations, &gce.RepresentativeScore)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan gap cluster example: %w", err)
		}
		examples = append(examples, gce)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	return gc, examples, nil
}
