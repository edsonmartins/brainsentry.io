package service

import (
	"testing"
)

func TestNewNLCypherService(t *testing.T) {
	svc := NewNLCypherService(nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.maxRetries != 3 {
		t.Errorf("expected 3 max retries, got %d", svc.maxRetries)
	}
}

func TestNLQueryResult_Structure(t *testing.T) {
	result := NLQueryResult{
		Query:           "find all Go memories",
		GeneratedCypher: "MATCH (m:Memory) WHERE m.tenantId = 't1' RETURN m.id",
		Results:         []map[string]any{{"id": "m1"}},
		Attempts:        1,
		Success:         true,
	}
	if result.Query != "find all Go memories" {
		t.Error("expected correct query")
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Attempts != 1 {
		t.Error("expected 1 attempt")
	}
	if len(result.Results) != 1 {
		t.Error("expected 1 result")
	}
}

func TestNLQueryResult_Failure(t *testing.T) {
	result := NLQueryResult{
		Query:        "complex query",
		Attempts:     3,
		Success:      false,
		ErrorMessage: "failed after 3 attempts",
	}
	if result.Success {
		t.Error("expected failure")
	}
	if result.ErrorMessage == "" {
		t.Error("expected error message")
	}
}

func TestGraphSchema_NotEmpty(t *testing.T) {
	if graphSchema == "" {
		t.Error("expected non-empty graph schema")
	}
	if !containsSubstr(graphSchema, "Memory") {
		t.Error("expected Memory in schema")
	}
	if !containsSubstr(graphSchema, "RELATED_TO") {
		t.Error("expected RELATED_TO in schema")
	}
}

func TestQueryNaturalLanguage_NilDeps(t *testing.T) {
	svc := NewNLCypherService(nil, nil)
	result, err := svc.QueryNaturalLanguage(nil, "test query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected failure without dependencies")
	}
	if result.ErrorMessage == "" {
		t.Error("expected error message")
	}
}
