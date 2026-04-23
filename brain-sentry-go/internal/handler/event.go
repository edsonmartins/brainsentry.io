package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/internal/service"
)

// EventHandler exposes Event endpoints.
type EventHandler struct {
	svc *service.EventService
}

// NewEventHandler builds the handler.
func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

// Record handles POST /v1/events
func (h *EventHandler) Record(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "event service not available")
		return
	}
	var req service.RecordEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	e, err := h.svc.Record(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(e)
}

// Get handles GET /v1/events/{id}
func (h *EventHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "event service not available")
		return
	}
	id := chi.URLParam(r, "id")
	e, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(e)
}

// List handles GET /v1/events
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "event service not available")
		return
	}
	q := r.URL.Query()
	f := postgres.EventFilter{
		EventType: q.Get("eventType"),
		EntityID:  q.Get("entityId"),
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.FromTime = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.ToTime = &t
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count":  len(list),
		"events": list,
	})
}

// Delete handles DELETE /v1/events/{id}
func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "event service not available")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Stats handles GET /v1/events/stats
func (h *EventHandler) Stats(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "event service not available")
		return
	}
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

// Extract handles POST /v1/events/extract
// Body: { content, sourceMemoryId }
func (h *EventHandler) Extract(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "event service not available")
		return
	}
	var req struct {
		Content        string `json:"content"`
		SourceMemoryID string `json:"sourceMemoryId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	events, err := h.svc.ExtractFromText(r.Context(), req.Content, req.SourceMemoryID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count":  len(events),
		"events": events,
	})
}
