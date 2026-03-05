package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

// BatchService handles bulk import/export of memories.
type BatchService struct {
	memoryRepo       *postgres.MemoryRepository
	embeddingService *EmbeddingService
	auditService     *AuditService
}

// NewBatchService creates a new BatchService.
func NewBatchService(
	memoryRepo *postgres.MemoryRepository,
	embeddingService *EmbeddingService,
	auditService *AuditService,
) *BatchService {
	return &BatchService{
		memoryRepo:       memoryRepo,
		embeddingService: embeddingService,
		auditService:     auditService,
	}
}

// BatchImportRequest represents a batch import request.
type BatchImportRequest struct {
	Memories       []dto.CreateMemoryRequest `json:"memories"`
	SkipDuplicates bool                      `json:"skipDuplicates,omitempty"`
	DryRun         bool                      `json:"dryRun,omitempty"`
}

// BatchImportResult contains the results of a batch import.
type BatchImportResult struct {
	Total     int      `json:"total"`
	Imported  int      `json:"imported"`
	Skipped   int      `json:"skipped"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
	DryRun    bool     `json:"dryRun"`
}

// BatchExportResult contains exported memories.
type BatchExportResult struct {
	Memories []dto.MemoryResponse `json:"memories"`
	Total    int                  `json:"total"`
	ExportedAt string             `json:"exportedAt"`
}

// Import imports a batch of memories.
func (s *BatchService) Import(ctx context.Context, req BatchImportRequest) (*BatchImportResult, error) {
	result := &BatchImportResult{
		Total:  len(req.Memories),
		DryRun: req.DryRun,
	}

	if len(req.Memories) == 0 {
		return result, nil
	}

	if len(req.Memories) > 1000 {
		return nil, fmt.Errorf("batch import limited to 1000 memories per request")
	}

	for i, memReq := range req.Memories {
		if memReq.Content == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("memory[%d]: content is required", i))
			continue
		}

		if req.DryRun {
			result.Imported++
			continue
		}

		m := &domain.Memory{
			Content:             memReq.Content,
			Summary:             memReq.Summary,
			Category:            memReq.Category,
			Importance:          memReq.Importance,
			MemoryType:          memReq.MemoryType,
			Tags:                memReq.Tags,
			SourceType:          memReq.SourceType,
			SourceReference:     memReq.SourceReference,
			CodeExample:         memReq.CodeExample,
			ProgrammingLanguage: memReq.ProgrammingLanguage,
			CreatedBy:           memReq.CreatedBy,
		}

		// Set defaults
		if m.Category == "" {
			m.Category = domain.CategoryKnowledge
		}
		if m.Importance == "" {
			m.Importance = domain.ImportanceMinor
		}
		if m.MemoryType == "" {
			m.MemoryType = domain.MemoryTypeSemantic
		}

		// Generate embedding
		if s.embeddingService != nil {
			m.Embedding = s.embeddingService.Embed(m.Content)
		}

		if err := s.memoryRepo.Create(ctx, m); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("memory[%d]: %s", i, err.Error()))
			slog.Warn("batch import failed for memory", "index", i, "error", err)
			continue
		}

		result.Imported++
	}

	return result, nil
}

// Export exports all memories for the current tenant.
func (s *BatchService) Export(ctx context.Context) (*BatchExportResult, error) {
	memories, err := s.memoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing memories for export: %w", err)
	}

	result := &BatchExportResult{
		Total:      len(memories),
		Memories:   make([]dto.MemoryResponse, 0, len(memories)),
		ExportedAt: fmt.Sprintf("%d", time.Now().Unix()),
	}

	for _, m := range memories {
		result.Memories = append(result.Memories, memoryToResponse(m))
	}

	return result, nil
}
