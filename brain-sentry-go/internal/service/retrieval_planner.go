package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	graphrepo "github.com/integraltech/brainsentry/internal/repository/graph"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// RetrievalPlannerService performs intent-aware retrieval with reflection loops.
type RetrievalPlannerService struct {
	openRouter       *OpenRouterService
	memoryRepo       *postgres.MemoryRepository
	memoryGraphRepo  *graphrepo.MemoryGraphRepository
	embeddingService *EmbeddingService
	maxRounds        int
	coverageTarget   float64 // 0-1, e.g. 0.8 = stop when 80% coverage
}

// NewRetrievalPlannerService creates a new RetrievalPlannerService.
func NewRetrievalPlannerService(
	openRouter *OpenRouterService,
	memoryRepo *postgres.MemoryRepository,
	memoryGraphRepo *graphrepo.MemoryGraphRepository,
	embeddingService *EmbeddingService,
) *RetrievalPlannerService {
	return &RetrievalPlannerService{
		openRouter:       openRouter,
		memoryRepo:       memoryRepo,
		memoryGraphRepo:  memoryGraphRepo,
		embeddingService: embeddingService,
		maxRounds:        3,
		coverageTarget:   0.8,
	}
}

// RetrievalPlan represents the planner's analysis of a query.
type RetrievalPlan struct {
	OriginalQuery string       `json:"originalQuery"`
	InfoNeeds     []InfoNeed   `json:"infoNeeds"`
	SubQueries    []SubQuery   `json:"subQueries"`
	Rounds        int          `json:"rounds"`
	Coverage      float64      `json:"coverage"`
	Results       []PlanResult `json:"results"`
	TotalTimeMs   int64        `json:"totalTimeMs"`
}

// InfoNeed represents a type of information needed to answer the query.
type InfoNeed struct {
	Type        string `json:"type"`        // e.g., "factual", "procedural", "contextual"
	Description string `json:"description"` // what info is needed
	Satisfied   bool   `json:"satisfied"`   // whether this need has been met
}

// SubQuery is a directed sub-query generated from an info need.
type SubQuery struct {
	Query    string `json:"query"`
	Purpose  string `json:"purpose"`  // which info need this targets
	ViewType string `json:"viewType"` // "semantic", "lexical", "graph"
}

// PlanResult is a memory retrieved by the planner with its source sub-query.
type PlanResult struct {
	MemoryID string  `json:"memoryId"`
	Content  string  `json:"content"`
	Score    float64 `json:"score"`
	Source   string  `json:"source"` // which sub-query found it
	Round    int     `json:"round"`
}

// PlanAndRetrieve performs intent-aware retrieval with reflection.
func (s *RetrievalPlannerService) PlanAndRetrieve(ctx context.Context, query string, limit int) (*RetrievalPlan, error) {
	start := time.Now()
	tenantID := tenant.FromContext(ctx)

	plan := &RetrievalPlan{
		OriginalQuery: query,
		Results:       make([]PlanResult, 0),
	}

	if s.openRouter == nil {
		// Fallback to simple search without planning
		return s.fallbackSearch(ctx, query, limit, plan, start)
	}

	// Phase 1: Analyze query intent and identify info needs
	infoNeeds, subQueries, err := s.analyzeIntent(ctx, query)
	if err != nil {
		slog.Warn("intent analysis failed, falling back", "error", err)
		return s.fallbackSearch(ctx, query, limit, plan, start)
	}
	plan.InfoNeeds = infoNeeds
	plan.SubQueries = subQueries

	seenIDs := make(map[string]bool)

	// Phase 2: Execute sub-queries across 3 views (semantic, lexical, graph)
	for round := 0; round < s.maxRounds; round++ {
		plan.Rounds = round + 1

		for _, sq := range subQueries {
			results := s.executeSubQuery(ctx, sq, tenantID, limit)
			for _, r := range results {
				if !seenIDs[r.MemoryID] {
					seenIDs[r.MemoryID] = true
					r.Round = round + 1
					plan.Results = append(plan.Results, r)
				}
			}
		}

		// Phase 3: Evaluate coverage
		coverage, satisfied := s.evaluateCoverage(ctx, plan)
		plan.Coverage = coverage
		for i, need := range plan.InfoNeeds {
			if satisfied[need.Type] {
				plan.InfoNeeds[i].Satisfied = true
			}
		}

		// Stop if coverage target met or enough results
		if coverage >= s.coverageTarget || len(plan.Results) >= limit*2 {
			break
		}

		// Phase 4: Generate gap-filling queries for unsatisfied needs
		gapQueries := s.generateGapQueries(ctx, plan)
		if len(gapQueries) == 0 {
			break
		}
		subQueries = gapQueries
	}

	// Sort and trim results
	sortPlanResults(plan.Results)
	if len(plan.Results) > limit {
		plan.Results = plan.Results[:limit]
	}

	plan.TotalTimeMs = time.Since(start).Milliseconds()
	return plan, nil
}

