//go:build integration

package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

func TestSemanticMemoryService_PostgresStoresSemanticAndProceduralMemories(t *testing.T) {
	pool, cleanup := setupAutoForgetIntegration(t)
	defer cleanup()

	ctx := tenant.WithTenant(context.Background(), "tenant-semantic-integration")
	ensureAutoForgetTenant(t, ctx, pool, "tenant-semantic-integration")

	repo := postgres.NewMemoryRepository(pool)
	sourceMemories := []*domain.Memory{
		{
			Content:    "Use transactional outbox for reliable event publishing.",
			Summary:    "outbox fact 1",
			Category:   domain.CategoryKnowledge,
			Importance: domain.ImportanceImportant,
		},
		{
			Content:    "Transactional outbox prevents event loss after database commit.",
			Summary:    "outbox fact 2",
			Category:   domain.CategoryKnowledge,
			Importance: domain.ImportanceImportant,
		},
		{
			Content:    "Implement outbox by writing events in the same transaction and publishing asynchronously.",
			Summary:    "outbox workflow",
			Category:   domain.CategoryKnowledge,
			Importance: domain.ImportanceImportant,
		},
	}
	for _, memory := range sourceMemories {
		if err := repo.Create(ctx, memory); err != nil {
			t.Fatalf("creating source memory %s: %v", memory.Summary, err)
		}
	}

	llm := &mockLLMProvider{
		name: "semantic-integration",
		response: `{
			"facts": [
				{"fact": "Transactional outbox prevents event loss after database commit.", "confidence": 0.93, "sourceCount": 3}
			],
			"workflows": [
				{"title": "Implement transactional outbox", "steps": [
					{"order": 1, "description": "Write event rows in the same transaction."},
					{"order": 2, "description": "Publish pending event rows asynchronously."}
				]}
			]
		}`,
	}
	svc := NewSemanticMemoryService(repo, llm, nil)

	result, err := svc.Consolidate(ctx, 3)
	if err != nil {
		t.Fatalf("Consolidate() error = %v", err)
	}
	if len(result.SemanticFacts) != 1 || len(result.Workflows) != 1 {
		t.Fatalf("unexpected semantic result: %#v", result)
	}

	memories, _, err := repo.List(ctx, 0, 20)
	if err != nil {
		t.Fatalf("listing memories: %v", err)
	}

	var semantic, procedural *domain.Memory
	for i := range memories {
		switch memories[i].MemoryType {
		case domain.MemoryTypeSemantic:
			if strings.HasPrefix(memories[i].Content, "Semantic consolidation:") {
				semantic = &memories[i]
			}
		case domain.MemoryTypeProcedural:
			if strings.HasPrefix(memories[i].Content, "Procedural consolidation:") {
				procedural = &memories[i]
			}
		}
	}
	if semantic == nil {
		t.Fatal("expected stored semantic consolidation memory")
	}
	if procedural == nil {
		t.Fatal("expected stored procedural consolidation memory")
	}
	assertMetadataType(t, semantic.Metadata, "semantic_consolidation")
	assertMetadataType(t, procedural.Metadata, "procedural_consolidation")
	if semantic.TenantID != "tenant-semantic-integration" || procedural.TenantID != "tenant-semantic-integration" {
		t.Fatalf("stored consolidation memories have wrong tenant: semantic=%s procedural=%s", semantic.TenantID, procedural.TenantID)
	}
}

