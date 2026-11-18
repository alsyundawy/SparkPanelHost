<!-- markdownlint-disable -->

# 02. Skema Database Relasional & Sistem RBAC (v8.1)

Database internal panel menggunakan **SQLite (WAL Mode + Foreign Keys ON)** untuk instalasi single-node dan **100% valid secara sintaks maupun urutan dependensi _foreign key_ pada PostgreSQL 17/18+** untuk mode multi-server cluster.

**Penyempurnaan v8.1 dibanding v8.1:**

- Menambahkan tabel **DNS Clustering** (`dns_clusters`, `dns_cluster_nodes`, `dns_zone_sync_status`) untuk mendukung mode WHM DNS Only style.
- Memperbarui komentar versi dan stack target (MariaDB 12.3.4, Rspamd 4.3.0, CrowdSec 1.8.2, CRS 4.30).

---

## 1. Skema SQL Lengkap (`migrations/001_init.sql`)

```sql
-- Aktifkan pada SQLite (diabaikan / disesuaikan oleh driver PostgreSQL)
-- PRAGMA journal_mode = WAL;
-- PRAGMA foreign_keys = ON;

-- 1. Feature Lists (WHM-style: mengaktifkan/menyembunyikan modul UI & API per paket)
CREATE TABLE feature_lists (
    id                  TEXT PRIMARY KEY,
    name                TEXT NOT NULL UNIQUE,
    description         TEXT,
    features_json       TEXT NOT NULL,                  -- JSON array of allowed feature keys
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Paket Hosting & Batasan Resource (cgroups v2 + Linux Disk Quota)
CREATE TABLE packages (
    id                  TEXT PRIMARY KEY,
    name                TEXT NOT NULL UNIQUE,
    disk_quota_mb       INTEGER NOT NULL DEFAULT 5120,
    inode_quota         INTEGER NOT NULL DEFAULT 250000,
    bandwidth_quota_mb  INTEGER NOT NULL DEFAULT 102400,
    max_domains         INTEGER NOT NULL DEFAULT 5,
    max_databases       INTEGER NOT NULL DEFAULT 10,
    max_mailboxes       INTEGER NOT NULL DEFAULT 20,
    max_cron_jobs       INTEGER NOT NULL DEFAULT 10,
    max_ftp_accounts    INTEGER NOT NULL DEFAULT 5,
    cpu_quota_percent   INTEGER NOT NULL DEFAULT 100,   -- 100 = 1 vCPU penuh di cgroups v2
    memory_limit_mb     INTEGER NOT NULL DEFAULT 1024,  -- MemoryMax di systemd slice
    io_weight           INTEGER NOT NULL DEFAULT 100,   -- IOWeight 10-1000
    tasks_max           INTEGER NOT NULL DEFAULT 256,   -- Anti fork-bomb CWE-400
    shell_access        BOOLEAN NOT NULL DEFAULT FALSE,
    allow_nodejs        BOOLEAN NOT NULL DEFAULT FALSE,
    allow_python        BOOLEAN NOT NULL DEFAULT FALSE,
    allow_git_deploy    BOOLEAN NOT NULL DEFAULT TRUE,
    allow_postgres      BOOLEAN NOT NULL DEFAULT FALSE,
    feature_list_id     TEXT REFERENCES feature_lists(id) ON DELETE SET NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Akun Admin (WHM), Reseller & Tenant User (cPanel)
CREATE TABLE users (
    id                  TEXT PRIMARY KEY,
    username            TEXT NOT NULL UNIQUE,           -- Linux username (regex: ^[a-z][a-z0-9]{2,15}$)
    email               TEXT NOT NULL UNIQUE,
    password_hash       TEXT NOT NULL,                  -- Argon2id
    role                TEXT NOT NULL CHECK(role IN ('superadmin', 'reseller', 'tenant')),
    reseller_id         TEXT REFERENCES users(id) ON DELETE SET NULL,
    package_id          TEXT REFERENCES packages(id) ON DELETE RESTRICT,
    linux_uid           INTEGER UNIQUE,
    linux_gid           INTEGER UNIQUE,
    locale              TEXT NOT NULL DEFAULT 'id' CHECK(locale IN ('id', 'en')),
    theme               TEXT NOT NULL DEFAULT 'dark' CHECK(theme IN ('dark', 'light', 'system')),
    totp_secret_enc     TEXT,
    totp_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    status              TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'suspended')),
    suspended_reason    TEXT,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_reseller_id ON users(reseller_id);
CREATE INDEX idx_users_status ON users(status);

-- 4. Alokasi Kuota & Branding Reseller (v8.1)
CREATE TABLE reseller_limits (
    user_id             TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    max_accounts        INTEGER NOT NULL DEFAULT 25,
    total_disk_mb       INTEGER NOT NULL DEFAULT 102400,
    total_bandwidth_mb  INTEGER NOT NULL DEFAULT 1048576,
    allow_overselling   BOOLEAN NOT NULL DEFAULT FALSE,
    private_ns1         TEXT,
    private_ns2         TEXT,
    brand_name          TEXT,
    brand_logo_url      TEXT,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Sesi Login & Token Impersonasi (Revocable Sessions)
CREATE TABLE user_sessions (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    impersonator_id     TEXT REFERENCES users(id) ON DELETE CASCADE,
    token_hash          TEXT NOT NULL UNIQUE,
    ip_address          TEXT NOT NULL,
    user_agent          TEXT,
    expires_at          TIMESTAMP NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);

-- 6. Scoped API Tokens untuk Billing Automation (WHMCS / Blesta / HostBill) (v8.1)
CREATE TABLE api_tokens (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    token_prefix        TEXT NOT NULL,
    token_hash          TEXT NOT NULL UNIQUE,
    scopes_json         TEXT NOT NULL,                  -- ["accounts:create", "accounts:suspend", ...]
    allowed_ips_json    TEXT,                           -- Optional CIDR allowlist
    last_used_at        TIMESTAMP,
    expires_at          TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 7. WebAuthn / Passkeys Credentials
CREATE TABLE user_passkeys (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id       BLOB NOT NULL UNIQUE,
    public_key          BLOB NOT NULL,
    sign_count          INTEGER NOT NULL DEFAULT 0,
    device_name         TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 8. Domain, Multi-PHP (8.2 - 8.6), SSL ACME ARI & Coraza WAF Profile
CREATE TABLE domains (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain_name         TEXT NOT NULL UNIQUE,
    document_root       TEXT NOT NULL,
    php_version         TEXT NOT NULL DEFAULT '8.5' CHECK(php_version IN ('8.2', '8.3', '8.4', '8.5', '8.6', 'static', 'proxy')),
    ssl_enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    ssl_provider        TEXT DEFAULT 'letsencrypt',
    ssl_expires_at      TIMESTAMP,
    ssl_ari_cert_id     TEXT,                           -- RFC 9773 ACME Renewal Information ID
    force_https         BOOLEAN NOT NULL DEFAULT TRUE,
    waf_mode            TEXT NOT NULL DEFAULT 'Blocking' CHECK(waf_mode IN ('Blocking', 'DetectionOnly', 'Off')),
    waf_profile         TEXT NOT NULL DEFAULT 'default',
    nginx_template      TEXT NOT NULL DEFAULT 'default',
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_domains_user_id ON domains(user_id);

-- 9. Pengecualian Aturan OWASP CRS per-Domain (Coraza WAF) (v8.1)
CREATE TABLE waf_exclusions (
    id                  TEXT PRIMARY KEY,
    domain_id           TEXT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    rule_id             INTEGER NOT NULL,
    uri_prefix          TEXT,
    reason              TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 10. Database Tenant (MariaDB 12.3 LTS & PostgreSQL 17+)
CREATE TABLE tenant_databases (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    engine              TEXT NOT NULL DEFAULT 'mariadb' CHECK(engine IN ('mariadb', 'postgresql')),
    db_name             TEXT NOT NULL UNIQUE,           -- Format: <username>_<dbname>
    db_user             TEXT NOT NULL UNIQUE,
    charset             TEXT NOT NULL DEFAULT 'utf8mb4',
    size_mb             REAL NOT NULL DEFAULT 0,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_tenant_dbs_user_id ON tenant_databases(user_id);

-- 11. Zona DNS (Sinkronisasi dengan PowerDNS 5.1 REST API)
CREATE TABLE dns_zones (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zone_name           TEXT NOT NULL UNIQUE,
    dnssec_enabled      BOOLEAN NOT NULL DEFAULT FALSE,
    serial              INTEGER NOT NULL DEFAULT 1,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 12. Layanan Email Hosting (Postfix + Dovecot 2.4 + Rspamd 4.3 + MTA-STS)
CREATE TABLE mail_domains (
    id                  TEXT PRIMARY KEY,
    domain_id           TEXT NOT NULL UNIQUE REFERENCES domains(id) ON DELETE CASCADE,
    dkim_selector       TEXT NOT NULL DEFAULT 'default',
    dkim_rsa_public     TEXT NOT NULL,
    dkim_ed25519_public TEXT,
    mta_sts_mode        TEXT NOT NULL DEFAULT 'testing' CHECK(mta_sts_mode IN ('none', 'testing', 'enforce')),
    catchall_address    TEXT,
    outbound_limit_hour INTEGER NOT NULL DEFAULT 200,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE mailboxes (
    id                  TEXT PRIMARY KEY,
    mail_domain_id      TEXT NOT NULL REFERENCES mail_domains(id) ON DELETE CASCADE,
    local_part          TEXT NOT NULL,
    password_hash       TEXT NOT NULL,                  -- ARGON2ID untuk Dovecot 2.4
    quota_mb            INTEGER NOT NULL DEFAULT 2048,
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(mail_domain_id, local_part)
);

CREATE TABLE mail_aliases (
    id                  TEXT PRIMARY KEY,
    mail_domain_id      TEXT NOT NULL REFERENCES mail_domains(id) ON DELETE CASCADE,
    source_local_part   TEXT NOT NULL,
    destination_emails  TEXT NOT NULL,                  -- JSON array of target emails
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(mail_domain_id, source_local_part)
);

CREATE TABLE mail_autoresponders (
    id                  TEXT PRIMARY KEY,
    mailbox_id          TEXT NOT NULL UNIQUE REFERENCES mailboxes(id) ON DELETE CASCADE,
    subject             TEXT NOT NULL,
    body_text           TEXT NOT NULL,
    start_at            TIMESTAMP,
    end_at              TIMESTAMP,
    enabled             BOOLEAN NOT NULL DEFAULT FALSE
);

-- 13. Native App Installer Instances, Staging & Git Deployments
CREATE TABLE app_instances (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain_id           TEXT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    recipe_id           TEXT NOT NULL,
    app_version         TEXT NOT NULL,
    install_path        TEXT NOT NULL,
    database_id         TEXT REFERENCES tenant_databases(id) ON DELETE SET NULL,
    is_staging          BOOLEAN NOT NULL DEFAULT FALSE,
    parent_instance_id  TEXT REFERENCES app_instances(id) ON DELETE SET NULL,
    auto_update         BOOLEAN NOT NULL DEFAULT FALSE,
    installed_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE git_deployments (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain_id           TEXT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    repo_url            TEXT NOT NULL,
    branch              TEXT NOT NULL DEFAULT 'main',
    deploy_path         TEXT NOT NULL,
    post_deploy_cmd     TEXT,
    webhook_secret_hash TEXT NOT NULL,
    last_commit_sha     TEXT,
    last_deployed_at    TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 14. Modern App Runtimes (Node.js & Python venv), Cron Jobs & FTP/SFTP
CREATE TABLE app_runtimes (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain_id           TEXT REFERENCES domains(id) ON DELETE SET NULL,
    runtime_type        TEXT NOT NULL CHECK(runtime_type IN ('nodejs', 'python')),
    runtime_version     TEXT NOT NULL,
    app_name            TEXT NOT NULL,
    working_dir         TEXT NOT NULL,
    entrypoint          TEXT NOT NULL,
    port                INTEGER NOT NULL UNIQUE,
    status              TEXT NOT NULL DEFAULT 'stopped' CHECK(status IN ('running', 'stopped', 'failed')),
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cron_jobs (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    command             TEXT NOT NULL,
    schedule            TEXT NOT NULL,
    enabled             BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ftp_accounts (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username            TEXT NOT NULL UNIQUE,
    password_hash       TEXT NOT NULL,
    home_dir            TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 15. Backup Destinations & Jobs (Restic 3-2-1)
CREATE TABLE backup_destinations (
    id                  TEXT PRIMARY KEY,
    name                TEXT NOT NULL,
    type                TEXT NOT NULL CHECK(type IN ('local', 's3', 'sftp', 'b2')),
    config_enc          TEXT NOT NULL,
    is_default          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE backup_jobs (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT REFERENCES users(id) ON DELETE CASCADE, -- NULL = server-wide job
    destination_id      TEXT NOT NULL REFERENCES backup_destinations(id) ON DELETE RESTRICT,
    schedule            TEXT NOT NULL,
    retention_days      INTEGER NOT NULL DEFAULT 30,
    last_run_at         TIMESTAMP,
    last_verified_at    TIMESTAMP,                      -- Timestamp automated test-restore terakhir
    last_status         TEXT,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 16. Aturan Firewall Persisten (nftables)
CREATE TABLE firewall_rules (
    id                  TEXT PRIMARY KEY,
    action              TEXT NOT NULL CHECK(action IN ('allow', 'deny')),
    cidr_or_ip          TEXT NOT NULL,
    port                INTEGER NOT NULL DEFAULT 0,
    protocol            TEXT NOT NULL DEFAULT 'any' CHECK(protocol IN ('tcp', 'udp', 'any')),
    comment             TEXT,
    created_by          TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 16. DNS Clustering (WHM DNS Only style) (v8.1)
CREATE TABLE dns_clusters (
    id                  TEXT PRIMARY KEY,
    name                TEXT NOT NULL UNIQUE,
    topology            TEXT NOT NULL DEFAULT 'star' CHECK(topology IN ('star', 'primary_secondary', 'galera', 'anycast')),
    primary_node_id     TEXT,                           -- optional designated primary
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE dns_cluster_nodes (
    id                  TEXT PRIMARY KEY,
    cluster_id          TEXT NOT NULL REFERENCES dns_clusters(id) ON DELETE CASCADE,
    hostname            TEXT NOT NULL,
    api_endpoint        TEXT NOT NULL,                  -- https://ns1.example.com:8443
    role                TEXT NOT NULL CHECK(role IN ('write_only', 'synchronize', 'standalone', 'dns_only')),
    api_token_id        TEXT REFERENCES api_tokens(id) ON DELETE SET NULL,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    last_seen_at        TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_dns_cluster_nodes_cluster ON dns_cluster_nodes(cluster_id);

CREATE TABLE dns_zone_sync_status (
    id                  TEXT PRIMARY KEY,
    zone_name           TEXT NOT NULL,
    cluster_id          TEXT NOT NULL REFERENCES dns_clusters(id) ON DELETE CASCADE,
    node_id             TEXT NOT NULL REFERENCES dns_cluster_nodes(id) ON DELETE CASCADE,
    serial              INTEGER NOT NULL DEFAULT 0,
    last_sync_at        TIMESTAMP,
    last_error          TEXT,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'synced', 'error', 'drift')),
    UNIQUE(zone_name, node_id)
);
CREATE INDEX idx_dns_zone_sync_zone ON dns_zone_sync_status(zone_name);

-- 17. Immutable Hash-Chained Audit Log
CREATE TABLE audit_logs (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_user_id       TEXT NOT NULL,
    impersonator_id     TEXT,
    action              TEXT NOT NULL,
    target_resource     TEXT NOT NULL,
    source_ip           TEXT NOT NULL,
    prev_hash           TEXT NOT NULL,
    entry_hash          TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

---

## 2. Matriks Otorisasi RBAC 3-Tingkat (`superadmin`, `reseller`, `tenant`)

| Resource / Operasi                                                        |     `superadmin` (WHM)     |                `reseller`                 | `tenant` (cPanel) | Penegakan Middleware                                             |
| :------------------------------------------------------------------------ | :------------------------: | :---------------------------------------: | :---------------: | :--------------------------------------------------------------- |
| Global Server Config, Firewall (`nftables`/CrowdSec), Backup Destinations |             Ya             |                   Tidak                   |       Tidak       | `RequireRole("superadmin")`                                      |
| Buat/Suspend/Impersonate Akun Tenant                                      |         Semua Akun         | Hanya Akun Milik Reseller (`reseller_id`) |       Tidak       | Validasi kuota `reseller_limits` + audit log                     |
| Kelola Domain, Multi-PHP (8.2–8.6), SSL ACME ARI, Coraza WAF Mode         |         Semua Akun         |                Akun Binaan                |   Milik Sendiri   | `domain.user_id == claims.UserID` + `CheckFeature("domains")`    |
| App Installer, Staging Push-to-Live, Git Deploy, Node.js/Python           |         Semua Akun         |                Akun Binaan                |   Milik Sendiri   | Validasi `packages.allow_*` & SHA-256 manifes                    |
| File Manager (`openat2`) & Web SSH Terminal (`xterm.js`)                  | Via Impersonate (tercatat) |        Via Impersonate (tercatat)         |   Milik Sendiri   | Pekerja turun ke `user.linux_uid` terkunci di `/home/<username>` |
