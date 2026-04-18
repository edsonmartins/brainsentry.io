package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

type TripletHandler struct {
	svc *service.TripletExtractionService
}

func NewTripletHandler(svc *service.TripletExtractionService) *TripletHandler {
	return &TripletHandler{svc: svc}
}

// Extract handles POST /v1/triplets/extract
// Body: {"content": "...", "memoryId": "optional-id"}
// Returns: extracted triplets.
func (h *TripletHandler) Extract(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "triplet extraction is not available (LLM disabled)")
		return
	}

	var req struct {
		Content  string `json:"content"`
		MemoryID string `json:"memoryId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	memoryID := req.MemoryID
	if memoryID == "" {
		memoryID = "inline"
	}

	triplets, err := h.svc.ExtractAndBuild(r.Context(), memoryID, req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"memoryId": memoryID,
		"count":    len(triplets),
		"triplets": triplets,
	})
}
