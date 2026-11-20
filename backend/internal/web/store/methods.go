package store

import (
	"context"
	"database/sql"
)

// User Methods
func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, role, reseller_id, package_id, linux_uid, linux_gid, locale, theme, totp_secret_enc, totp_enabled, status, suspended_reason, created_at
		FROM users WHERE username = ?
	`, username)
	return scanUser(row)
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, role, reseller_id, package_id, linux_uid, linux_gid, locale, theme, totp_secret_enc, totp_enabled, status, suspended_reason, created_at
		FROM users WHERE id = ?
	`, id)
	return scanUser(row)
}

func (s *Store) ListUsers(ctx context.Context, resellerID *string) ([]User, error) {
	query := `SELECT id, username, email, password_hash, role, reseller_id, package_id, linux_uid, linux_gid, locale, theme, totp_secret_enc, totp_enabled, status, suspended_reason, created_at FROM users`
	var rows *sql.Rows
	var err error
	if resellerID != nil {
		query += ` WHERE reseller_id = ? ORDER BY created_at DESC`
		rows, err = s.db.QueryContext(ctx, query, *resellerID)
	} else {
		query += ` ORDER BY created_at DESC`
		rows, err = s.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (s *Store) CreateUser(ctx context.Context, u *User) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash, role, reseller_id, package_id, linux_uid, linux_gid, locale, theme, totp_secret_enc, totp_enabled, status, suspended_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, u.ID, u.Username, u.Email, u.PasswordHash, u.Role, u.ResellerID, u.PackageID, u.LinuxUID, u.LinuxGID, u.Locale, u.Theme, u.TOTPSecretEnc, u.TOTPEnabled, u.Status, u.SuspendedReason)
	return err
}

func (s *Store) UpdateUserStatus(ctx context.Context, id, status string, reason *string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET status = ?, suspended_reason = ? WHERE id = ?`, status, reason, id)
	return err
}

func (s *Store) GetNextLinuxUID(ctx context.Context) (int, error) {
	var maxUID sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT MAX(linux_uid) FROM users WHERE linux_uid >= 2000`).Scan(&maxUID)
	if err != nil {
		return 0, err
	}
	if !maxUID.Valid || maxUID.Int64 < 2000 {
		return 2000, nil
	}
	return int(maxUID.Int64 + 1), nil
}

func scanUser(scanner interface {
	Scan(dest ...interface{}) error
}) (*User, error) {
	var u User
	err := scanner.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.ResellerID, &u.PackageID,
		&u.LinuxUID, &u.LinuxGID, &u.Locale, &u.Theme, &u.TOTPSecretEnc, &u.TOTPEnabled,
		&u.Status, &u.SuspendedReason, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Session Methods
func (s *Store) CreateSession(ctx context.Context, sess *UserSession) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_sessions (id, user_id, impersonator_id, token_hash, ip_address, user_agent, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sess.ID, sess.UserID, sess.ImpersonatorID, sess.TokenHash, sess.IPAddress, sess.UserAgent, sess.ExpiresAt)
	return err
}

