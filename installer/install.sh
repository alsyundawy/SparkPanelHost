#!/usr/bin/env bash
set -euo pipefail

# SparkPanelHost v0.1.0 Idempotent Server Installer
# Mendukung Ubuntu 26.04 LTS (Resolute) / 24.04 LTS (Noble) & Debian 13 (Trixie) / 12 (Bookworm)

PACKAGE_DEB=""
SKIP_REPOS=false
SKIP_SERVICES=false

while [[ $# -gt 0 ]]; do
	case "$1" in
	-p | --package | --deb)
		PACKAGE_DEB="$2"
		shift 2
		;;
	--skip-repos)
		SKIP_REPOS=true
		shift
		;;
	--skip-services)
		SKIP_SERVICES=true
		shift
		;;
	-h | --help)
		echo "Usage: $0 [OPTIONS]"
		echo "Options:"
		echo "  -p, --package, --deb <file.deb>  Install specified .deb package file directly"
		echo "  --skip-repos                     Skip adding external DEB822 repositories"
		echo "  --skip-services                  Skip starting/enabling systemd services"
		echo "  -h, --help                       Show this help message"
		exit 0
		;;
	*)
		echo "[ERROR] Unknown option: $1" >&2
		exit 1
		;;
	esac
done

if [[ ${EUID} -ne 0 ]]; then
	echo "[ERROR] Skrip installer wajib dijalankan sebagai root." >&2
	exit 1
fi

if [[ ! -r /etc/os-release ]]; then
	echo "[ERROR] /etc/os-release tidak ditemukan." >&2
	exit 1
fi

# shellcheck source=/dev/null
source /etc/os-release
OS_ID="${ID:-unknown}"
OS_CODENAME="${VERSION_CODENAME:-unknown}"
ARCH="$(dpkg --print-architecture)"

echo "==> Mendeteksi OS: ${OS_ID} (${OS_CODENAME}) [${ARCH}]"
case "${OS_ID}:${OS_CODENAME}" in
ubuntu:noble | ubuntu:resolute | debian:bookworm | debian:trixie)
	;;
*)
	echo "[ERROR] OS tidak didukung. Gunakan Ubuntu 24.04/26.04 LTS atau Debian 12/13." >&2
	exit 1
	;;
esac

TOTAL_RAM_MB="$(awk '/MemTotal/ {print int($2/1024)}' /proc/meminfo 2>/dev/null || echo "2048")"
if [[ ${TOTAL_RAM_MB} -lt 1800 ]]; then
	echo "[WARN] RAM terdeteksi ${TOTAL_RAM_MB} MB. Disarankan minimal 2048 MB RAM." >&2
fi

export DEBIAN_FRONTEND=noninteractive
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Auto-detect .deb package if not explicitly provided
if [[ -z ${PACKAGE_DEB} ]]; then
	if compgen -G "./sparkpanelhost*.deb" >/dev/null; then
		for f in ./sparkpanelhost*.deb; do
			if [[ -f ${f} ]]; then
				PACKAGE_DEB="${f}"
				break
			fi
		done
	elif compgen -G "${SCRIPT_DIR}/../dist/packages/sparkpanelhost*.deb" >/dev/null; then
		for f in "${SCRIPT_DIR}/../dist/packages"/sparkpanelhost*.deb; do
			if [[ -f ${f} ]]; then
				PACKAGE_DEB="${f}"
				break
			fi
		done
	elif compgen -G "${SCRIPT_DIR}/sparkpanelhost*.deb" >/dev/null; then
		for f in "${SCRIPT_DIR}"/sparkpanelhost*.deb; do
			if [[ -f ${f} ]]; then
				PACKAGE_DEB="${f}"
				break
			fi
		done
	fi
fi

echo "==> 1. Memperbarui paket dasar & menyiapkan disk staging (mitigasi tmpfs /tmp Debian 13)..."
apt-get update -y
apt-get install -y curl gnupg2 ca-certificates lsb-release apt-transport-https software-properties-common quota acl zstd unzip tar openssl jq

id -u panel-web &>/dev/null || useradd --system --home-dir /var/panel --shell /usr/sbin/nologin panel-web
id -u vmail &>/dev/null || useradd --system --uid 5000 --home-dir /var/vmail --shell /usr/sbin/nologin vmail

