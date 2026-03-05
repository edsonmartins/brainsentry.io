package service

import (
	"testing"
	"time"
)

func TestJWTService_GenerateAndValidate(t *testing.T) {
	svc := NewJWTService("test-secret-key-123", 1*time.Hour)

	token, err := svc.GenerateToken("user-1", "test@example.com", "tenant-1", []string{"USER", "ADMIN"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Errorf("expected UserID 'user-1', got '%s'", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.TenantID != "tenant-1" {
		t.Errorf("expected TenantID 'tenant-1', got '%s'", claims.TenantID)
	}
	if len(claims.Roles) != 2 || claims.Roles[0] != "USER" || claims.Roles[1] != "ADMIN" {
		t.Errorf("unexpected roles: %v", claims.Roles)
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret-key-123", -1*time.Hour) // negative duration = already expired

	token, err := svc.GenerateToken("user-1", "test@example.com", "tenant-1", nil)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = svc.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret-key-123", 1*time.Hour)

	_, err := svc.ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestJWTService_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-1", 1*time.Hour)
	svc2 := NewJWTService("secret-2", 1*time.Hour)

	token, err := svc1.GenerateToken("user-1", "test@example.com", "tenant-1", nil)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
