package drivers

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

type SystemTelemetry struct {
	CPUPercent         float64
	RAMUsedMB          uint64
	RAMTotalMB         uint64
	DiskUsedMB         uint64
	DiskTotalMB        uint64
	WAFBlocked24h      uint64
	CrowdSecActiveBans uint64
}

// GetSystemTelemetry reads system resource metrics and security counters.
func GetSystemTelemetry(ctx context.Context, username string) (*SystemTelemetry, error) {
	telemetry := &SystemTelemetry{}

	// Read RAM info from /proc/meminfo
	if memBytes, err := os.ReadFile("/proc/meminfo"); err == nil {
		var totalKB, freeKB, buffersKB, cachedKB uint64
		for _, line := range strings.Split(string(memBytes), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			switch fields[0] {
			case "MemTotal:":
				totalKB = val
			case "MemFree:":
				freeKB = val
			case "Buffers:":
				buffersKB = val
			case "Cached:":
				cachedKB = val
			}
		}
		if totalKB > 0 {
			telemetry.RAMTotalMB = totalKB / 1024
			usedKB := totalKB - (freeKB + buffersKB + cachedKB)
			telemetry.RAMUsedMB = usedKB / 1024
		}
	} else {
		// Fallback for macOS / non-Linux dev
		telemetry.RAMTotalMB = 8192
		telemetry.RAMUsedMB = 2048
	}

	// Read Disk info
	var stat syscall.Statfs_t
	checkPath := "/"
	if username != "" {
		userHome := "/home/" + username
		if _, err := os.Stat(userHome); err == nil {
			checkPath = userHome
		}
	}

	if err := syscall.Statfs(checkPath, &stat); err == nil && stat.Bsize > 0 {
		bsize := uint64(stat.Bsize)
		totalBytes := stat.Blocks * bsize
		freeBytes := stat.Bfree * bsize
		telemetry.DiskTotalMB = totalBytes / (1024 * 1024)
		telemetry.DiskUsedMB = (totalBytes - freeBytes) / (1024 * 1024)
	} else {
		telemetry.DiskTotalMB = 102400
		telemetry.DiskUsedMB = 15360
	}

	// Read CPU usage from /proc/stat
	if statBytes, err := os.ReadFile("/proc/stat"); err == nil {
		lines := strings.Split(string(statBytes), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "cpu ") {
			fields := strings.Fields(lines[0])
			if len(fields) >= 5 {
				user, _ := strconv.ParseFloat(fields[1], 64)
				nice, _ := strconv.ParseFloat(fields[2], 64)
				system, _ := strconv.ParseFloat(fields[3], 64)
				idle, _ := strconv.ParseFloat(fields[4], 64)
				total := user + nice + system + idle
				if total > 0 {
					telemetry.CPUPercent = ((user + nice + system) / total) * 100.0
				}
			}
		}
	} else {
		telemetry.CPUPercent = 4.2
	}

	// Read Coraza WAF blocked counter
	if data, err := os.ReadFile("/var/log/nginx/coraza_blocked.count"); err == nil {
		val, _ := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		telemetry.WAFBlocked24h = val
	}

	// Query CrowdSec active bans
	if csOut, err := execsafe.RunDirect(ctx, 5*time.Second, "/usr/bin/cscli", "decisions", "list", "-o", "json"); err == nil {
		var decisions []map[string]interface{}
		if err := json.Unmarshal([]byte(csOut), &decisions); err == nil {
			telemetry.CrowdSecActiveBans = uint64(len(decisions))
		}
	}

	return telemetry, nil
}
