package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
)

const apiTokenContextKey contextKey = "api_token"

func GetAPIToken(ctx context.Context) *store.APIToken {
	if val, ok := ctx.Value(apiTokenContextKey).(*store.APIToken); ok {
		return val
	}
	return nil
}

// RequireAPIScope validates API key from X-API-Key or Authorization header and checks scope & IP.
func (s *Server) RequireAPIScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken := r.Header.Get("X-API-Key")
			if rawToken == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					rawToken = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if rawToken == "" {
				http.Error(w, `{"error":"unauthorized: missing API key"}`, http.StatusUnauthorized)
				return
			}

			h := sha256.Sum256([]byte(rawToken))
			tokenHash := hex.EncodeToString(h[:])

			token, err := s.Store.GetAPITokenByHash(r.Context(), tokenHash)
			if err != nil || token == nil {
				http.Error(w, `{"error":"unauthorized: invalid API key"}`, http.StatusUnauthorized)
				return
			}

			// Check expiry
			if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
				http.Error(w, `{"error":"unauthorized: API key expired"}`, http.StatusUnauthorized)
				return
			}

			// Check IP restrictions
			if token.AllowedIPsJSON != nil && *token.AllowedIPsJSON != "" {
				var allowedIPs []string
				if err := json.Unmarshal([]byte(*token.AllowedIPsJSON), &allowedIPs); err == nil && len(allowedIPs) > 0 {
					clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
					if err != nil {
						clientIP = r.RemoteAddr
					}
					ip := net.ParseIP(clientIP)
					matched := false
					for _, allowed := range allowedIPs {
						if strings.Contains(allowed, "/") {
							if _, ipNet, err := net.ParseCIDR(allowed); err == nil && ipNet.Contains(ip) {
								matched = true
								break
							}
						} else if allowed == clientIP {
							matched = true
							break
						}
					}
					if !matched {
						http.Error(w, `{"error":"forbidden: client IP not allowed"}`, http.StatusForbidden)
						return
					}
				}
			}

			// Check scope
			var scopes []string
			if err := json.Unmarshal([]byte(token.ScopesJSON), &scopes); err == nil {
				hasScope := false
				for _, sc := range scopes {
					if sc == "*" || sc == "admin:*" || sc == scope {
						hasScope = true
						break
					}
				}
				if !hasScope {
					http.Error(w, `{"error":"forbidden: insufficient API key scope"}`, http.StatusForbidden)
					return
				}
			}

			// Update last used timestamp asynchronously
			go func() { // #nosec G118
				_ = s.Store.UpdateAPITokenLastUsed(context.Background(), token.ID)
			}()

			ctx := context.WithValue(r.Context(), apiTokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
