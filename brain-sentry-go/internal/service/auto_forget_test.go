package service

import (
	"testing"
)

func TestJaccardSimilarity_Identical(t *testing.T) {
	a := map[string]bool{"hello": true, "world": true}
	b := map[string]bool{"hello": true, "world": true}

	sim := jaccardSimilarity(a, b)
	if sim != 1.0 {
		t.Errorf("expected 1.0, got %f", sim)
	}
}

func TestJaccardSimilarity_Disjoint(t *testing.T) {
	a := map[string]bool{"hello": true}
	b := map[string]bool{"world": true}

	sim := jaccardSimilarity(a, b)
	if sim != 0.0 {
		t.Errorf("expected 0.0, got %f", sim)
	}
}

func TestJaccardSimilarity_Partial(t *testing.T) {
	a := map[string]bool{"hello": true, "world": true, "foo": true}
	b := map[string]bool{"hello": true, "world": true, "bar": true}

	sim := jaccardSimilarity(a, b)
	// intersection=2, union=4 → 0.5
	if sim != 0.5 {
		t.Errorf("expected 0.5, got %f", sim)
	}
}

func TestJaccardSimilarity_Empty(t *testing.T) {
	a := map[string]bool{}
	b := map[string]bool{}

	sim := jaccardSimilarity(a, b)
	if sim != 1.0 {
		t.Errorf("expected 1.0 for empty sets, got %f", sim)
	}
}

func TestTokenizeForJaccard(t *testing.T) {
	tokens := tokenizeForJaccard("Hello World! This is a test.")
	if !tokens["hello"] {
		t.Error("expected 'hello' token")
	}
	if !tokens["world"] {
		t.Error("expected 'world' token")
	}
	if !tokens["test"] {
		t.Error("expected 'test' token")
	}
	// Single char words should be excluded
	if tokens["a"] {
		t.Error("single char 'a' should be excluded")
	}
}

func TestDefaultAutoForgetConfig(t *testing.T) {
	cfg := DefaultAutoForgetConfig()
	if !cfg.TTLEnabled {
		t.Error("TTL should be enabled by default")
	}
	if !cfg.ContradictionEnabled {
		t.Error("Contradiction detection should be enabled by default")
	}
	if cfg.ContradictionThreshold != 0.9 {
		t.Errorf("expected threshold 0.9, got %f", cfg.ContradictionThreshold)
	}
	if cfg.MaxDeletesPerRun != 50 {
		t.Errorf("expected max 50 deletes, got %d", cfg.MaxDeletesPerRun)
	}
}
