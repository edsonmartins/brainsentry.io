package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// AutoForgetConfig holds configuration for auto-forget behavior.
type AutoForgetConfig struct {
	TTLEnabled             bool    // enable TTL-based expiry
	ContradictionEnabled   bool    // enable contradiction detection
	LowValueEnabled        bool    // enable low-value cleanup
	LowValueMaxAgeDays     int     // max age for low-value check (default 180)
	LowValueMaxImportance  string  // importance threshold (default MINOR)
	ContradictionThreshold float64 // Jaccard similarity threshold (default 0.9)
	MaxDeletesPerRun       int     // safety limit per run (default 50)
}

// DefaultAutoForgetConfig returns sensible defaults.
func DefaultAutoForgetConfig() AutoForgetConfig {
	return AutoForgetConfig{
		TTLEnabled: true,
		// Lexical similarity is not proof of contradiction. Automatic semantic
		// supersession is opt-in until a stronger, evidence-based detector is used.
		ContradictionEnabled:   false,
		LowValueEnabled:        true,
		LowValueMaxAgeDays:     180,
		LowValueMaxImportance:  "MINOR",
		ContradictionThreshold: 0.9,
		MaxDeletesPerRun:       50,
	}
}

// AutoForgetResult holds the result of an auto-forget run.
type AutoForgetResult struct {
	TTLExpired     int      `json:"ttl_expired"`
	Contradictions int      `json:"contradictions"`
	LowValue       int      `json:"low_value"`
	TotalDeleted   int      `json:"total_deleted"`
	DeletedIDs     []string `json:"deleted_ids,omitempty"`
	DryRun         bool     `json:"dry_run"`
}

type autoForgetMemoryRepository interface {
	List(ctx context.Context, page, size int) ([]domain.Memory, int64, error)
	Delete(ctx context.Context, id string) error
	SupersedeMemory(ctx context.Context, oldID, newID string) error
}

// AutoForgetService handles intelligent memory cleanup.
type AutoForgetService struct {
	memoryRepo   autoForgetMemoryRepository
	auditService *AuditService
	config       AutoForgetConfig
}

// NewAutoForgetService creates a new AutoForgetService.
func NewAutoForgetService(
	memoryRepo *postgres.MemoryRepository,
	auditService *AuditService,
	config AutoForgetConfig,
) *AutoForgetService {
	return &AutoForgetService{
		memoryRepo:   memoryRepo,
		auditService: auditService,
		config:       config,
	}
}

// Run executes all auto-forget mechanisms. If dryRun=true, no deletions occur.
func (s *AutoForgetService) Run(ctx context.Context, dryRun bool) (*AutoForgetResult, error) {
	result := &AutoForgetResult{DryRun: dryRun}
	var allDeletedIDs []string
	var runErrors []error
	maxDeletes := s.config.MaxDeletesPerRun
	if maxDeletes <= 0 {
		maxDeletes = 50
	}

	// 1. TTL-based expiry
	if s.config.TTLEnabled {
		ids, err := s.expireTTL(ctx, dryRun, maxDeletes-len(allDeletedIDs))
		if err != nil {
			slog.Warn("auto-forget TTL expiry failed", "error", err)
			runErrors = append(runErrors, fmt.Errorf("TTL expiry: %w", err))
		} else {
			result.TTLExpired = len(ids)
			allDeletedIDs = append(allDeletedIDs, ids...)
		}
	}

	// 2. Contradiction detection
	if s.config.ContradictionEnabled && len(allDeletedIDs) < maxDeletes {
		ids, err := s.detectContradictions(ctx, dryRun, maxDeletes-len(allDeletedIDs))
		if err != nil {
			slog.Warn("auto-forget contradiction detection failed", "error", err)
			runErrors = append(runErrors, fmt.Errorf("duplicate detection: %w", err))
		} else {
			result.Contradictions = len(ids)
			allDeletedIDs = append(allDeletedIDs, ids...)
		}
	}

	// 3. Low-value cleanup
	if s.config.LowValueEnabled && len(allDeletedIDs) < maxDeletes {
		ids, err := s.cleanupLowValue(ctx, dryRun, maxDeletes-len(allDeletedIDs))
		if err != nil {
			slog.Warn("auto-forget low-value cleanup failed", "error", err)
			runErrors = append(runErrors, fmt.Errorf("low-value cleanup: %w", err))
		} else {
			result.LowValue = len(ids)
			allDeletedIDs = append(allDeletedIDs, ids...)
		}
	}

	result.TotalDeleted = len(allDeletedIDs)
	result.DeletedIDs = allDeletedIDs

	if !dryRun && result.TotalDeleted > 0 {
		slog.Info("auto-forget completed",
			"ttl_expired", result.TTLExpired,
			"contradictions", result.Contradictions,
			"low_value", result.LowValue,
			"total_deleted", result.TotalDeleted,
		)

		if s.auditService != nil {
			auditCtx := tenant.WithTenant(context.Background(), tenant.FromContext(ctx))
			go s.auditService.LogError(auditCtx, "auto_forget",
				fmt.Sprintf("ttl=%d contradictions=%d low_value=%d total=%d",
					result.TTLExpired, result.Contradictions, result.LowValue, result.TotalDeleted))
		}
	}

	return result, errors.Join(runErrors...)
}

