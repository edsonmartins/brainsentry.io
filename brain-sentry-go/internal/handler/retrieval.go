package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// RetrievalHandler handles retrieval planner endpoints.
type RetrievalHandler struct {
	retrievalService *service.RetrievalPlannerService
}

// NewRetrievalHandler creates a new RetrievalHandler.
func NewRetrievalHandler(retrievalService *service.RetrievalPlannerService) *RetrievalHandler {
	return &RetrievalHandler{retrievalService: retrievalService}
}

// PlanSearch handles POST /v1/memories/plan-search — performs intent-aware multi-round retrieval.
func (h *RetrievalHandler) PlanSearch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	result, err := h.retrievalService.PlanAndRetrieve(r.Context(), req.Query, req.Limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "retrieval planning failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
