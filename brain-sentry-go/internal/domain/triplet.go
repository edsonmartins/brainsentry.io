package domain

import "time"

// Triplet represents an atomic (Subject, Predicate, Object) knowledge unit
// extracted from memory content. Each triplet is embeddable via its formatted
// text representation and enables graph-aware semantic search.
type Triplet struct {
	ID            string    `json:"id" db:"id"`
	TenantID      string    `json:"tenantId" db:"tenant_id"`
	MemoryID      string    `json:"memoryId" db:"memory_id"`         // source memory
	Subject       string    `json:"subject" db:"subject"`
	Predicate     string    `json:"predicate" db:"predicate"`
	Object        string    `json:"object" db:"object"`
	Text          string    `json:"text" db:"text"`                   // "subject→predicate→object"
	Embedding     []float32 `json:"-" db:"embedding"`
	Confidence    float64   `json:"confidence" db:"confidence"`       // 0-1, from LLM
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	// Feedback blending weight (0.5 neutral, >0.5 boosted, <0.5 penalized).
	FeedbackWeight float64 `json:"feedbackWeight" db:"feedback_weight"`
}
