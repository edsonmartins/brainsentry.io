package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

type OntologyHandler struct {
	svc *service.OntologyService
}

func NewOntologyHandler(svc *service.OntologyService) *OntologyHandler {
	return &OntologyHandler{svc: svc}
}

// Get handles GET /v1/ontology — returns the current ontology.
func (h *OntologyHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil || !h.svc.IsEnabled() {
		writeError(w, http.StatusServiceUnavailable, "no ontology is loaded")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.svc.Ontology())
}

// Set handles PUT /v1/ontology — replaces the in-memory ontology.
// Body: full Ontology JSON.
func (h *OntologyHandler) Set(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "ontology service is not available")
		return
	}

	var ont service.Ontology
	if err := json.NewDecoder(r.Body).Decode(&ont); err != nil {
		writeError(w, http.StatusBadRequest, "invalid ontology body")
		return
	}
	if err := h.svc.SetOntology(&ont); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":            "loaded",
		"name":              ont.Name,
		"version":           ont.Version,
		"entityTypes":       len(ont.EntityTypes),
		"entities":          len(ont.Entities),
		"relationshipTypes": len(ont.Relationships),
	})
}

// Resolve handles POST /v1/ontology/resolve
// Body: {"name": "postgres"}
// Returns canonical name + type if matched.
func (h *OntologyHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil || !h.svc.IsEnabled() {
		writeError(w, http.StatusServiceUnavailable, "no ontology is loaded")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	canonical, entityType, ok := h.svc.ResolveEntity(req.Name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"input":     req.Name,
		"matched":   ok,
		"canonical": canonical,
		"type":      entityType,
	})
}
