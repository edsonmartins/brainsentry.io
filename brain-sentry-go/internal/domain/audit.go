package domain

import (
	"encoding/json"
	"time"
)

// AuditLog records all system operations for traceability.
type AuditLog struct {
	ID               string          `json:"id" db:"id"`
	EventType        string          `json:"eventType" db:"event_type"`
	Timestamp        time.Time       `json:"timestamp" db:"timestamp"`
	UserID           string          `json:"userId,omitempty" db:"user_id"`
	SessionID        string          `json:"sessionId,omitempty" db:"session_id"`
	UserRequest      string          `json:"userRequest,omitempty" db:"user_request"`
	Decision         json.RawMessage `json:"decision,omitempty" db:"decision"`
	Reasoning        string          `json:"reasoning,omitempty" db:"reasoning"`
	Confidence       *float64        `json:"confidence,omitempty" db:"confidence"`
	InputData        json.RawMessage `json:"inputData,omitempty" db:"input_data"`
	OutputData       json.RawMessage `json:"outputData,omitempty" db:"output_data"`
	MemoriesAccessed []string        `json:"memoriesAccessed,omitempty" db:"-"`
	MemoriesCreated  []string        `json:"memoriesCreated,omitempty" db:"-"`
	MemoriesModified []string        `json:"memoriesModified,omitempty" db:"-"`
	LatencyMs        *int            `json:"latencyMs,omitempty" db:"latency_ms"`
	LLMCalls         *int            `json:"llmCalls,omitempty" db:"llm_calls"`
	TokensUsed       *int            `json:"tokensUsed,omitempty" db:"tokens_used"`
	Outcome          string          `json:"outcome,omitempty" db:"outcome"`
	ErrorMessage     string          `json:"errorMessage,omitempty" db:"error_message"`
	UserFeedback     json.RawMessage `json:"userFeedback,omitempty" db:"user_feedback"`
	TenantID         string          `json:"tenantId" db:"tenant_id"`
}
