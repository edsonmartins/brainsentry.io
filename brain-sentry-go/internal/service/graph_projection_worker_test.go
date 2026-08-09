package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

type fakeProjectionOutbox struct {
	events    []postgres.ProjectionEvent
	claimErr  error
	processed []string
	failed    []string
}

func (f *fakeProjectionOutbox) Claim(context.Context, int) ([]postgres.ProjectionEvent, error) {
	return f.events, f.claimErr
}
func (f *fakeProjectionOutbox) MarkProcessed(_ context.Context, id string) error {
	f.processed = append(f.processed, id)
	return nil
}
func (f *fakeProjectionOutbox) MarkFailed(_ context.Context, event postgres.ProjectionEvent, _ error) error {
	f.failed = append(f.failed, event.ID)
	return nil
}
func (f *fakeProjectionOutbox) RecoverLeases(context.Context) error { return nil }

type fakeProjectionMemories struct {
	memory *domain.Memory
	err    error
}

func (f *fakeProjectionMemories) FindByID(context.Context, string) (*domain.Memory, error) {
	return f.memory, f.err
}

type fakeProjectionGraph struct {
	saved  []string
	purged []string
	err    error
}

func (f *fakeProjectionGraph) SaveToGraph(_ context.Context, memory *domain.Memory) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, memory.ID)
	return nil
}
func (f *fakeProjectionGraph) CreateTagRelationships(context.Context, *domain.Memory) error {
	return f.err
}
func (f *fakeProjectionGraph) PurgeMemoryNodes(_ context.Context, _ string, ids []string) error {
	if f.err != nil {
		return f.err
	}
	f.purged = append(f.purged, ids...)
	return nil
}

func TestGraphProjectionWorkerProcessesUpsertAndDelete(t *testing.T) {
	outbox := &fakeProjectionOutbox{events: []postgres.ProjectionEvent{
		{ID: "event-upsert", AggregateID: "memory-1", TenantID: "tenant-1", Operation: "update"},
		{ID: "event-delete", AggregateID: "memory-1", TenantID: "tenant-1", Operation: "delete"},
	}}
	graph := &fakeProjectionGraph{}
	worker := &GraphProjectionWorker{
		outbox: outbox, memories: &fakeProjectionMemories{memory: &domain.Memory{ID: "memory-1"}}, graph: graph, batchSize: 10,
	}

	worker.processBatch(context.Background())

	if !reflect.DeepEqual(outbox.processed, []string{"event-upsert", "event-delete"}) || len(outbox.failed) != 0 {
		t.Fatalf("processed=%v failed=%v", outbox.processed, outbox.failed)
	}
	if !reflect.DeepEqual(graph.saved, []string{"memory-1"}) || !reflect.DeepEqual(graph.purged, []string{"memory-1"}) {
		t.Fatalf("saved=%v purged=%v", graph.saved, graph.purged)
	}
}

func TestGraphProjectionWorkerDoesNotResurrectDeletedMemory(t *testing.T) {
	graph := &fakeProjectionGraph{}
	worker := &GraphProjectionWorker{
		memories: &fakeProjectionMemories{err: errors.Join(errors.New("finding memory"), pgx.ErrNoRows)}, graph: graph,
	}
	event := postgres.ProjectionEvent{AggregateID: "deleted", TenantID: "tenant-1", Operation: "update"}

	if err := worker.project(context.Background(), event); err != nil {
		t.Fatalf("project() error = %v", err)
	}
	if !reflect.DeepEqual(graph.purged, []string{"deleted"}) || len(graph.saved) != 0 {
		t.Fatalf("saved=%v purged=%v", graph.saved, graph.purged)
	}
}

func TestGraphProjectionWorkerReschedulesFailure(t *testing.T) {
	outbox := &fakeProjectionOutbox{events: []postgres.ProjectionEvent{{ID: "event-1", AggregateID: "memory-1", Operation: "update"}}}
	worker := &GraphProjectionWorker{
		outbox:   outbox,
		memories: &fakeProjectionMemories{memory: &domain.Memory{ID: "memory-1"}},
		graph:    &fakeProjectionGraph{err: errors.New("graph unavailable")},
	}

	worker.processBatch(context.Background())

	if !reflect.DeepEqual(outbox.failed, []string{"event-1"}) || len(outbox.processed) != 0 {
		t.Fatalf("processed=%v failed=%v", outbox.processed, outbox.failed)
	}
}
