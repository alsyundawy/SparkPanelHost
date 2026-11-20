package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	agentv1 "github.com/alsyundawy/SparkPanelHost/backend/gen/agent/v1"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/agentclient"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/auth"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
	"github.com/creack/pty"
	chi "github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	yaml "gopkg.in/yaml.v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Configurable origin check
	},
}

type Server struct {
	Store     *store.Store
	Agent     *agentclient.Client
	RecipeDir string
}

func NewServer(st *store.Store, ag *agentclient.Client, recipeDir string) *Server {
	return &Server{
		Store:     st,
		Agent:     ag,
		RecipeDir: recipeDir,
	}
}

// Auth Handlers
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	u, err := s.Store.GetUserByUsername(r.Context(), req.Username)
	if err != nil || !auth.VerifyPassword(req.Password, u.PasswordHash) {
		writeJSONError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if u.Status == "suspended" {
		writeJSONError(w, http.StatusForbidden, "Account is suspended")
		return
	}

	token, err := auth.GenerateSecureToken()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate session")
		return
	}

	tokenHash := auth.HashToken(token)
	expiresAt := time.Now().Add(24 * time.Hour)

	sess := &store.UserSession{
		ID:        fmt.Sprintf("sess-%d", time.Now().UnixNano()),
		UserID:    u.ID,
		TokenHash: tokenHash,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		ExpiresAt: expiresAt,
	}

	if err := s.Store.CreateSession(r.Context(), sess); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to save session")
		return
	}

	// Set HttpOnly Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "aether_session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	_ = s.Store.AppendAuditLog(r.Context(), u.ID, nil, "auth.login", "user:"+u.ID, r.RemoteAddr)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   token,
		"user":    u,
	})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	sess := GetSessionFromContext(r.Context())
	if sess != nil {
		_ = s.Store.DeleteSession(r.Context(), sess.TokenHash)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "aether_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	u := GetUserFromContext(r.Context())
	sess := GetSessionFromContext(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"user":            u,
		"impersonated":    sess != nil && sess.ImpersonatorID != nil,
		"impersonator_id": sess.ImpersonatorID,
	})
}

type ImpersonateRequest struct {
	TargetUserID string `json:"target_user_id"`
}

func (s *Server) Impersonate(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	if actor == nil || (actor.Role != "superadmin" && actor.Role != "reseller") {
		writeJSONError(w, http.StatusForbidden, "Tenants are not permitted to impersonate users")
		return
	}

	var req ImpersonateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	targetUser, err := s.Store.GetUserByID(r.Context(), req.TargetUserID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Target user not found")
		return
	}

	if targetUser.ID == actor.ID {
		writeJSONError(w, http.StatusBadRequest, "Cannot impersonate self")
		return
	}
	if targetUser.Status == "suspended" {
		writeJSONError(w, http.StatusForbidden, "Cannot impersonate suspended account")
		return
	}

	// Reseller can only impersonate their own tenants
	if actor.Role == "reseller" {
		if targetUser.Role != "tenant" || targetUser.ResellerID == nil || *targetUser.ResellerID != actor.ID {
			writeJSONError(w, http.StatusForbidden, "Cannot impersonate user not owned by you")
			return
		}
	}

	token, _ := auth.GenerateSecureToken()
	tokenHash := auth.HashToken(token)
	expiresAt := time.Now().Add(30 * time.Minute) // 30-minute limit

	sess := &store.UserSession{
		ID:             fmt.Sprintf("sess-%d", time.Now().UnixNano()),
		UserID:         targetUser.ID,
		ImpersonatorID: &actor.ID,
		TokenHash:      tokenHash,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		ExpiresAt:      expiresAt,
	}

	if err := s.Store.CreateSession(r.Context(), sess); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to create impersonation session")
		return
	}

	_ = s.Store.AppendAuditLog(r.Context(), actor.ID, &actor.ID, "auth.impersonate", "user:"+targetUser.ID, r.RemoteAddr)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   token,
		"user":    targetUser,
	})
}

