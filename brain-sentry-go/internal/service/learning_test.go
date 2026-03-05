package service

import (
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

// newLearningServiceForTest creates a LearningService with nil repos for testing pure logic.
func newLearningServiceForTest(cfg LearningConfig) *LearningService {
	return &LearningService{
		config: cfg,
		stopCh: make(chan struct{}),
	}
}

// TestDefaultLearningConfig verifies all default values are set correctly.
func TestDefaultLearningConfig(t *testing.T) {
	cfg := DefaultLearningConfig()

	if cfg.PromoteHelpfulRate != 0.7 {
		t.Errorf("expected PromoteHelpfulRate 0.7, got %f", cfg.PromoteHelpfulRate)
	}
	if cfg.PromoteMinFeedback != 3 {
		t.Errorf("expected PromoteMinFeedback 3, got %d", cfg.PromoteMinFeedback)
	}
	if cfg.DemoteHelpfulRate != 0.3 {
		t.Errorf("expected DemoteHelpfulRate 0.3, got %f", cfg.DemoteHelpfulRate)
	}
	if cfg.DemoteMinFeedback != 5 {
		t.Errorf("expected DemoteMinFeedback 5, got %d", cfg.DemoteMinFeedback)
	}
	if cfg.ObsolescenceDays != 90 {
		t.Errorf("expected ObsolescenceDays 90, got %d", cfg.ObsolescenceDays)
	}
	if cfg.ProcessingInterval != 1*time.Hour {
		t.Errorf("expected ProcessingInterval 1h, got %v", cfg.ProcessingInterval)
	}
}

// TestShouldPromote covers the promotion logic branch conditions.
func TestShouldPromote(t *testing.T) {
	svc := newLearningServiceForTest(DefaultLearningConfig())

	tests := []struct {
		name        string
		helpful     int
		notHelpful  int
		expectPromote bool
	}{
		{
			name:          "too few feedback - not enough data",
			helpful:       2,
			notHelpful:    0,
			expectPromote: false,
		},
		{
			name:          "exactly min feedback, high rate",
			helpful:       3,
			notHelpful:    0,
			expectPromote: true,
		},
		{
			name:          "meets min feedback, rate exactly at threshold (0.7)",
			helpful:       7,
			notHelpful:    3,
			expectPromote: true,
		},
		{
			name:          "meets min feedback, rate below threshold",
			helpful:       6,
			notHelpful:    4,
			expectPromote: false,
		},
		{
			name:          "high feedback count, high rate",
			helpful:       10,
			notHelpful:    1,
			expectPromote: true,
		},
		{
			name:          "high feedback count, low rate",
			helpful:       1,
			notHelpful:    9,
			expectPromote: false,
		},
		{
			name:          "zero feedback",
			helpful:       0,
			notHelpful:    0,
			expectPromote: false,
		},
		{
			name:          "min feedback minus one, high rate",
			helpful:       2,
			notHelpful:    0,
			expectPromote: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &domain.Memory{
				HelpfulCount:    tt.helpful,
				NotHelpfulCount: tt.notHelpful,
			}
			got := svc.shouldPromote(m)
			if got != tt.expectPromote {
				t.Errorf("shouldPromote() = %v, want %v (helpful=%d, notHelpful=%d, rate=%.2f)",
					got, tt.expectPromote, tt.helpful, tt.notHelpful, m.HelpfulnessRate())
			}
		})
	}
}

// TestShouldPromoteCustomConfig verifies that custom thresholds are respected.
func TestShouldPromoteCustomConfig(t *testing.T) {
	cfg := LearningConfig{
		PromoteHelpfulRate: 0.9,
		PromoteMinFeedback: 10,
	}
	svc := newLearningServiceForTest(cfg)

	// 8 out of 10 = 0.8 rate, below custom threshold of 0.9
	m := &domain.Memory{HelpfulCount: 8, NotHelpfulCount: 2}
	if svc.shouldPromote(m) {
		t.Error("expected shouldPromote=false with 0.8 rate and 0.9 threshold")
	}

	// 9 out of 10 = 0.9 rate, exactly at threshold
	m2 := &domain.Memory{HelpfulCount: 9, NotHelpfulCount: 1}
	if !svc.shouldPromote(m2) {
		t.Error("expected shouldPromote=true with 0.9 rate at 0.9 threshold")
	}

	// Enough rate but not enough feedback (9 total, need 10)
	m3 := &domain.Memory{HelpfulCount: 9, NotHelpfulCount: 0}
	if svc.shouldPromote(m3) {
		t.Error("expected shouldPromote=false with insufficient feedback count")
	}
}

