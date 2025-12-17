package ingestion

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cgap/internal/model"
)

// Crawler fetches documents from various sources.
type Crawler interface {
	Crawl(ctx context.Context, source *model.Source) ([]RawDoc, error)
}

// RawDoc represents a document fetched from a source.
type RawDoc struct {
	URI         string
	Title       string
	Content     string
	Language    string
	PublishedAt string
}

// URLCrawler crawls static HTML/markdown from a URL.
type URLCrawler struct {
	client *http.Client
}

func NewURLCrawler() *URLCrawler {
	return &URLCrawler{
		client: &http.Client{},
	}
}

func (c *URLCrawler) Crawl(ctx context.Context, source *model.Source) ([]RawDoc, error) {
	url, ok := source.Config["url"].(string)
	if !ok {
		return nil, fmt.Errorf("source config missing url")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch failed: %d", resp.StatusCode)
	}

	return []RawDoc{
		{
			URI:      url,
			Title:    "Fetched Doc",
			Content:  "Content...",
			Language: "en",
		},
	}, nil
}

// Chunker splits documents into chunks.
type Chunker interface {
	Chunk(ctx context.Context, doc *model.Document, content string) ([]Chunk, error)
}

// Chunk represents a chunk of text to be indexed.
type Chunk struct {
	Ord         int
	Text        string
	TokenCount  int
	SectionPath string
	ScoreRaw    float32
}

// SimpleChunker uses fixed token size + overlap.
type SimpleChunker struct {
	chunkSize int
	overlap   int
}

func NewSimpleChunker(chunkSize, overlap int) *SimpleChunker {
	return &SimpleChunker{
		chunkSize: chunkSize,
		overlap:   overlap,
	}
}

func (c *SimpleChunker) Chunk(ctx context.Context, doc *model.Document, content string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []Chunk
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			chunks = append(chunks, Chunk{
				Ord:         i,
				Text:        line,
				TokenCount:  len(strings.Fields(line)),
				SectionPath: "root",
			})
		}
	}
	return chunks, nil
}

// IngestionPipeline orchestrates crawl -> chunk -> embed -> index.
type IngestionPipeline struct {
	crawler  Crawler
	chunker  Chunker
	embedder Embedder
	indexer  Indexer
}

func NewIngestionPipeline(crawler Crawler, chunker Chunker, embedder Embedder, indexer Indexer) *IngestionPipeline {
	return &IngestionPipeline{
		crawler:  crawler,
		chunker:  chunker,
		embedder: embedder,
		indexer:  indexer,
	}
}

func (p *IngestionPipeline) Run(ctx context.Context, source *model.Source, doc *model.Document) error {
	rawDocs, err := p.crawler.Crawl(ctx, source)
	if err != nil {
		return err
	}

	for _, raw := range rawDocs {
		chunks, err := p.chunker.Chunk(ctx, doc, raw.Content)
		if err != nil {
			return err
		}

		embeddings, err := p.embedder.Embed(ctx, chunks)
		if err != nil {
			return err
		}

		if err := p.indexer.Index(ctx, doc, chunks, embeddings); err != nil {
			return err
		}
	}

	return nil
}

// Embedder generates embeddings for chunks.
type Embedder interface {
	Embed(ctx context.Context, chunks []Chunk) ([][]float32, error)
}

// Indexer stores chunks and embeddings.
type Indexer interface {
	Index(ctx context.Context, doc *model.Document, chunks []Chunk, embeddings [][]float32) error
}
