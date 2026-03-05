package service

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// BenchmarkMetricType identifies what is being measured.
type BenchmarkMetricType string

const (
	MetricRecall    BenchmarkMetricType = "recall"
	MetricPrecision BenchmarkMetricType = "precision"
	MetricF1        BenchmarkMetricType = "f1"
	MetricMRR       BenchmarkMetricType = "mrr"       // Mean Reciprocal Rank
	MetricNDCG      BenchmarkMetricType = "ndcg"      // Normalized Discounted Cumulative Gain
	MetricLatencyP50 BenchmarkMetricType = "latency_p50"
	MetricLatencyP95 BenchmarkMetricType = "latency_p95"
	MetricLatencyP99 BenchmarkMetricType = "latency_p99"
	MetricThroughput BenchmarkMetricType = "throughput" // ops/sec
)

// BenchmarkQuery represents a test query with expected results.
type BenchmarkQuery struct {
	ID              string   `json:"id"`
	Query           string   `json:"query"`
	ExpectedIDs     []string `json:"expectedIds"`     // ground truth memory IDs
	RelevanceScores map[string]float64 `json:"relevanceScores,omitempty"` // per-ID relevance (for NDCG)
	Category        string   `json:"category,omitempty"`
}

// BenchmarkDataset holds a collection of test queries.
type BenchmarkDataset struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Queries     []BenchmarkQuery `json:"queries"`
	CreatedAt   time.Time        `json:"createdAt"`
}

// BenchmarkResult holds the results of a single query evaluation.
type BenchmarkResult struct {
	QueryID      string        `json:"queryId"`
	RetrievedIDs []string      `json:"retrievedIds"`
	Recall       float64       `json:"recall"`
	Precision    float64       `json:"precision"`
	F1           float64       `json:"f1"`
	MRR          float64       `json:"mrr"`
	NDCG         float64       `json:"ndcg"`
	Latency      time.Duration `json:"latency"`
}

// BenchmarkReport holds aggregated results across all queries.
type BenchmarkReport struct {
	DatasetName     string              `json:"datasetName"`
	TotalQueries    int                 `json:"totalQueries"`
	Metrics         map[BenchmarkMetricType]float64 `json:"metrics"`
	PerQuery        []BenchmarkResult   `json:"perQuery"`
	PerCategory     map[string]map[BenchmarkMetricType]float64 `json:"perCategory,omitempty"`
	Latencies       LatencyStats        `json:"latencies"`
	RunAt           time.Time           `json:"runAt"`
	Duration        time.Duration       `json:"duration"`
}

// LatencyStats holds latency distribution statistics.
type LatencyStats struct {
	Min    time.Duration `json:"min"`
	Max    time.Duration `json:"max"`
	Mean   time.Duration `json:"mean"`
	P50    time.Duration `json:"p50"`
	P95    time.Duration `json:"p95"`
	P99    time.Duration `json:"p99"`
}

// SearchFunc is the function signature for memory search to benchmark.
type SearchFunc func(query string, limit int) ([]string, time.Duration, error)

// BenchmarkService runs benchmarks against the memory system.
type BenchmarkService struct {
	defaultK int // default top-K for retrieval
}

// NewBenchmarkService creates a new BenchmarkService.
func NewBenchmarkService() *BenchmarkService {
	return &BenchmarkService{
		defaultK: 10,
	}
}