func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*UserSession, error) {
	var sess UserSession
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, impersonator_id, token_hash, ip_address, user_agent, expires_at, created_at
		FROM user_sessions WHERE token_hash = ? AND expires_at > CURRENT_TIMESTAMP
	`, tokenHash).Scan(
		&sess.ID, &sess.UserID, &sess.ImpersonatorID, &sess.TokenHash,
		&sess.IPAddress, &sess.UserAgent, &sess.ExpiresAt, &sess.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// Package Methods
func (s *Store) ListPackages(ctx context.Context) ([]Package, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, disk_quota_mb, inode_quota, bandwidth_quota_mb, max_domains, max_databases, max_mailboxes,
		       max_cron_jobs, max_ftp_accounts, cpu_quota_percent, memory_limit_mb, io_weight, tasks_max,
		       shell_access, allow_nodejs, allow_python, allow_git_deploy, allow_postgres, feature_list_id, created_at
		FROM packages ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pkgs []Package
	for rows.Next() {
		var p Package
		if err := rows.Scan(
			&p.ID, &p.Name, &p.DiskQuotaMB, &p.InodeQuota, &p.BandwidthQuotaMB, &p.MaxDomains,
			&p.MaxDatabases, &p.MaxMailboxes, &p.MaxCronJobs, &p.MaxFTPAccounts, &p.CPUQuotaPercent,
			&p.MemoryLimitMB, &p.IOWeight, &p.TasksMax, &p.ShellAccess, &p.AllowNodeJS,
			&p.AllowPython, &p.AllowGitDeploy, &p.AllowPostgres, &p.FeatureListID, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, rows.Err()
}

func (s *Store) CreatePackage(ctx context.Context, p *Package) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO packages (id, name, disk_quota_mb, inode_quota, bandwidth_quota_mb, max_domains, max_databases, max_mailboxes, max_cron_jobs, max_ftp_accounts, cpu_quota_percent, memory_limit_mb, io_weight, tasks_max, shell_access, allow_nodejs, allow_python, allow_git_deploy, allow_postgres, feature_list_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.ID, p.Name, p.DiskQuotaMB, p.InodeQuota, p.BandwidthQuotaMB, p.MaxDomains, p.MaxDatabases, p.MaxMailboxes, p.MaxCronJobs, p.MaxFTPAccounts, p.CPUQuotaPercent, p.MemoryLimitMB, p.IOWeight, p.TasksMax, p.ShellAccess, p.AllowNodeJS, p.AllowPython, p.AllowGitDeploy, p.AllowPostgres, p.FeatureListID)
	return err
}

// Domain Methods
func (s *Store) ListDomains(ctx context.Context, userID *string) ([]Domain, error) {
	query := `SELECT id, user_id, domain_name, document_root, php_version, ssl_enabled, ssl_provider, ssl_expires_at, force_https, waf_mode, waf_profile, nginx_template, created_at FROM domains`
	var rows *sql.Rows
	var err error
	if userID != nil {
		query += ` WHERE user_id = ? ORDER BY domain_name ASC`
		rows, err = s.db.QueryContext(ctx, query, *userID)
	} else {
		query += ` ORDER BY domain_name ASC`
		rows, err = s.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var d Domain
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.DomainName, &d.DocumentRoot, &d.PHPVersion, &d.SSLEnabled,
			&d.SSLProvider, &d.SSLExpiresAt, &d.ForceHTTPS, &d.WAFMode, &d.WAFProfile,
			&d.NginxTemplate, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (s *Store) CreateDomain(ctx context.Context, d *Domain) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO domains (id, user_id, domain_name, document_root, php_version, ssl_enabled, ssl_provider, ssl_expires_at, force_https, waf_mode, waf_profile, nginx_template)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.UserID, d.DomainName, d.DocumentRoot, d.PHPVersion, d.SSLEnabled, d.SSLProvider, d.SSLExpiresAt, d.ForceHTTPS, d.WAFMode, d.WAFProfile, d.NginxTemplate)
	return err
}

func (s *Store) GetDomainByID(ctx context.Context, id string) (*Domain, error) {
	var d Domain
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, domain_name, document_root, php_version, ssl_enabled, ssl_provider, ssl_expires_at, force_https, waf_mode, waf_profile, nginx_template, created_at
		FROM domains WHERE id = ?
	`, id).Scan(
		&d.ID, &d.UserID, &d.DomainName, &d.DocumentRoot, &d.PHPVersion, &d.SSLEnabled,
		&d.SSLProvider, &d.SSLExpiresAt, &d.ForceHTTPS, &d.WAFMode, &d.WAFProfile,
		&d.NginxTemplate, &d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) DeleteDomain(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM domains WHERE id = ?`, id)
	return err
}

// Database Methods
func (s *Store) ListDatabases(ctx context.Context, userID *string) ([]TenantDatabase, error) {
	query := `SELECT id, user_id, engine, db_name, db_user, charset, size_mb, created_at FROM tenant_databases`
	var rows *sql.Rows
	var err error
	if userID != nil {
		query += ` WHERE user_id = ? ORDER BY db_name ASC`
		rows, err = s.db.QueryContext(ctx, query, *userID)
	} else {
		query += ` ORDER BY db_name ASC`
		rows, err = s.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []TenantDatabase
	for rows.Next() {
		var db TenantDatabase
		if err := rows.Scan(&db.ID, &db.UserID, &db.Engine, &db.DBName, &db.DBUser, &db.Charset, &db.SizeMB, &db.CreatedAt); err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}
	return dbs, rows.Err()
}

func (s *Store) CreateTenantDatabase(ctx context.Context, db *TenantDatabase) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tenant_databases (id, user_id, engine, db_name, db_user, charset, size_mb)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, db.ID, db.UserID, db.Engine, db.DBName, db.DBUser, db.Charset, db.SizeMB)
	return err
}

func (s *Store) GetTenantDatabaseByID(ctx context.Context, id string) (*TenantDatabase, error) {
	var db TenantDatabase
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, engine, db_name, db_user, charset, size_mb, created_at
		FROM tenant_databases WHERE id = ?
	`, id).Scan(&db.ID, &db.UserID, &db.Engine, &db.DBName, &db.DBUser, &db.Charset, &db.SizeMB, &db.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (s *Store) DeleteTenantDatabase(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tenant_databases WHERE id = ?`, id)
	return err
}

// DNS Clustering Methods (WHM DNS Only style)
func (s *Store) ListDNSClusters(ctx context.Context) ([]DNSCluster, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, topology, primary_node_id, created_at FROM dns_clusters ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []DNSCluster
	for rows.Next() {
		var c DNSCluster
		if err := rows.Scan(&c.ID, &c.Name, &c.Topology, &c.PrimaryNodeID, &c.CreatedAt); err != nil {
			return nil, err
		}
		clusters = append(clusters, c)
	}
	return clusters, rows.Err()
}

func (s *Store) CreateDNSCluster(ctx context.Context, c *DNSCluster) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO dns_clusters (id, name, topology, primary_node_id)
		VALUES (?, ?, ?, ?)
	`, c.ID, c.Name, c.Topology, c.PrimaryNodeID)
	return err
}

