package dto

import (
	"encoding/json"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestCreateMemoryRequest_Unmarshal(t *testing.T) {
	body := `{"content":"test memory","category":"DECISION","importance":"CRITICAL","tags":["go","test"]}`
	var req CreateMemoryRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Content != "test memory" {
		t.Errorf("expected 'test memory', got '%s'", req.Content)
	}
	if req.Category != domain.CategoryDecision {
		t.Errorf("expected DECISION, got '%s'", req.Category)
	}
	if req.Importance != domain.ImportanceCritical {
		t.Errorf("expected CRITICAL, got '%s'", req.Importance)
	}
	if len(req.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(req.Tags))
	}
}

func TestCreateMemoryRequest_EmptyContent(t *testing.T) {
	body := `{"summary":"no content"}`
	var req CreateMemoryRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Content != "" {
		t.Error("expected empty content")
	}
}

func TestUpdateMemoryRequest_Unmarshal(t *testing.T) {
	body := `{"content":"updated","category":"PATTERN","changeReason":"fixing"}`
	var req UpdateMemoryRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Content != "updated" {
		t.Errorf("expected 'updated', got '%s'", req.Content)
	}
	if req.Category != domain.CategoryPattern {
		t.Errorf("expected PATTERN, got '%s'", req.Category)
	}
	if req.ChangeReason != "fixing" {
		t.Errorf("expected 'fixing', got '%s'", req.ChangeReason)
	}
}

func TestSearchRequest_Defaults(t *testing.T) {
	body := `{"query":"search term"}`
	var req SearchRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Query != "search term" {
		t.Errorf("expected 'search term', got '%s'", req.Query)
	}
	if req.Limit != 0 {
		t.Errorf("expected 0 (default), got %d", req.Limit)
	}
}

func TestSearchRequest_WithOptions(t *testing.T) {
	body := `{"query":"test","limit":5,"categories":["DECISION","PATTERN"],"includeRelated":true}`
	var req SearchRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Limit != 5 {
		t.Errorf("expected 5, got %d", req.Limit)
	}
	if len(req.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(req.Categories))
	}
	if !req.IncludeRelated {
		t.Error("expected includeRelated true")
	}
}

func TestInterceptRequest_Unmarshal(t *testing.T) {
	body := `{"prompt":"hello","userId":"u1","maxTokens":500,"forceDeepAnalysis":true}`
	var req InterceptRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Prompt != "hello" {
		t.Errorf("expected 'hello', got '%s'", req.Prompt)
	}
	if req.MaxTokens != 500 {
		t.Errorf("expected 500, got %d", req.MaxTokens)
	}
	if !req.ForceDeepAnalysis {
		t.Error("expected forceDeepAnalysis true")
	}
}

func TestLoginRequest_Unmarshal(t *testing.T) {
	body := `{"email":"test@example.com","password":"secret123"}`
	var req LoginRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Email != "test@example.com" {
		t.Errorf("expected 'test@example.com', got '%s'", req.Email)
	}
	if req.Password != "secret123" {
		t.Errorf("expected 'secret123', got '%s'", req.Password)
	}
}

func TestFlagMemoryRequest_Unmarshal(t *testing.T) {
	body := `{"reason":"incorrect info","correctedContent":"correct info"}`
	var req FlagMemoryRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Reason != "incorrect info" {
		t.Errorf("expected 'incorrect info', got '%s'", req.Reason)
	}
}

func TestCompressionRequest_Unmarshal(t *testing.T) {
	body := `{"messages":[{"role":"user","content":"hello"}],"tokenThreshold":1000}`
	var req CompressionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(req.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(req.Messages))
	}
	if req.TokenThreshold != 1000 {
		t.Errorf("expected 1000, got %d", req.TokenThreshold)
	}
}

func TestCreateTenantRequest_Unmarshal(t *testing.T) {
	body := `{"name":"Test Org","slug":"test-org","maxMemories":1000}`
	var req CreateTenantRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Name != "Test Org" {
		t.Errorf("expected 'Test Org', got '%s'", req.Name)
	}
	if req.Slug != "test-org" {
		t.Errorf("expected 'test-org', got '%s'", req.Slug)
	}
	if req.MaxMemories != 1000 {
		t.Errorf("expected 1000, got %d", req.MaxMemories)
	}
}

func TestCreateUserRequest_Unmarshal(t *testing.T) {
	body := `{"email":"user@test.com","password":"password123","tenantId":"t1","roles":["admin"]}`
	var req CreateUserRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.Email != "user@test.com" {
		t.Errorf("expected 'user@test.com', got '%s'", req.Email)
	}
	if len(req.Roles) != 1 || req.Roles[0] != "admin" {
		t.Errorf("expected ['admin'], got %v", req.Roles)
	}
}
