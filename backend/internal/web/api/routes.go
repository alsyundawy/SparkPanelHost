package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) Routes(staticDir string) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Security Headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	})

	apiCtx := &APIContext{Store: s.Store}

	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth
		r.Post("/auth/login", s.Login)

		// Billing Automation API (WHMCS / Blesta)
		r.Route("/billing", func(r chi.Router) {
			r.Use(s.RequireAPIScope("billing:manage"))
			r.Post("/accounts", s.BillingCreateAccount)
			r.Post("/accounts/{username}/suspend", s.BillingSuspendAccount)
			r.Post("/accounts/{username}/unsuspend", s.BillingUnsuspendAccount)
			r.Delete("/accounts/{username}", s.BillingTerminateAccount)
			r.Get("/accounts/{username}/usage", s.BillingGetUsage)
		})

		// Protected Endpoints
		r.Group(func(r chi.Router) {
			r.Use(apiCtx.AuthMiddleware)

			// Auth Session
			r.Post("/auth/logout", s.Logout)
			r.Get("/auth/me", s.Me)
			r.With(RequireRole("superadmin", "reseller")).Post("/auth/impersonate", s.Impersonate)

			// 2FA TOTP & Passkeys
			r.Post("/auth/2fa/totp/setup", s.SetupTOTP)
			r.Post("/auth/2fa/totp/enable", s.EnableTOTP)
			r.Post("/auth/2fa/totp/disable", s.DisableTOTP)
			r.Post("/auth/passkey/challenge", s.RegisterPasskeyChallenge)
			r.Post("/auth/passkey/register", s.RegisterPasskeyVerify)
			r.Get("/auth/passkeys", s.ListPasskeys)
			r.Delete("/auth/passkeys/{id}", s.DeletePasskey)

			// Users & Accounts
			r.Get("/users", s.ListUsers)
			r.With(RequireRole("superadmin", "reseller")).Post("/users", s.CreateUser)

			// Packages
			r.Get("/packages", func(w http.ResponseWriter, r *http.Request) {
				pkgs, err := s.Store.ListPackages(r.Context())
				if err != nil {
					writeJSONError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "packages": pkgs})
			})

			// Domains
			r.Get("/domains", s.ListDomains)
			r.Post("/domains", s.CreateDomain)
			r.Delete("/domains/{id}", s.DeleteDomain)

			// Databases
			r.Get("/databases", s.ListDatabases)
			r.Post("/databases", s.CreateDatabase)
			r.Delete("/databases/{id}", s.DeleteDatabase)

			// DNS Clustering
			r.Get("/dns/clusters", s.ListDNSClusters)
			r.With(RequireRole("superadmin")).Post("/dns/clusters", s.CreateDNSCluster)
			r.With(RequireRole("superadmin")).Post("/dns/clusters/nodes", s.RegisterDNSNode)
			r.Get("/dns/clusters/{name}/status", s.GetDNSClusterStatus)

			// Enterprise Mail
			r.Get("/mail/domains", s.ListMailDomains)
			r.Get("/mail/rbl", s.CheckRBL)

			// App Recipes
			r.Get("/apps/recipes", s.ListRecipes)

			// System Telemetry
			r.Get("/telemetry", s.GetTelemetry)
			r.Get("/telemetry/ws", s.StreamTelemetryWS)

			// Web Terminal WebSocket
			r.Get("/terminal/ws", s.WebTerminalWS)

			// Audit Logs
			r.With(RequireRole("superadmin")).Get("/audit-logs", s.ListAuditLogs)

			// Service Configuration Editor with 1-Click Rollback (superadmin only)
			r.Route("/config/{service}", func(cr chi.Router) {
				cr.Use(RequireRole("superadmin"))
				cr.Get("/", s.GetConfig)
				cr.Put("/", s.UpdateConfig)
				cr.Post("/rollback", s.RollbackConfig)
				cr.Get("/backups", s.ListConfigBackups)
			})
		})
	})

	// Static SPA File Server
	if staticDir != "" {
		fileServer(r, "/", http.Dir(staticDir))
	} else {
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>AetherPanel v8.1</title></head><body><h1>AetherPanel v8.1 Control Plane API Active</h1><p>Visit /api/v1/auth/me</p></body></html>`))
		})
	}

	return r
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))

		baseDir := filepath.Clean(string(root.(http.Dir)))
		cleanURLPath := filepath.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		fPath := filepath.Clean(filepath.Join(baseDir, cleanURLPath))
		rel, err := filepath.Rel(baseDir, fPath)
		if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check if file exists, fallback to index.html for SPA
		if _, err := os.Stat(fPath); os.IsNotExist(err) { // #nosec G703
			http.ServeFile(w, r, filepath.Join(baseDir, "index.html"))
			return
		}

		fs.ServeHTTP(w, r)
	})
}
