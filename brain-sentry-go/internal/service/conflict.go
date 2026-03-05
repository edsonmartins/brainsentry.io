package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

// ConflictService detects contradictions between memories.
type ConflictService struct {
	memoryRepo       *postgres.MemoryRepository
	openRouter       *OpenRouterService
	embeddingService *EmbeddingService
}

// NewConflictService creates a new ConflictService.
func NewConflictService(
	memoryRepo *postgres.MemoryRepository,
	openRouter *OpenRouterService,
	embeddingService *EmbeddingService,
) *ConflictService {
	return &ConflictService{
		memoryRepo:       memoryRepo,
		openRouter:       openRouter,
		embeddingService: embeddingService,
	}
}

// DetectConflicts scans memories for contradictions.
func (s *ConflictService) DetectConflicts(ctx context.Context, targetMemoryID string) ([]domain.ConflictResult, error) {
	if s.openRouter == nil {
		return nil, fmt.Errorf("LLM service required for conflict detection")
	}

	target, err := s.memoryRepo.FindByID(ctx, targetMemoryID)
	if err != nil {
		return nil, fmt.Errorf("target memory not found: %w", err)
	}

	// Find similar memories by category
	candidates, err := s.memoryRepo.FindByCategory(ctx, target.Category)
	if err != nil {
		return nil, fmt.Errorf("finding candidates: %w", err)
	}

	var conflicts []domain.ConflictResult

	for _, candidate := range candidates {
		if candidate.ID == targetMemoryID {
			continue
		}

		// Quick filter: only check memories with high embedding similarity
		sim := cosineSimilarity(target.Embedding, candidate.Embedding)
		if sim < 0.5 {
			continue
		}

		conflict, err := s.checkConflict(ctx, target, &candidate)
		if err != nil {
			slog.Warn("conflict check failed", "target", targetMemoryID, "candidate", candidate.ID, "error", err)
			continue
		}

		if conflict != nil {
			conflicts = append(conflicts, *conflict)
		}
	}

	return conflicts, nil
}

// ScanAllConflicts scans all tenant memories for conflicts.
func (s *ConflictService) ScanAllConflicts(ctx context.Context, maxPairs int) ([]domain.ConflictResult, error) {
	if s.openRouter == nil {
		return nil, fmt.Errorf("LLM service required for conflict detection")
	}

	if maxPairs <= 0 {
		maxPairs = 50
	}

	memories, err := s.memoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing memories: %w", err)
	}

	var conflicts []domain.ConflictResult
	checked := 0

	// Group by category for efficient comparison
	groups := make(map[domain.MemoryCategory][]domain.Memory)
	for _, m := range memories {
		groups[m.Category] = append(groups[m.Category], m)
	}

	for _, groupMemories := range groups {
		for i := 0; i < len(groupMemories)-1 && checked < maxPairs; i++ {
			for j := i + 1; j < len(groupMemories) && checked < maxPairs; j++ {
				sim := cosineSimilarity(groupMemories[i].Embedding, groupMemories[j].Embedding)
				if sim < 0.5 {
					continue
				}

				checked++
				conflict, err := s.checkConflict(ctx, &groupMemories[i], &groupMemories[j])
				if err != nil {
					continue
				}
				if conflict != nil {
					conflicts = append(conflicts, *conflict)
				}
			}
		}
	}

	return conflicts, nil
}

func (s *ConflictService) checkConflict(ctx context.Context, m1, m2 *domain.Memory) (*domain.ConflictResult, error) {
	prompt := fmt.Sprintf(`Analyze these two memories for contradictions or conflicts.

Memory 1 [%s/%s]:
%s

Memory 2 [%s/%s]:
%s

Respond with JSON only:
{
  "hasConflict": true/false,
  "conflictType": "CONTRADICTION|OUTDATED|INCONSISTENT|NONE",
  "description": "brief description of the conflict",
  "confidence": 0.0-1.0,
  "suggestion": "how to resolve the conflict"
}`,
		m1.Category, m1.Importance, contentForAnalysis(m1),
		m2.Category, m2.Importance, contentForAnalysis(m2))

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a conflict detection system. Analyze memories for contradictions. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		HasConflict  bool    `json:"hasConflict"`
		ConflictType string  `json:"conflictType"`
		Description  string  `json:"description"`
		Confidence   float64 `json:"confidence"`
		Suggestion   string  `json:"suggestion"`
	}

	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		return nil, fmt.Errorf("parsing conflict result: %w", err)
	}

	if !result.HasConflict || result.Confidence < 0.6 {
		return nil, nil
	}

	return &domain.ConflictResult{
		Memory1ID:      m1.ID,
		Memory2ID:      m2.ID,
		Memory1Summary: summaryOrContent(m1),
		Memory2Summary: summaryOrContent(m2),
		ConflictType:   result.ConflictType,
		Description:    result.Description,
		Confidence:     result.Confidence,
		Suggestion:     result.Suggestion,
	}, nil
}

func contentForAnalysis(m *domain.Memory) string {
	text := m.Content
	if len(text) > 500 {
		text = text[:500] + "..."
	}
	if m.CodeExample != "" {
		text += "\nCode: " + truncate(m.CodeExample, 200)
	}
	return text
}

func summaryOrContent(m *domain.Memory) string {
	if m.Summary != "" {
		return m.Summary
	}
	return truncate(m.Content, 200)
}
