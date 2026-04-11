package service

import (
	"context"
	"math"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
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

type fakeConsolidationMemoryRepo struct {
	memories []domain.Memory
	updated  []domain.Memory
	deleted  []string
}

func (f *fakeConsolidationMemoryRepo) FindAll(_ context.Context) ([]domain.Memory, error) {
	memories := make([]domain.Memory, len(f.memories))
	copy(memories, f.memories)
	return memories, nil
}

func (f *fakeConsolidationMemoryRepo) Update(_ context.Context, m *domain.Memory) error {
	copyMemory := *m
	f.updated = append(f.updated, copyMemory)
	return nil
}

func (f *fakeConsolidationMemoryRepo) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func TestConsolidateTenant_MergesSimilarMemories(t *testing.T) {
	repo := &fakeConsolidationMemoryRepo{
		memories: []domain.Memory{
			{
				ID:              "primary",
				Content:         "store memory metadata in jsonb",
				Summary:         "metadata in jsonb",
				Category:        domain.CategoryKnowledge,
				Importance:      domain.ImportanceMinor,
				Tags:            []string{"postgres"},
				Embedding:       []float32{1, 0, 0},
				Version:         2,
				AccessCount:     3,
				HelpfulCount:    1,
				NotHelpfulCount: 1,
			},
			{
				ID:              "secondary",
				Content:         "store metadata and tags in postgres jsonb",
				Summary:         "tags in jsonb",
				Category:        domain.CategoryKnowledge,
				Importance:      domain.ImportanceCritical,
				Tags:            []string{"tags"},
				Embedding:       []float32{1, 0, 0},
				AccessCount:     4,
				HelpfulCount:    2,
				NotHelpfulCount: 3,
			},
		},
	}
	llm := &mockLLMProvider{
		name:     "test",
		response: `{"content":"store memory metadata and tags in postgres jsonb","summary":"metadata and tags in jsonb"}`,
	}
	svc := NewConsolidationServiceWithLLM(repo, llm, nil, nil)
	ctx := tenant.WithTenant(context.Background(), "tenant-test")

	result, err := svc.ConsolidateTenant(ctx, 0.99)
	if err != nil {
		t.Fatalf("ConsolidateTenant() error = %v", err)
	}

	if result.Consolidated != 1 {
		t.Fatalf("expected one consolidation, got %d", result.Consolidated)
	}
	if len(result.MergedIDs) != 1 || result.MergedIDs[0] != "secondary" {
		t.Fatalf("merged IDs = %v", result.MergedIDs)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("expected one updated primary memory, got %d", len(repo.updated))
	}
	updated := repo.updated[0]
	if updated.ID != "primary" {
		t.Fatalf("updated ID = %s", updated.ID)
	}
	if updated.Content != "store memory metadata and tags in postgres jsonb" {
		t.Fatalf("merged content = %q", updated.Content)
	}
	if updated.Summary != "metadata and tags in jsonb" {
		t.Fatalf("merged summary = %q", updated.Summary)
	}
	if updated.Importance != domain.ImportanceCritical {
		t.Fatalf("expected higher importance to be preserved, got %s", updated.Importance)
	}
	if updated.AccessCount != 7 || updated.HelpfulCount != 3 || updated.NotHelpfulCount != 4 {
		t.Fatalf("expected counters to be summed, got access=%d helpful=%d notHelpful=%d",
			updated.AccessCount, updated.HelpfulCount, updated.NotHelpfulCount)
	}
	if updated.Version != 3 {
		t.Fatalf("expected version increment to 3, got %d", updated.Version)
	}
	assertHasTag(t, updated.Tags, "postgres")
	assertHasTag(t, updated.Tags, "tags")
	if len(repo.deleted) != 1 || repo.deleted[0] != "secondary" {
		t.Fatalf("deleted IDs = %v", repo.deleted)
	}
	if llm.callCount != 1 {
		t.Fatalf("expected one LLM call, got %d", llm.callCount)
	}
}

func TestConsolidateTenant_CompressesVerboseMemory(t *testing.T) {
	verbose := ""
	for i := 0; i < 250; i++ {
		verbose += "verbose memory detail "
	}
	repo := &fakeConsolidationMemoryRepo{
		memories: []domain.Memory{
			{
				ID:         "verbose-memory",
				Content:    verbose,
				Summary:    "verbose",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceMinor,
				Version:    1,
			},
		},
	}
	llm := &mockLLMProvider{
		name:     "test",
		response: `{"content":"compressed memory detail","summary":"compressed"}`,
	}
	svc := NewConsolidationServiceWithLLM(repo, llm, nil, nil)

	result, err := svc.ConsolidateTenant(context.Background(), 0.99)
	if err != nil {
		t.Fatalf("ConsolidateTenant() error = %v", err)
	}

	if result.Compressed != 1 {
		t.Fatalf("expected one compressed memory, got %d", result.Compressed)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("expected one updated compressed memory, got %d", len(repo.updated))
	}
	updated := repo.updated[0]
	if updated.Content != "compressed memory detail" {
		t.Fatalf("compressed content = %q", updated.Content)
	}
	if updated.Summary != "compressed" {
		t.Fatalf("compressed summary = %q", updated.Summary)
	}
	if updated.Version != 2 {
		t.Fatalf("expected version increment to 2, got %d", updated.Version)
	}
}

func assertHasTag(t *testing.T, tags []string, want string) {
	t.Helper()
	for _, tag := range tags {
		if tag == want {
			return
		}
	}
	t.Fatalf("expected tag %q in %v", want, tags)
}
