package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/service"
)

// ConflictHandler handles conflict detection endpoints.
type ConflictHandler struct {
	conflictService *service.ConflictService
}

// NewConflictHandler creates a new ConflictHandler.
func NewConflictHandler(conflictService *service.ConflictService) *ConflictHandler {
	return &ConflictHandler{conflictService: conflictService}
}

// DetectForMemory handles POST /v1/conflicts/detect/{memoryId}
func (h *ConflictHandler) DetectForMemory(w http.ResponseWriter, r *http.Request) {
	memoryID := chi.URLParam(r, "memoryId")
	if memoryID == "" {
		writeError(w, http.StatusBadRequest, "memoryId is required")
		return
	}

	conflicts, err := h.conflictService.DetectConflicts(r.Context(), memoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, conflicts)
}

// ScanAll handles POST /v1/conflicts/scan
func (h *ConflictHandler) ScanAll(w http.ResponseWriter, r *http.Request) {
	conflicts, err := h.conflictService.ScanAllConflicts(r.Context(), 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, conflicts)
}
