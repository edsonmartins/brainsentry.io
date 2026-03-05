package tenant

import (
	"context"
	"fmt"
	"regexp"
)

type contextKey struct{}

const DefaultTenantID = "a9f814d2-4dae-41f3-851b-8aa3d4706561"

// WithTenant returns a new context with the tenant ID set.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, contextKey{}, tenantID)
}

// FromContext extracts the tenant ID from the context.
// Returns the default tenant ID if not set.
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(contextKey{}).(string); ok && id != "" {
		return id
	}
	return DefaultTenantID
}

// HasTenant checks if a tenant ID is set in the context.
func HasTenant(ctx context.Context) bool {
	_, ok := ctx.Value(contextKey{}).(string)
	return ok
}

// tenantIDPattern validates that tenant IDs are alphanumeric with hyphens and underscores, max 64 chars.
var tenantIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// ValidateTenantID validates a tenant ID format.
// Must be alphanumeric with hyphens and underscores, max 64 characters.
func ValidateTenantID(id string) error {
	if id == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if !tenantIDPattern.MatchString(id) {
		return fmt.Errorf("tenant ID must be alphanumeric with hyphens and underscores, max 64 chars")
	}
	return nil
}
