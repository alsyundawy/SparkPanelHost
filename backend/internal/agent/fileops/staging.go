package fileops

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/netfetch"
)

// InstallAppArchive downloads an upstream archive safely, checks SHA-256, and extracts as tenant.
func InstallAppArchive(ctx context.Context, username, archiveURL, expectedSHA256, targetDir string, stripComponents uint32) error {
	cleanTarget := filepath.Clean(targetDir)
	if !strings.HasPrefix(cleanTarget, "/home/"+username) {
		return fmt.Errorf("security violation: target_dir %q must be in /home/%s", cleanTarget, username)
	}

	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("lookup user %s failed: %w", username, err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	stagingDir := fmt.Sprintf("/var/panel/tmp/%s", username)
	_ = os.MkdirAll(stagingDir, 0700)
	stagingFile := filepath.Join(stagingDir, fmt.Sprintf("pkg-%d.tmp", time.Now().UnixNano()))
	defer os.Remove(stagingFile)

	// Download via SSRF-safe dialer (anti-CWE-918)
	if err := netfetch.DownloadVerifiedUpstream(ctx, archiveURL, stagingFile, 2*1024*1024*1024); err != nil {
		return fmt.Errorf("safe download failed: %w", err)
	}

	// Verify SHA-256 (anti-CWE-494)
	if expectedSHA256 != "" {
		if err := VerifySHA256(stagingFile, expectedSHA256); err != nil {
			return err
		}
	}

	// Ensure destination directory exists and is owned by tenant
	_ = os.MkdirAll(cleanTarget, 0750)
	_, _ = execsafe.RunDirect(ctx, 10*time.Second, "/usr/bin/chown", fmt.Sprintf("%d:%d", uid, gid), cleanTarget)

	// Extract safely with openat2 and privilege drop (anti-CWE-22, CWE-59, CWE-409)
	return ExtractArchiveAsUser(stagingFile, cleanTarget, uid, gid, int(stripComponents))
}

// CloneAppStaging duplicates an application directory and handles database cloning with URL replacement.
func CloneAppStaging(ctx context.Context, username, sourceDir, targetDir, sourceDB, targetDB, oldURL, newURL string) error {
	cleanSource := filepath.Clean(sourceDir)
	cleanTarget := filepath.Clean(targetDir)

	if !strings.HasPrefix(cleanSource, "/home/"+username) || !strings.HasPrefix(cleanTarget, "/home/"+username) {
		return fmt.Errorf("security violation: source and target must reside under /home/%s", username)
	}

	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("lookup user %s failed: %w", username, err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	// Copy directory contents
	_ = os.MkdirAll(cleanTarget, 0750)
	err = filepath.Walk(cleanSource, func(currPath string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(cleanSource, currPath)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		destPath := filepath.Join(cleanTarget, rel)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0750)
		}

		// Don't copy symlinks pointing outside
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Read and optionally replace URLs
		data, err := os.ReadFile(currPath)
		if err != nil {
			return err
		}

		if oldURL != "" && newURL != "" && isTextFile(currPath) {
			data = bytes.ReplaceAll(data, []byte(oldURL), []byte(newURL))
		}

		return os.WriteFile(destPath, data, info.Mode().Perm())
	})
	if err != nil {
		return fmt.Errorf("staging file copy failed: %w", err)
	}

	// Chown target to user
	_, _ = execsafe.RunDirect(ctx, 30*time.Second, "/usr/bin/chown", "-R", fmt.Sprintf("%d:%d", uid, gid), cleanTarget)

	// Clone database if provided
	if sourceDB != "" && targetDB != "" {
		dumpSQL, err := execsafe.RunDirect(ctx, 10*time.Minute, "/usr/bin/mariadb-dump",
			"--single-transaction",
			"--quick",
			sourceDB,
		)
		if err != nil {
			return fmt.Errorf("staging db dump failed: %w", err)
		}

		if oldURL != "" && newURL != "" {
			dumpSQL = strings.ReplaceAll(dumpSQL, oldURL, newURL)
		}

		dumpFile := fmt.Sprintf("/var/panel/tmp/%s/staging-db-%d.sql", username, time.Now().UnixNano())
		if err := os.WriteFile(dumpFile, []byte(dumpSQL), 0600); err != nil {
			return err
		}
		defer os.Remove(dumpFile)

		// Create target DB and import
		createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", targetDB)
		_, _ = execsafe.RunDirect(ctx, 30*time.Second, "/usr/bin/mariadb", "-e", createSQL)

		// Import via mariadb
		_, err = execsafe.RunDirect(ctx, 10*time.Minute, "/usr/bin/mariadb", targetDB, "-e", fmt.Sprintf("source %s", dumpFile))
		if err != nil {
			return fmt.Errorf("staging db restore failed: %w", err)
		}
	}

	return nil
}

func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".php", ".json", ".sql", ".env", ".xml", ".html", ".htm", ".txt", ".yaml", ".yml", ".js", ".ts", ".css":
		return true
	default:
		return false
	}
}
