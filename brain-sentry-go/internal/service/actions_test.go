package service

import (
	"context"
	"testing"
	"time"
)

func TestActionService_CreateAndGet(t *testing.T) {
	svc := NewActionService()
	ctx := context.Background()

	action, err := svc.CreateAction(ctx, "Fix bug", "Fix the login bug", "agent-1", 5, []string{"bug"}, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Status != ActionPending {
		t.Errorf("expected pending, got %s", action.Status)
	}

	got, err := svc.GetAction(ctx, action.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Fix bug" {
		t.Errorf("expected 'Fix bug', got %s", got.Title)
	}
}

func TestActionService_UpdateStatus(t *testing.T) {
	svc := NewActionService()
	ctx := context.Background()

	action, _ := svc.CreateAction(ctx, "Task", "desc", "agent", 3, nil, "", nil)
	updated, err := svc.UpdateStatus(ctx, action.ID, ActionCompleted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != ActionCompleted {
		t.Errorf("expected completed, got %s", updated.Status)
	}
}

func TestActionService_DependencyPropagation(t *testing.T) {
	svc := NewActionService()
	ctx := context.Background()

	parent, _ := svc.CreateAction(ctx, "Parent", "parent task", "agent", 5, nil, "", nil)
	child, _ := svc.CreateAction(ctx, "Child", "depends on parent", "agent", 3, nil, "", []string{parent.ID})

	// Child should be pending initially
	if child.Status != ActionPending {
		t.Errorf("expected child pending, got %s", child.Status)
	}

	// Block child
	svc.UpdateStatus(ctx, child.ID, ActionBlocked)

	// Complete parent — should unblock child
	svc.UpdateStatus(ctx, parent.ID, ActionCompleted)

	got, _ := svc.GetAction(ctx, child.ID)
	if got.Status != ActionPending {
		t.Errorf("expected child unblocked to pending, got %s", got.Status)
	}
}

func TestActionService_Lease(t *testing.T) {
	svc := NewActionService()
	ctx := context.Background()

	action, _ := svc.CreateAction(ctx, "Task", "desc", "agent", 3, nil, "", nil)

	lease, err := svc.AcquireLease(ctx, action.ID, "agent-1", 10*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lease.HeldBy != "agent-1" {
		t.Errorf("expected heldBy agent-1, got %s", lease.HeldBy)
	}

	// Action should be in_progress
	got, _ := svc.GetAction(ctx, action.ID)
	if got.Status != ActionInProgress {
		t.Errorf("expected in_progress, got %s", got.Status)
	}

	// Another agent should fail
	_, err = svc.AcquireLease(ctx, action.ID, "agent-2", 10*time.Minute)
	if err == nil {
		t.Error("expected error when lease already held")
	}

	// Release
	err = svc.ReleaseLease(ctx, action.ID, "agent-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ = svc.GetAction(ctx, action.ID)
	if got.Status != ActionCompleted {
		t.Errorf("expected completed after release, got %s", got.Status)
	}
}

func TestActionService_LeaseExpiry(t *testing.T) {
	svc := NewActionService()
	ctx := context.Background()

	action, _ := svc.CreateAction(ctx, "Task", "desc", "agent", 3, nil, "", nil)

	// Acquire with minimum TTL then manually expire
	svc.AcquireLease(ctx, action.ID, "agent-1", time.Minute)

	// Manually expire the lease
	svc.mu.Lock()
	svc.leases[action.ID].ExpiresAt = time.Now().Add(-time.Second)
	svc.mu.Unlock()

	cleaned := svc.CleanupExpiredLeases(ctx)
	if cleaned != 1 {
		t.Errorf("expected 1 cleaned, got %d", cleaned)
	}

	got, _ := svc.GetAction(ctx, action.ID)
	if got.Status != ActionPending {
		t.Errorf("expected pending after expiry, got %s", got.Status)
	}
}

func TestActionService_ListWithFilter(t *testing.T) {
	svc := NewActionService()
	ctx := context.Background()

	svc.CreateAction(ctx, "A", "", "agent", 1, nil, "", nil)
	svc.CreateAction(ctx, "B", "", "agent", 1, nil, "", nil)
	b, _ := svc.CreateAction(ctx, "C", "", "agent", 1, nil, "", nil)
	svc.UpdateStatus(ctx, b.ID, ActionCompleted)

	all := svc.ListActions(ctx, nil)
	if len(all) != 3 {
		t.Errorf("expected 3 actions, got %d", len(all))
	}

	pending := ActionPending
	filtered := svc.ListActions(ctx, &pending)
	if len(filtered) != 2 {
		t.Errorf("expected 2 pending, got %d", len(filtered))
	}
}
