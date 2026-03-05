package service

import "testing"

func TestCleanJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"plain json",
			`{"key": "value"}`,
			`{"key": "value"}`,
		},
		{
			"markdown code block",
			"```json\n{\"key\": \"value\"}\n```",
			`{"key": "value"}`,
		},
		{
			"markdown code block no lang",
			"```\n{\"key\": \"value\"}\n```",
			`{"key": "value"}`,
		},
		{
			"with surrounding text",
			"Here is the result:\n{\"key\": \"value\"}\nDone.",
			`{"key": "value"}`,
		},
		{
			"nested braces",
			`{"outer": {"inner": "value"}}`,
			`{"outer": {"inner": "value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSON(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}
