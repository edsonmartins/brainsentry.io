package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// AuditService handles audit log operations.
type AuditService struct {
	auditRepo *postgres.AuditRepository
}

// NewAuditService creates a new AuditService.
func NewAuditService(auditRepo *postgres.AuditRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

// LogMemoryCreated logs a memory creation event asynchronously.
func (s *AuditService) LogMemoryCreated(ctx context.Context, m *domain.Memory) {
	outputData, _ := json.Marshal(map[string]any{
		"memoryId": m.ID,
		"category": m.Category,
		"importance": m.Importance,
	})

	log := &domain.AuditLog{
		EventType:       "memory_created",
		UserID:          m.CreatedBy,
		OutputData:      outputData,
		MemoriesCreated: []string{m.ID},
		Outcome:         "success",
		TenantID:        tenant.FromContext(ctx),
		Timestamp:       time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "memory_created")
	}
}

// LogMemoryUpdated logs a memory update event.
func (s *AuditService) LogMemoryUpdated(ctx context.Context, m *domain.Memory) {
	outputData, _ := json.Marshal(map[string]any{
		"memoryId": m.ID,
		"version":  m.Version,
	})

	log := &domain.AuditLog{
		EventType:        "memory_updated",
		OutputData:       outputData,
		MemoriesModified: []string{m.ID},
		Outcome:          "success",
		TenantID:         tenant.FromContext(ctx),
		Timestamp:        time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "memory_updated")
	}
}

// LogMemoryDeleted logs a memory deletion event.
func (s *AuditService) LogMemoryDeleted(ctx context.Context, memoryID string) {
	outputData, _ := json.Marshal(map[string]any{
		"memoryId": memoryID,
	})

	log := &domain.AuditLog{
		EventType:        "memory_deleted",
		OutputData:       outputData,
		MemoriesModified: []string{memoryID},
		Outcome:          "success",
		TenantID:         tenant.FromContext(ctx),
		Timestamp:        time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "memory_deleted")
	}
}

// LogInterception logs a context injection event.
func (s *AuditService) LogInterception(ctx context.Context, userID, sessionID, prompt string, enhanced bool, memoriesAccessed []string, latencyMs int, confidence float64, llmCalls, tokensUsed int) {
	inputData, _ := json.Marshal(map[string]any{"prompt": prompt})
	outputData, _ := json.Marshal(map[string]any{"enhanced": enhanced})

	conf := confidence
	lat := latencyMs
	llm := llmCalls
	tok := tokensUsed

	log := &domain.AuditLog{
		EventType:        "context_injection",
		UserID:           userID,
		SessionID:        sessionID,
		UserRequest:      prompt,
		InputData:        inputData,
		OutputData:       outputData,
		Confidence:       &conf,
		MemoriesAccessed: memoriesAccessed,
		LatencyMs:        &lat,
		LLMCalls:         &llm,
		TokensUsed:       &tok,
		Outcome:          "success",
		TenantID:         tenant.FromContext(ctx),
		Timestamp:        time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "context_injection")
	}
}

// LogError logs an error event.
func (s *AuditService) LogError(ctx context.Context, eventType, errorMessage string) {
	log := &domain.AuditLog{
		EventType:    eventType,
		ErrorMessage: errorMessage,
		Outcome:      "failed",
		TenantID:     tenant.FromContext(ctx),
		Timestamp:    time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create error audit log", "error", err)
	}
}

// ListByTenant returns audit logs for the current tenant.
func (s *AuditService) ListByTenant(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.auditRepo.ListByTenant(ctx, limit)
}

// FindByEventType returns audit logs by event type.
func (s *AuditService) FindByEventType(ctx context.Context, eventType string, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.auditRepo.FindByEventType(ctx, eventType, limit)
}

// FindByUserID returns audit logs for a user.
func (s *AuditService) FindByUserID(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.auditRepo.FindByUserID(ctx, userID, limit)
}

// FindBySessionID returns audit logs for a session.
func (s *AuditService) FindBySessionID(ctx context.Context, sessionID string, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.auditRepo.FindBySessionID(ctx, sessionID, limit)
}

// FindRecent returns the most recent audit logs.
func (s *AuditService) FindRecent(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.auditRepo.FindRecent(ctx, limit)
}

// FindByMemoryID returns audit logs related to a specific memory.
func (s *AuditService) FindByMemoryID(ctx context.Context, memoryID string, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.auditRepo.FindByMemoryID(ctx, memoryID, limit)
}

// FindByDateRange returns audit logs within a date range.
func (s *AuditService) FindByDateRange(ctx context.Context, from, to time.Time, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.auditRepo.FindByDateRange(ctx, from, to, limit)
}

// LogEntityExtraction logs an entity extraction event.
func (s *AuditService) LogEntityExtraction(ctx context.Context, memoryID string, entityCount, relationshipCount int) {
	outputData, _ := json.Marshal(map[string]any{
		"memoryId":          memoryID,
		"entityCount":       entityCount,
		"relationshipCount": relationshipCount,
	})

	log := &domain.AuditLog{
		EventType:  "entity_extraction",
		OutputData: outputData,
		Outcome:    "success",
		TenantID:   tenant.FromContext(ctx),
		Timestamp:  time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "entity_extraction")
	}
}

// LogRelationshipCreated logs a relationship creation event.
func (s *AuditService) LogRelationshipCreated(ctx context.Context, fromID, toID, relType string) {
	outputData, _ := json.Marshal(map[string]any{
		"fromMemoryId": fromID,
		"toMemoryId":   toID,
		"type":         relType,
	})

	log := &domain.AuditLog{
		EventType:  "relationship_created",
		OutputData: outputData,
		Outcome:    "success",
		TenantID:   tenant.FromContext(ctx),
		Timestamp:  time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "relationship_created")
	}
}

// LogMemoryFlagged logs a memory flagging event.
func (s *AuditService) LogMemoryFlagged(ctx context.Context, memoryID, reason, flaggedBy string) {
	outputData, _ := json.Marshal(map[string]any{
		"memoryId": memoryID,
		"reason":   reason,
	})

	log := &domain.AuditLog{
		EventType:        "memory_flagged",
		UserID:           flaggedBy,
		OutputData:       outputData,
		MemoriesModified: []string{memoryID},
		Outcome:          "success",
		TenantID:         tenant.FromContext(ctx),
		Timestamp:        time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "memory_flagged")
	}
}

// LogMemoryRollback logs a memory rollback event.
func (s *AuditService) LogMemoryRollback(ctx context.Context, memoryID string, targetVersion int) {
	outputData, _ := json.Marshal(map[string]any{
		"memoryId":      memoryID,
		"targetVersion": targetVersion,
	})

	log := &domain.AuditLog{
		EventType:        "memory_rollback",
		OutputData:       outputData,
		MemoriesModified: []string{memoryID},
		Outcome:          "success",
		TenantID:         tenant.FromContext(ctx),
		Timestamp:        time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to create audit log", "error", err, "event", "memory_rollback")
	}
}

// GetStats returns audit statistics.
func (s *AuditService) GetStats(ctx context.Context) (map[string]int64, error) {
	return s.auditRepo.CountByEventType(ctx)
}
