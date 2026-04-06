package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

type MeshHandler struct {
	svc *service.MeshSyncService
}

func NewMeshHandler(svc *service.MeshSyncService) *MeshHandler {
	return &MeshHandler{svc: svc}
}

func (h *MeshHandler) RegisterPeer(w http.ResponseWriter, r *http.Request) {
	var peer service.MeshPeer
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.RegisterPeer(r.Context(), peer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func (h *MeshHandler) ListPeers(w http.ResponseWriter, r *http.Request) {
	peers := h.svc.ListPeers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

func (h *MeshHandler) Sync(w http.ResponseWriter, r *http.Request) {
	var payload service.SyncPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	results := h.svc.SyncWithAllPeers(r.Context(), payload.Scope, payload.Items)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
