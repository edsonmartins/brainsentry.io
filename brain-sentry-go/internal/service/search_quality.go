package service

import (
	"math"
	"sort"
)

// SearchQualityMetrics holds standard IR evaluation metrics.
type SearchQualityMetrics struct {
	RecallAtK   float64 `json:"recallAtK"`
	PrecisionAtK float64 `json:"precisionAtK"`
	F1AtK       float64 `json:"f1AtK"`
	MRR         float64 `json:"mrr"`          // Mean Reciprocal Rank
	NDCGAtK     float64 `json:"ndcgAtK"`      // Normalized Discounted Cumulative Gain
	K           int     `json:"k"`
}

// LabeledQuery represents a query with known relevant document IDs for benchmarking.
type LabeledQuery struct {
	Query       string            `json:"query"`
	RelevantIDs []string          `json:"relevantIds"`
	Grades      map[string]int    `json:"grades,omitempty"` // graded relevance (0-3) per ID
}

// SearchBenchmarkResult holds results from running a benchmark suite.
type SearchBenchmarkResult struct {
	QueryCount     int                  `json:"queryCount"`
	AvgMetrics     SearchQualityMetrics `json:"avgMetrics"`
	PerQuery       []QueryResult        `json:"perQuery"`
	TotalTimeMs    int64                `json:"totalTimeMs"`
}

// QueryResult holds metrics for a single query.
type QueryResult struct {
	Query   string               `json:"query"`
	Metrics SearchQualityMetrics `json:"metrics"`
}

// ComputeRecallAtK computes recall@k: proportion of relevant documents retrieved in top-k.
func ComputeRecallAtK(retrieved []string, relevant []string, k int) float64 {
	if len(relevant) == 0 {
		return 1.0 // no relevant docs → perfect recall vacuously
	}

	relevantSet := toSet(relevant)
	hits := 0
	topK := min(k, len(retrieved))

	for i := 0; i < topK; i++ {
		if relevantSet[retrieved[i]] {
			hits++
		}
	}

	return float64(hits) / float64(len(relevant))
}

// ComputePrecisionAtK computes precision@k: proportion of top-k results that are relevant.
func ComputePrecisionAtK(retrieved []string, relevant []string, k int) float64 {
	relevantSet := toSet(relevant)
	hits := 0
	topK := min(k, len(retrieved))
	if topK == 0 {
		return 0
	}

	for i := 0; i < topK; i++ {
		if relevantSet[retrieved[i]] {
			hits++
		}
	}

	return float64(hits) / float64(topK)
}

// ComputeMRR computes Mean Reciprocal Rank: 1/rank of first relevant result.
func ComputeMRR(retrieved []string, relevant []string) float64 {
	relevantSet := toSet(relevant)
	for i, id := range retrieved {
		if relevantSet[id] {
			return 1.0 / float64(i+1)
		}
	}
	return 0
}

// ComputeNDCGAtK computes Normalized Discounted Cumulative Gain at k.
// Uses graded relevance if provided, otherwise binary (relevant=1, not=0).
func ComputeNDCGAtK(retrieved []string, relevant []string, grades map[string]int, k int) float64 {
	topK := min(k, len(retrieved))
	if topK == 0 {
		return 0
	}

	relevantSet := toSet(relevant)

	// DCG
	dcg := 0.0
	for i := 0; i < topK; i++ {
		var gain float64
		if grades != nil {
			if g, ok := grades[retrieved[i]]; ok {
				gain = float64(g)
			}
		} else if relevantSet[retrieved[i]] {
			gain = 1.0
		}
		dcg += gain / math.Log2(float64(i+2)) // i+2 because log2(1)=0
	}

	// Ideal DCG (sort grades descending)
	var idealGains []float64
	if grades != nil {
		for _, id := range relevant {
			if g, ok := grades[id]; ok {
				idealGains = append(idealGains, float64(g))
			} else {
				idealGains = append(idealGains, 1.0)
			}
		}
	} else {
		for range relevant {
			idealGains = append(idealGains, 1.0)
		}
	}

	sort.Float64s(idealGains)
	// Reverse
	for i, j := 0, len(idealGains)-1; i < j; i, j = i+1, j-1 {
		idealGains[i], idealGains[j] = idealGains[j], idealGains[i]
	}

	idcg := 0.0
	for i := 0; i < min(k, len(idealGains)); i++ {
		idcg += idealGains[i] / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0
	}

	return dcg / idcg
}

// ComputeSearchMetrics computes all metrics for a single query.
func ComputeSearchMetrics(retrieved []string, relevant []string, grades map[string]int, k int) SearchQualityMetrics {
	recall := ComputeRecallAtK(retrieved, relevant, k)
	precision := ComputePrecisionAtK(retrieved, relevant, k)
	mrr := ComputeMRR(retrieved, relevant)
	ndcg := ComputeNDCGAtK(retrieved, relevant, grades, k)

	f1 := 0.0
	if recall+precision > 0 {
		f1 = 2 * recall * precision / (recall + precision)
	}

	return SearchQualityMetrics{
		RecallAtK:    recall,
		PrecisionAtK: precision,
		F1AtK:        f1,
		MRR:          mrr,
		NDCGAtK:      ndcg,
		K:            k,
	}
}

// AverageMetrics computes the average of multiple SearchQualityMetrics.
func AverageMetrics(metrics []SearchQualityMetrics) SearchQualityMetrics {
	if len(metrics) == 0 {
		return SearchQualityMetrics{}
	}

	avg := SearchQualityMetrics{K: metrics[0].K}
	for _, m := range metrics {
		avg.RecallAtK += m.RecallAtK
		avg.PrecisionAtK += m.PrecisionAtK
		avg.F1AtK += m.F1AtK
		avg.MRR += m.MRR
		avg.NDCGAtK += m.NDCGAtK
	}

	n := float64(len(metrics))
	avg.RecallAtK /= n
	avg.PrecisionAtK /= n
	avg.F1AtK /= n
	avg.MRR /= n
	avg.NDCGAtK /= n

	return avg
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
