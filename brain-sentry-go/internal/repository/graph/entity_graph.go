package graph

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/integraltech/brainsentry/internal/dto"
)

// EntityGraphRepository handles entity and relationship storage in FalkorDB.
type EntityGraphRepository struct {
	client *Client
}

// NewEntityGraphRepository creates a new EntityGraphRepository.
func NewEntityGraphRepository(client *Client) *EntityGraphRepository {
	return &EntityGraphRepository{client: client}
}

// StoreEntity stores an entity node in the graph.
func (r *EntityGraphRepository) StoreEntity(ctx context.Context, name, entityType, tenantID, sourceMemoryID string, properties map[string]string) (string, error) {
	nodeID := "ent_" + uuid.New().String()[:12]

	propsStr := ""
	if len(properties) > 0 {
		parts := make([]string, 0, len(properties))
		for k, v := range properties {
			parts = append(parts, fmt.Sprintf("e.%s = '%s'", EscapeCypher(k), EscapeCypher(v)))
		}
		propsStr = ", " + strings.Join(parts, ", ")
	}

	label := EscapeCypherIdentifier(entityType)

	cypher := fmt.Sprintf(`CREATE (e:Entity:%s {
		id: '%s',
		name: '%s',
		type: '%s',
		tenantId: '%s',
		sourceMemoryId: '%s',
		createdAt: %d
		%s
	}) RETURN e.id as id`,
		label,
		EscapeCypher(nodeID),
		EscapeCypher(name),
		EscapeCypher(entityType),
		EscapeCypher(tenantID),
		EscapeCypher(sourceMemoryID),
		time.Now().UnixMilli(),
		propsStr,
	)

	_, err := r.client.Query(ctx, cypher)
	if err != nil {
		return "", fmt.Errorf("storing entity: %w", err)
	}

	// Create MENTIONS relationship from Memory to Entity
	mentionsCypher := fmt.Sprintf(`MATCH (m:Memory {id: '%s'}), (e:Entity {id: '%s'})
CREATE (m)-[:MENTIONS]->(e)`,
		EscapeCypher(sourceMemoryID),
		EscapeCypher(nodeID),
	)
	if _, err := r.client.Query(ctx, mentionsCypher); err != nil {
		slog.Warn("failed to create MENTIONS relationship", "error", err)
	}

	return nodeID, nil
}

// StoreRelationship stores a relationship edge between two entities.
func (r *EntityGraphRepository) StoreRelationship(ctx context.Context, sourceNodeID, targetNodeID, relType, tenantID string, properties map[string]string) error {
	propsStr := ""
	if len(properties) > 0 {
		parts := make([]string, 0, len(properties))
		for k, v := range properties {
			parts = append(parts, fmt.Sprintf("%s: '%s'", EscapeCypher(k), EscapeCypher(v)))
		}
		propsStr = ", " + strings.Join(parts, ", ")
	}

	relLabel := EscapeCypherIdentifier(relType)

	cypher := fmt.Sprintf(`MATCH (source:Entity {id: '%s'}), (target:Entity {id: '%s'})
CREATE (source)-[r:%s {
	tenantId: '%s',
	createdAt: %d
	%s
}]->(target)`,
		EscapeCypher(sourceNodeID),
		EscapeCypher(targetNodeID),
		relLabel,
		EscapeCypher(tenantID),
		time.Now().UnixMilli(),
		propsStr,
	)

	_, err := r.client.Query(ctx, cypher)
	if err != nil {
		return fmt.Errorf("storing relationship: %w", err)
	}

	return nil
}

// FindEntitiesByMemory returns entities extracted from a specific memory.
func (r *EntityGraphRepository) FindEntitiesByMemory(ctx context.Context, memoryID, tenantID string) ([]dto.EntityNode, error) {
	cypher := fmt.Sprintf(`MATCH (m:Memory {id: '%s'})-[:MENTIONS]->(e:Entity)
WHERE e.tenantId = '%s'
RETURN e.id as id, e.name as name, e.type as type, e.sourceMemoryId as sourceMemoryId`,
		EscapeCypher(memoryID),
		EscapeCypher(tenantID),
	)

	result, err := r.client.Query(ctx, cypher)
	if err != nil {
		return nil, fmt.Errorf("finding entities by memory: %w", err)
	}

	nodes := make([]dto.EntityNode, 0, len(result.Records))
	for _, rec := range result.Records {
		nodes = append(nodes, dto.EntityNode{
			ID:             GetString(rec.Values, "id"),
			Name:           GetString(rec.Values, "name"),
			Type:           GetString(rec.Values, "type"),
			SourceMemoryID: GetString(rec.Values, "sourceMemoryId"),
		})
	}

	return nodes, nil
}

