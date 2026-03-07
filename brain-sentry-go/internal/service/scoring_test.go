package service

import (
	"math"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestComputeHybridScore_HighSimilarity(t *testing.T) {
	m := &domain.Memory{
		Content:    "Go programming best practices",
		Importance: domain.ImportanceCritical,
		MemoryType: domain.MemoryTypeSemantic,
		CreatedAt:  time.Now(),
	}
	trace := ComputeHybridScore(m, 0.95, []string{"go", "programming"}, -1, nil, DefaultScoringWeights)
	if trace.FinalScore <= 0 {
		t.Error("expected positive score for high similarity")
	}
	if trace.SimBoost < 0.9 {
		t.Errorf("expected high sim boost for 0.95 cosine, got %.4f", trace.SimBoost)
	}
}

func TestComputeHybridScore_LowSimilarity(t *testing.T) {
	m := &domain.Memory{
		Content:    "Unrelated content about cooking",
		Importance: domain.ImportanceMinor,
		MemoryType: domain.MemoryTypeSemantic,
		CreatedAt:  time.Now(),
	}
	traceHigh := ComputeHybridScore(m, 0.95, nil, -1, nil, DefaultScoringWeights)
	traceLow := ComputeHybridScore(m, 0.1, nil, -1, nil, DefaultScoringWeights)
	if traceLow.FinalScore >= traceHigh.FinalScore {
		t.Errorf("low similarity (%.4f) should score less than high (%.4f)", traceLow.FinalScore, traceHigh.FinalScore)
	}
}

func TestComputeHybridScore_RecencyMatters(t *testing.T) {
	now := time.Now()
	recent := &domain.Memory{
		Content:    "Recent memory",
		Importance: domain.ImportanceImportant,
		MemoryType: domain.MemoryTypeSemantic,
		CreatedAt:  now,
	}
	old := &domain.Memory{
		Content:    "Old memory",
		Importance: domain.ImportanceImportant,
		MemoryType: domain.MemoryTypeSemantic,
		CreatedAt:  now.Add(-365 * 24 * time.Hour),
	}
	recentTrace := ComputeHybridScore(recent, 0.8, nil, -1, nil, DefaultScoringWeights)
	oldTrace := ComputeHybridScore(old, 0.8, nil, -1, nil, DefaultScoringWeights)
	if oldTrace.FinalScore >= recentTrace.FinalScore {
		t.Errorf("recent (%.4f) should score higher than old (%.4f)", recentTrace.FinalScore, oldTrace.FinalScore)
	}
}

func TestComputeHybridScore_GraphProximity(t *testing.T) {
	m := &domain.Memory{
		Content:    "Test memory",
		Importance: domain.ImportanceImportant,
		MemoryType: domain.MemoryTypeSemantic,
		CreatedAt:  time.Now(),
	}
	direct := ComputeHybridScore(m, 0.8, nil, 0, nil, DefaultScoringWeights)
	distant := ComputeHybridScore(m, 0.8, nil, 5, nil, DefaultScoringWeights)
	if distant.FinalScore >= direct.FinalScore {
		t.Errorf("direct graph (%.4f) should score higher than 5 hops (%.4f)", direct.FinalScore, distant.FinalScore)
	}
}

func TestComputeHybridScore_TagMatch(t *testing.T) {
	m := &domain.Memory{
		Content:    "Memory about Go",
		Tags:       []string{"go", "backend", "api"},
		Importance: domain.ImportanceImportant,
		MemoryType: domain.MemoryTypeSemantic,
		CreatedAt:  time.Now(),
	}
	withTags := ComputeHybridScore(m, 0.5, nil, -1, []string{"go", "api"}, DefaultScoringWeights)
	noTags := ComputeHybridScore(m, 0.5, nil, -1, []string{"python", "ml"}, DefaultScoringWeights)
	if noTags.FinalScore >= withTags.FinalScore {
		t.Errorf("matching tags (%.4f) should score higher than non-matching (%.4f)", withTags.FinalScore, noTags.FinalScore)
	}
}

func TestComputeHybridScore_ImportanceLevels(t *testing.T) {
	now := time.Now()
	critical := &domain.Memory{Content: "A", Importance: domain.ImportanceCritical, MemoryType: domain.MemoryTypeSemantic, CreatedAt: now}
	minor := &domain.Memory{Content: "A", Importance: domain.ImportanceMinor, MemoryType: domain.MemoryTypeSemantic, CreatedAt: now}
	cTrace := ComputeHybridScore(critical, 0.8, nil, -1, nil, DefaultScoringWeights)
	mTrace := ComputeHybridScore(minor, 0.8, nil, -1, nil, DefaultScoringWeights)
	if mTrace.FinalScore >= cTrace.FinalScore {
		t.Errorf("critical (%.4f) should score higher than minor (%.4f)", cTrace.FinalScore, mTrace.FinalScore)
	}
}

func TestComputeHybridScore_TraceHasAllFields(t *testing.T) {
	m := &domain.Memory{
		Content:         "Test",
		Tags:            []string{"test"},
		Importance:      domain.ImportanceImportant,
		MemoryType:      domain.MemoryTypeSemantic,
		EmotionalWeight: 0.5,
		CreatedAt:       time.Now(),
	}
	trace := ComputeHybridScore(m, 0.7, []string{"test"}, 1, []string{"test"}, DefaultScoringWeights)
	if trace.SimBoost <= 0 {
		t.Error("expected positive sim boost")
	}
	if trace.TokenOverlap <= 0 {
		t.Error("expected positive token overlap")
	}
	if trace.GraphProximity <= 0 {
		t.Error("expected positive graph proximity")
	}
	if trace.RecencyScore <= 0 {
		t.Error("expected positive recency")
	}
	if trace.TagMatchScore <= 0 {
		t.Error("expected positive tag match")
	}
	if trace.ImportanceScore <= 0 {
		t.Error("expected positive importance")
	}
	if trace.EmotionalBoost <= 1.0 {
		t.Error("expected emotional boost > 1.0")
	}
}

func TestTokenizeQuery(t *testing.T) {
	tokens := TokenizeQuery("How to implement authentication in Go?")
	if len(tokens) == 0 {
		t.Fatal("expected non-empty tokens")
	}
	// Stop words should be removed
	for _, tok := range tokens {
		if tok == "to" || tok == "in" || tok == "how" {
			// "how" is not a stop word, only "to" and "in" are
			if tok == "to" || tok == "in" {
				t.Errorf("stop word %q should be removed", tok)
			}
		}
	}
}

func TestTokenizeQuery_Empty(t *testing.T) {
	tokens := TokenizeQuery("")
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for empty query, got %d", len(tokens))
	}
}

