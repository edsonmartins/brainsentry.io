package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// VersionRepository handles memory version persistence.
type VersionRepository struct {
	pool *pgxpool.Pool
}

// NewVersionRepository creates a new VersionRepository.
func NewVersionRepository(pool *pgxpool.Pool) *VersionRepository {
	return &VersionRepository{pool: pool}
}

const versionColumns = `id, memory_id, version, content, summary, category, importance,
	metadata, code_example, changed_by, change_reason, change_type, created_at, tenant_id`

// Create inserts a new memory version.
func (r *VersionRepository) Create(ctx context.Context, v *domain.MemoryVersion) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	if v.TenantID == "" {
		v.TenantID = tenant.FromContext(ctx)
	}

	query := fmt.Sprintf(`INSERT INTO memory_versions (%s) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, versionColumns)

	_, err := r.pool.Exec(ctx, query,
		v.ID, v.MemoryID, v.Version, v.Content, v.Summary, v.Category, v.Importance,
		v.Metadata, v.CodeExample, v.ChangedBy, v.ChangeReason, v.ChangeType, v.CreatedAt, v.TenantID,
	)
	if err != nil {
		return fmt.Errorf("inserting version: %w", err)
	}
	return nil
}

// FindByMemoryID returns all versions for a memory.
func (r *VersionRepository) FindByMemoryID(ctx context.Context, memoryID string) ([]domain.MemoryVersion, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM memory_versions WHERE memory_id = $1 AND tenant_id = $2 ORDER BY version DESC`, versionColumns)

	rows, err := r.pool.Query(ctx, query, memoryID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("finding versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.MemoryVersion
	for rows.Next() {
		var v domain.MemoryVersion
		if err := rows.Scan(
			&v.ID, &v.MemoryID, &v.Version, &v.Content, &v.Summary, &v.Category, &v.Importance,
			&v.Metadata, &v.CodeExample, &v.ChangedBy, &v.ChangeReason, &v.ChangeType, &v.CreatedAt, &v.TenantID,
		); err != nil {
			return nil, fmt.Errorf("scanning version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// GetLatestVersion returns the latest version number for a memory.
func (r *VersionRepository) GetLatestVersion(ctx context.Context, memoryID string) (int, error) {
	tenantID := tenant.FromContext(ctx)
	var version int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(version), 0) FROM memory_versions WHERE memory_id = $1 AND tenant_id = $2`,
		memoryID, tenantID).Scan(&version)
	return version, err
}

// Count returns the number of versions for a memory.
func (r *VersionRepository) Count(ctx context.Context, memoryID string) (int64, error) {
	tenantID := tenant.FromContext(ctx)
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM memory_versions WHERE memory_id = $1 AND tenant_id = $2`,
		memoryID, tenantID).Scan(&count)
	return count, err
}

// DeleteByMemoryID deletes all versions for a memory.
func (r *VersionRepository) DeleteByMemoryID(ctx context.Context, memoryID string) error {
	tenantID := tenant.FromContext(ctx)
	_, err := r.pool.Exec(ctx,
		`DELETE FROM memory_versions WHERE memory_id = $1 AND tenant_id = $2`, memoryID, tenantID)
	return err
}

// CreateFromMemory creates a version snapshot from the current state of a memory.
func (r *VersionRepository) CreateFromMemory(ctx context.Context, m *domain.Memory, changeType, changeReason, changedBy string) error {
	latestVer, err := r.GetLatestVersion(ctx, m.ID)
	if err != nil {
		return err
	}

	var metadataBytes []byte
	if m.Metadata != nil {
		metadataBytes = m.Metadata
	}

	v := &domain.MemoryVersion{
		MemoryID:     m.ID,
		Version:      latestVer + 1,
		Content:      m.Content,
		Summary:      m.Summary,
		Category:     m.Category,
		Importance:   m.Importance,
		Metadata:     metadataBytes,
		CodeExample:  m.CodeExample,
		ChangedBy:    changedBy,
		ChangeReason: changeReason,
		ChangeType:   changeType,
		TenantID:     m.TenantID,
	}

	if err := r.Create(ctx, v); err != nil {
		return err
	}

	// Insert version tags
	if len(m.Tags) > 0 {
		if err := r.insertVersionTags(ctx, v.ID, m.Tags); err != nil {
			return err
		}
	}

	return nil
}

func (r *VersionRepository) insertVersionTags(ctx context.Context, versionID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO memory_version_tags (memory_version_id, tag) VALUES ")
	args := make([]any, 0, len(tags)*2)
	for i, tag := range tags {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, versionID, tag)
	}
	sb.WriteString(" ON CONFLICT DO NOTHING")
	_, err := r.pool.Exec(ctx, sb.String(), args...)
	return err
}

// MetadataToJSON helper.
func MetadataToJSON(m json.RawMessage) json.RawMessage {
	return m
}
