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

// --- Extended decay tests ---

func TestGetDecayRate_AllNineTypes(t *testing.T) {
	tests := []struct {
		memType domain.MemoryType
		rate    float64
	}{
		{domain.MemoryTypePersonality, 0.001},
		{domain.MemoryTypePreference, 0.003},
		{domain.MemoryTypeSemantic, 0.005},
		{domain.MemoryTypeProcedural, 0.005},
		{domain.MemoryTypeAssociative, 0.010},
		{domain.MemoryTypeEpisodic, 0.015},
		{domain.MemoryTypeEmotion, 0.020},
		{domain.MemoryTypeTask, 0.030},
		{domain.MemoryTypeThread, 0.050},
	}
	for _, tt := range tests {
		t.Run(string(tt.memType), func(t *testing.T) {
			rate := GetDecayRate(tt.memType)
			if rate != tt.rate {
				t.Errorf("GetDecayRate(%s) = %f, want %f", tt.memType, rate, tt.rate)
			}
		})
	}
}

func TestDecayRateOrdering(t *testing.T) {
	// Verify cognitive stability ordering: PERSONALITY < PREFERENCE < ... < THREAD
	orderedTypes := []domain.MemoryType{
		domain.MemoryTypePersonality,
		domain.MemoryTypePreference,
		domain.MemoryTypeSemantic,
		domain.MemoryTypeAssociative,
		domain.MemoryTypeEpisodic,
		domain.MemoryTypeEmotion,
		domain.MemoryTypeTask,
		domain.MemoryTypeThread,
	}
	for i := 1; i < len(orderedTypes); i++ {
		prevRate := GetDecayRate(orderedTypes[i-1])
		currRate := GetDecayRate(orderedTypes[i])
		if currRate < prevRate {
			t.Errorf("decay ordering violated: %s (%f) should have >= rate than %s (%f)",
				orderedTypes[i], currRate, orderedTypes[i-1], prevRate)
		}
	}
}

func TestComputeDecayedRelevance_AgeZero(t *testing.T) {
	now := time.Now()
	m := &domain.Memory{
		CreatedAt:      now,
		Importance:     domain.ImportanceCritical,
		MemoryType:     domain.MemoryTypeSemantic,
		AccessCount:    5,
		InjectionCount: 3,
	}
	score := ComputeDecayedRelevance(m, now)
	// decayFactor = exp(0) = 1.0
	// importanceFactor = 2.0
	// frequencyFactor = log(5+3+1) = log(9) ≈ 2.197
	// baseScore = 5*0.3 + 3*0.4 + 0*0.3 = 2.7
	expected := 2.7 * 1.0 * 2.0 * math.Log(9)
	if math.Abs(score-expected) > 0.01 {
		t.Errorf("score at age=0: got %f, want %f", score, expected)
	}
}

func TestComputeDecayedRelevance_ImportanceFactor_AllLevels(t *testing.T) {
	now := time.Now()
	created := now.Add(-10 * 24 * time.Hour)
	levels := []struct {
		importance domain.ImportanceLevel
		factor     float64
	}{
		{domain.ImportanceCritical, 2.0},
		{domain.ImportanceImportant, 1.5},
		{domain.ImportanceMinor, 1.0},
	}
	for _, tt := range levels {
		t.Run(string(tt.importance), func(t *testing.T) {
			m := &domain.Memory{
				CreatedAt:   created,
				Importance:  tt.importance,
				MemoryType:  domain.MemoryTypeSemantic,
				AccessCount: 5,
			}
			score := ComputeDecayedRelevance(m, now)
			// With same base score, decay, and frequency, the ratio should be proportional to importance factor
			if score <= 0 {
				t.Error("expected positive score")
			}
		})
	}
	// Verify CRITICAL > IMPORTANT > MINOR
	mCritical := &domain.Memory{CreatedAt: created, Importance: domain.ImportanceCritical, MemoryType: domain.MemoryTypeSemantic, AccessCount: 5}
	mImportant := &domain.Memory{CreatedAt: created, Importance: domain.ImportanceImportant, MemoryType: domain.MemoryTypeSemantic, AccessCount: 5}
	mMinor := &domain.Memory{CreatedAt: created, Importance: domain.ImportanceMinor, MemoryType: domain.MemoryTypeSemantic, AccessCount: 5}
	sCritical := ComputeDecayedRelevance(mCritical, now)
	sImportant := ComputeDecayedRelevance(mImportant, now)
	sMinor := ComputeDecayedRelevance(mMinor, now)
	if sCritical <= sImportant || sImportant <= sMinor {
		t.Errorf("importance ordering: CRITICAL(%f) > IMPORTANT(%f) > MINOR(%f)", sCritical, sImportant, sMinor)
	}
}

