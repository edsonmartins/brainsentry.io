package domain

import (
	"encoding/json"
	"time"
)

// Tenant represents a multi-tenancy organization.
type Tenant struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Slug        string          `json:"slug" db:"slug"`
	Description string          `json:"description,omitempty" db:"description"`
	Active      bool            `json:"active" db:"active"`
	MaxMemories int             `json:"maxMemories" db:"max_memories"`
	MaxUsers    int             `json:"maxUsers" db:"max_users"`
	Settings    json.RawMessage `json:"settings,omitempty" db:"settings"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
}