mkdir -p /var/panel/{data,tmp,backups} /var/panel/tmp/{nginx,nginx_proxy} /var/vmail /etc/sparkpanel /etc/aetherpanel /etc/coraza /etc/nftables.d /var/log/sparkpanel /var/log/aetherpanel
chown -R panel-web:panel-web /var/panel /etc/sparkpanel /etc/aetherpanel /var/log/sparkpanel /var/log/aetherpanel
chown -R vmail:vmail /var/vmail
chmod 0750 /var/panel /var/vmail /etc/sparkpanel /etc/aetherpanel /var/log/sparkpanel /var/log/aetherpanel

echo "==> 2. Menerapkan Kernel Hardening & TCP BBR (sysctl)..."
if [[ -f "${SCRIPT_DIR}/sysctl/99-aetherpanel.conf" ]]; then
	install -m 0644 "${SCRIPT_DIR}/sysctl/99-aetherpanel.conf" /etc/sysctl.d/99-sparkpanel.conf
	sysctl --system >/dev/null 2>&1 || true
fi

if [[ ${SKIP_REPOS} == "true" ]]; then
	echo "==> 3. Melewati penyiapan repositori eksternal (--skip-repos)..."
else
	echo "==> 3. Menyiapkan repositori DEB822 resmi (Multi-PHP 8.2-8.6, MariaDB 12.3, PowerDNS 5.1, Rspamd 4.3, CrowdSec 1.8)..."
	mkdir -p /usr/share/keyrings

	# 3a. Multi-PHP (Sury untuk Debian / Ondrej PPA untuk Ubuntu)
	if [[ ${OS_ID} == "debian" ]]; then
		curl -fsSL https://packages.sury.org/php/apt.gpg | gpg --dearmor --yes -o /usr/share/keyrings/sury-php.gpg
		cat <<EOF >/etc/apt/sources.list.d/sury-php.sources
Types: deb
URIs: https://packages.sury.org/php/
Suites: ${OS_CODENAME}
Components: main
Signed-By: /usr/share/keyrings/sury-php.gpg
EOF
	else
		add-apt-repository -y ppa:ondrej/php
	fi

	# 3b. MariaDB 12.3 LTS
	curl -fsSL https://mariadb.org/mariadb_release_signing_key.pgp | gpg --dearmor --yes -o /usr/share/keyrings/mariadb.gpg
	cat <<EOF >/etc/apt/sources.list.d/mariadb.sources
Types: deb
URIs: https://dlm.mariadb.com/repo/mariadb-server/12.3/repo/${OS_ID}
Suites: ${OS_CODENAME}
Components: main
Signed-By: /usr/share/keyrings/mariadb.gpg
EOF

	# 3c. PowerDNS Authoritative 5.1
	curl -fsSL https://repo.powerdns.com/FD380FBB-pub.asc | gpg --dearmor --yes -o /usr/share/keyrings/powerdns.gpg
	cat <<EOF >/etc/apt/sources.list.d/powerdns.sources
Types: deb
URIs: https://repo.powerdns.com/${OS_ID}
Suites: ${OS_CODENAME}-auth-51
Components: main
Signed-By: /usr/share/keyrings/powerdns.gpg
EOF

	# 3d. Rspamd 4.3+
	curl -fsSL https://rspamd.com/apt-stable/gpg.key | gpg --dearmor --yes -o /usr/share/keyrings/rspamd.gpg
	cat <<EOF >/etc/apt/sources.list.d/rspamd.sources
Types: deb
URIs: https://rspamd.com/apt-stable/
Suites: ${OS_CODENAME}
Components: main
Signed-By: /usr/share/keyrings/rspamd.gpg
EOF

	# 3e. CrowdSec 1.8+
	curl -fsSL https://packagecloud.io/crowdsec/crowdsec/gpgkey | gpg --dearmor --yes -o /usr/share/keyrings/crowdsec.gpg
	cat <<EOF >/etc/apt/sources.list.d/crowdsec.sources
Types: deb
URIs: https://packagecloud.io/crowdsec/crowdsec/${OS_ID}/
Suites: ${OS_CODENAME}
Components: main
Signed-By: /usr/share/keyrings/crowdsec.gpg
EOF

	apt-get update -y
