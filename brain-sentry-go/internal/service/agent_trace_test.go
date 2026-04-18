package service

import (
	"context"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

func tenantCtx(id string) context.Context {
	return tenant.WithTenant(context.Background(), id)
}

func TestAgentTrace_Record(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	ctx := tenantCtx("t1")

	trace, err := svc.Record(ctx, RecordTraceRequest{
		OriginFunction: "handle_query",
		WithMemory:     true,
		MemoryQuery:    "what is Go?",
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if trace.ID == "" {
		t.Error("expected generated ID")
	}
	if trace.TenantID != "t1" {
		t.Errorf("tenant mismatch: %s", trace.TenantID)
	}
	if trace.Status != domain.AgentTraceSuccess {
		t.Errorf("default status should be success, got %s", trace.Status)
	}
	if trace.Text == "" {
		t.Error("expected generated embeddable text")
	}
}

func TestAgentTrace_RecordRequiresFunction(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	_, err := svc.Record(tenantCtx("t1"), RecordTraceRequest{})
	if err == nil {
		t.Error("expected error for missing originFunction")
	}
}

func TestAgentTrace_GetAndList(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	ctx := tenantCtx("t1")

	for i := 0; i < 3; i++ {
		svc.Record(ctx, RecordTraceRequest{
			OriginFunction: "f",
			SessionID:      "s1",
		})
	}

	list := svc.List(ctx, ListFilter{SessionID: "s1"})
	if len(list) != 3 {
		t.Errorf("expected 3 traces, got %d", len(list))
	}

	got, err := svc.Get(ctx, list[0].ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != list[0].ID {
		t.Error("get returned wrong trace")
	}
}

func TestAgentTrace_TenantIsolation(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	svc.Record(tenantCtx("t1"), RecordTraceRequest{OriginFunction: "f"})
	svc.Record(tenantCtx("t2"), RecordTraceRequest{OriginFunction: "f"})

	l1 := svc.List(tenantCtx("t1"), ListFilter{})
	l2 := svc.List(tenantCtx("t2"), ListFilter{})

	if len(l1) != 1 || len(l2) != 1 {
		t.Errorf("expected 1 each, got t1=%d t2=%d", len(l1), len(l2))
	}
}

func TestAgentTrace_FilterByStatus(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	ctx := tenantCtx("t1")

	svc.Record(ctx, RecordTraceRequest{OriginFunction: "ok", Status: domain.AgentTraceSuccess})
	svc.Record(ctx, RecordTraceRequest{OriginFunction: "bad", Status: domain.AgentTraceError, ErrorMessage: "boom"})
	svc.Record(ctx, RecordTraceRequest{OriginFunction: "ok2", Status: domain.AgentTraceSuccess})

	errs := svc.List(ctx, ListFilter{Status: domain.AgentTraceError})
	if len(errs) != 1 {
		t.Errorf("expected 1 error trace, got %d", len(errs))
	}
	if errs[0].OriginFunction != "bad" {
		t.Errorf("wrong trace: %s", errs[0].OriginFunction)
	}
}

func TestAgentTrace_FilterBySet(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	ctx := tenantCtx("t1")

	svc.Record(ctx, RecordTraceRequest{
		OriginFunction: "a",
		BelongsToSets:  []string{"set1", "set2"},
	})
	svc.Record(ctx, RecordTraceRequest{
		OriginFunction: "b",
		BelongsToSets:  []string{"set3"},
	})

	filtered := svc.List(ctx, ListFilter{Set: "set1"})
	if len(filtered) != 1 {
		t.Errorf("expected 1 trace in set1, got %d", len(filtered))
	}
	if filtered[0].OriginFunction != "a" {
		t.Error("wrong trace returned")
	}
}

func TestAgentTrace_Delete(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	ctx := tenantCtx("t1")

	trace, _ := svc.Record(ctx, RecordTraceRequest{OriginFunction: "f"})
	if err := svc.Delete(ctx, trace.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := svc.Get(ctx, trace.ID); err == nil {
		t.Error("expected error after delete")
	}
}

func TestAgentTrace_Stats(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	ctx := tenantCtx("t1")

	svc.Record(ctx, RecordTraceRequest{OriginFunction: "a", WithMemory: true, DurationMs: 100})
	svc.Record(ctx, RecordTraceRequest{OriginFunction: "b", WithMemory: false, DurationMs: 200})
	svc.Record(ctx, RecordTraceRequest{OriginFunction: "c", Status: domain.AgentTraceError, DurationMs: 300})

	stats := svc.Stats(ctx)
	if stats.Total != 3 {
		t.Errorf("expected total 3, got %d", stats.Total)
	}
	if stats.Success != 2 {
		t.Errorf("expected 2 success, got %d", stats.Success)
	}
	if stats.Errors != 1 {
		t.Errorf("expected 1 error, got %d", stats.Errors)
	}
	if stats.WithMemory != 1 {
		t.Errorf("expected 1 withMemory, got %d", stats.WithMemory)
	}
	if stats.AvgDurationMs != 200 {
		t.Errorf("expected avg 200, got %f", stats.AvgDurationMs)
	}
}

func TestAgentTrace_MaxPerTenantEnforced(t *testing.T) {
	cfg := DefaultAgentTraceConfig()
	cfg.MaxPerTenant = 3
	svc := NewAgentTraceService(cfg)
	ctx := tenantCtx("t1")

	for i := 0; i < 5; i++ {
		svc.Record(ctx, RecordTraceRequest{OriginFunction: "f"})
	}

	list := svc.List(ctx, ListFilter{})
	if len(list) != 3 {
		t.Errorf("expected 3 traces (capped), got %d", len(list))
	}
}

func TestAgentTrace_DefaultLimit(t *testing.T) {
	svc := NewAgentTraceService(DefaultAgentTraceConfig())
	ctx := tenantCtx("t1")

	for i := 0; i < 150; i++ {
		svc.Record(ctx, RecordTraceRequest{OriginFunction: "f"})
	}

	list := svc.List(ctx, ListFilter{})
	if len(list) != 100 {
		t.Errorf("expected default limit 100, got %d", len(list))
	}
}
