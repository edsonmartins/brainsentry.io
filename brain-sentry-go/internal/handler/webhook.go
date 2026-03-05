package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/service"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// WebhookHandler handles webhook management endpoints.
type WebhookHandler struct {
	webhookService *service.WebhookService
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

// Register handles POST /v1/webhooks
func (h *WebhookHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL    string                     `json:"url"`
		Secret string                     `json:"secret,omitempty"`
		Events []domain.WebhookEventType  `json:"events,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	tenantID := tenant.FromContext(r.Context())
	webhook := h.webhookService.Register(r.Context(), tenantID, req.URL, req.Secret, req.Events)
	writeJSON(w, http.StatusCreated, webhook)
}

// Unregister handles DELETE /v1/webhooks/{id}
func (h *WebhookHandler) Unregister(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.webhookService.Unregister(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "webhook removed"})
}

// List handles GET /v1/webhooks
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())
	webhooks := h.webhookService.ListWebhooks(r.Context(), tenantID)
	writeJSON(w, http.StatusOK, webhooks)
}

// Deliveries handles GET /v1/webhooks/{id}/deliveries
func (h *WebhookHandler) Deliveries(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	deliveries := h.webhookService.GetDeliveries(r.Context(), id, 20)
	writeJSON(w, http.StatusOK, deliveries)
}