fi

echo "==> 4. Memasang Multi-PHP FPM (8.2 - 8.6)..."
for PHP_VER in 8.2 8.3 8.4 8.5 8.6; do
	if apt-cache show "php${PHP_VER}-fpm" &>/dev/null; then
		echo "    -> Memasang PHP ${PHP_VER}-FPM..."
		apt-get install -y "php${PHP_VER}-fpm" "php${PHP_VER}-cli" "php${PHP_VER}-mysql" "php${PHP_VER}-pgsql" "php${PHP_VER}-curl" "php${PHP_VER}-gd" "php${PHP_VER}-intl" "php${PHP_VER}-mbstring" "php${PHP_VER}-xml" "php${PHP_VER}-zip" "php${PHP_VER}-opcache" || true
	fi
done

echo "==> 5. Memasang Nginx, MariaDB 12.3 LTS, PowerDNS 5.1, Mail Stack, nftables, CrowdSec 1.8 & Restic..."
apt-get install -y nginx mariadb-server mariadb-client pdns-server pdns-backend-mysql postfix dovecot-core dovecot-imapd dovecot-pop3d dovecot-lmtpd dovecot-mysql rspamd redis-server nftables crowdsec crowdsec-firewall-bouncer-nftables certbot restic rclone || true

# Konfigurasi direktori temp Nginx berbasis disk (anti-tmpfs OOM Debian 13)
cat <<'EOF' >/etc/nginx/conf.d/00-sparkpanel-temp.conf
client_body_temp_path /var/panel/tmp/nginx 1 2;
proxy_temp_path /var/panel/tmp/nginx_proxy 1 2;
server_tokens off;
EOF

# Generate secret keys lokal jika belum ada
SECRETS_FILE="/etc/sparkpanel/secrets.env"
if [[ ! -f ${SECRETS_FILE} ]]; then
	PDNS_API_KEY="$(openssl rand -hex 32 2>/dev/null || tr -dc 'a-f0-9' </dev/urandom | head -c 64)"
	PANEL_MASTER_KEY="$(openssl rand -hex 32 2>/dev/null || tr -dc 'a-f0-9' </dev/urandom | head -c 64)"
	cat <<EOF >"${SECRETS_FILE}"
PDNS_API_URL=http://127.0.0.1:8081
PDNS_API_KEY=${PDNS_API_KEY}
PANEL_MASTER_KEY=${PANEL_MASTER_KEY}
EOF
	chown root:panel-web "${SECRETS_FILE}"
	chmod 0640 "${SECRETS_FILE}"
fi

# Backward compatibility secrets file
if [[ ! -f /etc/aetherpanel/secrets.env ]]; then
	ln -sf "${SECRETS_FILE}" /etc/aetherpanel/secrets.env 2>/dev/null || cp "${SECRETS_FILE}" /etc/aetherpanel/secrets.env
fi

# Terapkan aturan firewall atomik nftables
if [[ -f "${SCRIPT_DIR}/nftables/aetherpanel.nft" ]]; then
	install -m 0640 "${SCRIPT_DIR}/nftables/aetherpanel.nft" /etc/nftables.d/sparkpanel.nft
	nft -f /etc/nftables.d/sparkpanel.nft 2>/dev/null || true
fi

# 6. Konfigurasi PowerDNS REST API & MariaDB Hardening
if [[ -d /etc/powerdns ]]; then
	if ! grep -q "api=yes" /etc/powerdns/pdns.conf 2>/dev/null; then
		PDNS_KEY_VAL="$(grep '^PDNS_API_KEY=' "${SECRETS_FILE}" 2>/dev/null | cut -d'=' -f2- || echo "")"
		cat <<EOF >>/etc/powerdns/pdns.conf

# SparkPanelHost PowerDNS REST API Configuration
api=yes
api-key=${PDNS_KEY_VAL}
webserver=yes
webserver-address=127.0.0.1
webserver-port=8081
webserver-allow-from=127.0.0.1
EOF
	fi
fi

if [[ -d /etc/mysql/mariadb.conf.d ]]; then
	cat <<'EOF' >/etc/mysql/mariadb.conf.d/99-sparkpanel.cnf
