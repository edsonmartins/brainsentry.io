package service

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

type fakeSemanticMemoryRepo struct {
	memories []domain.Memory
	created  []domain.Memory
	pages    []int
}

func (f *fakeSemanticMemoryRepo) List(_ context.Context, page, _ int) ([]domain.Memory, int64, error) {
	f.pages = append(f.pages, page)
	memories := make([]domain.Memory, len(f.memories))
	copy(memories, f.memories)
	return memories, int64(len(memories)), nil
}

func (f *fakeSemanticMemoryRepo) Create(_ context.Context, m *domain.Memory) error {
	copyMemory := *m
	f.created = append(f.created, copyMemory)
	return nil
}

func TestSemanticMemoryConsolidate_ExtractsAndStoresFactsAndWorkflows(t *testing.T) {
	now := time.Now()
	repo := &fakeSemanticMemoryRepo{
		memories: []domain.Memory{
			{ID: "m1", Content: "Use transactional outbox for reliable event publishing.", Category: domain.CategoryKnowledge, CreatedAt: now},
			{ID: "m2", Content: "Transactional outbox prevents lost events when DB commit succeeds.", Category: domain.CategoryKnowledge, CreatedAt: now},
			{ID: "m3", Content: "Implement outbox by writing the event in the same transaction.", Category: domain.CategoryKnowledge, CreatedAt: now},
		},
	}
	llm := &mockLLMProvider{
		name: "test",
		response: `{
			"facts": [
				{"fact": "Transactional outbox prevents lost events.", "confidence": 0.92, "sourceCount": 3}
			],
			"workflows": [
				{"title": "Implement transactional outbox", "steps": [
					{"order": 1, "description": "Write the event in the same transaction."},
					{"order": 2, "description": "Publish pending events asynchronously."}
				]}
			]
		}`,
	}
	svc := NewSemanticMemoryServiceWithRepository(repo, llm, nil)
	ctx := tenant.WithTenant(context.Background(), "tenant-semantic")

	result, err := svc.Consolidate(ctx, 3)
	if err != nil {
		t.Fatalf("Consolidate() error = %v", err)
	}

	if result.MemoriesUsed != 3 {
		t.Fatalf("MemoriesUsed = %d, want 3", result.MemoriesUsed)
	}
	if len(result.SemanticFacts) != 1 {
		t.Fatalf("semantic facts = %d, want 1", len(result.SemanticFacts))
	}
	if result.SemanticFacts[0].Fact != "Transactional outbox prevents lost events." {
		t.Fatalf("fact = %q", result.SemanticFacts[0].Fact)
	}
	if result.SemanticFacts[0].Confidence != 0.92 {
		t.Fatalf("confidence = %f", result.SemanticFacts[0].Confidence)
	}
	if !reflect.DeepEqual(result.SemanticFacts[0].SourceMemoryIDs, []string{"m1", "m2", "m3"}) {
		t.Fatalf("source ids = %v", result.SemanticFacts[0].SourceMemoryIDs)
	}
	if len(result.Workflows) != 1 || len(result.Workflows[0].Steps) != 2 {
		t.Fatalf("workflows = %#v", result.Workflows)
	}
	if llm.callCount != 1 {
		t.Fatalf("LLM call count = %d, want 1", llm.callCount)
	}
	if !reflect.DeepEqual(repo.pages, []int{0}) {
		t.Fatalf("expected first page scan, got pages %v", repo.pages)
	}
	if len(repo.created) != 2 {
		t.Fatalf("created consolidated memories = %d, want 2", len(repo.created))
	}

	semantic := repo.created[0]
	if semantic.MemoryType != domain.MemoryTypeSemantic {
		t.Fatalf("semantic MemoryType = %s", semantic.MemoryType)
	}
	if semantic.TenantID != "tenant-semantic" {
		t.Fatalf("semantic TenantID = %s", semantic.TenantID)
	}
	var semanticMeta map[string]any
	if err := json.Unmarshal(semantic.Metadata, &semanticMeta); err != nil {
		t.Fatalf("semantic metadata JSON error = %v", err)
	}
	if semanticMeta["type"] != "semantic_consolidation" {
		t.Fatalf("semantic metadata type = %v", semanticMeta["type"])
	}

	procedural := repo.created[1]
	if procedural.MemoryType != domain.MemoryTypeProcedural {
		t.Fatalf("procedural MemoryType = %s", procedural.MemoryType)
	}
	if procedural.TenantID != "tenant-semantic" {
		t.Fatalf("procedural TenantID = %s", procedural.TenantID)
	}
	var proceduralMeta map[string]any
	if err := json.Unmarshal(procedural.Metadata, &proceduralMeta); err != nil {
		t.Fatalf("procedural metadata JSON error = %v", err)
	}
	if proceduralMeta["type"] != "procedural_consolidation" {
		t.Fatalf("procedural metadata type = %v", proceduralMeta["type"])
	}
}

func TestSemanticMemoryConsolidate_BelowMinimumDoesNotCallLLMOrStore(t *testing.T) {
	repo := &fakeSemanticMemoryRepo{
		memories: []domain.Memory{
			{ID: "m1", Content: "one memory", Category: domain.CategoryKnowledge},
			{ID: "m2", Content: "two memories", Category: domain.CategoryKnowledge},
		},
	}
	llm := &mockLLMProvider{name: "test", response: `{"facts":[],"workflows":[]}`}
	svc := NewSemanticMemoryServiceWithRepository(repo, llm, nil)

	result, err := svc.Consolidate(context.Background(), 3)
	if err != nil {
		t.Fatalf("Consolidate() error = %v", err)
	}

	if result.MemoriesUsed != 2 {
		t.Fatalf("MemoriesUsed = %d, want 2", result.MemoriesUsed)
	}
	if llm.callCount != 0 {
		t.Fatalf("LLM call count = %d, want 0", llm.callCount)
	}
	if len(repo.created) != 0 {
		t.Fatalf("created memories = %d, want 0", len(repo.created))
	}
	if !reflect.DeepEqual(repo.pages, []int{0}) {
		t.Fatalf("expected first page scan, got pages %v", repo.pages)
	}
}