// TestShouldDemote covers the demotion logic branch conditions.
func TestShouldDemote(t *testing.T) {
	svc := newLearningServiceForTest(DefaultLearningConfig())

	tests := []struct {
		name         string
		helpful      int
		notHelpful   int
		expectDemote bool
	}{
		{
			name:         "too few feedback - not enough data",
			helpful:      1,
			notHelpful:   3,
			expectDemote: false,
		},
		{
			name:         "exactly min feedback, low rate",
			helpful:      0,
			notHelpful:   5,
			expectDemote: true,
		},
		{
			name:         "meets min feedback, rate exactly at threshold (0.3)",
			helpful:      3,
			notHelpful:   7,
			expectDemote: true,
		},
		{
			name:         "meets min feedback, rate above threshold",
			helpful:      4,
			notHelpful:   6,
			expectDemote: false,
		},
		{
			name:         "high feedback count, low rate",
			helpful:      1,
			notHelpful:   9,
			expectDemote: true,
		},
		{
			name:         "high feedback count, high rate",
			helpful:      9,
			notHelpful:   1,
			expectDemote: false,
		},
		{
			name:         "zero feedback",
			helpful:      0,
			notHelpful:   0,
			expectDemote: false,
		},
		{
			name:         "min feedback minus one, low rate",
			helpful:      0,
			notHelpful:   4,
			expectDemote: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &domain.Memory{
				HelpfulCount:    tt.helpful,
				NotHelpfulCount: tt.notHelpful,
			}
			got := svc.shouldDemote(m)
			if got != tt.expectDemote {
				t.Errorf("shouldDemote() = %v, want %v (helpful=%d, notHelpful=%d, rate=%.2f)",
					got, tt.expectDemote, tt.helpful, tt.notHelpful, m.HelpfulnessRate())
			}
		})
	}
}

// TestShouldDemoteCustomConfig verifies that custom demotion thresholds are respected.
func TestShouldDemoteCustomConfig(t *testing.T) {
	cfg := LearningConfig{
		DemoteHelpfulRate: 0.2,
		DemoteMinFeedback: 8,
	}
	svc := newLearningServiceForTest(cfg)

	// 2 out of 8 = 0.25 rate, above custom threshold of 0.2 - should not demote
	m := &domain.Memory{HelpfulCount: 2, NotHelpfulCount: 6}
	if svc.shouldDemote(m) {
		t.Error("expected shouldDemote=false with 0.25 rate and 0.2 threshold")
	}

	// 1 out of 8 = 0.125 rate, below threshold of 0.2 - should demote
	m2 := &domain.Memory{HelpfulCount: 1, NotHelpfulCount: 7}
	if !svc.shouldDemote(m2) {
		t.Error("expected shouldDemote=true with 0.125 rate at 0.2 threshold")
	}

	// Enough rate ratio but not enough feedback (7 total, need 8)
	m3 := &domain.Memory{HelpfulCount: 0, NotHelpfulCount: 7}
	if svc.shouldDemote(m3) {
		t.Error("expected shouldDemote=false with insufficient feedback count")
	}
}

// TestIsObsolete covers the obsolescence detection logic.
func TestIsObsolete(t *testing.T) {
	svc := newLearningServiceForTest(DefaultLearningConfig()) // 90 days

	now := time.Now()
	recentTime := now.Add(-30 * 24 * time.Hour)  // 30 days ago - not obsolete
	oldTime := now.Add(-100 * 24 * time.Hour)     // 100 days ago - obsolete

	tests := []struct {
		name           string
		lastAccessedAt *time.Time
		createdAt      time.Time
		expectObsolete bool
	}{
		{
			name:           "recently accessed - not obsolete",
			lastAccessedAt: &recentTime,
			createdAt:      now.Add(-200 * 24 * time.Hour),
			expectObsolete: false,
		},
		{
			name:           "accessed long ago - obsolete",
			lastAccessedAt: &oldTime,
			createdAt:      now.Add(-200 * 24 * time.Hour),
			expectObsolete: true,
		},
		{
			name:           "never accessed, recently created",
			lastAccessedAt: nil,
			createdAt:      now.Add(-30 * 24 * time.Hour),
			expectObsolete: false,
		},
		{
			name:           "never accessed, created long ago",
			lastAccessedAt: nil,
			createdAt:      now.Add(-100 * 24 * time.Hour),
			expectObsolete: true,
		},
		{
			name:           "accessed exactly at obsolescence boundary",
			lastAccessedAt: func() *time.Time { t := now.Add(-91 * 24 * time.Hour); return &t }(),
			createdAt:      now.Add(-200 * 24 * time.Hour),
			expectObsolete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &domain.Memory{
				LastAccessedAt: tt.lastAccessedAt,
				CreatedAt:      tt.createdAt,
			}
			got := svc.isObsolete(m)
			if got != tt.expectObsolete {
				t.Errorf("isObsolete() = %v, want %v", got, tt.expectObsolete)
			}
		})
	}
}

