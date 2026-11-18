package drivers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	return os.MkdirAll(path, perm)
}

// WriteAtomic writes content to a temporary file and atomically renames it to destination.
func WriteAtomic(destPath, content string, perm os.FileMode) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	tmpFile := fmt.Sprintf("%s.tmp.%d", destPath, time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, []byte(content), perm); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, destPath); err != nil {
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
