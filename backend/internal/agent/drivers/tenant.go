package drivers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

type TenantLimits struct {
	DiskQuotaMB     uint64
	InodeQuota      uint64
	CPUQuotaPercent uint32
	MemoryLimitMB   uint64
	IOWeight        uint32
	TasksMax        uint32
	ShellAccess     bool
}

// ProvisionTenant provisions a Linux user account, quota, /var/panel/tmp, and cgroups v2 slice.
func ProvisionTenant(ctx context.Context, username string, uid, gid uint32, limits TenantLimits) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}

	shell := "/usr/sbin/nologin"
	if limits.ShellAccess {
		shell = "/bin/bash"
	}

	uidStr := strconv.FormatUint(uint64(uid), 10)
	gidStr := strconv.FormatUint(uint64(gid), 10)

	// useradd -u <uid> -g <gid> -m -d /home/<user> -s <shell> <user>
	_, err := execsafe.RunDirect(ctx, 30*time.Second, "/usr/sbin/useradd",
		"-u", uidStr,
		"-g", gidStr,
		"-m",
		"-d", "/home/"+username,
		"-s", shell,
		username,
	)
	if err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("useradd failed: %w", err)
	}

	// Prepare tenant home permissions (0750)
	_ = os.Chmod("/home/"+username, 0750)

	// Prepare /var/panel/tmp/<username> (0700) owned by tenant
	tmpDir := "/var/panel/tmp/" + username
	if err := os.MkdirAll(tmpDir, 0700); err == nil {
		_, _ = execsafe.RunDirect(ctx, 10*time.Second, "/usr/bin/chown", uidStr+":"+gidStr, tmpDir)
	}

	// Configure disk quota if tools exist
	if limits.DiskQuotaMB > 0 {
		_ = ApplyDiskQuota(ctx, username, limits.DiskQuotaMB, limits.InodeQuota)
	}

	// Configure systemd cgroups v2 slice
	return ApplyCgroupsSlice(ctx, uid, limits)
}

// UpdateTenantLimits updates disk quota and cgroups slice.
func UpdateTenantLimits(ctx context.Context, username string, uid uint32, limits TenantLimits) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	if limits.DiskQuotaMB > 0 {
		_ = ApplyDiskQuota(ctx, username, limits.DiskQuotaMB, limits.InodeQuota)
	}
	return ApplyCgroupsSlice(ctx, uid, limits)
}

// SuspendTenant locks/unlocks user account and terminates user processes.
func SuspendTenant(ctx context.Context, username string, suspended bool) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}

	if suspended {
		// Lock password and set nologin
		_, _ = execsafe.RunDirect(ctx, 10*time.Second, "/usr/sbin/usermod", "-L", "-s", "/usr/sbin/nologin", username)
		// Stop user systemd services or kill processes
		_, _ = execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "stop", "user-"+username+".slice")
	} else {
		// Unlock account
		_, _ = execsafe.RunDirect(ctx, 10*time.Second, "/usr/sbin/usermod", "-U", "-s", "/bin/bash", username)
	}
	return nil
}

// ApplyDiskQuota calls setquota for the tenant.
func ApplyDiskQuota(ctx context.Context, username string, diskMB, inodes uint64) error {
	// setquota -u <user> <soft_kb> <hard_kb> <soft_inodes> <hard_inodes> /
	softKB := strconv.FormatUint(diskMB*1024, 10)
	hardKB := strconv.FormatUint((diskMB+100)*1024, 10)
	softInodes := strconv.FormatUint(inodes, 10)
	hardInodes := strconv.FormatUint(inodes+10000, 10)

	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/sbin/setquota",
		"-u", username, softKB, hardKB, softInodes, hardInodes, "/",
	)
	return err
}

// ApplyCgroupsSlice generates systemd slice override for tenant UID.
func ApplyCgroupsSlice(ctx context.Context, uid uint32, limits TenantLimits) error {
	sliceDir := fmt.Sprintf("/etc/systemd/system/user-%d.slice.d", uid)
	confPath := sliceDir + "/50-aetherpanel-limits.conf"

	cpuQuota := 100
	if limits.CPUQuotaPercent > 0 && limits.CPUQuotaPercent <= 10000 {
		cpuQuota = int(limits.CPUQuotaPercent) // #nosec G115 - bounded range
	}
	memMB := 1024
	if limits.MemoryLimitMB > 0 && limits.MemoryLimitMB <= (1<<31-1) {
		memMB = int(limits.MemoryLimitMB) // #nosec G115 - bounded range
	}
	ioWeight := 100
	if limits.IOWeight > 0 && limits.IOWeight <= 10000 {
		ioWeight = int(limits.IOWeight) // #nosec G115 - bounded range
	}
	tasksMax := 256
	if limits.TasksMax > 0 && limits.TasksMax <= 1000000 {
		tasksMax = int(limits.TasksMax) // #nosec G115 - bounded range
	}

	content := fmt.Sprintf(`[Slice]
CPUQuota=%d%%
MemoryMax=%dM
IOWeight=%d
TasksMax=%d
`, cpuQuota, memMB, ioWeight, tasksMax)

	if err := WriteAtomic(confPath, content, 0644); err != nil {
		return err
	}

	return ReloadSystemd(ctx)
}

func isAlreadyExists(err error) bool {
	return err != nil && (contains(err.Error(), "already exists") || contains(err.Error(), "exit status 9"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && hasSubstr(s, substr)))
}

func hasSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
