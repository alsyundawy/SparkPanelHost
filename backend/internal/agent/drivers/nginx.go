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

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

type VHostConfig struct {
	Username      string
	DomainName    string
	DocumentRoot  string
	PHPVersion    string // "8.2", "8.3", "8.4", "8.5", "8.6", "static", "proxy"
	ProxyPort     uint32
	SSLEnabled    bool
	ForceHTTPS    bool
	WAFMode       string // "Blocking", "DetectionOnly", "Off"
	WAFProfile    string
	NginxTemplate string
}

func ValidateDomain(domain string) error {
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain name %q", domain)
	}
	return nil
}

// SyncVirtualHost renders hardened Nginx config, tests with nginx -t, and reloads.
func SyncVirtualHost(ctx context.Context, cfg VHostConfig) error {
	if err := ValidateDomain(cfg.DomainName); err != nil {
		return err
	}
	if err := ValidateUsername(cfg.Username); err != nil {
		return err
	}

	docRoot := filepath.Clean(cfg.DocumentRoot)
	if !strings.HasPrefix(docRoot, "/home/"+cfg.Username) && !strings.HasPrefix(docRoot, "/var/www") {
		return fmt.Errorf("security violation: document root %q must reside under /home/%s", docRoot, cfg.Username)
	}

	// Prepare directories
	_ = os.MkdirAll("/etc/nginx/sites-available", 0755)
	_ = os.MkdirAll("/etc/nginx/sites-enabled", 0755)
	_ = os.MkdirAll("/etc/nginx/coraza/exclusions", 0755)

	// Build Coraza WAF directives
	secRuleEngine := "Off"
	switch cfg.WAFMode {
	case "Blocking":
		secRuleEngine = "On"
	case "DetectionOnly":
		secRuleEngine = "DetectionOnly"
	}

	exclusionFile := fmt.Sprintf("/etc/nginx/coraza/exclusions/%s.conf", cfg.DomainName)
	if _, err := os.Stat(exclusionFile); os.IsNotExist(err) {
		_ = os.WriteFile(exclusionFile, []byte("# Coraza WAF Exclusions for "+cfg.DomainName+"\n"), 0644)
	}

	// SSL configuration paths
	sslCert := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", cfg.DomainName)
	sslKey := fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", cfg.DomainName)
	hasSSL := cfg.SSLEnabled
	if hasSSL {
		if _, err := os.Stat(sslCert); err != nil {
			hasSSL = false
		}
	}

	var sb strings.Builder

	// HTTP Server Block
	sb.WriteString(fmt.Sprintf(`# Managed by AetherPanel v8.1 - %s
server {
    listen 80;
    listen [::]:80;
    server_name %s www.%s;

    # ACME Challenge location
    location ^~ /.well-known/acme-challenge/ {
        root /var/panel/acme-challenge;
        allow all;
    }
`, cfg.DomainName, cfg.DomainName, cfg.DomainName))

	if hasSSL && cfg.ForceHTTPS {
		sb.WriteString(`
    location / {
        return 301 https://$host$request_uri;
    }
}
`)
	} else {
		writeServerBody(&sb, cfg, secRuleEngine, exclusionFile)
		sb.WriteString("}\n")
	}

	// HTTPS Server Block
	if hasSSL {
		sb.WriteString(fmt.Sprintf(`
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name %s www.%s;

    ssl_certificate %s;
    ssl_certificate_key %s;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:10m;
    ssl_session_tickets off;

    # HSTS & Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
`, cfg.DomainName, cfg.DomainName, sslCert, sslKey))

		writeServerBody(&sb, cfg, secRuleEngine, exclusionFile)
		sb.WriteString("}\n")
	}

	availablePath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", cfg.DomainName)
	enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", cfg.DomainName)

	// Backup existing file for rollback if test fails
	var oldContent []byte
	if data, err := os.ReadFile(availablePath); err == nil {
		oldContent = data
	}

	if err := WriteAtomic(availablePath, sb.String(), 0644); err != nil {
		return err
	}

	// Symlink to sites-enabled
	_ = os.Remove(enabledPath)
	if err := os.Symlink(availablePath, enabledPath); err != nil {
		return err
	}

	// Test configuration
	if _, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/sbin/nginx", "-t"); err != nil {
		// Rollback
		if len(oldContent) > 0 {
			_ = os.WriteFile(availablePath, oldContent, 0644) // #nosec G703,G306
		} else {
			_ = os.Remove(availablePath)
			_ = os.Remove(enabledPath)
		}
		return fmt.Errorf("nginx config test failed: %w", err)
	}

	// Reload nginx
	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", "nginx")
	return err
}

