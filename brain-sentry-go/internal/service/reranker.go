package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/integraltech/brainsentry/internal/domain"
)

// Reranker is the interface for post-retrieval reranking strategies.
type Reranker interface {
	Rerank(ctx context.Context, query string, memories []domain.Memory) ([]RankedMemory, error)
	Name() string
}

// RankedMemory is a memory with its reranking score.
type RankedMemory struct {
	Memory domain.Memory `json:"memory"`
	Score  float64       `json:"score"`
	Reason string        `json:"reason,omitempty"`
}

// NoOpReranker passes through results unchanged.
type NoOpReranker struct{}

func (r *NoOpReranker) Rerank(_ context.Context, _ string, memories []domain.Memory) ([]RankedMemory, error) {
	result := make([]RankedMemory, len(memories))
	for i, m := range memories {
		result[i] = RankedMemory{Memory: m, Score: 1.0 - float64(i)*0.01}
	}
	return result, nil
}

func (r *NoOpReranker) Name() string { return "noop" }

// BM25Reranker reranks using BM25-like term frequency scoring.
type BM25Reranker struct {
	k1 float64 // term saturation parameter (default 1.2)
	b  float64 // length normalization parameter (default 0.75)
}

// NewBM25Reranker creates a BM25 reranker with default parameters.
func NewBM25Reranker() *BM25Reranker {
	return &BM25Reranker{k1: 1.2, b: 0.75}
}

func (r *BM25Reranker) Rerank(_ context.Context, query string, memories []domain.Memory) ([]RankedMemory, error) {
	queryTerms := TokenizeQuery(query)
	if len(queryTerms) == 0 {
		result := make([]RankedMemory, len(memories))
		for i, m := range memories {
			result[i] = RankedMemory{Memory: m, Score: 0}
		}
		return result, nil
	}

	// Compute average document length
	totalLen := 0
	for _, m := range memories {
		totalLen += len(strings.Fields(m.Content))
	}
	avgDL := float64(totalLen) / float64(len(memories))
	if avgDL == 0 {
		avgDL = 1
	}

	// Compute IDF for each query term
	N := float64(len(memories))
	idf := make(map[string]float64)
	for _, term := range queryTerms {
		df := 0
		for _, m := range memories {
			if strings.Contains(strings.ToLower(m.Content+" "+m.Summary), term) {
				df++
			}
		}
		idf[term] = math.Log((N-float64(df)+0.5)/(float64(df)+0.5) + 1)
	}

	// Score each memory
	result := make([]RankedMemory, len(memories))
	for i, m := range memories {
		doc := strings.ToLower(m.Content + " " + m.Summary)
		docLen := float64(len(strings.Fields(doc)))
		score := 0.0

		for _, term := range queryTerms {
			tf := float64(strings.Count(doc, term))
			numerator := tf * (r.k1 + 1)
			denominator := tf + r.k1*(1-r.b+r.b*docLen/avgDL)
			score += idf[term] * numerator / denominator
		}

		result[i] = RankedMemory{Memory: m, Score: score, Reason: "bm25"}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result, nil
}

func (r *BM25Reranker) Name() string { return "bm25" }

// LLMReranker uses an LLM to score and rerank results.
type LLMReranker struct {
	openRouter *OpenRouterService
}

// NewLLMReranker creates an LLM-based reranker.
func NewLLMReranker(openRouter *OpenRouterService) *LLMReranker {
	return &LLMReranker{openRouter: openRouter}
}

func (r *LLMReranker) Rerank(ctx context.Context, query string, memories []domain.Memory) ([]RankedMemory, error) {
	if r.openRouter == nil || len(memories) == 0 {
		noop := &NoOpReranker{}
		return noop.Rerank(ctx, query, memories)
	}

	// Build candidates list
	var candidates string
	for i, m := range memories {
		summary := m.Summary
		if summary == "" {
			summary = truncate(m.Content, 200)
		}
		candidates += fmt.Sprintf("[%d] %s\n", i, summary)
	}

	prompt := fmt.Sprintf(`Given a query and candidate memories, score each memory's relevance from 0 to 1.

Query: %s

Candidates:
%s

Respond in JSON format only:
{"scores": [{"index": 0, "score": 0.9, "reason": "brief reason"}, ...]}`, query, candidates)

	response, err := r.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a relevance scoring system. Score memories by their relevance to the query. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		// Fallback to noop on error
		noop := &NoOpReranker{}
		return noop.Rerank(ctx, query, memories)
	}

	var scores struct {
		Scores []struct {
			Index  int     `json:"index"`
			Score  float64 `json:"score"`
			Reason string  `json:"reason"`
		} `json:"scores"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &scores); err != nil {
		noop := &NoOpReranker{}
		return noop.Rerank(ctx, query, memories)
	}

	// Map scores back to memories
	scoreMap := make(map[int]struct {
		score  float64
		reason string
	})
	for _, s := range scores.Scores {
		scoreMap[s.Index] = struct {
			score  float64
			reason string
		}{s.Score, s.Reason}
	}

	result := make([]RankedMemory, len(memories))
	for i, m := range memories {
		s, ok := scoreMap[i]
		if ok {
			result[i] = RankedMemory{Memory: m, Score: s.score, Reason: s.reason}
		} else {
			result[i] = RankedMemory{Memory: m, Score: 0.5, Reason: "unscored"}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result, nil
}

func (r *LLMReranker) Name() string { return "llm" }

// HybridScoreReranker reranks using the composite hybrid scoring system.
type HybridScoreReranker struct {
	weights ScoringWeights
}

// NewHybridScoreReranker creates a reranker based on composite hybrid scoring.
func NewHybridScoreReranker() *HybridScoreReranker {
	return &HybridScoreReranker{weights: DefaultScoringWeights}
}

func (r *HybridScoreReranker) Rerank(_ context.Context, query string, memories []domain.Memory) ([]RankedMemory, error) {
	queryTokens := TokenizeQuery(query)

	result := make([]RankedMemory, len(memories))
	for i, m := range memories {
		trace := ComputeHybridScore(&m, 0.5, queryTokens, -1, nil, r.weights)
		result[i] = RankedMemory{Memory: m, Score: trace.FinalScore, Reason: "hybrid"}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result, nil
}

func (r *HybridScoreReranker) Name() string { return "hybrid" }

// GetReranker returns a reranker by name.
func GetReranker(name string, openRouter *OpenRouterService) Reranker {
	switch strings.ToLower(name) {
	case "bm25":
		return NewBM25Reranker()
	case "llm":
		return NewLLMReranker(openRouter)
	case "hybrid":
		return NewHybridScoreReranker()
	default:
		return &NoOpReranker{}
	}
}
