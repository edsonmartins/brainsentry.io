package service

import (
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestFactAction_Constants(t *testing.T) {
	if FactActionAdd != "ADD" {
		t.Error("expected ADD")
	}
	if FactActionUpdate != "UPDATE" {
		t.Error("expected UPDATE")
	}
	if FactActionDelete != "DELETE" {
		t.Error("expected DELETE")
	}
	if FactActionNone != "NONE" {
		t.Error("expected NONE")
	}
}

func TestExtractedFact_Structure(t *testing.T) {
	fact := ExtractedFact{
		Subject:   "PostgreSQL",
		Predicate: "is",
		Object:    "a relational database",
		Context:   "database technologies",
	}
	if fact.Subject != "PostgreSQL" {
		t.Error("expected PostgreSQL")
	}
}

func TestFactDecision_Structure(t *testing.T) {
	decision := FactDecision{
		Fact:             ExtractedFact{Subject: "Go", Predicate: "is", Object: "compiled"},
		Action:           FactActionUpdate,
		Reason:           "fact already exists but needs update",
		ExistingMemoryID: "mem-123",
		MergedContent:    "Go is a compiled, statically typed language",
	}
	if decision.Action != FactActionUpdate {
		t.Error("expected UPDATE action")
	}
	if decision.ExistingMemoryID != "mem-123" {
		t.Error("expected mem-123")
	}
}

func TestReconciliationResult_Counters(t *testing.T) {
	result := ReconciliationResult{
		ExtractedFacts: 5,
		Added:          2,
		Updated:        1,
		Deleted:        1,
		Skipped:        1,
	}
	total := result.Added + result.Updated + result.Deleted + result.Skipped
	if total != result.ExtractedFacts {
		t.Errorf("expected total %d to equal extracted %d", total, result.ExtractedFacts)
	}
}

func TestNewReconciliationService_NilOpenRouter(t *testing.T) {
	svc := NewReconciliationService(nil, nil, nil)
	if svc == nil {
		t.Error("expected non-nil service")
	}
	if svc.openRouter != nil {
		t.Error("expected nil openRouter")
	}
}

func TestCreateFactMemoryRequest(t *testing.T) {
	req := createFactMemoryRequest("Go is compiled", "session-1")
	if req.Content != "Go is compiled" {
		t.Error("expected correct content")
	}
	if req.SourceType != "reconciliation" {
		t.Error("expected reconciliation source type")
	}
	if req.Metadata["sessionId"] != "session-1" {
		t.Error("expected session-1 in metadata")
	}
}

func TestUpdateFactMemoryRequest(t *testing.T) {
	req := updateFactMemoryRequest("updated content", "fact corrected")
	if req.Content != "updated content" {
		t.Error("expected correct content")
	}
	if req.ChangeReason != "fact reconciliation: fact corrected" {
		t.Errorf("unexpected change reason: %s", req.ChangeReason)
	}
}

// Compile-time checks
var _ = domain.Memory{}
