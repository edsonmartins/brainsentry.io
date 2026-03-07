package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

func makeTestMemories() []domain.Memory {
	return []domain.Memory{
		{ID: "1", Content: "Go is a compiled language for backend development", Summary: "Go backend", Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now()},
		{ID: "2", Content: "Python is great for data science and machine learning", Summary: "Python ML", Importance: domain.ImportanceMinor, MemoryType: domain.MemoryTypeSemantic, CreatedAt: time.Now()},
		{ID: "3", Content: "PostgreSQL database optimization and indexing strategies", Summary: "PostgreSQL optimization", Importance: domain.ImportanceCritical, MemoryType: domain.MemoryTypeProcedural, CreatedAt: time.Now()},
	}
}

func TestNoOpReranker(t *testing.T) {
	r := &NoOpReranker{}
	if r.Name() != "noop" {
		t.Error("expected noop name")
	}

	memories := makeTestMemories()
	results, err := r.Rerank(context.Background(), "test", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != len(memories) {
		t.Errorf("expected %d results, got %d", len(memories), len(results))
	}
	// Order should be preserved
	if results[0].Memory.ID != "1" {
		t.Error("expected order preserved")
	}
}

func TestBM25Reranker_RanksRelevant(t *testing.T) {
	r := NewBM25Reranker()
	if r.Name() != "bm25" {
		t.Error("expected bm25 name")
	}

	memories := makeTestMemories()
	results, err := r.Rerank(context.Background(), "Go backend development", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// Go memory should be ranked first
	if results[0].Memory.ID != "1" {
		t.Errorf("expected Go memory first, got %s", results[0].Memory.ID)
	}
}

func TestBM25Reranker_EmptyQuery(t *testing.T) {
	r := NewBM25Reranker()
	memories := makeTestMemories()
	results, err := r.Rerank(context.Background(), "", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestBM25Reranker_PostgreSQLQuery(t *testing.T) {
	r := NewBM25Reranker()
	memories := makeTestMemories()
	results, err := r.Rerank(context.Background(), "PostgreSQL indexing", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Memory.ID != "3" {
		t.Errorf("expected PostgreSQL memory first, got %s (score=%.4f)", results[0].Memory.ID, results[0].Score)
	}
}

func TestLLMReranker_FallbackWithoutOpenRouter(t *testing.T) {
	r := NewLLMReranker(nil)
	if r.Name() != "llm" {
		t.Error("expected llm name")
	}

	memories := makeTestMemories()
	results, err := r.Rerank(context.Background(), "test", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestHybridScoreReranker(t *testing.T) {
	r := NewHybridScoreReranker()
	if r.Name() != "hybrid" {
		t.Error("expected hybrid name")
	}

	memories := makeTestMemories()
	results, err := r.Rerank(context.Background(), "Go backend", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	// All should have positive scores
	for _, r := range results {
		if r.Score <= 0 {
			t.Errorf("expected positive score for %s, got %f", r.Memory.ID, r.Score)
		}
	}
}

func TestGetReranker_ByName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"bm25", "bm25"},
		{"llm", "llm"},
		{"hybrid", "hybrid"},
		{"noop", "noop"},
		{"unknown", "noop"},
		{"", "noop"},
	}

	for _, tt := range tests {
		r := GetReranker(tt.name, nil)
		if r.Name() != tt.expected {
			t.Errorf("GetReranker(%q) = %s, want %s", tt.name, r.Name(), tt.expected)
		}
	}
}

func TestRankedMemory_Structure(t *testing.T) {
	rm := RankedMemory{
		Memory: domain.Memory{ID: "m1"},
		Score:  0.85,
		Reason: "highly relevant",
	}
	if rm.Score != 0.85 {
		t.Error("expected 0.85")
	}
	if rm.Reason != "highly relevant" {
		t.Error("expected reason")
	}
}

// --- Extended reranker tests ---

func TestBM25Reranker_Parameters(t *testing.T) {
	r := NewBM25Reranker()
	if r.k1 != 1.2 {
		t.Errorf("expected k1=1.2, got %f", r.k1)
	}
	if r.b != 0.75 {
		t.Errorf("expected b=0.75, got %f", r.b)
	}
}

func TestBM25Reranker_SingleDocument(t *testing.T) {
	r := NewBM25Reranker()
	memories := []domain.Memory{
		{ID: "1", Content: "Go programming language", Summary: "Go lang", CreatedAt: time.Now()},
	}
	results, err := r.Rerank(context.Background(), "Go programming", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score for matching document, got %f", results[0].Score)
	}
}

func TestBM25Reranker_DocumentNotContainingTerm(t *testing.T) {
	r := NewBM25Reranker()
	memories := []domain.Memory{
		{ID: "1", Content: "Python machine learning", Summary: "Python ML", CreatedAt: time.Now()},
	}
	results, err := r.Rerank(context.Background(), "Rust systems programming", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Score != 0 {
		t.Errorf("expected score 0 for non-matching document, got %f", results[0].Score)
	}
}

func TestBM25Reranker_LengthNormalization(t *testing.T) {
	r := NewBM25Reranker()
	// Short doc and long doc both contain "Go" once
	short := domain.Memory{ID: "short", Content: "Go backend", Summary: "", CreatedAt: time.Now()}
	long := domain.Memory{ID: "long", Content: "Go " + strings.Repeat("other words ", 50), Summary: "", CreatedAt: time.Now()}
	results, err := r.Rerank(context.Background(), "Go backend", []domain.Memory{long, short})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Short doc should score higher due to length normalization (b=0.75)
	if results[0].Memory.ID != "short" {
		t.Errorf("shorter document should rank higher, got %s first", results[0].Memory.ID)
	}
}

func TestBM25Reranker_OrderIsDescending(t *testing.T) {
	r := NewBM25Reranker()
	memories := makeTestMemories()
	results, err := r.Rerank(context.Background(), "Go backend development programming", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not in descending order: [%d]=%f > [%d]=%f", i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestBM25Reranker_SummaryContribution(t *testing.T) {
	r := NewBM25Reranker()
	// Term is only in summary, not in content
	memories := []domain.Memory{
		{ID: "1", Content: "something else entirely", Summary: "Go backend", CreatedAt: time.Now()},
		{ID: "2", Content: "Python ML", Summary: "Python ML", CreatedAt: time.Now()},
	}
	results, err := r.Rerank(context.Background(), "Go backend", memories)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Memory.ID != "1" {
		t.Error("memory with matching summary should rank first")
	}
}

func TestHybridScoreReranker_UsesDefaultWeights(t *testing.T) {
	r := NewHybridScoreReranker()
	if r.weights.Similarity != DefaultScoringWeights.Similarity {
		t.Error("expected default weights")
	}
}

func TestLLMReranker_FallbackWithEmptyMemories(t *testing.T) {
	r := NewLLMReranker(nil)
	results, err := r.Rerank(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
