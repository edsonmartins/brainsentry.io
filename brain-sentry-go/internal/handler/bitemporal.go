package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

// BiTemporalHandler exposes "as of" time-travel queries over Memory.
type BiTemporalHandler struct {
	repo *postgres.MemoryRepository
}

// NewBiTemporalHandler wires the repository.
func NewBiTemporalHandler(repo *postgres.MemoryRepository) *BiTemporalHandler {
	return &BiTemporalHandler{repo: repo}
}

// AsOf handles GET /v1/memories/as-of?at=<RFC3339>&limit=N
func (h *BiTemporalHandler) AsOf(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "memory repository not available")
		return
	}
	v := r.URL.Query().Get("at")
	if v == "" {
		writeError(w, http.StatusBadRequest, "at (RFC3339 timestamp) is required")
		return
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid RFC3339 timestamp")
		return
	}
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	list, err := h.repo.FindAsOf(r.Context(), t, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"count":    len(list),
		"asOf":     t.UTC().Format(time.RFC3339),
		"memories": list,
	})
}
