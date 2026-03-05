package domain

import "testing"

func TestMemory_HelpfulnessRate(t *testing.T) {
	tests := []struct {
		name     string
		helpful  int
		notHelp  int
		expected float64
	}{
		{"no feedback", 0, 0, 0},
		{"all helpful", 10, 0, 1.0},
		{"all not helpful", 0, 10, 0},
		{"mixed", 7, 3, 0.7},
		{"one helpful", 1, 0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{HelpfulCount: tt.helpful, NotHelpfulCount: tt.notHelp}
			rate := m.HelpfulnessRate()
			if rate != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, rate)
			}
		})
	}
}

func TestMemory_RelevanceScore(t *testing.T) {
	m := &Memory{
		AccessCount:    10,
		InjectionCount: 5,
		HelpfulCount:   8,
		NotHelpfulCount: 2,
	}

	score := m.RelevanceScore()
	if score <= 0 {
		t.Errorf("expected positive relevance score, got %f", score)
	}

	// Memory with no activity should have 0 score
	m2 := &Memory{}
	if m2.RelevanceScore() != 0 {
		t.Errorf("expected 0 score for new memory, got %f", m2.RelevanceScore())
	}
}

func TestHindsightNote_PreventionEffectiveness(t *testing.T) {
	tests := []struct {
		name       string
		occurrence int
		prevention int
		expected   float64
	}{
		{"no occurrences", 0, 0, 0},
		{"all prevented", 10, 10, 1.0},
		{"none prevented", 10, 0, 0},
		{"partial", 10, 3, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HindsightNote{
				OccurrenceCount:        tt.occurrence,
				PreventionSuccessCount: tt.prevention,
			}
			eff := h.PreventionEffectiveness()
			if eff != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, eff)
			}
		})
	}
}
