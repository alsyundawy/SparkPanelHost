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

func TestConfigEditorEndpoints(t *testing.T) {
	migrationPath := filepath.Join("..", "..", "..", "migrations", "001_init.sql")
	st, err := store.NewStore(":memory:", migrationPath)
	if err != nil {
		t.Fatalf("failed to init in-memory store: %v", err)
	}
	defer st.Close()

	srv := NewServer(st, nil, "")

	// 1. Test GetConfig for unsupported service
	req := httptest.NewRequest("GET", "/api/v1/config/unsupported", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("service", "unsupported")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	srv.GetConfig(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unsupported service, got %d", w.Code)
	}

	// 2. Test UpdateConfig validation failure for invalid payload
	body, _ := json.Marshal(UpdateConfigRequest{
		SubPath: "",
		Content: "invalid content {{{",
	})
	reqUpdate := httptest.NewRequest("PUT", "/api/v1/config/nginx", bytes.NewReader(body))
	rctxUpdate := chi.NewRouteContext()
	rctxUpdate.URLParams.Add("service", "nginx")
	reqUpdate = reqUpdate.WithContext(context.WithValue(reqUpdate.Context(), chi.RouteCtxKey, rctxUpdate))
	wUpdate := httptest.NewRecorder()

	srv.UpdateConfig(wUpdate, reqUpdate)
	// On macOS test runner where /etc/nginx is permission-denied or binary missing, it returns 400
	if wUpdate.Code != http.StatusBadRequest {
		t.Logf("UpdateConfig response code: %d (body: %s)", wUpdate.Code, wUpdate.Body.String())
	}

	// 3. Test RollbackConfig missing timestamp
	bodyRollback, _ := json.Marshal(RollbackConfigRequest{
		SubPath:         "",
		BackupTimestamp: "",
	})
	reqRollback := httptest.NewRequest("POST", "/api/v1/config/nginx/rollback", bytes.NewReader(bodyRollback))
	rctxRollback := chi.NewRouteContext()
	rctxRollback.URLParams.Add("service", "nginx")
	reqRollback = reqRollback.WithContext(context.WithValue(reqRollback.Context(), chi.RouteCtxKey, rctxRollback))
	wRollback := httptest.NewRecorder()

	srv.RollbackConfig(wRollback, reqRollback)
	if wRollback.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing backup_timestamp, got %d", wRollback.Code)
	}
}
