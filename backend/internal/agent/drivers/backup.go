package drivers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

// CreateBackup runs a Restic snapshot for the user's home and data.
func CreateBackup(ctx context.Context, username, destinationID string, includeDBs, includeMail bool) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}

	userHome := "/home/" + username
	args := []string{
		"backup",
		userHome,
		"--tag", "user:" + username,
	}

	if includeMail {
		mailDir := "/var/vmail/" + username
		if _, err := os.Stat(mailDir); err == nil {
			args = append(args, mailDir)
		}
	}

	_, err := execsafe.RunDirect(ctx, 30*time.Minute, "/usr/bin/restic", args...)
	return err
}

// RestoreBackup restores a Restic snapshot directly to the user's home directory.
func RestoreBackup(ctx context.Context, username, snapshotID, destinationID string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	if snapshotID == "" {
		return fmt.Errorf("snapshot_id is required")
	}

	args := []string{
		"restore",
		snapshotID,
		"--target", "/home/" + username,
	}

	_, err := execsafe.RunDirect(ctx, 30*time.Minute, "/usr/bin/restic", args...)
	return err
}

// TestRestoreBackup executes an automated test-restore (3-2-1 verification) in a staging directory.
func TestRestoreBackup(ctx context.Context, snapshotID, destinationID string) error {
	if snapshotID == "" {
		return fmt.Errorf("snapshot_id is required")
	}

	testTarget := fmt.Sprintf("/var/panel/tmp/test-restore-%s-%d", snapshotID, time.Now().Unix())
	if err := os.MkdirAll(testTarget, 0700); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(testTarget)
	}()

	args := []string{
		"restore",
		snapshotID,
		"--target", testTarget,
		"--verify",
	}

	_, err := execsafe.RunDirect(ctx, 15*time.Minute, "/usr/bin/restic", args...)
	return err
}
