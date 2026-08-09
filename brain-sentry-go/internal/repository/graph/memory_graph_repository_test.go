package graph

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

type recordingGraphBackend struct {
	queries []string
	results []*QueryResult
	err     error
}

func (f *recordingGraphBackend) Query(_ context.Context, query string) (*QueryResult, error) {
	f.queries = append(f.queries, query)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.results) == 0 {
		return &QueryResult{}, nil
	}
	result := f.results[0]
	f.results = f.results[1:]
	return result, nil
}
func (f *recordingGraphBackend) Close() error { return nil }
func (f *recordingGraphBackend) Name() string { return "recording" }

func TestMemoryGraphWritesAreTenantScopedAndEscaped(t *testing.T) {
	backend := &recordingGraphBackend{}
	repo := &MemoryGraphRepository{client: backend}
	memory := &domain.Memory{ID: "memory'1", TenantID: "tenant'1", Content: "customer's fact", Tags: []string{"sales"}, Embedding: []float32{0.1, 0.2}}

	if err := repo.SaveToGraph(context.Background(), memory); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTagRelationships(context.Background(), memory); err != nil {
		t.Fatal(err)
	}
	if err := repo.PurgeMemoryNodes(context.Background(), memory.TenantID, []string{memory.ID}); err != nil {
		t.Fatal(err)
	}

	joined := strings.Join(backend.queries, "\n")
	if !strings.Contains(joined, "tenantId: 'tenant\\'1'") || !strings.Contains(joined, "m2.tenantId = 'tenant\\'1'") {
		t.Fatalf("tenant scope missing from Cypher:\n%s", joined)
	}
	if !strings.Contains(joined, "memory\\'1") || !strings.Contains(joined, "customer\\'s fact") {
		t.Fatalf("Cypher values were not escaped:\n%s", joined)
	}
	if !strings.Contains(joined, "m.embedding = vecf32([0.100000, 0.200000])") {
		t.Fatalf("embedding was not stored as a FalkorDB vector:\n%s", joined)
	}
}

func TestMemoryGraphVectorSearchFiltersTenantAndReturnsScores(t *testing.T) {
	backend := &recordingGraphBackend{results: []*QueryResult{{Records: []Record{{Values: map[string]any{"id": "m1", "score": 0.91}}}}}}
	repo := &MemoryGraphRepository{client: backend}

	ids, scores, err := repo.VectorSearch(context.Background(), []float32{0.1, 0.2}, 5, "tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "m1" || len(scores) != 1 || scores[0] != 0.91 {
		t.Fatalf("ids=%v scores=%v", ids, scores)
	}
	query := backend.queries[0]
	if !strings.Contains(query, "vecf32([0.100000, 0.200000])") || !strings.Contains(query, "node.tenantId = 'tenant-a'") {
		t.Fatalf("unexpected vector query: %s", query)
	}
}

func TestMemoryGraphVectorFailureUsesTenantScopedFallback(t *testing.T) {
	backend := &recordingGraphBackend{err: errors.New("vector unavailable")}
	repo := &MemoryGraphRepository{client: backend}

	if _, _, err := repo.VectorSearch(context.Background(), []float32{1}, 3, "tenant-a"); err == nil {
		t.Fatal("expected fallback error")
	}
	if len(backend.queries) != 2 || !strings.Contains(backend.queries[1], "m.tenantId = 'tenant-a'") {
		t.Fatalf("fallback was not tenant scoped: %v", backend.queries)
	}
}

func TestEnsureVectorIndexAcceptsOnlyKnownIdempotencyErrors(t *testing.T) {
	for _, message := range []string{"Attribute already exists", "Attribute 'embedding' is already indexed"} {
		backend := &recordingGraphBackend{err: errors.New(message)}
		if err := ensureVectorIndex(context.Background(), backend, 384); err != nil {
			t.Fatalf("known idempotency error %q must be accepted: %v", message, err)
		}
	}

	backend := &recordingGraphBackend{err: errors.New("ERR invalid vector index options")}
	if err := ensureVectorIndex(context.Background(), backend, 384); err == nil {
		t.Fatal("unexpected FalkorDB errors must not be hidden")
	}
}

func TestMemoryGraphTraversalMapsTenantScopedResults(t *testing.T) {
	backend := &recordingGraphBackend{results: []*QueryResult{
		{Records: []Record{{Values: map[string]any{"id": "related"}}}},
		{Records: []Record{{Values: map[string]any{
			"fromId": "m1", "toId": "m2", "fromSummary": "one", "toSummary": "two",
			"tag": "shared", "strength": 0.8, "type": "shared_tag",
		}}}},
	}}
	repo := &MemoryGraphRepository{client: backend}

	ids, err := repo.FindRelated(context.Background(), "m1", 0, "tenant-a")
	if err != nil || len(ids) != 1 || ids[0] != "related" {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	rels, err := repo.GetGraphRelationships(context.Background(), "tenant-a", 0)
	if err != nil || len(rels) != 1 || rels[0].Strength != 0.8 {
		t.Fatalf("relationships=%v err=%v", rels, err)
	}
	for _, query := range backend.queries {
		if !strings.Contains(query, "tenant-a") {
			t.Fatalf("tenant scope missing: %s", query)
		}
	}
	if err := repo.DeleteMemory(context.Background(), "m1"); err == nil {
		t.Fatal("unscoped graph deletion must be rejected")
	}
}

func TestGraphRAGTraversalAndDeduplication(t *testing.T) {
	direct := Record{Values: map[string]any{
		"memoryId": "direct", "summary": "Direct", "category": "KNOWLEDGE",
		"importance": "IMPORTANT", "hopDistance": int64(1), "path": []any{"seed", "direct"}, "score": 0.9,
	}}
	extended := Record{Values: map[string]any{
		"memoryId": "extended", "summary": "Extended", "category": "CONTEXT",
		"importance": "MINOR", "hopDistance": int64(2), "path": []any{"seed", "direct", "extended"}, "score": 0.7,
	}}
	backend := &recordingGraphBackend{results: []*QueryResult{
		{Records: []Record{direct}},
		{Records: []Record{direct, extended}},
		{Records: []Record{{Values: map[string]any{"id": "clustered"}}}},
	}}
	repo := &GraphRAGRepository{client: backend}
	indexCalls := 0
	repo.SetIndexInitializer(func(context.Context) error { indexCalls++; return nil })

	results, err := repo.EnrichContext(context.Background(), []string{"seed"}, "tenant-a")
	if err != nil || len(results) != 2 || results[0].MemoryID != "direct" || results[1].MemoryID != "extended" {
		t.Fatalf("results=%v err=%v", results, err)
	}
	if indexCalls != 2 || len(results[1].Path) != 3 {
		t.Fatalf("indexCalls=%d results=%v", indexCalls, results)
	}
	cluster, err := repo.GetCluster(context.Background(), "seed", "tenant-a", 0)
	if err != nil || len(cluster) != 1 || cluster[0] != "clustered" {
		t.Fatalf("cluster=%v err=%v", cluster, err)
	}
	if empty, err := repo.MultiHopSearch(context.Background(), nil, 0, 0, "tenant-a"); err != nil || empty != nil {
		t.Fatalf("empty seeds=%v err=%v", empty, err)
	}
}
