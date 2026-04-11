package service

import (
	"context"
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
	if svc.llm != nil {
		t.Error("expected nil llm")
	}
}

func TestNewReconciliationServiceWithLLM(t *testing.T) {
	llm := &mockLLMProvider{name: "test", response: `{"facts":[]}`}
	svc := NewReconciliationServiceWithLLM(llm, nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.llm != llm {
		t.Fatal("expected injected llm")
	}
}

func TestReconcileFacts_AddWithoutExistingMemory(t *testing.T) {
	llm := &mockLLMProvider{name: "test", response: `{"facts":[{"subject":"Brain Sentry","predicate":"uses","object":"tenant-scoped memory","context":"core product"}]}`}
	svc := NewReconciliationServiceWithLLM(llm, nil, nil)

	result, err := svc.ReconcileFacts(context.Background(), "Brain Sentry uses tenant-scoped memory.", "session-add")
	if err != nil {
		t.Fatalf("ReconcileFacts failed: %v", err)
	}

	if result.ExtractedFacts != 1 {
		t.Fatalf("expected 1 extracted fact, got %d", result.ExtractedFacts)
	}
	if result.Added != 1 || result.Updated != 0 || result.Deleted != 0 || result.Skipped != 0 {
		t.Fatalf("unexpected counters: %#v", result)
	}
	if len(result.Decisions) != 1 || result.Decisions[0].Action != FactActionAdd {
		t.Fatalf("expected ADD decision, got %#v", result.Decisions)
	}
	if llm.callCount != 1 {
		t.Fatalf("expected only extraction LLM call when no existing memories, got %d", llm.callCount)
	}
}

func TestReconciliationDecideAction_Update(t *testing.T) {
	llm := &mockLLMProvider{name: "test", response: `{"action":"UPDATE","reason":"new fact corrects old memory","existingMemoryId":"mem-1","mergedContent":"Brain Sentry uses autonomous context injection"}`}
	svc := NewReconciliationServiceWithLLM(llm, nil, nil)

	decision, err := svc.decideAction(context.Background(),
		ExtractedFact{Subject: "Brain Sentry", Predicate: "uses", Object: "autonomous context injection"},
		[]domain.Memory{{ID: "mem-1", Content: "Brain Sentry uses manual context lookup"}},
	)
	if err != nil {
		t.Fatalf("decideAction failed: %v", err)
	}

	if decision.Action != FactActionUpdate {
		t.Fatalf("expected UPDATE, got %#v", decision)
	}
	if decision.ExistingMemoryID != "mem-1" {
		t.Fatalf("expected existing memory mem-1, got %s", decision.ExistingMemoryID)
	}
	if decision.MergedContent == "" {
		t.Fatal("expected merged content")
	}
}

func TestReconciliationDecideAction_Delete(t *testing.T) {
	llm := &mockLLMProvider{name: "test", response: `{"action":"DELETE","reason":"new fact invalidates old memory","existingMemoryId":"mem-old"}`}
	svc := NewReconciliationServiceWithLLM(llm, nil, nil)

	decision, err := svc.decideAction(context.Background(),
		ExtractedFact{Subject: "Brain Sentry", Predicate: "replaces", Object: "legacy RAG"},
		[]domain.Memory{{ID: "mem-old", Content: "Brain Sentry is legacy RAG"}},
	)
	if err != nil {
		t.Fatalf("decideAction failed: %v", err)
	}

	if decision.Action != FactActionDelete {
		t.Fatalf("expected DELETE, got %#v", decision)
	}
	if decision.ExistingMemoryID != "mem-old" {
		t.Fatalf("expected existing memory mem-old, got %s", decision.ExistingMemoryID)
	}
}

func TestReconciliationDecideAction_None(t *testing.T) {
	llm := &mockLLMProvider{name: "test", response: `{"action":"NONE","reason":"fact already covered"}`}
	svc := NewReconciliationServiceWithLLM(llm, nil, nil)

	decision, err := svc.decideAction(context.Background(),
		ExtractedFact{Subject: "Brain Sentry", Predicate: "uses", Object: "tenant isolation"},
		[]domain.Memory{{ID: "mem-1", Content: "Brain Sentry uses tenant isolation"}},
	)
	if err != nil {
		t.Fatalf("decideAction failed: %v", err)
	}

	if decision.Action != FactActionNone {
		t.Fatalf("expected NONE, got %#v", decision)
	}
}

func TestReconciliationDecideAction_ParseErrorDefaultsToAdd(t *testing.T) {
	llm := &mockLLMProvider{name: "test", response: `not json`}
	svc := NewReconciliationServiceWithLLM(llm, nil, nil)

	decision, err := svc.decideAction(context.Background(),
		ExtractedFact{Subject: "Brain Sentry", Predicate: "uses", Object: "memory"},
		[]domain.Memory{{ID: "mem-1", Content: "existing memory"}},
	)
	if err != nil {
		t.Fatalf("decideAction failed: %v", err)
	}

	if decision.Action != FactActionAdd {
		t.Fatalf("expected ADD fallback, got %#v", decision)
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
