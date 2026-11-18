package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/auth"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
)

type contextKey string

const (
	userContextKey    contextKey = "auth_user"
	sessionContextKey contextKey = "auth_session"
)

type APIContext struct {
	Store *store.Store
}

func (a *APIContext) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""

		// Check Authorization Header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Fallback to Cookie
		if token == "" {
			if cookie, err := r.Cookie("aether_session"); err == nil {
				token = cookie.Value
			}
		}

		if token == "" {
			writeJSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		tokenHash := auth.HashToken(token)
		sess, err := a.Store.GetSessionByTokenHash(r.Context(), tokenHash)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		u, err := a.Store.GetUserByID(r.Context(), sess.UserID)
		if err != nil || u.Status == "suspended" {
			writeJSONError(w, http.StatusForbidden, "Account is suspended or inactive")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, u)
		ctx = context.WithValue(ctx, sessionContextKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := r.Context().Value(userContextKey).(*store.User)
			if !ok || u == nil {
				writeJSONError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			allowed := false
			for _, r := range roles {
				if u.Role == r {
					allowed = true
					break
				}
			}

			if !allowed {
				writeJSONError(w, http.StatusForbidden, "Insufficient permissions for this operation")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserFromContext(ctx context.Context) *store.User {
	if u, ok := ctx.Value(userContextKey).(*store.User); ok {
		return u
	}
	return nil
}

func GetSessionFromContext(ctx context.Context) *store.UserSession {
	if s, ok := ctx.Value(sessionContextKey).(*store.UserSession); ok {
		return s
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
