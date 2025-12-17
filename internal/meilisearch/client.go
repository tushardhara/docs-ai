package meilisearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cgap/internal/service"
)

// Client wraps Meilisearch HTTP client for search operations.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

// searchRequest matches Meilisearch search API request format
type searchRequest struct {
	Q       string   `json:"q"`
	Limit   int      `json:"limit,omitempty"`
	Filter  []string `json:"filter,omitempty"`
	Ranking []string `json:"rankingRules,omitempty"`
}

// searchResponse matches Meilisearch search API response format
type searchResponse struct {
	Hits    []searchHit `json:"hits"`
	Query   string      `json:"query"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	EstRC   int         `json:"estimatedTotalHits"`
	Process int         `json:"processingTimeMs"`
}

type searchHit struct {
	ID           string  `json:"id"`
	ChunkID      string  `json:"chunk_id"`
	DocumentID   string  `json:"document_id"`
	Text         string  `json:"text"`
	SectionPath  string  `json:"section_path,omitempty"`
	RankingScore float32 `json:"_rankingScore,omitempty"`
}

// Search queries Meilisearch and returns results.
func (c *Client) Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]service.SearchResult, error) {
	if topK == 0 {
		topK = 10
	}

	// Build filter array from filters map
	var filterArray []string
	for key, val := range filters {
		if val != nil {
			filterArray = append(filterArray, fmt.Sprintf("%s = %v", key, val))
		}
	}

	req := searchRequest{
		Q:      query,
		Limit:  topK,
		Filter: filterArray,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/indexes/%s/search", c.baseURL, index), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("meilisearch error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var searchResp searchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to service.SearchResult
	var results []service.SearchResult
	for _, hit := range searchResp.Hits {
		results = append(results, service.SearchResult{
			ID:   hit.ChunkID,
			Text: hit.Text,
			Metadata: map[string]any{
				"document_id":  hit.DocumentID,
				"section_path": hit.SectionPath,
				"chunk_id":     hit.ChunkID,
			},
			Score: hit.RankingScore,
		})
	}

	return results, nil
}
