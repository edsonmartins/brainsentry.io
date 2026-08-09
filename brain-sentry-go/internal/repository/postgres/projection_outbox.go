package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectionEvent struct {
	ID          string
	AggregateID string
	TenantID    string
	Operation   string
	Attempts    int
}

type ProjectionOutboxRepository struct {
	pool *pgxpool.Pool
}

func NewProjectionOutboxRepository(pool *pgxpool.Pool) *ProjectionOutboxRepository {
	return &ProjectionOutboxRepository{pool: pool}
}

func (r *ProjectionOutboxRepository) Claim(ctx context.Context, limit int) ([]ProjectionEvent, error) {
	if limit <= 0 {
		limit = 25
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `WITH claimed AS (
		SELECT id FROM projection_outbox
		WHERE status IN ('pending', 'processing') AND available_at <= NOW()
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED LIMIT $1
	)
	UPDATE projection_outbox o
	SET status = 'processing', attempts = attempts + 1, available_at = NOW() + INTERVAL '1 minute'
	FROM claimed WHERE o.id = claimed.id
	RETURNING o.id, o.aggregate_id, o.tenant_id, o.operation, o.attempts`, limit)
	if err != nil {
		return nil, fmt.Errorf("claiming projection events: %w", err)
	}
	defer rows.Close()
	var events []ProjectionEvent
	for rows.Next() {
		var event ProjectionEvent
		if err := rows.Scan(&event.ID, &event.AggregateID, &event.TenantID, &event.Operation, &event.Attempts); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *ProjectionOutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE projection_outbox SET status='processed', processed_at=NOW(), last_error=NULL WHERE id=$1`, id)
	return err
}

func (r *ProjectionOutboxRepository) MarkFailed(ctx context.Context, event ProjectionEvent, cause error) error {
	delay := time.Duration(1<<min(event.Attempts, 8)) * time.Second
	_, err := r.pool.Exec(ctx, `UPDATE projection_outbox
		SET status='pending', available_at=$2, last_error=$3 WHERE id=$1`,
		event.ID, time.Now().Add(delay), cause.Error())
	return err
}

func (r *ProjectionOutboxRepository) RecoverLeases(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `UPDATE projection_outbox SET status='pending'
		WHERE status='processing' AND available_at <= NOW()`)
	return err
}
