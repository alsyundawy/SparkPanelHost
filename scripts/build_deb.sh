#!/usr/bin/env bash
set -euo pipefail

# SparkPanelHost - Debian / Ubuntu Package (.deb) Builder
# Builds a production-ready .deb package conforming to Debian policy standards.

APP_NAME="sparkpanelhost"
VERSION_INPUT="${VERSION:-}"
ARCH="${ARCH:-amd64}"
OUTPUT_DIR="${OUTPUT_DIR:-dist/packages}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

while [[ $# -gt 0 ]]; do
	case "$1" in
	-v | --version)
		VERSION_INPUT="$2"
		shift 2
		;;
	-a | --arch)
		ARCH="$2"
		shift 2
		;;
	-o | --output)
		OUTPUT_DIR="$2"
		shift 2
		;;
	-h | --help)
		echo "Usage: $0 [OPTIONS]"
		echo "Options:"
		echo "  -v, --version <ver>   Package version (e.g. 0.1.0)"
		echo "  -a, --arch <arch>     Target architecture (default: amd64)"
		echo "  -o, --output <dir>    Output directory (default: dist/packages)"
		echo "  -h, --help            Show this help message"
		exit 0
		;;
	*)
		echo "[ERROR] Unknown option: $1" >&2
		exit 1
		;;
	esac
done

if [[ -z ${VERSION_INPUT} ]]; then
	VERSION_INPUT="$(git -C "${ROOT_DIR}" describe --tags --always 2>/dev/null || echo "0.1.0")"
fi

# Sanitize version for Debian package (remove leading 'v', ensure hyphen for revision if needed)
DEB_VERSION="${VERSION_INPUT#v}"
DEB_VERSION="${DEB_VERSION//_/-}"

# Debian policy: Version must start with a digit (0-9)
if [[ ! ${DEB_VERSION} =~ ^[0-9] ]]; then
	DEB_VERSION="0.1.0+git${DEB_VERSION}"
fi

if [[ ! ${DEB_VERSION} =~ - ]]; then
	DEB_VERSION="${DEB_VERSION}-1"
fi

PACKAGE_NAME="${APP_NAME}_${DEB_VERSION}_${ARCH}"
OUTPUT_DEB="${OUTPUT_DIR}/${PACKAGE_NAME}.deb"

echo "==> Building Debian Package: ${PACKAGE_NAME}.deb (Version: ${DEB_VERSION}, Arch: ${ARCH})"
mkdir -p "${OUTPUT_DIR}"

TMP_DIR="$(mktemp -d /tmp/sparkpanel-deb-build.XXXXXX)"
cleanup() {
	rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

STAGING="${TMP_DIR}/staging"
mkdir -p "${STAGING}"

# 1. Structure file tree
mkdir -p "${STAGING}/DEBIAN" \
	"${STAGING}/usr/local/bin" \
	"${STAGING}/usr/bin" \
	"${STAGING}/lib/systemd/system" \
	"${STAGING}/var/panel/web" \
	"${STAGING}/var/panel/recipes" \
	"${STAGING}/usr/share/sparkpanel/migrations" \
	"${STAGING}/usr/share/aetherpanel/migrations" \
	"${STAGING}/etc/sparkpanel" \
	"${STAGING}/etc/sysctl.d" \
	"${STAGING}/etc/nftables.d" \
	"${STAGING}/usr/share/doc/sparkpanelhost"

# 2. Check and copy binaries
BIN_AGENT="${ROOT_DIR}/dist/bin/panel-agent"
BIN_WEB="${ROOT_DIR}/dist/bin/panel-web"

if [[ ! -f ${BIN_AGENT} ]] || [[ ! -f ${BIN_WEB} ]]; then
	echo "==> Binaries missing in dist/bin/, attempting compilation..."
	mkdir -p "${ROOT_DIR}/dist/bin"
	GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo 'local')"
	(
		cd "${ROOT_DIR}/backend"
		CGO_ENABLED=0 GOOS=linux GOARCH="${ARCH}" go build -trimpath \
			-ldflags="-s -w -X main.Version=${DEB_VERSION} -X main.Commit=${GIT_COMMIT}" \
			-o "${BIN_AGENT}" ./cmd/panel-agent
		CGO_ENABLED=0 GOOS=linux GOARCH="${ARCH}" go build -trimpath \
			-ldflags="-s -w -X main.Version=${DEB_VERSION} -X main.Commit=${GIT_COMMIT}" \
			-o "${BIN_WEB}" ./cmd/panel-web
	)
fi

install -m 0750 "${BIN_AGENT}" "${STAGING}/usr/local/bin/panel-agent"
install -m 0755 "${BIN_WEB}" "${STAGING}/usr/local/bin/panel-web"

# Symlinks in /usr/bin
ln -sf /usr/local/bin/panel-agent "${STAGING}/usr/bin/sparkpanel-agent"
ln -sf /usr/local/bin/panel-web "${STAGING}/usr/bin/sparkpanel-web"
ln -sf /usr/local/bin/panel-web "${STAGING}/usr/bin/sparkpanel"

