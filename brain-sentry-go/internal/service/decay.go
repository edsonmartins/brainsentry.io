package service

import (
	"math"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

// DefaultDecayRates maps memory types to daily decay rates.
// Lower rate = slower decay = more stable memory.
var DefaultDecayRates = map[domain.MemoryType]float64{
	domain.MemoryTypePersonality: 0.001, // very stable (half-life ~693 days)
	domain.MemoryTypeSemantic:    0.005, // stable (half-life ~139 days)
	domain.MemoryTypeProcedural:  0.005, // stable (half-life ~139 days)
	domain.MemoryTypePreference:  0.003, // fairly stable (half-life ~231 days)
	domain.MemoryTypeEpisodic:    0.015, // moderate decay (half-life ~46 days)
	domain.MemoryTypeAssociative: 0.010, // moderate decay (half-life ~69 days)
	domain.MemoryTypeTask:        0.030, // fast decay (half-life ~23 days)
	domain.MemoryTypeThread:      0.050, // very fast decay (half-life ~14 days)
	domain.MemoryTypeEmotion:     0.020, // moderate (half-life ~35 days)
}

// GetDecayRate returns the decay rate for a memory type, falling back to default.
func GetDecayRate(memoryType domain.MemoryType) float64 {
	if rate, ok := DefaultDecayRates[memoryType]; ok {
		return rate
	}
	return 0.01 // default moderate decay
}

// ComputeDecayedRelevance computes the time-decayed relevance score for a memory.
// Formula: baseScore * exp(-decayRate * ageDays) * importanceFactor * log(frequency+1) * emotionalFactor
func ComputeDecayedRelevance(m *domain.Memory, now time.Time) float64 {
	baseScore := m.RelevanceScore()
	if baseScore <= 0 {
		baseScore = 0.1 // minimum base score
	}

	// Age in days from last access (or creation if never accessed)
	refTime := m.CreatedAt
	if m.LastAccessedAt != nil {
		refTime = *m.LastAccessedAt
	}
	ageDays := now.Sub(refTime).Hours() / 24.0
	if ageDays < 0 {
		ageDays = 0
	}

	// Decay rate (use memory's custom rate or default for type)
	rate := m.DecayRate
	if rate <= 0 {
		rate = GetDecayRate(m.MemoryType)
	}

	// Exponential decay
	decayFactor := math.Exp(-rate * ageDays)

	// Importance multiplier
	importanceFactor := 1.0
	switch m.Importance {
	case domain.ImportanceCritical:
		importanceFactor = 2.0
	case domain.ImportanceImportant:
		importanceFactor = 1.5
	case domain.ImportanceMinor:
		importanceFactor = 1.0
	}

	// Frequency boost: log(totalAccess + 1)
	totalAccess := m.AccessCount + m.InjectionCount
	frequencyFactor := math.Log(float64(totalAccess) + 1)
	if frequencyFactor < 0.5 {
		frequencyFactor = 0.5 // minimum frequency factor
	}

	// Emotional boost: memories with high |emotional_weight| decay slower
	emotionalFactor := 1.0 + 0.5*math.Abs(m.EmotionalWeight)

	return baseScore * decayFactor * importanceFactor * frequencyFactor * emotionalFactor
}

// IsExpired checks if a memory has passed its valid_to date.
func IsExpired(m *domain.Memory, now time.Time) bool {
	if m.ValidTo == nil {
		return false
	}
	return now.After(*m.ValidTo)
}

// IsActive checks if a memory is currently within its valid time range.
func IsActive(m *domain.Memory, now time.Time) bool {
	if m.ValidFrom != nil && now.Before(*m.ValidFrom) {
		return false
	}
	if m.ValidTo != nil && now.After(*m.ValidTo) {
		return false
	}
	return true
}
