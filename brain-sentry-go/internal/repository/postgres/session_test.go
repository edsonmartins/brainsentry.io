package postgres

import (
	"testing"
)

func TestSessionRepository_New(t *testing.T) {
	repo := NewSessionRepository(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}
