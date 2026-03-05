package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// BatchHandler handles batch import/export endpoints.
type BatchHandler struct {
	batchService *service.BatchService
}

// NewBatchHandler creates a new BatchHandler.
func NewBatchHandler(batchService *service.BatchService) *BatchHandler {
	return &BatchHandler{batchService: batchService}
}

// Import handles POST /v1/batch/import
func (h *BatchHandler) Import(w http.ResponseWriter, r *http.Request) {
	var req service.BatchImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Memories) == 0 {
		writeError(w, http.StatusBadRequest, "memories array is required")
		return
	}

	result, err := h.batchService.Import(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Export handles GET /v1/batch/export
func (h *BatchHandler) Export(w http.ResponseWriter, r *http.Request) {
	result, err := h.batchService.Export(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to export memories")
		return
	}

	writeJSON(w, http.StatusOK, result)
}
