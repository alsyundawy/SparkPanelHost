package execsafe

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// #nosec G101 -- allowlist of system binaries, not credentials
var allowedBinaries = map[string]string{
	"/usr/sbin/nginx":         "/usr/sbin/nginx",
	"/usr/bin/systemctl":      "/usr/bin/systemctl",
	"/usr/bin/systemd-run":    "/usr/bin/systemd-run",
	"/usr/sbin/useradd":       "/usr/sbin/useradd",
	"/usr/sbin/usermod":       "/usr/sbin/usermod",
	"/usr/sbin/userdel":       "/usr/sbin/userdel",
	"/usr/sbin/setquota":      "/usr/sbin/setquota",
	"/usr/sbin/chpasswd":      "/usr/sbin/chpasswd",
	"/usr/bin/chown":          "/usr/bin/chown",
	"/usr/bin/chmod":          "/usr/bin/chmod",
	"/usr/bin/rspamadm":       "/usr/bin/rspamadm",
	"/usr/bin/doveadm":        "/usr/bin/doveadm",
	"/usr/sbin/postmap":       "/usr/sbin/postmap",
	"/usr/bin/mariadb":        "/usr/bin/mariadb",
	"/usr/bin/mysql":          "/usr/bin/mysql",
	"/usr/bin/mariadb-dump":   "/usr/bin/mariadb-dump",
	"/usr/bin/mariadb-backup": "/usr/bin/mariadb-backup",
	"/usr/bin/psql":           "/usr/bin/psql",
	"/usr/bin/pg_dump":        "/usr/bin/pg_dump",
	"/usr/bin/createdb":       "/usr/bin/createdb",
	"/usr/bin/createuser":     "/usr/bin/createuser",
	"/usr/bin/dropdb":         "/usr/bin/dropdb",
	"/usr/bin/dropuser":       "/usr/bin/dropuser",
	"/usr/bin/certbot":        "/usr/bin/certbot",
	"/usr/bin/restic":         "/usr/bin/restic",
	"/usr/bin/rclone":         "/usr/bin/rclone",
	"/usr/sbin/nft":           "/usr/sbin/nft",
	"/usr/bin/cscli":          "/usr/bin/cscli",
	"/usr/bin/pdnsutil":       "/usr/bin/pdnsutil",
	"/usr/bin/crontab":        "/usr/bin/crontab",
	"/usr/bin/git":            "/usr/bin/git",
	"/usr/bin/php":            "/usr/bin/php",
	"/usr/bin/php8.2":         "/usr/bin/php8.2",
	"/usr/bin/php8.3":         "/usr/bin/php8.3",
	"/usr/bin/php8.4":         "/usr/bin/php8.4",
	"/usr/bin/php8.5":         "/usr/bin/php8.5",
	"/usr/bin/php8.6":         "/usr/bin/php8.6",
	"/usr/sbin/php-fpm8.2":    "/usr/sbin/php-fpm8.2",
	"/usr/sbin/php-fpm8.3":    "/usr/sbin/php-fpm8.3",
	"/usr/sbin/php-fpm8.4":    "/usr/sbin/php-fpm8.4",
	"/usr/sbin/php-fpm8.5":    "/usr/sbin/php-fpm8.5",
	"/usr/sbin/php-fpm8.6":    "/usr/sbin/php-fpm8.6",
	"/usr/sbin/mariadbd":      "/usr/sbin/mariadbd",
	"/usr/local/bin/nginx":    "/usr/local/bin/nginx",
	"/usr/local/bin/php":      "/usr/local/bin/php",
	"/usr/local/bin/mariadb":  "/usr/local/bin/mariadb",
}

var cleanEnv = []string{
	"PATH=/usr/sbin:/usr/bin:/sbin:/bin",
	"LC_ALL=C.UTF-8",
	"LANG=C.UTF-8",
}

func RunDirect(ctx context.Context, timeout time.Duration, binary string, args ...string) (string, error) {
	canonicalBin, ok := allowedBinaries[binary]
	if !ok {
		return "", fmt.Errorf("security violation: biner %q tidak terdaftar pada allowlist", binary)
	}

	cleanArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, "\x00\r\n") {
			return "", fmt.Errorf("security violation: argument contains invalid control characters")
		}
		cleanArgs[i] = arg
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, canonicalBin, cleanArgs...) // #nosec G204
	cmd.Env = cleanEnv

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("eksekusi %s gagal (%v): %s", canonicalBin, err, stderr.String())
	}
	return stdout.String(), nil
}
