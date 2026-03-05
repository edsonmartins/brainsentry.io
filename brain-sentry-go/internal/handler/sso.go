package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// SSOHandler handles SSO authentication endpoints.
type SSOHandler struct {
	ssoService *service.SSOService
}

// NewSSOHandler creates a new SSOHandler.
func NewSSOHandler(ssoService *service.SSOService) *SSOHandler {
	return &SSOHandler{ssoService: ssoService}
}

// GetAuthURL handles GET /v1/auth/sso/authorize
func (h *SSOHandler) GetAuthURL(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())
	state := r.URL.Query().Get("state")
	if state == "" {
		state = "default"
	}

	url, err := h.ssoService.GetAuthorizationURL(tenantID, state)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// Callback handles POST /v1/auth/sso/callback
func (h *SSOHandler) Callback(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())

	var req struct {
		Code  string `json:"code"`
		State string `json:"state,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "authorization code is required")
		return
	}

	resp, err := h.ssoService.HandleOIDCCallback(r.Context(), tenantID, req.Code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetConfig handles GET /v1/auth/sso/config
func (h *SSOHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())
	cfg := h.ssoService.GetConfig(tenantID)
	if cfg == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"enabled":  false,
			"provider": nil,
		})
		return
	}

	// Don't expose secrets
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":  cfg.Enabled,
		"provider": cfg.Provider,
		"issuerUrl": cfg.IssuerURL,
		"scopes":    cfg.Scopes,
	})
}
