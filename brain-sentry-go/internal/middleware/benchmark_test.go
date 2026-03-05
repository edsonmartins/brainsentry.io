package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10000,
		BurstSize:         10000,
	})
	bucket := rl.getBucket("bench-tenant")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.allow()
	}
}

func BenchmarkRateLimiter_GetBucket(b *testing.B) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1000,
		BurstSize:         100,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.getBucket("tenant-1")
	}
}

func BenchmarkCORS_Middleware(b *testing.B) {
	h := CORS([]string{"http://localhost:3000"}, []string{"GET", "POST"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
	}
}

func BenchmarkRequireRole_Allowed(b *testing.B) {
	h := RequireRole(RoleAdmin)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &AuthClaims{Roles: []string{"ADMIN"}}
	ctx := context.WithValue(req.Context(), authContextKey{}, claims)
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
	}
}
