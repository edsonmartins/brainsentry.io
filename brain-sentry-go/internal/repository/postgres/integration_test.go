//go:build integration

package postgres

import (
	"context"
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
	tenantRepo := NewTenantRepository(testPool)
	tenantRepo.Create(ctx, &domain.Tenant{
		ID:     "test-tenant-integration",
		Name:   "Integration Test",
		Slug:   fmt.Sprintf("integration-%d", time.Now().UnixNano()),
		Active: true,
	})

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
