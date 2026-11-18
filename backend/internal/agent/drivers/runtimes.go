package drivers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

type AppRuntimeConfig struct {
	Username    string
	RuntimeType string // "nodejs" or "python"
	Action      string // "start", "stop", "restart", "setup"
	AppName     string
	WorkingDir  string
	Entrypoint  string
	Port        uint32
}

// ManageAppRuntime orchestrates systemd unit generation and execution for Node.js & Python tenant apps.
func ManageAppRuntime(ctx context.Context, cfg AppRuntimeConfig) error {
	if err := ValidateUsername(cfg.Username); err != nil {
		return err
	}
	cleanAppName := strings.ToLower(strings.TrimSpace(cfg.AppName))
	if cleanAppName == "" {
		return fmt.Errorf("app_name cannot be empty")
	}

	serviceName := fmt.Sprintf("aether-app-%s-%s.service", cfg.Username, cleanAppName)
	servicePath := filepath.Join("/etc/systemd/system", serviceName)

	if cfg.Action == "setup" {
		cleanWorkDir := filepath.Clean(cfg.WorkingDir)
		if !strings.HasPrefix(cleanWorkDir, "/home/"+cfg.Username) {
			return fmt.Errorf("security violation: working_dir %q must be inside /home/%s", cleanWorkDir, cfg.Username)
		}

		unitContent := fmt.Sprintf(`[Unit]
Description=AetherPanel %s App - %s (%s)
After=network.target

[Service]
Type=simple
User=%s
Group=%s
WorkingDirectory=%s
ExecStart=%s
Restart=always
RestartSec=5
Environment=PORT=%d
Environment=NODE_ENV=production
Environment=PYTHONUNBUFFERED=1

[Install]
WantedBy=multi-user.target
`, cfg.RuntimeType, cleanAppName, cfg.Username, cfg.Username, cfg.Username, cleanWorkDir, cfg.Entrypoint, cfg.Port)

		if err := WriteAtomic(servicePath, unitContent, 0644); err != nil {
			return err
		}

		if err := ReloadSystemd(ctx); err != nil {
			return err
		}

		_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "enable", serviceName)
		_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "start", serviceName)
		return err
	}

	switch cfg.Action {
	case "start", "stop", "restart":
		_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", cfg.Action, serviceName)
		return err
	case "delete", "remove":
		_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "stop", serviceName)
		_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "disable", serviceName)
		_ = os.Remove(servicePath)
		return ReloadSystemd(ctx)
	default:
		return fmt.Errorf("unsupported action %q", cfg.Action)
	}
}

// SyncCronJob configures scheduled commands for a tenant user via crontab.
func SyncCronJob(ctx context.Context, username, action, schedule, command string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}

	// Read existing crontab
	existing, _ := execsafe.RunDirect(ctx, 10*time.Second, "/usr/bin/crontab", "-u", username, "-l")
	lines := strings.Split(existing, "\n")

	cleanCmd := strings.TrimSpace(command)
	targetLine := fmt.Sprintf("%s %s", schedule, cleanCmd)

	var newLines []string
	found := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if trimmed != "" {
				newLines = append(newLines, l)
			}
			continue
		}
		if strings.Contains(trimmed, cleanCmd) {
			if action != "remove" && action != "disable" {
				newLines = append(newLines, targetLine)
			}
			found = true
		} else {
			newLines = append(newLines, l)
		}
	}

	if action == "add" && !found {
		newLines = append(newLines, targetLine)
	}

	tmpFile := fmt.Sprintf("/var/panel/tmp/%s/cron-%d.tmp", username, time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, []byte(strings.Join(newLines, "\n")+"\n"), 0600); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	_, err := execsafe.RunDirect(ctx, 10*time.Second, "/usr/bin/crontab", "-u", username, tmpFile)
	return err
}

// ProvisionFTPAccount configures FTP access for a user.
func ProvisionFTPAccount(ctx context.Context, username, ftpUser, passwordHash, homeDir string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	cleanHome := filepath.Clean(homeDir)
	if !strings.HasPrefix(cleanHome, "/home/"+username) {
		return fmt.Errorf("security violation: FTP home %q must be in /home/%s", cleanHome, username)
	}
	// For pure-ftpd or vsftpd, we ensure permissions and directory exist
	return os.MkdirAll(cleanHome, 0755)
}

// DeployGitRepository performs git clone or git pull without shell execution.
func DeployGitRepository(ctx context.Context, username, repoURL, branch, deployDir string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	cleanDeploy := filepath.Clean(deployDir)
	if !strings.HasPrefix(cleanDeploy, "/home/"+username) {
		return fmt.Errorf("security violation: deploy_dir %q must be in /home/%s", cleanDeploy, username)
	}

	if branch == "" {
		branch = "main"
	}

	gitDir := filepath.Join(cleanDeploy, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Existing repo: git pull
		_, err := execsafe.RunDirect(ctx, 5*time.Minute, "/usr/bin/git",
			"-C", cleanDeploy,
			"pull", "origin", branch,
		)
		return err
	}

	// New repo: git clone
	_ = os.MkdirAll(cleanDeploy, 0755)
	_, err := execsafe.RunDirect(ctx, 5*time.Minute, "/usr/bin/git",
		"clone",
		"--branch", branch,
		"--depth", "1",
		repoURL,
		cleanDeploy,
	)
	return err
}
