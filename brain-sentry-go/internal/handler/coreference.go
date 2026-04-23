package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// CoreferenceHandler exposes the coreference resolution endpoint.
type CoreferenceHandler struct {
	svc *service.CoreferenceService
}

// NewCoreferenceHandler wires the service.
func NewCoreferenceHandler(svc *service.CoreferenceService) *CoreferenceHandler {
	return &CoreferenceHandler{svc: svc}
}

// Resolve handles POST /v1/extract/resolve-coreferences
// Body: { "content": "..." }
func (h *CoreferenceHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "coreference service not available")
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	res, err := h.svc.Resolve(r.Context(), req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
