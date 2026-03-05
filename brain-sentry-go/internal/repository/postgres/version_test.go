package postgres

import (
	"testing"
)

func TestVersionRepository_New(t *testing.T) {
	repo := NewVersionRepository(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}
