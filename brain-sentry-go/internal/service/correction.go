package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// CorrectionService handles memory corrections, flagging, and rollbacks.
type CorrectionService struct {
	memoryRepo  *postgres.MemoryRepository
	versionRepo *postgres.VersionRepository
	auditService *AuditService
}

// NewCorrectionService creates a new CorrectionService.
func NewCorrectionService(
	memoryRepo *postgres.MemoryRepository,
	versionRepo *postgres.VersionRepository,
	auditService *AuditService,
) *CorrectionService {
	return &CorrectionService{
		memoryRepo:  memoryRepo,
		versionRepo: versionRepo,
		auditService: auditService,
	}
}

// FlagMemory flags a memory as incorrect and creates a correction record.
func (s *CorrectionService) FlagMemory(ctx context.Context, memoryID string, req dto.FlagMemoryRequest) (*domain.MemoryCorrection, error) {
	m, err := s.memoryRepo.FindByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	// Update memory validation status
	m.ValidationStatus = domain.ValidationFlagged
	m.Version++
	if err := s.memoryRepo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("flagging memory: %w", err)
	}

	correction := &domain.MemoryCorrection{
		ID:               uuid.New().String(),
		MemoryID:         memoryID,
		Status:           domain.CorrectionPending,
		Reason:           req.Reason,
		CorrectedContent: req.CorrectedContent,
		FlaggedBy:        req.FlaggedBy,
		PreviousVersion:  m.Version - 1,
		TenantID:         tenant.FromContext(ctx),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Audit
	if s.auditService != nil {
		go s.auditService.LogMemoryFlagged(
			tenant.WithTenant(context.Background(), correction.TenantID),
			memoryID, req.Reason, req.FlaggedBy,
		)
	}

	return correction, nil
}

// ReviewCorrection reviews a pending correction (approve or reject).
func (s *CorrectionService) ReviewCorrection(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error) {
	m, err := s.memoryRepo.FindByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	switch req.Action {
	case "approve":
		m.ValidationStatus = domain.ValidationApproved
		slog.Info("correction approved", "memoryId", memoryID, "reviewer", req.ReviewedBy)
	case "reject":
		// Restore to approved status
		m.ValidationStatus = domain.ValidationApproved
		slog.Info("correction rejected", "memoryId", memoryID, "reviewer", req.ReviewedBy)
	default:
		return nil, fmt.Errorf("invalid action: %s (must be 'approve' or 'reject')", req.Action)
	}

	m.Version++
	if err := s.memoryRepo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("updating memory: %w", err)
	}

	return m, nil
}

// RollbackMemory rolls back a memory to a specific version.
func (s *CorrectionService) RollbackMemory(ctx context.Context, memoryID string, targetVersion int) (*domain.Memory, error) {
	if s.versionRepo == nil {
		return nil, fmt.Errorf("version history not available")
	}

	// Get the target version
	versions, err := s.versionRepo.FindByMemoryID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("getting version history: %w", err)
	}

	var targetVer *domain.MemoryVersion
	for i := range versions {
		if versions[i].Version == targetVersion {
			targetVer = &versions[i]
			break
		}
	}

	if targetVer == nil {
		return nil, fmt.Errorf("version %d not found for memory %s", targetVersion, memoryID)
	}

	// Get current memory
	m, err := s.memoryRepo.FindByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	// Archive current version before rollback
	if s.versionRepo != nil {
		bgCtx := tenant.WithTenant(context.Background(), m.TenantID)
		if err := s.versionRepo.CreateFromMemory(bgCtx, m, "rollback", fmt.Sprintf("rolling back to version %d", targetVersion), ""); err != nil {
			slog.Warn("failed to archive version before rollback", "error", err)
		}
	}

	// Apply rollback
	m.Content = targetVer.Content
	m.Summary = targetVer.Summary
	m.Category = targetVer.Category
	m.Importance = targetVer.Importance
	m.Tags = targetVer.Tags
	m.ValidationStatus = domain.ValidationApproved
	m.Version++

	if err := s.memoryRepo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("applying rollback: %w", err)
	}

	// Audit
	if s.auditService != nil {
		go s.auditService.LogMemoryRollback(
			tenant.WithTenant(context.Background(), m.TenantID),
			memoryID, targetVersion,
		)
	}

	return m, nil
}
