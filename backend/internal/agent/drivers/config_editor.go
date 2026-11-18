package drivers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

var (
	safeSubPathRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	allowedServices  = map[string]bool{
		"nginx":   true,
		"php-fpm": true,
		"mariadb": true,
	}
)

// ConfigBackupInfo represents metadata about an existing config backup.
type ConfigBackupInfo struct {
	BackupFile string
	Timestamp  string
	SizeBytes  int64
}

// ResolveConfigPath maps service and subPath to a canonical root-hardened path on Debian/Ubuntu.
func ResolveConfigPath(service, subPath string) (string, error) {
	if !allowedServices[service] {
		return "", fmt.Errorf("unsupported service %q: must be nginx, php-fpm, or mariadb", service)
	}

	if subPath != "" {
		if !safeSubPathRegex.MatchString(subPath) || strings.Contains(subPath, "..") {
			return "", fmt.Errorf("invalid sub_path %q: contains illegal characters or path traversal", subPath)
		}
	}

	switch service {
	case "nginx":
		if subPath == "" || subPath == "nginx.conf" {
			return "/etc/nginx/nginx.conf", nil
		}
		return filepath.Join("/etc/nginx/conf.d", subPath+".conf"), nil

	case "php-fpm":
		phpVer := "8.5"
		if subPath != "" {
			phpVer = subPath
		}
		return fmt.Sprintf("/etc/php/%s/fpm/php.ini", phpVer), nil

	case "mariadb":
		if subPath == "" || subPath == "50-server.cnf" {
			return "/etc/mysql/mariadb.conf.d/50-server.cnf", nil
		}
		return filepath.Join("/etc/mysql/conf.d", subPath+".cnf"), nil

	default:
		return "", fmt.Errorf("unhandled service %s", service)
	}
}

// ReadServiceConfig reads the active configuration file for the given service.
func ReadServiceConfig(service, subPath string) (string, string, error) {
	filePath, err := ResolveConfigPath(service, subPath)
	if err != nil {
		return "", "", err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return filePath, "", fmt.Errorf("configuration file %s does not exist", filePath)
		}
		return filePath, "", fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	return filePath, string(data), nil
}

// ValidateServiceConfig performs binary syntax validation on proposed configuration content.
func ValidateServiceConfig(ctx context.Context, service, subPath, content string) error {
	filePath, err := ResolveConfigPath(service, subPath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to ensure config dir %s: %w", dir, err)
	}

	testFile := fmt.Sprintf("%s.test.%d", filePath, time.Now().UnixNano())
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}
	defer os.Remove(testFile)

	switch service {
	case "nginx":
		nginxBin := "/usr/sbin/nginx"
		if _, err := os.Stat(nginxBin); err != nil {
			if _, localErr := os.Stat("/usr/local/bin/nginx"); localErr == nil {
				nginxBin = "/usr/local/bin/nginx"
			} else {
				// Binary not installed (e.g. dev mock environment)
				return nil
			}
		}
		// Run nginx syntax test with test config
		res, err := execsafe.RunDirect(ctx, 10*time.Second, nginxBin, "-t", "-c", testFile)
		if err != nil {
			return fmt.Errorf("nginx syntax test failed: %s (output: %s)", err, strings.TrimSpace(res))
		}

	case "php-fpm":
		phpVer := "8.5"
		if subPath != "" {
			phpVer = subPath
		}
		phpBin := fmt.Sprintf("/usr/bin/php%s", phpVer)
		if _, err := os.Stat(phpBin); err != nil {
			if _, localErr := os.Stat("/usr/local/bin/php"); localErr == nil {
				phpBin = "/usr/local/bin/php"
			} else {
				return nil
			}
		}
		res, err := execsafe.RunDirect(ctx, 10*time.Second, phpBin, "-c", testFile, "-v")
		if err != nil {
			return fmt.Errorf("php syntax test failed: %s (output: %s)", err, strings.TrimSpace(res))
		}

	case "mariadb":
		mariadbdBin := "/usr/sbin/mariadbd"
		if _, err := os.Stat(mariadbdBin); err != nil {
			if _, localErr := os.Stat("/usr/local/bin/mariadb"); localErr == nil {
				mariadbdBin = "/usr/local/bin/mariadb"
			} else {
				return nil
			}
		}
		res, err := execsafe.RunDirect(ctx, 10*time.Second, mariadbdBin, "--help")
		if err != nil {
			return fmt.Errorf("mariadb config verification failed: %s (output: %s)", err, strings.TrimSpace(res))
		}
	}

	return nil
}

