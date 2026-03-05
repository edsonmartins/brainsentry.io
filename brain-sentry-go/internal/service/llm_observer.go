package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LLMCallEvent represents a single LLM API call with full tracing data.
type LLMCallEvent struct {
	ID            string        `json:"id"`
	Model         string        `json:"model"`
	Operation     string        `json:"operation"` // e.g. "analyze_importance", "extract_entities"
	InputTokens   int           `json:"inputTokens"`
	OutputTokens  int           `json:"outputTokens"`
	TotalTokens   int           `json:"totalTokens"`
	Latency       time.Duration `json:"latencyMs"`
	Success       bool          `json:"success"`
	ErrorMessage  string        `json:"errorMessage,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
	TenantID      string        `json:"tenantId,omitempty"`
	SessionID     string        `json:"sessionId,omitempty"`
	EstimatedCost float64       `json:"estimatedCost"` // USD
}

// LLMObserver is the interface for LLM observability backends.
type LLMObserver interface {
	OnLLMCall(ctx context.Context, event LLMCallEvent)
	Flush(ctx context.Context) error
}

// LogObserver logs LLM calls to structured logger (default backend).
type LogObserver struct{}

func (o *LogObserver) OnLLMCall(_ context.Context, event LLMCallEvent) {
	slog.Info("llm_call",
		"operation", event.Operation,
		"model", event.Model,
		"latencyMs", event.Latency.Milliseconds(),
		"tokens", event.TotalTokens,
		"success", event.Success,
		"cost", event.EstimatedCost,
	)
}

func (o *LogObserver) Flush(_ context.Context) error { return nil }

// MetricsObserver collects aggregate metrics for LLM calls.
type MetricsObserver struct {
	mu             sync.RWMutex
	totalCalls     int64
	totalTokens    int64
	totalLatencyMs int64
	totalCost      float64
	totalErrors    int64
	byOperation    map[string]*OperationMetrics
}

// OperationMetrics holds per-operation aggregate metrics.
type OperationMetrics struct {
	Calls      int64   `json:"calls"`
	Tokens     int64   `json:"tokens"`
	LatencyMs  int64   `json:"latencyMs"`
	Errors     int64   `json:"errors"`
	Cost       float64 `json:"cost"`
	AvgLatency float64 `json:"avgLatencyMs"`
}

// NewMetricsObserver creates a new metrics collecting observer.
func NewMetricsObserver() *MetricsObserver {
	return &MetricsObserver{
		byOperation: make(map[string]*OperationMetrics),
	}
}

func (o *MetricsObserver) OnLLMCall(_ context.Context, event LLMCallEvent) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.totalCalls++
	o.totalTokens += int64(event.TotalTokens)
	o.totalLatencyMs += event.Latency.Milliseconds()
	o.totalCost += event.EstimatedCost

	if !event.Success {
		o.totalErrors++
	}

	op, ok := o.byOperation[event.Operation]
	if !ok {
		op = &OperationMetrics{}
		o.byOperation[event.Operation] = op
	}
	op.Calls++
	op.Tokens += int64(event.TotalTokens)
	op.LatencyMs += event.Latency.Milliseconds()
	op.Cost += event.EstimatedCost
	if !event.Success {
		op.Errors++
	}
	if op.Calls > 0 {
		op.AvgLatency = float64(op.LatencyMs) / float64(op.Calls)
	}
}

func (o *MetricsObserver) Flush(_ context.Context) error { return nil }

// Summary returns aggregate metrics.
func (o *MetricsObserver) Summary() LLMMetricsSummary {
	o.mu.RLock()
	defer o.mu.RUnlock()

	avg := float64(0)
	if o.totalCalls > 0 {
		avg = float64(o.totalLatencyMs) / float64(o.totalCalls)
	}

	ops := make(map[string]OperationMetrics, len(o.byOperation))
	for k, v := range o.byOperation {
		ops[k] = *v
	}

	return LLMMetricsSummary{
		TotalCalls:      o.totalCalls,
		TotalTokens:     o.totalTokens,
		TotalCost:       o.totalCost,
		TotalErrors:     o.totalErrors,
		AvgLatencyMs:    avg,
		ByOperation:     ops,
	}
}

// LLMMetricsSummary holds aggregate LLM metrics.
type LLMMetricsSummary struct {
	TotalCalls   int64                      `json:"totalCalls"`
	TotalTokens  int64                      `json:"totalTokens"`
	TotalCost    float64                    `json:"totalCost"`
	TotalErrors  int64                      `json:"totalErrors"`
	AvgLatencyMs float64                    `json:"avgLatencyMs"`
	ByOperation  map[string]OperationMetrics `json:"byOperation"`
}

// BufferedObserver batches events and flushes to a delegate periodically.
type BufferedObserver struct {
	mu       sync.Mutex
	delegate LLMObserver
	buffer   []LLMCallEvent
	maxSize  int
}

// NewBufferedObserver creates a buffered observer that flushes at maxSize events.
func NewBufferedObserver(delegate LLMObserver, maxSize int) *BufferedObserver {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &BufferedObserver{
		delegate: delegate,
		buffer:   make([]LLMCallEvent, 0, maxSize),
		maxSize:  maxSize,
	}
}

func (o *BufferedObserver) OnLLMCall(ctx context.Context, event LLMCallEvent) {
	o.mu.Lock()
	o.buffer = append(o.buffer, event)
	shouldFlush := len(o.buffer) >= o.maxSize
	o.mu.Unlock()

	if shouldFlush {
		o.Flush(ctx)
	}
}

func (o *BufferedObserver) Flush(ctx context.Context) error {
	o.mu.Lock()
	events := make([]LLMCallEvent, len(o.buffer))
	copy(events, o.buffer)
	o.buffer = o.buffer[:0]
	o.mu.Unlock()

	for _, event := range events {
		o.delegate.OnLLMCall(ctx, event)
	}
	return o.delegate.Flush(ctx)
}

// MultiObserver fans out events to multiple observers.
type MultiObserver struct {
	observers []LLMObserver
}

// NewMultiObserver creates an observer that sends to all delegates.
func NewMultiObserver(observers ...LLMObserver) *MultiObserver {
	return &MultiObserver{observers: observers}
}

func (o *MultiObserver) OnLLMCall(ctx context.Context, event LLMCallEvent) {
	for _, obs := range o.observers {
		obs.OnLLMCall(ctx, event)
	}
}

func (o *MultiObserver) Flush(ctx context.Context) error {
	for _, obs := range o.observers {
		if err := obs.Flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

// EstimateCost estimates the USD cost for a given model and token counts.
// Prices are approximate and should be updated periodically.
func EstimateCost(model string, inputTokens, outputTokens int) float64 {
	// Approximate pricing per 1M tokens (input/output)
	prices := map[string][2]float64{
		"gpt-4":            {30.0, 60.0},
		"gpt-4-turbo":      {10.0, 30.0},
		"gpt-4o":           {2.5, 10.0},
		"gpt-4o-mini":      {0.15, 0.6},
		"gpt-3.5-turbo":    {0.5, 1.5},
		"claude-3-opus":    {15.0, 75.0},
		"claude-3-sonnet":  {3.0, 15.0},
		"claude-3-haiku":   {0.25, 1.25},
		"claude-3.5-sonnet":{3.0, 15.0},
	}

	p, ok := prices[model]
	if !ok {
		p = [2]float64{1.0, 2.0} // default estimate
	}

	inputCost := float64(inputTokens) / 1_000_000 * p[0]
	outputCost := float64(outputTokens) / 1_000_000 * p[1]
	return inputCost + outputCost
}