func TestComputeTokenOverlap(t *testing.T) {
	overlap := computeTokenOverlap("Go programming best practices for backend", "", []string{"go", "backend"})
	if overlap != 1.0 {
		t.Errorf("expected 1.0 overlap, got %f", overlap)
	}
}

func TestComputeTagMatch(t *testing.T) {
	score := computeTagMatch([]string{"go", "api"}, []string{"go", "api", "db"})
	expected := 2.0 / 3.0
	if score < expected-0.01 || score > expected+0.01 {
		t.Errorf("expected %.2f, got %.2f", expected, score)
	}
}

func TestSigmoid_Midpoint(t *testing.T) {
	val := sigmoid(0.5)
	if val < 0.49 || val > 0.51 {
		t.Errorf("sigmoid(0.5) should be ~0.5, got %f", val)
	}
}

// --- Exact math verification tests ---

func TestSigmoid_KnownValues(t *testing.T) {
	// sigmoid(x) = 1 / (1 + exp(-6*(x-0.5)))
	tests := []struct {
		x    float64
		want float64
		tol  float64
	}{
		{0.0, 1.0 / (1.0 + math.Exp(3.0)), 0.001},   // ~0.0474
		{0.5, 0.5, 0.001},                              // exact midpoint
		{1.0, 1.0 / (1.0 + math.Exp(-3.0)), 0.001},   // ~0.9526
		{-1.0, 1.0 / (1.0 + math.Exp(9.0)), 0.0001},  // very small
		{2.0, 1.0 / (1.0 + math.Exp(-9.0)), 0.0001},  // very close to 1
	}
	for _, tt := range tests {
		got := sigmoid(tt.x)
		if math.Abs(got-tt.want) > tt.tol {
			t.Errorf("sigmoid(%f) = %f, want ~%f (tol %f)", tt.x, got, tt.want, tt.tol)
		}
	}
}