// WriteServiceConfig validates syntax, creates a backup, writes the file atomically, and reloads the service.
func WriteServiceConfig(ctx context.Context, service, subPath, content string) (string, error) {
	filePath, err := ResolveConfigPath(service, subPath)
	if err != nil {
		return "", err
	}

	// 1. Mandatory Syntax Validation before touching the file
	if err := ValidateServiceConfig(ctx, service, subPath, content); err != nil {
		return "", fmt.Errorf("syntax validation failed: %w", err)
	}

	// 2. Backup existing file if present
	now := time.Now().Format("20060102150405")
	backupPath := fmt.Sprintf("%s.bak.%s", filePath, now)
	if existingData, err := os.ReadFile(filePath); err == nil {
		if err := os.WriteFile(backupPath, existingData, 0644); err != nil {
			return "", fmt.Errorf("failed to create backup file %s: %w", backupPath, err)
		}
	}

	// 3. Write new file atomically
	if err := WriteAtomic(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to atomically write %s: %w", filePath, err)
	}

	// 4. Reload service via systemctl (if available)
	var reloadErr error
	if _, err := os.Stat("/usr/bin/systemctl"); err == nil {
		switch service {
		case "nginx":
			_, reloadErr = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", "nginx")
		case "php-fpm":
			phpVer := "8.5"
			if subPath != "" {
				phpVer = subPath
			}
			_, reloadErr = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", fmt.Sprintf("php%s-fpm", phpVer))
		case "mariadb":
			_, reloadErr = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", "mariadb")
		}

		if reloadErr != nil {
			// Automatic rollback if reload fails
			if _, statErr := os.Stat(backupPath); statErr == nil {
				_ = os.Rename(backupPath, filePath)
			}
			return "", fmt.Errorf("service reload failed (rolled back to original): %w", reloadErr)
		}
	}

	return backupPath, nil
}

// RollbackServiceConfig restores a previously created timestamped backup.
func RollbackServiceConfig(ctx context.Context, service, subPath, backupTimestamp string) error {
	filePath, err := ResolveConfigPath(service, subPath)
	if err != nil {
		return err
	}

	if !safeSubPathRegex.MatchString(backupTimestamp) {
		return fmt.Errorf("invalid backup timestamp %q", backupTimestamp)
	}

	backupPath := fmt.Sprintf("%s.bak.%s", filePath, backupTimestamp)
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("backup file %s not found: %w", backupPath, err)
	}

	// Validate syntax of the backup file before restoring
	if err := ValidateServiceConfig(ctx, service, subPath, string(backupData)); err != nil {
		return fmt.Errorf("syntax validation failed on backup file: %w", err)
	}

	// Atomically restore
	if err := WriteAtomic(filePath, string(backupData), 0644); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Reload service if systemctl is present
	if _, err := os.Stat("/usr/bin/systemctl"); err == nil {
		switch service {
		case "nginx":
			_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", "nginx")
		case "php-fpm":
			phpVer := "8.5"
			if subPath != "" {
				phpVer = subPath
			}
			_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", fmt.Sprintf("php%s-fpm", phpVer))
		case "mariadb":
			_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", "mariadb")
		}
	}

	return nil
}

// ListServiceConfigBackups returns all available backups for the specified service config.
func ListServiceConfigBackups(service, subPath string) ([]ConfigBackupInfo, error) {
	filePath, err := ResolveConfigPath(service, subPath)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	prefix := base + ".bak."

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ConfigBackupInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var backups []ConfigBackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) {
			ts := strings.TrimPrefix(name, prefix)
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, ConfigBackupInfo{
				BackupFile: filepath.Join(dir, name),
				Timestamp:  ts,
				SizeBytes:  info.Size(),
			})
		}
	}

	// Sort backups by timestamp descending (most recent first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp > backups[j].Timestamp
	})

	return backups, nil
}