// RunBenchmark evaluates a search function against a dataset.
func (s *BenchmarkService) RunBenchmark(dataset *BenchmarkDataset, searchFn SearchFunc, k int) (*BenchmarkReport, error) {
	if k <= 0 {
		k = s.defaultK
	}

	start := time.Now()
	report := &BenchmarkReport{
		DatasetName:  dataset.Name,
		TotalQueries: len(dataset.Queries),
		Metrics:      make(map[BenchmarkMetricType]float64),
		PerCategory:  make(map[string]map[BenchmarkMetricType]float64),
		RunAt:        start,
	}

	var (
		totalRecall    float64
		totalPrecision float64
		totalF1        float64
		totalMRR       float64
		totalNDCG      float64
		latencies      []time.Duration
	)

	// Category accumulators
	catMetrics := make(map[string]*categoryAccumulator)

	for _, q := range dataset.Queries {
		retrievedIDs, latency, err := searchFn(q.Query, k)
		if err != nil {
			return nil, fmt.Errorf("search failed for query %s: %w", q.ID, err)
		}

		result := s.evaluateQuery(q, retrievedIDs, latency)
		report.PerQuery = append(report.PerQuery, result)

		totalRecall += result.Recall
		totalPrecision += result.Precision
		totalF1 += result.F1
		totalMRR += result.MRR
		totalNDCG += result.NDCG
		latencies = append(latencies, latency)

		// Per-category
		if q.Category != "" {
			acc, ok := catMetrics[q.Category]
			if !ok {
				acc = &categoryAccumulator{}
				catMetrics[q.Category] = acc
			}
			acc.recall += result.Recall
			acc.precision += result.Precision
			acc.f1 += result.F1
			acc.mrr += result.MRR
			acc.ndcg += result.NDCG
			acc.count++
		}
	}

	n := float64(len(dataset.Queries))
	if n > 0 {
		report.Metrics[MetricRecall] = totalRecall / n
		report.Metrics[MetricPrecision] = totalPrecision / n
		report.Metrics[MetricF1] = totalF1 / n
		report.Metrics[MetricMRR] = totalMRR / n
		report.Metrics[MetricNDCG] = totalNDCG / n
	}

	// Latency stats
	report.Latencies = computeLatencyStats(latencies)
	report.Metrics[MetricLatencyP50] = float64(report.Latencies.P50.Milliseconds())
	report.Metrics[MetricLatencyP95] = float64(report.Latencies.P95.Milliseconds())
	report.Metrics[MetricLatencyP99] = float64(report.Latencies.P99.Milliseconds())

	duration := time.Since(start)
	report.Duration = duration
	if duration > 0 {
		report.Metrics[MetricThroughput] = n / duration.Seconds()
	}

	// Per-category aggregation
	for cat, acc := range catMetrics {
		cn := float64(acc.count)
		report.PerCategory[cat] = map[BenchmarkMetricType]float64{
			MetricRecall:    acc.recall / cn,
			MetricPrecision: acc.precision / cn,
			MetricF1:        acc.f1 / cn,
			MetricMRR:       acc.mrr / cn,
			MetricNDCG:      acc.ndcg / cn,
		}
	}

	return report, nil
}

// evaluateQuery computes metrics for a single query.
func (s *BenchmarkService) evaluateQuery(q BenchmarkQuery, retrievedIDs []string, latency time.Duration) BenchmarkResult {
	result := BenchmarkResult{
		QueryID:      q.ID,
		RetrievedIDs: retrievedIDs,
		Latency:      latency,
	}

	expectedSet := make(map[string]bool, len(q.ExpectedIDs))
	for _, id := range q.ExpectedIDs {
		expectedSet[id] = true
	}

	if len(q.ExpectedIDs) == 0 {
		return result
	}

	// Recall & Precision
	hits := 0
	for _, id := range retrievedIDs {
		if expectedSet[id] {
			hits++
		}
	}

	result.Recall = float64(hits) / float64(len(q.ExpectedIDs))
	if len(retrievedIDs) > 0 {
		result.Precision = float64(hits) / float64(len(retrievedIDs))
	}

	// F1
	if result.Recall+result.Precision > 0 {
		result.F1 = 2 * (result.Recall * result.Precision) / (result.Recall + result.Precision)
	}

	// MRR (first relevant result position)
	for i, id := range retrievedIDs {
		if expectedSet[id] {
			result.MRR = 1.0 / float64(i+1)
			break
		}
	}

	// NDCG
	result.NDCG = computeNDCG(q, retrievedIDs)

	return result
}

