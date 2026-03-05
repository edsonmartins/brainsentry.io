package cache

import (
	"testing"
)

// TestNewRedisCache_InvalidAddress verifies that NewRedisCache returns nil
// when Redis is unreachable (non-fatal behaviour).
func TestNewRedisCache_InvalidAddress(t *testing.T) {
	// Port 1 is almost certainly not listening anywhere; the Ping inside
	// NewRedisCache will time out / refuse and the function must return nil.
	c := NewRedisCache("127.0.0.1:1", "", 0)
	if c != nil {
		// Clean up the connection that somehow succeeded (shouldn't happen).
		_ = c.Close()
		t.Fatal("expected NewRedisCache to return nil for unreachable address")
	}
}

// TestNewRedisCache_EmptyAddress verifies that an empty address also returns nil.
func TestNewRedisCache_EmptyAddress(t *testing.T) {
	c := NewRedisCache("", "", 0)
	if c != nil {
		_ = c.Close()
		t.Fatal("expected NewRedisCache to return nil for empty address")
	}
}

// TestNewRedisCache_InvalidHostname verifies that a hostname that cannot be
// resolved also causes the constructor to return nil.
func TestNewRedisCache_InvalidHostname(t *testing.T) {
	c := NewRedisCache("this-host-does-not-exist.invalid:6379", "", 0)
	if c != nil {
		_ = c.Close()
		t.Fatal("expected NewRedisCache to return nil for unresolvable hostname")
	}
}

// TestEmbeddingKey verifies the cache-key helper produces the expected format.
func TestEmbeddingKey(t *testing.T) {
	tests := []struct {
		tenantID string
		text     string
		wantPrefix string
	}{
		{"tenant-1", "hello world", "emb:tenant-1:"},
		{"", "hello", "emb::"},
		{"t", "", "emb:t:"},
	}

	for _, tc := range tests {
		key := EmbeddingKey(tc.tenantID, tc.text)
		if len(key) == 0 {
			t.Errorf("EmbeddingKey(%q, %q) returned empty string", tc.tenantID, tc.text)
		}
		// The key must start with "emb:<tenantID>:"
		if len(key) < len(tc.wantPrefix) || key[:len(tc.wantPrefix)] != tc.wantPrefix {
			t.Errorf("EmbeddingKey(%q, %q) = %q, want prefix %q",
				tc.tenantID, tc.text, key, tc.wantPrefix)
		}
	}
}

// TestEmbeddingKey_Deterministic verifies that the same inputs always produce
// the same key.
func TestEmbeddingKey_Deterministic(t *testing.T) {
	k1 := EmbeddingKey("tenant-abc", "some text")
	k2 := EmbeddingKey("tenant-abc", "some text")
	if k1 != k2 {
		t.Errorf("EmbeddingKey is not deterministic: %q != %q", k1, k2)
	}
}

// TestEmbeddingKey_DifferentInputsProduceDifferentKeys verifies that distinct
// inputs do not collide.
func TestEmbeddingKey_DifferentInputsProduceDifferentKeys(t *testing.T) {
	k1 := EmbeddingKey("tenant-1", "text-a")
	k2 := EmbeddingKey("tenant-1", "text-b")
	k3 := EmbeddingKey("tenant-2", "text-a")

	if k1 == k2 {
		t.Errorf("expected different keys for different text: both are %q", k1)
	}
	if k1 == k3 {
		t.Errorf("expected different keys for different tenant: both are %q", k1)
	}
}
