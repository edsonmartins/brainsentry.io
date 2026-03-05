package service

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword verifies that HashPassword produces a valid bcrypt hash.
func TestHashPassword(t *testing.T) {
	password := "s3cur3P@ssw0rd"

	hash, err := HashPassword(password, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	// Verify the hash is valid bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Errorf("bcrypt comparison failed for produced hash: %v", err)
	}
}

// TestHashPassword_WrongPassword verifies that a wrong password does not match the hash.
func TestHashPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrong-password"))
	if err == nil {
		t.Error("expected mismatch for wrong password, but comparison succeeded")
	}
}

// TestHashPassword_EmptyPassword verifies that an empty password can be hashed.
func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashPassword failed for empty string: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash for empty password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("")); err != nil {
		t.Errorf("bcrypt comparison failed for empty password hash: %v", err)
	}
}

// TestHashPassword_DifferentCosts verifies that different bcrypt costs produce different hashes.
func TestHashPassword_DifferentCosts(t *testing.T) {
	password := "same-password"

	hash1, err := HashPassword(password, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashPassword cost=%d failed: %v", bcrypt.MinCost, err)
	}

	hash2, err := HashPassword(password, bcrypt.MinCost+1)
	if err != nil {
		t.Fatalf("HashPassword cost=%d failed: %v", bcrypt.MinCost+1, err)
	}

	// Hashes should be different (bcrypt includes salt)
	if hash1 == hash2 {
		t.Error("expected different hashes for different bcrypt costs")
	}

	// But both should still verify correctly
	if err := bcrypt.CompareHashAndPassword([]byte(hash1), []byte(password)); err != nil {
		t.Errorf("hash1 failed verification: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash2), []byte(password)); err != nil {
		t.Errorf("hash2 failed verification: %v", err)
	}
}

// TestHashPassword_Uniqueness verifies that the same password hashed twice produces different hashes (due to salt).
func TestHashPassword_Uniqueness(t *testing.T) {
	password := "same-password"

	hash1, err := HashPassword(password, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("first HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("second HashPassword failed: %v", err)
	}

	// bcrypt uses a random salt, so hashes must differ
	if hash1 == hash2 {
		t.Error("expected different hashes for the same password (bcrypt salt should differ)")
	}
}

// TestAuthService_RefreshToken_WithJWT tests RefreshToken using the JWT service directly,
// bypassing the database dependency of the AuthService.
func TestAuthService_RefreshToken_WithJWT(t *testing.T) {
	jwtSvc := NewJWTService("refresh-secret-key", 1*time.Hour)

	// Generate a refresh token directly
	refreshToken, err := jwtSvc.GenerateRefreshToken("user-42", "refresh@example.com", "tenant-99", []string{"ADMIN"})
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Validate the refresh token
	claims, err := jwtSvc.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("ValidateToken on refresh token failed: %v", err)
	}

	if claims.UserID != "user-42" {
		t.Errorf("expected UserID 'user-42', got '%s'", claims.UserID)
	}
	if claims.Email != "refresh@example.com" {
		t.Errorf("expected Email 'refresh@example.com', got '%s'", claims.Email)
	}
	if claims.TenantID != "tenant-99" {
		t.Errorf("expected TenantID 'tenant-99', got '%s'", claims.TenantID)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "ADMIN" {
		t.Errorf("unexpected roles: %v", claims.Roles)
	}
}

// TestAuthService_RefreshToken_Expiry verifies that refresh tokens expire after the configured duration.
func TestAuthService_RefreshToken_Expiry(t *testing.T) {
	// Use a negative duration to create an already-expired token
	jwtSvc := NewJWTService("secret", -1*time.Hour)

	refreshToken, err := jwtSvc.GenerateRefreshToken("user-1", "test@example.com", "tenant-1", nil)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	_, err = jwtSvc.ValidateToken(refreshToken)
	if err == nil {
		t.Fatal("expected error when validating expired refresh token")
	}
}

// TestErrConstants verifies that auth error variables are non-nil and distinct.
func TestErrConstants(t *testing.T) {
	if ErrInvalidCredentials == nil {
		t.Error("ErrInvalidCredentials should not be nil")
	}
	if ErrAccountDisabled == nil {
		t.Error("ErrAccountDisabled should not be nil")
	}
	if ErrUserNotFound == nil {
		t.Error("ErrUserNotFound should not be nil")
	}

	// Errors should be distinct
	if ErrInvalidCredentials == ErrAccountDisabled {
		t.Error("ErrInvalidCredentials and ErrAccountDisabled should be different errors")
	}
	if ErrInvalidCredentials == ErrUserNotFound {
		t.Error("ErrInvalidCredentials and ErrUserNotFound should be different errors")
	}
	if ErrAccountDisabled == ErrUserNotFound {
		t.Error("ErrAccountDisabled and ErrUserNotFound should be different errors")
	}
}

// TestErrConstants_Messages verifies that the error messages are meaningful.
func TestErrConstants_Messages(t *testing.T) {
	tests := []struct {
		err     error
		wantMsg string
	}{
		{ErrInvalidCredentials, "invalid email or password"},
		{ErrAccountDisabled, "account is disabled"},
		{ErrUserNotFound, "user not found"},
	}

	for _, tt := range tests {
		t.Run(tt.wantMsg, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("expected error message %q, got %q", tt.wantMsg, tt.err.Error())
			}
		})
	}
}
