package domain

import "time"

// AgentTraceStatus represents the outcome of an agent call.
type AgentTraceStatus string

const (
	AgentTraceSuccess AgentTraceStatus = "success"
	AgentTraceError   AgentTraceStatus = "error"
)

// AgentTrace is a formal schema for procedural memory — captures a single
// agent function call with its inputs, outputs, memory context, and outcome.
//
// Unlike Memory (which represents knowledge), AgentTrace represents a workflow
// event. Enables post-hoc analysis ("why did the agent fail?", "which query
// patterns work best?") without ad-hoc metadata.
type AgentTrace struct {
	ID               string           `json:"id" db:"id"`
	TenantID         string           `json:"tenantId" db:"tenant_id"`
	SessionID        string           `json:"sessionId,omitempty" db:"session_id"`
	AgentID          string           `json:"agentId,omitempty" db:"agent_id"`
	OriginFunction   string           `json:"originFunction" db:"origin_function"`
	WithMemory       bool             `json:"withMemory" db:"with_memory"`
	MemoryQuery      string           `json:"memoryQuery,omitempty" db:"memory_query"`
	MethodParams     map[string]any   `json:"methodParams,omitempty" db:"method_params"`
	MethodReturn     any              `json:"methodReturn,omitempty" db:"method_return"`
	MemoryContext    string           `json:"memoryContext,omitempty" db:"memory_context"`
	Status           AgentTraceStatus `json:"status" db:"status"`
	ErrorMessage     string           `json:"errorMessage,omitempty" db:"error_message"`
	Text             string           `json:"text" db:"text"` // embeddable summary
	DurationMs       int64            `json:"durationMs" db:"duration_ms"`
	CreatedAt        time.Time        `json:"createdAt" db:"created_at"`
	MemoryIDs        []string         `json:"memoryIds,omitempty" db:"memory_ids"` // memories used as context
	BelongsToSets    []string         `json:"belongsToSets,omitempty" db:"belongs_to_sets"`
}
