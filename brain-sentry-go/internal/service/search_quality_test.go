package service

import (
	"math"
	"testing"
)

func TestRecallAtK_AllRelevant(t *testing.T) {
	retrieved := []string{"a", "b", "c"}
	relevant := []string{"a", "b", "c"}

	recall := ComputeRecallAtK(retrieved, relevant, 10)
	if recall != 1.0 {
		t.Errorf("expected 1.0, got %f", recall)
	}
}

func TestRecallAtK_Partial(t *testing.T) {
	retrieved := []string{"a", "x", "b", "y", "c"}
	relevant := []string{"a", "b", "c", "d"}

	recall := ComputeRecallAtK(retrieved, relevant, 5)
	// 3 out of 4 relevant found
	if math.Abs(recall-0.75) > 0.001 {
		t.Errorf("expected 0.75, got %f", recall)
	}
}

func TestRecallAtK_None(t *testing.T) {
	retrieved := []string{"x", "y", "z"}
	relevant := []string{"a", "b"}

	recall := ComputeRecallAtK(retrieved, relevant, 3)
	if recall != 0.0 {
		t.Errorf("expected 0.0, got %f", recall)
	}
}

func TestPrecisionAtK(t *testing.T) {
	retrieved := []string{"a", "x", "b", "y", "z"}
	relevant := []string{"a", "b"}

	precision := ComputePrecisionAtK(retrieved, relevant, 5)
	// 2 relevant in top 5 → 0.4
	if math.Abs(precision-0.4) > 0.001 {
		t.Errorf("expected 0.4, got %f", precision)
	}
}

func TestMRR_FirstPosition(t *testing.T) {
	retrieved := []string{"a", "b", "c"}
	relevant := []string{"a"}

	mrr := ComputeMRR(retrieved, relevant)
	if mrr != 1.0 {
		t.Errorf("expected 1.0, got %f", mrr)
	}
}

func TestMRR_ThirdPosition(t *testing.T) {
	retrieved := []string{"x", "y", "a"}
	relevant := []string{"a"}

	mrr := ComputeMRR(retrieved, relevant)
	expected := 1.0 / 3.0
	if math.Abs(mrr-expected) > 0.001 {
		t.Errorf("expected %f, got %f", expected, mrr)
	}
}

func TestMRR_NotFound(t *testing.T) {
	retrieved := []string{"x", "y", "z"}
	relevant := []string{"a"}

	mrr := ComputeMRR(retrieved, relevant)
	if mrr != 0.0 {
		t.Errorf("expected 0.0, got %f", mrr)
	}
}

func TestNDCGAtK_Perfect(t *testing.T) {
	retrieved := []string{"a", "b", "c"}
	relevant := []string{"a", "b", "c"}

	ndcg := ComputeNDCGAtK(retrieved, relevant, nil, 3)
	if math.Abs(ndcg-1.0) > 0.001 {
		t.Errorf("expected 1.0, got %f", ndcg)
	}
}

func TestNDCGAtK_Graded(t *testing.T) {
	// Ideal order: c(3), b(2), a(1)
	// Retrieved:   c(3), b(2), a(1) → perfect
	retrieved := []string{"c", "b", "a"}
	relevant := []string{"a", "b", "c"}
	grades := map[string]int{"a": 1, "b": 2, "c": 3}

	ndcg := ComputeNDCGAtK(retrieved, relevant, grades, 3)
	if math.Abs(ndcg-1.0) > 0.001 {
		t.Errorf("expected 1.0 for ideal ordering, got %f", ndcg)
	}
}

func TestComputeSearchMetrics_Full(t *testing.T) {
	retrieved := []string{"a", "x", "b"}
	relevant := []string{"a", "b", "c"}

	metrics := ComputeSearchMetrics(retrieved, relevant, nil, 3)

	// Recall: 2/3
	if math.Abs(metrics.RecallAtK-2.0/3.0) > 0.001 {
		t.Errorf("expected recall 0.667, got %f", metrics.RecallAtK)
	}
	// Precision: 2/3
	if math.Abs(metrics.PrecisionAtK-2.0/3.0) > 0.001 {
		t.Errorf("expected precision 0.667, got %f", metrics.PrecisionAtK)
	}
	// MRR: 1/1 (first result is relevant)
	if metrics.MRR != 1.0 {
		t.Errorf("expected MRR 1.0, got %f", metrics.MRR)
	}
	// F1 should be non-zero
	if metrics.F1AtK == 0 {
		t.Error("expected non-zero F1")
	}
}

func TestAverageMetrics(t *testing.T) {
	m1 := SearchQualityMetrics{RecallAtK: 0.8, PrecisionAtK: 0.6, MRR: 1.0, NDCGAtK: 0.9, K: 10}
	m2 := SearchQualityMetrics{RecallAtK: 0.6, PrecisionAtK: 0.4, MRR: 0.5, NDCGAtK: 0.7, K: 10}

	avg := AverageMetrics([]SearchQualityMetrics{m1, m2})

	if math.Abs(avg.RecallAtK-0.7) > 0.001 {
		t.Errorf("expected avg recall 0.7, got %f", avg.RecallAtK)
	}
	if math.Abs(avg.MRR-0.75) > 0.001 {
		t.Errorf("expected avg MRR 0.75, got %f", avg.MRR)
	}
}
