package service

import (
	"sort"

	"github.com/integraltech/brainsentry/internal/domain"
)

// RRFConfig configures Reciprocal Rank Fusion scoring.
type RRFConfig struct {
	K              int     // RRF constant (default 60, higher = more uniform blending)
	VectorWeight   float64 // weight for vector search stream
	TextWeight     float64 // weight for full-text search stream
	GraphWeight    float64 // weight for graph search stream
	MaxPerSession  int     // max results per session for diversity (0 = no limit)
}

// DefaultRRFConfig returns standard RRF parameters.
func DefaultRRFConfig() RRFConfig {
	return RRFConfig{
		K:             60,
		VectorWeight:  0.5,
		TextWeight:    0.3,
		GraphWeight:   0.2,
		MaxPerSession: 3,
	}
}

// RRFResult holds a memory with its RRF score and source tracking.
type RRFResult struct {
	Memory      domain.Memory
	RRFScore    float64
	VectorRank  int // -1 if not in stream
	TextRank    int
	GraphRank   int
	SessionID   string
}

// ComputeRRF performs Reciprocal Rank Fusion across multiple ranked lists.
// Each list is a ranked slice of memories (best first).
// Returns merged results sorted by RRF score descending.
func ComputeRRF(
	vectorResults []domain.Memory,
	textResults []domain.Memory,
	graphResults []domain.Memory,
	config RRFConfig,
) []RRFResult {
	// Build score map
	scores := make(map[string]*RRFResult)

	addStream := func(results []domain.Memory, weight float64, setRank func(*RRFResult, int)) {
		for rank, m := range results {
			entry, exists := scores[m.ID]
			if !exists {
				entry = &RRFResult{
					Memory:     m,
					VectorRank: -1,
					TextRank:   -1,
					GraphRank:  -1,
				}
				scores[m.ID] = entry
			}
			rrfContribution := weight * (1.0 / float64(config.K+rank+1))
			entry.RRFScore += rrfContribution
			setRank(entry, rank)
		}
	}

	addStream(vectorResults, config.VectorWeight, func(r *RRFResult, rank int) { r.VectorRank = rank })
	addStream(textResults, config.TextWeight, func(r *RRFResult, rank int) { r.TextRank = rank })
	addStream(graphResults, config.GraphWeight, func(r *RRFResult, rank int) { r.GraphRank = rank })

	// Collect and sort
	results := make([]RRFResult, 0, len(scores))
	for _, r := range scores {
		results = append(results, *r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RRFScore > results[j].RRFScore
	})

	// Apply session diversity if configured
	if config.MaxPerSession > 0 {
		results = enforceSessionDiversity(results, config.MaxPerSession)
	}

	return results
}

// enforceSessionDiversity caps results per session to avoid skewing results.
func enforceSessionDiversity(results []RRFResult, maxPerSession int) []RRFResult {
	sessionCounts := make(map[string]int)
	var diverse []RRFResult

	for _, r := range results {
		sessionKey := r.Memory.CreatedBy // using CreatedBy as session proxy
		if sessionKey == "" {
			sessionKey = r.Memory.TenantID
		}

		if sessionCounts[sessionKey] < maxPerSession {
			diverse = append(diverse, r)
			sessionCounts[sessionKey]++
		}
	}

	return diverse
}
