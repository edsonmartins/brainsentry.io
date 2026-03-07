package handler

import (
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// ReflectionHandler handles reflection endpoints.
type ReflectionHandler struct {
	reflectionService *service.ReflectionService
}

// NewReflectionHandler creates a new ReflectionHandler.
func NewReflectionHandler(reflectionService *service.ReflectionService) *ReflectionHandler {
	return &ReflectionHandler{reflectionService: reflectionService}
}

// RunReflection handles POST /v1/reflect — performs a reflection cycle.
func (h *ReflectionHandler) RunReflection(w http.ResponseWriter, r *http.Request) {
	result, err := h.reflectionService.RunReflection(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reflection failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
