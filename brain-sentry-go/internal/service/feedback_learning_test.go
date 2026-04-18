package service

import (
	"math"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestFeedbackLearning_NeutralWhenNoFeedback(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())
	m := &domain.Memory{HelpfulCount: 0, NotHelpfulCount: 0}

	w := s.ComputeWeight(m)
	if w != 0.5 {
		t.Errorf("expected neutral 0.5, got %f", w)
	}
}

func TestFeedbackLearning_AllHelpful(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())
	m := &domain.Memory{HelpfulCount: 10, NotHelpfulCount: 0}

	w := s.ComputeWeight(m)
	// (10+1)/(10+0+2) = 11/12 ≈ 0.917
	if math.Abs(w-11.0/12.0) > 0.001 {
		t.Errorf("expected ~0.917, got %f", w)
	}
}

func TestFeedbackLearning_AllNotHelpful(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())
	m := &domain.Memory{HelpfulCount: 0, NotHelpfulCount: 5}

	w := s.ComputeWeight(m)
	// (0+1)/(0+5+2) = 1/7 ≈ 0.143
	if math.Abs(w-1.0/7.0) > 0.001 {
		t.Errorf("expected ~0.143, got %f", w)
	}
}

func TestFeedbackLearning_BalancedIsNeutral(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())
	m := &domain.Memory{HelpfulCount: 10, NotHelpfulCount: 10}

	w := s.ComputeWeight(m)
	// (10+1)/(10+10+2) = 11/22 = 0.5
	if math.Abs(w-0.5) > 0.001 {
		t.Errorf("expected 0.5 for balanced, got %f", w)
	}
}

func TestFeedbackLearning_MinFeedbackThreshold(t *testing.T) {
	cfg := DefaultFeedbackLearningConfig()
	cfg.MinFeedback = 5
	s := NewFeedbackLearningService(cfg)

	m := &domain.Memory{HelpfulCount: 3, NotHelpfulCount: 1}
	w := s.ComputeWeight(m)
	if w != cfg.NeutralWeight {
		t.Errorf("below threshold should return neutral, got %f", w)
	}
}

func TestFeedbackLearning_BlendBoost(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())

	boosted := s.BlendScore(1.0, 1.0)
	// alpha=0.3, weight=1.0 → multiplier = 1 + 0.3*(2*1 - 1) = 1.3
	if math.Abs(boosted-1.3) > 0.001 {
		t.Errorf("expected 1.3, got %f", boosted)
	}
}

func TestFeedbackLearning_BlendPenalize(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())

	penalized := s.BlendScore(1.0, 0.0)
	// alpha=0.3, weight=0.0 → multiplier = 1 + 0.3*(-1) = 0.7
	if math.Abs(penalized-0.7) > 0.001 {
		t.Errorf("expected 0.7, got %f", penalized)
	}
}

func TestFeedbackLearning_BlendNeutral(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())

	neutral := s.BlendScore(1.5, 0.5)
	if math.Abs(neutral-1.5) > 0.001 {
		t.Errorf("neutral weight should not change score, got %f", neutral)
	}
}

func TestFeedbackLearning_ApplyToTrace(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())

	trace := &ScoreTrace{FinalScore: 1.0}
	m := &domain.Memory{HelpfulCount: 10, NotHelpfulCount: 0}

	blended := s.ApplyFeedbackToTrace(trace, m)
	if blended == 1.0 {
		t.Error("trace should have been updated")
	}
	if trace.FinalScore != blended {
		t.Error("trace.FinalScore should match returned value")
	}
}

func TestFeedbackLearning_NilSafe(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())

	if w := s.ComputeWeight(nil); w != 0.5 {
		t.Errorf("nil memory should return neutral, got %f", w)
	}
	if score := s.ApplyFeedbackToTrace(nil, nil); score != 0 {
		t.Errorf("nil inputs should return 0, got %f", score)
	}
}

func TestFeedbackLearning_ClampsWeight(t *testing.T) {
	s := NewFeedbackLearningService(DefaultFeedbackLearningConfig())

	// Weight > 1 should clamp to 1
	high := s.BlendScore(1.0, 2.0)
	expected := s.BlendScore(1.0, 1.0)
	if math.Abs(high-expected) > 0.001 {
		t.Errorf("weight > 1 should clamp, got %f vs %f", high, expected)
	}

	// Weight < 0 should clamp to 0
	low := s.BlendScore(1.0, -0.5)
	expected = s.BlendScore(1.0, 0.0)
	if math.Abs(low-expected) > 0.001 {
		t.Errorf("weight < 0 should clamp, got %f vs %f", low, expected)
	}
}