func TestSigmoid_IsMonotonic(t *testing.T) {
	prev := sigmoid(-10)
	for x := -9.0; x <= 10.0; x += 0.5 {
		curr := sigmoid(x)
		if curr < prev {
			t.Errorf("sigmoid is not monotonic: sigmoid(%f) = %f < sigmoid(%f) = %f", x, curr, x-0.5, prev)
		}
		prev = curr
	}
}

func TestSimBoost_Formula(t *testing.T) {
	// SimBoost = 1 - exp(-5 * cosineSim)
	tests := []struct {
		cosineSim float64
		want      float64
	}{
		{0.0, 0.0},
		{0.5, 1.0 - math.Exp(-2.5)},    // ~0.9179
		{1.0, 1.0 - math.Exp(-5.0)},     // ~0.9933
		{0.1, 1.0 - math.Exp(-0.5)},     // ~0.3935
	}
	for _, tt := range tests {
		m := &domain.Memory{
			Content:    "test",
			Importance: domain.ImportanceImportant,
			MemoryType: domain.MemoryTypeSemantic,
			CreatedAt:  time.Now(),
		}
		trace := ComputeHybridScore(m, tt.cosineSim, nil, -1, nil, DefaultScoringWeights)
		if math.Abs(trace.SimBoost-tt.want) > 0.001 {
			t.Errorf("SimBoost for cosineSim=%f: got %f, want %f", tt.cosineSim, trace.SimBoost, tt.want)
		}
	}
}

func TestImportanceScore_AllLevels(t *testing.T) {
	tests := []struct {
		importance domain.ImportanceLevel
		want       float64
	}{
		{domain.ImportanceCritical, 1.0},
		{domain.ImportanceImportant, 0.7},
		{domain.ImportanceMinor, 0.4},
		{domain.ImportanceLevel("UNKNOWN"), 0.4}, // falls into default
	}
	for _, tt := range tests {
		t.Run(string(tt.importance), func(t *testing.T) {
			m := &domain.Memory{
				Content:    "test",
				Importance: tt.importance,
				MemoryType: domain.MemoryTypeSemantic,
				CreatedAt:  time.Now(),
			}
			trace := ComputeHybridScore(m, 0.5, nil, -1, nil, DefaultScoringWeights)
			if trace.ImportanceScore != tt.want {
				t.Errorf("ImportanceScore for %s = %f, want %f", tt.importance, trace.ImportanceScore, tt.want)
			}
		})
	}
}

func TestComputeTagMatch_NilQueryTags(t *testing.T) {
	score := computeTagMatch([]string{"go"}, nil)
	if score != 0.5 {
		t.Errorf("expected 0.5 (neutral) for nil queryTags, got %f", score)
	}
}

func TestComputeTagMatch_EmptyMemoryTags(t *testing.T) {
	score := computeTagMatch(nil, []string{"go"})
	if score != 0.0 {
		t.Errorf("expected 0.0 for empty memory tags, got %f", score)
	}
}

func TestComputeTagMatch_CaseInsensitive(t *testing.T) {
	score := computeTagMatch([]string{"GO", "API"}, []string{"go", "api"})
	if score != 1.0 {
		t.Errorf("expected 1.0 for case-insensitive match, got %f", score)
	}
}

func TestComputeTagMatch_PartialMatch(t *testing.T) {
	score := computeTagMatch([]string{"go", "api", "db"}, []string{"go", "db", "python"})
	expected := 2.0 / 3.0
	if math.Abs(score-expected) > 0.001 {
		t.Errorf("expected %f, got %f", expected, score)
	}
}

func TestComputeTokenOverlap_PartialMatch(t *testing.T) {
	overlap := computeTokenOverlap("go backend api", "", []string{"go", "python"})
	if overlap != 0.5 {
		t.Errorf("expected 0.5 for 1/2 match, got %f", overlap)
	}
}

func TestComputeTokenOverlap_NoMatch(t *testing.T) {
	overlap := computeTokenOverlap("python ml", "", []string{"go", "java"})
	if overlap != 0.0 {
		t.Errorf("expected 0.0, got %f", overlap)
	}
}

func TestComputeTokenOverlap_EmptyQueryTokens(t *testing.T) {
	overlap := computeTokenOverlap("some content", "", []string{})
	if overlap != 0.0 {
		t.Errorf("expected 0.0 for empty tokens, got %f", overlap)
	}
}