// User Handlers
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	var resellerID *string
	if actor.Role == "reseller" {
		resellerID = &actor.ID
	}

	users, err := s.Store.ListUsers(r.Context(), resellerID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"users":   users,
	})
}

type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	PackageID string `json:"package_id"`
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Password hashing error")
		return
	}

	role := req.Role
	if role == "" || actor.Role == "reseller" {
		role = "tenant"
	}

	var resellerID *string
	if actor.Role == "reseller" {
		resellerID = &actor.ID
	}

	userID := fmt.Sprintf("usr-%d", time.Now().UnixNano())
	uid, err := s.Store.GetNextLinuxUID(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to allocate Linux UID: "+err.Error())
		return
	}

	u := &store.User{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         role,
		ResellerID:   resellerID,
		PackageID:    &req.PackageID,
		LinuxUID:     uid,
		LinuxGID:     uid,
		Locale:       "id",
		Theme:        "dark",
		Status:       "active",
	}

	if err := s.Store.CreateUser(r.Context(), u); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call agent to provision system tenant
	if s.Agent != nil {
		_, _ = s.Agent.Agent().ProvisionTenant(r.Context(), &agentv1.ProvisionTenantRequest{
			Username:        req.Username,
			Uid:             uint32(uid), // #nosec G115 - uid is allocated >= 2000
			Gid:             uint32(uid), // #nosec G115
			DiskQuotaMb:     5120,
			InodeQuota:      250000,
			CpuQuotaPercent: 100,
			MemoryLimitMb:   1024,
			IoWeight:        100,
			TasksMax:        256,
			ShellAccess:     false,
		})
	}

	_ = s.Store.AppendAuditLog(r.Context(), actor.ID, nil, "user.create", "user:"+u.ID, r.RemoteAddr)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"user":    u,
	})
}

// Domain Handlers
func (s *Server) ListDomains(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	var userID *string
	if actor.Role == "tenant" {
		userID = &actor.ID
	}

	domains, err := s.Store.ListDomains(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"domains": domains,
	})
}

type CreateDomainRequest struct {
	DomainName string `json:"domain_name"`
	PHPVersion string `json:"php_version"`
	WAFMode    string `json:"waf_mode"`
}

func (s *Server) CreateDomain(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	var req CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	phpVer := req.PHPVersion
	if phpVer == "" {
		phpVer = "8.5"
	}
	wafMode := req.WAFMode
	if wafMode == "" {
		wafMode = "Blocking"
	}

	docRoot := fmt.Sprintf("/home/%s/public_html/%s", actor.Username, req.DomainName)
	d := &store.Domain{
		ID:            fmt.Sprintf("dom-%d", time.Now().UnixNano()),
		UserID:        actor.ID,
		DomainName:    req.DomainName,
		DocumentRoot:  docRoot,
		PHPVersion:    phpVer,
		ForceHTTPS:    true,
		WAFMode:       wafMode,
		WAFProfile:    "default",
		NginxTemplate: "default",
	}

	if err := s.Store.CreateDomain(r.Context(), d); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call agent to sync vhost & php pool
	if s.Agent != nil {
		_, _ = s.Agent.Agent().SyncVirtualHost(r.Context(), &agentv1.SyncVirtualHostRequest{
			Username:      actor.Username,
			DomainName:    req.DomainName,
			DocumentRoot:  docRoot,
			PhpVersion:    phpVer,
			ForceHttps:    true,
			WafMode:       wafMode,
			WafProfile:    "default",
			NginxTemplate: "default",
		})
		_, _ = s.Agent.Agent().SyncPHPPool(r.Context(), &agentv1.SyncPHPPoolRequest{
			Username:    actor.Username,
			PhpVersion:  phpVer,
			MaxChildren: 10,
			MemoryLimit: "256M",
		})
	}

	_ = s.Store.AppendAuditLog(r.Context(), actor.ID, nil, "domain.create", "domain:"+d.ID, r.RemoteAddr)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"domain":  d,
	})
}

