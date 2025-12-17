package meilisearch

import (
	"context"
	"fmt"

	"cgap/internal/service"
)

// Client wraps Meilisearch HTTP client for search operations.
type Client struct {
	baseURL string
	apiKey  string
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// Search queries Meilisearch.
func (c *Client) Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]service.SearchResult, error) {
	// TODO: Implement Meilisearch HTTP client
	// POST /indexes/{index}/search with query + filters
	// Parse response and return SearchResult array

	_ = ctx
	_ = index
	_ = query
	_ = topK
	_ = filters

	return nil, fmt.Errorf("not implemented")
}