func (s *Store) ListDNSClusterNodes(ctx context.Context, clusterID string) ([]DNSClusterNode, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, cluster_id, hostname, api_endpoint, role, api_token_id, is_active, last_seen_at, created_at
		FROM dns_cluster_nodes WHERE cluster_id = ? ORDER BY hostname ASC
	`, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []DNSClusterNode
	for rows.Next() {
		var n DNSClusterNode
		if err := rows.Scan(&n.ID, &n.ClusterID, &n.Hostname, &n.APIEndpoint, &n.Role, &n.APITokenID, &n.IsActive, &n.LastSeenAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (s *Store) CreateDNSClusterNode(ctx context.Context, n *DNSClusterNode) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO dns_cluster_nodes (id, cluster_id, hostname, api_endpoint, role, api_token_id, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, n.ID, n.ClusterID, n.Hostname, n.APIEndpoint, n.Role, n.APITokenID, n.IsActive)
	return err
}

// Mail Methods
func (s *Store) ListMailDomains(ctx context.Context) ([]MailDomain, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, domain_id, dkim_selector, dkim_rsa_public, dkim_ed25519_public, mta_sts_mode, catchall_address, outbound_limit_hour, created_at
		FROM mail_domains ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []MailDomain
	for rows.Next() {
		var md MailDomain
		if err := rows.Scan(&md.ID, &md.DomainID, &md.DKIMSelector, &md.DKIMRSAPublic, &md.DKIMEd25519Public, &md.MTASTSMode, &md.CatchallAddress, &md.OutboundLimitHour, &md.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, md)
	}
	return domains, rows.Err()
}

func (s *Store) CreateMailDomain(ctx context.Context, md *MailDomain) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO mail_domains (id, domain_id, dkim_selector, dkim_rsa_public, dkim_ed25519_public, mta_sts_mode, catchall_address, outbound_limit_hour)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, md.ID, md.DomainID, md.DKIMSelector, md.DKIMRSAPublic, md.DKIMEd25519Public, md.MTASTSMode, md.CatchallAddress, md.OutboundLimitHour)
	return err
}

func (s *Store) ListMailboxes(ctx context.Context, mailDomainID string) ([]Mailbox, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, mail_domain_id, local_part, quota_mb, active, created_at
		FROM mailboxes WHERE mail_domain_id = ? ORDER BY local_part ASC
	`, mailDomainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boxes []Mailbox
	for rows.Next() {
		var b Mailbox
		if err := rows.Scan(&b.ID, &b.MailDomainID, &b.LocalPart, &b.QuotaMB, &b.Active, &b.CreatedAt); err != nil {
			return nil, err
		}
		boxes = append(boxes, b)
	}
	return boxes, rows.Err()
}

