package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
