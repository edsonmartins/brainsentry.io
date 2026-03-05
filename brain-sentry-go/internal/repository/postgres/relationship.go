package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// RelationshipRepository handles memory relationship persistence.
type RelationshipRepository struct {
	pool *pgxpool.Pool
}

// NewRelationshipRepository creates a new RelationshipRepository.
func NewRelationshipRepository(pool *pgxpool.Pool) *RelationshipRepository {
	return &RelationshipRepository{pool: pool}
}

const relColumns = `id, from_memory_id, to_memory_id, type, frequency, severity, strength, description, created_at, last_used_at, tenant_id`

func (r *RelationshipRepository) scanRel(dest *domain.MemoryRelationship, scanFn func(dest ...any) error) error {
	return scanFn(
		&dest.ID, &dest.FromMemoryID, &dest.ToMemoryID, &dest.Type,
		&dest.Frequency, &dest.Severity, &dest.Strength, &dest.Description,
		&dest.CreatedAt, &dest.LastUsedAt, &dest.TenantID,
	)
}

// Create inserts a new memory relationship.
func (r *RelationshipRepository) Create(ctx context.Context, rel *domain.MemoryRelationship) error {
	query := fmt.Sprintf(`INSERT INTO memory_relationships (%s) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, relColumns)
	_, err := r.pool.Exec(ctx, query,
		rel.ID, rel.FromMemoryID, rel.ToMemoryID, rel.Type,
		rel.Frequency, rel.Severity, rel.Strength, rel.Description,
		rel.CreatedAt, rel.LastUsedAt, rel.TenantID,
	)
	if err != nil {
		return fmt.Errorf("creating relationship: %w", err)
	}
	return nil
}

// Update updates a relationship.
func (r *RelationshipRepository) Update(ctx context.Context, rel *domain.MemoryRelationship) error {
	query := `UPDATE memory_relationships SET frequency=$1, strength=$2, last_used_at=$3, description=$4 WHERE id=$5`
	_, err := r.pool.Exec(ctx, query, rel.Frequency, rel.Strength, rel.LastUsedAt, rel.Description, rel.ID)
	return err
}

// FindByFromAndTo finds a relationship between two memories.
func (r *RelationshipRepository) FindByFromAndTo(ctx context.Context, fromID, toID string) (*domain.MemoryRelationship, error) {
	query := fmt.Sprintf(`SELECT %s FROM memory_relationships WHERE from_memory_id = $1 AND to_memory_id = $2`, relColumns)
	var rel domain.MemoryRelationship
	err := r.scanRel(&rel, r.pool.QueryRow(ctx, query, fromID, toID).Scan)
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// FindByFromMemoryID returns outgoing relationships from a memory.
func (r *RelationshipRepository) FindByFromMemoryID(ctx context.Context, memoryID string) ([]domain.MemoryRelationship, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM memory_relationships WHERE from_memory_id = $1 AND tenant_id = $2 ORDER BY strength DESC`, relColumns)
	return r.queryRels(ctx, query, memoryID, tenantID)
}

// FindByToMemoryID returns incoming relationships to a memory.
func (r *RelationshipRepository) FindByToMemoryID(ctx context.Context, memoryID string) ([]domain.MemoryRelationship, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM memory_relationships WHERE to_memory_id = $1 AND tenant_id = $2 ORDER BY strength DESC`, relColumns)
	return r.queryRels(ctx, query, memoryID, tenantID)
}

// ListByTenant returns all relationships for the tenant.
func (r *RelationshipRepository) ListByTenant(ctx context.Context) ([]domain.MemoryRelationship, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM memory_relationships WHERE tenant_id = $1 ORDER BY created_at DESC`, relColumns)
	return r.queryRels(ctx, query, tenantID)
}

// FindRelatedWithMinStrength returns relationships with minimum strength.
func (r *RelationshipRepository) FindRelatedWithMinStrength(ctx context.Context, memoryID string, minStrength float64) ([]domain.MemoryRelationship, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM memory_relationships
		WHERE (from_memory_id = $1 OR to_memory_id = $1) AND tenant_id = $2 AND strength >= $3
		ORDER BY strength DESC`, relColumns)
	return r.queryRels(ctx, query, memoryID, tenantID, minStrength)
}

// UpdateStrength updates the strength of a relationship.
func (r *RelationshipRepository) UpdateStrength(ctx context.Context, id string, strength float64) (*domain.MemoryRelationship, error) {
	now := time.Now()
	query := fmt.Sprintf(`UPDATE memory_relationships SET strength = $1, last_used_at = $2 WHERE id = $3
		RETURNING %s`, relColumns)
	var rel domain.MemoryRelationship
	err := r.scanRel(&rel, r.pool.QueryRow(ctx, query, strength, now, id).Scan)
	if err != nil {
		return nil, fmt.Errorf("updating strength: %w", err)
	}
	return &rel, nil
}

// DeleteByFromAndTo deletes a relationship between two memories.
func (r *RelationshipRepository) DeleteByFromAndTo(ctx context.Context, fromID, toID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM memory_relationships WHERE from_memory_id = $1 AND to_memory_id = $2`, fromID, toID)
	return err
}

// DeleteByMemoryID deletes all relationships for a memory.
func (r *RelationshipRepository) DeleteByMemoryID(ctx context.Context, memoryID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM memory_relationships WHERE from_memory_id = $1 OR to_memory_id = $1`, memoryID)
	return err
}

func (r *RelationshipRepository) queryRels(ctx context.Context, query string, args ...any) ([]domain.MemoryRelationship, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying relationships: %w", err)
	}
	defer rows.Close()

	var rels []domain.MemoryRelationship
	for rows.Next() {
		var rel domain.MemoryRelationship
		if err := rows.Scan(
			&rel.ID, &rel.FromMemoryID, &rel.ToMemoryID, &rel.Type,
			&rel.Frequency, &rel.Severity, &rel.Strength, &rel.Description,
			&rel.CreatedAt, &rel.LastUsedAt, &rel.TenantID,
		); err != nil {
			return nil, fmt.Errorf("scanning relationship: %w", err)
		}
		rels = append(rels, rel)
	}
	return rels, nil
}
