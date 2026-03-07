package domain

import "time"

// User represents a system user with authentication and multi-tenancy.
type User struct {
	ID            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Name          string    `json:"name,omitempty" db:"name"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	TenantID      string    `json:"tenantId" db:"tenant_id"`
	Roles         []string  `json:"roles" db:"-"`
	Active        bool      `json:"active" db:"active"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
	EmailVerified bool      `json:"emailVerified" db:"email_verified"`
	Metadata      *string   `json:"metadata,omitempty" db:"metadata"`
}
