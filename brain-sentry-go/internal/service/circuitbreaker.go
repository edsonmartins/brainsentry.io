package service

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // failing, reject calls
	CircuitHalfOpen                     // testing recovery
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	Name             string        // service name for logging
	FailureThreshold int           // failures before opening
	SuccessThreshold int           // successes in half-open before closing
	OpenTimeout      time.Duration // how long to stay open before half-open
	MaxRetries       int           // max retries with backoff
	BaseBackoff      time.Duration // initial backoff duration
	MaxBackoff       time.Duration // maximum backoff duration
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:             name,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		OpenTimeout:      30 * time.Second,
		MaxRetries:       3,
		BaseBackoff:      500 * time.Millisecond,
		MaxBackoff:       10 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern with exponential backoff + jitter.
type CircuitBreaker struct {
	mu               sync.RWMutex
	config           CircuitBreakerConfig
	state            CircuitState
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	lastStateChange  time.Time
	totalCalls       int64
	totalFailures    int64
	totalSuccesses   int64
	consecutiveOk    int
}

// NewCircuitBreaker creates a new circuit breaker with the given config.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:          config,
		state:           CircuitClosed,
		lastStateChange: time.Now(),
	}
}

// Execute runs fn through the circuit breaker with retry and backoff.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := cb.canExecute(); err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= cb.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := cb.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := fn(ctx)
		if err == nil {
			cb.recordSuccess()
			return nil
		}

		lastErr = err
		slog.Debug("circuit breaker call failed",
			"service", cb.config.Name,
			"attempt", attempt+1,
			"error", err,
		)
	}

	cb.recordFailure()
	return fmt.Errorf("[%s] circuit breaker: %w", cb.config.Name, lastErr)
}

// ExecuteWithResult runs fn and returns result + error through the circuit breaker.
func ExecuteWithResult[T any](cb *CircuitBreaker, ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T

	if err := cb.canExecute(); err != nil {
		return zero, err
	}

	var lastErr error
	for attempt := 0; attempt <= cb.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := cb.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := fn(ctx)
		if err == nil {
			cb.recordSuccess()
			return result, nil
		}

		lastErr = err
	}

	cb.recordFailure()
	return zero, fmt.Errorf("[%s] circuit breaker: %w", cb.config.Name, lastErr)
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns circuit breaker statistics.
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return CircuitBreakerStats{
		Name:            cb.config.Name,
		State:           cb.state.String(),
		TotalCalls:      cb.totalCalls,
		TotalSuccesses:  cb.totalSuccesses,
		TotalFailures:   cb.totalFailures,
		FailureCount:    cb.failureCount,
		LastFailureTime: cb.lastFailureTime,
		LastStateChange: cb.lastStateChange,
	}
}

// CircuitBreakerStats holds stats for monitoring.
type CircuitBreakerStats struct {
	Name            string    `json:"name"`
	State           string    `json:"state"`
	TotalCalls      int64     `json:"totalCalls"`
	TotalSuccesses  int64     `json:"totalSuccesses"`
	TotalFailures   int64     `json:"totalFailures"`
	FailureCount    int       `json:"failureCount"`
	LastFailureTime time.Time `json:"lastFailureTime,omitempty"`
	LastStateChange time.Time `json:"lastStateChange"`
}

func (cb *CircuitBreaker) canExecute() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalCalls++

	switch cb.state {
	case CircuitClosed:
		return nil

	case CircuitOpen:
		// Check if open timeout has elapsed
		if time.Since(cb.lastStateChange) >= cb.config.OpenTimeout {
			cb.setState(CircuitHalfOpen)
			slog.Info("circuit breaker transitioning to half-open", "service", cb.config.Name)
			return nil
		}
		return fmt.Errorf("[%s] circuit breaker is OPEN, rejecting call", cb.config.Name)

	case CircuitHalfOpen:
		return nil // allow probe calls
	}

	return nil
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalSuccesses++
	cb.consecutiveOk++

	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.setState(CircuitClosed)
			cb.failureCount = 0
			cb.successCount = 0
			cb.consecutiveOk = 0
			slog.Info("circuit breaker recovered, closing", "service", cb.config.Name)
		}
	case CircuitClosed:
		// Reset failure count on success
		if cb.consecutiveOk > cb.config.FailureThreshold {
			cb.failureCount = 0
			cb.consecutiveOk = 0
		}
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalFailures++
	cb.failureCount++
	cb.consecutiveOk = 0
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitClosed:
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.setState(CircuitOpen)
			slog.Warn("circuit breaker opened", "service", cb.config.Name, "failures", cb.failureCount)
		}
	case CircuitHalfOpen:
		cb.setState(CircuitOpen)
		cb.successCount = 0
		slog.Warn("circuit breaker probe failed, reopening", "service", cb.config.Name)
	}
}

func (cb *CircuitBreaker) setState(state CircuitState) {
	cb.state = state
	cb.lastStateChange = time.Now()
}

// calculateBackoff returns exponential backoff with jitter.
func (cb *CircuitBreaker) calculateBackoff(attempt int) time.Duration {
	base := cb.config.BaseBackoff
	for i := 1; i < attempt; i++ {
		base *= 2
		if base > cb.config.MaxBackoff {
			base = cb.config.MaxBackoff
			break
		}
	}
	// Add jitter: ±25%
	jitter := time.Duration(float64(base) * (0.75 + rand.Float64()*0.5))
	return jitter
}

// CircuitBreakerRegistry manages multiple circuit breakers.
type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
}

// NewCircuitBreakerRegistry creates a new registry.
func NewCircuitBreakerRegistry() *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Get returns a circuit breaker by name, creating it with defaults if not found.
func (r *CircuitBreakerRegistry) Get(name string) *CircuitBreaker {
	r.mu.RLock()
	if cb, ok := r.breakers[name]; ok {
		r.mu.RUnlock()
		return cb
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after write lock
	if cb, ok := r.breakers[name]; ok {
		return cb
	}

	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig(name))
	r.breakers[name] = cb
	return cb
}

// Register adds a circuit breaker with custom config.
func (r *CircuitBreakerRegistry) Register(config CircuitBreakerConfig) *CircuitBreaker {
	r.mu.Lock()
	defer r.mu.Unlock()
	cb := NewCircuitBreaker(config)
	r.breakers[config.Name] = cb
	return cb
}

// AllStats returns stats for all circuit breakers.
func (r *CircuitBreakerRegistry) AllStats() []CircuitBreakerStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stats := make([]CircuitBreakerStats, 0, len(r.breakers))
	for _, cb := range r.breakers {
		stats = append(stats, cb.Stats())
	}
	return stats
}
