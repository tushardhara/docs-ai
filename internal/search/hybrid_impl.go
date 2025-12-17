package search

import (
	"context"

	"cgap/internal/service"
)

// Hybrid blends two search providers (primary preferred, secondary for recall).
type Hybrid struct {
	primary   service.Search
	secondary service.Search
}

func NewHybrid(primary, secondary service.Search) *Hybrid {
	return &Hybrid{primary: primary, secondary: secondary}
}

// Search queries primary then secondary and merges results up to topK.
func (h *Hybrid) Search(ctx context.Context, index, query string, topK int, filters map[string]any) ([]service.SearchResult, error) {
	if topK <= 0 {
		topK = 10
	}

	// Try primary
	pRes, pErr := h.primary.Search(ctx, index, query, topK, filters)

	// Always try secondary to improve recall
	sRes, sErr := h.secondary.Search(ctx, index, query, topK, filters)

	// If both failed, return one of the errors
	if pErr != nil && sErr != nil {
		return nil, pErr
	}

	// Merge results: prefer primary ordering, then fill with secondary (dedupe by ID)
	seen := make(map[string]bool)
	out := make([]service.SearchResult, 0, topK)

	for _, r := range pRes {
		if len(out) >= topK {
			break
		}
		if !seen[r.ID] {
			seen[r.ID] = true
			out = append(out, r)
		}
	}

	for _, r := range sRes {
		if len(out) >= topK {
			break
		}
		if !seen[r.ID] {
			seen[r.ID] = true
			out = append(out, r)
		}
	}

	return out, nil
}