# 3. Systemd units
if [[ -f "${ROOT_DIR}/installer/systemd/panel-agent.service" ]]; then
	install -m 0644 "${ROOT_DIR}/installer/systemd/panel-agent.service" "${STAGING}/lib/systemd/system/panel-agent.service"
fi
if [[ -f "${ROOT_DIR}/installer/systemd/panel-web.service" ]]; then
	install -m 0644 "${ROOT_DIR}/installer/systemd/panel-web.service" "${STAGING}/lib/systemd/system/panel-web.service"
fi

# 4. Frontend static files
if [[ -d "${ROOT_DIR}/dist/web" ]] && [[ -f "${ROOT_DIR}/dist/web/index.html" ]]; then
	cp -r "${ROOT_DIR}/dist/web/"* "${STAGING}/var/panel/web/"
elif [[ -d "${ROOT_DIR}/frontend/dist" ]] && [[ -f "${ROOT_DIR}/frontend/dist/index.html" ]]; then
	cp -r "${ROOT_DIR}/frontend/dist/"* "${STAGING}/var/panel/web/"
else
	echo "==> Frontend bundle missing in dist/web, attempting build..."
	(
		cd "${ROOT_DIR}/frontend"
		npm ci
		npm run build
	)
	cp -r "${ROOT_DIR}/frontend/dist/"* "${STAGING}/var/panel/web/"
fi

# 5. Recipes
if [[ -d "${ROOT_DIR}/recipes" ]]; then
	cp -r "${ROOT_DIR}/recipes/"* "${STAGING}/var/panel/recipes/"
fi

# 6. Migrations
if [[ -f "${ROOT_DIR}/backend/migrations/001_init.sql" ]]; then
	install -m 0644 "${ROOT_DIR}/backend/migrations/001_init.sql" "${STAGING}/usr/share/sparkpanel/migrations/001_init.sql"
	install -m 0644 "${ROOT_DIR}/backend/migrations/001_init.sql" "${STAGING}/usr/share/aetherpanel/migrations/001_init.sql"
fi

# 7. Config template, sysctl, nftables
cat <<'EOF' >"${STAGING}/etc/sparkpanel/config.yaml.example"
# SparkPanelHost Configuration Profile
server:
  listen: ":2087"
  agent_socket: "/run/panel/agent.sock"
  database: "/var/panel/data/sparkpanel.db"
  migrations: "/usr/share/sparkpanel/migrations/001_init.sql"
  static_dir: "/var/panel/web"
  recipe_dir: "/var/panel/recipes"
EOF
chmod 0640 "${STAGING}/etc/sparkpanel/config.yaml.example"

if [[ -f "${ROOT_DIR}/installer/sysctl/99-aetherpanel.conf" ]]; then
	install -m 0644 "${ROOT_DIR}/installer/sysctl/99-aetherpanel.conf" "${STAGING}/etc/sysctl.d/99-sparkpanel.conf"
fi
if [[ -f "${ROOT_DIR}/installer/nftables/aetherpanel.nft" ]]; then
	install -m 0640 "${ROOT_DIR}/installer/nftables/aetherpanel.nft" "${STAGING}/etc/nftables.d/sparkpanel.nft"
fi

# 8. Documentation & License
if [[ -f "${ROOT_DIR}/LICENSE" ]]; then
	install -m 0644 "${ROOT_DIR}/LICENSE" "${STAGING}/usr/share/doc/sparkpanelhost/copyright"
fi
if [[ -f "${ROOT_DIR}/README.md" ]]; then
	install -m 0644 "${ROOT_DIR}/README.md" "${STAGING}/usr/share/doc/sparkpanelhost/README.md"
fi

# 9. Maintainer scripts in DEBIAN/
cat <<EOF >"${STAGING}/DEBIAN/control"
Package: ${APP_NAME}
Version: ${DEB_VERSION}
Section: web
Priority: optional
Architecture: ${ARCH}
Maintainer: SparkPanelHost Project <https://github.com/alsyundawy/SparkPanelHost>
Depends: systemd, curl, ca-certificates, tar, zstd, jq, openssl
Recommends: nginx, mariadb-server, redis-server, nftables, restic, rclone, crowdsec, certbot
Suggests: postfix, dovecot-core, rspamd, pdns-server, pdns-backend-mysql
Provides: aetherpanel
Conflicts: aetherpanel
Replaces: aetherpanel
Description: SparkPanelHost All-in-One Linux Web Hosting Control Panel
 Modern all-in-one web hosting control panel for Ubuntu 26.04/24.04 LTS
 and Debian 13/12. Features dual-plane privilege separation (panel-agent root
 gRPC daemon + panel-web unprivileged user), Coraza WAF, Multi-PHP pool manager,
 physical MariaDB backups, and PowerDNS clustering.
EOF

cat <<'EOF' >"${STAGING}/DEBIAN/preinst"
#!/bin/sh
set -e

# Ensure unprivileged system users exist
if ! id -u panel-web >/dev/null 2>&1; then
	useradd --system --home-dir /var/panel --shell /usr/sbin/nologin panel-web || true
fi

