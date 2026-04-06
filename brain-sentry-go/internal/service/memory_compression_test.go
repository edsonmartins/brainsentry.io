package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestMemoryCompressionService_CompressWithLLM(t *testing.T) {
	mockResponse := `{
		"facts": ["PostgreSQL uses pgvector for embeddings", "Vector search uses cosine similarity"],
		"concepts": ["PostgreSQL", "pgvector", "cosine similarity"],
		"narrative": "The system uses PostgreSQL with pgvector extension for embedding storage and cosine similarity search.",
		"importance": 7,
		"title": "PostgreSQL vector search setup",
		"files": ["internal/repository/memory.go"]
	}`

	mock := &mockLLMProvider{name: "test", response: mockResponse}
	svc := NewMemoryCompressionService(mock)

	data, err := svc.Compress(context.Background(), "We set up PostgreSQL with pgvector for vector search using cosine similarity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Facts) != 2 {
		t.Errorf("expected 2 facts, got %d", len(data.Facts))
	}
	if len(data.Concepts) != 3 {
		t.Errorf("expected 3 concepts, got %d", len(data.Concepts))
	}
	if data.Importance != 7 {
		t.Errorf("expected importance 7, got %d", data.Importance)
	}
	if data.Title != "PostgreSQL vector search setup" {
		t.Errorf("unexpected title: %s", data.Title)
	}
	if len(data.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(data.Files))
	}
}

func TestMemoryCompressionService_FallbackWithoutLLM(t *testing.T) {
	svc := NewMemoryCompressionService(nil)

	data, err := svc.Compress(context.Background(), "PostgreSQL uses pgvector for embeddings")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Importance != 5 {
		t.Errorf("expected default importance 5, got %d", data.Importance)
	}
	if len(data.Facts) == 0 {
		t.Error("expected at least one fact in fallback")
	}
	if data.Narrative == "" {
		t.Error("expected non-empty narrative in fallback")
	}
}

func TestMemoryCompressionService_SelfCorrection(t *testing.T) {
	callCount := 0
	mock := &mockLLMProvider{name: "test"}

	// First call returns invalid JSON, second returns valid
	origChat := mock.Chat
	_ = origChat
	invalidThenValid := &selfCorrectingMock{
		responses: []string{
			`not valid json at all`,
			`{"facts": ["test fact"], "concepts": ["test"], "narrative": "test narrative", "importance": 5, "title": "test"}`,
		},
	}

	svc := NewMemoryCompressionService(invalidThenValid)
	data, err := svc.Compress(context.Background(), "test content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Importance != 5 {
		t.Errorf("expected importance 5, got %d", data.Importance)
	}
	if invalidThenValid.callCount != 2 {
		t.Errorf("expected 2 calls (retry), got %d", invalidThenValid.callCount)
	}
	_ = callCount
}

type selfCorrectingMock struct {
	responses []string
	callCount int
}

func (m *selfCorrectingMock) Name() string { return "self-correcting" }
func (m *selfCorrectingMock) Chat(ctx context.Context, msgs []ChatMessage) (string, error) {
	idx := m.callCount
	m.callCount++
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return m.responses[len(m.responses)-1], nil
}

func TestMemoryCompressionService_EnrichMemory(t *testing.T) {
	svc := NewMemoryCompressionService(nil)

	m := &domain.Memory{
		Content: "test content",
		Tags:    []string{"existing"},
	}

	data := &CompressedMemoryData{
		Facts:      []string{"fact1"},
		Concepts:   []string{"Go", "PostgreSQL"},
		Narrative:  "A test narrative",
		Importance: 8,
		Title:      "Test Title",
	}

	svc.EnrichMemory(m, data)

	if m.Summary != "A test narrative" {
		t.Errorf("expected narrative as summary, got: %s", m.Summary)
	}

	// Check metadata
	var meta map[string]any
	if err := json.Unmarshal(m.Metadata, &meta); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}

	if meta["compressionTitle"] != "Test Title" {
		t.Errorf("expected compressionTitle in metadata")
	}

	// Check tags include concepts
	if len(m.Tags) != 3 { // existing + go + postgresql
		t.Errorf("expected 3 tags, got %d: %v", len(m.Tags), m.Tags)
	}
}
