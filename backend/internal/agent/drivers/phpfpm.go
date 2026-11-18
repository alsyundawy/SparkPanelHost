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

var supportedPHPVersions = map[string]bool{
	"8.2": true,
	"8.3": true,
	"8.4": true,
	"8.5": true,
	"8.6": true,
}

type PHPPoolConfig struct {
	Username    string
	PHPVersion  string
	MaxChildren uint32
	MemoryLimit string
	CustomINI   map[string]string
}

// SyncPHPPool writes an isolated PHP-FPM pool config for a tenant and reloads the service.
func SyncPHPPool(ctx context.Context, cfg PHPPoolConfig) error {
	if err := ValidateUsername(cfg.Username); err != nil {
		return err
	}

	phpVer := cfg.PHPVersion
	if phpVer == "" {
		phpVer = "8.5"
	}
	if !supportedPHPVersions[phpVer] {
		return fmt.Errorf("unsupported PHP version %q: supported are 8.2, 8.3, 8.4, 8.5, 8.6", phpVer)
	}

	maxChildren := cfg.MaxChildren
	if maxChildren == 0 {
		maxChildren = 10
	}
	memLimit := cfg.MemoryLimit
	if memLimit == "" {
		memLimit = "256M"
	}

	poolDir := fmt.Sprintf("/etc/php/%s/fpm/pool.d", phpVer)
	_ = os.MkdirAll(poolDir, 0755)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`; Managed by AetherPanel v8.1
[%s]
user = %s
group = %s

listen = /run/php/php%s-fpm-%s.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = ondemand
pm.max_children = %d
pm.process_idle_timeout = 10s
pm.max_requests = 1000

; Hardened Security Isolation
php_admin_value[memory_limit] = %s
php_admin_value[open_basedir] = /home/%s:/var/panel/tmp/%s:/tmp
php_admin_value[upload_tmp_dir] = /var/panel/tmp/%s
php_admin_value[session.save_path] = /var/panel/tmp/%s
php_admin_value[disable_functions] = exec,passthru,shell_exec,system,proc_open,popen,curl_multi_exec,parse_ini_file,show_source
php_admin_flag[expose_php] = Off
php_admin_flag[allow_url_fopen] = On
php_admin_flag[allow_url_include] = Off

`, cfg.Username, cfg.Username, cfg.Username, phpVer, cfg.Username, maxChildren, memLimit, cfg.Username, cfg.Username, cfg.Username, cfg.Username))

	for k, v := range cfg.CustomINI {
		cleanK := strings.ReplaceAll(k, "\n", "")
		cleanV := strings.ReplaceAll(v, "\n", "")
		sb.WriteString(fmt.Sprintf("php_admin_value[%s] = %s\n", cleanK, cleanV))
	}

	poolFile := filepath.Join(poolDir, fmt.Sprintf("%s.conf", cfg.Username))
	if err := WriteAtomic(poolFile, sb.String(), 0644); err != nil {
		return err
	}

	serviceName := fmt.Sprintf("php%s-fpm", phpVer)
	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", serviceName)
	return err
}

// RemovePHPPool removes pool config for tenant and reloads.
func RemovePHPPool(ctx context.Context, username, phpVer string) error {
	if !supportedPHPVersions[phpVer] {
		return fmt.Errorf("unsupported PHP version %q", phpVer)
	}
	poolFile := fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", phpVer, username)
	_ = os.Remove(poolFile)

	serviceName := fmt.Sprintf("php%s-fpm", phpVer)
	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", serviceName)
	return err
}
