package service

import (
	"context"
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

func TestDefaultSessionConfig(t *testing.T) {
	cfg := DefaultSessionConfig()
	if cfg.DefaultTTL != 2*time.Hour {
		t.Errorf("expected 2h TTL, got %v", cfg.DefaultTTL)
	}
	if cfg.MaxIdleTime != 30*time.Minute {
		t.Errorf("expected 30m idle, got %v", cfg.MaxIdleTime)
	}
	if cfg.CleanupInterval != 5*time.Minute {
		t.Errorf("expected 5m cleanup, got %v", cfg.CleanupInterval)
	}
}

func TestSessionService_CreateSession(t *testing.T) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	ctx := tenant.WithTenant(context.Background(), "test-tenant")

	session := svc.CreateSession(ctx, "user-1")
	if session.ID == "" {
		t.Fatal("expected session ID")
	}
	if session.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", session.UserID)
	}
	if session.TenantID != "test-tenant" {
		t.Errorf("expected test-tenant, got %s", session.TenantID)
	}
	if session.Status != domain.SessionActive {
		t.Errorf("expected ACTIVE, got %s", session.Status)
	}
	if !session.IsActive() {
		t.Error("expected session to be active")
	}
}

func TestSessionService_GetSession(t *testing.T) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	ctx := tenant.WithTenant(context.Background(), "t1")
	session := svc.CreateSession(ctx, "u1")

	got, err := svc.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != session.ID {
		t.Errorf("expected %s, got %s", session.ID, got.ID)
	}
}

func TestSessionService_GetSession_NotFound(t *testing.T) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	_, err := svc.GetSession(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSessionService_TouchSession(t *testing.T) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	ctx := tenant.WithTenant(context.Background(), "t1")
	session := svc.CreateSession(ctx, "u1")

	originalActivity := session.LastActivityAt
	time.Sleep(1 * time.Millisecond) // ensure time difference

	if err := svc.TouchSession(ctx, session.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetSession(ctx, session.ID)
	if !got.LastActivityAt.After(originalActivity) {
		t.Error("expected LastActivityAt to be updated")
	}
}

func TestSessionService_EndSession(t *testing.T) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	ctx := tenant.WithTenant(context.Background(), "t1")
	session := svc.CreateSession(ctx, "u1")

	if err := svc.EndSession(ctx, session.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetSession(ctx, session.ID)
	if got.Status != domain.SessionCompleted {
		t.Errorf("expected COMPLETED, got %s", got.Status)
	}
	if got.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

func TestSessionService_ListActiveSessions(t *testing.T) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	ctx := tenant.WithTenant(context.Background(), "t1")
	ctx2 := tenant.WithTenant(context.Background(), "t2")

	svc.CreateSession(ctx, "u1")
	svc.CreateSession(ctx, "u2")
	svc.CreateSession(ctx2, "u3")

	active := svc.ListActiveSessions(ctx)
	if len(active) != 2 {
		t.Errorf("expected 2 active sessions for t1, got %d", len(active))
	}
}

func TestSessionService_IncrementCounters(t *testing.T) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	session := svc.CreateSession(context.Background(), "u1")

	svc.IncrementMemoryCount(session.ID)
	svc.IncrementMemoryCount(session.ID)
	svc.IncrementInterceptionCount(session.ID)

	got, _ := svc.GetSession(context.Background(), session.ID)
	if got.MemoryCount != 2 {
		t.Errorf("expected 2 memories, got %d", got.MemoryCount)
	}
	if got.InterceptionCount != 1 {
		t.Errorf("expected 1 interception, got %d", got.InterceptionCount)
	}
}

func TestSession_IsExpired(t *testing.T) {
	s := &domain.Session{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		Status:    domain.SessionActive,
	}
	if !s.IsExpired() {
		t.Error("expected session to be expired")
	}
}

func TestSession_IsActive(t *testing.T) {
	s := &domain.Session{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Status:    domain.SessionActive,
	}
	if !s.IsActive() {
		t.Error("expected session to be active")
	}

	s.Status = domain.SessionCompleted
	if s.IsActive() {
		t.Error("completed session should not be active")
	}
}
