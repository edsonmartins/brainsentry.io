package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/service"
)

// CorrectionHandler handles memory correction endpoints.
type CorrectionHandler struct {
	correctionService *service.CorrectionService
}

// NewCorrectionHandler creates a new CorrectionHandler.
func NewCorrectionHandler(correctionService *service.CorrectionService) *CorrectionHandler {
	return &CorrectionHandler{correctionService: correctionService}
}

// Flag handles POST /v1/memories/{id}/flag
func (h *CorrectionHandler) Flag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.FlagMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	correction, err := h.correctionService.FlagMemory(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to flag memory")
		return
	}

	writeJSON(w, http.StatusOK, correction)
}

// Review handles POST /v1/memories/{id}/review
func (h *CorrectionHandler) Review(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.ReviewCorrectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		writeError(w, http.StatusBadRequest, "action must be 'approve' or 'reject'")
		return
	}

	m, err := h.correctionService.ReviewCorrection(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to review correction")
		return
	}

	writeJSON(w, http.StatusOK, m)
}

// Rollback handles POST /v1/memories/{id}/rollback
func (h *CorrectionHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		TargetVersion int `json:"targetVersion"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TargetVersion <= 0 {
		// Try query param
		if v := r.URL.Query().Get("version"); v != "" {
			req.TargetVersion, _ = strconv.Atoi(v)
		}
	}

	if req.TargetVersion <= 0 {
		writeError(w, http.StatusBadRequest, "targetVersion is required")
		return
	}

	m, err := h.correctionService.RollbackMemory(r.Context(), id, req.TargetVersion)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, m)
}
