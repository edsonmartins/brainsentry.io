package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

// MetadataKeyBelongsToSets is the JSON metadata key used to store NodeSet membership.
const MetadataKeyBelongsToSets = "belongsToSets"

// NodeSetService manages flexible multi-set membership for memories.
// Uses the existing Memory.Metadata JSON field (no schema migration required).
type NodeSetService struct {
	memoryRepo *postgres.MemoryRepository
}

// NewNodeSetService creates a new NodeSetService.
func NewNodeSetService(memoryRepo *postgres.MemoryRepository) *NodeSetService {
	return &NodeSetService{memoryRepo: memoryRepo}
}

// GetMemorySets returns the sets a memory belongs to.
func (s *NodeSetService) GetMemorySets(m *domain.Memory) []string {
	if m == nil || len(m.Metadata) == 0 {
		return nil
	}

	var meta map[string]any
	if err := json.Unmarshal(m.Metadata, &meta); err != nil {
		return nil
	}

	raw, ok := meta[MetadataKeyBelongsToSets]
	if !ok {
		return nil
	}

	return normalizeSetList(raw)
}

// AddToSet adds a memory to one or more sets (idempotent).
func (s *NodeSetService) AddToSet(ctx context.Context, memoryID string, setNames ...string) error {
	if len(setNames) == 0 {
		return nil
	}

	m, err := s.memoryRepo.FindByID(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("find memory: %w", err)
	}

	existing := s.GetMemorySets(m)
	existingSet := make(map[string]bool, len(existing))
	for _, s := range existing {
		existingSet[s] = true
	}

	merged := make([]string, 0, len(existing)+len(setNames))
	merged = append(merged, existing...)
	for _, name := range setNames {
		name = normalizeSetName(name)
		if name == "" || existingSet[name] {
			continue
		}
		merged = append(merged, name)
		existingSet[name] = true
	}

	sort.Strings(merged)
	return s.writeSets(ctx, m, merged)
}

// RemoveFromSet removes a memory from one or more sets.
func (s *NodeSetService) RemoveFromSet(ctx context.Context, memoryID string, setNames ...string) error {
	if len(setNames) == 0 {
		return nil
	}

	m, err := s.memoryRepo.FindByID(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("find memory: %w", err)
	}

	existing := s.GetMemorySets(m)
	toRemove := make(map[string]bool, len(setNames))
	for _, name := range setNames {
		toRemove[normalizeSetName(name)] = true
	}

	remaining := make([]string, 0, len(existing))
	for _, name := range existing {
		if !toRemove[name] {
			remaining = append(remaining, name)
		}
	}

	return s.writeSets(ctx, m, remaining)
}

// writeSets persists the set list into the memory's metadata.
func (s *NodeSetService) writeSets(ctx context.Context, m *domain.Memory, sets []string) error {
	meta := make(map[string]any)
	if len(m.Metadata) > 0 {
		_ = json.Unmarshal(m.Metadata, &meta)
	}

	if len(sets) == 0 {
		delete(meta, MetadataKeyBelongsToSets)
	} else {
		meta[MetadataKeyBelongsToSets] = sets
	}

	raw, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	m.Metadata = raw

	return s.memoryRepo.Update(ctx, m)
}

// FilterBySet returns memories that belong to the given set.
// Works on an in-memory slice — callers filter fetched results, not at DB level.
// For large result sets, implement JSONB query at repo layer.
func (s *NodeSetService) FilterBySet(memories []domain.Memory, setName string) []domain.Memory {
	if setName == "" {
		return memories
	}
	setName = normalizeSetName(setName)

	result := make([]domain.Memory, 0, len(memories))
	for _, m := range memories {
		sets := s.GetMemorySets(&m)
		if containsString(sets, setName) {
			result = append(result, m)
		}
	}
	return result
}

// ListSets returns all distinct set names observed across the provided memories.
func (s *NodeSetService) ListSets(memories []domain.Memory) []string {
	seen := make(map[string]int)
	for _, m := range memories {
		for _, name := range s.GetMemorySets(&m) {
			seen[name]++
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// normalizeSetName canonicalizes a set name: lowercased, trimmed, collapsed.
func normalizeSetName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return ""
	}
	return name
}

// normalizeSetList accepts either []string or []any from JSON unmarshal.
func normalizeSetList(raw any) []string {
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
