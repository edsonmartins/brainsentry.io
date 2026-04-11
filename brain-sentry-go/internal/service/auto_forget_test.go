package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestJaccardSimilarity_Identical(t *testing.T) {
	a := map[string]bool{"hello": true, "world": true}
	b := map[string]bool{"hello": true, "world": true}

	sim := jaccardSimilarity(a, b)
	if sim != 1.0 {
		t.Errorf("expected 1.0, got %f", sim)
	}
}

func TestJaccardSimilarity_Disjoint(t *testing.T) {
	a := map[string]bool{"hello": true}
	b := map[string]bool{"world": true}

	sim := jaccardSimilarity(a, b)
	if sim != 0.0 {
		t.Errorf("expected 0.0, got %f", sim)
	}
}

func TestJaccardSimilarity_Partial(t *testing.T) {
	a := map[string]bool{"hello": true, "world": true, "foo": true}
	b := map[string]bool{"hello": true, "world": true, "bar": true}

	sim := jaccardSimilarity(a, b)
	// intersection=2, union=4 → 0.5
	if sim != 0.5 {
		t.Errorf("expected 0.5, got %f", sim)
	}
}

func TestJaccardSimilarity_Empty(t *testing.T) {
	a := map[string]bool{}
	b := map[string]bool{}

	sim := jaccardSimilarity(a, b)
	if sim != 1.0 {
		t.Errorf("expected 1.0 for empty sets, got %f", sim)
	}
}

func TestTokenizeForJaccard(t *testing.T) {
	tokens := tokenizeForJaccard("Hello World! This is a test.")
	if !tokens["hello"] {
		t.Error("expected 'hello' token")
	}
	if !tokens["world"] {
		t.Error("expected 'world' token")
	}
	if !tokens["test"] {
		t.Error("expected 'test' token")
	}
	// Single char words should be excluded
	if tokens["a"] {
		t.Error("single char 'a' should be excluded")
	}
}

func TestDefaultAutoForgetConfig(t *testing.T) {
	cfg := DefaultAutoForgetConfig()
	if !cfg.TTLEnabled {
		t.Error("TTL should be enabled by default")
	}
	if !cfg.ContradictionEnabled {
		t.Error("Contradiction detection should be enabled by default")
	}
	if cfg.ContradictionThreshold != 0.9 {
		t.Errorf("expected threshold 0.9, got %f", cfg.ContradictionThreshold)
	}
	if cfg.MaxDeletesPerRun != 50 {
		t.Errorf("expected max 50 deletes, got %d", cfg.MaxDeletesPerRun)
	}
}

type fakeAutoForgetMemoryRepo struct {
	memories []domain.Memory

	listPages     []int
	deletedIDs    []string
	supersededIDs [][2]string
}

func (f *fakeAutoForgetMemoryRepo) List(_ context.Context, page, _ int) ([]domain.Memory, int64, error) {
	f.listPages = append(f.listPages, page)

	memories := make([]domain.Memory, len(f.memories))
	copy(memories, f.memories)
	return memories, int64(len(memories)), nil
}

func (f *fakeAutoForgetMemoryRepo) Delete(_ context.Context, id string) error {
	f.deletedIDs = append(f.deletedIDs, id)
	return nil
}

func (f *fakeAutoForgetMemoryRepo) SupersedeMemory(_ context.Context, oldID, newID string) error {
	f.supersededIDs = append(f.supersededIDs, [2]string{oldID, newID})
	return nil
}

func TestAutoForgetRun_DryRunDoesNotDeleteExpiredMemory(t *testing.T) {
	now := time.Now()
	expired := now.Add(-time.Hour)
	repo := &fakeAutoForgetMemoryRepo{
		memories: []domain.Memory{
			{
				ID:      "expired-memory",
				ValidTo: &expired,
			},
		},
	}
	svc := &AutoForgetService{
		memoryRepo: repo,
		config: AutoForgetConfig{
			TTLEnabled:             true,
			ContradictionEnabled:   false,
			LowValueEnabled:        false,
			ContradictionThreshold: 0.9,
			MaxDeletesPerRun:       10,
		},
	}

	result, err := svc.Run(context.Background(), true)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.TTLExpired != 1 || result.TotalDeleted != 1 {
		t.Fatalf("expected dry-run to report one expired memory, got ttl=%d total=%d", result.TTLExpired, result.TotalDeleted)
	}
	if len(repo.deletedIDs) != 0 {
		t.Fatalf("dry-run deleted memories: %v", repo.deletedIDs)
	}
	if !reflect.DeepEqual(repo.listPages, []int{0}) {
		t.Fatalf("expected auto-forget to scan first page, got pages %v", repo.listPages)
	}
}

