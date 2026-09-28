package middleware

import (
	"context"
	"net/http"
	"strings"

	"ecommerce/backend/internal/auth"
	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/utils"
)

type contextKey string

const (
	ContextUserID contextKey = "user_id"
	ContextEmail  contextKey = "email"
	ContextRole   contextKey = "role"
)

// RequireAuth validates the Bearer JWT and injects user info into the
// request context. Responds 401 if missing/invalid.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				utils.Error(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ParseToken(jwtSecret, tokenString)
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin must be chained after RequireAuth. Responds 403 if the
// authenticated user is not an admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ContextRole).(models.Role)
		if !ok || role != models.RoleAdmin {
			utils.Error(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserIDFromContext extracts the authenticated user's ID.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ContextUserID).(int64)
	return id, ok
}
