package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/service"
)

type AgentTraceHandler struct {
	svc *service.AgentTraceService
}

func NewAgentTraceHandler(svc *service.AgentTraceService) *AgentTraceHandler {
	return &AgentTraceHandler{svc: svc}
}

// Record handles POST /v1/traces
// Body: RecordTraceRequest fields.
func (h *AgentTraceHandler) Record(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "agent trace service is not available")
		return
	}

	var req service.RecordTraceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	trace, err := h.svc.Record(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trace)
}

// List handles GET /v1/traces
// Query params: sessionId, agentId, status, set, limit.
func (h *AgentTraceHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "agent trace service is not available")
		return
	}

	filter := service.ListFilter{
		SessionID: r.URL.Query().Get("sessionId"),
		AgentID:   r.URL.Query().Get("agentId"),
		Status:    domain.AgentTraceStatus(r.URL.Query().Get("status")),
		Set:       r.URL.Query().Get("set"),
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			filter.Limit = n
		}
	}

	traces := h.svc.List(r.Context(), filter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"count":  len(traces),
		"traces": traces,
	})
}

// Get handles GET /v1/traces/{id}
func (h *AgentTraceHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "agent trace service is not available")
		return
	}

	id := chi.URLParam(r, "id")
	trace, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trace)
}

// Delete handles DELETE /v1/traces/{id}
func (h *AgentTraceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "agent trace service is not available")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Stats handles GET /v1/traces/stats
func (h *AgentTraceHandler) Stats(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "agent trace service is not available")
		return
	}

	stats := h.svc.Stats(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
