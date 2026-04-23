package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// PolicyRepository persists Policy records.
type PolicyRepository struct {
	pool *pgxpool.Pool
}

// NewPolicyRepository builds the repository.
func NewPolicyRepository(pool *pgxpool.Pool) *PolicyRepository {
	return &PolicyRepository{pool: pool}
}

const policyColumns = `id, tenant_id, name, description, category, severity, rule_type,
	rule_config, enabled, created_at, updated_at, version`

func scanPolicy(row pgx.Row) (*domain.Policy, error) {
	var p domain.Policy
	err := row.Scan(&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Category,
		&p.Severity, &p.RuleType, &p.RuleConfig, &p.Enabled, &p.CreatedAt, &p.UpdatedAt, &p.Version)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Create inserts a new Policy.
func (r *PolicyRepository) Create(ctx context.Context, p *domain.Policy) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.TenantID == "" {
		p.TenantID = tenant.FromContext(ctx)
	}
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	if p.Version == 0 {
		p.Version = 1
	}
	if p.Severity == "" {
		p.Severity = domain.PolicyWarning
	}
	if len(p.RuleConfig) == 0 {
		p.RuleConfig = []byte("{}")
	}

	query := fmt.Sprintf(`INSERT INTO policies (%s) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, policyColumns)
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.TenantID, p.Name, p.Description, p.Category, p.Severity, p.RuleType,
		p.RuleConfig, p.Enabled, p.CreatedAt, p.UpdatedAt, p.Version)
	return err
}

// Update modifies an existing policy and bumps its version.
func (r *PolicyRepository) Update(ctx context.Context, p *domain.Policy) error {
	tenantID := tenant.FromContext(ctx)
	p.UpdatedAt = time.Now()
	p.Version++
	_, err := r.pool.Exec(ctx,
		`UPDATE policies SET name=$1, description=$2, category=$3, severity=$4, rule_type=$5,
		 rule_config=$6, enabled=$7, updated_at=$8, version=$9
		 WHERE id=$10 AND tenant_id=$11`,
		p.Name, p.Description, p.Category, p.Severity, p.RuleType,
		p.RuleConfig, p.Enabled, p.UpdatedAt, p.Version, p.ID, tenantID)
	return err
}

// Delete removes a policy.
func (r *PolicyRepository) Delete(ctx context.Context, id string) error {
	tenantID := tenant.FromContext(ctx)
	_, err := r.pool.Exec(ctx, `DELETE FROM policies WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return err
}

// FindByID returns a single policy.
func (r *PolicyRepository) FindByID(ctx context.Context, id string) (*domain.Policy, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM policies WHERE id=$1 AND tenant_id=$2`, policyColumns)
	p, err := scanPolicy(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("policy not found: %s", id)
		}
		return nil, err
	}
	return p, nil
}

// ListForCategory returns enabled policies targeting a decision category.
// Passing the special "*" category also matches; empty category returns all.
func (r *PolicyRepository) ListForCategory(ctx context.Context, category string) ([]*domain.Policy, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM policies
		WHERE tenant_id = $1 AND enabled = TRUE AND (category = $2 OR category = '*')
		ORDER BY severity DESC, created_at ASC`, policyColumns)
	rows, err := r.pool.Query(ctx, query, tenantID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Policy
	for rows.Next() {
		p, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListAll returns all policies for the tenant, newest first.
func (r *PolicyRepository) ListAll(ctx context.Context) ([]*domain.Policy, error) {
	tenantID := tenant.FromContext(ctx)
	query := fmt.Sprintf(`SELECT %s FROM policies WHERE tenant_id=$1 ORDER BY created_at DESC`, policyColumns)
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Policy
	for rows.Next() {
		p, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
