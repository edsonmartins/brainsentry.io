package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

// BatchSearchService executes multiple queries in one call and returns a
// multi-dimensional result: for each memory, a score per query.
//
// Useful for UX patterns like "which of these 5 query formulations works best?"
// or batch re-ranking where relevance must be compared across queries.
type BatchSearchService struct {
	memoryService *MemoryService
}

// NewBatchSearchService creates a new BatchSearchService.
func NewBatchSearchService(memoryService *MemoryService) *BatchSearchService {
	return &BatchSearchService{memoryService: memoryService}
}

// BatchSearchRequest contains multiple queries and a shared result limit.
type BatchSearchRequest struct {
	Queries []string `json:"queries"`
	Limit   int      `json:"limit,omitempty"` // per-query limit (default 10)
	Tags    []string `json:"tags,omitempty"`
}

// BatchScore captures a memory's relevance across all queries in the batch.
type BatchScore struct {
	MemoryID      string    `json:"memoryId"`
	Summary       string    `json:"summary,omitempty"`
	Category      string    `json:"category,omitempty"`
	PerQuery      []float64 `json:"perQuery"`       // relevance per query (same order as request)
	MatchedQueries []int    `json:"matchedQueries"` // indices of queries where memory scored above threshold
	Mean          float64   `json:"mean"`
	Max           float64   `json:"max"`
}

// BatchSearchResponse is the multi-dimensional result.
type BatchSearchResponse struct {
	Queries     []string     `json:"queries"`
	Results     []BatchScore `json:"results"`
	SearchTimeMs int64       `json:"searchTimeMs"`
}

// MatchThreshold is the score above which a memory is considered a match for a query.
const batchMatchThreshold = 0.05

// Search runs all queries in parallel (goroutines) and aggregates per-memory scores.
func (s *BatchSearchService) Search(ctx context.Context, req BatchSearchRequest) (*BatchSearchResponse, error) {
	start := time.Now()

	if len(req.Queries) == 0 {
		return nil, fmt.Errorf("at least one query is required")
	}
	if len(req.Queries) > 20 {
		return nil, fmt.Errorf("too many queries (max 20)")
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Run queries in parallel
	type perQueryResult struct {
		index  int
		memory map[string]float64 // memoryID → relevance
		err    error
	}

	results := make(chan perQueryResult, len(req.Queries))
	for i, q := range req.Queries {
		go func(idx int, query string) {
			resp, err := s.memoryService.SearchMemories(ctx, dto.SearchRequest{
				Query: query,
				Tags:  req.Tags,
				Limit: limit * 2,
			})
			if err != nil {
				results <- perQueryResult{index: idx, err: err}
				return
			}
			scores := make(map[string]float64, len(resp.Results))
			for _, r := range resp.Results {
				scores[r.ID] = r.RelevanceScore
			}
			results <- perQueryResult{index: idx, memory: scores}
		}(i, q)
	}

	// Collect with error propagation
	perQueryScores := make([]map[string]float64, len(req.Queries))
	for i := 0; i < len(req.Queries); i++ {
		r := <-results
		if r.err != nil {
			return nil, fmt.Errorf("query %d failed: %w", r.index, r.err)
		}
		perQueryScores[r.index] = r.memory
	}

	// Collect unique memory IDs and fetch details
	uniqueIDs := make(map[string]bool)
	for _, scores := range perQueryScores {
		for id := range scores {
			uniqueIDs[id] = true
		}
	}

	batchResults := make([]BatchScore, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		memory, err := s.memoryService.GetMemory(ctx, id)
		if err != nil || memory == nil {
			continue
		}

		score := BatchScore{
			MemoryID: id,
			Summary:  memory.Summary,
			Category: string(memory.Category),
			PerQuery: make([]float64, len(req.Queries)),
		}

		var sum, maxScore float64
		for qIdx, qScores := range perQueryScores {
			relevance := qScores[id] // 0 if not matched
			score.PerQuery[qIdx] = relevance
			sum += relevance
			if relevance > maxScore {
				maxScore = relevance
			}
			if relevance > batchMatchThreshold {
				score.MatchedQueries = append(score.MatchedQueries, qIdx)
			}
		}
		score.Mean = sum / float64(len(req.Queries))
		score.Max = maxScore
		batchResults = append(batchResults, score)
	}

	// Sort by max score descending
	sort.Slice(batchResults, func(i, j int) bool {
		return batchResults[i].Max > batchResults[j].Max
	})

	// Cap at per-query limit × queries
	cap := limit * len(req.Queries)
	if len(batchResults) > cap {
		batchResults = batchResults[:cap]
	}

	return &BatchSearchResponse{
		Queries:     req.Queries,
		Results:     batchResults,
		SearchTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

// Ensure domain import is always used (BatchScore may reference it indirectly).
var _ = domain.ImportanceCritical