func (s *RetrievalPlannerService) analyzeIntent(ctx context.Context, query string) ([]InfoNeed, []SubQuery, error) {
	prompt := fmt.Sprintf(`Analyze the following query and identify:
1. What types of information are needed to fully answer it
2. Sub-queries to retrieve each type of information

Respond in JSON format only:
{
  "infoNeeds": [
    {"type": "factual|procedural|contextual|relational|temporal", "description": "what info is needed"}
  ],
  "subQueries": [
    {"query": "specific search query", "purpose": "which info need this targets", "viewType": "semantic|lexical|graph"}
  ]
}

Query: %s`, query)

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a retrieval planning system. Analyze queries and plan retrieval strategies. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		InfoNeeds  []InfoNeed `json:"infoNeeds"`
		SubQueries []SubQuery `json:"subQueries"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		return nil, nil, fmt.Errorf("parsing intent analysis: %w", err)
	}

	return result.InfoNeeds, result.SubQueries, nil
}

func (s *RetrievalPlannerService) executeSubQuery(ctx context.Context, sq SubQuery, tenantID string, limit int) []PlanResult {
	var results []PlanResult

	switch sq.ViewType {
	case "semantic":
		if s.embeddingService != nil && s.memoryGraphRepo != nil {
			embedding := s.embeddingService.Embed(sq.Query)
			ids, scores, err := s.memoryGraphRepo.VectorSearch(ctx, embedding, limit, tenantID)
			if err == nil {
				for i, id := range ids {
					m, err := s.memoryRepo.FindByID(ctx, id)
					if err != nil || isInactiveMemory(m, time.Now()) {
						continue
					}
					results = append(results, PlanResult{
						MemoryID: id,
						Content:  truncate(m.Content, 300),
						Score:    scores[i],
						Source:   "semantic:" + sq.Purpose,
					})
				}
			}
		}

	case "lexical":
		memories, err := s.memoryRepo.FullTextSearch(ctx, sq.Query, limit)
		if err == nil {
			for _, m := range memories {
				if isInactiveMemory(&m, time.Now()) {
					continue
				}
				results = append(results, PlanResult{
					MemoryID: m.ID,
					Content:  truncate(m.Content, 300),
					Score:    0.5, // default score for lexical
					Source:   "lexical:" + sq.Purpose,
				})
			}
		}

	case "graph":
		// Use full-text as proxy for graph queries
		memories, err := s.memoryRepo.FullTextSearch(ctx, sq.Query, limit)
		if err == nil {
			for _, m := range memories {
				if isInactiveMemory(&m, time.Now()) {
					continue
				}
				results = append(results, PlanResult{
					MemoryID: m.ID,
					Content:  truncate(m.Content, 300),
					Score:    0.4,
					Source:   "graph:" + sq.Purpose,
				})
			}
		}
	}

	return results
}

func (s *RetrievalPlannerService) evaluateCoverage(ctx context.Context, plan *RetrievalPlan) (float64, map[string]bool) {
	if len(plan.InfoNeeds) == 0 {
		return 1.0, nil
	}

	satisfied := make(map[string]bool)

	// Simple heuristic: check if any results match each info need
	for _, need := range plan.InfoNeeds {
		for _, r := range plan.Results {
			if r.Source != "" && strings.Contains(r.Source, need.Type) {
				satisfied[need.Type] = true
				break
			}
		}
		// Also check if results content relates to the need description
		if !satisfied[need.Type] {
			queryTokens := TokenizeQuery(need.Description)
			for _, r := range plan.Results {
				overlap := computeTokenOverlap(r.Content, "", queryTokens)
				if overlap > 0.3 {
					satisfied[need.Type] = true
					break
				}
			}
		}
	}

	coverage := float64(len(satisfied)) / float64(len(plan.InfoNeeds))
	return coverage, satisfied
}

func (s *RetrievalPlannerService) generateGapQueries(ctx context.Context, plan *RetrievalPlan) []SubQuery {
	var gaps []SubQuery
	for _, need := range plan.InfoNeeds {
		if !need.Satisfied {
			// Generate a more specific query for unsatisfied needs
			gaps = append(gaps, SubQuery{
				Query:    need.Description,
				Purpose:  need.Type,
				ViewType: "lexical", // try different view
			})
		}
	}
	return gaps
}

func (s *RetrievalPlannerService) fallbackSearch(ctx context.Context, query string, limit int, plan *RetrievalPlan, start time.Time) (*RetrievalPlan, error) {
	memories, err := s.memoryRepo.FullTextSearch(ctx, query, limit)
	if err != nil {
		return plan, err
	}
	for _, m := range memories {
		plan.Results = append(plan.Results, PlanResult{
			MemoryID: m.ID,
			Content:  truncate(m.Content, 300),
			Score:    0.5,
			Source:   "fallback",
			Round:    1,
		})
	}
	plan.Rounds = 1
	plan.Coverage = 1.0
	plan.TotalTimeMs = time.Since(start).Milliseconds()
	return plan, nil
}

func sortPlanResults(results []PlanResult) {
	for i := 1; i < len(results); i++ {
		key := results[i]
		j := i - 1
		for j >= 0 && results[j].Score < key.Score {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = key
	}
}