func TestAutoForgetRun_DeletesExpiredTTLMemory(t *testing.T) {
	now := time.Now()
	expired := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	repo := &fakeAutoForgetMemoryRepo{
		memories: []domain.Memory{
			{
				ID:      "expired-memory",
				ValidTo: &expired,
			},
			{
				ID:      "active-memory",
				ValidTo: &future,
			},
		},
	}
	svc := &AutoForgetService{
		memoryRepo: repo,
		config: AutoForgetConfig{
			TTLEnabled:             true,
			ContradictionEnabled:   false,
			LowValueEnabled:        false,
			ContradictionThreshold: 0.9,
			MaxDeletesPerRun:       10,
		},
	}

	result, err := svc.Run(context.Background(), false)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.TTLExpired != 1 || result.TotalDeleted != 1 {
		t.Fatalf("expected one deleted expired memory, got ttl=%d total=%d", result.TTLExpired, result.TotalDeleted)
	}
	if !reflect.DeepEqual(repo.deletedIDs, []string{"expired-memory"}) {
		t.Fatalf("deleted IDs = %v", repo.deletedIDs)
	}
}

func TestAutoForgetRun_SupersedesOlderDuplicateMemory(t *testing.T) {
	now := time.Now()
	repo := &fakeAutoForgetMemoryRepo{
		memories: []domain.Memory{
			{
				ID:        "older-memory",
				Content:   "use postgres jsonb metadata for memory storage",
				Category:  domain.CategoryKnowledge,
				CreatedAt: now.Add(-time.Hour),
			},
			{
				ID:        "newer-memory",
				Content:   "use postgres jsonb metadata for memory storage",
				Category:  domain.CategoryKnowledge,
				CreatedAt: now,
			},
		},
	}
	svc := &AutoForgetService{
		memoryRepo: repo,
		config: AutoForgetConfig{
			TTLEnabled:             false,
			ContradictionEnabled:   true,
			LowValueEnabled:        false,
			ContradictionThreshold: 1.0,
			MaxDeletesPerRun:       10,
		},
	}

	result, err := svc.Run(context.Background(), false)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Contradictions != 1 || result.TotalDeleted != 1 {
		t.Fatalf("expected one superseded duplicate, got contradictions=%d total=%d", result.Contradictions, result.TotalDeleted)
	}
	if !reflect.DeepEqual(repo.supersededIDs, [][2]string{{"older-memory", "newer-memory"}}) {
		t.Fatalf("superseded IDs = %v", repo.supersededIDs)
	}
}

func TestAutoForgetRun_DeletesLowValueMemory(t *testing.T) {
	now := time.Now()
	recentAccess := now.Add(-time.Hour)
	repo := &fakeAutoForgetMemoryRepo{
		memories: []domain.Memory{
			{
				ID:          "low-value-memory",
				Importance:  domain.ImportanceMinor,
				CreatedAt:   now.AddDate(0, 0, -200),
				AccessCount: 0,
			},
			{
				ID:             "recently-accessed-memory",
				Importance:     domain.ImportanceMinor,
				CreatedAt:      now.AddDate(0, 0, -200),
				LastAccessedAt: &recentAccess,
				AccessCount:    0,
			},
			{
				ID:          "important-memory",
				Importance:  domain.ImportanceImportant,
				CreatedAt:   now.AddDate(0, 0, -200),
				AccessCount: 0,
			},
		},
	}
	svc := &AutoForgetService{
		memoryRepo: repo,
		config: AutoForgetConfig{
			TTLEnabled:            false,
			ContradictionEnabled:  false,
			LowValueEnabled:       true,
			LowValueMaxAgeDays:    180,
			LowValueMaxImportance: string(domain.ImportanceMinor),
			MaxDeletesPerRun:      10,
		},
	}

	result, err := svc.Run(context.Background(), false)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.LowValue != 1 || result.TotalDeleted != 1 {
		t.Fatalf("expected one low-value memory deleted, got low_value=%d total=%d", result.LowValue, result.TotalDeleted)
	}
	if !reflect.DeepEqual(repo.deletedIDs, []string{"low-value-memory"}) {
		t.Fatalf("deleted IDs = %v", repo.deletedIDs)
	}
}

func TestAutoForgetRun_RespectsMaxDeletesPerRun(t *testing.T) {
	now := time.Now()
	expired := now.Add(-time.Hour)
	repo := &fakeAutoForgetMemoryRepo{
		memories: []domain.Memory{
			{ID: "expired-memory-1", ValidTo: &expired},
			{ID: "expired-memory-2", ValidTo: &expired},
		},
	}
	svc := &AutoForgetService{
		memoryRepo: repo,
		config: AutoForgetConfig{
			TTLEnabled:             true,
			ContradictionEnabled:   false,
			LowValueEnabled:        false,
			ContradictionThreshold: 0.9,
			MaxDeletesPerRun:       1,
		},
	}

	result, err := svc.Run(context.Background(), false)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.TotalDeleted != 1 {
		t.Fatalf("expected max one deletion, got %d", result.TotalDeleted)
	}
	if !reflect.DeepEqual(repo.deletedIDs, []string{"expired-memory-1"}) {
		t.Fatalf("deleted IDs = %v", repo.deletedIDs)
	}
}
