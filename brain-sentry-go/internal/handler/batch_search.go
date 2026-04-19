package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

type BatchSearchHandler struct {
	svc *service.BatchSearchService
}

func NewBatchSearchHandler(svc *service.BatchSearchService) *BatchSearchHandler {
	return &BatchSearchHandler{svc: svc}
}

// Search handles POST /v1/memories/batch-search
// Body: {"queries": ["q1", "q2"], "limit": 10, "tags": [...]}
func (h *BatchSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "batch search is not available")
		return
	}

	var req service.BatchSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Search(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
