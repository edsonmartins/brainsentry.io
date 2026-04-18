package service

import "testing"

func TestGenerateIdentityID_Deterministic(t *testing.T) {
	id1 := GenerateIdentityID("Entity", "PostgreSQL", "TECHNOLOGY")
	id2 := GenerateIdentityID("Entity", "PostgreSQL", "TECHNOLOGY")

	if id1 != id2 {
		t.Errorf("expected same UUID for same inputs, got %s vs %s", id1, id2)
	}
}

func TestGenerateIdentityID_Normalization(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"PostgreSQL", "postgresql"},
		{"PostgreSQL", "  PostgreSQL  "},
		{"New York City", "new_york_city"},
	}

	for _, tt := range tests {
		id1 := GenerateIdentityID("Entity", tt.a, "TECHNOLOGY")
		id2 := GenerateIdentityID("Entity", tt.b, "TECHNOLOGY")
		if id1 != id2 {
			t.Errorf("expected same UUID for normalized equivalents %q / %q, got %s vs %s",
				tt.a, tt.b, id1, id2)
		}
	}
}

func TestGenerateIdentityID_DifferentInputs(t *testing.T) {
	a := GenerateIdentityID("Entity", "PostgreSQL", "TECHNOLOGY")
	b := GenerateIdentityID("Entity", "MySQL", "TECHNOLOGY")

	if a == b {
		t.Error("expected different UUIDs for different names")
	}
}

func TestGenerateIdentityID_TypeMatters(t *testing.T) {
	a := GenerateIdentityID("Entity", "Python", "TECHNOLOGY")
	b := GenerateIdentityID("Entity", "Python", "LANGUAGE")

	if a == b {
		t.Error("expected different UUIDs for different types")
	}
}

func TestGenerateIdentityID_ClassNameMatters(t *testing.T) {
	a := GenerateIdentityID("Entity", "X", "Y")
	b := GenerateIdentityID("Relationship", "X", "Y")

	if a == b {
		t.Error("expected different UUIDs for different class names")
	}
}

func TestGenerateEntityID_Matches(t *testing.T) {
	direct := GenerateEntityID("Go", "LANGUAGE")
	viaIdentity := GenerateIdentityID("Entity", "Go", "LANGUAGE")

	if direct != viaIdentity {
		t.Errorf("GenerateEntityID should match identity path, got %s vs %s", direct, viaIdentity)
	}
}

func TestGenerateRelationshipID_DirectionMatters(t *testing.T) {
	forward := GenerateRelationshipID("A", "B", "uses")
	reverse := GenerateRelationshipID("B", "A", "uses")

	if forward == reverse {
		t.Error("expected different UUIDs for forward vs reverse relationships")
	}
}

func TestGenerateTripletID_Deterministic(t *testing.T) {
	a := GenerateTripletID("PostgreSQL", "supports", "JSON")
	b := GenerateTripletID("PostgreSQL", "supports", "JSON")

	if a != b {
		t.Error("expected same UUID for same triplet")
	}
}

func TestFormatTripletText(t *testing.T) {
	text := FormatTripletText("PostgreSQL", "supports", "JSON")
	expected := "PostgreSQL→supports→JSON"

	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}
}

func TestFormatTripletText_TrimsSpaces(t *testing.T) {
	text := FormatTripletText("  A  ", "  B  ", "  C  ")
	expected := "A→B→C"

	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}
}

func TestNormalizeIdentityValue(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"PostgreSQL", "postgresql"},
		{"  spaced  ", "spaced"},
		{"hello-world", "hello_world"},
		{"foo__bar", "foo_bar"},
		{"__leading", "leading"},
	}
	for _, tt := range tests {
		got := normalizeIdentityValue(tt.input)
		if got != tt.expected {
			t.Errorf("normalize(%q): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}
