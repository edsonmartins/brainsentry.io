package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// NLQueryHandler handles natural language query endpoints.
type NLQueryHandler struct {
	nlCypherService *service.NLCypherService
}

// NewNLQueryHandler creates a new NLQueryHandler.
func NewNLQueryHandler(nlCypherService *service.NLCypherService) *NLQueryHandler {
	return &NLQueryHandler{nlCypherService: nlCypherService}
}

// Query handles POST /v1/graph/nl-query — translates NL to Cypher and queries the graph.
func (h *NLQueryHandler) Query(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Question string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	result, err := h.nlCypherService.QueryNaturalLanguage(r.Context(), req.Question)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
