package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/oopbest/ecommerce-app/pkg/security"
)

type contextKey string

const UserContextKey = contextKey("user_claims")

// AuthMiddleware ตรวจสอบ JWT Bearer Token ใน Header
func AuthMiddleware(jwtSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			// ตรวจสอบ Format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				sendError(w, http.StatusUnauthorized, "Invalid authorization format. Expected 'Bearer <token>'")
				return
			}

			tokenString := parts[1]
			claims, err := security.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				sendError(w, http.StatusUnauthorized, "Invalid or expired token: "+err.Error())
				return
			}

			// ฝัง Claims (UserID, Email, Role) เข้าไปใน Request Context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next(w, r.WithContext(ctx))
		}
	}
}

// RequireRole Middleware สำหรับตรวจสอบสิทธิ์ (Role-Based Access Control)
func RequireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*security.CustomClaims)
		if !ok || claims == nil {
			sendError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if claims.Role != requiredRole {
			sendError(w, http.StatusForbidden, "Forbidden: insufficient permissions, '"+requiredRole+"' role required")
			return
		}

		next(w, r)
	}
}

// GetUserClaims Helper ฟังก์ชันสำหรับดึง Claims ของ User ออกจาก Context
func GetUserClaims(r *http.Request) (*security.CustomClaims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(*security.CustomClaims)
	return claims, ok
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