if ! id -u vmail >/dev/null 2>&1; then
	useradd --system --uid 5000 --home-dir /var/vmail --shell /usr/sbin/nologin vmail || true
fi

exit 0
EOF

cat <<'EOF' >"${STAGING}/DEBIAN/postinst"
#!/bin/sh
set -e

# Ensure essential directories exist with strict permissions
mkdir -p /var/panel/data /var/panel/tmp /var/panel/backups /var/panel/tmp/nginx /var/panel/tmp/nginx_proxy
mkdir -p /etc/sparkpanel /var/log/sparkpanel /var/vmail

chown -R panel-web:panel-web /var/panel /etc/sparkpanel /var/log/sparkpanel
chown -R vmail:vmail /var/vmail
chmod 0750 /var/panel /etc/sparkpanel /var/log/sparkpanel /var/vmail

# Backward compatibility symlink for /etc/aetherpanel
if [ ! -d /etc/aetherpanel ] && [ ! -L /etc/aetherpanel ]; then
	ln -sf /etc/sparkpanel /etc/aetherpanel || true
fi
if [ ! -d /var/log/aetherpanel ] && [ ! -L /var/log/aetherpanel ]; then
	ln -sf /var/log/sparkpanel /var/log/aetherpanel || true
fi

# Default configuration template
if [ ! -f /etc/sparkpanel/config.yaml ] && [ -f /etc/sparkpanel/config.yaml.example ]; then
	cp /etc/sparkpanel/config.yaml.example /etc/sparkpanel/config.yaml
	chown panel-web:panel-web /etc/sparkpanel/config.yaml
	chmod 0640 /etc/sparkpanel/config.yaml
fi

# Generate secrets.env if missing
if [ ! -f /etc/sparkpanel/secrets.env ]; then
	PDNS_KEY=$(openssl rand -hex 32 2>/dev/null || tr -dc 'a-f0-9' < /dev/urandom | head -c 64)
	MASTER_KEY=$(openssl rand -hex 32 2>/dev/null || tr -dc 'a-f0-9' < /dev/urandom | head -c 64)
	cat <<SECRETS >/etc/sparkpanel/secrets.env
PDNS_API_URL=http://127.0.0.1:8081
PDNS_API_KEY=${PDNS_KEY}
PANEL_MASTER_KEY=${MASTER_KEY}
SECRETS
	chown root:panel-web /etc/sparkpanel/secrets.env
	chmod 0640 /etc/sparkpanel/secrets.env
fi

# Reload systemd and enable services if systemd is active
if [ -d /run/systemd/system ]; then
	systemctl daemon-reload || true
	systemctl enable panel-agent.service panel-web.service || true
fi

exit 0
EOF

cat <<'EOF' >"${STAGING}/DEBIAN/prerm"
#!/bin/sh
set -e

# Stop running services before upgrade/removal
if [ -d /run/systemd/system ]; then
	systemctl stop panel-web.service 2>/dev/null || true
	systemctl stop panel-agent.service 2>/dev/null || true
fi

exit 0
EOF

cat <<'EOF' >"${STAGING}/DEBIAN/postrm"
#!/bin/sh
set -e

if [ -d /run/systemd/system ]; then
	systemctl daemon-reload || true
fi

exit 0
EOF

chmod 0755 "${STAGING}/DEBIAN/preinst" "${STAGING}/DEBIAN/postinst" "${STAGING}/DEBIAN/prerm" "${STAGING}/DEBIAN/postrm"
chmod 0644 "${STAGING}/DEBIAN/control"

# 10. Generate md5sums
(
	cd "${STAGING}"
	find . -type f ! -path "./DEBIAN/*" -print0 | xargs -0 md5sum | sed 's| \./| |' >"${STAGING}/DEBIAN/md5sums"
	chmod 0644 "${STAGING}/DEBIAN/md5sums"
)

# 11. Build .deb package
echo "==> Packing into ${OUTPUT_DEB}..."
if command -v dpkg-deb >/dev/null 2>&1; then
	dpkg-deb --build --root-owner-group "${STAGING}" "${OUTPUT_DEB}"
else
	# Fallback for systems without dpkg-deb (e.g. macOS): assemble standard Debian ar archive
	PKG_TMP="${TMP_DIR}/ar_pkg"
	mkdir -p "${PKG_TMP}"
	echo "2.0" >"${PKG_TMP}/debian-binary"
	tar -czf "${PKG_TMP}/control.tar.gz" -C "${STAGING}/DEBIAN" .
	tar -czf "${PKG_TMP}/data.tar.gz" --exclude="./DEBIAN" -C "${STAGING}" .
	ar -rc "${OUTPUT_DEB}" "${PKG_TMP}/debian-binary" "${PKG_TMP}/control.tar.gz" "${PKG_TMP}/data.tar.gz"
fi

# 12. Checksum
(
	cd "${OUTPUT_DIR}"
	sha256sum "$(basename "${OUTPUT_DEB}")" >"$(basename "${OUTPUT_DEB}").sha256"
)

echo "==> Debian package successfully generated:"
echo "    -> ${OUTPUT_DEB}"
echo "    -> ${OUTPUT_DEB}.sha256"
