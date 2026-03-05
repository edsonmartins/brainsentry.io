package tenant

import (
	"context"
	"testing"
)

func TestWithTenantAndFromContext(t *testing.T) {
	ctx := context.Background()
	tenantID := "test-tenant-123"

	ctx = WithTenant(ctx, tenantID)
	result := FromContext(ctx)

	if result != tenantID {
		t.Errorf("expected '%s', got '%s'", tenantID, result)
	}
}

func TestFromContext_Default(t *testing.T) {
	ctx := context.Background()
	result := FromContext(ctx)

	if result != DefaultTenantID {
		t.Errorf("expected default tenant '%s', got '%s'", DefaultTenantID, result)
	}
}

func TestFromContext_Empty(t *testing.T) {
	ctx := WithTenant(context.Background(), "")
	result := FromContext(ctx)

	// Empty string should fall back to default
	if result != DefaultTenantID {
		t.Errorf("expected default tenant for empty string, got '%s'", result)
	}
}

func TestMultipleTenants(t *testing.T) {
	ctx1 := WithTenant(context.Background(), "tenant-1")
	ctx2 := WithTenant(context.Background(), "tenant-2")

	if FromContext(ctx1) != "tenant-1" {
		t.Error("ctx1 should have tenant-1")
	}
	if FromContext(ctx2) != "tenant-2" {
		t.Error("ctx2 should have tenant-2")
	}
}
