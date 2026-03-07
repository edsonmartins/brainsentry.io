package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintTable_Basic(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"ID", "Name"}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	printTable(&buf, headers, rows)
	out := buf.String()
	if !strings.Contains(out, "ID") {
		t.Error("expected header ID")
	}
	if !strings.Contains(out, "Alice") {
		t.Error("expected Alice in output")
	}
	if !strings.Contains(out, "Bob") {
		t.Error("expected Bob in output")
	}
}

func TestPrintTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	printTable(&buf, []string{"ID"}, nil)
	out := buf.String()
	if !strings.Contains(out, "ID") {
		t.Error("expected header even with no rows")
	}
}

func TestPrintJSON_Basic(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	if err := printJSON(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"key": "value"`) {
		t.Error("expected JSON key-value pair")
	}
}

func TestTruncate_Short(t *testing.T) {
	result := truncate("hello", 10)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestTruncate_Long(t *testing.T) {
	result := truncate("hello world", 8)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got '%s'", result)
	}
}

func TestTruncate_Exact(t *testing.T) {
	result := truncate("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestTruncate_VeryShortMax(t *testing.T) {
	result := truncate("hello world", 3)
	if result != "hel" {
		t.Errorf("expected 'hel', got '%s'", result)
	}
}