[mysqld]
local_infile = 0
symbolic-links = 0
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
EOF
fi

# 7. Memasang Paket SparkPanelHost (.deb) atau Salin Biner Manual
if [[ -n ${PACKAGE_DEB} && -f ${PACKAGE_DEB} ]]; then
	echo "==> 7. Memasang paket SparkPanelHost Debian (.deb): ${PACKAGE_DEB}..."
	apt-get install -y "${PACKAGE_DEB}" || (dpkg -i "${PACKAGE_DEB}" && apt-get install -f -y)
else
	echo "==> 7. Memasang berkas migrasi, resep & unit systemd manual..."
	mkdir -p /usr/share/sparkpanel/migrations /usr/share/aetherpanel/migrations /var/panel/recipes
	if [[ -f "${SCRIPT_DIR}/../backend/migrations/001_init.sql" ]]; then
		install -m 0644 "${SCRIPT_DIR}/../backend/migrations/001_init.sql" /usr/share/sparkpanel/migrations/001_init.sql
		install -m 0644 "${SCRIPT_DIR}/../backend/migrations/001_init.sql" /usr/share/aetherpanel/migrations/001_init.sql
	fi

	if [[ -d "${SCRIPT_DIR}/../recipes" ]]; then
		cp -r "${SCRIPT_DIR}/../recipes/"* /var/panel/recipes/ 2>/dev/null || true
		chown -R panel-web:panel-web /var/panel/recipes
	fi

	# Pasang biner jika build dist/bin tersedia
	if [[ -f "${SCRIPT_DIR}/../dist/bin/panel-agent" ]]; then
		install -m 0750 -o root -g root "${SCRIPT_DIR}/../dist/bin/panel-agent" /usr/local/bin/panel-agent
	fi
	if [[ -f "${SCRIPT_DIR}/../dist/bin/panel-web" ]]; then
		install -m 0755 -o root -g root "${SCRIPT_DIR}/../dist/bin/panel-web" /usr/local/bin/panel-web
	fi

	# Pasang unit systemd
	if [[ -f "${SCRIPT_DIR}/systemd/panel-agent.service" ]]; then
		install -m 0644 "${SCRIPT_DIR}/systemd/panel-agent.service" /etc/systemd/system/panel-agent.service
	fi
	if [[ -f "${SCRIPT_DIR}/systemd/panel-web.service" ]]; then
		install -m 0644 "${SCRIPT_DIR}/systemd/panel-web.service" /etc/systemd/system/panel-web.service
	fi
fi

# Symlinks in /usr/bin
if [[ -f /usr/local/bin/panel-web ]]; then
	ln -sf /usr/local/bin/panel-web /usr/bin/sparkpanel-web
	ln -sf /usr/local/bin/panel-web /usr/bin/sparkpanel
fi
if [[ -f /usr/local/bin/panel-agent ]]; then
	ln -sf /usr/local/bin/panel-agent /usr/bin/sparkpanel-agent
fi

# 8. Aktivasi Layanan Systemd
if [[ ${SKIP_SERVICES} == "false" ]] && [[ -d /run/systemd/system ]]; then
	echo "==> 8. Mengaktifkan layanan systemd inti..."
	systemctl daemon-reload || true
	systemctl enable --now mariadb nginx pdns postfix dovecot redis-server rspamd nftables crowdsec || true

	if [[ -f /usr/local/bin/panel-agent && -f /usr/local/bin/panel-web ]]; then
		systemctl enable --now panel-agent panel-web || true
		echo "==> Daemon SparkPanelHost (panel-agent & panel-web) berhasil diaktifkan!"
	fi
else
	echo "==> 8. Melewati aktivasi systemd (--skip-services atau lingkungan container non-systemd)..."
fi

SERVER_IP="$(hostname -I 2>/dev/null | awk '{print $1}' || echo "IP_SERVER")"

echo "=========================================================================="
echo "  Selamat! SparkPanelHost v0.1.0 berhasil dipasang pada server Anda.     "
echo "  URL Control Panel Admin: http://${SERVER_IP}:2087/                    "
echo "  Kredensial awal tercatat pada log journalctl: journalctl -u panel-web  "
echo "=========================================================================="
