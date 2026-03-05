package domain

import "time"

// SessionStatus represents the lifecycle state of a session.
type SessionStatus string

const (
	SessionActive    SessionStatus = "ACTIVE"
	SessionPaused    SessionStatus = "PAUSED"
	SessionCompleted SessionStatus = "COMPLETED"
	SessionExpired   SessionStatus = "EXPIRED"
)

// Session tracks a user interaction session.
type Session struct {
	ID              string        `json:"id" db:"id"`
	UserID          string        `json:"userId" db:"user_id"`
	TenantID        string        `json:"tenantId" db:"tenant_id"`
	Status          SessionStatus `json:"status" db:"status"`
	StartedAt       time.Time     `json:"startedAt" db:"started_at"`
	LastActivityAt  time.Time     `json:"lastActivityAt" db:"last_activity_at"`
	EndedAt         *time.Time    `json:"endedAt,omitempty" db:"ended_at"`
	ExpiresAt       time.Time     `json:"expiresAt" db:"expires_at"`
	MemoryCount     int           `json:"memoryCount" db:"memory_count"`
	InterceptionCount int         `json:"interceptionCount" db:"interception_count"`
	NoteCount       int           `json:"noteCount" db:"note_count"`
	Metadata        map[string]any `json:"metadata,omitempty" db:"-"`
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive checks if the session is active and not expired.
func (s *Session) IsActive() bool {
	return s.Status == SessionActive && !s.IsExpired()
}
