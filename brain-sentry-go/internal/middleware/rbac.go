package middleware

import (
	"net/http"
)

const (
	RoleAdmin    = "ADMIN"
	RoleUser     = "USER"
	RoleReadonly = "READONLY"
)

// RequireRole returns a middleware that checks if the authenticated user has one of the required roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, `{"error":"unauthorized","status":401}`, http.StatusUnauthorized)
				return
			}

			for _, role := range claims.Roles {
				if allowed[role] {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"forbidden: insufficient permissions","status":403}`, http.StatusForbidden)
		})
	}
}
