package service

import (
	"testing"
)

func sampleOntology() *Ontology {
	return &Ontology{
		Name:    "test",
		Version: "1.0",
		EntityTypes: []OntologyEntityType{
			{Name: "TECHNOLOGY", Description: "Technologies and tools"},
			{Name: "PERSON"},
			{Name: "LANGUAGE"},
		},
		Entities: []OntologyEntity{
			{Name: "PostgreSQL", Type: "TECHNOLOGY", Aliases: []string{"postgres", "postgre"}},
			{Name: "Go", Type: "LANGUAGE", Aliases: []string{"golang"}},
		},
		Relationships: []OntologyRelationship{
			{Name: "uses", SourceType: "*", TargetType: "TECHNOLOGY"},
			{Name: "implements", SourceType: "LANGUAGE", TargetType: "TECHNOLOGY"},
			{Name: "authored_by", SourceType: "TECHNOLOGY", TargetType: "PERSON"},
		},
	}
}

func TestOntology_DisabledByDefault(t *testing.T) {
	svc := NewOntologyService()
	if svc.IsEnabled() {
		t.Error("service should be disabled without ontology")
	}
	if svc.IsValidType("ANYTHING") == false {
		t.Error("disabled service should be permissive")
	}
}

func TestOntology_SetAndQuery(t *testing.T) {
	svc := NewOntologyService()
	if err := svc.SetOntology(sampleOntology()); err != nil {
		t.Fatalf("SetOntology: %v", err)
	}
	if !svc.IsEnabled() {
		t.Error("expected enabled after SetOntology")
	}

	types := svc.AllowedTypes()
	if len(types) != 3 {
		t.Errorf("expected 3 types, got %d", len(types))
	}

	rels := svc.AllowedRelationships()
	if len(rels) != 3 {
		t.Errorf("expected 3 relationships, got %d", len(rels))
	}
}

func TestOntology_ResolveExactMatch(t *testing.T) {
	svc := NewOntologyService()
	svc.SetOntology(sampleOntology())

	canonical, typ, ok := svc.ResolveEntity("PostgreSQL")
	if !ok || canonical != "PostgreSQL" || typ != "TECHNOLOGY" {
		t.Errorf("exact match failed: got canonical=%q type=%q ok=%v", canonical, typ, ok)
	}
}

func TestOntology_ResolveAlias(t *testing.T) {
	svc := NewOntologyService()
	svc.SetOntology(sampleOntology())

	canonical, typ, ok := svc.ResolveEntity("postgres")
	if !ok || canonical != "PostgreSQL" || typ != "TECHNOLOGY" {
		t.Errorf("alias resolution failed: got canonical=%q type=%q ok=%v", canonical, typ, ok)
	}
}

func TestOntology_ResolveCaseInsensitive(t *testing.T) {
	svc := NewOntologyService()
	svc.SetOntology(sampleOntology())

	canonical, _, ok := svc.ResolveEntity("GOLANG")
	if !ok || canonical != "Go" {
		t.Errorf("case-insensitive failed: got canonical=%q ok=%v", canonical, ok)
	}
}

func TestOntology_ResolveNoMatch(t *testing.T) {
	svc := NewOntologyService()
	svc.SetOntology(sampleOntology())

	_, _, ok := svc.ResolveEntity("SomethingCompletelyDifferent")
	if ok {
		t.Error("expected no match")
	}
}

func TestOntology_ValidType(t *testing.T) {
	svc := NewOntologyService()
	svc.SetOntology(sampleOntology())

	if !svc.IsValidType("TECHNOLOGY") {
		t.Error("TECHNOLOGY should be valid")
	}
	if svc.IsValidType("INVALID") {
		t.Error("INVALID should not be valid")
	}
}

func TestOntology_ValidRelationship(t *testing.T) {
	svc := NewOntologyService()
	svc.SetOntology(sampleOntology())

	tests := []struct {
		name       string
		src, tgt   string
		expected   bool
		comment    string
	}{
		{"uses", "PERSON", "TECHNOLOGY", true, "wildcard source matches"},
		{"implements", "LANGUAGE", "TECHNOLOGY", true, "exact type match"},
		{"implements", "PERSON", "TECHNOLOGY", false, "wrong source type"},
		{"unknown", "A", "B", false, "unknown relationship"},
		{"uses", "PERSON", "PERSON", false, "target must be TECHNOLOGY"},
	}

	for _, tt := range tests {
		got := svc.IsValidRelationship(tt.name, tt.src, tt.tgt)
		if got != tt.expected {
			t.Errorf("%s (%s): expected %v got %v (%s → %s)",
				tt.comment, tt.name, tt.expected, got, tt.src, tt.tgt)
		}
	}
}

func TestOntology_Levenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		d    int
	}{
		{"kitten", "sitting", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"a", "b", 1},
	}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.d {
			t.Errorf("levenshtein(%q,%q) = %d, expected %d", c.a, c.b, got, c.d)
		}
	}
}

func TestOntology_Similarity(t *testing.T) {
	// identical → 1.0
	if s := similarity("abc", "abc"); s != 1.0 {
		t.Errorf("identical should be 1.0, got %f", s)
	}
	// empty → 0.0
	if s := similarity("", "abc"); s != 0.0 {
		t.Errorf("empty should be 0.0, got %f", s)
	}
	// partial
	if s := similarity("postgres", "postgresql"); s < 0.5 {
		t.Errorf("similar words should score > 0.5, got %f", s)
	}
}
