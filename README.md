<!-- markdownlint-disable -->

# SparkPanelHost (v0.1.0)

[![MegaLinter](https://github.com/alsyundawy/SparkPanelHost/actions/workflows/megalinter.yml/badge.svg)](https://github.com/alsyundawy/SparkPanelHost/actions/workflows/megalinter.yml)
[![CI Security & Verification](https://github.com/alsyundawy/SparkPanelHost/actions/workflows/ci.yml)](https://github.com/alsyundawy/SparkPanelHost/actions/workflows/ci.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

**SparkPanelHost** adalah _All-in-One Linux Web Hosting Control Panel_ modern, aman, dan berkinerja tinggi yang dirancang sebagai alternatif tangguh untuk cPanel/WHM, Plesk, dan CloudPanel. Dibangun secara native untuk **Ubuntu 26.04 LTS (Resolute) / 24.04 LTS (Noble)** dan **Debian 13 (Trixie) / 12 (Bookworm)**, SparkPanelHost siap digunakan untuk hosting personal, reseller hosting (RBAC 3-tingkat), hingga otomasi billing komersial (WHMCS/Blesta).

---

## Fitur Utama & Keunggulan Arsitektur

1. **Dual-Plane Privilege Separation (Zero-Trust):**
   - `sparkpanel-web` (Go): Control plane unprivileged tanpa akses root, tanpa sudo, berjalan dalam systemd sandbox terisolasi.
   - `sparkpanel-agent` (Go): Daemon sistem berhak akses root yang dikeraskan (`ProtectSystem=strict`, `MemoryDenyWriteExecute=yes`), hanya mendengarkan Unix Domain Socket `/run/panel/agent.sock`, dan memverifikasi identitas proses via kernel `SO_PEERCRED`.
2. **Zero-Shell Execution Engine (`execsafe`):**
   - Larangan mutlak terhadap konkatenasi shell `/bin/sh -c` (anti CWE-78). Seluruh eksekusi proses OS menggunakan `exec.CommandContext` dengan argumen array terpisah dan lingkungan tersterilisasi deterministik (`PATH=/usr/sbin:/usr/bin:/sbin:/bin`, `LC_ALL=C.UTF-8`).
3. **Kernel-Enforced Path Traversal Immunity (`unix.Openat2`):**
   - Ekstraksi arsip `.tar.gz` dan `.zip` diamankan secara langsung di level kernel menggunakan `unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS` (anti CWE-22, CWE-59, CWE-362), disertai limit reader anti-Zip-Bomb (CWE-409).
4. **Authoritative DNS Clustering (Gaya WHM DNS Only):**
   - Integrasi PowerDNS Authoritative 5.1+ via REST API murni, mendukung topologi Star Topology, replikasi TSIG/AXFR, 1-click DNSSEC, dan ACME DNS-01.
5. **Modern Stack Web Server & WAF:**
   - Nginx Mainline + Coraza WAF (Go-native memory safe) + OWASP CRS v4 dengan profil pengecualian per-CMS (WordPress, Joomla, Drupal, Roundcube).
6. **Multi-PHP Runtime Manager:**
   - Dukungan PHP terisolasi per-user: PHP 8.2 (EOL warning), 8.3, 8.4, 8.5 (default), dan 8.6 GA dengan FPM pool mandiri.
7. **Enterprise Mail Stack:**
   - Postfix + Dovecot 2.4.5+ + Rspamd 4.3+ dengan dual-signing DKIM (RSA-2048 + Ed25519), ARC, DMARC, MTA-STS, dan monitor live RBL.
8. **Modern UI/UX Frontend (React 19 + Tailwind CSS v4):**
   - Bento Grid modular, OLED Dark Mode default + Light Mode WCAG AAA, Command Palette (`Cmd+K`), Web Terminal berbasis xterm.js (PTY terkunci UID tenant), dan Monaco Editor.

---

## Struktur Monorepo

```text
SparkPanelHost/
├── .github/
│   └── workflows/
│       ├── megalinter.yml              # MegaLinter analysis (Go, TS, Bash, YAML, Markdown)
│       ├── ci.yml                      # CI pipeline: test -race, staticcheck, tsc, build
│       ├── ci-security.yml             # SAST (gosec, govulncheck) & race detector
│       ├── auto-update-manifests.yml   # Auto-update katalog resep aplikasi resmi
│       └── release.yml                 # Packaging rilis tag otomatis tar.zst
├── api/
│   └── proto/agent/v1/agent.proto      # 32 gRPC contracts daemon
├── backend/
│   ├── cmd/
│   │   ├── panel-agent/main.go         # Privileged daemon root
│   │   └── panel-web/main.go           # Unprivileged web control plane
│   └── internal/
│       ├── agent/                      # drivers, execsafe, fileops, netfetch, server
│       └── web/                        # api, auth, store, agentclient, workers
├── frontend/                           # React 19 + TypeScript + Vite + Tailwind CSS v4
├── installer/
│   ├── install.sh                      # Idempotent DEB822 installer script
│   ├── nftables/                       # Aturan firewall atomik inet
│   └── systemctl/                      # Hardened systemd unit templates
├── recipes/                            # Katalog 1-Click App Installer (JSON Schema v7)
├── deploy/vagrant/                     # Multi-Node Cluster Vagrant VM testing harness
├── docs/                               # Dokumentasi spesifikasi arsitektur komprehensif
├── LICENSE                             # GNU General Public License v3.0 (GPL-3.0)
└── Makefile                            # Build automation & orchestration
```

---

## Memulai & Menjalankan (Quickstart)

### Prasyarat Pembangunan Lokal

- **Go:** 1.23+
- **Node.js:** 20+ / 22+ (npm 10+)
- **OS Target:** Ubuntu 24.04/26.04 LTS atau Debian 12/13

### Kompilasi dari Sumber

```bash
# 1. Clone repository
git clone https://github.com/alsyundawy/SparkPanelHost.git
cd SparkPanelHost

# 2. Build seluruh biner dan frontend
make build-all

# 3. Jalankan pengujian unit & static analysis
make test
```

---

## Lisensi

Proyek ini dilisensikan di bawah ketentuan **GNU General Public License v3.0 (GPL-3.0-or-later)**. Lihat berkas [LICENSE](LICENSE) untuk informasi hak cipta dan lisensi lengkap.
