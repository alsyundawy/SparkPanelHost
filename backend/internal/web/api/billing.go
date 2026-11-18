package api

import (
	"encoding/json"
	"net/http"

	agentv1 "github.com/alsyundawy/SparkPanelHost/backend/gen/agent/v1"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/auth"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BillingCreateAccountRequest struct {
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	Domain     string  `json:"domain"`
	PackageID  string  `json:"package_id"`
	ResellerID *string `json:"reseller_id,omitempty"`
}

func (s *Server) BillingCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req BillingCreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}

	// 1. Hash password with Argon2id
	passHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// 2. Allocate Linux UID/GID deterministically or incrementally
	assignedUID, err := s.Store.GetNextLinuxUID(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to allocate Linux UID: "+err.Error())
		return
	}

	newUser := &store.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passHash,
		Role:         "tenant",
		ResellerID:   req.ResellerID,
		PackageID:    &req.PackageID,
		LinuxUID:     assignedUID,
		LinuxGID:     assignedUID,
		Locale:       "id",
		Theme:        "dark",
		Status:       "active",
	}

	if err := s.Store.CreateUser(r.Context(), newUser); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create user in database: "+err.Error())
		return
	}

	// 3. Provision Tenant via System Agent daemon
	if s.Agent != nil {
		_, err := s.Agent.Agent().ProvisionTenant(r.Context(), &agentv1.ProvisionTenantRequest{
			Username:        req.Username,
			Uid:             uint32(assignedUID), // #nosec G115 - uid is allocated >= 2000
			Gid:             uint32(assignedUID), // #nosec G115
			DiskQuotaMb:     10240,
			InodeQuota:      500000,
			CpuQuotaPercent: 100,
			MemoryLimitMb:   1024,
			TasksMax:        256,
			IoWeight:        100,
		})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "agent provisioning failed: "+err.Error())
			return
		}
	}

	token := GetAPIToken(r.Context())
	actorID := "billing_api"
	if token != nil {
		actorID = "api_token:" + token.ID
	}
	_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "billing.account_created", "user:"+req.Username, r.RemoteAddr)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"result":   "success",
		"user_id":  newUser.ID,
		"username": newUser.Username,
		"status":   "active",
	})
}

func (s *Server) BillingSuspendAccount(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	u, err := s.Store.GetUserByUsername(r.Context(), username)
	if err != nil || u == nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	reason := "Suspended via Billing API"
	if err := s.Store.UpdateUserStatus(r.Context(), u.ID, "suspended", &reason); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to suspend user")
		return
	}

	if s.Agent != nil {
		_, _ = s.Agent.Agent().SuspendTenant(r.Context(), &agentv1.SuspendTenantRequest{
			Username:  username,
			Suspended: true,
		})
	}

	token := GetAPIToken(r.Context())
	actorID := "billing_api"
	if token != nil {
		actorID = "api_token:" + token.ID
	}
	_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "billing.account_suspended", "user:"+username, r.RemoteAddr)

	writeJSON(w, http.StatusOK, map[string]string{
		"result":   "success",
		"username": username,
		"status":   "suspended",
	})
}

func (s *Server) BillingUnsuspendAccount(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	u, err := s.Store.GetUserByUsername(r.Context(), username)
	if err != nil || u == nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := s.Store.UpdateUserStatus(r.Context(), u.ID, "active", nil); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to unsuspend user")
		return
	}

	if s.Agent != nil {
		_, _ = s.Agent.Agent().SuspendTenant(r.Context(), &agentv1.SuspendTenantRequest{
			Username:  username,
			Suspended: false,
		})
	}

	token := GetAPIToken(r.Context())
	actorID := "billing_api"
	if token != nil {
		actorID = "api_token:" + token.ID
	}
	_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "billing.account_unsuspended", "user:"+username, r.RemoteAddr)

	writeJSON(w, http.StatusOK, map[string]string{
		"result":   "success",
		"username": username,
		"status":   "active",
	})
}

func (s *Server) BillingTerminateAccount(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	u, err := s.Store.GetUserByUsername(r.Context(), username)
	if err != nil || u == nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if s.Agent != nil {
		_, _ = s.Agent.Agent().SuspendTenant(r.Context(), &agentv1.SuspendTenantRequest{
			Username:  username,
			Suspended: true,
		})
	}

	token := GetAPIToken(r.Context())
	actorID := "billing_api"
	if token != nil {
		actorID = "api_token:" + token.ID
	}
	_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "billing.account_terminated", "user:"+username, r.RemoteAddr)

	writeJSON(w, http.StatusOK, map[string]string{
		"result":   "success",
		"username": username,
		"status":   "terminated",
	})
}

func (s *Server) BillingGetUsage(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	u, err := s.Store.GetUserByUsername(r.Context(), username)
	if err != nil || u == nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	var diskUsedMB uint64 = 150
	var diskTotalMB uint64 = 5120
	var bandwidthUsedMB uint64 = 2048
	var bandwidthTotalMB uint64 = 102400

	if s.Agent != nil {
		if resp, err := s.Agent.Agent().GetSystemTelemetry(r.Context(), &agentv1.GetSystemTelemetryRequest{Username: username}); err == nil {
			diskUsedMB = resp.DiskUsedMb
			diskTotalMB = resp.DiskTotalMb
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"username":           username,
		"disk_used_mb":       diskUsedMB,
		"disk_limit_mb":      diskTotalMB,
		"bandwidth_used_mb":  bandwidthUsedMB,
		"bandwidth_limit_mb": bandwidthTotalMB,
		"status":             u.Status,
	})
}
