package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

type QueryRouterHandler struct {
	svc *service.QueryRouterService
}

func NewQueryRouterHandler(svc *service.QueryRouterService) *QueryRouterHandler {
	return &QueryRouterHandler{svc: svc}
}

// Classify handles POST /v1/router/classify
// Body: {"query": "..."}
// Returns: the RouterDecision with chosen strategy and confidence.
func (h *QueryRouterHandler) Classify(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "query router is not available")
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	decision := h.svc.Classify(req.Query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}
