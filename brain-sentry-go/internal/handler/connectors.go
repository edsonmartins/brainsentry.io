package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/integraltech/brainsentry/internal/service"
)

// ConnectorsHandler handles connector endpoints.
type ConnectorsHandler struct {
	connectorService *service.ConnectorService
	registry         *service.ConnectorRegistry
}

// NewConnectorsHandler creates a new ConnectorsHandler.
func NewConnectorsHandler(connectorService *service.ConnectorService, registry *service.ConnectorRegistry) *ConnectorsHandler {
	return &ConnectorsHandler{
		connectorService: connectorService,
		registry:         registry,
	}
}

// List handles GET /v1/connectors — lists all registered connectors.
func (h *ConnectorsHandler) List(w http.ResponseWriter, r *http.Request) {
	connectors := h.registry.List()
	names := make([]string, 0, len(connectors))
	for name := range connectors {
		names = append(names, name)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connectors": names,
	})
}

// Sync handles POST /v1/connectors/{name}/sync — triggers sync for a specific connector.
func (h *ConnectorsHandler) Sync(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "connector name is required")
		return
	}

	sinceStr := r.URL.Query().Get("since")
	var since *time.Time
	if sinceStr != "" {
		t, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid since format, use RFC3339")
			return
		}
		since = &t
	}

	result, err := h.connectorService.SyncConnector(r.Context(), name, since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sync failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SyncAll handles POST /v1/connectors/sync-all — syncs all connectors.
func (h *ConnectorsHandler) SyncAll(w http.ResponseWriter, r *http.Request) {
	results := h.connectorService.SyncAll(r.Context(), nil)
	writeJSON(w, http.StatusOK, results)
}
