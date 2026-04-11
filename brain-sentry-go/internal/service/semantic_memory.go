package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// SemanticFact represents an atomic fact extracted from memories with confidence scoring.
type SemanticFact struct {
	ID              string   `json:"id"`
	Fact            string   `json:"fact"`
	Confidence      float64  `json:"confidence"` // 0-1
	SourceMemoryIDs []string `json:"sourceMemoryIds"`
	CreatedAt       string   `json:"createdAt"`
	AccessCount     int      `json:"accessCount"`
}

// ProceduralStep represents a step in a procedural workflow.
type ProceduralStep struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
}

// ProceduralWorkflow represents an extracted workflow procedure.
type ProceduralWorkflow struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Steps           []ProceduralStep `json:"steps"`
	SourceMemoryIDs []string         `json:"sourceMemoryIds"`
	CreatedAt       string           `json:"createdAt"`
}

// SemanticMemoryService consolidates memories into higher-order semantic facts and procedural workflows.
type SemanticMemoryService struct {
	memoryRepo semanticMemoryRepository
	llm        LLMProvider
	auditSvc   *AuditService
}

type semanticMemoryRepository interface {
	List(ctx context.Context, page, size int) ([]domain.Memory, int64, error)
	Create(ctx context.Context, m *domain.Memory) error
}

// NewSemanticMemoryService creates a new SemanticMemoryService.
func NewSemanticMemoryService(
	memoryRepo *postgres.MemoryRepository,
	llm LLMProvider,
	auditSvc *AuditService,
) *SemanticMemoryService {
	return NewSemanticMemoryServiceWithRepository(memoryRepo, llm, auditSvc)
}

func NewSemanticMemoryServiceWithRepository(
	memoryRepo semanticMemoryRepository,
	llm LLMProvider,
	auditSvc *AuditService,
) *SemanticMemoryService {
	return &SemanticMemoryService{
		memoryRepo: memoryRepo,
		llm:        llm,
		auditSvc:   auditSvc,
	}
}

// SemanticConsolidationResult holds the outcome of a consolidation run.
type SemanticConsolidationResult struct {
	SemanticFacts []SemanticFact       `json:"semanticFacts"`
	Workflows     []ProceduralWorkflow `json:"workflows"`
	MemoriesUsed  int                  `json:"memoriesUsed"`
	DurationMs    int64                `json:"durationMs"`
}

const semanticExtractionPrompt = `Given a collection of related memories, extract:
1. Atomic factual statements that are consistently mentioned across multiple memories (semantic facts)
2. Procedural workflows or step-by-step processes described in the memories

Respond with valid JSON only:
{
  "facts": [
    {"fact": "statement of fact", "confidence": 0.9, "sourceCount": 3}
  ],
  "workflows": [
    {"title": "workflow name", "steps": [{"order": 1, "description": "step description"}]}
  ]
}`

// Consolidate analyzes recent memories and extracts semantic facts and procedural workflows.
func (s *SemanticMemoryService) Consolidate(ctx context.Context, minMemories int) (*SemanticConsolidationResult, error) {
	start := time.Now()

	if minMemories <= 0 {
		minMemories = 5
	}

	// Get memories for consolidation
	memories, _, err := s.memoryRepo.List(ctx, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("listing memories for consolidation: %w", err)
	}

	if len(memories) < minMemories {
		return &SemanticConsolidationResult{
			MemoriesUsed: len(memories),
			DurationMs:   time.Since(start).Milliseconds(),
		}, nil
	}

	// Group by category for targeted extraction
	groups := groupByCategory(memories)

	var allFacts []SemanticFact
	var allWorkflows []ProceduralWorkflow

	for category, mems := range groups {
		if len(mems) < 3 {
			continue
		}

		facts, workflows, err := s.extractFromGroup(ctx, category, mems)
		if err != nil {
			slog.Warn("consolidation extraction failed for category",
				"category", category, "error", err)
			continue
		}

		allFacts = append(allFacts, facts...)
		allWorkflows = append(allWorkflows, workflows...)
	}

	result := &SemanticConsolidationResult{
		SemanticFacts: allFacts,
		Workflows:     allWorkflows,
		MemoriesUsed:  len(memories),
		DurationMs:    time.Since(start).Milliseconds(),
	}

	// Store results as consolidated memories
	if len(allFacts) > 0 || len(allWorkflows) > 0 {
		s.storeConsolidation(ctx, result)
	}

	slog.Info("semantic consolidation completed",
		"facts", len(allFacts),
		"workflows", len(allWorkflows),
		"memoriesUsed", len(memories),
		"durationMs", result.DurationMs,
	)

	return result, nil
}

