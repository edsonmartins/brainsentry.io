package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// SessionRepository handles session persistence in PostgreSQL.
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

// Create inserts a new session.
func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, tenant_id, status, started_at, last_activity_at, ended_at, expires_at, memory_count, interception_count, note_count)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.UserID, s.TenantID, s.Status, s.StartedAt, s.LastActivityAt, s.EndedAt, s.ExpiresAt,
		s.MemoryCount, s.InterceptionCount, s.NoteCount,
	)
	if err != nil {
		return fmt.Errorf("inserting session: %w", err)
	}
	return nil
}

// FindByID retrieves a session by ID.
func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	var s domain.Session
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, tenant_id, status, started_at, last_activity_at, ended_at, expires_at, memory_count, interception_count, note_count
		FROM sessions WHERE id = $1`, id).Scan(
		&s.ID, &s.UserID, &s.TenantID, &s.Status, &s.StartedAt, &s.LastActivityAt, &s.EndedAt, &s.ExpiresAt,
		&s.MemoryCount, &s.InterceptionCount, &s.NoteCount,
	)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}
	return &s, nil
}

// Update updates a session.
func (r *SessionRepository) Update(ctx context.Context, s *domain.Session) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET status=$1, last_activity_at=$2, ended_at=$3, expires_at=$4, memory_count=$5, interception_count=$6, note_count=$7
		WHERE id=$8`,
		s.Status, s.LastActivityAt, s.EndedAt, s.ExpiresAt, s.MemoryCount, s.InterceptionCount, s.NoteCount, s.ID,
	)
	if err != nil {
		return fmt.Errorf("updating session: %w", err)
	}
	return nil
}

// FindActiveByTenant returns all active sessions for a tenant.
func (r *SessionRepository) FindActiveByTenant(ctx context.Context) ([]*domain.Session, error) {
	tenantID := tenant.FromContext(ctx)
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, tenant_id, status, started_at, last_activity_at, ended_at, expires_at, memory_count, interception_count, note_count
		FROM sessions WHERE tenant_id = $1 AND status = 'ACTIVE' ORDER BY last_activity_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var s domain.Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.TenantID, &s.Status, &s.StartedAt, &s.LastActivityAt, &s.EndedAt, &s.ExpiresAt,
			&s.MemoryCount, &s.InterceptionCount, &s.NoteCount); err != nil {
			return nil, fmt.Errorf("scanning session: %w", err)
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

// ExpireOld expires sessions past their TTL or idle too long.
func (r *SessionRepository) ExpireOld(ctx context.Context, maxIdleTime time.Duration) (int64, error) {
	now := time.Now()
	idleThreshold := now.Add(-maxIdleTime)

	tag, err := r.pool.Exec(ctx,
		`UPDATE sessions SET status = 'EXPIRED', ended_at = $1
		WHERE status = 'ACTIVE' AND (expires_at < $1 OR last_activity_at < $2)`,
		now, idleThreshold,
	)
	if err != nil {
		return 0, fmt.Errorf("expiring sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}

// DeleteOldCompleted removes completed/expired sessions older than the given duration.
func (r *SessionRepository) DeleteOldCompleted(ctx context.Context, olderThan time.Duration) (int64, error) {
	threshold := time.Now().Add(-olderThan)
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM sessions WHERE status IN ('COMPLETED', 'EXPIRED') AND ended_at < $1`, threshold)
	if err != nil {
		return 0, fmt.Errorf("deleting old sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}
