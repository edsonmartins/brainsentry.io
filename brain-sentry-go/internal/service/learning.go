package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// LearningConfig holds configuration for the learning system.
type LearningConfig struct {
	// Promotion/demotion thresholds
	PromoteHelpfulRate   float64       // Min helpfulness rate to promote (default: 0.7)
	PromoteMinFeedback   int           // Min feedback count to consider promotion (default: 3)
	DemoteHelpfulRate    float64       // Max helpfulness rate to demote (default: 0.3)
	DemoteMinFeedback    int           // Min feedback count to consider demotion (default: 5)
	ObsolescenceDays     int           // Days without access before marking obsolete (default: 90)
	ProcessingInterval   time.Duration // How often to run the learning cycle (default: 1h)
}

// DefaultLearningConfig returns a LearningConfig with sensible defaults.
func DefaultLearningConfig() LearningConfig {
	return LearningConfig{
		PromoteHelpfulRate: 0.7,
		PromoteMinFeedback: 3,
		DemoteHelpfulRate:  0.3,
		DemoteMinFeedback:  5,
		ObsolescenceDays:   90,
		ProcessingInterval: 1 * time.Hour,
	}
}

// LearningService handles auto-promotion/demotion and obsolescence detection.
type LearningService struct {
	memoryRepo   *postgres.MemoryRepository
	tenantRepo   *postgres.TenantRepository
	auditService *AuditService
	config       LearningConfig
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewLearningService creates a new LearningService.
func NewLearningService(
	memoryRepo *postgres.MemoryRepository,
	tenantRepo *postgres.TenantRepository,
	auditService *AuditService,
	config LearningConfig,
) *LearningService {
	return &LearningService{
		memoryRepo:   memoryRepo,
		tenantRepo:   tenantRepo,
		auditService: auditService,
		config:       config,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the background learning cycle.
// It accepts fallback tenant IDs used when the tenant repository is not available.
func (s *LearningService) Start(fallbackTenantIDs []string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.config.ProcessingInterval)
		defer ticker.Stop()

		// Run immediately on start
		for _, tid := range s.discoverTenants(fallbackTenantIDs) {
			s.processTenant(tid)
		}

		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				for _, tid := range s.discoverTenants(fallbackTenantIDs) {
					s.processTenant(tid)
				}
			}
		}
	}()
}

// discoverTenants returns all active tenant IDs from the database, falling back to provided IDs.
func (s *LearningService) discoverTenants(fallback []string) []string {
	if s.tenantRepo == nil {
		return fallback
	}
	tenants, err := s.tenantRepo.List(context.Background())
	if err != nil {
		slog.Warn("failed to discover tenants, using fallback", "error", err)
		return fallback
	}
	ids := make([]string, 0, len(tenants))
	for _, t := range tenants {
		ids = append(ids, t.ID)
	}
	if len(ids) == 0 {
		return fallback
	}
	return ids
}

// Stop gracefully stops the learning cycle.
func (s *LearningService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// ProcessTenantNow runs the learning cycle for a specific tenant immediately.
func (s *LearningService) ProcessTenantNow(ctx context.Context) error {
	tenantID := tenant.FromContext(ctx)
	return s.processTenantCtx(ctx, tenantID)
}

func (s *LearningService) processTenant(tenantID string) {
	ctx := tenant.WithTenant(context.Background(), tenantID)
	if err := s.processTenantCtx(ctx, tenantID); err != nil {
		slog.Warn("learning cycle failed", "tenant", tenantID, "error", err)
	}
}

func (s *LearningService) processTenantCtx(ctx context.Context, tenantID string) error {
	memories, err := s.memoryRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	var promoted, demoted, obsoleted int

	for i := range memories {
		m := &memories[i]
		changed := false

		// Auto-promotion: high helpfulness -> increase importance
		if s.shouldPromote(m) {
			newImportance := promoteImportance(m.Importance)
			if newImportance != m.Importance {
				m.Importance = newImportance
				changed = true
				promoted++
			}
		}

		// Auto-demotion: low helpfulness -> decrease importance
		if s.shouldDemote(m) {
			newImportance := demoteImportance(m.Importance)
			if newImportance != m.Importance {
				m.Importance = newImportance
				changed = true
				demoted++
			}
		}

		// Obsolescence: no access in N days -> mark as MINOR
		if s.isObsolete(m) && m.Importance != domain.ImportanceMinor {
			m.Importance = domain.ImportanceMinor
			changed = true
			obsoleted++
		}

		if changed {
			m.Version++
			if err := s.memoryRepo.Update(ctx, m); err != nil {
				slog.Warn("failed to update memory during learning cycle",
					"memoryId", m.ID, "error", err)
			}
		}
	}

	if promoted > 0 || demoted > 0 || obsoleted > 0 {
		slog.Info("learning cycle completed",
			"tenant", tenantID,
			"promoted", promoted,
			"demoted", demoted,
			"obsoleted", obsoleted,
			"total", len(memories),
		)
	}

	return nil
}

func (s *LearningService) shouldPromote(m *domain.Memory) bool {
	totalFeedback := m.HelpfulCount + m.NotHelpfulCount
	if totalFeedback < s.config.PromoteMinFeedback {
		return false
	}
	return m.HelpfulnessRate() >= s.config.PromoteHelpfulRate
}

func (s *LearningService) shouldDemote(m *domain.Memory) bool {
	totalFeedback := m.HelpfulCount + m.NotHelpfulCount
	if totalFeedback < s.config.DemoteMinFeedback {
		return false
	}
	return m.HelpfulnessRate() <= s.config.DemoteHelpfulRate
}

func (s *LearningService) isObsolete(m *domain.Memory) bool {
	if m.LastAccessedAt == nil {
		// Use created_at as reference if never accessed
		return time.Since(m.CreatedAt) > time.Duration(s.config.ObsolescenceDays)*24*time.Hour
	}
	return time.Since(*m.LastAccessedAt) > time.Duration(s.config.ObsolescenceDays)*24*time.Hour
}

func promoteImportance(current domain.ImportanceLevel) domain.ImportanceLevel {
	switch current {
	case domain.ImportanceMinor:
		return domain.ImportanceImportant
	case domain.ImportanceImportant:
		return domain.ImportanceCritical
	default:
		return current // Already critical
	}
}

func demoteImportance(current domain.ImportanceLevel) domain.ImportanceLevel {
	switch current {
	case domain.ImportanceCritical:
		return domain.ImportanceImportant
	case domain.ImportanceImportant:
		return domain.ImportanceMinor
	default:
		return current // Already minor
	}
}
