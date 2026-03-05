package service

import (
	"context"
	"log/slog"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/repository/graph"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// EntityGraphService handles entity extraction and graph storage.
type EntityGraphService struct {
	entityGraphRepo *graph.EntityGraphRepository
	openRouter      *OpenRouterService
	auditService    *AuditService
}

// NewEntityGraphService creates a new EntityGraphService.
func NewEntityGraphService(
	entityGraphRepo *graph.EntityGraphRepository,
	openRouter *OpenRouterService,
	auditService *AuditService,
) *EntityGraphService {
	return &EntityGraphService{
		entityGraphRepo: entityGraphRepo,
		openRouter:      openRouter,
		auditService:    auditService,
	}
}

// ExtractAndStoreEntities extracts entities from memory content and stores them in the graph.
func (s *EntityGraphService) ExtractAndStoreEntities(ctx context.Context, m *domain.Memory) error {
	if s.openRouter == nil {
		return nil
	}

	tenantID := tenant.FromContext(ctx)

	result, err := s.openRouter.ExtractEntities(ctx, m.Content)
	if err != nil {
		slog.Warn("entity extraction failed", "error", err, "memoryId", m.ID)
		return err
	}

	if len(result.Entities) == 0 {
		return nil
	}

	// Store entity nodes
	nodeIDs := make(map[string]string) // entity name -> node ID
	for _, entity := range result.Entities {
		nodeID, err := s.entityGraphRepo.StoreEntity(ctx, entity.Name, entity.Type, tenantID, m.ID, entity.Properties)
		if err != nil {
			slog.Warn("failed to store entity", "error", err, "entity", entity.Name)
			continue
		}
		nodeIDs[entity.Name] = nodeID
	}

	// Store relationships
	for _, rel := range result.Relationships {
		sourceID, sourceOK := nodeIDs[rel.Source]
		targetID, targetOK := nodeIDs[rel.Target]
		if !sourceOK || !targetOK {
			continue
		}
		if err := s.entityGraphRepo.StoreRelationship(ctx, sourceID, targetID, rel.Type, tenantID, rel.Properties); err != nil {
			slog.Warn("failed to store relationship", "error", err, "source", rel.Source, "target", rel.Target)
		}
	}

	// Audit
	if s.auditService != nil {
		go s.auditService.LogEntityExtraction(
			tenant.WithTenant(context.Background(), tenantID),
			m.ID, len(result.Entities), len(result.Relationships),
		)
	}

	return nil
}

// FindEntitiesByMemory returns entities for a specific memory.
func (s *EntityGraphService) FindEntitiesByMemory(ctx context.Context, memoryID string) ([]dto.EntityNode, error) {
	tenantID := tenant.FromContext(ctx)
	return s.entityGraphRepo.FindEntitiesByMemory(ctx, memoryID, tenantID)
}

// FindRelationshipsByMemory returns relationships for a memory's entities.
func (s *EntityGraphService) FindRelationshipsByMemory(ctx context.Context, memoryID string) ([]dto.EntityEdge, error) {
	tenantID := tenant.FromContext(ctx)
	return s.entityGraphRepo.FindRelationshipsByMemory(ctx, memoryID, tenantID)
}

// SearchEntities searches for entities by name.
func (s *EntityGraphService) SearchEntities(ctx context.Context, searchTerm string, limit int) ([]dto.EntityNode, error) {
	tenantID := tenant.FromContext(ctx)
	return s.entityGraphRepo.SearchEntities(ctx, searchTerm, tenantID, limit)
}

// GetKnowledgeGraph returns the full knowledge graph for visualization.
func (s *EntityGraphService) GetKnowledgeGraph(ctx context.Context, limit int) (*dto.KnowledgeGraphResponse, error) {
	tenantID := tenant.FromContext(ctx)
	return s.entityGraphRepo.GetKnowledgeGraph(ctx, tenantID, limit)
}
