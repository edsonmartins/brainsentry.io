package service

import (
	"context"
	"testing"
)

func TestTripletExtraction_NoLLM(t *testing.T) {
	s := NewTripletExtractionService(nil)
	triplets, err := s.ExtractFromContent(context.Background(), "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if triplets != nil {
		t.Errorf("expected nil triplets when LLM is nil, got %d", len(triplets))
	}
}

func TestTripletExtraction_EmptyContent(t *testing.T) {
	mock := &mockLLMProvider{name: "test", response: `{"triplets":[]}`}
	s := NewTripletExtractionService(mock)

	triplets, err := s.ExtractFromContent(context.Background(), "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(triplets) != 0 {
		t.Errorf("expected 0 triplets for empty content, got %d", len(triplets))
	}
	if mock.callCount != 0 {
		t.Errorf("should not call LLM for empty content, got %d calls", mock.callCount)
	}
}

func TestTripletExtraction_ValidResponse(t *testing.T) {
	response := `{
		"triplets": [
			{"subject": "PostgreSQL", "predicate": "supports", "object": "JSON", "confidence": 0.95},
			{"subject": "Redis", "predicate": "used_for", "object": "caching", "confidence": 0.9}
		]
	}`
	mock := &mockLLMProvider{name: "test", response: response}
	s := NewTripletExtractionService(mock)

	triplets, err := s.ExtractFromContent(context.Background(), "PostgreSQL supports JSON. Redis is used for caching.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(triplets) != 2 {
		t.Errorf("expected 2 triplets, got %d", len(triplets))
	}
	if triplets[0].Subject != "PostgreSQL" {
		t.Errorf("unexpected subject: %s", triplets[0].Subject)
	}
}

func TestTripletExtraction_FiltersInvalid(t *testing.T) {
	response := `{
		"triplets": [
			{"subject": "PostgreSQL", "predicate": "supports", "object": "JSON", "confidence": 0.9},
			{"subject": "", "predicate": "x", "object": "y", "confidence": 0.5},
			{"subject": "x", "predicate": "", "object": "y", "confidence": 0.5}
		]
	}`
	mock := &mockLLMProvider{name: "test", response: response}
	s := NewTripletExtractionService(mock)

	triplets, err := s.ExtractFromContent(context.Background(), "some text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(triplets) != 1 {
		t.Errorf("expected 1 valid triplet after filtering, got %d", len(triplets))
	}
}

func TestTripletExtraction_ClampsConfidence(t *testing.T) {
	response := `{
		"triplets": [
			{"subject": "A", "predicate": "x", "object": "B", "confidence": 0},
			{"subject": "C", "predicate": "y", "object": "D", "confidence": 1.5}
		]
	}`
	mock := &mockLLMProvider{name: "test", response: response}
	s := NewTripletExtractionService(mock)

	triplets, err := s.ExtractFromContent(context.Background(), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both should default to 0.5 due to invalid confidence
	for _, tr := range triplets {
		if tr.Confidence != 0.5 {
			t.Errorf("invalid confidence should default to 0.5, got %f", tr.Confidence)
		}
	}
}

func TestTripletExtraction_BuildTriplets(t *testing.T) {
	mock := &mockLLMProvider{name: "test"}
	s := NewTripletExtractionService(mock)

	extracted := []ExtractedTriplet{
		{Subject: "A", Predicate: "uses", Object: "B", Confidence: 0.9},
		{Subject: "A", Predicate: "uses", Object: "B", Confidence: 0.7}, // dup
	}

	triplets := s.BuildTriplets(context.Background(), "mem-1", extracted)

	if len(triplets) != 1 {
		t.Errorf("expected dedup to 1 triplet, got %d", len(triplets))
	}

	if triplets[0].MemoryID != "mem-1" {
		t.Errorf("expected memoryId mem-1, got %s", triplets[0].MemoryID)
	}
	if triplets[0].Text != "A→uses→B" {
		t.Errorf("expected text A→uses→B, got %s", triplets[0].Text)
	}
	if triplets[0].FeedbackWeight != 0.5 {
		t.Errorf("expected neutral feedback weight 0.5, got %f", triplets[0].FeedbackWeight)
	}
	if triplets[0].ID == "" {
		t.Error("expected non-empty UUID")
	}
}

func TestTripletExtraction_ExtractAndBuild(t *testing.T) {
	response := `{"triplets": [{"subject": "Go", "predicate": "is", "object": "typed", "confidence": 0.9}]}`
	mock := &mockLLMProvider{name: "test", response: response}
	s := NewTripletExtractionService(mock)

	triplets, err := s.ExtractAndBuild(context.Background(), "mem-1", "Go is statically typed.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(triplets) != 1 {
		t.Errorf("expected 1 triplet, got %d", len(triplets))
	}
}

func TestTripletExtraction_InvalidJSONRetries(t *testing.T) {
	mock := &selfCorrectingMock{
		responses: []string{
			"not valid json",
			`{"triplets":[{"subject":"A","predicate":"b","object":"C","confidence":0.9}]}`,
		},
	}
	s := NewTripletExtractionService(mock)

	triplets, err := s.ExtractFromContent(context.Background(), "some text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(triplets) != 1 {
		t.Errorf("expected 1 triplet after retry, got %d", len(triplets))
	}
	if mock.callCount != 2 {
		t.Errorf("expected 2 LLM calls (initial + retry), got %d", mock.callCount)
	}
}
