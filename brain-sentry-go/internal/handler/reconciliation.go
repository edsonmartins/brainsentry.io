package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// ReconciliationHandler handles reconciliation endpoints.
type ReconciliationHandler struct {
	reconciliationService *service.ReconciliationService
}

// NewReconciliationHandler creates a new ReconciliationHandler.
func NewReconciliationHandler(reconciliationService *service.ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{reconciliationService: reconciliationService}
}

// Reconcile handles POST /v1/reconcile — extracts facts from content and reconciles with existing memories.
func (h *ReconciliationHandler) Reconcile(w http.ResponseWriter, r *http.Request) {
	if h.reconciliationService == nil {
		writeError(w, http.StatusServiceUnavailable, "reconciliation service is not available")
		return
	}

	var req struct {
		Content   string `json:"content"`
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	result, err := h.reconciliationService.ReconcileFacts(r.Context(), req.Content, req.SessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reconciliation failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
