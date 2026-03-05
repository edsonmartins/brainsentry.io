package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestMemoryResponse_Marshal(t *testing.T) {
	resp := MemoryResponse{
		ID:       "m1",
		Content:  "test",
		Category: domain.CategoryDecision,
		TenantID: "t1",
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestMemoryResponse_WithRelatedMemories(t *testing.T) {
	resp := MemoryResponse{
		ID:      "m1",
		Content: "test",
		RelatedMemories: []RelatedMemoryRef{
			{ID: "m2", Summary: "related", RelationshipType: domain.RelRelatedTo, Strength: 0.8},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if _, ok := result["relatedMemories"]; !ok {
		t.Error("expected relatedMemories in JSON output")
	}
}

func TestSearchResponse_Marshal(t *testing.T) {
	resp := SearchResponse{
		Results:      []MemoryResponse{{ID: "m1", Content: "test"}},
		Total:        1,
		SearchTimeMs: 42,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["searchTimeMs"] != float64(42) {
		t.Errorf("expected searchTimeMs 42, got %v", result["searchTimeMs"])
	}
	if result["total"] != float64(1) {
		t.Errorf("expected total 1, got %v", result["total"])
	}
}

func TestErrorResponse_Marshal(t *testing.T) {
	resp := ErrorResponse{
		Error:         "Bad Request",
		Message:       "content is required",
		Status:        400,
		ErrorCode:     "validation",
		ErrorCategory: "VALIDATION",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["errorCode"] != "validation" {
		t.Errorf("expected errorCode 'validation', got '%v'", result["errorCode"])
	}
	if result["errorCategory"] != "VALIDATION" {
		t.Errorf("expected errorCategory 'VALIDATION', got '%v'", result["errorCategory"])
	}
}

func TestMemoryListResponse_Marshal(t *testing.T) {
	resp := MemoryListResponse{
		Memories:      []MemoryResponse{},
		Page:          0,
		Size:          20,
		TotalElements: 0,
		TotalPages:    0,
		HasNext:       false,
		HasPrevious:   false,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if _, ok := result["memories"]; !ok {
		t.Error("expected memories field")
	}
	if _, ok := result["totalElements"]; !ok {
		t.Error("expected totalElements field")
	}
}

func TestRelatedMemoryRef_Marshal(t *testing.T) {
	ref := RelatedMemoryRef{
		ID:               "m2",
		Summary:          "related memory",
		RelationshipType: domain.RelUsedWith,
		Strength:         0.9,
	}
	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["relationshipType"] != "USED_WITH" {
		t.Errorf("expected 'USED_WITH', got '%v'", result["relationshipType"])
	}
	if result["strength"] != 0.9 {
		t.Errorf("expected 0.9, got %v", result["strength"])
	}
}
