package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	agentv1 "github.com/alsyundawy/SparkPanelHost/backend/gen/agent/v1"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/drivers"
	chi "github.com/go-chi/chi/v5"
)

type UpdateConfigRequest struct {
	SubPath string `json:"sub_path"`
	Content string `json:"content"`
}

type RollbackConfigRequest struct {
	SubPath         string `json:"sub_path"`
	BackupTimestamp string `json:"backup_timestamp"`
}

// GetConfig reads the active configuration for the specified service.
func (s *Server) GetConfig(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")
	subPath := r.URL.Query().Get("sub_path")

	if s.Agent != nil {
		resp, err := s.Agent.Agent().ReadServiceConfig(r.Context(), &agentv1.ReadServiceConfigRequest{
			Service: service,
			SubPath: subPath,
		})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to read service config: %v", err))
			return
		}
		if !resp.GetSuccess() {
			writeJSONError(w, http.StatusBadRequest, resp.GetMessage())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":   true,
			"service":   service,
			"sub_path":  subPath,
			"file_path": resp.GetFilePath(),
			"content":   resp.GetContent(),
		})
		return
	}

	// Dev / Fallback direct driver execution
	filePath, content, err := drivers.ReadServiceConfig(service, subPath)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"service":   service,
		"sub_path":  subPath,
		"file_path": filePath,
		"content":   content,
	})
}

// UpdateConfig validates syntax, creates a backup, and atomically writes the configuration.
func (s *Server) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u := GetUserFromContext(r.Context())
	actorID := ""
	if u != nil {
		actorID = u.ID
	}

	if s.Agent != nil {
		resp, err := s.Agent.Agent().WriteServiceConfig(r.Context(), &agentv1.WriteServiceConfigRequest{
			Service: service,
			SubPath: req.SubPath,
			Content: req.Content,
		})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("daemon error: %v", err))
			return
		}
		if !resp.GetSuccess() {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("syntax error: %s", resp.GetMessage()))
			return
		}

		if s.Store != nil {
			_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "config.update", fmt.Sprintf("%s:%s", service, req.SubPath), r.RemoteAddr)
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":     true,
			"message":     resp.GetMessage(),
			"backup_path": resp.GetBackupPath(),
		})
		return
	}

	// Fallback direct driver execution
	backupPath, err := drivers.WriteServiceConfig(r.Context(), service, req.SubPath, req.Content)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("syntax/write error: %v", err))
		return
	}

	if s.Store != nil {
		_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "config.update", fmt.Sprintf("%s:%s", service, req.SubPath), r.RemoteAddr)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"message":     "Configuration updated and service reloaded successfully",
		"backup_path": backupPath,
	})
}

// RollbackConfig restores a previously created timestamped configuration backup.
func (s *Server) RollbackConfig(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")

	var req RollbackConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.BackupTimestamp == "" {
		writeJSONError(w, http.StatusBadRequest, "backup_timestamp is required")
		return
	}

	u := GetUserFromContext(r.Context())
	actorID := ""
	if u != nil {
		actorID = u.ID
	}

	if s.Agent != nil {
		resp, err := s.Agent.Agent().RollbackServiceConfig(r.Context(), &agentv1.RollbackServiceConfigRequest{
			Service:         service,
			SubPath:         req.SubPath,
			BackupTimestamp: req.BackupTimestamp,
		})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("daemon error: %v", err))
			return
		}
		if !resp.GetSuccess() {
			writeJSONError(w, http.StatusBadRequest, resp.GetMessage())
			return
		}

		if s.Store != nil {
			_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "config.rollback", fmt.Sprintf("%s:%s (ts: %s)", service, req.SubPath, req.BackupTimestamp), r.RemoteAddr)
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": resp.GetMessage(),
		})
		return
	}

	// Fallback direct driver execution
	if err := drivers.RollbackServiceConfig(r.Context(), service, req.SubPath, req.BackupTimestamp); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if s.Store != nil {
		_ = s.Store.AppendAuditLog(r.Context(), actorID, nil, "config.rollback", fmt.Sprintf("%s:%s (ts: %s)", service, req.SubPath, req.BackupTimestamp), r.RemoteAddr)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Config rolled back to %s successfully", req.BackupTimestamp),
	})
}

// ListConfigBackups lists available backups for a service configuration file.
func (s *Server) ListConfigBackups(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")
	subPath := r.URL.Query().Get("sub_path")

	if s.Agent != nil {
		resp, err := s.Agent.Agent().ListServiceConfigBackups(r.Context(), &agentv1.ListServiceConfigBackupsRequest{
			Service: service,
			SubPath: subPath,
		})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("daemon error: %v", err))
			return
		}
		if !resp.GetSuccess() {
			writeJSONError(w, http.StatusBadRequest, resp.GetMessage())
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"backups": resp.GetBackups(),
		})
		return
	}

	// Fallback direct driver execution
	backups, err := drivers.ListServiceConfigBackups(service, subPath)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"backups": backups,
	})
}
