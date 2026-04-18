package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/internal/service"
)

type NodeSetHandler struct {
	svc        *service.NodeSetService
	memoryRepo *postgres.MemoryRepository
}

func NewNodeSetHandler(svc *service.NodeSetService, memoryRepo *postgres.MemoryRepository) *NodeSetHandler {
	return &NodeSetHandler{svc: svc, memoryRepo: memoryRepo}
}

// AddToSet handles POST /v1/memories/{id}/sets
// Body: {"sets": ["set1", "set2"]}
func (h *NodeSetHandler) AddToSet(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "node set service is not available")
		return
	}

	id := chi.URLParam(r, "id")
	var req struct {
		Sets []string `json:"sets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.AddToSet(r.Context(), id, req.Sets...); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return current sets
	memory, err := h.memoryRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "memory not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"memoryId": id,
		"sets":     h.svc.GetMemorySets(memory),
	})
}

// RemoveFromSet handles DELETE /v1/memories/{id}/sets
// Body: {"sets": ["set1"]}
func (h *NodeSetHandler) RemoveFromSet(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "node set service is not available")
		return
	}

	id := chi.URLParam(r, "id")
	var req struct {
		Sets []string `json:"sets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.RemoveFromSet(r.Context(), id, req.Sets...); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	memory, err := h.memoryRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "memory not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"memoryId": id,
		"sets":     h.svc.GetMemorySets(memory),
	})
}

// GetSets handles GET /v1/memories/{id}/sets
func (h *NodeSetHandler) GetSets(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "node set service is not available")
		return
	}

	id := chi.URLParam(r, "id")
	memory, err := h.memoryRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "memory not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"memoryId": id,
		"sets":     h.svc.GetMemorySets(memory),
	})
}
