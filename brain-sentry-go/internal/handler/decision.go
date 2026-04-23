package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/internal/service"
)

// DecisionHandler exposes Decision endpoints.
type DecisionHandler struct {
	svc *service.DecisionService
}

// NewDecisionHandler constructs the handler.
func NewDecisionHandler(svc *service.DecisionService) *DecisionHandler {
	return &DecisionHandler{svc: svc}
}

// Record handles POST /v1/decisions
func (h *DecisionHandler) Record(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	var req service.RecordDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	d, err := h.svc.Record(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(d)
}

// Get handles GET /v1/decisions/{id}
func (h *DecisionHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	id := chi.URLParam(r, "id")
	d, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(d)
}

// List handles GET /v1/decisions
func (h *DecisionHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	q := r.URL.Query()
	f := postgres.DecisionFilter{
		Category:  q.Get("category"),
		AgentID:   q.Get("agentId"),
		SessionID: q.Get("sessionId"),
		Outcome:   domain.DecisionOutcome(q.Get("outcome")),
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Offset = n
		}
	}
	if v := q.Get("as_of"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.AsOf = &t
		}
	}
	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count":     len(list),
		"decisions": list,
	})
}

// Precedents handles GET /v1/decisions/{id}/precedents
func (h *DecisionHandler) Precedents(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	id := chi.URLParam(r, "id")
	limit := 5
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.svc.FindPrecedentsForDecision(r.Context(), id, limit)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count":      len(list),
		"precedents": list,
	})
}

// SearchPrecedents handles POST /v1/decisions/precedents
// Body: { category, scenario, limit }
func (h *DecisionHandler) SearchPrecedents(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	var req struct {
		Category string `json:"category"`
		Scenario string `json:"scenario"`
		Limit    int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Category == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}
	list, err := h.svc.FindPrecedentsForCategory(r.Context(), req.Category, req.Scenario, req.Limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count":      len(list),
		"precedents": list,
	})
}

// CausalChain handles GET /v1/decisions/{id}/causal-chain
func (h *DecisionHandler) CausalChain(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	id := chi.URLParam(r, "id")
	maxDepth := 5
	if v := r.URL.Query().Get("maxDepth"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxDepth = n
		}
	}
	chain, err := h.svc.CausalChain(r.Context(), id, maxDepth)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count": len(chain),
		"chain": chain,
	})
}

// Influence handles GET /v1/decisions/{id}/influence
func (h *DecisionHandler) Influence(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	id := chi.URLParam(r, "id")
	report, err := h.svc.AnalyzeInfluence(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

// Supersede handles POST /v1/decisions/{id}/supersede
// Body: { newId }
func (h *DecisionHandler) Supersede(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "decision service not available")
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		NewID string `json:"newId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewID == "" {
		writeError(w, http.StatusBadRequest, "newId is required")
		return
	}
	if err := h.svc.Supersede(r.Context(), id, req.NewID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
