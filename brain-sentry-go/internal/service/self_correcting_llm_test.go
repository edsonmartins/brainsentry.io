package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestSelfCorrectingLLM_ValidFirstAttempt(t *testing.T) {
	mock := &mockLLMProvider{name: "test", response: `{"name": "test", "value": 42}`}
	sc := NewSelfCorrectingLLM(mock, 2)

	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	err := sc.ChatJSON(context.Background(), []ChatMessage{{Role: "user", Content: "test"}}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test" || result.Value != 42 {
		t.Errorf("unexpected result: %+v", result)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
}

func TestSelfCorrectingLLM_RetryOnInvalidJSON(t *testing.T) {
	mock := &selfCorrectingMock{
		responses: []string{
			"not json",
			`{"name": "fixed"}`,
		},
	}
	sc := NewSelfCorrectingLLM(mock, 2)

	var result struct {
		Name string `json:"name"`
	}

	err := sc.ChatJSON(context.Background(), []ChatMessage{{Role: "user", Content: "test"}}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "fixed" {
		t.Errorf("expected 'fixed', got %s", result.Name)
	}
	if mock.callCount != 2 {
		t.Errorf("expected 2 calls, got %d", mock.callCount)
	}
}

func TestSelfCorrectingLLM_CustomValidation(t *testing.T) {
	mock := &selfCorrectingMock{
		responses: []string{
			`{"value": 0}`,  // fails validation
			`{"value": 10}`, // passes
		},
	}
	sc := NewSelfCorrectingLLM(mock, 2)

	validate := func(raw json.RawMessage) error {
		var v struct{ Value int }
		json.Unmarshal(raw, &v)
		if v.Value < 1 {
			return fmt.Errorf("value must be >= 1")
		}
		return nil
	}

	result, err := sc.ChatWithValidation(context.Background(),
		[]ChatMessage{{Role: "user", Content: "test"}}, validate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct{ Value int }
	json.Unmarshal(result, &parsed)
	if parsed.Value != 10 {
		t.Errorf("expected value 10, got %d", parsed.Value)
	}
}

func TestSelfCorrectingLLM_AllAttemptsFail(t *testing.T) {
	mock := &mockLLMProvider{name: "test", response: "never valid json {{{{"}
	sc := NewSelfCorrectingLLM(mock, 1)

	_, err := sc.ChatWithValidation(context.Background(),
		[]ChatMessage{{Role: "user", Content: "test"}}, nil)
	if err == nil {
		t.Fatal("expected error when all attempts fail")
	}
}

func TestSelfCorrectingLLM_Name(t *testing.T) {
	mock := &mockLLMProvider{name: "openrouter"}
	sc := NewSelfCorrectingLLM(mock, 2)
	if sc.Name() != "self-correcting-openrouter" {
		t.Errorf("unexpected name: %s", sc.Name())
	}
}
