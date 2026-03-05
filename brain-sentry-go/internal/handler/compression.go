package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/service"
)

// CompressionHandler handles context compression endpoints.
type CompressionHandler struct {
	compressionService *service.CompressionService
}

// NewCompressionHandler creates a new CompressionHandler.
func NewCompressionHandler(compressionService *service.CompressionService) *CompressionHandler {
	return &CompressionHandler{compressionService: compressionService}
}

// Compress handles POST /v1/compression/compress
func (h *CompressionHandler) Compress(w http.ResponseWriter, r *http.Request) {
	var req struct {
		dto.CompressionRequest
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "messages are required")
		return
	}

	result, err := h.compressionService.Compress(r.Context(), req.CompressionRequest, req.SessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "compression failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetSessionSummaries handles GET /v1/compression/session/{sessionId}
func (h *CompressionHandler) GetSessionSummaries(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	summaries, err := h.compressionService.GetSessionSummaries(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get summaries")
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

// GetLatestSummary handles GET /v1/compression/session/{sessionId}/latest
func (h *CompressionHandler) GetLatestSummary(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	summary, err := h.compressionService.GetLatestSummary(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no summary found")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