func writeServerBody(sb *strings.Builder, cfg VHostConfig, secRuleEngine, exclusionFile string) {
	docRoot := filepath.Clean(cfg.DocumentRoot)
	sb.WriteString(fmt.Sprintf(`
    root %s;
    index index.php index.html index.htm;

    # Staging temp buffer anti-Debian 13 tmpfs OOM
    client_body_temp_path /var/panel/tmp/%s/nginx 1 2;
    client_max_body_size 256M;

    # Anti-CVE Oct 2026: Safe URI handling
    merge_slashes on;
    disable_symlinks if_not_owner from=%s;

    # Coraza WAF Configuration
    coraza_enabled on;
    coraza_rules 'SecRuleEngine %s';
    include %s;

    access_log /var/log/nginx/%s_access.log;
    error_log /var/log/nginx/%s_error.log warn;

    location / {
`, docRoot, cfg.Username, docRoot, secRuleEngine, exclusionFile, cfg.DomainName, cfg.DomainName))

	if cfg.PHPVersion == "proxy" && cfg.ProxyPort > 0 {
		sb.WriteString(fmt.Sprintf(`        proxy_pass http://127.0.0.1:%d;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
`, cfg.ProxyPort))
	} else {
		sb.WriteString(`        try_files $uri $uri/ /index.php?$args;
    }
`)
	}

	// PHP-FPM Location
	if cfg.PHPVersion != "static" && cfg.PHPVersion != "proxy" {
		phpVer := cfg.PHPVersion
		if phpVer == "" {
			phpVer = "8.5"
		}
		sb.WriteString(fmt.Sprintf(`
    location ~ \.php$ {
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        try_files $fastcgi_script_name =404;
        fastcgi_pass unix:/run/php/php%s-fpm-%s.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
        fastcgi_read_timeout 300;
        fastcgi_buffer_size 128k;
        fastcgi_buffers 256 16k;
    }
`, phpVer, cfg.Username))
	}

	// Deny hidden files (.git, .env, .htaccess)
	sb.WriteString(`
    location ~ /\.(?!well-known).* {
        deny all;
        access_log off;
        log_not_found off;
    }
`)
}

// DeleteVirtualHost removes site configs and reloads Nginx.
func DeleteVirtualHost(ctx context.Context, username, domain string) error {
	availablePath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)
	enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", domain)
	exclusionPath := fmt.Sprintf("/etc/nginx/coraza/exclusions/%s.conf", domain)

	_ = os.Remove(enabledPath)
	_ = os.Remove(availablePath)
	_ = os.Remove(exclusionPath)

	if _, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/sbin/nginx", "-t"); err != nil {
		return err
	}
	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", "nginx")
	return err
}

// IssueSSLCertificate executes certbot in non-interactive mode.
func IssueSSLCertificate(ctx context.Context, domain string, challengeType string, wildcard bool) error {
	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--register-unsafely-without-email",
		"-d", domain,
	}

	if wildcard {
		args = append(args, "-d", "*."+domain)
	}

	if challengeType == "dns-01" {
		args = append(args, "--preferred-challenges", "dns", "--manual")
	} else {
		args = append(args, "--webroot", "-w", "/var/panel/acme-challenge")
	}

	_, err := execsafe.RunDirect(ctx, 120*time.Second, "/usr/bin/certbot", args...)
	return err
}

// ManageWAFExclusion adds or removes CRS rules for a domain.
func ManageWAFExclusion(ctx context.Context, domain, action string, ruleID uint32, uriPrefix string) error {
	exclusionPath := fmt.Sprintf("/etc/nginx/coraza/exclusions/%s.conf", domain)
	content, _ := os.ReadFile(exclusionPath)
	lines := strings.Split(string(content), "\n")

	ruleDirective := fmt.Sprintf("SecRuleRemoveById %d", ruleID)
	if uriPrefix != "" {
		ruleDirective = fmt.Sprintf("SecRule REQUEST_URI \"@beginsWith %s\" \"id:%d,phase:1,pass,nolog,ctl:ruleRemoveById=%d\"", uriPrefix, ruleID+1000000, ruleID)
	}

	var newLines []string
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, fmt.Sprintf("%d", ruleID)) {
			if action == "remove" {
				continue
			}
			found = true
		}
		if trimmed != "" {
			newLines = append(newLines, line)
		}
	}

	if action == "add" && !found {
		newLines = append(newLines, ruleDirective)
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	if err := WriteAtomic(exclusionPath, newContent, 0644); err != nil {
		return err
	}

	_, err := execsafe.RunDirect(ctx, 15*time.Second, "/usr/bin/systemctl", "reload", "nginx")
	return err
}
