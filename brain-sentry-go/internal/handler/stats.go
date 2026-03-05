package handler

import (
	"net/http"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

// StatsHandler handles statistics endpoints.
type StatsHandler struct {
	memoryRepo *postgres.MemoryRepository
	auditRepo  *postgres.AuditRepository
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(memoryRepo *postgres.MemoryRepository, auditRepo *postgres.AuditRepository) *StatsHandler {
	return &StatsHandler{
		memoryRepo: memoryRepo,
		auditRepo:  auditRepo,
	}
}

// Overview handles GET /v1/stats/overview
func (h *StatsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	totalMemories, _ := h.memoryRepo.Count(ctx)
	memoriesByCategory, _ := h.memoryRepo.CountByCategory(ctx)
	memoriesByImportance, _ := h.memoryRepo.CountByImportance(ctx)
	requestsToday, _ := h.auditRepo.CountToday(ctx)
	totalInjections, _ := h.auditRepo.CountByEventTypeValue(ctx, "context_injection")
	avgLatency, _ := h.auditRepo.AverageLatency(ctx)
	activeMemories24h, _ := h.memoryRepo.CountActiveRecent(ctx)

	var injectionRate float64
	if requestsToday > 0 {
		injectionRate = float64(totalInjections) / float64(requestsToday) * 100
	}

	resp := dto.StatsResponse{
		TotalMemories:        totalMemories,
		MemoriesByCategory:   memoriesByCategory,
		MemoriesByImportance: memoriesByImportance,
		RequestsToday:        requestsToday,
		TotalInjections:      totalInjections,
		InjectionRate:        injectionRate,
		AvgLatencyMs:         avgLatency,
		ActiveMemories24h:    activeMemories24h,
	}

	writeJSON(w, http.StatusOK, resp)
}

// TopPatterns handles GET /v1/stats/top-patterns
func (h *StatsHandler) TopPatterns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get top categories by injection count
	byCat, _ := h.memoryRepo.CountByCategory(ctx)

	patterns := make([]map[string]any, 0, len(byCat))
	for cat, count := range byCat {
		patterns = append(patterns, map[string]any{
			"category": cat,
			"count":    count,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"patterns": patterns,
		"total":    len(patterns),
	})
}

// HealthStats handles GET /v1/stats/health
func (h *StatsHandler) HealthStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "UP",
	})
}
