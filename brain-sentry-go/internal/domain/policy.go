package domain

import (
	"encoding/json"
	"time"
)

// PolicySeverity indicates the impact of a policy violation.
type PolicySeverity string

const (
	PolicyInfo     PolicySeverity = "info"
	PolicyWarning  PolicySeverity = "warning"
	PolicyError    PolicySeverity = "error"
	PolicyCritical PolicySeverity = "critical"
)

// PolicyRuleType enumerates supported rule kinds evaluated by the engine.
type PolicyRuleType string

const (
	PolicyMinConfidence   PolicyRuleType = "min_confidence"
	PolicyRequiresMemory  PolicyRuleType = "requires_memory"
	PolicyRequiresEntity  PolicyRuleType = "requires_entity"
	PolicyForbiddenOutcome PolicyRuleType = "forbidden_outcome"
	PolicyRequiresReasoning PolicyRuleType = "requires_reasoning"
	PolicyCategoryBlocked  PolicyRuleType = "category_blocked"
)

// Policy is a tenant-scoped, versioned governance rule that the PolicyEngine
// evaluates against Decision records. Policies are first-class objects so they
// can be queried, audited, and superseded rather than baked into code.
type Policy struct {
	ID          string          `json:"id" db:"id"`
	TenantID    string          `json:"tenantId" db:"tenant_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Category    string          `json:"category" db:"category"`
	Severity    PolicySeverity  `json:"severity" db:"severity"`
	RuleType    PolicyRuleType  `json:"ruleType" db:"rule_type"`
	RuleConfig  json.RawMessage `json:"ruleConfig" db:"rule_config"`
	Enabled     bool            `json:"enabled" db:"enabled"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
	Version     int             `json:"version" db:"version"`
}

// PolicyViolation is an evaluation result referencing the failing policy.
type PolicyViolation struct {
	PolicyID   string         `json:"policyId"`
	PolicyName string         `json:"policyName"`
	Severity   PolicySeverity `json:"severity"`
	Message    string         `json:"message"`
}
