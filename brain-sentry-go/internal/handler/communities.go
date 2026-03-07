package handler

import (
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// CommunitiesHandler handles community detection endpoints.
type CommunitiesHandler struct {
	louvainService *service.LouvainService
}

// NewCommunitiesHandler creates a new CommunitiesHandler.
func NewCommunitiesHandler(louvainService *service.LouvainService) *CommunitiesHandler {
	return &CommunitiesHandler{louvainService: louvainService}
}

// DetectCommunities handles GET /v1/graph/communities — runs Louvain community detection.
func (h *CommunitiesHandler) DetectCommunities(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())

	result, err := h.louvainService.DetectCommunities(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "community detection failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
