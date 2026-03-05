package mcp

import (
	"testing"
)

// TestNewMCPError verifies that NewMCPError constructs an RPCError with the
// correct code, message, and Data map containing both "errorCategory" and
// "errorType" keys.
func TestNewMCPError(t *testing.T) {
	const code = -32602
	const msg = "something went wrong"
	const cat = ErrCategoryValidation

	rpcErr := NewMCPError(code, msg, cat)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != code {
		t.Errorf("expected code %d, got %d", code, rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}

	dataMap, ok := rpcErr.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected Data to be map[string]any, got %T", rpcErr.Data)
	}
	if dataMap["errorCategory"] != cat {
		t.Errorf("expected errorCategory %q, got %v", cat, dataMap["errorCategory"])
	}
	if dataMap["errorType"] != cat {
		t.Errorf("expected errorType %q, got %v", cat, dataMap["errorType"])
	}
}

// TestErrValidation verifies the ErrValidation constructor.
func TestErrValidation(t *testing.T) {
	msg := "field 'content' is required"
	rpcErr := ErrValidation(msg)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != -32602 {
		t.Errorf("expected code -32602, got %d", rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}
	assertCategory(t, rpcErr, ErrCategoryValidation)
}

// TestErrNotFound verifies the ErrNotFound constructor.
func TestErrNotFound(t *testing.T) {
	msg := "memory with id abc123 not found"
	rpcErr := ErrNotFound(msg)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != -32602 {
		t.Errorf("expected code -32602, got %d", rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}
	assertCategory(t, rpcErr, ErrCategoryNotFound)
}

// TestErrAuthorization verifies the ErrAuthorization constructor.
func TestErrAuthorization(t *testing.T) {
	msg := "unauthorized access"
	rpcErr := ErrAuthorization(msg)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != -32600 {
		t.Errorf("expected code -32600, got %d", rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}
	assertCategory(t, rpcErr, ErrCategoryAuthorization)
}

// TestErrInternal verifies the ErrInternal constructor.
func TestErrInternal(t *testing.T) {
	msg := "database connection failed"
	rpcErr := ErrInternal(msg)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != -32603 {
		t.Errorf("expected code -32603, got %d", rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}
	assertCategory(t, rpcErr, ErrCategoryInternal)
}

// TestErrTenant verifies the ErrTenant constructor.
func TestErrTenant(t *testing.T) {
	msg := "tenant not configured"
	rpcErr := ErrTenant(msg)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != -32600 {
		t.Errorf("expected code -32600, got %d", rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}
	assertCategory(t, rpcErr, ErrCategoryTenant)
}

// TestErrRateLimit verifies the ErrRateLimit constructor.
func TestErrRateLimit(t *testing.T) {
	msg := "too many requests"
	rpcErr := ErrRateLimit(msg)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != -32600 {
		t.Errorf("expected code -32600, got %d", rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}
	assertCategory(t, rpcErr, ErrCategoryRateLimit)
}

// TestErrTimeout verifies the ErrTimeout constructor.
func TestErrTimeout(t *testing.T) {
	msg := "request timed out"
	rpcErr := ErrTimeout(msg)

	if rpcErr == nil {
		t.Fatal("expected non-nil RPCError")
	}
	if rpcErr.Code != -32603 {
		t.Errorf("expected code -32603, got %d", rpcErr.Code)
	}
	if rpcErr.Message != msg {
		t.Errorf("expected message %q, got %q", msg, rpcErr.Message)
	}
	assertCategory(t, rpcErr, ErrCategoryTimeout)
}

// TestErrorCategoryConstants verifies that all category constant values remain
// stable (guards against accidental renames/typos).
func TestErrorCategoryConstants(t *testing.T) {
	cases := []struct {
		name     string
		got      string
		expected string
	}{
		{"ErrCategoryValidation", ErrCategoryValidation, "validation"},
		{"ErrCategoryAuthorization", ErrCategoryAuthorization, "authorization"},
		{"ErrCategoryNotFound", ErrCategoryNotFound, "not_found"},
		{"ErrCategoryInternal", ErrCategoryInternal, "internal"},
		{"ErrCategoryTenant", ErrCategoryTenant, "tenant"},
		{"ErrCategoryRateLimit", ErrCategoryRateLimit, "rate_limit"},
		{"ErrCategoryTimeout", ErrCategoryTimeout, "timeout"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.got)
			}
		})
	}
}

// TestErrDataBothKeys verifies that every error constructor populates both
// "errorCategory" and "errorType" keys in the Data map.
func TestErrDataBothKeys(t *testing.T) {
	constructors := []struct {
		name string
		err  *RPCError
	}{
		{"ErrValidation", ErrValidation("v")},
		{"ErrNotFound", ErrNotFound("nf")},
		{"ErrAuthorization", ErrAuthorization("auth")},
		{"ErrInternal", ErrInternal("int")},
		{"ErrTenant", ErrTenant("ten")},
		{"ErrRateLimit", ErrRateLimit("rl")},
		{"ErrTimeout", ErrTimeout("to")},
	}

	for _, tc := range constructors {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatal("expected non-nil error")
			}
			dataMap, ok := tc.err.Data.(map[string]any)
			if !ok {
				t.Fatalf("expected Data to be map[string]any, got %T", tc.err.Data)
			}
			if _, exists := dataMap["errorCategory"]; !exists {
				t.Error("Data map missing key 'errorCategory'")
			}
			if _, exists := dataMap["errorType"]; !exists {
				t.Error("Data map missing key 'errorType'")
			}
			// Both keys must hold the same category string
			if dataMap["errorCategory"] != dataMap["errorType"] {
				t.Errorf("errorCategory %q != errorType %q",
					dataMap["errorCategory"], dataMap["errorType"])
			}
		})
	}
}

// assertCategory is a helper that checks the Data map of an RPCError contains
// the expected category under both "errorCategory" and "errorType" keys.
func assertCategory(t *testing.T, rpcErr *RPCError, expected string) {
	t.Helper()
	dataMap, ok := rpcErr.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected Data to be map[string]any, got %T", rpcErr.Data)
	}
	if dataMap["errorCategory"] != expected {
		t.Errorf("expected errorCategory %q, got %v", expected, dataMap["errorCategory"])
	}
	if dataMap["errorType"] != expected {
		t.Errorf("expected errorType %q, got %v", expected, dataMap["errorType"])
	}
}
