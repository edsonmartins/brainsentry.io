package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/internal/service"
)

type FeedbackLearningHandler struct {
	svc        *service.FeedbackLearningService
	memoryRepo *postgres.MemoryRepository
}

func NewFeedbackLearningHandler(svc *service.FeedbackLearningService, memoryRepo *postgres.MemoryRepository) *FeedbackLearningHandler {
	return &FeedbackLearningHandler{svc: svc, memoryRepo: memoryRepo}
}

// GetWeight handles GET /v1/memories/{id}/feedback-weight
// Returns the current feedback-derived weight for a memory.
func (h *FeedbackLearningHandler) GetWeight(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "feedback learning is not available")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "memory id is required")
		return
	}

	memory, err := h.memoryRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "memory not found")
		return
	}

	weight := h.svc.ComputeWeight(memory)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"memoryId":        id,
		"helpfulCount":    memory.HelpfulCount,
		"notHelpfulCount": memory.NotHelpfulCount,
		"feedbackWeight":  weight,
		"alpha":           h.svc.Config().Alpha,
	})
}
