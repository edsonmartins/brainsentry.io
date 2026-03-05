package service

import (
	"context"
	"testing"
	"time"
)

func TestLogObserver_OnLLMCall(t *testing.T) {
	obs := &LogObserver{}
	obs.OnLLMCall(context.Background(), LLMCallEvent{
		Operation: "test",
		Model:     "gpt-4o",
		Success:   true,
		Latency:   100 * time.Millisecond,
	})
	// No panic = pass
}

func TestLogObserver_Flush(t *testing.T) {
	obs := &LogObserver{}
	if err := obs.Flush(context.Background()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMetricsObserver_Aggregates(t *testing.T) {
	obs := NewMetricsObserver()
	ctx := context.Background()

	obs.OnLLMCall(ctx, LLMCallEvent{
		Operation:   "analyze",
		TotalTokens: 100,
		Latency:     50 * time.Millisecond,
		Success:     true,
		EstimatedCost: 0.001,
	})
	obs.OnLLMCall(ctx, LLMCallEvent{
		Operation:   "analyze",
		TotalTokens: 200,
		Latency:     100 * time.Millisecond,
		Success:     false,
		EstimatedCost: 0.002,
	})
	obs.OnLLMCall(ctx, LLMCallEvent{
		Operation:   "extract",
		TotalTokens: 150,
		Latency:     75 * time.Millisecond,
		Success:     true,
		EstimatedCost: 0.0015,
	})

	summary := obs.Summary()
	if summary.TotalCalls != 3 {
		t.Errorf("expected 3 calls, got %d", summary.TotalCalls)
	}
	if summary.TotalTokens != 450 {
		t.Errorf("expected 450 tokens, got %d", summary.TotalTokens)
	}
	if summary.TotalErrors != 1 {
		t.Errorf("expected 1 error, got %d", summary.TotalErrors)
	}
	if summary.TotalCost < 0.004 {
		t.Errorf("expected cost >= 0.004, got %f", summary.TotalCost)
	}

	analyzeOp, ok := summary.ByOperation["analyze"]
	if !ok {
		t.Fatal("expected analyze operation")
	}
	if analyzeOp.Calls != 2 {
		t.Errorf("expected 2 analyze calls, got %d", analyzeOp.Calls)
	}
	if analyzeOp.Errors != 1 {
		t.Errorf("expected 1 analyze error, got %d", analyzeOp.Errors)
	}
}

func TestBufferedObserver_FlushesAtThreshold(t *testing.T) {
	metrics := NewMetricsObserver()
	buffered := NewBufferedObserver(metrics, 2)
	ctx := context.Background()

	buffered.OnLLMCall(ctx, LLMCallEvent{Operation: "op1", Success: true})
	// Not yet flushed
	if metrics.Summary().TotalCalls != 0 {
		t.Error("expected 0 calls before flush threshold")
	}

	buffered.OnLLMCall(ctx, LLMCallEvent{Operation: "op2", Success: true})
	// Should have flushed
	if metrics.Summary().TotalCalls != 2 {
		t.Errorf("expected 2 calls after flush, got %d", metrics.Summary().TotalCalls)
	}
}

func TestBufferedObserver_ManualFlush(t *testing.T) {
	metrics := NewMetricsObserver()
	buffered := NewBufferedObserver(metrics, 100)
	ctx := context.Background()

	buffered.OnLLMCall(ctx, LLMCallEvent{Operation: "op1", Success: true})
	buffered.Flush(ctx)

	if metrics.Summary().TotalCalls != 1 {
		t.Errorf("expected 1 call after manual flush, got %d", metrics.Summary().TotalCalls)
	}
}

func TestMultiObserver_FansOut(t *testing.T) {
	m1 := NewMetricsObserver()
	m2 := NewMetricsObserver()
	multi := NewMultiObserver(m1, m2)
	ctx := context.Background()

	multi.OnLLMCall(ctx, LLMCallEvent{Operation: "op", TotalTokens: 100, Success: true})

	if m1.Summary().TotalCalls != 1 {
		t.Error("expected 1 call on m1")
	}
	if m2.Summary().TotalCalls != 1 {
		t.Error("expected 1 call on m2")
	}
}

func TestEstimateCost_KnownModel(t *testing.T) {
	cost := EstimateCost("gpt-4o", 1000, 500)
	if cost <= 0 {
		t.Error("expected positive cost")
	}
}

func TestEstimateCost_UnknownModel(t *testing.T) {
	cost := EstimateCost("unknown-model", 1000, 500)
	if cost <= 0 {
		t.Error("expected positive cost even for unknown model")
	}
}

func TestLLMCallEvent_Structure(t *testing.T) {
	event := LLMCallEvent{
		ID:            "evt-1",
		Model:         "gpt-4o",
		Operation:     "analyze",
		InputTokens:   500,
		OutputTokens:  200,
		TotalTokens:   700,
		Latency:       150 * time.Millisecond,
		Success:       true,
		Timestamp:     time.Now(),
		TenantID:      "t1",
		SessionID:     "s1",
		EstimatedCost: 0.005,
	}
	if event.TotalTokens != 700 {
		t.Error("expected 700 tokens")
	}
}

func TestNewBufferedObserver_DefaultSize(t *testing.T) {
	obs := NewBufferedObserver(&LogObserver{}, 0)
	if obs.maxSize != 100 {
		t.Errorf("expected default 100, got %d", obs.maxSize)
	}
}
