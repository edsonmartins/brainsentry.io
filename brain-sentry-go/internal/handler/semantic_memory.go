package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/integraltech/brainsentry/internal/service"
)

type SemanticMemoryHandler struct {
	svc *service.SemanticMemoryService
}

func NewSemanticMemoryHandler(svc *service.SemanticMemoryService) *SemanticMemoryHandler {
	return &SemanticMemoryHandler{svc: svc}
}

func (h *SemanticMemoryHandler) Consolidate(w http.ResponseWriter, r *http.Request) {
	minMemories := 5
	if v := r.URL.Query().Get("minMemories"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			minMemories = n
		}
	}

	result, err := h.svc.Consolidate(r.Context(), minMemories)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
