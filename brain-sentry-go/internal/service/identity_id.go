package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// NamespaceCognitive is the namespace UUID for BrainSentry cognitive entities.
// Derived from UUID v5 of "brainsentry.cognitive" under DNS namespace — stable across runs.
var NamespaceCognitive = uuid.NewSHA1(uuid.NameSpaceDNS, []byte("brainsentry.cognitive"))

// nonAlphanumeric removes non-word characters for canonical normalization.
var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// normalizeIdentityValue produces a canonical lowercase representation of
// an identity field value. Same logical entity → same normalized string → same UUID5.
func normalizeIdentityValue(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = nonAlphanumeric.ReplaceAllString(v, "_")
	v = strings.Trim(v, "_")
	return v
}

// GenerateIdentityID creates a deterministic UUID5 from a class name and identity field values.
// Same inputs always produce the same UUID, enabling idempotent upserts and natural deduplication.
//
// Example:
//   GenerateIdentityID("Entity", "PostgreSQL", "TECHNOLOGY") -> always same UUID
//   GenerateIdentityID("Entity", "postgres", "TECHNOLOGY")   -> same UUID as above (normalization)
func GenerateIdentityID(className string, identityValues ...string) uuid.UUID {
	parts := make([]string, 0, len(identityValues)+1)
	parts = append(parts, strings.ToLower(className))
	for _, v := range identityValues {
		parts = append(parts, normalizeIdentityValue(v))
	}
	key := strings.Join(parts, "|")
	return uuid.NewSHA1(NamespaceCognitive, []byte(key))
}

// GenerateEntityID returns the deterministic UUID for an entity with given name and type.
func GenerateEntityID(name, entityType string) uuid.UUID {
	return GenerateIdentityID("Entity", name, entityType)
}

// GenerateRelationshipID returns the deterministic UUID for a relationship
// between source and target entities of a given type.
// Direction matters: (A, uses, B) != (B, uses, A).
func GenerateRelationshipID(sourceID, targetID, relType string) uuid.UUID {
	return GenerateIdentityID("Relationship", sourceID, relType, targetID)
}

// GenerateTripletID returns the deterministic UUID for a triplet (S, P, O).
func GenerateTripletID(subject, predicate, object string) uuid.UUID {
	return GenerateIdentityID("Triplet", subject, predicate, object)
}

// FormatTripletText builds the embeddable text representation of a triplet
// using the pattern "subject→predicate→object" for vector indexing.
func FormatTripletText(subject, predicate, object string) string {
	return fmt.Sprintf("%s→%s→%s",
		strings.TrimSpace(subject),
		strings.TrimSpace(predicate),
		strings.TrimSpace(object),
	)
}
