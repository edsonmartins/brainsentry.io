package service

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

type testAnswer struct {
	Question   string   `json:"question"`
	Confidence float64  `json:"confidence"`
	Tags       []string `json:"tags,omitempty"`
}

func TestStructuredOutput_HappyPath(t *testing.T) {
	mock := &mockLLMProvider{
		name:     "t",
		response: `{"question":"What is Go?","confidence":0.95,"tags":["language","compiled"]}`,
	}

	var a testAnswer
	err := StructuredOutput[testAnswer](context.Background(), mock, "Tell me about Go.", &a, StructuredOutputOption{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Question != "What is Go?" {
		t.Errorf("unexpected question: %q", a.Question)
	}
	if a.Confidence != 0.95 {
		t.Errorf("unexpected confidence: %f", a.Confidence)
	}
	if len(a.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(a.Tags))
	}
}

func TestStructuredOutput_RetryOnInvalidJSON(t *testing.T) {
	mock := &selfCorrectingMock{
		responses: []string{
			"not json",
			`{"question":"q","confidence":0.5}`,
		},
	}

	var a testAnswer
	err := StructuredOutput[testAnswer](context.Background(), mock, "prompt", &a, StructuredOutputOption{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Question != "q" {
		t.Errorf("unexpected question: %q", a.Question)
	}
	if mock.callCount != 2 {
		t.Errorf("expected 2 calls (retry), got %d", mock.callCount)
	}
}

func TestStructuredOutput_NilLLM(t *testing.T) {
	var a testAnswer
	err := StructuredOutput[testAnswer](context.Background(), nil, "x", &a, StructuredOutputOption{})
	if err == nil {
		t.Error("expected error when llm is nil")
	}
}

func TestStructuredOutput_NilResult(t *testing.T) {
	mock := &mockLLMProvider{name: "t", response: "{}"}
	err := StructuredOutput[testAnswer](context.Background(), mock, "x", nil, StructuredOutputOption{})
	if err == nil {
		t.Error("expected error when result pointer is nil")
	}
}

func TestStructuredOutput_ExhaustsRetries(t *testing.T) {
	mock := &mockLLMProvider{name: "t", response: "never valid"}

	var a testAnswer
	err := StructuredOutput[testAnswer](context.Background(), mock, "x", &a, StructuredOutputOption{MaxRetries: 1})
	if err == nil {
		t.Error("expected error after exhausting retries")
	}
	// 1 initial + 1 retry = 2 calls
	if mock.callCount != 2 {
		t.Errorf("expected 2 calls, got %d", mock.callCount)
	}
}

func TestDescribeSchema_Struct(t *testing.T) {
	schema := describeSchema(reflect.TypeOf(testAnswer{}))

	if !strings.Contains(schema, `"question": string`) {
		t.Errorf("schema missing question field:\n%s", schema)
	}
	if !strings.Contains(schema, `"confidence": number`) {
		t.Errorf("schema missing confidence field:\n%s", schema)
	}
	if !strings.Contains(schema, `(optional)`) {
		t.Errorf("schema missing optional marker for tags:\n%s", schema)
	}
}

func TestDescribeSchema_PrimitiveTypes(t *testing.T) {
	tests := []struct {
		value    any
		expected string
	}{
		{"", "string"},
		{0, "integer"},
		{0.0, "number"},
		{false, "boolean"},
	}
	for _, tt := range tests {
		got := describeSchema(reflect.TypeOf(tt.value))
		if got != tt.expected {
			t.Errorf("describeSchema(%T): expected %q, got %q", tt.value, tt.expected, got)
		}
	}
}

func TestDescribeSchema_Slice(t *testing.T) {
	got := describeSchema(reflect.TypeOf([]string{}))
	if !strings.HasPrefix(got, "array of") {
		t.Errorf("expected 'array of ...', got %q", got)
	}
}
