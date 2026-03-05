package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status.
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "brainsentry",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	// HTTPRequestDuration tracks request latency by method and path.
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "brainsentry",
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration in seconds",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method", "path"})

	// HTTPRequestsInFlight tracks the number of in-flight requests.
	HTTPRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "brainsentry",
		Name:      "http_requests_in_flight",
		Help:      "Number of HTTP requests currently being processed",
	})

	// MemoriesTotal tracks total memories created.
	MemoriesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "brainsentry",
		Name:      "memories_created_total",
		Help:      "Total number of memories created",
	})

	// InterceptionsTotal tracks total interceptions.
	InterceptionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "brainsentry",
		Name:      "interceptions_total",
		Help:      "Total number of interceptions",
	}, []string{"enhanced"})

	// LLMCallsTotal tracks LLM API calls.
	LLMCallsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "brainsentry",
		Name:      "llm_calls_total",
		Help:      "Total number of LLM API calls",
	})

	// LLMCallDuration tracks LLM call latency.
	LLMCallDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "brainsentry",
		Name:      "llm_call_duration_seconds",
		Help:      "LLM API call duration in seconds",
		Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
	})
)

// Metrics returns a middleware that collects Prometheus metrics.
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			HTTPRequestsInFlight.Inc()

			rw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			HTTPRequestsInFlight.Dec()
			duration := time.Since(start).Seconds()

			path := normalizePath(r.URL.Path)
			status := strconv.Itoa(rw.statusCode)

			HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		})
	}
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizePath reduces cardinality by replacing IDs with placeholders.
func normalizePath(path string) string {
	// Simple normalization: keep first 3 segments, replace UUIDs
	segments := splitPath(path)
	for i, s := range segments {
		if looksLikeID(s) {
			segments[i] = ":id"
		}
	}
	result := "/"
	for _, s := range segments {
		if s != "" {
			result += s + "/"
		}
	}
	if len(result) > 1 {
		result = result[:len(result)-1]
	}
	return result
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func looksLikeID(s string) bool {
	if len(s) < 8 {
		return false
	}
	// UUID pattern: contains hyphens and hex chars
	hyphens := 0
	for _, c := range s {
		if c == '-' {
			hyphens++
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return hyphens >= 3
}
