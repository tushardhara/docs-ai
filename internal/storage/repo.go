package storage

import (
	"context"

	"cgap/internal/model"
)

// Database repository interfaces decoupled from implementation.

type ProjectRepo interface {
	GetByID(ctx context.Context, id string) (*model.Project, error)
	GetBySlug(ctx context.Context, slug string) (*model.Project, error)
	Create(ctx context.Context, p *model.Project) error
	Update(ctx context.Context, p *model.Project) error
}

type DocumentRepo interface {
	GetByID(ctx context.Context, id string) (*model.Document, error)
	GetByURI(ctx context.Context, projectID, uri string) (*model.Document, error)
	Create(ctx context.Context, d *model.Document) error
	List(ctx context.Context, projectID string, limit, offset int) ([]*model.Document, error)
}

type ChunkRepo interface {
	GetByID(ctx context.Context, id string) (*model.Chunk, error)
	CreateBatch(ctx context.Context, chunks []*model.Chunk) error
	ListByDocument(ctx context.Context, documentID string) ([]*model.Chunk, error)
}

type ThreadRepo interface {
	GetByID(ctx context.Context, id string) (*model.Thread, error)
	Create(ctx context.Context, t *model.Thread) error
	Update(ctx context.Context, t *model.Thread) error
}

type MessageRepo interface {
	GetByID(ctx context.Context, id string) (*model.Message, error)
	Create(ctx context.Context, m *model.Message) error
	ListByThread(ctx context.Context, threadID string, limit, offset int) ([]*model.Message, error)
}

type AnswerRepo interface {
	Create(ctx context.Context, a *model.Answer) error
	GetByMessageID(ctx context.Context, messageID string) (*model.Answer, error)
}

type CitationRepo interface {
	CreateBatch(ctx context.Context, citations []*model.Citation) error
	ListByAnswer(ctx context.Context, answerID string) ([]*model.Citation, error)
}

type AnalyticsRepo interface {
	RecordEvent(ctx context.Context, e *model.AnalyticsEvent) error
	CountQuestions(ctx context.Context, projectID string, from, to string) (int, error)
	CountUncertain(ctx context.Context, projectID string, from, to string) (int, error)
}

type GapRepo interface {
	CreateCandidate(ctx context.Context, gc *model.GapCandidate) error
	CreateCluster(ctx context.Context, gc *model.GapCluster) error
	CreateExample(ctx context.Context, gce *model.GapClusterExample) error
	ListClusters(ctx context.Context, projectID, window string) ([]*model.GapCluster, error)
	GetClusterDetail(ctx context.Context, clusterID string) (*model.GapCluster, []*model.GapClusterExample, error)
}

// Store aggregates all repos.
type Store interface {
	Projects() ProjectRepo
	Documents() DocumentRepo
	Chunks() ChunkRepo
	Threads() ThreadRepo
	Messages() MessageRepo
	Answers() AnswerRepo
	Citations() CitationRepo
	Analytics() AnalyticsRepo
	Gaps() GapRepo
	Close() error
}
