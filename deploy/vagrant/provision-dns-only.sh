#!/usr/bin/env bash
# ==============================================================================
# AetherPanel v8.1 — Automated DNS-Only Node Provisioner
# Installs and hardens PowerDNS Authoritative Server with REST API enabled
# Target OS: Debian 12/13 & Ubuntu 24.04/26.04
# Conforms to strict bash-pro guidelines: ShellCheck clean, defensive execution
# ==============================================================================

set -Eeuo pipefail
shopt -s inherit_errexit 2>/dev/null || true

export DEBIAN_FRONTEND=noninteractive
export LC_ALL=C.UTF-8
export PATH="/usr/sbin:/usr/bin:/sbin:/bin"

CLUSTER_API_KEY="dns-cluster-secret-key-v8.1"
PDNS_CONF="/etc/powerdns/pdns.conf"
PDNS_DB_DIR="/var/lib/powerdns"
PDNS_DB_FILE="${PDNS_DB_DIR}/pdns.sqlite3"

echo "==> [DNS-Only Node] Starting automated provisioning..."

# 1. Update package lists and install PowerDNS Authoritative with SQLite3 backend
echo "==> [DNS-Only Node] Installing PowerDNS Authoritative and dependencies..."
apt-get update -qq
apt-get install -y -qq \
	pdns-server \
	pdns-backend-sqlite3 \
	sqlite3 \
	curl \
	jq \
	nftables \
	ca-certificates

# 2. Disable systemd-resolved DNS stub listener to avoid port 53 collision if active
if systemctl is-active --quiet systemd-resolved 2>/dev/null; then
	echo "==> [DNS-Only Node] Configuring systemd-resolved to release port 53..."
	mkdir -p /etc/systemd/resolved.conf.d
	cat <<'EOF' >/etc/systemd/resolved.conf.d/disable-stub.conf
[Resolve]
DNSStubListener=no
EOF
	systemctl restart systemd-resolved || true
fi

# 3. Initialize SQLite3 database schema for PowerDNS Authoritative if missing
mkdir -p "${PDNS_DB_DIR}"
if [[ ! -s ${PDNS_DB_FILE} ]]; then
	echo "==> [DNS-Only Node] Initializing PowerDNS SQLite3 database schema..."
	sqlite3 "${PDNS_DB_FILE}" <<'EOF'
CREATE TABLE IF NOT EXISTS domains (
  id                    INTEGER PRIMARY KEY,
  name                  VARCHAR(255) NOT NULL COLLATE NOCASE,
  master                VARCHAR(128) DEFAULT NULL,
  last_check            INTEGER DEFAULT NULL,
  type                  VARCHAR(8) NOT NULL,
  notified_serial       INTEGER DEFAULT NULL,
  account               VARCHAR(40) DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS name_index ON domains(name);

CREATE TABLE IF NOT EXISTS records (
  id                    INTEGER PRIMARY KEY,
  domain_id             INTEGER DEFAULT NULL,
  name                  VARCHAR(255) DEFAULT NULL,
  type                  VARCHAR(10) DEFAULT NULL,
  content               VARCHAR(65535) DEFAULT NULL,
  ttl                   INTEGER DEFAULT NULL,
  prio                  INTEGER DEFAULT NULL,
  disabled              BOOLEAN DEFAULT 0,
  ordername             VARCHAR(255),
  auth                  BOOLEAN DEFAULT 1,
  FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS nametype_index ON records(name,type);
CREATE INDEX IF NOT EXISTS domain_id ON records(domain_id);
CREATE INDEX IF NOT EXISTS ordername ON records(ordername);

CREATE TABLE IF NOT EXISTS supermasters (
  ip                    VARCHAR(64) NOT NULL,
  nameserver            VARCHAR(255) NOT NULL COLLATE NOCASE,
  account               VARCHAR(40) NOT NULL
);

CREATE TABLE IF NOT EXISTS comments (
  id                    INTEGER PRIMARY KEY,
  domain_id             INTEGER NOT NULL,
  name                  VARCHAR(255) NOT NULL,
  type                  VARCHAR(10) NOT NULL,
  modified_at           INTEGER NOT NULL,
  account               VARCHAR(40) DEFAULT NULL,
  comment               VARCHAR(65535) NOT NULL,
  FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS domainmetadata (
  id                    INTEGER PRIMARY KEY,
  domain_id             INT NOT NULL,
  kind                  VARCHAR(32) COLLATE NOCASE,
  content               TEXT,
  FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cryptokeys (
  id                    INTEGER PRIMARY KEY,
  domain_id             INT NOT NULL,
  flags                 INT NOT NULL,
  active                BOOL,
  published             BOOL DEFAULT 1,
  content               TEXT,
  FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tsigkeys (
  id                    INTEGER PRIMARY KEY,
  name                  VARCHAR(255) COLLATE NOCASE,
  algorithm             VARCHAR(50) COLLATE NOCASE,
  secret                VARCHAR(255)
);
CREATE UNIQUE INDEX IF NOT EXISTS namealgoindex ON tsigkeys(name, algorithm);
EOF
fi

# Set appropriate permissions for PowerDNS daemon user
PDNS_USER="pdns"
if id -u "_powerdns" >/dev/null 2>&1; then
	PDNS_USER="_powerdns"
fi
chown -R "${PDNS_USER}:${PDNS_USER}" "${PDNS_DB_DIR}"
chmod 0750 "${PDNS_DB_DIR}"
chmod 0640 "${PDNS_DB_FILE}"

# 4. Generate hardened PowerDNS configuration with REST API enabled
echo "==> [DNS-Only Node] Configuring PowerDNS Authoritative and REST API..."
cat <<EOF >"${PDNS_CONF}"
# AetherPanel v8.1 — Hardened PowerDNS Authoritative Config (DNS-Only Node)
launch=gsqlite3
gsqlite3-database=${PDNS_DB_FILE}

# Local & Network Binding
local-address=0.0.0.0,::
local-port=53

# REST API & Webserver Configuration
webserver=yes
webserver-address=0.0.0.0
webserver-port=8081
webserver-allow-from=127.0.0.1,::1,192.168.56.0/24,10.0.0.0/8
api=yes
api-key=${CLUSTER_API_KEY}

# Security & Process Hardening
setuid=${PDNS_USER}
setgid=${PDNS_USER}
security-poll-suffix=
guardian=yes
daemon=yes
version-string=powerdns
EOF

chown "root:${PDNS_USER}" "${PDNS_CONF}"
chmod 0640 "${PDNS_CONF}"

# 5. Restart and enable PowerDNS service
echo "==> [DNS-Only Node] Starting PowerDNS service..."
systemctl restart pdns
systemctl enable pdns

echo "==> [DNS-Only Node] DNS-Only Provisioning Complete! PowerDNS API listening on 0.0.0.0:8081"
