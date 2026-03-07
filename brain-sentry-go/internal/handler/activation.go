package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// ActivationHandler handles spreading activation endpoints.
type ActivationHandler struct {
	activationService *service.SpreadingActivationService
}

// NewActivationHandler creates a new ActivationHandler.
func NewActivationHandler(activationService *service.SpreadingActivationService) *ActivationHandler {
	return &ActivationHandler{activationService: activationService}
}

// Activate handles POST /v1/memories/activate — propagates activation from seed memories.
func (h *ActivationHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SeedIDs         []string  `json:"seedIds"`
		SeedActivations []float64 `json:"seedActivations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.SeedIDs) == 0 {
		writeError(w, http.StatusBadRequest, "seedIds is required")
		return
	}

	// Default activations to 1.0 if not provided
	if len(req.SeedActivations) == 0 {
		req.SeedActivations = make([]float64, len(req.SeedIDs))
		for i := range req.SeedActivations {
			req.SeedActivations[i] = 1.0
		}
	}

	tenantID := tenant.FromContext(r.Context())

	result, err := h.activationService.Spread(r.Context(), req.SeedIDs, req.SeedActivations, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "activation failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
