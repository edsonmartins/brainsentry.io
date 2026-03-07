package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/service"
)

// CrossSessionHandler handles cross-session endpoints.
type CrossSessionHandler struct {
	crossSessionService *service.CrossSessionService
}

// NewCrossSessionHandler creates a new CrossSessionHandler.
func NewCrossSessionHandler(crossSessionService *service.CrossSessionService) *CrossSessionHandler {
	return &CrossSessionHandler{crossSessionService: crossSessionService}
}

// GetSessionEvents handles GET /v1/sessions/{id}/events — returns events for a session.
func (h *CrossSessionHandler) GetSessionEvents(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session id is required")
		return
	}

	events := h.crossSessionService.GetSessionEvents(sessionID)
	writeJSON(w, http.StatusOK, events)
}

// GetCrossContext handles GET /v1/sessions/{id}/cross-context — returns cross-session context.
func (h *CrossSessionHandler) GetCrossContext(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session id is required")
		return
	}

	result, err := h.crossSessionService.OnSessionStart(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cross-context retrieval failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
