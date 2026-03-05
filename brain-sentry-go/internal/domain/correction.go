package domain

import "time"

// MemoryCorrection represents a correction/flag on a memory.
type MemoryCorrection struct {
	ID              string           `json:"id" db:"id"`
	MemoryID        string           `json:"memoryId" db:"memory_id"`
	Status          CorrectionStatus `json:"status" db:"status"`
	Reason          string           `json:"reason" db:"reason"`
	CorrectedContent string          `json:"correctedContent,omitempty" db:"corrected_content"`
	FlaggedBy       string           `json:"flaggedBy" db:"flagged_by"`
	ReviewedBy      string           `json:"reviewedBy,omitempty" db:"reviewed_by"`
	ReviewNotes     string           `json:"reviewNotes,omitempty" db:"review_notes"`
	PreviousVersion int              `json:"previousVersion" db:"previous_version"`
	TenantID        string           `json:"tenantId" db:"tenant_id"`
	CreatedAt       time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time        `json:"updatedAt" db:"updated_at"`
}
