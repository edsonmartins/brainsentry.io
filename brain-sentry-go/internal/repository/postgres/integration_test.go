//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "brainsentry_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("failed to start container: %v\n", err)
		os.Exit(1)
	}
	defer container.Terminate(ctx)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/brainsentry_test?sslmode=disable", host, port.Port())

	testPool, err = NewPool(ctx, dsn, 5, 1)
	if err != nil {
		fmt.Printf("failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	// Run migrations
	if err := runMigrations(ctx, testPool); err != nil {
		fmt.Printf("failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migration, err := os.ReadFile("migrations/000001_init_schema.up.sql")
	if err != nil {
		return fmt.Errorf("reading migration: %w", err)
	}
	_, err = pool.Exec(ctx, string(migration))
	return err
}

func testContext() context.Context {
	return tenant.WithTenant(context.Background(), "test-tenant-integration")
}

func ensureTenantExists(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	repo := NewTenantRepository(testPool)
	_ = repo.Create(ctx, &domain.Tenant{
		ID:     id,
		Name:   "Integration Test " + id,
		Slug:   fmt.Sprintf("%s-%d", id, time.Now().UnixNano()),
		Active: true,
	})
}

// --- Tenant Repository Tests ---

func TestTenantRepository_CRUD(t *testing.T) {
	repo := NewTenantRepository(testPool)
	ctx := testContext()

	// Create
	tn := &domain.Tenant{
		Name:   "Test Tenant",
		Slug:   fmt.Sprintf("test-tenant-%d", time.Now().UnixNano()),
		Active: true,
	}
	err := repo.Create(ctx, tn)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if tn.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	// Get by ID
	found, err := repo.FindByID(ctx, tn.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "Test Tenant" {
		t.Errorf("expected name 'Test Tenant', got '%s'", found.Name)
	}

	// Update
	found.Name = "Updated Tenant"
	err = repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := repo.FindByID(ctx, tn.ID)
	if updated.Name != "Updated Tenant" {
		t.Errorf("expected updated name, got '%s'", updated.Name)
	}

	// List
	tenants, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tenants) == 0 {
		t.Error("expected at least one tenant")
	}

	// Delete
	err = repo.Delete(ctx, tn.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// --- Memory Repository Tests ---

func TestMemoryRepository_CRUD(t *testing.T) {
	repo := NewMemoryRepository(testPool)
	ctx := testContext()

	// Ensure tenant exists
	ensureTenantExists(t, ctx, "test-tenant-integration")

	// Create
	m := &domain.Memory{
		Content:    "Go is a statically typed, compiled programming language.",
		Summary:    "About Go language",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceImportant,
		Tags:       []string{"golang", "programming"},
	}

	err := repo.Create(ctx, m)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if m.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	// Find by ID
	found, err := repo.FindByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Content != m.Content {
		t.Errorf("content mismatch")
	}
	if found.Category != domain.CategoryKnowledge {
		t.Errorf("expected KNOWLEDGE category, got %s", found.Category)
	}

	// Update
	found.Summary = "Updated summary about Go"
	found.Version++
	err = repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := repo.FindByID(ctx, m.ID)
	if updated.Summary != "Updated summary about Go" {
		t.Errorf("summary not updated")
	}

	// List
	memories, total, err := repo.List(ctx, 0, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total == 0 || len(memories) == 0 {
		t.Error("expected at least one memory")
	}

	// Full text search
	results, err := repo.FullTextSearch(ctx, "Go programming", 5)
	if err != nil {
		t.Fatalf("FullTextSearch failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search results")
	}

	// Find by category
	byCategory, err := repo.FindByCategory(ctx, domain.CategoryKnowledge)
	if err != nil {
		t.Fatalf("FindByCategory failed: %v", err)
	}
	if len(byCategory) == 0 {
		t.Error("expected memories in KNOWLEDGE category")
	}

	// Feedback
	err = repo.RecordFeedback(ctx, m.ID, true)
	if err != nil {
		t.Fatalf("RecordFeedback failed: %v", err)
	}

	feedbackMem, _ := repo.FindByID(ctx, m.ID)
	if feedbackMem.HelpfulCount != 1 {
		t.Errorf("expected helpful_count=1, got %d", feedbackMem.HelpfulCount)
	}

	// Delete
	err = repo.Delete(ctx, m.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.FindByID(ctx, m.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestMemoryRepository_PreservesFieldsAndTenantIsolation(t *testing.T) {
	repo := NewMemoryRepository(testPool)
	ctxTenant1 := tenant.WithTenant(context.Background(), "tenant-integrity-a")
	ctxTenant2 := tenant.WithTenant(context.Background(), "tenant-integrity-b")

	ensureTenantExists(t, ctxTenant1, "tenant-integrity-a")
	ensureTenantExists(t, ctxTenant2, "tenant-integrity-b")

	validFrom := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Second)
	validTo := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	metadata := map[string]any{
		"origin": "integration-test",
		"scope":  "memory-integrity",
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}

	m := &domain.Memory{
		Content:             "Memory with metadata, tags and provenance",
		Summary:             "Integrity memory",
		Category:            domain.CategoryKnowledge,
		Importance:          domain.ImportanceCritical,
		Tags:                []string{"integrity", "tenant-a"},
		Metadata:            metadataJSON,
		SourceType:          "manual",
		SourceReference:     "integration-test",
		CreatedBy:           "qa@brainsentry.io",
		CodeExample:         "fmt.Println(\"ok\")",
		ProgrammingLanguage: "go",
		MemoryType:          domain.MemoryTypeSemantic,
		EmotionalWeight:     0.75,
		SimHash:             "abc123def456",
		ValidFrom:           &validFrom,
		ValidTo:             &validTo,
		DecayRate:           0.005,
	}

	if err := repo.Create(ctxTenant1, m); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(ctxTenant1, m.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.TenantID != "tenant-integrity-a" {
		t.Fatalf("expected tenant-integrity-a, got %s", found.TenantID)
	}
	if len(found.Tags) != 2 || found.Tags[0] != "integrity" || found.Tags[1] != "tenant-a" {
		t.Fatalf("tags not persisted correctly: %#v", found.Tags)
	}
	if found.SourceType != "manual" || found.SourceReference != "integration-test" {
		t.Fatalf("provenance not preserved: %s / %s", found.SourceType, found.SourceReference)
	}
	if found.SimHash != "abc123def456" {
		t.Fatalf("simhash not preserved: %s", found.SimHash)
	}
	if found.EmotionalWeight != 0.75 {
		t.Fatalf("expected emotional weight 0.75, got %f", found.EmotionalWeight)
	}
	if found.ValidFrom == nil || !found.ValidFrom.Equal(validFrom) {
		t.Fatalf("validFrom not preserved: %#v", found.ValidFrom)
	}
	if found.ValidTo == nil || !found.ValidTo.Equal(validTo) {
		t.Fatalf("validTo not preserved: %#v", found.ValidTo)
	}

	var decoded map[string]any
	if err := json.Unmarshal(found.Metadata, &decoded); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}
	if decoded["origin"] != "integration-test" || decoded["scope"] != "memory-integrity" {
		t.Fatalf("metadata not preserved: %#v", decoded)
	}

	if _, err := repo.FindByID(ctxTenant2, m.ID); err == nil {
		t.Fatal("expected tenant isolation to block FindByID from another tenant")
	}

	memoriesTenant2, totalTenant2, err := repo.List(ctxTenant2, 0, 20)
	if err != nil {
		t.Fatalf("List for tenant 2 failed: %v", err)
	}
	for _, memory := range memoriesTenant2 {
		if memory.ID == m.ID {
			t.Fatal("memory from tenant A leaked into tenant B list")
		}
	}
	if totalTenant2 < int64(len(memoriesTenant2)) {
		t.Fatalf("inconsistent pagination count for tenant 2: total=%d len=%d", totalTenant2, len(memoriesTenant2))
	}
}

func TestMemoryRepository_ExpireAndSupersedeLifecycle(t *testing.T) {
	repo := NewMemoryRepository(testPool)
	ctx := tenant.WithTenant(context.Background(), "tenant-lifecycle")
	ensureTenantExists(t, ctx, "tenant-lifecycle")

	expiredAt := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	activeUntil := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)

	expired := &domain.Memory{
		Content:    "Expired memory",
		Summary:    "Expired memory",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceMinor,
		ValidTo:    &expiredAt,
	}
	current := &domain.Memory{
		Content:    "Current memory",
		Summary:    "Current memory",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceImportant,
		ValidTo:    &activeUntil,
	}
	replacement := &domain.Memory{
		Content:    "Replacement memory",
		Summary:    "Replacement memory",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceCritical,
	}

	for _, memory := range []*domain.Memory{expired, current, replacement} {
		if err := repo.Create(ctx, memory); err != nil {
			t.Fatalf("Create failed for %s: %v", memory.Summary, err)
		}
	}

	rowsAffected, err := repo.ExpireStaleMemories(ctx)
	if err != nil {
		t.Fatalf("ExpireStaleMemories failed: %v", err)
	}
	if rowsAffected == 0 {
		t.Fatal("expected at least one expired memory to be soft-deleted")
	}
	if _, err := repo.FindByID(ctx, expired.ID); err == nil {
		t.Fatal("expired memory should not be retrievable after expiration")
	}

	if err := repo.SupersedeMemory(ctx, current.ID, replacement.ID); err != nil {
		t.Fatalf("SupersedeMemory failed: %v", err)
	}

	superseded, err := repo.FindByID(ctx, current.ID)
	if err != nil {
		t.Fatalf("FindByID current failed: %v", err)
	}
	if superseded.SupersededBy != replacement.ID {
		t.Fatalf("expected supersededBy=%s, got %s", replacement.ID, superseded.SupersededBy)
	}
	if superseded.ValidTo == nil {
		t.Fatal("expected superseded memory to receive validTo")
	}

	active, err := repo.FindActiveMemories(ctx, 10)
	if err != nil {
		t.Fatalf("FindActiveMemories failed: %v", err)
	}
	for _, memory := range active {
		if memory.ID == expired.ID {
			t.Fatal("expired memory leaked into active memories")
		}
		if memory.ID == current.ID {
			t.Fatal("superseded memory leaked into active memories")
		}
	}
}

// --- Audit Repository Tests ---

func TestAuditRepository_CRUD(t *testing.T) {
	repo := NewAuditRepository(testPool)
	ctx := testContext()

	log := &domain.AuditLog{
		EventType: "test_event",
		UserID:    "test-user",
		SessionID: "test-session",
		Outcome:   "success",
		TenantID:  "test-tenant-integration",
		Timestamp: time.Now(),
	}

	err := repo.Create(ctx, log)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List by tenant
	logs, err := repo.ListByTenant(ctx, 10)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(logs) == 0 {
		t.Error("expected at least one audit log")
	}

	// Find by event type
	byEvent, err := repo.FindByEventType(ctx, "test_event", 10)
	if err != nil {
		t.Fatalf("FindByEventType failed: %v", err)
	}
	if len(byEvent) == 0 {
		t.Error("expected audit logs for test_event")
	}

	// Find by session
	bySession, err := repo.FindBySessionID(ctx, "test-session", 10)
	if err != nil {
		t.Fatalf("FindBySessionID failed: %v", err)
	}
	if len(bySession) == 0 {
		t.Error("expected audit logs for test-session")
	}

	// Count by event type
	counts, err := repo.CountByEventType(ctx)
	if err != nil {
		t.Fatalf("CountByEventType failed: %v", err)
	}
	if counts["test_event"] == 0 {
		t.Error("expected count > 0 for test_event")
	}
}

// --- Note Repository Tests ---

func TestNoteRepository_CRUD(t *testing.T) {
	repo := NewNoteRepository(testPool)
	ctx := testContext()

	// Create note
	note := &domain.Note{
		SessionID:     "test-session",
		Type:          domain.NoteInsight,
		Title:         "Test Insight",
		Content:       "This is a test insight note",
		Category:      domain.NoteCategoryGeneric,
		Severity:      domain.SeverityMedium,
		AutoGenerated: true,
	}

	err := repo.CreateNote(ctx, note)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}
	if note.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	// List notes
	notes, err := repo.ListNotes(ctx, 10)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	if len(notes) == 0 {
		t.Error("expected at least one note")
	}

	// Create hindsight note
	hn := &domain.HindsightNote{
		SessionID:       "test-session",
		Title:           "Test Error",
		Severity:        domain.SeverityHigh,
		ErrorType:       "NullPointerException",
		ErrorMessage:    "null pointer at line 42",
		Resolution:      "Check for nil before accessing",
		OccurrenceCount: 1,
		AutoGenerated:   true,
	}

	err = repo.CreateHindsightNote(ctx, hn)
	if err != nil {
		t.Fatalf("CreateHindsightNote failed: %v", err)
	}

	// Search hindsight notes
	results, err := repo.SearchHindsightNotes(ctx, "null pointer", "test-tenant-integration", 5)
	if err != nil {
		t.Fatalf("SearchHindsightNotes failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search results for 'null pointer'")
	}

	// Find by session
	sessionNotes, err := repo.FindNotesBySession(ctx, "test-session", 10)
	if err != nil {
		t.Fatalf("FindNotesBySession failed: %v", err)
	}
	if len(sessionNotes) == 0 {
		t.Error("expected notes for test-session")
	}
}

// --- User Repository Tests ---

func TestUserRepository_CRUD(t *testing.T) {
	repo := NewUserRepository(testPool)
	ctx := testContext()

	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	user := &domain.User{
		Email:        email,
		Name:         "Test User",
		PasswordHash: "$2a$10$examplehashhere",
		TenantID:     "test-tenant-integration",
		Roles:        []string{"USER"},
		Active:       true,
		CreatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Find by email
	found, err := repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found.Name != "Test User" {
		t.Errorf("expected 'Test User', got '%s'", found.Name)
	}

	// Find by ID
	byID, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if byID.Email != email {
		t.Errorf("email mismatch")
	}
}
