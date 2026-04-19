package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/service"
)

type SessionCacheHandler struct {
	svc *service.SessionMemoryCache
}

func NewSessionCacheHandler(svc *service.SessionMemoryCache) *SessionCacheHandler {
	return &SessionCacheHandler{svc: svc}
}

// Push handles POST /v1/session-cache/{sessionId}
// Body: SessionInteraction (query, response, memoryIds, metadata).
func (h *SessionCacheHandler) Push(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "session cache is not available")
		return
	}

	sessionID := chi.URLParam(r, "sessionId")
	var it service.SessionInteraction
	if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.svc.Push(r.Context(), sessionID, it); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// List handles GET /v1/session-cache/{sessionId}?limit=10
func (h *SessionCacheHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "session cache is not available")
		return
	}

	sessionID := chi.URLParam(r, "sessionId")
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}

	items, err := h.svc.Recent(r.Context(), sessionID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"sessionId":    sessionID,
		"count":        len(items),
		"interactions": items,
	})
}

// Clear handles DELETE /v1/session-cache/{sessionId}
func (h *SessionCacheHandler) Clear(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "session cache is not available")
		return
	}
	sessionID := chi.URLParam(r, "sessionId")
	if err := h.svc.Clear(r.Context(), sessionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Cognify handles POST /v1/session-cache/{sessionId}/cognify?clear=true
// Persists cached interactions into permanent memories.
func (h *SessionCacheHandler) Cognify(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "session cache is not available")
		return
	}
	sessionID := chi.URLParam(r, "sessionId")
	clearAfter := r.URL.Query().Get("clear") == "true"

	result, err := h.svc.Cognify(r.Context(), sessionID, clearAfter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListSessions handles GET /v1/session-cache — returns all active session IDs.
func (h *SessionCacheHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "session cache is not available")
		return
	}
	keys, err := h.svc.ListSessions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"count":    len(keys),
		"sessions": keys,
	})
}