// TestIsObsoleteCustomDays verifies custom obsolescence period is used.
func TestIsObsoleteCustomDays(t *testing.T) {
	cfg := LearningConfig{ObsolescenceDays: 7}
	svc := newLearningServiceForTest(cfg)

	now := time.Now()

	// 6 days ago - not obsolete with 7-day threshold
	recent := now.Add(-6 * 24 * time.Hour)
	m := &domain.Memory{LastAccessedAt: &recent, CreatedAt: now.Add(-30 * 24 * time.Hour)}
	if svc.isObsolete(m) {
		t.Error("expected not obsolete when last accessed 6 days ago with 7-day threshold")
	}

	// 8 days ago - obsolete with 7-day threshold
	old := now.Add(-8 * 24 * time.Hour)
	m2 := &domain.Memory{LastAccessedAt: &old, CreatedAt: now.Add(-30 * 24 * time.Hour)}
	if !svc.isObsolete(m2) {
		t.Error("expected obsolete when last accessed 8 days ago with 7-day threshold")
	}
}

// TestPromoteImportance verifies importance level transitions upward.
func TestPromoteImportance(t *testing.T) {
	tests := []struct {
		input    domain.ImportanceLevel
		expected domain.ImportanceLevel
	}{
		{domain.ImportanceMinor, domain.ImportanceImportant},
		{domain.ImportanceImportant, domain.ImportanceCritical},
		{domain.ImportanceCritical, domain.ImportanceCritical}, // already at max
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := promoteImportance(tt.input)
			if got != tt.expected {
				t.Errorf("promoteImportance(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

// TestDemoteImportance verifies importance level transitions downward.
func TestDemoteImportance(t *testing.T) {
	tests := []struct {
		input    domain.ImportanceLevel
		expected domain.ImportanceLevel
	}{
		{domain.ImportanceCritical, domain.ImportanceImportant},
		{domain.ImportanceImportant, domain.ImportanceMinor},
		{domain.ImportanceMinor, domain.ImportanceMinor}, // already at min
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := demoteImportance(tt.input)
			if got != tt.expected {
				t.Errorf("demoteImportance(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

// TestPromoteDemoteSymmetry verifies that promote followed by demote returns to the original level.
func TestPromoteDemoteSymmetry(t *testing.T) {
	levels := []domain.ImportanceLevel{
		domain.ImportanceMinor,
		domain.ImportanceImportant,
	}

	for _, level := range levels {
		promoted := promoteImportance(level)
		demoted := demoteImportance(promoted)
		if demoted != level {
			t.Errorf("promote then demote of %s: got %s, want %s", level, demoted, level)
		}
	}
}

// TestPromoteImportanceIdempotentAtMax verifies that promoting Critical returns Critical.
func TestPromoteImportanceIdempotentAtMax(t *testing.T) {
	for i := 0; i < 5; i++ {
		result := promoteImportance(domain.ImportanceCritical)
		if result != domain.ImportanceCritical {
			t.Errorf("repeated promote of Critical should stay Critical, got %s", result)
		}
	}
}

// TestDemoteImportanceIdempotentAtMin verifies that demoting Minor returns Minor.
func TestDemoteImportanceIdempotentAtMin(t *testing.T) {
	for i := 0; i < 5; i++ {
		result := demoteImportance(domain.ImportanceMinor)
		if result != domain.ImportanceMinor {
			t.Errorf("repeated demote of Minor should stay Minor, got %s", result)
		}
	}
}