// expireTTL deletes memories whose ValidTo has passed.
func (s *AutoForgetService) expireTTL(ctx context.Context, dryRun bool, limit int) ([]string, error) {
	now := time.Now()
	memories, err := s.listAll(ctx, 500)
	if err != nil {
		return nil, err
	}

	var expired []string
	for _, m := range memories {
		if m.ValidTo != nil && now.After(*m.ValidTo) && m.DeletedAt == nil {
			expired = append(expired, m.ID)
			if len(expired) >= limit {
				break
			}
		}
	}

	if !dryRun {
		deleted := expired[:0]
		for _, id := range expired {
			if err := s.memoryRepo.Delete(ctx, id); err != nil {
				slog.Warn("failed to delete expired memory", "id", id, "error", err)
				continue
			}
			deleted = append(deleted, id)
		}
		expired = deleted
	}

	return expired, nil
}

// detectContradictions finds memories with very similar content (Jaccard > threshold)
// and marks the older one as superseded.
func (s *AutoForgetService) detectContradictions(ctx context.Context, dryRun bool, limit int) ([]string, error) {
	memories, err := s.listAll(ctx, 200)
	if err != nil {
		return nil, err
	}

	type memoryText struct {
		id         string
		tokens     map[string]bool
		created    time.Time
		category   domain.MemoryCategory
		normalized string
	}

	var items []memoryText
	for _, m := range memories {
		if m.DeletedAt != nil || m.SupersededBy != "" {
			continue
		}
		tokens := tokenizeForJaccard(m.Content)
		items = append(items, memoryText{
			id:         m.ID,
			tokens:     tokens,
			created:    m.CreatedAt,
			category:   m.Category,
			normalized: normalizeMemoryText(m.Content),
		})
	}

	var superseded []string
	seen := make(map[string]bool)

	for i := 0; i < len(items); i++ {
		if seen[items[i].id] {
			continue
		}
		for j := i + 1; j < len(items); j++ {
			if seen[items[j].id] {
				continue
			}
			// Only compare same category
			if items[i].category != items[j].category {
				continue
			}

			sim := jaccardSimilarity(items[i].tokens, items[j].tokens)
			// Similar wording can express the opposite fact (especially around
			// negation). Only exact normalized duplicates are safe to supersede
			// automatically; semantic contradictions require explicit review.
			if sim >= s.config.ContradictionThreshold && items[i].normalized == items[j].normalized {
				// Keep newer, supersede older
				olderID := items[i].id
				newerID := items[j].id
				if items[i].created.After(items[j].created) {
					olderID = items[j].id
					newerID = items[i].id
				}

				if !seen[olderID] {
					seen[olderID] = true
					superseded = append(superseded, olderID)

					if !dryRun {
						if err := s.memoryRepo.SupersedeMemory(ctx, olderID, newerID); err != nil {
							slog.Warn("failed to supersede contradicting memory",
								"older", olderID, "newer", newerID, "error", err)
							seen[olderID] = false
							superseded = superseded[:len(superseded)-1]
							continue
						}
					}

					if len(superseded) >= limit {
						return superseded, nil
					}
				}
			}
		}
	}

	return superseded, nil
}

// cleanupLowValue removes old memories with low importance and no recent access.
func (s *AutoForgetService) cleanupLowValue(ctx context.Context, dryRun bool, limit int) ([]string, error) {
	memories, err := s.listAll(ctx, 500)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	maxAge := time.Duration(s.config.LowValueMaxAgeDays) * 24 * time.Hour

	var lowValue []string
	for _, m := range memories {
		if m.DeletedAt != nil || m.SupersededBy != "" {
			continue
		}

		age := now.Sub(m.CreatedAt)
		if age < maxAge {
			continue
		}

		// Check importance
		if string(m.Importance) != s.config.LowValueMaxImportance {
			continue
		}

		// Check last access - if accessed recently, skip
		lastAccess := m.LastAccessedAt
		if lastAccess != nil && now.Sub(*lastAccess) < maxAge/2 {
			continue
		}

		// Check access count - if frequently accessed, skip
		if m.AccessCount > 3 {
			continue
		}
		if m.HelpfulCount > 0 || m.InjectionCount > 0 {
			continue
		}

		lowValue = append(lowValue, m.ID)
		if len(lowValue) >= limit {
			break
		}
	}

	if !dryRun {
		deleted := lowValue[:0]
		for _, id := range lowValue {
			if err := s.memoryRepo.Delete(ctx, id); err != nil {
				slog.Warn("failed to delete low-value memory", "id", id, "error", err)
				continue
			}
			deleted = append(deleted, id)
		}
		lowValue = deleted
	}

	return lowValue, nil
}

func (s *AutoForgetService) listAll(ctx context.Context, pageSize int) ([]domain.Memory, error) {
	var all []domain.Memory
	for page := 0; ; page++ {
		batch, total, err := s.memoryRepo.List(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) == 0 || int64(len(all)) >= total {
			return all, nil
		}
	}
}

func normalizeMemoryText(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(text)), " ")
}

// tokenize splits text into lowercase word tokens.
func tokenizeForJaccard(text string) map[string]bool {
	words := strings.Fields(strings.ToLower(text))
	tokens := make(map[string]bool, len(words))
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'")
		if len(w) > 1 {
			tokens[w] = true
		}
	}
	return tokens
}

// jaccardSimilarity computes the Jaccard similarity between two token sets.
func jaccardSimilarity(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	intersection := 0
	for k := range a {
		if b[k] {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}
