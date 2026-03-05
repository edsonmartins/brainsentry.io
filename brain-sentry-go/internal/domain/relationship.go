package domain

import "time"

// MemoryRelationship represents a first-class relationship between memories.
type MemoryRelationship struct {
	ID           string           `json:"id" db:"id"`
	FromMemoryID string           `json:"fromMemoryId" db:"from_memory_id"`
	ToMemoryID   string           `json:"toMemoryId" db:"to_memory_id"`
	Type         RelationshipType `json:"type" db:"type"`
	Frequency    int              `json:"frequency" db:"frequency"`
	Severity     string           `json:"severity,omitempty" db:"severity"`
	Strength     float64          `json:"strength" db:"strength"`
	Description  string           `json:"description,omitempty" db:"description"`
	CreatedAt    time.Time        `json:"createdAt" db:"created_at"`
	LastUsedAt   *time.Time       `json:"lastUsedAt,omitempty" db:"last_used_at"`
	TenantID     string           `json:"tenantId" db:"tenant_id"`
}

// MemoryVersion tracks historical versions of memories.
type MemoryVersion struct {
	ID           string          `json:"id" db:"id"`
	MemoryID     string          `json:"memoryId" db:"memory_id"`
	Version      int             `json:"version" db:"version"`
	Content      string          `json:"content" db:"content"`
	Summary      string          `json:"summary,omitempty" db:"summary"`
	Category     MemoryCategory  `json:"category,omitempty" db:"category"`
	Importance   ImportanceLevel `json:"importance,omitempty" db:"importance"`
	Metadata     []byte          `json:"metadata,omitempty" db:"metadata"`
	Tags         []string        `json:"tags,omitempty" db:"-"`
	CodeExample  string          `json:"codeExample,omitempty" db:"code_example"`
	ChangedBy    string          `json:"changedBy,omitempty" db:"changed_by"`
	ChangeReason string          `json:"changeReason,omitempty" db:"change_reason"`
	ChangeType   string          `json:"changeType,omitempty" db:"change_type"`
	CreatedAt    time.Time       `json:"createdAt" db:"created_at"`
	TenantID     string          `json:"tenantId" db:"tenant_id"`
}

// ContextSummary tracks compression events.
type ContextSummary struct {
	ID                   string    `json:"id" db:"id"`
	TenantID             string    `json:"tenantId" db:"tenant_id"`
	SessionID            string    `json:"sessionId" db:"session_id"`
	OriginalTokenCount   int       `json:"originalTokenCount" db:"original_token_count"`
	CompressedTokenCount int       `json:"compressedTokenCount" db:"compressed_token_count"`
	CompressionRatio     float64   `json:"compressionRatio" db:"compression_ratio"`
	Summary              string    `json:"summary" db:"summary"`
	Goals                []string  `json:"goals,omitempty" db:"-"`
	Decisions            []string  `json:"decisions,omitempty" db:"-"`
	Errors               []string  `json:"errors,omitempty" db:"-"`
	Todos                []string  `json:"todos,omitempty" db:"-"`
	RecentWindowSize     int       `json:"recentWindowSize" db:"recent_window_size"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at"`
	ModelUsed            string    `json:"modelUsed,omitempty" db:"model_used"`
	CompressionMethod    string    `json:"compressionMethod,omitempty" db:"compression_method"`
}
