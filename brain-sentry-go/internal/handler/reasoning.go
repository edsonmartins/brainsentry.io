package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// ReasoningHandler groups inference engines (abductive, deductive, ...).
type ReasoningHandler struct {
	abductive *service.AbductiveReasoner
}

// NewReasoningHandler wires engines.
func NewReasoningHandler(abductive *service.AbductiveReasoner) *ReasoningHandler {
	return &ReasoningHandler{abductive: abductive}
}

// Abduce handles POST /v1/reasoning/abduce
func (h *ReasoningHandler) Abduce(w http.ResponseWriter, r *http.Request) {
	if h.abductive == nil {
		writeError(w, http.StatusServiceUnavailable, "abductive reasoner not available")
		return
	}
	var req service.AbduceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.abductive.Abduce(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
