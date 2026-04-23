package domain

import (
	"encoding/json"
	"time"
)

// DecisionOutcome is the resolved state of a recorded decision.
type DecisionOutcome string

const (
	DecisionApproved DecisionOutcome = "approved"
	DecisionRejected DecisionOutcome = "rejected"
	DecisionDeferred DecisionOutcome = "deferred"
	DecisionPending  DecisionOutcome = "pending"
)

// Decision is a first-class, auditable record of a reasoning act performed by
// an agent or human operator. Unlike Memory (knowledge) and AgentTrace (call),
// a Decision captures *why* an outcome was chosen, links to the entities and
// parent decisions that influenced it, and is the anchor for precedent search,
// causal-chain analysis, and policy enforcement.
type Decision struct {
	ID               string          `json:"id" db:"id"`
	TenantID         string          `json:"tenantId" db:"tenant_id"`
	Category         string          `json:"category" db:"category"`
	Scenario         string          `json:"scenario" db:"scenario"`
	Reasoning        string          `json:"reasoning" db:"reasoning"`
	Outcome          DecisionOutcome `json:"outcome" db:"outcome"`
	Confidence       float64         `json:"confidence" db:"confidence"`
	AgentID          string          `json:"agentId,omitempty" db:"agent_id"`
	SessionID        string          `json:"sessionId,omitempty" db:"session_id"`
	ParentDecisionID string          `json:"parentDecisionId,omitempty" db:"parent_decision_id"`
	EntityIDs        []string        `json:"entityIds,omitempty" db:"-"`
	MemoryIDs        []string        `json:"memoryIds,omitempty" db:"-"`
	PolicyViolations []string        `json:"policyViolations,omitempty" db:"-"`
	Embedding        []float32       `json:"-" db:"embedding"`
	Metadata         json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time       `json:"createdAt" db:"created_at"`
	ValidFrom        *time.Time      `json:"validFrom,omitempty" db:"valid_from"`
	ValidUntil       *time.Time      `json:"validUntil,omitempty" db:"valid_until"`
	RecordedAt       time.Time       `json:"recordedAt" db:"recorded_at"`
	SupersededBy     string          `json:"supersededBy,omitempty" db:"superseded_by"`
}

// DecisionPrecedent is a Decision returned as a precedent match, paired with
// a similarity score.
type DecisionPrecedent struct {
	Decision   *Decision `json:"decision"`
	Similarity float64   `json:"similarity"`
}

// CausalNode is a single node in a decision's causal chain, capturing the
// decision itself and its relationship to the target decision (root, parent,
// sibling, child, descendant).
type CausalNode struct {
	Decision *Decision `json:"decision"`
	Depth    int       `json:"depth"`
	Relation string    `json:"relation"`
}
