package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

type CascadeExtractionHandler struct {
	svc *service.CascadeEntityExtractionService
}

func NewCascadeExtractionHandler(svc *service.CascadeEntityExtractionService) *CascadeExtractionHandler {
	return &CascadeExtractionHandler{svc: svc}
}

// Extract handles POST /v1/cascade-extract
// Body: {"content": "..."}
// Returns: entities + relationships from 3-pass cascade extraction.
func (h *CascadeExtractionHandler) Extract(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "cascade extraction is not available (LLM disabled)")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	result, err := h.svc.Extract(r.Context(), req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
