package service

import (
	"math"
	"strings"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

// ScoringWeights defines the weights for each scoring component.
type ScoringWeights struct {
	Similarity     float64 `json:"similarity"`     // α: semantic similarity boost
	TokenOverlap   float64 `json:"tokenOverlap"`   // β: lexical token overlap
	GraphProximity float64 `json:"graphProximity"`  // γ: graph distance
	Recency        float64 `json:"recency"`         // δ: time recency
	TagMatch       float64 `json:"tagMatch"`        // ε: tag relevance
	Importance     float64 `json:"importance"`      // ζ: importance level
}

// DefaultScoringWeights provides balanced default weights.
var DefaultScoringWeights = ScoringWeights{
	Similarity:     0.30,
	TokenOverlap:   0.15,
	GraphProximity: 0.10,
	Recency:        0.15,
	TagMatch:       0.10,
	Importance:     0.20,
}

// ScoreTrace provides an explainable breakdown of the scoring computation.
type ScoreTrace struct {
	FinalScore      float64 `json:"finalScore"`
	SimBoost        float64 `json:"simBoost"`
	TokenOverlap    float64 `json:"tokenOverlap"`
	GraphProximity  float64 `json:"graphProximity"`
	RecencyScore    float64 `json:"recencyScore"`
	TagMatchScore   float64 `json:"tagMatchScore"`
	ImportanceScore float64 `json:"importanceScore"`
	DecayFactor     float64 `json:"decayFactor"`
	EmotionalBoost  float64 `json:"emotionalBoost"`
}

// ComputeHybridScore computes a composite score with explainable trace.
// cosineSim: cosine similarity from vector search (0-1)
// queryTokens: tokenized query for overlap calculation
// graphHops: number of hops from query seeds (-1 if not applicable)
// queryTags: tags from query context
func ComputeHybridScore(
	m *domain.Memory,
	cosineSim float64,
	queryTokens []string,
	graphHops int,
	queryTags []string,
	weights ScoringWeights,
) ScoreTrace {
	now := time.Now()
	trace := ScoreTrace{}

	// 1. Non-linear similarity boost: 1 - e^(-τ·sim) — more discriminative than raw cosine
	tau := 5.0
	trace.SimBoost = 1.0 - math.Exp(-tau*cosineSim)

	// 2. Token overlap (Jaccard-like)
	trace.TokenOverlap = computeTokenOverlap(m.Content, m.Summary, queryTokens)

	// 3. Graph proximity (penalize distant hops)
	if graphHops >= 0 {
		trace.GraphProximity = 1.0 / (1.0 + float64(graphHops))
	} else {
		trace.GraphProximity = 0.5 // neutral for non-graph results
	}

	// 4. Recency score (exponential decay of age)
	refTime := m.CreatedAt
	if m.LastAccessedAt != nil {
		refTime = *m.LastAccessedAt
	}
	ageDays := now.Sub(refTime).Hours() / 24.0
	if ageDays < 0 {
		ageDays = 0
	}
	trace.RecencyScore = math.Exp(-0.02 * ageDays) // half-life ~35 days

	// 5. Tag match score
	trace.TagMatchScore = computeTagMatch(m.Tags, queryTags)

	// 6. Importance multiplier
	switch m.Importance {
	case domain.ImportanceCritical:
		trace.ImportanceScore = 1.0
	case domain.ImportanceImportant:
		trace.ImportanceScore = 0.7
	default:
		trace.ImportanceScore = 0.4
	}

	// Decay factor from temporal decay
	rate := m.DecayRate
	if rate <= 0 {
		rate = GetDecayRate(m.MemoryType)
	}
	trace.DecayFactor = math.Exp(-rate * ageDays)

	// Emotional boost
	trace.EmotionalBoost = 1.0 + 0.3*math.Abs(m.EmotionalWeight)

	// Composite: sigmoid(weighted sum) * decay * emotional
	rawScore := weights.Similarity*trace.SimBoost +
		weights.TokenOverlap*trace.TokenOverlap +
		weights.GraphProximity*trace.GraphProximity +
		weights.Recency*trace.RecencyScore +
		weights.TagMatch*trace.TagMatchScore +
		weights.Importance*trace.ImportanceScore

	// Sigmoid normalization to 0-1 range
	trace.FinalScore = sigmoid(rawScore) * trace.DecayFactor * trace.EmotionalBoost

	return trace
}

// sigmoid maps a value to (0, 1) range.
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-6.0*(x-0.5)))
}

// computeTokenOverlap calculates Jaccard-like overlap between memory content and query tokens.
func computeTokenOverlap(content, summary string, queryTokens []string) float64 {
	if len(queryTokens) == 0 {
		return 0
	}

	text := strings.ToLower(content + " " + summary)
	matchCount := 0
	for _, qt := range queryTokens {
		if strings.Contains(text, strings.ToLower(qt)) {
			matchCount++
		}
	}
	return float64(matchCount) / float64(len(queryTokens))
}

// computeTagMatch calculates the proportion of query tags that match memory tags.
func computeTagMatch(memoryTags, queryTags []string) float64 {
	if len(queryTags) == 0 {
		return 0.5 // neutral when no query tags
	}
	if len(memoryTags) == 0 {
		return 0
	}

	tagSet := make(map[string]bool, len(memoryTags))
	for _, t := range memoryTags {
		tagSet[strings.ToLower(t)] = true
	}

	matchCount := 0
	for _, qt := range queryTags {
		if tagSet[strings.ToLower(qt)] {
			matchCount++
		}
	}
	return float64(matchCount) / float64(len(queryTags))
}

// TokenizeQuery splits a query into searchable tokens.
func TokenizeQuery(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	// Filter out stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "shall": true, "can": true, "to": true,
		"of": true, "in": true, "for": true, "on": true, "with": true,
		"at": true, "by": true, "from": true, "it": true, "this": true,
		"that": true, "and": true, "or": true, "but": true, "not": true,
		"if": true, "then": true, "so": true, "as": true, "up": true,
	}

	var tokens []string
	for _, w := range words {
		if len(w) >= 2 && !stopWords[w] {
			tokens = append(tokens, w)
		}
	}
	return tokens
}
