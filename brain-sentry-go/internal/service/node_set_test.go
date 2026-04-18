package service

import (
	"encoding/json"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestNodeSet_GetMemorySets_Empty(t *testing.T) {
	svc := NewNodeSetService(nil)
	m := &domain.Memory{}
	if len(svc.GetMemorySets(m)) != 0 {
		t.Error("expected empty for memory without metadata")
	}
}

func TestNodeSet_GetMemorySets_WithSets(t *testing.T) {
	svc := NewNodeSetService(nil)
	raw, _ := json.Marshal(map[string]any{
		MetadataKeyBelongsToSets: []string{"project-x", "team-alpha"},
	})
	m := &domain.Memory{Metadata: raw}

	sets := svc.GetMemorySets(m)
	if len(sets) != 2 {
		t.Errorf("expected 2 sets, got %d", len(sets))
	}
}

func TestNodeSet_GetMemorySets_FromInterfaceSlice(t *testing.T) {
	svc := NewNodeSetService(nil)
	raw, _ := json.Marshal(map[string]any{
		MetadataKeyBelongsToSets: []any{"a", "b", 42, ""},
	})
	m := &domain.Memory{Metadata: raw}

	sets := svc.GetMemorySets(m)
	if len(sets) != 2 {
		t.Errorf("expected 2 string sets (skipping non-string/empty), got %d: %v", len(sets), sets)
	}
}

func TestNodeSet_FilterBySet(t *testing.T) {
	svc := NewNodeSetService(nil)

	rawA, _ := json.Marshal(map[string]any{
		MetadataKeyBelongsToSets: []string{"project-x"},
	})
	rawB, _ := json.Marshal(map[string]any{
		MetadataKeyBelongsToSets: []string{"team-alpha"},
	})

	memories := []domain.Memory{
		{ID: "1", Metadata: rawA},
		{ID: "2", Metadata: rawB},
		{ID: "3"},
	}

	filtered := svc.FilterBySet(memories, "project-x")
	if len(filtered) != 1 || filtered[0].ID != "1" {
		t.Errorf("expected only '1', got %+v", filtered)
	}
}

func TestNodeSet_FilterBySet_EmptyReturnsAll(t *testing.T) {
	svc := NewNodeSetService(nil)
	memories := []domain.Memory{{ID: "1"}, {ID: "2"}}

	filtered := svc.FilterBySet(memories, "")
	if len(filtered) != 2 {
		t.Errorf("empty filter should return all, got %d", len(filtered))
	}
}

func TestNodeSet_ListSets(t *testing.T) {
	svc := NewNodeSetService(nil)

	rawA, _ := json.Marshal(map[string]any{
		MetadataKeyBelongsToSets: []string{"a", "b"},
	})
	rawB, _ := json.Marshal(map[string]any{
		MetadataKeyBelongsToSets: []string{"b", "c"},
	})

	memories := []domain.Memory{
		{ID: "1", Metadata: rawA},
		{ID: "2", Metadata: rawB},
	}

	sets := svc.ListSets(memories)
	if len(sets) != 3 {
		t.Errorf("expected 3 unique sets, got %d: %v", len(sets), sets)
	}
}

func TestNormalizeSetName(t *testing.T) {
	tests := []struct{ input, expected string }{
		{"Project-X", "project-x"},
		{"  Team  ", "team"},
		{"", ""},
		{"ALPHA", "alpha"},
	}
	for _, tt := range tests {
		got := normalizeSetName(tt.input)
		if got != tt.expected {
			t.Errorf("normalize(%q): got %q expected %q", tt.input, got, tt.expected)
		}
	}
}
