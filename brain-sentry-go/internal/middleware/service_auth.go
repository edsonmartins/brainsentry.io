package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type serviceAuthContextKey string

const ServiceClaimsKey serviceAuthContextKey = "serviceClaims"

// ServiceClaims represents the claims in a service-to-service JWT.
type ServiceClaims struct {
	jwt.RegisteredClaims
}

// ServiceAuth middleware validates service-to-service JWTs on integration endpoints.
// Uses SQUADX_SERVICE_SECRET (shared HMAC secret).
func ServiceAuth(next http.Handler) http.Handler {
	secret := os.Getenv("SQUADX_SERVICE_SECRET")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" || len(secret) < 32 {
			http.Error(w, `{"error":"service auth not configured"}`, http.StatusServiceUnavailable)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenString, &ServiceClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":"invalid service token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*ServiceClaims)
		if !ok {
			http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		// Verify expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			http.Error(w, `{"error":"token expired"}`, http.StatusUnauthorized)
			return
		}

		// Store claims in context for downstream handlers
		ctx := context.WithValue(r.Context(), ServiceClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetServiceClaims extracts service claims from the request context.
func GetServiceClaims(r *http.Request) *ServiceClaims {
	claims, _ := r.Context().Value(ServiceClaimsKey).(*ServiceClaims)
	return claims
}
