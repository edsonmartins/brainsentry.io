package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles POST /v1/auth/login
//
//	@Summary	Authenticate user
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		dto.LoginRequest	true	"Login credentials"
//	@Success	200		{object}	dto.LoginResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Router		/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, service.ErrAccountDisabled):
			writeError(w, http.StatusForbidden, "account is disabled")
		default:
			writeError(w, http.StatusInternalServerError, "authentication failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// DemoLogin handles POST /v1/auth/demo — creates or retrieves the demo user and returns a token.
func (h *AuthHandler) DemoLogin(w http.ResponseWriter, r *http.Request) {
	resp, err := h.authService.DemoLogin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "demo login failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Logout handles POST /v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Stateless JWT - client discards token
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// Refresh handles POST /v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refreshToken is required")
		return
	}

	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
