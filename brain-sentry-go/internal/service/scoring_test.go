package service

import (
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
