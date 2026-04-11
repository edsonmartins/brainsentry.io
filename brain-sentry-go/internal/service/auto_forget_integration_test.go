//go:build integration

package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

func setupAutoForgetIntegration(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()
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
		t.Fatalf("starting postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("getting container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("getting container port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/brainsentry_test?sslmode=disable", host, port.Port())
	pool, err := postgres.NewPool(ctx, dsn, 5, 1)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("connecting to postgres: %v", err)
	}

	if err := runAutoForgetIntegrationMigrations(ctx, pool); err != nil {
		pool.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("running migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
		_ = container.Terminate(context.Background())
	}
	return pool, cleanup
}

func runAutoForgetIntegrationMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("resolving current filename")
	}
	migrationPath := filepath.Join(filepath.Dir(filename), "..", "repository", "postgres", "migrations", "000001_init_schema.up.sql")
	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("reading migration: %w", err)
	}
	_, err = pool.Exec(ctx, string(migration))
	return err
}

func ensureAutoForgetTenant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id string) {
	t.Helper()
	repo := postgres.NewTenantRepository(pool)
	_ = repo.Create(ctx, &domain.Tenant{
		ID:     id,
		Name:   "Auto Forget Integration " + id,
		Slug:   fmt.Sprintf("%s-%d", id, time.Now().UnixNano()),
		Active: true,
	})
}

func TestAutoForgetService_PostgresDryRunDoesNotMutate(t *testing.T) {
	pool, cleanup := setupAutoForgetIntegration(t)
	defer cleanup()

	ctx := tenant.WithTenant(context.Background(), "tenant-auto-forget-dry-run")
	ensureAutoForgetTenant(t, ctx, pool, "tenant-auto-forget-dry-run")

	repo := postgres.NewMemoryRepository(pool)
	expiredAt := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	expired := &domain.Memory{
		Content:    "expired dry-run memory",
		Summary:    "expired dry-run memory",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceMinor,
		ValidTo:    &expiredAt,
	}
	if err := repo.Create(ctx, expired); err != nil {
		t.Fatalf("creating expired memory: %v", err)
	}

	svc := NewAutoForgetService(repo, nil, AutoForgetConfig{
		TTLEnabled:             true,
		ContradictionEnabled:   false,
		LowValueEnabled:        false,
		ContradictionThreshold: 0.9,
		MaxDeletesPerRun:       10,
	})

	result, err := svc.Run(ctx, true)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.TTLExpired != 1 || result.TotalDeleted != 1 {
		t.Fatalf("expected dry-run to report one candidate, got ttl=%d total=%d", result.TTLExpired, result.TotalDeleted)
	}
	if _, err := repo.FindByID(ctx, expired.ID); err != nil {
		t.Fatalf("dry-run should not delete expired memory: %v", err)
	}
}

func TestAutoForgetService_PostgresRealRunMutatesTTLContradictionAndLowValue(t *testing.T) {
	pool, cleanup := setupAutoForgetIntegration(t)
	defer cleanup()

	ctx := tenant.WithTenant(context.Background(), "tenant-auto-forget-real-run")
	ensureAutoForgetTenant(t, ctx, pool, "tenant-auto-forget-real-run")

	repo := postgres.NewMemoryRepository(pool)
	now := time.Now().UTC().Truncate(time.Second)
	expiredAt := now.Add(-time.Hour)

	expired := &domain.Memory{
		Content:    "expired real-run memory",
		Summary:    "expired real-run memory",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceMinor,
		ValidTo:    &expiredAt,
	}
	duplicateOlder := &domain.Memory{
		Content:    "use transactional outbox for reliable event publishing",
		Summary:    "outbox older",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceImportant,
	}
	duplicateNewer := &domain.Memory{
		Content:    "use transactional outbox for reliable event publishing",
		Summary:    "outbox newer",
		Category:   domain.CategoryKnowledge,
		Importance: domain.ImportanceImportant,
	}
	lowValue := &domain.Memory{
		Content:    "low value memory that should age out",
		Summary:    "low value",
		Category:   domain.CategoryReference,
		Importance: domain.ImportanceMinor,
	}
	active := &domain.Memory{
		Content:    "active important memory should remain",
		Summary:    "active",
		Category:   domain.CategoryDecision,
		Importance: domain.ImportanceCritical,
	}

	for _, memory := range []*domain.Memory{expired, duplicateOlder, duplicateNewer, lowValue, active} {
		if err := repo.Create(ctx, memory); err != nil {
			t.Fatalf("creating %s: %v", memory.Summary, err)
		}
	}

	if _, err := pool.Exec(ctx, `UPDATE memories SET created_at = $1, updated_at = $1 WHERE id = $2`, now.AddDate(0, 0, -200), lowValue.ID); err != nil {
		t.Fatalf("backdating low-value memory: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE memories SET created_at = $1, updated_at = $1 WHERE id = $2`, now.Add(-2*time.Hour), duplicateOlder.ID); err != nil {
		t.Fatalf("backdating duplicate older memory: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE memories SET created_at = $1, updated_at = $1 WHERE id = $2`, now.Add(-time.Hour), duplicateNewer.ID); err != nil {
		t.Fatalf("backdating duplicate newer memory: %v", err)
	}

	svc := NewAutoForgetService(repo, nil, AutoForgetConfig{
		TTLEnabled:             true,
		ContradictionEnabled:   true,
		LowValueEnabled:        true,
		LowValueMaxAgeDays:     180,
		LowValueMaxImportance:  string(domain.ImportanceMinor),
		ContradictionThreshold: 1.0,
		MaxDeletesPerRun:       10,
	})

	result, err := svc.Run(ctx, false)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.TTLExpired != 1 || result.Contradictions != 1 || result.LowValue != 1 || result.TotalDeleted != 3 {
		t.Fatalf("unexpected result: %#v", result)
	}

	if _, err := repo.FindByID(ctx, expired.ID); err == nil {
		t.Fatal("expired memory should be soft-deleted")
	}
	if _, err := repo.FindByID(ctx, lowValue.ID); err == nil {
		t.Fatal("low-value memory should be soft-deleted")
	}

	superseded, err := repo.FindByID(ctx, duplicateOlder.ID)
	if err != nil {
		t.Fatalf("superseded older duplicate should remain retrievable: %v", err)
	}
	if superseded.SupersededBy != duplicateNewer.ID {
		t.Fatalf("expected older duplicate supersededBy=%s, got %s", duplicateNewer.ID, superseded.SupersededBy)
	}
	if superseded.ValidTo == nil {
		t.Fatal("expected superseded duplicate to receive validTo")
	}

	if _, err := repo.FindByID(ctx, duplicateNewer.ID); err != nil {
		t.Fatalf("newer duplicate should remain active: %v", err)
	}
	if _, err := repo.FindByID(ctx, active.ID); err != nil {
		t.Fatalf("active memory should remain: %v", err)
	}
}
