package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db       *sql.DB
	auditMut sync.Mutex
}

// NewStore opens SQLite with WAL mode, foreign keys, and executes migrations.
func NewStore(dbPath, migrationPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	// Execute migrations
	if migrationPath != "" {
		sqlContent, err := os.ReadFile(migrationPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migrations file %s: %w", migrationPath, err)
		}

		queries := strings.Split(string(sqlContent), ";")
		for _, q := range queries {
			// Strip SQL comments from query
			var cleanLines []string
			for _, line := range strings.Split(q, "\n") {
				trimLine := strings.TrimSpace(line)
				if !strings.HasPrefix(trimLine, "--") && trimLine != "" {
					cleanLines = append(cleanLines, line)
				}
			}
			trimmed := strings.TrimSpace(strings.Join(cleanLines, "\n"))
			if trimmed == "" {
				continue
			}
			if _, err := db.Exec(trimmed); err != nil {
				// Ignore table already exists
				if !strings.Contains(err.Error(), "already exists") {
					return nil, fmt.Errorf("migration query error (%v): %s", err, trimmed)
				}
			}
		}
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

// Models
type User struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	Role            string    `json:"role"` // 'superadmin', 'reseller', 'tenant'
	ResellerID      *string   `json:"reseller_id,omitempty"`
	PackageID       *string   `json:"package_id,omitempty"`
	LinuxUID        int       `json:"linux_uid"`
	LinuxGID        int       `json:"linux_gid"`
	Locale          string    `json:"locale"`
	Theme           string    `json:"theme"`
	TOTPSecretEnc   *string   `json:"-"`
	TOTPEnabled     bool      `json:"totp_enabled"`
	Status          string    `json:"status"` // 'active', 'suspended'
	SuspendedReason *string   `json:"suspended_reason,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type Package struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	DiskQuotaMB      int64     `json:"disk_quota_mb"`
	InodeQuota       int64     `json:"inode_quota"`
	BandwidthQuotaMB int64     `json:"bandwidth_quota_mb"`
	MaxDomains       int       `json:"max_domains"`
	MaxDatabases     int       `json:"max_databases"`
	MaxMailboxes     int       `json:"max_mailboxes"`
	MaxCronJobs      int       `json:"max_cron_jobs"`
	MaxFTPAccounts   int       `json:"max_ftp_accounts"`
	CPUQuotaPercent  int       `json:"cpu_quota_percent"`
	MemoryLimitMB    int64     `json:"memory_limit_mb"`
	IOWeight         int       `json:"io_weight"`
	TasksMax         int       `json:"tasks_max"`
	ShellAccess      bool      `json:"shell_access"`
	AllowNodeJS      bool      `json:"allow_nodejs"`
	AllowPython      bool      `json:"allow_python"`
	AllowGitDeploy   bool      `json:"allow_git_deploy"`
	AllowPostgres    bool      `json:"allow_postgres"`
	FeatureListID    *string   `json:"feature_list_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type FeatureList struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	FeaturesJSON string    `json:"features_json"`
	CreatedAt    time.Time `json:"created_at"`
}

type ResellerLimits struct {
	UserID           string    `json:"user_id"`
	MaxAccounts      int       `json:"max_accounts"`
	TotalDiskMB      int64     `json:"total_disk_mb"`
	TotalBandwidthMB int64     `json:"total_bandwidth_mb"`
	AllowOverselling bool      `json:"allow_overselling"`
	PrivateNS1       *string   `json:"private_ns1,omitempty"`
	PrivateNS2       *string   `json:"private_ns2,omitempty"`
	BrandName        *string   `json:"brand_name,omitempty"`
	BrandLogoURL     *string   `json:"brand_logo_url,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserSession struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ImpersonatorID *string   `json:"impersonator_id,omitempty"`
	TokenHash      string    `json:"-"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type UserPasskey struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	CredentialID []byte    `json:"credential_id"`
	PublicKey    []byte    `json:"public_key"`
	SignCount    int       `json:"sign_count"`
	DeviceName   string    `json:"device_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type APIToken struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Name           string     `json:"name"`
	TokenPrefix    string     `json:"token_prefix"`
	TokenHash      string     `json:"-"`
	ScopesJSON     string     `json:"scopes_json"`
	AllowedIPsJSON *string    `json:"allowed_ips_json,omitempty"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Domain struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	DomainName    string     `json:"domain_name"`
	DocumentRoot  string     `json:"document_root"`
	PHPVersion    string     `json:"php_version"`
	SSLEnabled    bool       `json:"ssl_enabled"`
	SSLProvider   *string    `json:"ssl_provider,omitempty"`
	SSLExpiresAt  *time.Time `json:"ssl_expires_at,omitempty"`
	ForceHTTPS    bool       `json:"force_https"`
	WAFMode       string     `json:"waf_mode"`
	WAFProfile    string     `json:"waf_profile"`
	NginxTemplate string     `json:"nginx_template"`
	CreatedAt     time.Time  `json:"created_at"`
}

type TenantDatabase struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Engine    string    `json:"engine"` // 'mariadb' or 'postgresql'
	DBName    string    `json:"db_name"`
	DBUser    string    `json:"db_user"`
	Charset   string    `json:"charset"`
	SizeMB    float64   `json:"size_mb"`
	CreatedAt time.Time `json:"created_at"`
}