// FindRelationshipsByMemory returns relationships between entities from a memory.
func (r *EntityGraphRepository) FindRelationshipsByMemory(ctx context.Context, memoryID, tenantID string) ([]dto.EntityEdge, error) {
	cypher := fmt.Sprintf(`MATCH (m:Memory {id: '%s'})-[:MENTIONS]->(e1:Entity)-[rel]->(e2:Entity)
WHERE e1.tenantId = '%s' AND e2.tenantId = '%s'
RETURN id(rel) as id, e1.id as sourceId, e2.id as targetId,
       e1.name as sourceName, e2.name as targetName, type(rel) as type`,
		EscapeCypher(memoryID),
		EscapeCypher(tenantID),
		EscapeCypher(tenantID),
	)

	result, err := r.client.Query(ctx, cypher)
	if err != nil {
		return nil, fmt.Errorf("finding relationships by memory: %w", err)
	}

	edges := make([]dto.EntityEdge, 0, len(result.Records))
	for _, rec := range result.Records {
		edges = append(edges, dto.EntityEdge{
			ID:         fmt.Sprintf("%v", rec.Values["id"]),
			SourceID:   GetString(rec.Values, "sourceId"),
			TargetID:   GetString(rec.Values, "targetId"),
			SourceName: GetString(rec.Values, "sourceName"),
			TargetName: GetString(rec.Values, "targetName"),
			Type:       GetString(rec.Values, "type"),
		})
	}

	return edges, nil
}

// SearchEntities searches for entities by name (case-insensitive CONTAINS).
func (r *EntityGraphRepository) SearchEntities(ctx context.Context, searchTerm, tenantID string, limit int) ([]dto.EntityNode, error) {
	if limit <= 0 {
		limit = 20
	}

	cypher := fmt.Sprintf(`MATCH (e:Entity)
WHERE e.tenantId = '%s' AND toLower(e.name) CONTAINS toLower('%s')
RETURN e.id as id, e.name as name, e.type as type, e.sourceMemoryId as sourceMemoryId
LIMIT %d`,
		EscapeCypher(tenantID),
		EscapeCypher(searchTerm),
		limit,
	)

	result, err := r.client.Query(ctx, cypher)
	if err != nil {
		return nil, fmt.Errorf("searching entities: %w", err)
	}

	nodes := make([]dto.EntityNode, 0, len(result.Records))
	for _, rec := range result.Records {
		nodes = append(nodes, dto.EntityNode{
			ID:             GetString(rec.Values, "id"),
			Name:           GetString(rec.Values, "name"),
			Type:           GetString(rec.Values, "type"),
			SourceMemoryID: GetString(rec.Values, "sourceMemoryId"),
		})
	}

	return nodes, nil
}

// GetKnowledgeGraph returns all entities and relationships for visualization.
func (r *EntityGraphRepository) GetKnowledgeGraph(ctx context.Context, tenantID string, limit int) (*dto.KnowledgeGraphResponse, error) {
	if limit <= 0 {
		limit = 100
	}

	// Get entities
	nodesCypher := fmt.Sprintf(`MATCH (e:Entity)
WHERE e.tenantId = '%s'
RETURN e.id as id, e.name as name, e.type as type, e.sourceMemoryId as sourceMemoryId
LIMIT %d`,
		EscapeCypher(tenantID),
		limit,
	)

	nodesResult, err := r.client.Query(ctx, nodesCypher)
	if err != nil {
		return nil, fmt.Errorf("getting knowledge graph nodes: %w", err)
	}

	nodes := make([]dto.EntityNode, 0, len(nodesResult.Records))
	for _, rec := range nodesResult.Records {
		nodes = append(nodes, dto.EntityNode{
			ID:             GetString(rec.Values, "id"),
			Name:           GetString(rec.Values, "name"),
			Type:           GetString(rec.Values, "type"),
			SourceMemoryID: GetString(rec.Values, "sourceMemoryId"),
		})
	}

	// Get edges
	edgesCypher := fmt.Sprintf(`MATCH (e1:Entity)-[r]->(e2:Entity)
WHERE e1.tenantId = '%s' AND e2.tenantId = '%s'
RETURN id(r) as id, e1.id as sourceId, e2.id as targetId,
       e1.name as sourceName, e2.name as targetName, type(r) as type
LIMIT %d`,
		EscapeCypher(tenantID),
		EscapeCypher(tenantID),
		limit,
	)

	edgesResult, err := r.client.Query(ctx, edgesCypher)
	if err != nil {
		return nil, fmt.Errorf("getting knowledge graph edges: %w", err)
	}

	edges := make([]dto.EntityEdge, 0, len(edgesResult.Records))
	for _, rec := range edgesResult.Records {
		edges = append(edges, dto.EntityEdge{
			ID:         fmt.Sprintf("%v", rec.Values["id"]),
			SourceID:   GetString(rec.Values, "sourceId"),
			TargetID:   GetString(rec.Values, "targetId"),
			SourceName: GetString(rec.Values, "sourceName"),
			TargetName: GetString(rec.Values, "targetName"),
			Type:       GetString(rec.Values, "type"),
		})
	}

	return &dto.KnowledgeGraphResponse{
		Nodes:      nodes,
		Edges:      edges,
		TotalNodes: len(nodes),
		TotalEdges: len(edges),
	}, nil
}