func TestComputeDecayedRelevance_FrequencyFloor(t *testing.T) {
	now := time.Now()
	m := &domain.Memory{
		CreatedAt:      now,
		Importance:     domain.ImportanceMinor,
		MemoryType:     domain.MemoryTypeSemantic,
		AccessCount:    0,
		InjectionCount: 0,
	}
	score := ComputeDecayedRelevance(m, now)
	// log(0+0+1) = log(1) = 0, floored to 0.5
	// baseScore = 0, floored to 0.1
	expected := 0.1 * 1.0 * 1.0 * 0.5 * 1.0 // baseScore * decay * importance * frequency * emotional
	if math.Abs(score-expected) > 0.01 {
		t.Errorf("frequency floor: got %f, want %f", score, expected)
	}
}

func TestComputeDecayedRelevance_LastAccessOverridesCreatedAt(t *testing.T) {
	now := time.Now()
	created := now.Add(-100 * 24 * time.Hour)
	lastAccess := now.Add(-1 * 24 * time.Hour)
	m := &domain.Memory{
		CreatedAt:      created,
		LastAccessedAt: &lastAccess,
		Importance:     domain.ImportanceImportant,
		MemoryType:     domain.MemoryTypeSemantic,
		AccessCount:    5,
	}
	mNoAccess := &domain.Memory{
		CreatedAt:   created,
		Importance:  domain.ImportanceImportant,
		MemoryType:  domain.MemoryTypeSemantic,
		AccessCount: 5,
	}
	scoreAccessed := ComputeDecayedRelevance(m, now)
	scoreNoAccess := ComputeDecayedRelevance(mNoAccess, now)
	if scoreAccessed <= scoreNoAccess {
		t.Errorf("recently accessed (%f) should score higher than old (%f)", scoreAccessed, scoreNoAccess)
	}
}

func TestComputeDecayedRelevance_AllTypeComparison(t *testing.T) {
	now := time.Now()
	created := now.Add(-30 * 24 * time.Hour) // 30 days ago
	types := []domain.MemoryType{
		domain.MemoryTypePersonality,
		domain.MemoryTypePreference,
		domain.MemoryTypeSemantic,
		domain.MemoryTypeAssociative,
		domain.MemoryTypeEpisodic,
		domain.MemoryTypeTask,
		domain.MemoryTypeThread,
	}
	var prevScore float64
	for i, mt := range types {
		m := &domain.Memory{
			CreatedAt:   created,
			Importance:  domain.ImportanceImportant,
			MemoryType:  mt,
			AccessCount: 5,
		}
		score := ComputeDecayedRelevance(m, now)
		if i > 0 && score > prevScore {
			t.Errorf("type %s (rate %f, score %f) should decay faster than previous (score %f)",
				mt, GetDecayRate(mt), score, prevScore)
		}
		prevScore = score
	}
}

func TestComputeDecayedRelevance_BaseScoreFloor(t *testing.T) {
	now := time.Now()
	m := &domain.Memory{
		CreatedAt:       now,
		Importance:      domain.ImportanceMinor,
		MemoryType:      domain.MemoryTypeSemantic,
		AccessCount:     0,
		InjectionCount:  0,
		HelpfulCount:    0,
		NotHelpfulCount: 0,
	}
	score := ComputeDecayedRelevance(m, now)
	if score <= 0 {
		t.Error("score should be positive even with zero activity (baseScore floor = 0.1)")
	}
}

func TestComputeDecayedRelevance_NegativeAge_ClampedToZero(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	m := &domain.Memory{
		CreatedAt:   future,
		Importance:  domain.ImportanceImportant,
		MemoryType:  domain.MemoryTypeSemantic,
		AccessCount: 5,
	}
	score := ComputeDecayedRelevance(m, now)
	// Age should be clamped to 0, not negative (which would increase the score exponentially)
	if score <= 0 {
		t.Error("expected positive score")
	}
	// Score should be the same as age=0
	mNow := &domain.Memory{
		CreatedAt:   now,
		Importance:  domain.ImportanceImportant,
		MemoryType:  domain.MemoryTypeSemantic,
		AccessCount: 5,
	}
	scoreNow := ComputeDecayedRelevance(mNow, now)
	if math.Abs(score-scoreNow) > 0.01 {
		t.Errorf("future memory (%f) should equal current (%f) due to age clamping", score, scoreNow)
	}
}
