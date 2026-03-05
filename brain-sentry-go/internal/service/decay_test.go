package service

import (
	"math"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestGetDecayRate_KnownTypes(t *testing.T) {
	tests := []struct {
		memType domain.MemoryType
		rate    float64
	}{
		{domain.MemoryTypePersonality, 0.001},
		{domain.MemoryTypeThread, 0.050},
		{domain.MemoryTypeTask, 0.030},
		{domain.MemoryTypeSemantic, 0.005},
	}

	for _, tt := range tests {
		rate := GetDecayRate(tt.memType)
		if rate != tt.rate {
			t.Errorf("GetDecayRate(%s) = %f, want %f", tt.memType, rate, tt.rate)
		}
	}
}

func TestGetDecayRate_UnknownType(t *testing.T) {
	rate := GetDecayRate(domain.MemoryType("UNKNOWN"))
	if rate != 0.01 {
		t.Errorf("expected default 0.01, got %f", rate)
	}
}

func TestComputeDecayedRelevance_RecentMemory(t *testing.T) {
	now := time.Now()
	m := &domain.Memory{
		CreatedAt:   now,
		Importance:  domain.ImportanceCritical,
		MemoryType:  domain.MemoryTypeSemantic,
		AccessCount: 10,
	}
	score := ComputeDecayedRelevance(m, now)
	if score <= 0 {
		t.Error("expected positive score for recent memory")
	}
}

func TestComputeDecayedRelevance_OldMemory(t *testing.T) {
	now := time.Now()
	old := now.Add(-365 * 24 * time.Hour) // 1 year ago
	m := &domain.Memory{
		CreatedAt:   old,
		Importance:  domain.ImportanceMinor,
		MemoryType:  domain.MemoryTypeThread,
		AccessCount: 1,
	}
	score := ComputeDecayedRelevance(m, now)
	// Thread with 0.05 decay rate after 365 days = exp(-0.05*365) ≈ 0
	if score > 0.5 {
		t.Errorf("expected very low score for old thread memory, got %f", score)
	}
}

func TestComputeDecayedRelevance_PersonalityDecaysSlowly(t *testing.T) {
	now := time.Now()
	created := now.Add(-180 * 24 * time.Hour) // 6 months ago
	personality := &domain.Memory{
		CreatedAt:   created,
		Importance:  domain.ImportanceImportant,
		MemoryType:  domain.MemoryTypePersonality,
		AccessCount: 5,
	}
	thread := &domain.Memory{
		CreatedAt:   created,
		Importance:  domain.ImportanceImportant,
		MemoryType:  domain.MemoryTypeThread,
		AccessCount: 5,
	}
	pScore := ComputeDecayedRelevance(personality, now)
	tScore := ComputeDecayedRelevance(thread, now)
	if pScore <= tScore {
		t.Errorf("personality (%.4f) should have higher score than thread (%.4f) after 6 months", pScore, tScore)
	}
}

func TestComputeDecayedRelevance_EmotionalBoost(t *testing.T) {
	now := time.Now()
	created := now.Add(-30 * 24 * time.Hour)
	neutral := &domain.Memory{
		CreatedAt:       created,
		MemoryType:      domain.MemoryTypeSemantic,
		AccessCount:     5,
		EmotionalWeight: 0,
	}
	emotional := &domain.Memory{
		CreatedAt:       created,
		MemoryType:      domain.MemoryTypeSemantic,
		AccessCount:     5,
		EmotionalWeight: 0.9,
	}
	nScore := ComputeDecayedRelevance(neutral, now)
	eScore := ComputeDecayedRelevance(emotional, now)
	if eScore <= nScore {
		t.Errorf("emotional memory (%.4f) should score higher than neutral (%.4f)", eScore, nScore)
	}
}

func TestComputeDecayedRelevance_CustomDecayRate(t *testing.T) {
	now := time.Now()
	created := now.Add(-100 * 24 * time.Hour)
	m := &domain.Memory{
		CreatedAt:   created,
		MemoryType:  domain.MemoryTypeSemantic,
		DecayRate:   0.1, // very fast custom decay
		AccessCount: 5,
	}
	score := ComputeDecayedRelevance(m, now)
	// exp(-0.1 * 100) = exp(-10) ≈ 0.0000454
	expected := math.Exp(-0.1 * 100)
	if score > expected*100 { // rough upper bound
		t.Errorf("custom high decay rate should produce low score, got %f", score)
	}
}

func TestIsExpired_NoValidTo(t *testing.T) {
	m := &domain.Memory{}
	if IsExpired(m, time.Now()) {
		t.Error("memory without valid_to should not be expired")
	}
}

func TestIsExpired_FutureValidTo(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	m := &domain.Memory{ValidTo: &future}
	if IsExpired(m, time.Now()) {
		t.Error("memory with future valid_to should not be expired")
	}
}

func TestIsExpired_PastValidTo(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	m := &domain.Memory{ValidTo: &past}
	if !IsExpired(m, time.Now()) {
		t.Error("memory with past valid_to should be expired")
	}
}

func TestIsActive_WithinRange(t *testing.T) {
	now := time.Now()
	from := now.Add(-24 * time.Hour)
	to := now.Add(24 * time.Hour)
	m := &domain.Memory{ValidFrom: &from, ValidTo: &to}
	if !IsActive(m, now) {
		t.Error("memory within range should be active")
	}
}

func TestIsActive_BeforeRange(t *testing.T) {
	now := time.Now()
	from := now.Add(24 * time.Hour)
	m := &domain.Memory{ValidFrom: &from}
	if IsActive(m, now) {
		t.Error("memory before valid_from should not be active")
	}
}

func TestIsActive_NoConstraints(t *testing.T) {
	m := &domain.Memory{}
	if !IsActive(m, time.Now()) {
		t.Error("memory with no time constraints should be active")
	}
}
