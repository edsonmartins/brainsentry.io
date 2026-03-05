package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/service"
)

// SessionHandler handles session lifecycle endpoints.
type SessionHandler struct {
	sessionService *service.SessionService
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

// Create handles POST /v1/sessions
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	session := h.sessionService.CreateSession(r.Context(), req.UserID)
	writeJSON(w, http.StatusCreated, session)
}

// Get handles GET /v1/sessions/{id}
func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	session, err := h.sessionService.GetSession(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// Touch handles POST /v1/sessions/{id}/touch
func (h *SessionHandler) Touch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.sessionService.TouchSession(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "session touched"})
}

// End handles POST /v1/sessions/{id}/end
func (h *SessionHandler) End(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.sessionService.EndSession(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "session ended"})
}

// ListActive handles GET /v1/sessions/active
func (h *SessionHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	sessions := h.sessionService.ListActiveSessions(r.Context())
	writeJSON(w, http.StatusOK, sessions)
}
