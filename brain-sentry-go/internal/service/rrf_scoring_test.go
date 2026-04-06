package service

import (
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func makeMemory(id string) domain.Memory {
	return domain.Memory{ID: id, Content: "test " + id}
}

func TestComputeRRF_SingleStream(t *testing.T) {
	vector := []domain.Memory{makeMemory("a"), makeMemory("b"), makeMemory("c")}
	config := DefaultRRFConfig()
	config.MaxPerSession = 0

	results := ComputeRRF(vector, nil, nil, config)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Memory.ID != "a" {
		t.Errorf("expected 'a' first, got %s", results[0].Memory.ID)
	}
	if results[0].VectorRank != 0 {
		t.Errorf("expected VectorRank 0, got %d", results[0].VectorRank)
	}
}

func TestComputeRRF_MultiStream(t *testing.T) {
	vector := []domain.Memory{makeMemory("a"), makeMemory("b")}
	text := []domain.Memory{makeMemory("b"), makeMemory("c")}
	graph := []domain.Memory{makeMemory("c"), makeMemory("a")}

	config := DefaultRRFConfig()
	config.MaxPerSession = 0

	results := ComputeRRF(vector, text, graph, config)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// All memories should appear in results
	ids := make(map[string]bool)
	for _, r := range results {
		ids[r.Memory.ID] = true
	}
	for _, id := range []string{"a", "b", "c"} {
		if !ids[id] {
			t.Errorf("missing memory %s in results", id)
		}
	}

	// Results that appear in multiple streams should score higher
	for _, r := range results {
		if r.RRFScore <= 0 {
			t.Errorf("expected positive RRF score for %s, got %f", r.Memory.ID, r.RRFScore)
		}
	}
}

func TestComputeRRF_SessionDiversity(t *testing.T) {
	m1 := domain.Memory{ID: "1", Content: "test", CreatedBy: "session-a"}
	m2 := domain.Memory{ID: "2", Content: "test", CreatedBy: "session-a"}
	m3 := domain.Memory{ID: "3", Content: "test", CreatedBy: "session-a"}
	m4 := domain.Memory{ID: "4", Content: "test", CreatedBy: "session-a"}
	m5 := domain.Memory{ID: "5", Content: "test", CreatedBy: "session-b"}

	vector := []domain.Memory{m1, m2, m3, m4, m5}
	config := RRFConfig{K: 60, VectorWeight: 1.0, MaxPerSession: 2}

	results := ComputeRRF(vector, nil, nil, config)

	sessionACount := 0
	for _, r := range results {
		if r.Memory.CreatedBy == "session-a" {
			sessionACount++
		}
	}

	if sessionACount > 2 {
		t.Errorf("expected max 2 from session-a, got %d", sessionACount)
	}
}

func TestDefaultRRFConfig(t *testing.T) {
	c := DefaultRRFConfig()
	total := c.VectorWeight + c.TextWeight + c.GraphWeight
	if total < 0.99 || total > 1.01 {
		t.Errorf("weights should sum to 1.0, got %f", total)
	}
	if c.K != 60 {
		t.Errorf("expected K=60, got %d", c.K)
	}
	if c.MaxPerSession != 3 {
		t.Errorf("expected MaxPerSession=3, got %d", c.MaxPerSession)
	}
}
