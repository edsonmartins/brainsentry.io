package service

import (
	"strings"
	"testing"
)

func TestPrivacyStripping_PrivateTags(t *testing.T) {
	svc := NewPrivacyStrippingService()
	content := "Hello <private>my secret data</private> world"
	stripped, result := svc.Strip(content)

	if strings.Contains(stripped, "secret data") {
		t.Error("private tag content should be stripped")
	}
	if !strings.Contains(stripped, "Hello") {
		t.Error("non-private content should remain")
	}
	if result.ItemsRemoved == 0 {
		t.Error("expected items removed > 0")
	}
}

func TestPrivacyStripping_EnvVars(t *testing.T) {
	svc := NewPrivacyStrippingService()
	content := "Config: export API_KEY=sk-abc123def456 and DATABASE_URL=postgres://user:pass@host/db"
	stripped, result := svc.Strip(content)

	if strings.Contains(stripped, "sk-abc123def456") {
		t.Error("API key should be stripped")
	}
	if strings.Contains(stripped, "postgres://user:pass") {
		t.Error("database URL should be stripped")
	}
	if result.ItemsRemoved == 0 {
		t.Error("expected items removed")
	}
}

func TestPrivacyStripping_SecretPatterns(t *testing.T) {
	svc := NewPrivacyStrippingService()

	tests := []struct {
		name    string
		input   string
		shouldStrip bool
	}{
		{"GitHub PAT", "token: ghp_abcdefghijklmnopqrstuvwxyz1234567890", true},
		{"AWS key", "key: AKIAIOSFODNN7EXAMPLE", true},
		{"Slack bot", "xoxb-123456-789012-abcdef", true},
		{"Normal text", "This is just normal text without secrets", false},
	}

	for _, tt := range tests {
		stripped, result := svc.Strip(tt.input)
		if tt.shouldStrip && result.ItemsRemoved == 0 {
			t.Errorf("%s: expected stripping but nothing was stripped", tt.name)
		}
		if !tt.shouldStrip && stripped != tt.input {
			t.Errorf("%s: unexpected stripping of clean text", tt.name)
		}
	}
}

func TestPrivacyStripping_ContainsSensitive(t *testing.T) {
	svc := NewPrivacyStrippingService()

	if !svc.ContainsSensitive("<private>secret</private>") {
		t.Error("should detect private tags")
	}
	if !svc.ContainsSensitive("export API_KEY=mykey123456789012345") {
		t.Error("should detect env vars")
	}
	if svc.ContainsSensitive("This is clean text") {
		t.Error("should not flag clean text")
	}
}

func TestPrivacyStripping_StripBeforeStorage(t *testing.T) {
	svc := NewPrivacyStrippingService()

	clean := svc.StripBeforeStorage("Just normal text")
	if clean != "Just normal text" {
		t.Error("clean text should pass through unchanged")
	}

	dirty := svc.StripBeforeStorage("Here is <private>a secret</private>")
	if strings.Contains(dirty, "a secret") {
		t.Error("private content should be stripped for storage")
	}
}