func (s *SemanticMemoryService) extractFromGroup(ctx context.Context, category string, memories []domain.Memory) ([]SemanticFact, []ProceduralWorkflow, error) {
	if s.llm == nil {
		return nil, nil, nil
	}

	// Build content summary for LLM
	var contentBuilder string
	var sourceIDs []string
	for i, m := range memories {
		if i >= 20 { // limit to 20 memories per group
			break
		}
		contentBuilder += fmt.Sprintf("Memory %d [%s]: %s\n\n", i+1, m.Category, truncateForLLM(m.Content, 500))
		sourceIDs = append(sourceIDs, m.ID)
	}

	response, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: semanticExtractionPrompt},
		{Role: "user", Content: fmt.Sprintf("Category: %s\n\nMemories:\n%s", category, contentBuilder)},
	})
	if err != nil {
		return nil, nil, err
	}

	var parsed struct {
		Facts []struct {
			Fact        string  `json:"fact"`
			Confidence  float64 `json:"confidence"`
			SourceCount int     `json:"sourceCount"`
		} `json:"facts"`
		Workflows []struct {
			Title string `json:"title"`
			Steps []struct {
				Order       int    `json:"order"`
				Description string `json:"description"`
			} `json:"steps"`
		} `json:"workflows"`
	}

	if err := json.Unmarshal([]byte(cleanJSON(response)), &parsed); err != nil {
		return nil, nil, fmt.Errorf("parsing consolidation response: %w", err)
	}

	now := time.Now().Format(time.RFC3339)

	var facts []SemanticFact
	for _, f := range parsed.Facts {
		facts = append(facts, SemanticFact{
			ID:              fmt.Sprintf("sf-%s-%d", category[:3], len(facts)),
			Fact:            f.Fact,
			Confidence:      f.Confidence,
			SourceMemoryIDs: sourceIDs,
			CreatedAt:       now,
		})
	}

	var workflows []ProceduralWorkflow
	for _, w := range parsed.Workflows {
		steps := make([]ProceduralStep, len(w.Steps))
		for i, s := range w.Steps {
			steps[i] = ProceduralStep{Order: s.Order, Description: s.Description}
		}
		workflows = append(workflows, ProceduralWorkflow{
			ID:              fmt.Sprintf("pw-%s-%d", category[:3], len(workflows)),
			Title:           w.Title,
			Steps:           steps,
			SourceMemoryIDs: sourceIDs,
			CreatedAt:       now,
		})
	}

	return facts, workflows, nil
}

func (s *SemanticMemoryService) storeConsolidation(ctx context.Context, result *SemanticConsolidationResult) {
	tenantID := tenant.FromContext(ctx)

	// Store semantic facts as a consolidated memory
	if len(result.SemanticFacts) > 0 {
		factsJSON, _ := json.Marshal(result.SemanticFacts)
		meta, _ := json.Marshal(map[string]any{
			"type":          "semantic_consolidation",
			"factCount":     len(result.SemanticFacts),
			"semanticFacts": result.SemanticFacts,
		})

		m := &domain.Memory{
			Content:    fmt.Sprintf("Semantic consolidation: %d facts extracted", len(result.SemanticFacts)),
			Summary:    string(factsJSON),
			Category:   domain.CategoryKnowledge,
			Importance: domain.ImportanceImportant,
			MemoryType: domain.MemoryTypeSemantic,
			Metadata:   meta,
			TenantID:   tenantID,
		}
		if err := s.memoryRepo.Create(ctx, m); err != nil {
			slog.Warn("failed to store semantic consolidation", "error", err)
		}
	}

	// Store procedural workflows
	if len(result.Workflows) > 0 {
		meta, _ := json.Marshal(map[string]any{
			"type":          "procedural_consolidation",
			"workflowCount": len(result.Workflows),
			"workflows":     result.Workflows,
		})

		m := &domain.Memory{
			Content:    fmt.Sprintf("Procedural consolidation: %d workflows extracted", len(result.Workflows)),
			Category:   domain.CategoryKnowledge,
			Importance: domain.ImportanceImportant,
			MemoryType: domain.MemoryTypeProcedural,
			Metadata:   meta,
			TenantID:   tenantID,
		}
		if err := s.memoryRepo.Create(ctx, m); err != nil {
			slog.Warn("failed to store procedural consolidation", "error", err)
		}
	}
}

func groupByCategory(memories []domain.Memory) map[string][]domain.Memory {
	groups := make(map[string][]domain.Memory)
	for _, m := range memories {
		cat := string(m.Category)
		if cat == "" {
			cat = "GENERAL"
		}
		groups[cat] = append(groups[cat], m)
	}
	return groups
}