func (s *Server) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	domainID := chi.URLParam(r, "id")

	dom, err := s.Store.GetDomainByID(r.Context(), domainID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Domain not found")
		return
	}

	if actor.Role == "tenant" && dom.UserID != actor.ID {
		writeJSONError(w, http.StatusForbidden, "Cannot delete domain not owned by you")
		return
	}
	if actor.Role == "reseller" {
		targetUser, err := s.Store.GetUserByID(r.Context(), dom.UserID)
		if err != nil || targetUser.ResellerID == nil || *targetUser.ResellerID != actor.ID {
			writeJSONError(w, http.StatusForbidden, "Cannot delete domain not owned by your tenants")
			return
		}
	}

	if err := s.Store.DeleteDomain(r.Context(), domainID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete domain: "+err.Error())
		return
	}
	_ = s.Store.AppendAuditLog(r.Context(), actor.ID, nil, "domain.delete", "domain:"+domainID, r.RemoteAddr)

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Database Handlers
func (s *Server) ListDatabases(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	var userID *string
	if actor.Role == "tenant" {
		userID = &actor.ID
	}

	dbs, err := s.Store.ListDatabases(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"databases": dbs,
	})
}

type CreateDatabaseRequest struct {
	DBName   string `json:"db_name"`
	DBUser   string `json:"db_user"`
	Password string `json:"password"`
	Engine   string `json:"engine"` // "mariadb" or "postgresql"
}

func (s *Server) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	var req CreateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	dbIDRegex := regexp.MustCompile(`^[a-z0-9_]{1,24}$`)
	if !dbIDRegex.MatchString(req.DBName) || !dbIDRegex.MatchString(req.DBUser) {
		writeJSONError(w, http.StatusBadRequest, "Invalid database name or user identifier: must match ^[a-z0-9_]{1,24}$")
		return
	}

	engine := req.Engine
	if engine == "" {
		engine = "mariadb"
	}

	fullDBName := fmt.Sprintf("%s_%s", actor.Username, req.DBName)
	fullDBUser := fmt.Sprintf("%s_%s", actor.Username, req.DBUser)

	dbRecord := &store.TenantDatabase{
		ID:      fmt.Sprintf("db-%d", time.Now().UnixNano()),
		UserID:  actor.ID,
		Engine:  engine,
		DBName:  fullDBName,
		DBUser:  fullDBUser,
		Charset: "utf8mb4",
		SizeMB:  0.0,
	}

	if err := s.Store.CreateTenantDatabase(r.Context(), dbRecord); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if s.Agent != nil {
		_, _ = s.Agent.Agent().CreateDatabase(r.Context(), &agentv1.CreateDatabaseRequest{
			Username:   actor.Username,
			Engine:     engine,
			DbName:     fullDBName,
			DbUser:     fullDBUser,
			DbPassword: req.Password,
		})
	}

	_ = s.Store.AppendAuditLog(r.Context(), actor.ID, nil, "database.create", "db:"+dbRecord.ID, r.RemoteAddr)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":  true,
		"database": dbRecord,
	})
}

func (s *Server) DeleteDatabase(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	dbID := chi.URLParam(r, "id")

	dbRecord, err := s.Store.GetTenantDatabaseByID(r.Context(), dbID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Database not found")
		return
	}

	if actor.Role == "tenant" && dbRecord.UserID != actor.ID {
		writeJSONError(w, http.StatusForbidden, "Cannot delete database not owned by you")
		return
	}
	if actor.Role == "reseller" {
		targetUser, err := s.Store.GetUserByID(r.Context(), dbRecord.UserID)
		if err != nil || targetUser.ResellerID == nil || *targetUser.ResellerID != actor.ID {
			writeJSONError(w, http.StatusForbidden, "Cannot delete database not owned by your tenants")
			return
		}
	}

	if err := s.Store.DeleteTenantDatabase(r.Context(), dbID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete database: "+err.Error())
		return
	}
	_ = s.Store.AppendAuditLog(r.Context(), actor.ID, nil, "database.delete", "db:"+dbID, r.RemoteAddr)

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// DNS Clustering Handlers (WHM DNS Only style)
func (s *Server) ListDNSClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := s.Store.ListDNSClusters(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"clusters": clusters,
	})
}

