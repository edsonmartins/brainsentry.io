package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/service"
)

// InterceptionHandler handles prompt interception endpoints.
type InterceptionHandler struct {
	interceptionService *service.InterceptionService
}

// NewInterceptionHandler creates a new InterceptionHandler.
func NewInterceptionHandler(interceptionService *service.InterceptionService) *InterceptionHandler {
	return &InterceptionHandler{interceptionService: interceptionService}
}

// Intercept handles POST /v1/intercept
//
//	@Summary		Intercept and enhance prompt with context
//	@Description	Analyzes a prompt, searches for relevant memories and notes, and injects context
//	@Tags			Interception
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.InterceptRequest	true	"Interception request"
//	@Success		200		{object}	dto.InterceptResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/intercept [post]
func (h *InterceptionHandler) Intercept(w http.ResponseWriter, r *http.Request) {
	var req dto.InterceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	resp, err := h.interceptionService.Intercept(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "interception failed")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
