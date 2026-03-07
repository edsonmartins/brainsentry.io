package commands

import (
	"testing"
	"time"
)

func TestValidateTenantID_ValidUUID(t *testing.T) {
	a := &App{TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561"}
	if err := a.validateTenantID(); err != nil {
		t.Errorf("expected valid UUID, got error: %v", err)
	}
}

func TestValidateTenantID_InvalidUUID(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
	}{
		{"empty", ""},
		{"short", "abc"},
		{"no dashes", "a9f814d24dae41f3851b8aa3d4706561"},
		{"invalid chars", "g9f814d2-4dae-41f3-851b-8aa3d4706561"},
		{"too long", "a9f814d2-4dae-41f3-851b-8aa3d4706561-extra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{TenantID: tt.tenantID}
			if err := a.validateTenantID(); err == nil {
				t.Errorf("expected error for tenant ID %q", tt.tenantID)
			}
		})
	}
}

func TestNewContext_HasTimeout(t *testing.T) {
	a := &App{
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Timeout:  5 * time.Second,
	}
	ctx, cancel := a.newContext()
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > 5*time.Second {
		t.Errorf("expected ~5s remaining, got %v", remaining)
	}
}

func TestNewContext_DefaultTimeout(t *testing.T) {
	a := &App{
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
	}
	ctx, cancel := a.newContext()
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 25*time.Second || remaining > 30*time.Second {
		t.Errorf("expected ~30s remaining (default), got %v", remaining)
	}
}

func TestNewContext_CustomTimeout(t *testing.T) {
	a := &App{
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Timeout:  10 * time.Second,
	}
	ctx, cancel := a.newContext()
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 5*time.Second || remaining > 10*time.Second {
		t.Errorf("expected ~10s, got %v", remaining)
	}
}

func TestAppDefaults(t *testing.T) {
	a := &App{}
	if a.Timeout != 0 {
		t.Error("default Timeout should be zero (uses defaultTimeout)")
	}
	if a.Output != "" {
		t.Error("default Output should be empty string")
	}
}

func TestMaxImportFileSize(t *testing.T) {
	expected := int64(100 * 1024 * 1024)
	if maxImportFileSize != expected {
		t.Errorf("expected maxImportFileSize=%d, got %d", expected, maxImportFileSize)
	}
}
