package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// ConsolidationService consolidates similar memories and compresses verbose ones.
type ConsolidationService struct {
	memoryRepo      *postgres.MemoryRepository
	openRouter      *OpenRouterService
	embeddingService *EmbeddingService
	auditService    *AuditService
}

// NewConsolidationService creates a new ConsolidationService.
func NewConsolidationService(
	memoryRepo *postgres.MemoryRepository,
	openRouter *OpenRouterService,
	embeddingService *EmbeddingService,
	auditService *AuditService,
) *ConsolidationService {
	return &ConsolidationService{
		memoryRepo:       memoryRepo,
		openRouter:       openRouter,
		embeddingService: embeddingService,
		auditService:     auditService,
	}
}

// ConsolidationResult contains the results of a consolidation run.
type ConsolidationResult struct {
	Consolidated int      `json:"consolidated"`
	Compressed   int      `json:"compressed"`
	MergedIDs    []string `json:"mergedIds,omitempty"`
}

// ConsolidateTenant finds and merges similar memories for the current tenant.
func (s *ConsolidationService) ConsolidateTenant(ctx context.Context, similarityThreshold float64) (*ConsolidationResult, error) {
	if s.openRouter == nil {
		return &ConsolidationResult{}, nil
	}

	memories, err := s.memoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing memories: %w", err)
	}

	if len(memories) < 2 {
		return &ConsolidationResult{}, nil
	}

	result := &ConsolidationResult{}

	// Group by category for more efficient comparison
	groups := make(map[domain.MemoryCategory][]domain.Memory)
	for _, m := range memories {
		groups[m.Category] = append(groups[m.Category], m)
	}

	for _, groupMemories := range groups {
		if len(groupMemories) < 2 {
			continue
		}

		// Find similar pairs using embedding similarity
		merged := make(map[string]bool)
		for i := 0; i < len(groupMemories)-1; i++ {
			if merged[groupMemories[i].ID] {
				continue
			}
			for j := i + 1; j < len(groupMemories); j++ {
				if merged[groupMemories[j].ID] {
					continue
				}

				sim := cosineSimilarity(groupMemories[i].Embedding, groupMemories[j].Embedding)
				if sim >= similarityThreshold {
					err := s.mergeMemories(ctx, &groupMemories[i], &groupMemories[j])
					if err != nil {
						slog.Warn("failed to merge memories",
							"id1", groupMemories[i].ID,
							"id2", groupMemories[j].ID,
							"error", err)
						continue
					}
					merged[groupMemories[j].ID] = true
					result.Consolidated++
					result.MergedIDs = append(result.MergedIDs, groupMemories[j].ID)
				}
			}
		}
	}

	// Compress verbose memories
	for _, m := range memories {
		if len(m.Content) > 2000 && s.openRouter != nil {
			compressed, err := s.compressMemory(ctx, &m)
			if err != nil {
				slog.Warn("failed to compress memory", "id", m.ID, "error", err)
				continue
			}
			if compressed {
				result.Compressed++
			}
		}
	}

	return result, nil
}

func (s *ConsolidationService) mergeMemories(ctx context.Context, primary, secondary *domain.Memory) error {
	// Use LLM to merge content
	prompt := fmt.Sprintf(`Merge these two similar memories into one concise memory. Keep all unique information.

Memory 1:
%s

Memory 2:
%s

Respond with JSON:
{"content": "merged content", "summary": "brief summary"}`, primary.Content, secondary.Content)

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a memory consolidation system. Merge similar memories while preserving all unique information. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return fmt.Errorf("LLM merge failed: %w", err)
	}

	var merged struct {
		Content string `json:"content"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &merged); err != nil {
		return fmt.Errorf("parsing merged content: %w", err)
	}

	// Update primary memory with merged content
	primary.Content = merged.Content
	if merged.Summary != "" {
		primary.Summary = merged.Summary
	}

	// Combine tags
	tagSet := make(map[string]bool)
	for _, t := range primary.Tags {
		tagSet[t] = true
	}
	for _, t := range secondary.Tags {
		tagSet[t] = true
	}
	primary.Tags = make([]string, 0, len(tagSet))
	for t := range tagSet {
		primary.Tags = append(primary.Tags, t)
	}

	// Keep higher importance
	if importanceRank(secondary.Importance) > importanceRank(primary.Importance) {
		primary.Importance = secondary.Importance
	}

	// Sum feedback counts
	primary.HelpfulCount += secondary.HelpfulCount
	primary.NotHelpfulCount += secondary.NotHelpfulCount
	primary.AccessCount += secondary.AccessCount

	// Regenerate embedding
	if s.embeddingService != nil {
		primary.Embedding = s.embeddingService.Embed(primary.Content)
	}

	primary.Version++
	if err := s.memoryRepo.Update(ctx, primary); err != nil {
		return fmt.Errorf("updating primary: %w", err)
	}

	// Delete secondary
	tenantID := tenant.FromContext(ctx)
	deleteCtx := tenant.WithTenant(context.Background(), tenantID)
	if err := s.memoryRepo.Delete(deleteCtx, secondary.ID); err != nil {
		return fmt.Errorf("deleting secondary: %w", err)
	}

	return nil
}

func (s *ConsolidationService) compressMemory(ctx context.Context, m *domain.Memory) (bool, error) {
	prompt := fmt.Sprintf(`Compress this memory content while preserving all key information, code examples, and technical details. Reduce verbosity.

Content:
%s

Respond with JSON:
{"content": "compressed content", "summary": "brief summary"}`, m.Content)

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a content compressor. Reduce verbosity while keeping all essential information. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return false, err
	}

	var compressed struct {
		Content string `json:"content"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &compressed); err != nil {
		return false, err
	}

	// Only apply if actually shorter
	if len(compressed.Content) >= len(m.Content) {
		return false, nil
	}

	m.Content = compressed.Content
	if compressed.Summary != "" {
		m.Summary = compressed.Summary
	}
	m.Version++

	if s.embeddingService != nil {
		m.Embedding = s.embeddingService.Embed(m.Content)
	}

	if err := s.memoryRepo.Update(ctx, m); err != nil {
		return false, err
	}

	return true, nil
}

func importanceRank(level domain.ImportanceLevel) int {
	switch level {
	case domain.ImportanceCritical:
		return 3
	case domain.ImportanceImportant:
		return 2
	case domain.ImportanceMinor:
		return 1
	default:
		return 0
	}
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
