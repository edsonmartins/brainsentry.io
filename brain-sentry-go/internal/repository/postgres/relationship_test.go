package postgres

import (
	"testing"
)

func TestRelationshipRepository_New(t *testing.T) {
	repo := NewRelationshipRepository(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
	if repo.pool != nil {
		t.Error("expected nil pool when initialized with nil")
	}
}