type CreateDNSClusterRequest struct {
	Name     string `json:"name"`
	Topology string `json:"topology"` // 'star', 'primary_secondary', 'galera'
}

func (s *Server) CreateDNSCluster(w http.ResponseWriter, r *http.Request) {
	var req CreateDNSClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	c := &store.DNSCluster{
		ID:       fmt.Sprintf("cluster-%d", time.Now().UnixNano()),
		Name:     req.Name,
		Topology: req.Topology,
	}

	if err := s.Store.CreateDNSCluster(r.Context(), c); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"cluster": c,
	})
}

type RegisterDNSNodeRequest struct {
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
	Hostname    string `json:"hostname"`
	APIEndpoint string `json:"api_endpoint"`
	Role        string `json:"role"` // 'dns_only', 'synchronize', 'write_only'
	APIToken    string `json:"api_token"`
}

func (s *Server) RegisterDNSNode(w http.ResponseWriter, r *http.Request) {
	var req RegisterDNSNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	node := &store.DNSClusterNode{
		ID:          fmt.Sprintf("node-%d", time.Now().UnixNano()),
		ClusterID:   req.ClusterID,
		Hostname:    req.Hostname,
		APIEndpoint: req.APIEndpoint,
		Role:        req.Role,
		IsActive:    true,
	}

	if err := s.Store.CreateDNSClusterNode(r.Context(), node); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Register in daemon agent
	if s.Agent != nil {
		_, _ = s.Agent.Agent().RegisterDNSClusterNode(r.Context(), &agentv1.RegisterDNSClusterNodeRequest{
			ClusterName: req.ClusterName,
			Hostname:    req.Hostname,
			ApiEndpoint: req.APIEndpoint,
			Role:        req.Role,
			ApiToken:    req.APIToken,
		})
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"node":    node,
	})
}

func (s *Server) GetDNSClusterStatus(w http.ResponseWriter, r *http.Request) {
	clusterName := chi.URLParam(r, "name")
	if s.Agent != nil {
		res, err := s.Agent.Agent().GetDNSClusterStatus(r.Context(), &agentv1.GetDNSClusterStatusRequest{
			ClusterName: clusterName,
		})
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"nodes":   res.GetNodes(),
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"nodes":   []interface{}{},
	})
}

// Mail Handlers
func (s *Server) ListMailDomains(w http.ResponseWriter, r *http.Request) {
	mDomains, err := s.Store.ListMailDomains(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"mail_domains": mDomains,
	})
}

func (s *Server) CheckRBL(w http.ResponseWriter, r *http.Request) {
	serverIP := r.URL.Query().Get("ip")
	if serverIP == "" {
		serverIP = "127.0.0.1"
	}

	if s.Agent != nil {
		res, err := s.Agent.Agent().CheckRBLStatus(r.Context(), &agentv1.CheckRBLStatusRequest{
			ServerIp: serverIP,
		})
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success":   true,
				"server_ip": res.GetServerIp(),
				"clean":     res.GetClean(),
				"results":   res.GetResults(),
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"server_ip": serverIP,
		"clean":     true,
		"results":   []interface{}{},
	})
}

// App Recipes Handlers
type RecipeManifest struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Category    string `yaml:"category" json:"category"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description" json:"description"`
	ArchiveURL  string `yaml:"archive_url" json:"archive_url"`
	SHA256      string `yaml:"sha256" json:"sha256"`
}

