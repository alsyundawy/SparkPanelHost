package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
	chi "github.com/go-chi/chi/v5"
)

func TestSecurityAndIsolationHandlers(t *testing.T) {
	migrationPath := filepath.Join("..", "..", "..", "migrations", "001_init.sql")
	st, err := store.NewStore(":memory:", migrationPath)
	if err != nil {
		t.Fatalf("failed to init in-memory store: %v", err)
	}
	defer st.Close()

	srv := NewServer(st, nil, "")

	// Seed users: superadmin, reseller, tenant1, tenant2 with required schema fields
	uAdmin := &store.User{
		ID:           "usr-admin",
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "hash",
		Role:         "superadmin",
		LinuxUID:     1000,
		LinuxGID:     1000,
		Locale:       "id",
		Theme:        "dark",
		Status:       "active",
	}
	if err := st.CreateUser(context.Background(), uAdmin); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	resellerID := "usr-reseller"
	uReseller := &store.User{
		ID:           resellerID,
		Username:     "reseller",
		Email:        "reseller@example.com",
		PasswordHash: "hash",
		Role:         "reseller",
		LinuxUID:     2001,
		LinuxGID:     2001,
		Locale:       "id",
		Theme:        "dark",
		Status:       "active",
	}
	if err := st.CreateUser(context.Background(), uReseller); err != nil {
		t.Fatalf("failed to create reseller: %v", err)
	}

	uTenant1 := &store.User{
		ID:           "usr-tenant1",
		Username:     "tenant1",
		Email:        "tenant1@example.com",
		PasswordHash: "hash",
		Role:         "tenant",
		ResellerID:   &resellerID,
		LinuxUID:     2002,
		LinuxGID:     2002,
		Locale:       "id",
		Theme:        "dark",
		Status:       "active",
	}
	if err := st.CreateUser(context.Background(), uTenant1); err != nil {
		t.Fatalf("failed to create tenant1: %v", err)
	}

	uTenant2 := &store.User{
		ID:           "usr-tenant2",
		Username:     "tenant2",
		Email:        "tenant2@example.com",
		PasswordHash: "hash",
		Role:         "tenant",
		LinuxUID:     2003,
		LinuxGID:     2003,
		Locale:       "id",
		Theme:        "dark",
		Status:       "active",
	}
	if err := st.CreateUser(context.Background(), uTenant2); err != nil {
		t.Fatalf("failed to create tenant2: %v", err)
	}

	// 1. Tenant attempting to impersonate -> 403 Forbidden
	bodyImp, _ := json.Marshal(ImpersonateRequest{TargetUserID: "usr-admin"})
	reqImp := httptest.NewRequest("POST", "/api/v1/auth/impersonate", bytes.NewReader(bodyImp))
	reqImp = reqImp.WithContext(context.WithValue(reqImp.Context(), userContextKey, uTenant1))
	wImp := httptest.NewRecorder()
	srv.Impersonate(wImp, reqImp)
	if wImp.Code != http.StatusForbidden {
		t.Errorf("expected 403 when tenant tries to impersonate, got %d", wImp.Code)
	}

	// 2. Reseller attempting to impersonate a tenant belonging to another reseller/admin -> 403 Forbidden
	bodyImp2, _ := json.Marshal(ImpersonateRequest{TargetUserID: "usr-tenant2"})
	reqImp2 := httptest.NewRequest("POST", "/api/v1/auth/impersonate", bytes.NewReader(bodyImp2))
	reqImp2 = reqImp2.WithContext(context.WithValue(reqImp2.Context(), userContextKey, uReseller))
	wImp2 := httptest.NewRecorder()
	srv.Impersonate(wImp2, reqImp2)
	if wImp2.Code != http.StatusForbidden {
		t.Errorf("expected 403 when reseller tries to impersonate unowned tenant, got %d", wImp2.Code)
	}

	// 3. Database identifier validation against SQL injection metacharacters (aaPanel CVE-2025-12914)
	bodyBadDB, _ := json.Marshal(CreateDatabaseRequest{
		DBName:   "test; DROP TABLE users;--",
		DBUser:   "test' OR '1'='1",
		Password: "password123",
	})
	reqBadDB := httptest.NewRequest("POST", "/api/v1/databases", bytes.NewReader(bodyBadDB))
	reqBadDB = reqBadDB.WithContext(context.WithValue(reqBadDB.Context(), userContextKey, uTenant1))
	wBadDB := httptest.NewRecorder()
	srv.CreateDatabase(wBadDB, reqBadDB)
	if wBadDB.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for SQLi characters in database name/user, got %d", wBadDB.Code)
	}

	// 4. Tenant boundary on domain deletion
	dom := &store.Domain{
		ID:            "dom-tenant2",
		UserID:        "usr-tenant2",
		DomainName:    "tenant2.example.com",
		DocumentRoot:  "/home/tenant2/public_html",
		PHPVersion:    "8.5",
		WAFMode:       "Blocking",
		WAFProfile:    "default",
		NginxTemplate: "default",
	}
	if err := st.CreateDomain(context.Background(), dom); err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	reqDelDom := httptest.NewRequest("DELETE", "/api/v1/domains/dom-tenant2", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "dom-tenant2")
	reqDelDom = reqDelDom.WithContext(context.WithValue(reqDelDom.Context(), chi.RouteCtxKey, rctx))
	reqDelDom = reqDelDom.WithContext(context.WithValue(reqDelDom.Context(), userContextKey, uTenant1))
	wDelDom := httptest.NewRecorder()
	srv.DeleteDomain(wDelDom, reqDelDom)
	if wDelDom.Code != http.StatusForbidden {
		t.Errorf("expected 403 when tenant1 tries to delete tenant2 domain, got %d", wDelDom.Code)
	}

	// 5. Tenant boundary on database deletion
	db := &store.TenantDatabase{
		ID:      "db-tenant2",
		UserID:  "usr-tenant2",
		Engine:  "mariadb",
		DBName:  "tenant2_app",
		DBUser:  "tenant2_user",
		Charset: "utf8mb4",
		SizeMB:  1.0,
	}
	if err := st.CreateTenantDatabase(context.Background(), db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	reqDelDB := httptest.NewRequest("DELETE", "/api/v1/databases/db-tenant2", nil)
	rctxDB := chi.NewRouteContext()
	rctxDB.URLParams.Add("id", "db-tenant2")
	reqDelDB = reqDelDB.WithContext(context.WithValue(reqDelDB.Context(), chi.RouteCtxKey, rctxDB))
	reqDelDB = reqDelDB.WithContext(context.WithValue(reqDelDB.Context(), userContextKey, uTenant1))
	wDelDB := httptest.NewRecorder()
	srv.DeleteDatabase(wDelDB, reqDelDB)
	if wDelDB.Code != http.StatusForbidden {
		t.Errorf("expected 403 when tenant1 tries to delete tenant2 database, got %d", wDelDB.Code)
	}

	// 6. Tenant denied web terminal shell access
	reqTerm := httptest.NewRequest("GET", "/api/v1/terminal/ws", nil)
	reqTerm = reqTerm.WithContext(context.WithValue(reqTerm.Context(), userContextKey, uTenant1))
	wTerm := httptest.NewRecorder()
	srv.WebTerminalWS(wTerm, reqTerm)
	if wTerm.Code != http.StatusForbidden {
		t.Errorf("expected 403 when tenant accesses web terminal, got %d", wTerm.Code)
	}
}
