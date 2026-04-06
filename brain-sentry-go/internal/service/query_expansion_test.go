package service

import (
	"context"
	"testing"
)

func TestQueryExpansionService_ExpandWithLLM(t *testing.T) {
	mockResponse := `{
		"reformulations": [
			"how does PostgreSQL pgvector search work",
			"vector similarity search PostgreSQL",
			"cosine similarity embedding lookup"
		],
		"entities": ["PostgreSQL", "pgvector"]
	}`

	mock := &mockLLMProvider{name: "test", response: mockResponse}
	svc := NewQueryExpansionService(mock)

	result := svc.Expand(context.Background(), "how does vector search work")

	if result.Original != "how does vector search work" {
		t.Errorf("expected original query preserved, got: %s", result.Original)
	}
	if len(result.Reformulations) != 3 {
		t.Errorf("expected 3 reformulations, got %d", len(result.Reformulations))
	}
	if len(result.Entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(result.Entities))
	}
}

func TestQueryExpansionService_FallbackWithoutLLM(t *testing.T) {
	svc := NewQueryExpansionService(nil)

	result := svc.Expand(context.Background(), "how does the vector search algorithm work")

	if result.Original != "how does the vector search algorithm work" {
		t.Errorf("expected original query preserved")
	}
	// Should have at least one reformulation (stop words removed)
	if len(result.Reformulations) == 0 {
		t.Error("expected at least one fallback reformulation")
	}
}

func TestQueryExpansionService_ShortQuery(t *testing.T) {
	svc := NewQueryExpansionService(nil)

	result := svc.Expand(context.Background(), "pgvector")

	if result.Original != "pgvector" {
		t.Errorf("expected original query preserved")
	}
	// Short query — no reformulations
	if len(result.Reformulations) != 0 {
		t.Errorf("expected no reformulations for short query, got %d", len(result.Reformulations))
	}
}

func TestQueryExpansionService_CapsReformulations(t *testing.T) {
	mockResponse := `{
		"reformulations": ["a", "b", "c", "d", "e", "f", "g"],
		"entities": []
	}`

	mock := &mockLLMProvider{name: "test", response: mockResponse}
	svc := NewQueryExpansionService(mock)

	result := svc.Expand(context.Background(), "test query with many words")

	if len(result.Reformulations) > 5 {
		t.Errorf("expected max 5 reformulations, got %d", len(result.Reformulations))
	}
}
