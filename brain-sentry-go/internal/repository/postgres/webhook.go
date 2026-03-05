package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/integraltech/brainsentry/internal/domain"
)

// WebhookRepository handles webhook persistence in PostgreSQL.
type WebhookRepository struct {
	pool *pgxpool.Pool
}

// NewWebhookRepository creates a new WebhookRepository.
func NewWebhookRepository(pool *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{pool: pool}
}

// Create inserts a new webhook with its event subscriptions.
func (r *WebhookRepository) Create(ctx context.Context, wh *domain.Webhook) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO webhooks (id, tenant_id, url, secret, active, created_at, updated_at, last_error, fail_count)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		wh.ID, wh.TenantID, wh.URL, wh.Secret, wh.Active, wh.CreatedAt, wh.UpdatedAt, wh.LastError, wh.FailCount,
	)
	if err != nil {
		return fmt.Errorf("inserting webhook: %w", err)
	}

	if err := r.insertEvents(ctx, tx, wh.ID, wh.Events); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// FindByID retrieves a webhook by ID.
func (r *WebhookRepository) FindByID(ctx context.Context, id string) (*domain.Webhook, error) {
	var wh domain.Webhook
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, url, secret, active, created_at, updated_at, last_error, fail_count
		FROM webhooks WHERE id = $1`, id).Scan(
		&wh.ID, &wh.TenantID, &wh.URL, &wh.Secret, &wh.Active, &wh.CreatedAt, &wh.UpdatedAt, &wh.LastError, &wh.FailCount,
	)
	if err != nil {
		return nil, fmt.Errorf("finding webhook: %w", err)
	}

	events, err := r.loadEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	wh.Events = events
	return &wh, nil
}

// FindByTenant returns all webhooks for a tenant.
func (r *WebhookRepository) FindByTenant(ctx context.Context, tenantID string) ([]*domain.Webhook, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, url, secret, active, created_at, updated_at, last_error, fail_count
		FROM webhooks WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		var wh domain.Webhook
		if err := rows.Scan(&wh.ID, &wh.TenantID, &wh.URL, &wh.Secret, &wh.Active, &wh.CreatedAt, &wh.UpdatedAt, &wh.LastError, &wh.FailCount); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		events, _ := r.loadEvents(ctx, wh.ID)
		wh.Events = events
		webhooks = append(webhooks, &wh)
	}
	return webhooks, nil
}

// Update updates a webhook (fail_count, active, last_error, updated_at).
func (r *WebhookRepository) Update(ctx context.Context, wh *domain.Webhook) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE webhooks SET active=$1, updated_at=$2, last_error=$3, fail_count=$4 WHERE id=$5`,
		wh.Active, wh.UpdatedAt, wh.LastError, wh.FailCount, wh.ID,
	)
	if err != nil {
		return fmt.Errorf("updating webhook: %w", err)
	}
	return nil
}

// Delete removes a webhook.
func (r *WebhookRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting webhook: %w", err)
	}
	return nil
}

// CreateDelivery records a webhook delivery attempt.
func (r *WebhookRepository) CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhook_deliveries (id, webhook_id, event, payload, status_code, success, error, timestamp, latency_ms)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		d.ID, d.WebhookID, d.Event, d.Payload, d.StatusCode, d.Success, d.Error, d.Timestamp, d.LatencyMs,
	)
	if err != nil {
		return fmt.Errorf("inserting delivery: %w", err)
	}
	return nil
}

// FindDeliveries returns recent deliveries for a webhook.
func (r *WebhookRepository) FindDeliveries(ctx context.Context, webhookID string, limit int) ([]domain.WebhookDelivery, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, webhook_id, event, payload, status_code, success, error, timestamp, latency_ms
		FROM webhook_deliveries WHERE webhook_id = $1 ORDER BY timestamp DESC LIMIT $2`, webhookID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []domain.WebhookDelivery
	for rows.Next() {
		var d domain.WebhookDelivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.Event, &d.Payload, &d.StatusCode, &d.Success, &d.Error, &d.Timestamp, &d.LatencyMs); err != nil {
			return nil, fmt.Errorf("scanning delivery: %w", err)
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, nil
}

func (r *WebhookRepository) loadEvents(ctx context.Context, webhookID string) ([]domain.WebhookEventType, error) {
	rows, err := r.pool.Query(ctx, `SELECT event_type FROM webhook_events WHERE webhook_id = $1`, webhookID)
	if err != nil {
		return nil, fmt.Errorf("loading webhook events: %w", err)
	}
	defer rows.Close()

	var events []domain.WebhookEventType
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		events = append(events, domain.WebhookEventType(e))
	}
	return events, nil
}

func (r *WebhookRepository) insertEvents(ctx context.Context, tx pgx.Tx, webhookID string, events []domain.WebhookEventType) error {
	if len(events) == 0 {
		return nil
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO webhook_events (webhook_id, event_type) VALUES ")
	args := make([]any, 0, len(events)*2)
	for i, e := range events {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, webhookID, string(e))
	}
	sb.WriteString(" ON CONFLICT DO NOTHING")
	_, err := tx.Exec(ctx, sb.String(), args...)
	return err
}
