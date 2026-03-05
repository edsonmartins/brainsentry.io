package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedOnInit(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("test"))
	if cb.State() != CircuitClosed {
		t.Errorf("expected CLOSED, got %s", cb.State())
	}
}

func TestCircuitBreaker_SuccessKeepsClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("test"))
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected CLOSED after success, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:             "test",
		FailureThreshold: 3,
		SuccessThreshold: 1,
		OpenTimeout:      1 * time.Second,
		MaxRetries:       0, // no retries for fast test
		BaseBackoff:      1 * time.Millisecond,
		MaxBackoff:       10 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("service down")
	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.State() != CircuitOpen {
		t.Errorf("expected OPEN after %d failures, got %s", 3, cb.State())
	}
}

func TestCircuitBreaker_RejectsWhenOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:             "test",
		FailureThreshold: 1,
		OpenTimeout:      10 * time.Second,
		MaxRetries:       0,
		BaseBackoff:      1 * time.Millisecond,
		MaxBackoff:       10 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Trip the breaker
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	// Should reject
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err == nil {
		t.Error("expected rejection when open")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      50 * time.Millisecond,
		MaxRetries:       0,
		BaseBackoff:      1 * time.Millisecond,
		MaxBackoff:       10 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Trip
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	if cb.State() != CircuitOpen {
		t.Fatal("expected OPEN")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to half-open and succeed
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error after half-open: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected CLOSED after successful probe, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailReopens(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:             "test",
		FailureThreshold: 1,
		SuccessThreshold: 2,
		OpenTimeout:      50 * time.Millisecond,
		MaxRetries:       0,
		BaseBackoff:      1 * time.Millisecond,
		MaxBackoff:       10 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Trip
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	time.Sleep(60 * time.Millisecond)

	// Probe fails
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("still failing")
	})

	if cb.State() != CircuitOpen {
		t.Errorf("expected OPEN after failed probe, got %s", cb.State())
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("test-stats"))
	cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
	cb.Execute(context.Background(), func(ctx context.Context) error { return nil })

	stats := cb.Stats()
	if stats.Name != "test-stats" {
		t.Errorf("expected name test-stats, got %s", stats.Name)
	}
	if stats.TotalCalls != 2 {
		t.Errorf("expected 2 calls, got %d", stats.TotalCalls)
	}
	if stats.TotalSuccesses != 2 {
		t.Errorf("expected 2 successes, got %d", stats.TotalSuccesses)
	}
}

func TestCircuitBreakerRegistry_GetOrCreate(t *testing.T) {
	reg := NewCircuitBreakerRegistry()
	cb1 := reg.Get("svc-a")
	cb2 := reg.Get("svc-a")
	if cb1 != cb2 {
		t.Error("expected same instance for same name")
	}
	cb3 := reg.Get("svc-b")
	if cb1 == cb3 {
		t.Error("expected different instance for different name")
	}
}

func TestCircuitBreakerRegistry_AllStats(t *testing.T) {
	reg := NewCircuitBreakerRegistry()
	reg.Get("svc-a")
	reg.Get("svc-b")
	stats := reg.AllStats()
	if len(stats) != 2 {
		t.Errorf("expected 2 stats, got %d", len(stats))
	}
}

func TestCircuitState_String(t *testing.T) {
	if CircuitClosed.String() != "CLOSED" {
		t.Error("expected CLOSED")
	}
	if CircuitOpen.String() != "OPEN" {
		t.Error("expected OPEN")
	}
	if CircuitHalfOpen.String() != "HALF_OPEN" {
		t.Error("expected HALF_OPEN")
	}
}

func TestExecuteWithResult_Success(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("test"))
	result, err := ExecuteWithResult(cb, context.Background(), func(ctx context.Context) (string, error) {
		return "hello", nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected hello, got %s", result)
	}
}

func TestExecuteWithResult_Failure(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.MaxRetries = 0
	cb := NewCircuitBreaker(config)
	_, err := ExecuteWithResult(cb, context.Background(), func(ctx context.Context) (string, error) {
		return "", errors.New("fail")
	})
	if err == nil {
		t.Error("expected error")
	}
}