func (s *Store) CreateMailbox(ctx context.Context, mb *Mailbox) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO mailboxes (id, mail_domain_id, local_part, password_hash, quota_mb, active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, mb.ID, mb.MailDomainID, mb.LocalPart, mb.PasswordHash, mb.QuotaMB, mb.Active)
	return err
}

func (s *Store) ListAuditLogs(ctx context.Context, limit int) ([]AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, actor_user_id, impersonator_id, action, target_resource, source_ip, prev_hash, entry_hash, created_at
		FROM audit_logs ORDER BY id DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.ActorUserID, &l.ImpersonatorID, &l.Action, &l.TargetResource, &l.SourceIP, &l.PrevHash, &l.EntryHash, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// TOTP & Passkeys Methods
func (s *Store) UpdateUserTOTP(ctx context.Context, userID, secretEnc string, enabled bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET totp_secret_enc = ?, totp_enabled = ? WHERE id = ?
	`, secretEnc, enabled, userID)
	return err
}

func (s *Store) CreateUserPasskey(ctx context.Context, p *UserPasskey) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_passkeys (id, user_id, credential_id, public_key, sign_count, device_name)
		VALUES (?, ?, ?, ?, ?, ?)
	`, p.ID, p.UserID, p.CredentialID, p.PublicKey, p.SignCount, p.DeviceName)
	return err
}

func (s *Store) GetUserPasskeys(ctx context.Context, userID string) ([]UserPasskey, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, credential_id, public_key, sign_count, device_name, created_at
		FROM user_passkeys WHERE user_id = ? ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passkeys []UserPasskey
	for rows.Next() {
		var p UserPasskey
		if err := rows.Scan(&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey, &p.SignCount, &p.DeviceName, &p.CreatedAt); err != nil {
			return nil, err
		}
		passkeys = append(passkeys, p)
	}
	return passkeys, rows.Err()
}

func (s *Store) GetPasskeyByCredentialID(ctx context.Context, credID []byte) (*UserPasskey, error) {
	var p UserPasskey
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, credential_id, public_key, sign_count, device_name, created_at
		FROM user_passkeys WHERE credential_id = ?
	`, credID).Scan(&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey, &p.SignCount, &p.DeviceName, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) UpdatePasskeySignCount(ctx context.Context, credID []byte, count int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_passkeys SET sign_count = ? WHERE credential_id = ?
	`, count, credID)
	return err
}

func (s *Store) DeleteUserPasskey(ctx context.Context, id, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM user_passkeys WHERE id = ? AND user_id = ?
	`, id, userID)
	return err
}

// API Token Methods
func (s *Store) GetAPITokenByHash(ctx context.Context, tokenHash string) (*APIToken, error) {
	var t APIToken
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, token_prefix, token_hash, scopes_json, allowed_ips_json, last_used_at, expires_at, created_at
		FROM api_tokens WHERE token_hash = ?
	`, tokenHash).Scan(&t.ID, &t.UserID, &t.Name, &t.TokenPrefix, &t.TokenHash, &t.ScopesJSON, &t.AllowedIPsJSON, &t.LastUsedAt, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) UpdateAPITokenLastUsed(ctx context.Context, tokenID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE api_tokens SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?
	`, tokenID)
	return err
}
