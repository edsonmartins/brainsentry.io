package domain

import (
	"encoding/json"
	"time"
)

// Memory is the core entity with vector embeddings for semantic search.
type Memory struct {
	ID                  string           `json:"id" db:"id"`
	Content             string           `json:"content" db:"content"`
	Summary             string           `json:"summary,omitempty" db:"summary"`
	Category            MemoryCategory   `json:"category,omitempty" db:"category"`
	Importance          ImportanceLevel  `json:"importance,omitempty" db:"importance"`
	ValidationStatus    ValidationStatus `json:"validationStatus,omitempty" db:"validation_status"`
	Embedding           []float32        `json:"-" db:"embedding"`
	Metadata            json.RawMessage  `json:"metadata,omitempty" db:"metadata"`
	Tags                []string         `json:"tags,omitempty" db:"-"`
	SourceType          string           `json:"sourceType,omitempty" db:"source_type"`
	SourceReference     string           `json:"sourceReference,omitempty" db:"source_reference"`
	CreatedBy           string           `json:"createdBy,omitempty" db:"created_by"`
	TenantID            string           `json:"tenantId" db:"tenant_id"`
	CreatedAt           time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time        `json:"updatedAt" db:"updated_at"`
	LastAccessedAt      *time.Time       `json:"lastAccessedAt,omitempty" db:"last_accessed_at"`
	Version             int              `json:"version" db:"version"`
	AccessCount         int              `json:"accessCount" db:"access_count"`
	InjectionCount      int              `json:"injectionCount" db:"injection_count"`
	HelpfulCount        int              `json:"helpfulCount" db:"helpful_count"`
	NotHelpfulCount     int              `json:"notHelpfulCount" db:"not_helpful_count"`
	CodeExample         string           `json:"codeExample,omitempty" db:"code_example"`
	ProgrammingLanguage string           `json:"programmingLanguage,omitempty" db:"programming_language"`
	MemoryType          MemoryType       `json:"memoryType,omitempty" db:"memory_type"`
	DeletedAt           *time.Time       `json:"deletedAt,omitempty" db:"deleted_at"`
	EmotionalWeight     float64          `json:"emotionalWeight" db:"emotional_weight"`
	SimHash             string           `json:"simHash,omitempty" db:"sim_hash"`
	ValidFrom           *time.Time       `json:"validFrom,omitempty" db:"valid_from"`
	ValidTo             *time.Time       `json:"validTo,omitempty" db:"valid_to"`
	DecayRate           float64          `json:"decayRate" db:"decay_rate"`
	SupersededBy        string           `json:"supersededBy,omitempty" db:"superseded_by"`
}

// HelpfulnessRate returns the ratio of helpful feedback.
func (m *Memory) HelpfulnessRate() float64 {
	total := m.HelpfulCount + m.NotHelpfulCount
	if total == 0 {
		return 0
	}
	return float64(m.HelpfulCount) / float64(total)
}

// RelevanceScore combines access frequency, injection rate, and helpfulness.
func (m *Memory) RelevanceScore() float64 {
	accessScore := float64(m.AccessCount) * 0.3
	injectionScore := float64(m.InjectionCount) * 0.4
	helpfulScore := m.HelpfulnessRate() * 0.3
	return accessScore + injectionScore + helpfulScore
}
