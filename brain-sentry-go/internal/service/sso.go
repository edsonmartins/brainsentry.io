package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// SSOProvider represents a supported SSO provider type.
type SSOProvider string

const (
	SSOProviderOIDC SSOProvider = "OIDC"
	SSOProviderSAML SSOProvider = "SAML"
)

// SSOConfig holds SSO provider configuration.
type SSOConfig struct {
	Provider     SSOProvider `json:"provider" yaml:"provider"`
	Enabled      bool        `json:"enabled" yaml:"enabled"`
	ClientID     string      `json:"clientId" yaml:"clientId"`
	ClientSecret string      `json:"clientSecret" yaml:"clientSecret"`
	IssuerURL    string      `json:"issuerUrl" yaml:"issuerUrl"`
	RedirectURL  string      `json:"redirectUrl" yaml:"redirectUrl"`
	Scopes       []string    `json:"scopes" yaml:"scopes"`
	// SAML-specific
	MetadataURL    string `json:"metadataUrl,omitempty" yaml:"metadataUrl"`
	EntityID       string `json:"entityId,omitempty" yaml:"entityId"`
	CertificatePEM string `json:"certificatePem,omitempty" yaml:"certificatePem"`
}

// SSOService handles SSO authentication via OIDC/SAML.
type SSOService struct {
	configs    map[string]*SSOConfig // tenantID -> config
	jwtService *JWTService
	client     *http.Client
}

// NewSSOService creates a new SSOService.
func NewSSOService(jwtService *JWTService) *SSOService {
	return &SSOService{
		configs:    make(map[string]*SSOConfig),
		jwtService: jwtService,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// ConfigureTenant sets the SSO configuration for a tenant.
func (s *SSOService) ConfigureTenant(tenantID string, config *SSOConfig) {
	s.configs[tenantID] = config
}

// GetConfig returns the SSO config for a tenant.
func (s *SSOService) GetConfig(tenantID string) *SSOConfig {
	return s.configs[tenantID]
}

// IsEnabled checks if SSO is enabled for a tenant.
func (s *SSOService) IsEnabled(tenantID string) bool {
	cfg := s.configs[tenantID]
	return cfg != nil && cfg.Enabled
}

// SSOLoginResponse represents the response from SSO login.
type SSOLoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	ExternalID   string `json:"externalId"`
}

// GetAuthorizationURL returns the SSO authorization URL for a tenant.
func (s *SSOService) GetAuthorizationURL(tenantID, state string) (string, error) {
	cfg := s.configs[tenantID]
	if cfg == nil || !cfg.Enabled {
		return "", fmt.Errorf("SSO not configured for tenant %s", tenantID)
	}

	switch cfg.Provider {
	case SSOProviderOIDC:
		return s.getOIDCAuthURL(cfg, state)
	case SSOProviderSAML:
		return "", fmt.Errorf("SAML SSO requires SP-initiated login via /sso/saml/login")
	default:
		return "", fmt.Errorf("unsupported SSO provider: %s", cfg.Provider)
	}
}

// HandleOIDCCallback processes the OIDC callback with authorization code.
func (s *SSOService) HandleOIDCCallback(ctx context.Context, tenantID, code string) (*SSOLoginResponse, error) {
	cfg := s.configs[tenantID]
	if cfg == nil || !cfg.Enabled || cfg.Provider != SSOProviderOIDC {
		return nil, fmt.Errorf("OIDC not configured for tenant %s", tenantID)
	}

	// Exchange code for tokens
	tokenResp, err := s.exchangeOIDCCode(ctx, cfg, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	// Get user info
	userInfo, err := s.getOIDCUserInfo(ctx, cfg, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("user info failed: %w", err)
	}

	// Generate internal JWT
	token, err := s.jwtService.GenerateToken(userInfo.Sub, userInfo.Email, tenantID, []string{"USER"})
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(userInfo.Sub, userInfo.Email, tenantID, []string{"USER"})
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &SSOLoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		Provider:     string(SSOProviderOIDC),
		ExternalID:   userInfo.Sub,
	}, nil
}

func (s *SSOService) getOIDCAuthURL(cfg *SSOConfig, state string) (string, error) {
	scopes := "openid email profile"
	if len(cfg.Scopes) > 0 {
		scopes = ""
		for i, scope := range cfg.Scopes {
			if i > 0 {
				scopes += " "
			}
			scopes += scope
		}
	}

	url := fmt.Sprintf("%s/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		cfg.IssuerURL, cfg.ClientID, cfg.RedirectURL, scopes, state)

	return url, nil
}

type oidcTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

func (s *SSOService) exchangeOIDCCode(ctx context.Context, cfg *SSOConfig, code string) (*oidcTokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  cfg.RedirectURL,
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.IssuerURL+"/token", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp oidcTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	return &tokenResp, nil
}

type oidcUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *SSOService) getOIDCUserInfo(ctx context.Context, cfg *SSOConfig, accessToken string) (*oidcUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.IssuerURL+"/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("userinfo endpoint failed", "status", resp.StatusCode)
		return nil, fmt.Errorf("userinfo returned %d", resp.StatusCode)
	}

	var userInfo oidcUserInfo
	if err := json.Unmarshal(respBody, &userInfo); err != nil {
		return nil, fmt.Errorf("parsing userinfo: %w", err)
	}

	return &userInfo, nil
}
