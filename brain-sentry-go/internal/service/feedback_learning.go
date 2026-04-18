package service

import (
	"math"

	"github.com/integraltech/brainsentry/internal/domain"
)

// FeedbackLearningConfig controls how user feedback blends into search scoring.
//
// Alpha is the feedback influence weight in [0, 1]:
//   - 0.0: ignore feedback entirely (pure similarity/relevance)
//   - 0.3: default — modest feedback influence
//   - 1.0: feedback-only ranking
//
// NeutralWeight is the feedback weight assumed when a memory has no feedback yet.
// 0.5 is neutral (no boost, no penalty).
type FeedbackLearningConfig struct {
	Alpha         float64 // feedback influence in blended scoring
	NeutralWeight float64 // default weight for memories with no feedback
	MinFeedback   int     // minimum helpful+notHelpful count to trust the weight
}

// DefaultFeedbackLearningConfig returns sensible defaults.
func DefaultFeedbackLearningConfig() FeedbackLearningConfig {
	return FeedbackLearningConfig{
		Alpha:         0.3,
		NeutralWeight: 0.5,
		MinFeedback:   2,
	}
}

// FeedbackLearningService computes feedback-derived weights from memory
// HelpfulCount/NotHelpfulCount and blends them into search scoring.
//
// The weight is a Wilson-like smoothed ratio:
//
//   weight = (helpful + 1) / (helpful + notHelpful + 2)
//
// This maps:
//   - no feedback            → 0.5 (neutral)
//   - 1 helpful, 0 not       → 0.667
//   - 10 helpful, 0 not      → 0.917
//   - 0 helpful, 5 not       → 0.143
//   - 10 helpful, 10 not     → 0.5 (neutral)
type FeedbackLearningService struct {
	config FeedbackLearningConfig
}

// NewFeedbackLearningService creates a new FeedbackLearningService.
func NewFeedbackLearningService(config FeedbackLearningConfig) *FeedbackLearningService {
	if config.NeutralWeight == 0 {
		config.NeutralWeight = 0.5
	}
	return &FeedbackLearningService{config: config}
}

// ComputeWeight returns the feedback-derived weight for a memory in [0, 1].
// Memories without enough feedback return NeutralWeight.
func (s *FeedbackLearningService) ComputeWeight(m *domain.Memory) float64 {
	if m == nil {
		return s.config.NeutralWeight
	}

	total := m.HelpfulCount + m.NotHelpfulCount
	if total < s.config.MinFeedback {
		return s.config.NeutralWeight
	}

	// Laplace-smoothed ratio in [0, 1]
	return float64(m.HelpfulCount+1) / float64(total+2)
}

// BlendScore applies feedback influence to a base score.
//
// Formula:
//   blended = baseScore * (1 + alpha * (2*weight - 1))
//
// Where (2*weight - 1) maps weight in [0,1] to a signed multiplier in [-1, +1]:
//   - weight=1.0 (all helpful)   → +alpha boost
//   - weight=0.5 (neutral)       → no change
//   - weight=0.0 (all not helpful) → -alpha penalty
//
// This preserves ranking when feedback is absent but allows strong feedback to reorder.
func (s *FeedbackLearningService) BlendScore(baseScore, feedbackWeight float64) float64 {
	// Clamp weight to [0, 1]
	w := math.Max(0, math.Min(1, feedbackWeight))
	multiplier := 1.0 + s.config.Alpha*(2.0*w-1.0)
	return baseScore * multiplier
}

// ApplyFeedbackToTrace enriches an existing ScoreTrace by applying feedback
// to its FinalScore. Returns the blended score and updates trace in place.
func (s *FeedbackLearningService) ApplyFeedbackToTrace(trace *ScoreTrace, m *domain.Memory) float64 {
	if trace == nil || m == nil {
		return 0
	}
	weight := s.ComputeWeight(m)
	blended := s.BlendScore(trace.FinalScore, weight)
	trace.FinalScore = blended
	return blended
}

// Config returns the current configuration.
func (s *FeedbackLearningService) Config() FeedbackLearningConfig {
	return s.config
}
