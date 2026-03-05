package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/pkg/tenant"
)

// TenantExtractor returns a middleware that extracts the tenant ID from the request
// and injects it into the context. Priority: X-Tenant-ID header > query param > JWT claim > default.
func TenantExtractor(defaultTenantID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tenantID string

			// 1. X-Tenant-ID header (highest priority)
			if h := r.Header.Get("X-Tenant-ID"); h != "" {
				tenantID = h
			}

			// 2. Query parameter
			if tenantID == "" {
				if q := r.URL.Query().Get("tenant"); q != "" {
					tenantID = q
				}
			}

			// 3. JWT claim
			if tenantID == "" {
				if claims := ClaimsFromContext(r.Context()); claims != nil && claims.TenantID != "" {
					tenantID = claims.TenantID
				}
			}

			// 4. Default
			if tenantID == "" {
				tenantID = defaultTenantID
			}

			// Validate tenant ID format
			if err := tenant.ValidateTenantID(tenantID); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			ctx := tenant.WithTenant(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
