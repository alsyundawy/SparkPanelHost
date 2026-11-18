package execsafe

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

var allowedBinaries = map[string]bool{
	"/usr/sbin/nginx":         true,
	"/usr/bin/systemctl":      true,
	"/usr/bin/systemd-run":    true,
	"/usr/sbin/useradd":       true,
	"/usr/sbin/usermod":       true,
	"/usr/sbin/userdel":       true,
	"/usr/sbin/setquota":      true,
	"/usr/sbin/chpasswd":      true,
	"/usr/bin/chown":          true,
	"/usr/bin/chmod":          true,
	"/usr/bin/rspamadm":       true,
	"/usr/bin/doveadm":        true,
	"/usr/sbin/postmap":       true,
	"/usr/bin/mariadb":        true,
	"/usr/bin/mysql":          true,
	"/usr/bin/mariadb-dump":   true,
	"/usr/bin/mariadb-backup": true,
	"/usr/bin/psql":           true,
	"/usr/bin/pg_dump":        true,
	"/usr/bin/createdb":       true,
	"/usr/bin/createuser":     true,
	"/usr/bin/dropdb":         true,
	"/usr/bin/dropuser":       true,
	"/usr/bin/certbot":        true,
	"/usr/bin/restic":         true,
	"/usr/bin/rclone":         true,
	"/usr/sbin/nft":           true,
	"/usr/bin/cscli":          true,
	"/usr/bin/pdnsutil":       true,
	"/usr/bin/crontab":        true,
	"/usr/bin/git":            true,
	"/usr/bin/php":            true,
	"/usr/bin/php8.2":         true,
	"/usr/bin/php8.3":         true,
	"/usr/bin/php8.4":         true,
	"/usr/bin/php8.5":         true,
	"/usr/bin/php8.6":         true,
	"/usr/sbin/php-fpm8.2":    true,
	"/usr/sbin/php-fpm8.3":    true,
	"/usr/sbin/php-fpm8.4":    true,
	"/usr/sbin/php-fpm8.5":    true,
	"/usr/sbin/php-fpm8.6":    true,
	"/usr/sbin/mariadbd":      true,
	"/usr/local/bin/nginx":    true,
	"/usr/local/bin/php":      true,
	"/usr/local/bin/mariadb":  true,
}

var cleanEnv = []string{
	"PATH=/usr/sbin:/usr/bin:/sbin:/bin",
	"LC_ALL=C.UTF-8",
	"LANG=C.UTF-8",
}

func RunDirect(ctx context.Context, timeout time.Duration, binary string, args ...string) (string, error) {
	if !allowedBinaries[binary] {
		return "", fmt.Errorf("security violation: biner %q tidak terdaftar pada allowlist", binary)
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, binary, args...)
	cmd.Env = cleanEnv

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("eksekusi %s gagal (%v): %s", binary, err, stderr.String())
	}
	return stdout.String(), nil
}
