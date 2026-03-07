package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/service"
	"github.com/integraltech/brainsentry/internal/middleware"
)

// ProfileHandler handles profile endpoints.
type ProfileHandler struct {
	profileService *service.ProfileService
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// GetProfile handles GET /v1/profile — generates profile for the authenticated user.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profile, err := h.profileService.GenerateProfile(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate profile: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// GetProfileByUser handles GET /v1/profile/{userId} — generates profile for a specific user.
func (h *ProfileHandler) GetProfileByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "userId is required")
		return
	}

	profile, err := h.profileService.GenerateProfile(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate profile: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, profile)
}
