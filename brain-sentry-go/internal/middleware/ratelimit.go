package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiterConfig configures the rate limiter.
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func newTokenBucket(maxTokens float64, refillRate float64) *tokenBucket {
	return &tokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter stores per-tenant rate limiters.
type RateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
	config  RateLimiterConfig
}

// NewRateLimiter creates a new in-memory rate limiter.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60
	}
	if cfg.BurstSize <= 0 {
		cfg.BurstSize = cfg.RequestsPerMinute
	}
	return &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		config:  cfg,
	}
}

func (rl *RateLimiter) getBucket(tenantID string) *tokenBucket {
	rl.mu.RLock()
	b, ok := rl.buckets[tenantID]
	rl.mu.RUnlock()
	if ok {
		return b
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()
	// Double-check after acquiring write lock
	if b, ok := rl.buckets[tenantID]; ok {
		return b
	}
	b = newTokenBucket(float64(rl.config.BurstSize), float64(rl.config.RequestsPerMinute)/60.0)
	rl.buckets[tenantID] = b
	return b
}

// RateLimit returns an HTTP middleware that rate-limits requests per tenant.
func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract tenant from header or query (middleware runs before tenant extractor populates context)
			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				tenantID = r.URL.Query().Get("tenantId")
			}
			if tenantID == "" {
				tenantID = "default"
			}

			bucket := rl.getBucket(tenantID)
			if !bucket.allow() {
				w.Header().Set("Retry-After", "60")
				http.Error(w, fmt.Sprintf(`{"error":"rate limit exceeded","status":429,"tenant":"%s"}`, tenantID), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
