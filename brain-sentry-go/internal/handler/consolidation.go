package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// ConsolidationHandler handles consolidation endpoints.
type ConsolidationHandler struct {
	consolidationService *service.ConsolidationService
}

// NewConsolidationHandler creates a new ConsolidationHandler.
func NewConsolidationHandler(consolidationService *service.ConsolidationService) *ConsolidationHandler {
	return &ConsolidationHandler{consolidationService: consolidationService}
}

// Consolidate handles POST /v1/consolidate — merges similar memories and compresses verbose ones.
func (h *ConsolidationHandler) Consolidate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SimilarityThreshold float64 `json:"similarityThreshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default threshold if no body provided
		req.SimilarityThreshold = 0.85
	}

	if req.SimilarityThreshold <= 0 || req.SimilarityThreshold > 1 {
		req.SimilarityThreshold = 0.85
	}

	result, err := h.consolidationService.ConsolidateTenant(r.Context(), req.SimilarityThreshold)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "consolidation failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
