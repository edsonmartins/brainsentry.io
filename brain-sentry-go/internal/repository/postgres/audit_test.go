package postgres

import (
	"testing"
)

func TestAuditColumns_HasCorrectCount(t *testing.T) {
	// 18 columns in auditColumns
	expected := 18
	count := countColumns(auditColumns)
	if count != expected {
		t.Errorf("expected %d audit columns, got %d", expected, count)
	}
}

func TestAuditColumns_ContainsRequiredFields(t *testing.T) {
	required := []string{"id", "event_type", "timestamp", "user_id", "session_id", "tenant_id"}
	for _, field := range required {
		if !containsSubstring(auditColumns, field) {
			t.Errorf("auditColumns should contain %s", field)
		}
	}
}

func TestToJSON_Nil(t *testing.T) {
	result := ToJSON(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestToJSON_ValidMap(t *testing.T) {
	data := map[string]any{"action": "test"}
	result := ToJSON(data)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if string(result) != `{"action":"test"}` {
		t.Errorf("unexpected JSON: %s", result)
	}
}
