package service

import (
	"testing"
	"time"
)

func TestNewBenchmarkService(t *testing.T) {
	svc := NewBenchmarkService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.defaultK != 10 {
		t.Errorf("expected default K=10, got %d", svc.defaultK)
	}
}

func TestRunBenchmark_PerfectRecall(t *testing.T) {
	svc := NewBenchmarkService()
	dataset := &BenchmarkDataset{
		Name: "test",
		Queries: []BenchmarkQuery{
			{
				ID:          "q1",
				Query:       "test query",
				ExpectedIDs: []string{"m1", "m2", "m3"},
			},
		},
	}

	// Perfect search: returns all expected
	searchFn := func(query string, limit int) ([]string, time.Duration, error) {
		return []string{"m1", "m2", "m3"}, 5 * time.Millisecond, nil
	}

	report, err := svc.RunBenchmark(dataset, searchFn, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Metrics[MetricRecall] != 1.0 {
		t.Errorf("expected recall 1.0, got %f", report.Metrics[MetricRecall])
	}
	if report.Metrics[MetricPrecision] != 1.0 {
		t.Errorf("expected precision 1.0, got %f", report.Metrics[MetricPrecision])
	}
	if report.Metrics[MetricF1] != 1.0 {
		t.Errorf("expected F1 1.0, got %f", report.Metrics[MetricF1])
	}
	if report.Metrics[MetricMRR] != 1.0 {
		t.Errorf("expected MRR 1.0, got %f", report.Metrics[MetricMRR])
	}
}

func TestRunBenchmark_PartialRecall(t *testing.T) {
	svc := NewBenchmarkService()
	dataset := &BenchmarkDataset{
		Name: "test",
		Queries: []BenchmarkQuery{
			{
				ID:          "q1",
				Query:       "test",
				ExpectedIDs: []string{"m1", "m2", "m3", "m4"},
			},
		},
	}

	// Returns 2 of 4 expected + 1 extra
	searchFn := func(query string, limit int) ([]string, time.Duration, error) {
		return []string{"m1", "m2", "extra"}, 3 * time.Millisecond, nil
	}

	report, err := svc.RunBenchmark(dataset, searchFn, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Recall: 2/4 = 0.5
	if report.Metrics[MetricRecall] != 0.5 {
		t.Errorf("expected recall 0.5, got %f", report.Metrics[MetricRecall])
	}
	// Precision: 2/3 ≈ 0.667
	expectedPrec := 2.0 / 3.0
	if report.Metrics[MetricPrecision] < expectedPrec-0.01 || report.Metrics[MetricPrecision] > expectedPrec+0.01 {
		t.Errorf("expected precision ~0.667, got %f", report.Metrics[MetricPrecision])
	}
}

func TestRunBenchmark_EmptyResults(t *testing.T) {
	svc := NewBenchmarkService()
	dataset := &BenchmarkDataset{
		Name: "test",
		Queries: []BenchmarkQuery{
			{
				ID:          "q1",
				Query:       "test",
				ExpectedIDs: []string{"m1"},
			},
		},
	}

	searchFn := func(query string, limit int) ([]string, time.Duration, error) {
		return nil, 1 * time.Millisecond, nil
	}

	report, err := svc.RunBenchmark(dataset, searchFn, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Metrics[MetricRecall] != 0 {
		t.Errorf("expected recall 0, got %f", report.Metrics[MetricRecall])
	}
}

func TestMRR(t *testing.T) {
	svc := NewBenchmarkService()

	q := BenchmarkQuery{
		ID:          "q1",
		Query:       "test",
		ExpectedIDs: []string{"m3"},
	}

	// Relevant item at position 3 (0-indexed: 2) → MRR = 1/3
	result := svc.evaluateQuery(q, []string{"m1", "m2", "m3", "m4"}, 0)
	expectedMRR := 1.0 / 3.0
	if result.MRR < expectedMRR-0.01 || result.MRR > expectedMRR+0.01 {
		t.Errorf("expected MRR ~0.333, got %f", result.MRR)
	}
}

func TestNDCG_BinaryRelevance(t *testing.T) {
	q := BenchmarkQuery{
		ID:          "q1",
		ExpectedIDs: []string{"m1", "m2"},
	}

	// Perfect ranking
	ndcg := computeNDCG(q, []string{"m1", "m2", "m3"})
	if ndcg != 1.0 {
		t.Errorf("expected NDCG 1.0 for perfect ranking, got %f", ndcg)
	}

	// Imperfect ranking (relevant at positions 2,3 instead of 1,2)
	ndcg2 := computeNDCG(q, []string{"m3", "m1", "m2"})
	if ndcg2 >= 1.0 {
		t.Errorf("expected NDCG < 1.0 for imperfect ranking, got %f", ndcg2)
	}
	if ndcg2 <= 0 {
		t.Errorf("expected NDCG > 0, got %f", ndcg2)
	}
}

func TestNDCG_GradedRelevance(t *testing.T) {
	q := BenchmarkQuery{
		ID:          "q1",
		ExpectedIDs: []string{"m1", "m2"},
		RelevanceScores: map[string]float64{
			"m1": 3.0, // highly relevant
			"m2": 1.0, // somewhat relevant
		},
	}

	// Best ranking: most relevant first
	bestNDCG := computeNDCG(q, []string{"m1", "m2"})
	// Worst ranking: least relevant first
	worstNDCG := computeNDCG(q, []string{"m2", "m1"})

	if bestNDCG < worstNDCG {
		t.Errorf("best ranking (%.4f) should have higher NDCG than worst (%.4f)", bestNDCG, worstNDCG)
	}
}

func TestLatencyStats(t *testing.T) {
	latencies := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
		100 * time.Millisecond, // outlier
	}

	stats := computeLatencyStats(latencies)
	if stats.Min != 1*time.Millisecond {
		t.Errorf("expected min 1ms, got %s", stats.Min)
	}
	if stats.Max != 100*time.Millisecond {
		t.Errorf("expected max 100ms, got %s", stats.Max)
	}
	if stats.P50 > 10*time.Millisecond {
		t.Errorf("expected P50 < 10ms, got %s", stats.P50)
	}
	if stats.P95 < 5*time.Millisecond {
		t.Errorf("expected P95 >= 5ms, got %s", stats.P95)
	}
}

func TestLatencyStats_Empty(t *testing.T) {
	stats := computeLatencyStats(nil)
	if stats.Min != 0 || stats.Max != 0 {
		t.Error("expected zero stats for empty input")
	}
}

func TestGenerateSyntheticDataset(t *testing.T) {
	dataset := GenerateSyntheticDataset("test", 20, 3, []string{"factual", "temporal", "preference"})
	if dataset.Name != "test" {
		t.Errorf("expected name 'test', got %s", dataset.Name)
	}
	if len(dataset.Queries) != 20 {
		t.Errorf("expected 20 queries, got %d", len(dataset.Queries))
	}
	if len(dataset.Queries[0].ExpectedIDs) != 3 {
		t.Errorf("expected 3 expected IDs, got %d", len(dataset.Queries[0].ExpectedIDs))
	}
	// Check categories are distributed
	cats := make(map[string]int)
	for _, q := range dataset.Queries {
		cats[q.Category]++
	}
	if len(cats) != 3 {
		t.Errorf("expected 3 categories, got %d", len(cats))
	}
}

func TestRunBenchmark_WithCategories(t *testing.T) {
	svc := NewBenchmarkService()
	dataset := &BenchmarkDataset{
		Name: "categorized",
		Queries: []BenchmarkQuery{
			{ID: "q1", Query: "a", ExpectedIDs: []string{"m1"}, Category: "factual"},
			{ID: "q2", Query: "b", ExpectedIDs: []string{"m2"}, Category: "factual"},
			{ID: "q3", Query: "c", ExpectedIDs: []string{"m3"}, Category: "temporal"},
		},
	}

	searchFn := func(query string, limit int) ([]string, time.Duration, error) {
		switch query {
		case "a":
			return []string{"m1"}, 1 * time.Millisecond, nil
		case "b":
			return []string{"x"}, 1 * time.Millisecond, nil // miss
		default:
			return []string{"m3"}, 1 * time.Millisecond, nil
		}
	}

	report, err := svc.RunBenchmark(dataset, searchFn, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check per-category
	if len(report.PerCategory) != 2 {
		t.Errorf("expected 2 categories, got %d", len(report.PerCategory))
	}

	factual := report.PerCategory["factual"]
	if factual[MetricRecall] != 0.5 {
		t.Errorf("expected factual recall 0.5, got %f", factual[MetricRecall])
	}

	temporal := report.PerCategory["temporal"]
	if temporal[MetricRecall] != 1.0 {
		t.Errorf("expected temporal recall 1.0, got %f", temporal[MetricRecall])
	}
}

func TestFormatReport(t *testing.T) {
	report := &BenchmarkReport{
		DatasetName:  "test",
		TotalQueries: 10,
		Metrics: map[BenchmarkMetricType]float64{
			MetricRecall:     0.85,
			MetricPrecision:  0.90,
			MetricF1:         0.87,
			MetricMRR:        0.75,
			MetricNDCG:       0.82,
			MetricThroughput: 100.0,
		},
		Latencies: LatencyStats{
			P50: 5 * time.Millisecond,
			P95: 20 * time.Millisecond,
			P99: 50 * time.Millisecond,
		},
		Duration: 100 * time.Millisecond,
	}

	text := FormatReport(report)
	if text == "" {
		t.Fatal("expected non-empty report")
	}
	if !containsSubstr(text, "test") {
		t.Error("expected dataset name in report")
	}
	if !containsSubstr(text, "0.8500") {
		t.Error("expected recall in report")
	}
}

func TestBenchmarkMetricTypes(t *testing.T) {
	types := []BenchmarkMetricType{
		MetricRecall, MetricPrecision, MetricF1, MetricMRR, MetricNDCG,
		MetricLatencyP50, MetricLatencyP95, MetricLatencyP99, MetricThroughput,
	}
	seen := make(map[BenchmarkMetricType]bool)
	for _, mt := range types {
		if seen[mt] {
			t.Errorf("duplicate metric type: %s", mt)
		}
		seen[mt] = true
	}
	if len(types) != 9 {
		t.Errorf("expected 9 metric types, got %d", len(types))
	}
}
