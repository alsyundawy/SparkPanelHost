package drivers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

var usernameRegex = regexp.MustCompile(`^[a-z][a-z0-9]{2,15}$`)

// ValidateUsername ensures the username complies with strict Linux username rules (3-16 chars, lowercase alphanumeric).
func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("invalid username %q: must match ^[a-z][a-z0-9]{2,15}$", username)
	}
	return nil
}

// EnsureDir creates a directory with specific permissions and ownership if not exists.
func EnsureDir(path string, perm os.FileMode) error {
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid directory path: %s", path)
	}
	return os.MkdirAll(cleanPath, perm)
}

// WriteAtomic writes content to a temporary file and atomically renames it to destination.
func WriteAtomic(destPath, content string, perm os.FileMode) error {
	cleanDest := filepath.Clean(destPath)
	if !filepath.IsAbs(cleanDest) || strings.Contains(cleanDest, "..") {
		return fmt.Errorf("invalid destination path: %s", destPath)
	}

	dir := filepath.Dir(cleanDest)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	tmpFile := filepath.Clean(fmt.Sprintf("%s.tmp.%d", cleanDest, time.Now().UnixNano()))
	if !strings.HasPrefix(tmpFile, dir) {
		return fmt.Errorf("security violation: temp file path escaped directory")
	}

	if err := os.WriteFile(tmpFile, []byte(content), perm); err != nil { // #nosec G306,G703
		return err
	}

	if err := os.Rename(tmpFile, cleanDest); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	return nil
}

// ReloadSystemd calls systemctl daemon-reload via execsafe.
func ReloadSystemd(ctx context.Context) error {
	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "daemon-reload")
	return err
}
