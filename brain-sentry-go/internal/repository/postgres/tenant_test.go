package postgres

import (
	"testing"
)

func TestTenantRepository_New(t *testing.T) {
	repo := NewTenantRepository(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}
