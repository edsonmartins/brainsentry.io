package handler

import (
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

func TestBuildTimelineGraphPreservesVersionsAndMemoryIdentity(t *testing.T) {
	t0 := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	history := []postgres.MemoryHistoryEntry{
		{Memory: domain.Memory{ID: "m1", Summary: "before", Version: 1, RecordedAt: t0}, Operation: "create", SystemFrom: t0, SystemTo: &t1},
		{Memory: domain.Memory{ID: "m1", Summary: "after", Version: 2, RecordedAt: t0}, Operation: "update", SystemFrom: t1},
	}

	nodes, edges := buildTimelineGraph(history)
	if len(nodes) != 2 || len(edges) != 1 {
		t.Fatalf("nodes=%d edges=%d, want 2 and 1", len(nodes), len(edges))
	}
	if nodes[0].ID == nodes[1].ID || nodes[0].MemoryID != "m1" || nodes[1].MemoryID != "m1" {
		t.Fatalf("timeline identities are not version-safe: %#v", nodes)
	}
	if !nodes[1].RecordedAt.Equal(t1) || nodes[1].Operation != "update" {
		t.Fatalf("system time/operation lost: %#v", nodes[1])
	}
	if edges[0].Type != "VERSION" || edges[0].Source != nodes[0].ID || edges[0].Target != nodes[1].ID {
		t.Fatalf("unexpected version edge: %#v", edges[0])
	}
}

func TestBuildTimelineGraphLinksCrossMemorySupersession(t *testing.T) {
	t0 := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	history := []postgres.MemoryHistoryEntry{
		{Memory: domain.Memory{ID: "old", Summary: "old", Version: 1, SupersededBy: "new"}, Operation: "update", SystemFrom: t0},
		{Memory: domain.Memory{ID: "new", Summary: "new", Version: 1}, Operation: "create", SystemFrom: t0.Add(time.Minute)},
	}

	_, edges := buildTimelineGraph(history)
	if len(edges) != 1 || edges[0].Type != "SUPERSEDES" {
		t.Fatalf("expected one supersession edge, got %#v", edges)
	}
}
