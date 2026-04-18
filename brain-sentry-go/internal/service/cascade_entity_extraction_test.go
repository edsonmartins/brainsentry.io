package service

import (
	"context"
	"testing"
)

// multiResponseMock returns a different response per call, cycling.
type multiResponseMock struct {
	responses []string
	calls     int
}

func (m *multiResponseMock) Name() string { return "multi" }
func (m *multiResponseMock) Chat(ctx context.Context, msgs []ChatMessage) (string, error) {
	idx := m.calls
	m.calls++
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	return m.responses[idx], nil
}

func TestCascadeExtraction_NoLLM(t *testing.T) {
	s := NewCascadeEntityExtractionService(nil)
	result, err := s.Extract(context.Background(), "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entities) != 0 || len(result.Relationships) != 0 {
		t.Error("expected empty result when LLM is nil")
	}
}

func TestCascadeExtraction_NodesOnly(t *testing.T) {
	// Only one entity → cascade stops after pass 1
	mock := &multiResponseMock{
		responses: []string{
			`{"entities":[{"name":"PostgreSQL","type":"TECHNOLOGY"}]}`,
		},
	}
	s := NewCascadeEntityExtractionService(mock)

	result, err := s.Extract(context.Background(), "PostgreSQL is a database.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(result.Entities))
	}
	if len(result.Relationships) != 0 {
		t.Errorf("expected no relationships, got %d", len(result.Relationships))
	}
	if result.PassCount != 1 {
		t.Errorf("expected passCount 1, got %d", result.PassCount)
	}
}

func TestCascadeExtraction_FullPipeline(t *testing.T) {
	// Response 1: nodes
	// Response 2: edges (source/target pairs)
	// Response 3+: relationship name per edge
	mock := &multiResponseMock{
		responses: []string{
			`{"entities":[
				{"name":"Go","type":"LANGUAGE"},
				{"name":"PostgreSQL","type":"TECHNOLOGY"}
			]}`,
			`{"edges":[{"source":"Go","target":"PostgreSQL"}]}`,
			`{"relationship":"connects_to"}`,
		},
	}
	s := NewCascadeEntityExtractionService(mock)

	result, err := s.Extract(context.Background(), "Go connects to PostgreSQL.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(result.Entities))
	}
	if len(result.Relationships) != 1 {
		t.Errorf("expected 1 relationship, got %d", len(result.Relationships))
	}
	if result.PassCount != 3 {
		t.Errorf("expected passCount 3, got %d", result.PassCount)
	}

	rel := result.Relationships[0]
	if rel.Source != "Go" || rel.Target != "PostgreSQL" {
		t.Errorf("unexpected rel: %+v", rel)
	}
	if rel.Type != "connects_to" {
		t.Errorf("expected connects_to, got %s", rel.Type)
	}
}

func TestCascadeExtraction_FiltersInvalidEdges(t *testing.T) {
	// Pass 2 suggests edge with entity not in pass 1 → should be filtered
	mock := &multiResponseMock{
		responses: []string{
			`{"entities":[{"name":"Go","type":"LANGUAGE"}]}`,
			`{"edges":[{"source":"Go","target":"UnknownEntity"}]}`,
			`{"relationship":"x"}`,
		},
	}
	s := NewCascadeEntityExtractionService(mock)

	result, err := s.Extract(context.Background(), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only 1 entity → cascade stops at pass 1 (len<2)
	if len(result.Relationships) != 0 {
		t.Errorf("expected 0 relationships (only 1 entity), got %d", len(result.Relationships))
	}
}

func TestCascadeExtraction_EdgeSelfLoopFiltered(t *testing.T) {
	mock := &multiResponseMock{
		responses: []string{
			`{"entities":[{"name":"A","type":"X"},{"name":"B","type":"X"}]}`,
			`{"edges":[{"source":"A","target":"A"},{"source":"A","target":"B"}]}`,
			`{"relationship":"related_to"}`,
		},
	}
	s := NewCascadeEntityExtractionService(mock)

	result, err := s.Extract(context.Background(), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Self-loop should be filtered, only A→B remains
	if len(result.Relationships) != 1 {
		t.Errorf("expected 1 relationship (self-loop filtered), got %d", len(result.Relationships))
	}
}

func TestCascadeExtraction_DedupsEntities(t *testing.T) {
	mock := &multiResponseMock{
		responses: []string{
			`{"entities":[
				{"name":"Go","type":"LANGUAGE"},
				{"name":"go","type":"LANGUAGE"},
				{"name":"Go","type":"LANGUAGE"}
			]}`,
		},
	}
	s := NewCascadeEntityExtractionService(mock)

	result, err := s.Extract(context.Background(), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entities) != 1 {
		t.Errorf("expected dedup to 1 entity, got %d", len(result.Entities))
	}
}

func TestCascadeExtraction_NormalizesRelationshipName(t *testing.T) {
	mock := &multiResponseMock{
		responses: []string{
			`{"entities":[{"name":"A","type":"X"},{"name":"B","type":"X"}]}`,
			`{"edges":[{"source":"A","target":"B"}]}`,
			`{"relationship":"Depends On"}`,
		},
	}
	s := NewCascadeEntityExtractionService(mock)

	result, err := s.Extract(context.Background(), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Relationships) != 1 {
		t.Fatalf("expected 1 relationship")
	}
	if result.Relationships[0].Type != "depends_on" {
		t.Errorf("expected depends_on (normalized), got %s", result.Relationships[0].Type)
	}
}