func TestConsolidationService_PostgresMergesAndSoftDeletesSecondary(t *testing.T) {
	pool, cleanup := setupAutoForgetIntegration(t)
	defer cleanup()

	ctx := tenant.WithTenant(context.Background(), "tenant-consolidation-integration")
	ensureAutoForgetTenant(t, ctx, pool, "tenant-consolidation-integration")

	repo := postgres.NewMemoryRepository(pool)
	older := &domain.Memory{
		Content:         "use postgres jsonb metadata for memory storage",
		Summary:         "older",
		Category:        domain.CategoryKnowledge,
		Importance:      domain.ImportanceCritical,
		Tags:            []string{"metadata", "jsonb"},
		Embedding:       []float32{1, 0, 0},
		AccessCount:     4,
		HelpfulCount:    2,
		NotHelpfulCount: 1,
	}
	newer := &domain.Memory{
		Content:      "use postgres jsonb metadata for memory storage",
		Summary:      "newer",
		Category:     domain.CategoryKnowledge,
		Importance:   domain.ImportanceMinor,
		Tags:         []string{"postgres"},
		Embedding:    []float32{1, 0, 0},
		AccessCount:  3,
		HelpfulCount: 1,
	}
	for _, memory := range []*domain.Memory{older, newer} {
		if err := repo.Create(ctx, memory); err != nil {
			t.Fatalf("creating memory %s: %v", memory.Summary, err)
		}
	}

	now := time.Now().UTC().Truncate(time.Second)
	if _, err := pool.Exec(ctx, `UPDATE memories SET created_at = $1, updated_at = $1 WHERE id = $2`, now.Add(-2*time.Hour), older.ID); err != nil {
		t.Fatalf("backdating older memory: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE memories SET created_at = $1, updated_at = $1 WHERE id = $2`, now.Add(-time.Hour), newer.ID); err != nil {
		t.Fatalf("backdating newer memory: %v", err)
	}

	llm := &mockLLMProvider{
		name:     "consolidation-integration",
		response: `{"content":"use postgres jsonb metadata for memory storage with explicit tags","summary":"merged postgres metadata"}`,
	}
	svc := NewConsolidationService(repo, nil, nil, nil)
	svc.llm = llm

	result, err := svc.ConsolidateTenant(ctx, 0.99)
	if err != nil {
		t.Fatalf("ConsolidateTenant() error = %v", err)
	}
	if result.Consolidated != 1 {
		t.Fatalf("Consolidated = %d, want 1", result.Consolidated)
	}
	if len(result.MergedIDs) != 1 || result.MergedIDs[0] != older.ID {
		t.Fatalf("MergedIDs = %v, want [%s]", result.MergedIDs, older.ID)
	}

	updated, err := repo.FindByID(ctx, newer.ID)
	if err != nil {
		t.Fatalf("newer primary should remain: %v", err)
	}
	if updated.Content != "use postgres jsonb metadata for memory storage with explicit tags" {
		t.Fatalf("updated content = %q", updated.Content)
	}
	if updated.Importance != domain.ImportanceCritical {
		t.Fatalf("expected critical importance to be preserved, got %s", updated.Importance)
	}
	if updated.AccessCount != 7 || updated.HelpfulCount != 3 || updated.NotHelpfulCount != 1 {
		t.Fatalf("expected counters to be summed, got access=%d helpful=%d notHelpful=%d",
			updated.AccessCount, updated.HelpfulCount, updated.NotHelpfulCount)
	}
	assertContainsTag(t, updated.Tags, "metadata")
	assertContainsTag(t, updated.Tags, "jsonb")
	assertContainsTag(t, updated.Tags, "postgres")

	if _, err := repo.FindByID(ctx, older.ID); err == nil {
		t.Fatal("older secondary memory should be soft-deleted")
	}
}

func TestCrossSessionService_PostgresCreatesRedactedEpisodicMemory(t *testing.T) {
	pool, cleanup := setupAutoForgetIntegration(t)
	defer cleanup()

	ctx := tenant.WithTenant(context.Background(), "tenant-cross-session-integration")
	ensureAutoForgetTenant(t, ctx, pool, "tenant-cross-session-integration")

	repo := postgres.NewMemoryRepository(pool)
	svc := NewCrossSessionService(repo, nil, nil)
	svc.redactionLevel = RedactionPartial

	if _, err := svc.OnSessionStart(ctx, "session-cross-1"); err != nil {
		t.Fatalf("OnSessionStart() error = %v", err)
	}
	svc.RecordEvent(
		"session-cross-1",
		domain.ObservationDecision,
		"Use queue for async memory writes",
		"Decision owner jane.doe@example.com chose queue based async memory writes.",
		map[string]any{"source": "integration"},
	)

	result, err := svc.OnSessionEnd(ctx, "session-cross-1")
	if err != nil {
		t.Fatalf("OnSessionEnd() error = %v", err)
	}
	if result.EventsRecorded != 1 || result.ObservationsFound != 1 || result.EntriesCreated != 1 {
		t.Fatalf("unexpected cross-session result: %#v", result)
	}

	memories, _, err := repo.List(ctx, 0, 10)
	if err != nil {
		t.Fatalf("listing memories: %v", err)
	}
	var stored *domain.Memory
	for i := range memories {
		if memories[i].MemoryType == domain.MemoryTypeEpisodic && hasTag(memories[i].Tags, "session:session-cross-1") {
			stored = &memories[i]
			break
		}
	}
	if stored == nil {
		t.Fatalf("expected stored episodic cross-session memory in %#v", memories)
	}
	if stored.TenantID != "tenant-cross-session-integration" {
		t.Fatalf("stored tenant = %s", stored.TenantID)
	}
	if stored.Importance != domain.ImportanceCritical {
		t.Fatalf("decision observation should be critical, got %s", stored.Importance)
	}
	if !hasTag(stored.Tags, "cross-session") || !hasTag(stored.Tags, string(domain.ObservationDecision)) {
		t.Fatalf("stored tags missing cross-session provenance: %v", stored.Tags)
	}
	if strings.Contains(stored.Content, "jane.doe@example.com") {
		t.Fatalf("PII email was not masked: %q", stored.Content)
	}
	if !strings.Contains(stored.Content, "[EMAIL]") {
		t.Fatalf("expected masked email marker in content: %q", stored.Content)
	}
	if !strings.Contains(stored.Summary, "Use queue for async memory writes") {
		t.Fatalf("summary missing event title: %q", stored.Summary)
	}
}

func assertMetadataType(t *testing.T, raw json.RawMessage, want string) {
	t.Helper()
	var meta map[string]any
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("metadata JSON error: %v", err)
	}
	if meta["type"] != want {
		t.Fatalf("metadata type = %v, want %s", meta["type"], want)
	}
}

func assertContainsTag(t *testing.T, tags []string, want string) {
	t.Helper()
	for _, tag := range tags {
		if tag == want {
			return
		}
	}
	t.Fatalf("expected tag %q in %v", want, tags)
}
