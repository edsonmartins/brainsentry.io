package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	graphrepo "github.com/integraltech/brainsentry/internal/repository/graph"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
	"github.com/jackc/pgx/v5"
)

type GraphProjectionWorker struct {
	outbox     projectionOutbox
	memories   projectionMemoryRepository
	graph      projectionGraph
	cancel     context.CancelFunc
	wait       sync.WaitGroup
	batchSize  int
	pollPeriod time.Duration
}

type projectionOutbox interface {
	Claim(context.Context, int) ([]postgres.ProjectionEvent, error)
	MarkProcessed(context.Context, string) error
	MarkFailed(context.Context, postgres.ProjectionEvent, error) error
	RecoverLeases(context.Context) error
}

type projectionMemoryRepository interface {
	FindByID(context.Context, string) (*domain.Memory, error)
}

type projectionGraph interface {
	PurgeMemoryNodes(context.Context, string, []string) error
	SaveToGraph(context.Context, *domain.Memory) error
	CreateTagRelationships(context.Context, *domain.Memory) error
}

func NewGraphProjectionWorker(outbox *postgres.ProjectionOutboxRepository, memories *postgres.MemoryRepository, graph *graphrepo.MemoryGraphRepository) *GraphProjectionWorker {
	return &GraphProjectionWorker{outbox: outbox, memories: memories, graph: graph, batchSize: 25, pollPeriod: time.Second}
}

func (w *GraphProjectionWorker) Start(parent context.Context) {
	if w == nil || w.outbox == nil || w.memories == nil || w.graph == nil || w.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	w.cancel = cancel
	w.wait.Add(1)
	go func() {
		defer w.wait.Done()
		_ = w.outbox.RecoverLeases(ctx)
		ticker := time.NewTicker(w.pollPeriod)
		defer ticker.Stop()
		for {
			w.processBatch(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (w *GraphProjectionWorker) Stop() {
	if w == nil || w.cancel == nil {
		return
	}
	w.cancel()
	w.wait.Wait()
}

func (w *GraphProjectionWorker) processBatch(ctx context.Context) {
	events, err := w.outbox.Claim(ctx, w.batchSize)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Warn("failed to claim graph projection events", "error", err)
		}
		return
	}
	for _, event := range events {
		if err := w.project(ctx, event); err != nil {
			slog.Warn("graph projection failed", "eventId", event.ID, "memoryId", event.AggregateID, "error", err)
			if markErr := w.outbox.MarkFailed(ctx, event, err); markErr != nil {
				slog.Warn("failed to reschedule graph projection", "eventId", event.ID, "error", markErr)
			}
			continue
		}
		if err := w.outbox.MarkProcessed(ctx, event.ID); err != nil {
			slog.Warn("failed to acknowledge graph projection", "eventId", event.ID, "error", err)
		}
	}
}

func (w *GraphProjectionWorker) project(ctx context.Context, event postgres.ProjectionEvent) error {
	tenantCtx := tenant.WithTenant(ctx, event.TenantID)
	if event.Operation == "delete" {
		return w.graph.PurgeMemoryNodes(tenantCtx, event.TenantID, []string{event.AggregateID})
	}
	memory, err := w.memories.FindByID(tenantCtx, event.AggregateID)
	if err != nil {
		// An older upsert event may be claimed after a later canonical delete.
		// Project current canonical state, not stale event payload, so ordering
		// across workers cannot resurrect a deleted graph node.
		if errors.Is(err, pgx.ErrNoRows) {
			return w.graph.PurgeMemoryNodes(tenantCtx, event.TenantID, []string{event.AggregateID})
		}
		return err
	}
	if err := w.graph.SaveToGraph(tenantCtx, memory); err != nil {
		return err
	}
	return w.graph.CreateTagRelationships(tenantCtx, memory)
}