// computeNDCG calculates Normalized Discounted Cumulative Gain.
func computeNDCG(q BenchmarkQuery, retrievedIDs []string) float64 {
	if len(q.ExpectedIDs) == 0 || len(retrievedIDs) == 0 {
		return 0
	}

	// Build relevance map
	relevance := make(map[string]float64)
	if len(q.RelevanceScores) > 0 {
		relevance = q.RelevanceScores
	} else {
		// Binary relevance: 1 if expected, 0 otherwise
		for _, id := range q.ExpectedIDs {
			relevance[id] = 1.0
		}
	}

	// DCG
	dcg := 0.0
	for i, id := range retrievedIDs {
		rel := relevance[id]
		dcg += (math.Pow(2, rel) - 1) / math.Log2(float64(i+2)) // log2(i+2) since i is 0-indexed
	}

	// Ideal DCG (sort by relevance descending)
	idealRels := make([]float64, 0, len(relevance))
	for _, rel := range relevance {
		idealRels = append(idealRels, rel)
	}
	sort.Float64s(idealRels)
	// Reverse
	for i, j := 0, len(idealRels)-1; i < j; i, j = i+1, j-1 {
		idealRels[i], idealRels[j] = idealRels[j], idealRels[i]
	}

	idcg := 0.0
	for i, rel := range idealRels {
		if i >= len(retrievedIDs) {
			break
		}
		idcg += (math.Pow(2, rel) - 1) / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0
	}

	return dcg / idcg
}

// computeLatencyStats calculates latency distribution statistics.
func computeLatencyStats(latencies []time.Duration) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	// Sort latencies
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var total time.Duration
	for _, l := range sorted {
		total += l
	}

	return LatencyStats{
		Min:  sorted[0],
		Max:  sorted[len(sorted)-1],
		Mean: total / time.Duration(len(sorted)),
		P50:  percentile(sorted, 50),
		P95:  percentile(sorted, 95),
		P99:  percentile(sorted, 99),
	}
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(p)/100.0*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

type categoryAccumulator struct {
	recall    float64
	precision float64
	f1        float64
	mrr       float64
	ndcg      float64
	count     int
}

// GenerateSyntheticDataset creates a configurable synthetic benchmark dataset.
func GenerateSyntheticDataset(name string, queryCount int, expectedPerQuery int, categories []string) *BenchmarkDataset {
	dataset := &BenchmarkDataset{
		Name:        name,
		Description: fmt.Sprintf("Synthetic dataset with %d queries, %d expected per query", queryCount, expectedPerQuery),
		CreatedAt:   time.Now(),
	}

	for i := 0; i < queryCount; i++ {
		q := BenchmarkQuery{
			ID:    fmt.Sprintf("q%d", i),
			Query: fmt.Sprintf("synthetic query %d about topic %d", i, i%5),
		}

		// Generate expected IDs
		for j := 0; j < expectedPerQuery; j++ {
			q.ExpectedIDs = append(q.ExpectedIDs, fmt.Sprintf("m%d-%d", i, j))
		}

		// Assign category if available
		if len(categories) > 0 {
			q.Category = categories[i%len(categories)]
		}

		dataset.Queries = append(dataset.Queries, q)
	}

	return dataset
}

// FormatReport returns a human-readable summary of benchmark results.
func FormatReport(report *BenchmarkReport) string {
	var sb fmt.Stringer = &reportFormatter{report: report}
	return sb.String()
}

type reportFormatter struct {
	report *BenchmarkReport
}

func (f *reportFormatter) String() string {
	r := f.report
	return fmt.Sprintf(`Benchmark Report: %s
Queries: %d | Duration: %s | Throughput: %.1f q/s

Retrieval Metrics:
  Recall:    %.4f
  Precision: %.4f
  F1:        %.4f
  MRR:       %.4f
  NDCG:      %.4f

Latency:
  P50: %s | P95: %s | P99: %s
  Min: %s | Max: %s | Mean: %s`,
		r.DatasetName,
		r.TotalQueries, r.Duration, r.Metrics[MetricThroughput],
		r.Metrics[MetricRecall], r.Metrics[MetricPrecision],
		r.Metrics[MetricF1], r.Metrics[MetricMRR], r.Metrics[MetricNDCG],
		r.Latencies.P50, r.Latencies.P95, r.Latencies.P99,
		r.Latencies.Min, r.Latencies.Max, r.Latencies.Mean,
	)
}
