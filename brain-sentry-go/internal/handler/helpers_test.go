package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(rr, http.StatusOK, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp["key"] != "value" {
		t.Errorf("expected 'value', got '%s'", resp["key"])
	}
}

func TestWriteDomainErrorMappings(t *testing.T) {
	tests := []struct {
		err      error
		status   int
		category string
		code     string
	}{
		{domain.NewNotFoundError("missing"), http.StatusNotFound, "NOT_FOUND", "not_found"},
		{domain.NewValidationError("invalid"), http.StatusBadRequest, "VALIDATION", "validation"},
		{domain.NewConflictError("conflict"), http.StatusConflict, "CONFLICT", "conflict"},
		{&domain.DomainError{Err: domain.ErrAlreadyExists, Message: "exists", Code: "already_exists"}, http.StatusConflict, "CONFLICT", "already_exists"},
		{&domain.DomainError{Err: domain.ErrUnauthorized, Message: "login", Code: "unauthorized"}, http.StatusUnauthorized, "AUTH", "unauthorized"},
		{&domain.DomainError{Err: domain.ErrForbidden, Message: "denied", Code: "forbidden"}, http.StatusForbidden, "AUTH", "forbidden"},
		{&domain.DomainError{Err: domain.ErrRateLimited, Message: "slow down", Code: "rate_limited"}, http.StatusTooManyRequests, "INTERNAL", "rate_limited"},
		{domain.NewInternalError("broken"), http.StatusInternalServerError, "INTERNAL", "internal"},
	}

	for _, tc := range tests {
		rr := httptest.NewRecorder()
		writeDomainError(rr, tc.err)
		var response dto.ErrorResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if rr.Code != tc.status || response.ErrorCategory != tc.category || response.ErrorCode != tc.code {
			t.Fatalf("err=%v response=%+v status=%d", tc.err, response, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	writeDomainError(rr, errors.New("opaque"))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("opaque errors must map to 500, got %d", rr.Code)
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeError(rr, http.StatusBadRequest, "something went wrong")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var resp dto.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.Status != 400 {
		t.Errorf("expected status 400 in body, got %d", resp.Status)
	}
	if resp.Message != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%s'", resp.Message)
	}
	if resp.Error != "Bad Request" {
		t.Errorf("expected error 'Bad Request', got '%s'", resp.Error)
	}
}

func TestWriteJSON_StatusCodes(t *testing.T) {
	codes := []int{http.StatusCreated, http.StatusAccepted, http.StatusNoContent}
	for _, code := range codes {
		rr := httptest.NewRecorder()
		writeJSON(rr, code, map[string]string{})
		if rr.Code != code {
			t.Errorf("expected status %d, got %d", code, rr.Code)
		}
	}
}
