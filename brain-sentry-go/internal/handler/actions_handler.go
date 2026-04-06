package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/service"
)

type ActionsHandler struct {
	svc *service.ActionService
}

func NewActionsHandler(svc *service.ActionService) *ActionsHandler {
	return &ActionsHandler{svc: svc}
}

func (h *ActionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		CreatedBy   string   `json:"createdBy"`
		Priority    int      `json:"priority"`
		Tags        []string `json:"tags"`
		ParentID    string   `json:"parentId"`
		DependsOn   []string `json:"dependsOn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	action, err := h.svc.CreateAction(r.Context(), req.Title, req.Description, req.CreatedBy, req.Priority, req.Tags, req.ParentID, req.DependsOn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(action)
}

func (h *ActionsHandler) List(w http.ResponseWriter, r *http.Request) {
	var statusFilter *service.ActionStatus
	if s := r.URL.Query().Get("status"); s != "" {
		status := service.ActionStatus(s)
		statusFilter = &status
	}

	actions := h.svc.ListActions(r.Context(), statusFilter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actions)
}

func (h *ActionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	action, err := h.svc.GetAction(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(action)
}

func (h *ActionsHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	action, err := h.svc.UpdateStatus(r.Context(), id, service.ActionStatus(req.Status))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(action)
}

func (h *ActionsHandler) AcquireLease(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		AgentID string `json:"agentId"`
		TTLMin  int    `json:"ttlMinutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(req.TTLMin) * time.Minute
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	lease, err := h.svc.AcquireLease(r.Context(), id, req.AgentID, ttl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lease)
}

func (h *ActionsHandler) ReleaseLease(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		AgentID   string `json:"agentId"`
		Completed bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.ReleaseLease(r.Context(), id, req.AgentID, req.Completed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
