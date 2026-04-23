package domain

import (
	"encoding/json"
	"time"
)

// EventParticipant names a party involved in an event with its role.
type EventParticipant struct {
	EntityID string `json:"entityId"`
	Role     string `json:"role,omitempty"`
	Label    string `json:"label,omitempty"`
}

// Event is a typed occurrence with time and participants. Distinct from
// Memory (knowledge) and Decision (reasoning), events answer "what happened,
// who was involved, and when?". Useful for query patterns like "all contract
// signings in the last quarter".
type Event struct {
	ID             string             `json:"id" db:"id"`
	TenantID       string             `json:"tenantId" db:"tenant_id"`
	EventType      string             `json:"eventType" db:"event_type"`
	Title          string             `json:"title" db:"title"`
	Description    string             `json:"description" db:"description"`
	OccurredAt     time.Time          `json:"occurredAt" db:"occurred_at"`
	Participants   []EventParticipant `json:"participants" db:"-"`
	Attributes     json.RawMessage    `json:"attributes,omitempty" db:"attributes"`
	SourceMemoryID string             `json:"sourceMemoryId,omitempty" db:"source_memory_id"`
	Embedding      []float32          `json:"-" db:"embedding"`
	CreatedAt      time.Time          `json:"createdAt" db:"created_at"`
}
