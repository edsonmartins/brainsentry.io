package graph

import (
	"context"
	"strings"
	"testing"
)

func TestEntityGraphWritesAreValidAndTenantScoped(t *testing.T) {
	backend := &recordingGraphBackend{}
	repo := &EntityGraphRepository{client: backend}

	entityID, err := repo.StoreEntity(context.Background(), "Customer's API", "System", "tenant-a", "memory-a", map[string]string{"owner-name": "O'Reilly"})
	if err != nil {
		t.Fatal(err)
	}
	if entityID == "" || len(backend.queries) != 2 {
		t.Fatalf("entity=%q queries=%d", entityID, len(backend.queries))
	}
	createQuery := backend.queries[0]
	if !strings.Contains(createQuery, "}) SET e.`owner-name` = 'O\\'Reilly' RETURN") {
		t.Fatalf("additional properties must use a valid SET clause: %s", createQuery)
	}
	mentionsQuery := backend.queries[1]
	if strings.Count(mentionsQuery, "tenantId: 'tenant-a'") != 2 {
		t.Fatalf("MENTIONS endpoints must both be tenant scoped: %s", mentionsQuery)
	}

	if err := repo.StoreRelationship(context.Background(), "source", "target", "DEPENDS-ON", "tenant-a", map[string]string{"reason": "required"}); err != nil {
		t.Fatal(err)
	}
	relQuery := backend.queries[2]
	if strings.Count(relQuery, "tenantId: 'tenant-a'") < 3 || !strings.Contains(relQuery, "[r:`DEPENDS-ON`") {
		t.Fatalf("relationship was not safely tenant scoped: %s", relQuery)
	}
}

func TestEntityGraphReadsMapRecordsAndScopeTenant(t *testing.T) {
	node := Record{Values: map[string]any{"id": "e1", "name": "API", "type": "System", "sourceMemoryId": "m1"}}
	edge := Record{Values: map[string]any{"id": int64(7), "sourceId": "e1", "targetId": "e2", "sourceName": "API", "targetName": "DB", "type": "USES"}}
	backend := &recordingGraphBackend{results: []*QueryResult{
		{Records: []Record{node}},
		{Records: []Record{edge}},
		{Records: []Record{node}},
		{Records: []Record{node}},
		{Records: []Record{edge}},
	}}
	repo := &EntityGraphRepository{client: backend}

	if nodes, err := repo.FindEntitiesByMemory(context.Background(), "m1", "tenant-a"); err != nil || len(nodes) != 1 || nodes[0].ID != "e1" {
		t.Fatalf("nodes=%v err=%v", nodes, err)
	}
	if edges, err := repo.FindRelationshipsByMemory(context.Background(), "m1", "tenant-a"); err != nil || len(edges) != 1 || edges[0].SourceID != "e1" {
		t.Fatalf("edges=%v err=%v", edges, err)
	}
	if nodes, err := repo.SearchEntities(context.Background(), "api", "tenant-a", 0); err != nil || len(nodes) != 1 {
		t.Fatalf("search=%v err=%v", nodes, err)
	}
	graph, err := repo.GetKnowledgeGraph(context.Background(), "tenant-a", 0)
	if err != nil || graph.TotalNodes != 1 || graph.TotalEdges != 1 {
		t.Fatalf("graph=%v err=%v", graph, err)
	}
	for _, query := range backend.queries {
		if !strings.Contains(query, "tenant-a") {
			t.Fatalf("tenant scope missing: %s", query)
		}
	}
}
