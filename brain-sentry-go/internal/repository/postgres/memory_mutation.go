package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (r *MemoryRepository) recordCanonicalMutation(ctx context.Context, tx pgx.Tx, m *domain.Memory, operation, changeReason string) error {
	now := m.UpdatedAt
	if now.IsZero() {
		now = time.Now()
	}

	if operation != "create" {
		if _, err := tx.Exec(ctx, `UPDATE memory_history SET system_to = $1
			WHERE memory_id = $2 AND tenant_id = $3 AND system_to IS NULL`, now, m.ID, m.TenantID); err != nil {
			return fmt.Errorf("closing bitemporal history: %w", err)
		}
	}

	snapshot, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("encoding bitemporal snapshot: %w", err)
	}
	tags := m.Tags
	if tags == nil {
		tags = []string{}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO memory_history
		(memory_id, tenant_id, system_from, valid_from, valid_to, operation, snapshot, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.TenantID, now, m.ValidFrom, m.ValidTo, operation, snapshot, tags); err != nil {
		return fmt.Errorf("inserting bitemporal history: %w", err)
	}

	versionID := uuid.New().String()
	if _, err := tx.Exec(ctx, `INSERT INTO memory_versions
		(id, memory_id, version, content, summary, category, importance, metadata,
		 code_example, changed_by, change_reason, change_type, created_at, tenant_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		versionID, m.ID, m.Version, m.Content, m.Summary, m.Category, m.Importance,
		m.Metadata, m.CodeExample, m.CreatedBy, changeReason, operation, now, m.TenantID); err != nil {
		return fmt.Errorf("inserting atomic memory version: %w", err)
	}
	for _, tag := range m.Tags {
		if _, err := tx.Exec(ctx, `INSERT INTO memory_version_tags (memory_version_id, tag)
			VALUES ($1,$2) ON CONFLICT DO NOTHING`, versionID, tag); err != nil {
			return fmt.Errorf("inserting atomic version tag: %w", err)
		}
	}

	eventType := "memory_" + operation + "d"
	if operation == "delete" {
		eventType = "memory_deleted"
	}
	auditID := uuid.New().String()
	output, _ := json.Marshal(map[string]any{"memoryId": m.ID, "version": m.Version})
	if _, err := tx.Exec(ctx, `INSERT INTO audit_logs
		(id, event_type, timestamp, user_id, output_data, outcome, tenant_id)
		VALUES ($1,$2,$3,$4,$5,'success',$6)`,
		auditID, eventType, now, m.CreatedBy, output, m.TenantID); err != nil {
		return fmt.Errorf("inserting atomic audit: %w", err)
	}
	linkTable := "audit_memories_modified"
	if operation == "create" {
		linkTable = "audit_memories_created"
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s (audit_log_id, memory_id) VALUES ($1,$2)`, linkTable), auditID, m.ID); err != nil {
		return fmt.Errorf("linking atomic audit: %w", err)
	}

	outboxPayload, _ := json.Marshal(map[string]any{"memoryId": m.ID})
	if _, err := tx.Exec(ctx, `INSERT INTO projection_outbox
		(id, aggregate_type, aggregate_id, tenant_id, operation, payload)
		VALUES ($1,'memory',$2,$3,$4,$5)`,
		uuid.New().String(), m.ID, m.TenantID, operation, outboxPayload); err != nil {
		return fmt.Errorf("inserting graph projection outbox event: %w", err)
	}
	return nil
}

func loadTagsTx(ctx context.Context, tx pgx.Tx, memoryID string) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT tag FROM memory_tags WHERE memory_id = $1 ORDER BY tag`, memoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}
