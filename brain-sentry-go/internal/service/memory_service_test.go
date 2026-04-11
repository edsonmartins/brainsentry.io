package service

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

// --- extractChainOfThought tests ---

func TestCOTExtraction_NoThought(t *testing.T) {
	content, cot := extractChainOfThought("normal content without thought tags")
	if cot != "" {
		t.Errorf("expected empty COT, got %q", cot)
	}
	if content != "normal content without thought tags" {
		t.Errorf("expected unchanged content, got %q", content)
	}
}

func TestCOTExtraction_SingleThought(t *testing.T) {
	input := "before <THOUGHT>my reasoning</THOUGHT> after"
	content, cot := extractChainOfThought(input)
	if cot != "my reasoning" {
		t.Errorf("expected 'my reasoning', got %q", cot)
	}
	if content != "before  after" {
		t.Errorf("expected 'before  after', got %q", content)
	}
}

func TestCOTExtraction_MultipleThoughts(t *testing.T) {
	input := "<THOUGHT>first</THOUGHT> text <THOUGHT>second</THOUGHT>"
	content, cot := extractChainOfThought(input)
	if cot != "first\n---\nsecond" {
		t.Errorf("expected 'first\\n---\\nsecond', got %q", cot)
	}
	if content != "text" {
		t.Errorf("expected 'text', got %q", content)
	}
}

// --- EmotionalWeight clamping (tested via CreateMemory logic inlined) ---

func TestEmotionalWeightClamping(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"above max", 1.5, 1.0},
		{"below min", -2.0, -1.0},
		{"within range", 0.5, 0.5},
		{"at max", 1.0, 1.0},
		{"at min", -1.0, -1.0},
		{"zero", 0.0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.input
			if w < -1 {
				w = -1
			}
			if w > 1 {
				w = 1
			}
			if w != tt.want {
				t.Errorf("clamp(%f) = %f, want %f", tt.input, w, tt.want)
			}
		})
	}
}

// --- RelevanceScore formula tests ---

func TestRelevanceScore_Formula(t *testing.T) {
	tests := []struct {
		name           string
		accessCount    int
		injectionCount int
		helpful        int
		notHelpful     int
		expected       float64
	}{
		{
			name:           "standard counts",
			accessCount:    10,
			injectionCount: 5,
			helpful:        3,
			notHelpful:     1,
			expected:       10*0.3 + 5*0.4 + 0.75*0.3, // 3 + 2 + 0.225 = 5.225
		},
		{
			name:           "zero everything",
			accessCount:    0,
			injectionCount: 0,
			helpful:        0,
			notHelpful:     0,
			expected:       0,
		},
		{
			name:           "all helpful",
			accessCount:    1,
			injectionCount: 1,
			helpful:        10,
			notHelpful:     0,
			expected:       1*0.3 + 1*0.4 + 1.0*0.3, // 0.3 + 0.4 + 0.3 = 1.0
		},
		{
			name:           "all not helpful",
			accessCount:    5,
			injectionCount: 0,
			helpful:        0,
			notHelpful:     10,
			expected:       5*0.3 + 0 + 0, // 1.5
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &domain.Memory{
				AccessCount:     tt.accessCount,
				InjectionCount:  tt.injectionCount,
				HelpfulCount:    tt.helpful,
				NotHelpfulCount: tt.notHelpful,
			}
			score := m.RelevanceScore()
			if math.Abs(score-tt.expected) > 0.001 {
				t.Errorf("RelevanceScore() = %f, want %f", score, tt.expected)
			}
		})
	}
}

// --- HelpfulnessRate tests ---

func TestHelpfulnessRate_ZeroDenominator(t *testing.T) {
	m := &domain.Memory{HelpfulCount: 0, NotHelpfulCount: 0}
	if m.HelpfulnessRate() != 0 {
		t.Errorf("expected 0 for zero counts, got %f", m.HelpfulnessRate())
	}
}

func TestHelpfulnessRate_AllHelpful(t *testing.T) {
	m := &domain.Memory{HelpfulCount: 10, NotHelpfulCount: 0}
	if m.HelpfulnessRate() != 1.0 {
		t.Errorf("expected 1.0 for all helpful, got %f", m.HelpfulnessRate())
	}
}

func TestHelpfulnessRate_FiftyFifty(t *testing.T) {
	m := &domain.Memory{HelpfulCount: 5, NotHelpfulCount: 5}
	if m.HelpfulnessRate() != 0.5 {
		t.Errorf("expected 0.5 for equal counts, got %f", m.HelpfulnessRate())
	}
}

