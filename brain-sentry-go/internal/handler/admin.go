package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// AdminHandler handles admin/operational endpoints.
type AdminHandler struct {
	cbRegistry  *service.CircuitBreakerRegistry
	llmObserver *service.MetricsObserver
	piiService  *service.PIIService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(cbRegistry *service.CircuitBreakerRegistry, llmObserver *service.MetricsObserver, piiService *service.PIIService) *AdminHandler {
	return &AdminHandler{
		cbRegistry:  cbRegistry,
		llmObserver: llmObserver,
		piiService:  piiService,
	}
}

// GetCircuitBreakers handles GET /v1/admin/circuit-breakers — returns circuit breaker stats.
func (h *AdminHandler) GetCircuitBreakers(w http.ResponseWriter, r *http.Request) {
	if h.cbRegistry == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	stats := h.cbRegistry.AllStats()
	writeJSON(w, http.StatusOK, stats)
}

// GetLLMMetrics handles GET /v1/admin/llm-metrics — returns LLM usage metrics.
func (h *AdminHandler) GetLLMMetrics(w http.ResponseWriter, r *http.Request) {
	if h.llmObserver == nil {
		writeJSON(w, http.StatusOK, map[string]any{})
		return
	}
	summary := h.llmObserver.Summary()
	writeJSON(w, http.StatusOK, summary)
}

// ScanPII handles POST /v1/pii/scan — scans text for personally identifiable information.
func (h *AdminHandler) ScanPII(w http.ResponseWriter, r *http.Request) {
	if h.piiService == nil {
		writeError(w, http.StatusServiceUnavailable, "PII service not available")
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	masked := h.piiService.Mask(req.Text)
	writeJSON(w, http.StatusOK, map[string]string{
		"original": req.Text,
		"masked":   masked,
	})
}
