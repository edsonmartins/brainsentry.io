package service

import (
	"math"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

// TestCosineSimilarity_KnownVectors tests cosine similarity with analytically known results.
func TestCosineSimilarity_KnownVectors(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
		delta    float64
	}{
		{
			name:     "identical unit vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{1, 0, 0},
			expected: 1.0,
			delta:    1e-9,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{0, 1, 0},
			expected: 0.0,
			delta:    1e-9,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{-1, 0, 0},
			expected: -1.0,
			delta:    1e-9,
		},
		{
			name:     "45-degree angle gives 1/sqrt(2)",
			a:        []float32{1, 1, 0},
			b:        []float32{1, 0, 0},
			expected: 1.0 / math.Sqrt(2),
			delta:    1e-6,
		},
		{
			name:     "scaled vectors have same similarity",
			a:        []float32{2, 0, 0},
			b:        []float32{5, 0, 0},
			expected: 1.0,
			delta:    1e-9,
		},
		{
			name:     "all-ones vectors in 3D are identical direction",
			a:        []float32{1, 1, 1},
			b:        []float32{1, 1, 1},
			expected: 1.0,
			delta:    1e-6,
		},
		{
			name:     "2D vectors at 90 degrees",
			a:        []float32{0, 1},
			b:        []float32{1, 0},
			expected: 0.0,
			delta:    1e-9,
		},
		{
			name:     "mixed sign vectors",
			a:        []float32{1, -1},
			b:        []float32{-1, 1},
			expected: -1.0,
			delta:    1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			if math.Abs(result-tt.expected) > tt.delta {
				t.Errorf("cosineSimilarity(%v, %v) = %f, want %f (delta %g)",
					tt.a, tt.b, result, tt.expected, tt.delta)
			}
		})
	}
}

// TestCosineSimilarity_EdgeCases tests edge cases with empty and zero vectors.
func TestCosineSimilarity_EdgeCases(t *testing.T) {
	t.Run("empty vectors return 0", func(t *testing.T) {
		result := cosineSimilarity([]float32{}, []float32{})
		if result != 0.0 {
			t.Errorf("expected 0 for empty vectors, got %f", result)
		}
	})

	t.Run("nil vectors return 0", func(t *testing.T) {
		result := cosineSimilarity(nil, nil)
		if result != 0.0 {
			t.Errorf("expected 0 for nil vectors, got %f", result)
		}
	})

	t.Run("different length vectors return 0", func(t *testing.T) {
		result := cosineSimilarity([]float32{1, 0}, []float32{1, 0, 0})
		if result != 0.0 {
			t.Errorf("expected 0 for different-length vectors, got %f", result)
		}
	})

	t.Run("zero vector returns 0", func(t *testing.T) {
		result := cosineSimilarity([]float32{0, 0, 0}, []float32{1, 2, 3})
		if result != 0.0 {
			t.Errorf("expected 0 when one vector is zero, got %f", result)
		}
	})

	t.Run("both zero vectors return 0", func(t *testing.T) {
		result := cosineSimilarity([]float32{0, 0}, []float32{0, 0})
		if result != 0.0 {
			t.Errorf("expected 0 for both zero vectors, got %f", result)
		}
	})

	t.Run("single element identical", func(t *testing.T) {
		result := cosineSimilarity([]float32{5}, []float32{5})
		if math.Abs(result-1.0) > 1e-9 {
			t.Errorf("expected 1.0 for identical single-element vectors, got %f", result)
		}
	})

	t.Run("single element opposite", func(t *testing.T) {
		result := cosineSimilarity([]float32{3}, []float32{-3})
		if math.Abs(result-(-1.0)) > 1e-9 {
			t.Errorf("expected -1.0 for opposite single-element vectors, got %f", result)
		}
	})
}

// TestCosineSimilarity_Symmetry verifies that cosine similarity is symmetric (a,b == b,a).
func TestCosineSimilarity_Symmetry(t *testing.T) {
	a := []float32{0.5, 0.3, 0.8, 0.1}
	b := []float32{0.2, 0.9, 0.4, 0.7}

	ab := cosineSimilarity(a, b)
	ba := cosineSimilarity(b, a)

	if math.Abs(ab-ba) > 1e-9 {
		t.Errorf("cosineSimilarity is not symmetric: f(a,b)=%f, f(b,a)=%f", ab, ba)
	}
}

// TestCosineSimilarity_ResultRange verifies the result is always in [-1, 1].
func TestCosineSimilarity_ResultRange(t *testing.T) {
	vectors := [][]float32{
		{1, 0, 0},
		{-1, 0, 0},
		{0, 1, 0},
		{1, 1, 1},
		{-1, -1, -1},
		{0.5, 0.5, 0.707},
		{100, 200, 300},
	}

	for i, a := range vectors {
		for j, b := range vectors {
			result := cosineSimilarity(a, b)
			if result < -1.0-1e-9 || result > 1.0+1e-9 {
				t.Errorf("cosineSimilarity(vectors[%d], vectors[%d]) = %f out of range [-1, 1]",
					i, j, result)
			}
		}
	}
}

// TestImportanceRank verifies the ordering of importance levels.
func TestImportanceRank(t *testing.T) {
	tests := []struct {
		level    domain.ImportanceLevel
		expected int
	}{
		{domain.ImportanceCritical, 3},
		{domain.ImportanceImportant, 2},
		{domain.ImportanceMinor, 1},
		{"UNKNOWN", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			got := importanceRank(tt.level)
			if got != tt.expected {
				t.Errorf("importanceRank(%s) = %d, want %d", tt.level, got, tt.expected)
			}
		})
	}
}

// TestImportanceRank_Ordering verifies that the rank values are strictly ordered.
func TestImportanceRank_Ordering(t *testing.T) {
	if importanceRank(domain.ImportanceCritical) <= importanceRank(domain.ImportanceImportant) {
		t.Error("CRITICAL should rank higher than IMPORTANT")
	}
	if importanceRank(domain.ImportanceImportant) <= importanceRank(domain.ImportanceMinor) {
		t.Error("IMPORTANT should rank higher than MINOR")
	}
	if importanceRank(domain.ImportanceCritical) <= importanceRank(domain.ImportanceMinor) {
		t.Error("CRITICAL should rank higher than MINOR")
	}
}

// TestImportanceRank_UnknownLevels verifies that unknown levels return a rank lower than MINOR.
func TestImportanceRank_UnknownLevels(t *testing.T) {
	unknowns := []domain.ImportanceLevel{"SUPER", "MEGA", "none", ""}
	for _, level := range unknowns {
		rank := importanceRank(level)
		if rank >= importanceRank(domain.ImportanceMinor) {
			t.Errorf("unknown importance level %q should rank below MINOR, got %d", level, rank)
		}
	}
}
