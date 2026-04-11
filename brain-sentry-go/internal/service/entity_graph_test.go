package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

type fakeEntityExtractor struct {
	result *EntityExtractionResult
	err    error
}

func (f fakeEntityExtractor) ExtractEntities(context.Context, string) (*EntityExtractionResult, error) {
	return f.result, f.err
}

type storedEntity struct {
	name           string
	entityType     string
	tenantID       string
	sourceMemoryID string
	properties     map[string]string
}

type storedRelationship struct {
	sourceNodeID string
	targetNodeID string
	relType      string
	tenantID     string
	properties   map[string]string
}

type fakeEntityGraphRepo struct {
	entityIDs     map[string]string
	entities      []storedEntity
	relationships []storedRelationship
}

func (f *fakeEntityGraphRepo) StoreEntity(_ context.Context, name, entityType, tenantID, sourceMemoryID string, properties map[string]string) (string, error) {
	f.entities = append(f.entities, storedEntity{
		name:           name,
		entityType:     entityType,
		tenantID:       tenantID,
		sourceMemoryID: sourceMemoryID,
		properties:     properties,
	})
	if id, ok := f.entityIDs[name]; ok {
		return id, nil
	}
	return "node-" + name, nil
}

func (f *fakeEntityGraphRepo) StoreRelationship(_ context.Context, sourceNodeID, targetNodeID, relType, tenantID string, properties map[string]string) error {
	f.relationships = append(f.relationships, storedRelationship{
		sourceNodeID: sourceNodeID,
		targetNodeID: targetNodeID,
		relType:      relType,
		tenantID:     tenantID,
		properties:   properties,
	})
	return nil
}

func (f *fakeEntityGraphRepo) FindEntitiesByMemory(context.Context, string, string) ([]dto.EntityNode, error) {
	return nil, nil
}

func (f *fakeEntityGraphRepo) FindRelationshipsByMemory(context.Context, string, string) ([]dto.EntityEdge, error) {
	return nil, nil
}

func (f *fakeEntityGraphRepo) SearchEntities(context.Context, string, string, int) ([]dto.EntityNode, error) {
	return nil, nil
}

func (f *fakeEntityGraphRepo) GetKnowledgeGraph(context.Context, string, int) (*dto.KnowledgeGraphResponse, error) {
	return nil, nil
}

func TestEntityGraphService_ExtractAndStoreEntities(t *testing.T) {
	repo := &fakeEntityGraphRepo{
		entityIDs: map[string]string{
			"PostgreSQL":                "node-postgres",
			"Transactional Outbox":      "node-outbox",
			"Missing Relationship Node": "node-missing",
		},
	}
	svc := &EntityGraphService{
		entityGraphRepo: repo,
		entityExtractor: fakeEntityExtractor{
			result: &EntityExtractionResult{
				Entities: []ExtractedEntity{
					{Name: "PostgreSQL", Type: "TECHNOLOGY", Properties: map[string]string{"role": "database"}},
					{Name: "Transactional Outbox", Type: "CONCEPT", Properties: map[string]string{"scope": "reliability"}},
				},
				Relationships: []ExtractedRelationship{
					{Source: "Transactional Outbox", Target: "PostgreSQL", Type: "USES", Properties: map[string]string{"reason": "storage"}},
					{Source: "Transactional Outbox", Target: "Missing Relationship Node", Type: "RELATED_TO"},
				},
			},
		},
	}
	ctx := tenant.WithTenant(context.Background(), "tenant-graph")
	memory := &domain.Memory{ID: "memory-1", Content: "Transactional outbox uses PostgreSQL."}

	if err := svc.ExtractAndStoreEntities(ctx, memory); err != nil {
		t.Fatalf("ExtractAndStoreEntities() error = %v", err)
	}

	if len(repo.entities) != 2 {
		t.Fatalf("stored entities = %d, want 2", len(repo.entities))
	}
	if repo.entities[0].tenantID != "tenant-graph" || repo.entities[0].sourceMemoryID != "memory-1" {
		t.Fatalf("entity tenant/source mismatch: %#v", repo.entities[0])
	}
	if !reflect.DeepEqual(repo.entities[0].properties, map[string]string{"role": "database"}) {
		t.Fatalf("entity properties = %#v", repo.entities[0].properties)
	}
	if len(repo.relationships) != 1 {
		t.Fatalf("stored relationships = %d, want 1: %#v", len(repo.relationships), repo.relationships)
	}
	if repo.relationships[0].sourceNodeID != "node-outbox" || repo.relationships[0].targetNodeID != "node-postgres" {
		t.Fatalf("relationship nodes = %#v", repo.relationships[0])
	}
	if repo.relationships[0].relType != "USES" || repo.relationships[0].tenantID != "tenant-graph" {
		t.Fatalf("relationship type/tenant = %#v", repo.relationships[0])
	}
}

func TestEntityGraphService_ExtractAndStoreEntities_NoExtractorOrRepoNoops(t *testing.T) {
	memory := &domain.Memory{ID: "memory-1", Content: "content"}

	if err := (&EntityGraphService{}).ExtractAndStoreEntities(context.Background(), memory); err != nil {
		t.Fatalf("no deps should no-op, got %v", err)
	}
	if err := (&EntityGraphService{entityExtractor: fakeEntityExtractor{result: &EntityExtractionResult{}}}).ExtractAndStoreEntities(context.Background(), memory); err != nil {
		t.Fatalf("missing repo should no-op, got %v", err)
	}
}

func TestEntityGraphService_ExtractAndStoreEntities_ExtractorError(t *testing.T) {
	svc := &EntityGraphService{
		entityGraphRepo: &fakeEntityGraphRepo{},
		entityExtractor: fakeEntityExtractor{err: errors.New("llm unavailable")},
	}

	if err := svc.ExtractAndStoreEntities(context.Background(), &domain.Memory{ID: "m1", Content: "content"}); err == nil {
		t.Fatal("expected extractor error")
	}
}