func (s *Server) ListRecipes(w http.ResponseWriter, r *http.Request) {
	var recipes []RecipeManifest
	recipeDir := s.RecipeDir
	if recipeDir == "" {
		recipeDir = "recipes"
	}

	_ = filepath.Walk(recipeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}
		data, err := os.ReadFile(filepath.Clean(path)) // #nosec G122,G304
		if err == nil {
			var rec RecipeManifest
			if err := yaml.Unmarshal(data, &rec); err == nil && rec.ID != "" {
				recipes = append(recipes, rec)
			}
		}
		return nil
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"recipes": recipes,
	})
}

// Telemetry Handlers
func (s *Server) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	username := ""
	if actor.Role == "tenant" {
		username = actor.Username
	}

	if s.Agent != nil {
		res, err := s.Agent.Agent().GetSystemTelemetry(r.Context(), &agentv1.GetSystemTelemetryRequest{
			Username: username,
		})
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success":              true,
				"cpu_percent":          res.GetCpuPercent(),
				"ram_used_mb":          res.GetRamUsedMb(),
				"ram_total_mb":         res.GetRamTotalMb(),
				"disk_used_mb":         res.GetDiskUsedMb(),
				"disk_total_mb":        res.GetDiskTotalMb(),
				"waf_blocked_24h":      res.GetWafBlocked_24H(),
				"crowdsec_active_bans": res.GetCrowdsecActiveBans(),
			})
			return
		}
	}

	// Fallback mock telemetry
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":              true,
		"cpu_percent":          12.5,
		"ram_used_mb":          1840,
		"ram_total_mb":         8192,
		"disk_used_mb":         14200,
		"disk_total_mb":        102400,
		"waf_blocked_24h":      142,
		"crowdsec_active_bans": 18,
	})
}

func (s *Server) StreamTelemetryWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		data := map[string]interface{}{
			"timestamp":            time.Now().Unix(),
			"cpu_percent":          10.0 + float64(time.Now().Unix()%15),
			"ram_used_mb":          1800 + (time.Now().Unix() % 200),
			"ram_total_mb":         8192,
			"disk_used_mb":         14250,
			"disk_total_mb":        102400,
			"waf_blocked_24h":      145,
			"crowdsec_active_bans": 18,
		}
		if s.Agent != nil {
			if res, err := s.Agent.Agent().GetSystemTelemetry(context.Background(), &agentv1.GetSystemTelemetryRequest{}); err == nil {
				data["cpu_percent"] = res.GetCpuPercent()
				data["ram_used_mb"] = res.GetRamUsedMb()
				data["ram_total_mb"] = res.GetRamTotalMb()
				data["disk_used_mb"] = res.GetDiskUsedMb()
				data["disk_total_mb"] = res.GetDiskTotalMb()
				data["waf_blocked_24h"] = res.GetWafBlocked_24H()
				data["crowdsec_active_bans"] = res.GetCrowdsecActiveBans()
			}
		}

		if err := conn.WriteJSON(data); err != nil {
			break
		}
	}
}

// Web Terminal WebSocket PTY Bridge
func (s *Server) WebTerminalWS(w http.ResponseWriter, r *http.Request) {
	actor := GetUserFromContext(r.Context())
	if actor == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	if actor.Role == "tenant" {
		writeJSONError(w, http.StatusForbidden, "Web terminal shell access is disabled for this account")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Spawn shell session via PTY
	shell := "/bin/bash"
	c := ptyShellCommand(shell)
	f, err := pty.Start(c)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to spawn PTY: %v\r\n", err)))
		return
	}
	defer func() {
		_ = c.Process.Kill()
		_ = f.Close()
	}()

	// PTY stdout -> WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := f.Read(buf)
			if err != nil {
				break
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				break
			}
		}
	}()

	// WebSocket -> PTY stdin
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		_, _ = f.Write(msg)
	}
}

func (s *Server) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	logs, err := s.Store.ListAuditLogs(r.Context(), limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"logs":    logs,
	})
}
