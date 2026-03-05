package postgres

import (
	"testing"
)

func TestWebhookRepository_New(t *testing.T) {
	repo := NewWebhookRepository(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}