type DNSZone struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ZoneName      string    `json:"zone_name"`
	DNSSECEnabled bool      `json:"dnssec_enabled"`
	Serial        int       `json:"serial"`
	CreatedAt     time.Time `json:"created_at"`
}

type DNSCluster struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Topology      string    `json:"topology"` // 'star', 'primary_secondary', 'galera', 'anycast'
	PrimaryNodeID *string   `json:"primary_node_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type DNSClusterNode struct {
	ID          string     `json:"id"`
	ClusterID   string     `json:"cluster_id"`
	Hostname    string     `json:"hostname"`
	APIEndpoint string     `json:"api_endpoint"`
	Role        string     `json:"role"` // 'write_only', 'synchronize', 'standalone', 'dns_only'
	APITokenID  *string    `json:"api_token_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DNSZoneSyncStatus struct {
	ID         string     `json:"id"`
	ZoneName   string     `json:"zone_name"`
	ClusterID  string     `json:"cluster_id"`
	NodeID     string     `json:"node_id"`
	Serial     int        `json:"serial"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	LastError  *string    `json:"last_error,omitempty"`
	Status     string     `json:"status"` // 'pending', 'synced', 'error', 'drift'
}

type MailDomain struct {
	ID                string    `json:"id"`
	DomainID          string    `json:"domain_id"`
	DKIMSelector      string    `json:"dkim_selector"`
	DKIMRSAPublic     string    `json:"dkim_rsa_public"`
	DKIMEd25519Public *string   `json:"dkim_ed25519_public,omitempty"`
	MTASTSMode        string    `json:"mta_sts_mode"`
	CatchallAddress   *string   `json:"catchall_address,omitempty"`
	OutboundLimitHour int       `json:"outbound_limit_hour"`
	CreatedAt         time.Time `json:"created_at"`
}

type Mailbox struct {
	ID           string    `json:"id"`
	MailDomainID string    `json:"mail_domain_id"`
	LocalPart    string    `json:"local_part"`
	PasswordHash string    `json:"-"`
	QuotaMB      int64     `json:"quota_mb"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type MailAlias struct {
	ID                string    `json:"id"`
	MailDomainID      string    `json:"mail_domain_id"`
	SourceLocalPart   string    `json:"source_local_part"`
	DestinationEmails string    `json:"destination_emails"` // JSON array
	Active            bool      `json:"active"`
	CreatedAt         time.Time `json:"created_at"`
}

type AuditLog struct {
	ID             int64     `json:"id"`
	ActorUserID    string    `json:"actor_user_id"`
	ImpersonatorID *string   `json:"impersonator_id,omitempty"`
	Action         string    `json:"action"`
	TargetResource string    `json:"target_resource"`
	SourceIP       string    `json:"source_ip"`
	PrevHash       string    `json:"prev_hash"`
	EntryHash      string    `json:"entry_hash"`
	CreatedAt      time.Time `json:"created_at"`
}

// AppendAuditLog computes immutable hash chain entry and saves it to audit_logs table.
func (s *Store) AppendAuditLog(ctx context.Context, actorID string, impersonatorID *string, action, targetResource, sourceIP string) error {
	s.auditMut.Lock()
	defer s.auditMut.Unlock()

	var prevHash string
	err := s.db.QueryRowContext(ctx, "SELECT entry_hash FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&prevHash)
	if err == sql.ErrNoRows {
		prevHash = strings.Repeat("0", 64) // Genesis hash
	} else if err != nil {
		return err
	}

	impStr := ""
	if impersonatorID != nil {
		impStr = *impersonatorID
	}
	now := time.Now().UTC()
	timestampStr := now.Format(time.RFC3339Nano)

	// Hash formula: hex(sha256(prev_hash + actor + impersonator + action + target + ip + timestamp))
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", prevHash, actorID, impStr, action, targetResource, sourceIP, timestampStr)
	h := sha256.Sum256([]byte(raw))
	entryHash := hex.EncodeToString(h[:])

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO audit_logs (actor_user_id, impersonator_id, action, target_resource, source_ip, prev_hash, entry_hash, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, actorID, impersonatorID, action, targetResource, sourceIP, prevHash, entryHash, now)
	return err
}
