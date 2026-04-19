package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

// Ontology is a lightweight JSON-based vocabulary that constrains entity
// extraction. It lists canonical entity types, canonical names with aliases,
// and allowed relationship types.
//
// Unlike full RDF/OWL (which requires external libraries and complex parsing),
// this format captures 80% of the value: domain-specific vocabulary control,
// canonical names, and relationship constraints.
type Ontology struct {
	Name          string                  `json:"name"`
	Version       string                  `json:"version"`
	EntityTypes   []OntologyEntityType    `json:"entityTypes"`
	Entities      []OntologyEntity        `json:"entities,omitempty"`      // canonical entities with aliases
	Relationships []OntologyRelationship  `json:"relationships"`
}

// OntologyEntityType represents a type in the vocabulary.
type OntologyEntityType struct {
	Name        string   `json:"name"`        // e.g., "TECHNOLOGY"
	Description string   `json:"description,omitempty"`
	ParentType  string   `json:"parentType,omitempty"` // inheritance
	Examples    []string `json:"examples,omitempty"`
}

// OntologyEntity represents a canonical named entity.
// Extracted entities whose names fuzzy-match any alias resolve to this canonical.
type OntologyEntity struct {
	Name     string   `json:"name"`              // canonical name
	Type     string   `json:"type"`              // must be an OntologyEntityType
	Aliases  []string `json:"aliases,omitempty"` // alternative forms (case-insensitive)
}

// OntologyRelationship declares an allowed relationship between entity types.
type OntologyRelationship struct {
	Name       string   `json:"name"`       // e.g., "uses", "depends_on"
	SourceType string   `json:"sourceType"` // allowed source entity type (or "*")
	TargetType string   `json:"targetType"` // allowed target entity type (or "*")
	Symmetric  bool     `json:"symmetric,omitempty"`
}

// OntologyService loads and queries an ontology. Safe for concurrent use.
type OntologyService struct {
	mu            sync.RWMutex
	ontology      *Ontology
	typeIndex     map[string]*OntologyEntityType // lowercase name → type
	entityIndex   map[string]*OntologyEntity      // lowercase alias → canonical entity
	relIndex      map[string][]*OntologyRelationship // lowercase name → relationships
}

// NewOntologyService creates an empty (no-op) service.
// Use LoadFromFile or SetOntology to activate it.
func NewOntologyService() *OntologyService {
	return &OntologyService{
		typeIndex:   map[string]*OntologyEntityType{},
		entityIndex: map[string]*OntologyEntity{},
		relIndex:    map[string][]*OntologyRelationship{},
	}
}

// LoadFromFile loads an ontology from a JSON file path.
// Empty path is a no-op (service stays disabled).
func (s *OntologyService) LoadFromFile(path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read ontology file: %w", err)
	}

	var ont Ontology
	if err := json.Unmarshal(data, &ont); err != nil {
		return fmt.Errorf("parse ontology: %w", err)
	}
	return s.SetOntology(&ont)
}

// SetOntology replaces the current ontology and rebuilds indexes.
func (s *OntologyService) SetOntology(ont *Ontology) error {
	if ont == nil {
		return fmt.Errorf("ontology is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ontology = ont
	s.typeIndex = make(map[string]*OntologyEntityType, len(ont.EntityTypes))
	s.entityIndex = make(map[string]*OntologyEntity, len(ont.Entities))
	s.relIndex = make(map[string][]*OntologyRelationship, len(ont.Relationships))

	for i := range ont.EntityTypes {
		t := &ont.EntityTypes[i]
		s.typeIndex[strings.ToLower(t.Name)] = t
	}

	for i := range ont.Entities {
		e := &ont.Entities[i]
		s.entityIndex[strings.ToLower(e.Name)] = e
		for _, alias := range e.Aliases {
			s.entityIndex[strings.ToLower(alias)] = e
		}
	}

	for i := range ont.Relationships {
		r := &ont.Relationships[i]
		key := strings.ToLower(r.Name)
		s.relIndex[key] = append(s.relIndex[key], r)
	}

	return nil
}

// IsEnabled reports whether the service has an ontology loaded.
func (s *OntologyService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ontology != nil
}

// Ontology returns the loaded ontology (or nil).
func (s *OntologyService) Ontology() *Ontology {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ontology
}

// ResolveEntity tries to map an extracted entity name to a canonical ontology entity.
// Returns (canonicalName, type, true) on match; ("", "", false) if no match.
// Uses exact match on name and aliases (case-insensitive), then fuzzy match.
func (s *OntologyService) ResolveEntity(name string) (string, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ontology == nil {
		return "", "", false
	}

	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return "", "", false
	}

	// Exact match (name or alias)
	if e, ok := s.entityIndex[key]; ok {
		return e.Name, e.Type, true
	}

	// Fuzzy match (Levenshtein ratio >= 0.85 on canonical names)
	best := ""
	bestType := ""
	bestScore := 0.0
	for k, e := range s.entityIndex {
		score := similarity(key, k)
		if score > bestScore {
			bestScore = score
			best = e.Name
			bestType = e.Type
		}
	}
	if bestScore >= 0.85 {
		return best, bestType, true
	}
	return "", "", false
}

// IsValidType reports whether the type is declared in the ontology.
func (s *OntologyService) IsValidType(typeName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ontology == nil {
		return true // disabled → permissive
	}
	_, ok := s.typeIndex[strings.ToLower(typeName)]
	return ok
}

// IsValidRelationship reports whether the relationship name is allowed between two types.
// Empty types are treated as wildcards. Unknown relationships return false.
func (s *OntologyService) IsValidRelationship(name, sourceType, targetType string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ontology == nil {
		return true // permissive when disabled
	}

	rels := s.relIndex[strings.ToLower(name)]
	if len(rels) == 0 {
		return false
	}

	src := strings.ToUpper(sourceType)
	tgt := strings.ToUpper(targetType)

	for _, r := range rels {
		rSrc := strings.ToUpper(r.SourceType)
		rTgt := strings.ToUpper(r.TargetType)

		srcMatch := rSrc == "*" || rSrc == src || src == ""
		tgtMatch := rTgt == "*" || rTgt == tgt || tgt == ""

		if srcMatch && tgtMatch {
			return true
		}
		if r.Symmetric && rSrc == tgt && rTgt == src {
			return true
		}
	}
	return false
}

// AllowedTypes returns the sorted list of entity type names in the ontology.
func (s *OntologyService) AllowedTypes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ontology == nil {
		return nil
	}
	names := make([]string, 0, len(s.ontology.EntityTypes))
	for _, t := range s.ontology.EntityTypes {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return names
}

// AllowedRelationships returns the sorted list of relationship names.
func (s *OntologyService) AllowedRelationships() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ontology == nil {
		return nil
	}
	seen := make(map[string]bool)
	for _, r := range s.ontology.Relationships {
		seen[r.Name] = true
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// similarity returns a ratio in [0, 1] based on Levenshtein distance.
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	dist := levenshtein(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	return 1.0 - float64(dist)/float64(maxLen)
}

// levenshtein computes edit distance with O(min(a, b)) space.
func levenshtein(a, b string) int {
	if len(a) < len(b) {
		a, b = b, a
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(
				prev[j]+1,      // deletion
				curr[j-1]+1,    // insertion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
