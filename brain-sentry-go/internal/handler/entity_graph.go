package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/service"
)

// EntityGraphHandler handles entity graph endpoints.
type EntityGraphHandler struct {
	entityGraphService *service.EntityGraphService
	memoryService      *service.MemoryService
}

// NewEntityGraphHandler creates a new EntityGraphHandler.
func NewEntityGraphHandler(entityGraphService *service.EntityGraphService, memoryService *service.MemoryService) *EntityGraphHandler {
	return &EntityGraphHandler{
		entityGraphService: entityGraphService,
		memoryService:      memoryService,
	}
}

// GetEntitiesByMemory handles GET /v1/entity-graph/memory/{memoryId}/entities
func (h *EntityGraphHandler) GetEntitiesByMemory(w http.ResponseWriter, r *http.Request) {
	memoryID := chi.URLParam(r, "memoryId")
	entities, err := h.entityGraphService.FindEntitiesByMemory(r.Context(), memoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get entities")
		return
	}
	writeJSON(w, http.StatusOK, entities)
}

// GetRelationshipsByMemory handles GET /v1/entity-graph/memory/{memoryId}/relationships
func (h *EntityGraphHandler) GetRelationshipsByMemory(w http.ResponseWriter, r *http.Request) {
	memoryID := chi.URLParam(r, "memoryId")
	rels, err := h.entityGraphService.FindRelationshipsByMemory(r.Context(), memoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get relationships")
		return
	}
	writeJSON(w, http.StatusOK, rels)
}

// SearchEntities handles GET /v1/entity-graph/search
func (h *EntityGraphHandler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	entities, err := h.entityGraphService.SearchEntities(r.Context(), query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	writeJSON(w, http.StatusOK, entities)
}

// GetKnowledgeGraph handles GET /v1/entity-graph/knowledge-graph
func (h *EntityGraphHandler) GetKnowledgeGraph(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	graph, err := h.entityGraphService.GetKnowledgeGraph(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get knowledge graph")
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

// ExtractEntities handles POST /v1/entity-graph/extract/{memoryId}
func (h *EntityGraphHandler) ExtractEntities(w http.ResponseWriter, r *http.Request) {
	memoryID := chi.URLParam(r, "memoryId")

	m, err := h.memoryService.GetMemory(r.Context(), memoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, "memory not found")
		return
	}

	if err := h.entityGraphService.ExtractAndStoreEntities(r.Context(), m); err != nil {
		writeError(w, http.StatusInternalServerError, "entity extraction failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "entities extracted successfully"})
}

// BatchExtract handles POST /v1/entity-graph/extract-batch
func (h *EntityGraphHandler) BatchExtract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MemoryIDs []string `json:"memoryIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	processed := 0
	for _, id := range req.MemoryIDs {
		m, err := h.memoryService.GetMemory(r.Context(), id)
		if err != nil {
			continue
		}
		if err := h.entityGraphService.ExtractAndStoreEntities(r.Context(), m); err == nil {
			processed++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "batch extraction completed",
		"processed": processed,
		"total":     len(req.MemoryIDs),
	})
}