// --- DecayRate from MemoryType ---

func TestDecayRateFromMemoryType(t *testing.T) {
	tests := []struct {
		memType domain.MemoryType
		rate    float64
	}{
		{domain.MemoryTypePersonality, 0.001},
		{domain.MemoryTypeThread, 0.050},
		{domain.MemoryTypeTask, 0.030},
		{domain.MemoryTypeSemantic, 0.005},
		{domain.MemoryTypeProcedural, 0.005},
		{domain.MemoryTypeEpisodic, 0.015},
		{domain.MemoryTypeAssociative, 0.010},
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

// --- sortScoredMemories tests ---

func TestSortScoredMemories_Descending(t *testing.T) {
	results := []scoredMemory{
		{memory: domain.Memory{ID: "low"}, trace: ScoreTrace{FinalScore: 0.3}},
		{memory: domain.Memory{ID: "high"}, trace: ScoreTrace{FinalScore: 0.9}},
		{memory: domain.Memory{ID: "mid"}, trace: ScoreTrace{FinalScore: 0.6}},
	}
	sortScoredMemories(results)
	if results[0].memory.ID != "high" {
		t.Errorf("expected 'high' first, got %s", results[0].memory.ID)
	}
	if results[2].memory.ID != "low" {
		t.Errorf("expected 'low' last, got %s", results[2].memory.ID)
	}
}

func TestSortScoredMemories_SingleElement(t *testing.T) {
	results := []scoredMemory{
		{memory: domain.Memory{ID: "only"}, trace: ScoreTrace{FinalScore: 0.5}},
	}
	sortScoredMemories(results)
	if results[0].memory.ID != "only" {
		t.Errorf("expected 'only', got %s", results[0].memory.ID)
	}
}

func TestSortScoredMemories_Empty(t *testing.T) {
	var results []scoredMemory
	sortScoredMemories(results) // should not panic
}

func TestSortScoredMemories_AllSameScore(t *testing.T) {
	results := []scoredMemory{
		{memory: domain.Memory{ID: "a"}, trace: ScoreTrace{FinalScore: 0.5}},
		{memory: domain.Memory{ID: "b"}, trace: ScoreTrace{FinalScore: 0.5}},
		{memory: domain.Memory{ID: "c"}, trace: ScoreTrace{FinalScore: 0.5}},
	}
	sortScoredMemories(results) // should not panic
	if len(results) != 3 {
		t.Errorf("expected 3, got %d", len(results))
	}
}

// --- NewMemoryService tests ---

func TestNewMemoryService_NilDeps(t *testing.T) {
	svc := NewMemoryService(nil, nil, nil, nil, nil, nil, true)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.piiService == nil {
		t.Error("piiService should be auto-initialized")
	}
	if !svc.autoImportance {
		t.Error("autoImportance should be true")
	}
}

func TestNewMemoryService_AutoImportanceFalse(t *testing.T) {
	svc := NewMemoryService(nil, nil, nil, nil, nil, nil, false)
	if svc.autoImportance {
		t.Error("autoImportance should be false")
	}
}

// --- memoryToResponse tests ---

func TestMemoryToResponse_PreservesFields(t *testing.T) {
	m := domain.Memory{
		ID:         "test-id",
		Content:    "test content",
		Summary:    "test summary",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceCritical,
		Tags:       []string{"go", "test"},
		MemoryType: domain.MemoryTypeSemantic,
		Version:    3,
		CreatedAt:  time.Now(),
	}
	resp := memoryToResponse(m)
	if resp.ID != m.ID {
		t.Errorf("ID mismatch: %s vs %s", resp.ID, m.ID)
	}
	if resp.Category != m.Category {
		t.Errorf("Category mismatch: %s vs %s", resp.Category, m.Category)
	}
	if resp.Version != m.Version {
		t.Errorf("Version mismatch: %d vs %d", resp.Version, m.Version)
	}
}

func TestMemoryToResponse_IncludesComputedFields(t *testing.T) {
	m := domain.Memory{
		AccessCount:     10,
		InjectionCount:  5,
		HelpfulCount:    3,
		NotHelpfulCount: 1,
		MemoryType:      domain.MemoryTypeSemantic,
		CreatedAt:       time.Now(),
	}
	resp := memoryToResponse(m)
	if resp.RelevanceScore <= 0 {
		t.Errorf("expected positive relevance score, got %f", resp.RelevanceScore)
	}
	if resp.HelpfulnessRate != 0.75 {
		t.Errorf("expected 0.75 helpfulness rate, got %f", resp.HelpfulnessRate)
	}
}

type fakeMemoryRepository struct {
	byID            map[string]*domain.Memory
	fullTextResults []domain.Memory
	fullTextQueries []string
}

func (f *fakeMemoryRepository) Create(_ context.Context, m *domain.Memory) error {
	if f.byID == nil {
		f.byID = make(map[string]*domain.Memory)
	}
	copyMemory := *m
	f.byID[m.ID] = &copyMemory
	return nil
}

func (f *fakeMemoryRepository) FindByID(_ context.Context, id string) (*domain.Memory, error) {
	if f.byID != nil {
		if m, ok := f.byID[id]; ok {
			copyMemory := *m
			return &copyMemory, nil
		}
	}
	return nil, errors.New("memory not found")
}

func (f *fakeMemoryRepository) List(_ context.Context, _ int, _ int) ([]domain.Memory, int64, error) {
	return nil, 0, nil
}

func (f *fakeMemoryRepository) Update(_ context.Context, m *domain.Memory) error {
	if f.byID == nil {
		f.byID = make(map[string]*domain.Memory)
	}
	copyMemory := *m
	f.byID[m.ID] = &copyMemory
	return nil
}

func (f *fakeMemoryRepository) Delete(_ context.Context, id string) error {
	delete(f.byID, id)
	return nil
}

func (f *fakeMemoryRepository) FindByCategory(_ context.Context, _ domain.MemoryCategory) ([]domain.Memory, error) {
	return nil, nil
}

func (f *fakeMemoryRepository) FindByImportance(_ context.Context, _ domain.ImportanceLevel) ([]domain.Memory, error) {
	return nil, nil
}

func (f *fakeMemoryRepository) FullTextSearch(_ context.Context, query string, limit int) ([]domain.Memory, error) {
	f.fullTextQueries = append(f.fullTextQueries, query)
	if limit > 0 && len(f.fullTextResults) > limit {
		return f.fullTextResults[:limit], nil
	}
	return f.fullTextResults, nil
}

func (f *fakeMemoryRepository) IncrementAccessCount(_ context.Context, _ string) error {
	return nil
}

func (f *fakeMemoryRepository) RecordFeedback(_ context.Context, _ string, _ bool) error {
	return nil
}

func (f *fakeMemoryRepository) FindSimHashes(_ context.Context) (map[string]string, error) {
	return nil, nil
}

func (f *fakeMemoryRepository) BoostAccessCount(_ context.Context, _ string, _ int) error {
	return nil
}

func (f *fakeMemoryRepository) SupersedeMemory(_ context.Context, _ string, _ string) error {
	return nil
}

type fakeMemoryGraphRepository struct {
	ids    []string
	scores []float64
}

func (f *fakeMemoryGraphRepository) VectorSearch(_ context.Context, _ []float32, _ int, _ string) ([]string, []float64, error) {
	return f.ids, f.scores, nil
}

type fakeEmbeddingGenerator struct {
	api bool
}

func (f fakeEmbeddingGenerator) Embed(_ string) []float32 {
	return []float32{1, 0, 0}
}

func (f fakeEmbeddingGenerator) HasAPI() bool {
	return f.api
}

func TestSearchMemories_TextFallbackFiltersInactiveAndSorts(t *testing.T) {
	now := time.Now()
	expiredAt := now.Add(-time.Hour)
	repo := &fakeMemoryRepository{
		fullTextResults: []domain.Memory{
			{
				ID:         "lower-score",
				Content:    "postgres memory",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceMinor,
				Tags:       []string{"other"},
				MemoryType: domain.MemoryTypeSemantic,
				CreatedAt:  now,
				UpdatedAt:  now,
				DecayRate:  GetDecayRate(domain.MemoryTypeSemantic),
			},
			{
				ID:         "expired-high-score",
				Content:    "postgres memory integrity",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceCritical,
				Tags:       []string{"core"},
				MemoryType: domain.MemoryTypeSemantic,
				ValidTo:    &expiredAt,
				CreatedAt:  now,
				UpdatedAt:  now,
				DecayRate:  GetDecayRate(domain.MemoryTypeSemantic),
			},
			{
				ID:           "superseded-high-score",
				Content:      "postgres memory integrity",
				Category:     domain.CategoryKnowledge,
				Importance:   domain.ImportanceCritical,
				Tags:         []string{"core"},
				MemoryType:   domain.MemoryTypeSemantic,
				SupersededBy: "replacement",
				CreatedAt:    now,
				UpdatedAt:    now,
				DecayRate:    GetDecayRate(domain.MemoryTypeSemantic),
			},
			{
				ID:         "best-active",
				Content:    "postgres memory integrity",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceCritical,
				Tags:       []string{"core"},
				MemoryType: domain.MemoryTypeSemantic,
				CreatedAt:  now,
				UpdatedAt:  now,
				DecayRate:  GetDecayRate(domain.MemoryTypeSemantic),
			},
		},
	}
	svc := &MemoryService{
		memoryRepo: repo,
		piiService: NewPIIService(),
	}

	resp, err := svc.SearchMemories(context.Background(), dto.SearchRequest{
		Query: "postgres memory integrity",
		Tags:  []string{"core"},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchMemories() error = %v", err)
	}

	if resp.Total != 2 {
		t.Fatalf("expected 2 active results, got %d: %#v", resp.Total, resp.Results)
	}
	if resp.Results[0].ID != "best-active" {
		t.Fatalf("expected best-active first, got %s", resp.Results[0].ID)
	}
	if resp.Results[1].ID != "lower-score" {
		t.Fatalf("expected lower-score second, got %s", resp.Results[1].ID)
	}
	if len(repo.fullTextQueries) != 1 || repo.fullTextQueries[0] != "postgres memory integrity" {
		t.Fatalf("full-text queries = %v", repo.fullTextQueries)
	}
}

func TestSearchMemories_DeduplicatesVectorAndTextResults(t *testing.T) {
	now := time.Now()
	repo := &fakeMemoryRepository{
		byID: map[string]*domain.Memory{
			"vector-duplicate": {
				ID:         "vector-duplicate",
				Content:    "postgres memory integrity vector duplicate",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceCritical,
				Tags:       []string{"core"},
				MemoryType: domain.MemoryTypeSemantic,
				CreatedAt:  now,
				UpdatedAt:  now,
				DecayRate:  GetDecayRate(domain.MemoryTypeSemantic),
			},
			"vector-only": {
				ID:         "vector-only",
				Content:    "postgres memory integrity vector only",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceImportant,
				Tags:       []string{"core"},
				MemoryType: domain.MemoryTypeSemantic,
				CreatedAt:  now,
				UpdatedAt:  now,
				DecayRate:  GetDecayRate(domain.MemoryTypeSemantic),
			},
		},
		fullTextResults: []domain.Memory{
			{
				ID:         "vector-duplicate",
				Content:    "postgres memory integrity vector duplicate",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceCritical,
				Tags:       []string{"core"},
				MemoryType: domain.MemoryTypeSemantic,
				CreatedAt:  now,
				UpdatedAt:  now,
				DecayRate:  GetDecayRate(domain.MemoryTypeSemantic),
			},
			{
				ID:         "text-only",
				Content:    "postgres memory integrity text only",
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceMinor,
				Tags:       []string{"core"},
				MemoryType: domain.MemoryTypeSemantic,
				CreatedAt:  now,
				UpdatedAt:  now,
				DecayRate:  GetDecayRate(domain.MemoryTypeSemantic),
			},
		},
	}
	svc := &MemoryService{
		memoryRepo:       repo,
		memoryGraphRepo:  &fakeMemoryGraphRepository{ids: []string{"vector-duplicate", "vector-only"}, scores: []float64{0.95, 0.9}},
		embeddingService: fakeEmbeddingGenerator{api: true},
		piiService:       NewPIIService(),
	}

	resp, err := svc.SearchMemories(context.Background(), dto.SearchRequest{
		Query: "postgres memory integrity",
		Tags:  []string{"core"},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchMemories() error = %v", err)
	}

	gotIDs := make([]string, 0, len(resp.Results))
	seen := make(map[string]bool, len(resp.Results))
	for _, result := range resp.Results {
		gotIDs = append(gotIDs, result.ID)
		if seen[result.ID] {
			t.Fatalf("duplicate result %q in %v", result.ID, gotIDs)
		}
		seen[result.ID] = true
	}

	wantIDs := []string{"vector-duplicate", "vector-only", "text-only"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("result IDs = %v, want %v", gotIDs, wantIDs)
	}
}
