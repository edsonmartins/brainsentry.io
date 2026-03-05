package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// UserRepository handles user persistence in PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// FindByEmail finds a user by email (case-insensitive).
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, password_hash, tenant_id, active, created_at, last_login_at, email_verified, metadata
		FROM users
		WHERE LOWER(email) = LOWER($1)`

	var user domain.User
	var lastLogin *time.Time
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash,
		&user.TenantID, &user.Active, &user.CreatedAt,
		&lastLogin, &user.EmailVerified, &user.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	user.LastLoginAt = lastLogin

	// Load roles
	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return &user, nil
}

// FindByID finds a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	tenantID := tenant.FromContext(ctx)
	query := `
		SELECT id, email, name, password_hash, tenant_id, active, created_at, last_login_at, email_verified, metadata
		FROM users
		WHERE id = $1 AND tenant_id = $2`

	var user domain.User
	var lastLogin *time.Time
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash,
		&user.TenantID, &user.Active, &user.CreatedAt,
		&lastLogin, &user.EmailVerified, &user.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	user.LastLoginAt = lastLogin

	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return &user, nil
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO users (id, email, name, password_hash, tenant_id, active, created_at, email_verified, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.Exec(ctx, query,
		user.ID, user.Email, user.Name, user.PasswordHash,
		user.TenantID, user.Active, user.CreatedAt,
		user.EmailVerified, user.Metadata,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	// Insert roles
	if len(user.Roles) > 0 {
		if err := r.insertRoles(ctx, tx, user.ID, user.Roles); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// UpdateLastLogin updates the last login timestamp.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string, t time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET last_login_at = $1 WHERE id = $2`, t, id)
	return err
}

// ListByTenant returns all users for a tenant.
func (r *UserRepository) ListByTenant(ctx context.Context) ([]domain.User, error) {
	tenantID := tenant.FromContext(ctx)
	query := `
		SELECT id, email, name, password_hash, tenant_id, active, created_at, last_login_at, email_verified, metadata
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		var lastLogin *time.Time
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.PasswordHash,
			&user.TenantID, &user.Active, &user.CreatedAt,
			&lastLogin, &user.EmailVerified, &user.Metadata,
		); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		user.LastLoginAt = lastLogin
		users = append(users, user)
	}

	// Load roles for all users
	for i := range users {
		roles, err := r.loadRoles(ctx, users[i].ID)
		if err != nil {
			return nil, err
		}
		users[i].Roles = roles
	}

	return users, nil
}

func (r *UserRepository) loadRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT role FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("loading roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scanning role: %w", err)
		}
		roles = append(roles, role)
	}
	if roles == nil {
		roles = []string{}
	}
	return roles, nil
}

func (r *UserRepository) insertRoles(ctx context.Context, tx pgx.Tx, userID string, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO user_roles (user_id, role) VALUES ")
	args := make([]any, 0, len(roles)*2)
	for i, role := range roles {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, userID, role)
	}
	_, err := tx.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("inserting roles: %w", err)
	}
	return nil
}