func TestComputeTokenOverlap_NilQueryTokens(t *testing.T) {
	overlap := computeTokenOverlap("some content", "", nil)
	if overlap != 0.0 {
		t.Errorf("expected 0.0 for nil tokens, got %f", overlap)
	}
}

func TestGraphProximity_ZeroHops(t *testing.T) {
	m := &domain.Memory{Content: "test", Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now()}
	trace := ComputeHybridScore(m, 0.5, nil, 0, nil, DefaultScoringWeights)
	if trace.GraphProximity != 1.0 {
		t.Errorf("expected 1.0 for 0 hops, got %f", trace.GraphProximity)
	}
}

func TestGraphProximity_NegativeHops_Neutral(t *testing.T) {
	m := &domain.Memory{Content: "test", Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now()}
	trace := ComputeHybridScore(m, 0.5, nil, -1, nil, DefaultScoringWeights)
	if trace.GraphProximity != 0.5 {
		t.Errorf("expected 0.5 (neutral) for -1 hops, got %f", trace.GraphProximity)
	}
}

func TestGraphProximity_MultipleHops(t *testing.T) {
	m := &domain.Memory{Content: "test", Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now()}
	tests := []struct {
		hops int
		want float64
	}{
		{0, 1.0},
		{1, 0.5},
		{2, 1.0 / 3.0},
		{5, 1.0 / 6.0},
	}
	for _, tt := range tests {
		trace := ComputeHybridScore(m, 0.5, nil, tt.hops, nil, DefaultScoringWeights)
		if math.Abs(trace.GraphProximity-tt.want) > 0.001 {
			t.Errorf("GraphProximity(%d hops) = %f, want %f", tt.hops, trace.GraphProximity, tt.want)
		}
	}
}

func TestEmotionalBoost_Negative(t *testing.T) {
	m := &domain.Memory{Content: "test", Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now(), EmotionalWeight: -0.8}
	trace := ComputeHybridScore(m, 0.5, nil, -1, nil, DefaultScoringWeights)
	expected := 1.0 + 0.3*0.8 // uses abs
	if math.Abs(trace.EmotionalBoost-expected) > 0.001 {
		t.Errorf("EmotionalBoost = %f, want %f", trace.EmotionalBoost, expected)
	}
}

func TestEmotionalBoost_Zero(t *testing.T) {
	m := &domain.Memory{Content: "test", Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now(), EmotionalWeight: 0}
	trace := ComputeHybridScore(m, 0.5, nil, -1, nil, DefaultScoringWeights)
	if trace.EmotionalBoost != 1.0 {
		t.Errorf("EmotionalBoost = %f, want 1.0", trace.EmotionalBoost)
	}
}

func TestComputeHybridScore_CustomWeights_SimilarityOnly(t *testing.T) {
	m := &domain.Memory{Content: "test", Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now()}
	// Only similarity weight
	weights := ScoringWeights{Similarity: 1.0, TokenOverlap: 0, GraphProximity: 0, Recency: 0, TagMatch: 0, Importance: 0}
	traceHigh := ComputeHybridScore(m, 0.95, nil, -1, nil, weights)
	traceLow := ComputeHybridScore(m, 0.1, nil, -1, nil, weights)
	if traceLow.FinalScore >= traceHigh.FinalScore {
		t.Errorf("with similarity-only weights, high sim (%f) should beat low sim (%f)", traceHigh.FinalScore, traceLow.FinalScore)
	}
}

func TestTokenizeQuery_StopWordRemoval(t *testing.T) {
	tokens := TokenizeQuery("the to in for on with at by from")
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens (all stop words), got %d: %v", len(tokens), tokens)
	}
}

func TestTokenizeQuery_MinLengthFilter(t *testing.T) {
	tokens := TokenizeQuery("a go programming")
	// "a" is 1 char (filtered), "go" is 2 chars (kept), "programming" is 11 chars (kept)
	if len(tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
}

func TestDefaultScoringWeights_SumCloseToOne(t *testing.T) {
	w := DefaultScoringWeights
	sum := w.Similarity + w.TokenOverlap + w.GraphProximity + w.Recency + w.TagMatch + w.Importance
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("default weights sum = %f, expected ~1.0", sum)
	}
}
