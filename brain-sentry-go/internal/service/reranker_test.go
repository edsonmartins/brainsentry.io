package service

import (
	"context"
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
