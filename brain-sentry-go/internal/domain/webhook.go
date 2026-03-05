package domain

import "time"

// WebhookEventType represents types of events that can trigger webhooks.
type WebhookEventType string

const (
	WebhookMemoryCreated    WebhookEventType = "memory.created"
	WebhookMemoryUpdated    WebhookEventType = "memory.updated"
	WebhookMemoryDeleted    WebhookEventType = "memory.deleted"
	WebhookMemoryFlagged    WebhookEventType = "memory.flagged"
	WebhookInterception     WebhookEventType = "interception.completed"
	WebhookSessionStarted   WebhookEventType = "session.started"
	WebhookSessionEnded     WebhookEventType = "session.ended"
	WebhookConflictDetected WebhookEventType = "conflict.detected"
)

// Webhook represents a registered webhook endpoint.
type Webhook struct {
	ID        string             `json:"id" db:"id"`
	TenantID  string             `json:"tenantId" db:"tenant_id"`
	URL       string             `json:"url" db:"url"`
	Secret    string             `json:"secret,omitempty" db:"secret"`
	Events    []WebhookEventType `json:"events" db:"-"`
	Active    bool               `json:"active" db:"active"`
	CreatedAt time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt" db:"updated_at"`
	LastError string             `json:"lastError,omitempty" db:"last_error"`
	FailCount int                `json:"failCount" db:"fail_count"`
}

// WebhookDelivery tracks a webhook delivery attempt.
type WebhookDelivery struct {
	ID         string           `json:"id"`
	WebhookID  string           `json:"webhookId"`
	Event      WebhookEventType `json:"event"`
	Payload    string           `json:"payload"`
	StatusCode int              `json:"statusCode"`
	Success    bool             `json:"success"`
	Error      string           `json:"error,omitempty"`
	Timestamp  time.Time        `json:"timestamp"`
	LatencyMs  int64            `json:"latencyMs"`
}

// ConflictResult represents a detected conflict between memories.
type ConflictResult struct {
	Memory1ID   string  `json:"memory1Id"`
	Memory2ID   string  `json:"memory2Id"`
	Memory1Summary string `json:"memory1Summary"`
	Memory2Summary string `json:"memory2Summary"`
	ConflictType   string `json:"conflictType"`
	Description    string `json:"description"`
	Confidence     float64 `json:"confidence"`
	Suggestion     string `json:"suggestion"`
}
