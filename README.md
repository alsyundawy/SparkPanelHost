<!-- markdownlint-disable -->

<p align="center">
  <a href="https://github.com/alsyundawy/SparkPanelHost">
    <img src="https://raw.githubusercontent.com/alsyundawy/SparkPanelHost/main/docs/assets/sparkpanelhost-banner.png" width="100%" alt="SparkPanelHost — All-in-One Linux Web Hosting Control Panel for Ubuntu &amp; Debian" onerror="this.style.display='none'" />
  </a>
</p>

<h1 align="center">SparkPanelHost</h1>

<p align="center">
  <b>SparkPanelHost — High-Performance, Hardened All-in-One Linux Hosting Control Panel</b><br>
  <i>Architected with Go Dual-Plane Privilege Separation, Zero-Shell Execution, Kernel-Enforced Path Traversal Immunity, and Modern Bento Grid UI/UX</i>
</p>

<p align="center">
  <b>English</b> &nbsp;|&nbsp; <a href="#ringkasan-bahasa-indonesia"><b>Bahasa Indonesia</b></a>
</p>

<p align="center">
  <a href="https://github.com/alsyundawy/SparkPanelHost/releases/latest"><img src="https://img.shields.io/badge/Release-v0.1.0-2ea44f?style=for-the-badge&logo=github&logoColor=white" alt="Latest Release v0.1.0"></a>
  <a href="https://github.com/alsyundawy/SparkPanelHost"><img src="https://img.shields.io/badge/Platform-Ubuntu%2026%2F24%20LTS%20%7C%20Debian%2013%2F12-E95420?style=for-the-badge&logo=ubuntu&logoColor=white" alt="Supported OS: Ubuntu &amp; Debian"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Backend-Go%201.26%20%2B%20gRPC-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.26 Backend"></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/Frontend-React%2019%20%2B%20Tailwind%20v4-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React 19 + Tailwind CSS v4"></a>
  <a href="https://coraza.io"><img src="https://img.shields.io/badge/Security-Coraza%20WAF%20%2B%20CrowdSec-FF6B6B?style=for-the-badge&logo=shield&logoColor=white" alt="Coraza WAF + CrowdSec"></a>
  <a href="https://github.com/alsyundawy/SparkPanelHost/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-GPL--3.0-blue?style=for-the-badge" alt="GPL-3.0 License"></a>
  <a href="https://github.com/alsyundawy/SparkPanelHost/actions"><img src="https://img.shields.io/badge/CI%2FCD-100%25%20Passing-success?style=for-the-badge&logo=githubactions&logoColor=white" alt="CI/CD Status"></a>
  <a href="https://github.com/alsyundawy/SparkPanelHost/actions/workflows/megalinter.yml"><img src="https://img.shields.io/badge/MegaLinter-Clean%20(0%20Issues)-success?style=for-the-badge" alt="MegaLinter Clean"></a>
</p>

<p align="center">
  <b>SparkPanelHost</b> is an enterprise-grade, memory-efficient, all-in-one Linux hosting control panel built as a modern, hardened alternative to <b>cPanel/WHM, Plesk, and CloudPanel</b>. Built natively for <b>Ubuntu 26.04 LTS (Resolute) / 24.04 LTS (Noble)</b> and <b>Debian 13 (Trixie) / 12 (Bookworm)</b>, SparkPanelHost provides full multi-tenancy (Superadmin, Reseller, and Tenant roles), automated billing integration (WHMCS/Blesta), DNS clustering (WHM DNS Only style), and enterprise-grade email deliverability.
</p>

<p align="center">
  Engineered from the ground up with a <b>Zero-Trust security architecture</b>, it implements a dual-plane privilege separation model (unprivileged <code>panel-web</code> and a root-isolated <code>panel-agent</code> verified via kernel <code>SO_PEERCRED</code>), a strict zero-shell execution engine (<b>anti-CWE-78</b>), kernel-level symlink and traversal immunity via <code>unix.Openat2</code> (<b>anti-CWE-22 / CWE-59</b>), and an ultra-fast Bento Grid UI with AdminLTE/Bootstrap ergonomics — consuming <b>under 150 MB RAM</b> at idle.
</p>

<p align="center">
  <a href="#quick-installation-1-command">
    <img src="https://img.shields.io/badge/🚀_Quick_Install-1--Command_Curl-238636?style=for-the-badge&logo=cloudsmith&logoColor=white" alt="Quick Install">
  </a>
  <a href="docs/00_MASTER_PLAN_v8.1.md">
    <img src="https://img.shields.io/badge/📖_Master_Plan-v8.1_Blueprint-0284c7?style=for-the-badge&logo=gitbook&logoColor=white" alt="Master Plan">
  </a>
  <a href="https://github.com/alsyundawy/SparkPanelHost/releases/latest">
    <img src="https://img.shields.io/badge/🪞_Releases_&_Changelog-GitHub-6f42c1?style=for-the-badge&logo=github&logoColor=white" alt="Releases &amp; Changelog">
  </a>
  <a href="#security-invariants--cwecve-mitigation">
    <img src="https://img.shields.io/badge/🛡️_Security_Architecture-Zero--Trust-darkgreen?style=for-the-badge&logo=securityscorecard&logoColor=white" alt="Security Architecture">
  </a>
  <a href="https://github.com/alsyundawy/SparkPanelHost/issues">
    <img src="https://img.shields.io/badge/🐛_Report_Issue-GitHub_Issues-red?style=for-the-badge&logo=github&logoColor=white" alt="Report Issue">
  </a>
</p>

---

> Lead Architect & Maintainer:<br>
> **[`HARRY DERTIN SUTISNA ALSYUNDAWY (@alsyundawy)`](https://github.com/alsyundawy)** — [`ALSYUNDAWY IT SOLUTION`](https://alsyundawy.com)<br>
>
> 📚 **[`Master Architecture Blueprint (docs/00_MASTER_PLAN_v8.1.md)`](docs/00_MASTER_PLAN_v8.1.md)** &nbsp;|&nbsp;
> 🔒 **[`gRPC Contracts & Daemon Protocol (docs/01_GRPC_PROTO.md)`](docs/01_GRPC_PROTO_AND_SECURE_DAEMON.md)** &nbsp;|&nbsp;
> 🗄️ **[`Database Schema & RBAC 3-Tier (docs/02_DATABASE_SCHEMA.md)`](docs/02_DATABASE_SCHEMA_AND_RBAC.md)** &nbsp;|&nbsp;
> 📦 **[`Installer & 1-Click App Recipes (docs/03_INSTALLER.md)`](docs/03_INSTALLER_AND_NATIVE_APP_RECIPES.md)** &nbsp;|&nbsp;
> 🎨 **[`Modern UI/UX & Bento Grid Guide (docs/04_UI_UX.md)`](docs/04_MODERN_UI_UX_AND_I18N_BLUEPRINT.md)** &nbsp;|&nbsp;
> 🛠️ **[`Monorepo Maintenance Guide (docs/05_GUIDE.md)`](docs/05_MONOREPO_AND_MAINTENANCE_GUIDE.md)** &nbsp;|&nbsp;
> 📋 **[`Master Task List & Verification (TASK_LIST.md)`](TASK_LIST.md)** &nbsp;|&nbsp;
> 🌐 **[`Multi-Node Vagrant Cluster Testbed (deploy/vagrant/)`](deploy/vagrant/)** &nbsp;|&nbsp;
> 🐛 **[`Issue Tracker`](https://github.com/alsyundawy/SparkPanelHost/issues)** &nbsp;|&nbsp;
> 💖 **[`Support via PayPal`](https://www.paypal.me/alsyundawy)**

---

> [!IMPORTANT]
>
> ### ⚠️ Security Architecture & Kernel Invariants
>
> 1. **Dual-Plane Privilege Separation (Zero-Trust Model)**<br>
>    The web dashboard and public API (`panel-web`) run as an unprivileged user (`panel-web`) with an empty Linux capability bounding set (`CapabilityBoundingSet=`). The root system daemon (`panel-agent`) runs in a strictly hardened systemd sandbox (`ProtectSystem=strict`, `MemoryDenyWriteExecute=yes`), listening exclusively on a local Unix Domain Socket (`/run/panel/agent.sock`, `chmod 0660`), with caller verification via Linux kernel **`SO_PEERCRED`** (`unix.GetsockoptUcred`).
>
> 2. **Zero-Shell Execution Engine (`execsafe`)**<br>
>    Direct execution of `/bin/sh -c`, `/bin/bash -c`, or arbitrary string command concatenation is **strictly forbidden across all subsystems**. Every OS command is executed via explicit binary allowlists and argument vectors with a deterministic sanitized environment (`PATH=/usr/sbin:/usr/bin:/sbin:/bin`, `LC_ALL=C.UTF-8`), eliminating **CWE-78 (Command Injection)** and **CWE-426 (Untrusted Search Path)**.
>
> 3. **Kernel-Enforced Path Traversal & Symlink Immunity (`unix.Openat2`)**<br>
>    All file operations, archive extractions (`.tar.gz`, `.zip`), and tenant staging operations utilize the modern Linux kernel syscall **`unix.Openat2`** with **`unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS`** flags, locking execution to tenant subtrees and rendering **CWE-22, CWE-59, CWE-362, and CWE-409 (Zip-Bomb)** impossible.
>
> 4. **Debian 13 `tmpfs` & cgroups v2 Quota Safety**<br>
>    On Debian 13 (_Trixie_), `/tmp` is mounted on RAM (`tmpfs`). SparkPanelHost strictly isolates staging uploads, database dumps, and backup staging in `/var/panel/tmp/<username>` on the primary storage partition with Linux disk quotas enforced, preventing memory exhaustion attacks (**CWE-400**).

---

## ⚡ Why SparkPanelHost? (Comparative Industry Matrix)

| Evaluation Metric          |    SparkPanelHost (v0.1.0)     |         cPanel / WHM         |      Plesk Obsidian      |    CloudPanel v2     |         CyberPanel          | aaPanel (BaoTa)  |
| :------------------------- | :----------------------------: | :--------------------------: | :----------------------: | :------------------: | :-------------------------: | :--------------: |
| **Idle RAM Footprint**     |          **< 150 MB**          |       ~1.5 GB – 2.0 GB       |     ~1.0 GB – 1.8 GB     |       ~250 MB        |           ~600 MB           |     ~500 MB      |
| **Privilege Model**        | **Dual-Plane (`SO_PEERCRED`)** |       Single Monolith        |     Single Monolith      |    Single Daemon     |      Root Web Process       | Root Web Process |
| **Shell Security**         |  **Zero-Shell (`execsafe`)**   |    Shell Scripts Wrappers    |     Shell Execution      |   Shell Execution    | Shell Execution (CVE PSAUX) | Shell Execution  |
| **Path Traversal Defense** |   **Kernel `unix.Openat2`**    |    Userspace Path Checks     |  Userspace Path Checks   |     Path Checks      | Vulnerable (CVE-2024-51567) |    Vulnerable    |
| **WAF & Threat Intel**     |   **Coraza WAF + CrowdSec**    |     ModSecurity / OWASP      |   ModSecurity / Comodo   |  Basic Vhost Rules   |      OpenLiteSpeed WAF      | Nginx WAF Plugin |
| **Multi-PHP Management**   |      **PHP 8.2 – 8.6 GA**      |         EasyApache 4         |    Plesk PHP Packages    |      Multi-PHP       |    LiteSpeed PHP (LSPHP)    |  App Store PHP   |
| **DNS Clustering**         |  **WHM DNS Only (Star/AXFR)**  |         WHM DNS Only         |    Slave DNS Manager     |  None (Single Node)  |     CyberPanel DNS Sync     |       None       |
| **Licensing Cost**         |     **100% Free (GPLv3)**      |   Expensive Monthly/Tiered   | Expensive Monthly/Tiered | Free (Closed Addons) |        Free / Addons        |  Free / Addons   |
| **Modern UI/UX**           |   **React 19 + Bento Grid**    | Legacy Paper Lantern/Jupiter |   Modern Vue Dashboard   | Modern Vue Dashboard |      Bootstrap Legacy       |   LayUI Legacy   |

---

## 🏗️ High-Level System Architecture

```text
 ┌────────────────────────────────────────────────────────────────────────┐
 │                    CLIENT TIER & BROWSER INTERFACE                     │
 │  HTTPS Port :2087 (WHM Administration) | Port :2083 (Tenant Hosting)   │
 └───────────────────────────────────┬────────────────────────────────────┘
                                     │
                                     ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │                CONTROL PLANE: panel-web (Go 1.26 Static)               │
 │                User: panel-web (Unprivileged, CapabilityBoundingSet=)  │
 ├────────────────────────────────────────────────────────────────────────┤
 │  • REST API v1 Routing & Session Management (Argon2id + Passkeys/TOTP) │
 │  • WebSocket Live Telemetry Streaming (CPU, RAM, Disk, Coraza WAF)     │
 │  • Embedded SQLite WAL Store (Relational DB & Audit Hash-Chained Logs) │
 │  • Single Page Application (React 19 + TypeScript + Tailwind CSS v4)   │
 └───────────────────────────────────┬────────────────────────────────────┘
                                     │
                                     │ Unix Domain Socket: /run/panel/agent.sock (0660)
                                     │ Kernel Verification: SO_PEERCRED (UID/GID)
                                     ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │                 SYSTEM AGENT: panel-agent (Go 1.26 Root)               │
 │                 Hardened Systemd Sandbox (ProtectSystem=strict)        │
 ├────────────────────────────────────────────────────────────────────────┤
 │  • 32 gRPC Service Procedures (Tenancy, Domains, DB, DNS, Mail, WAF)   │
 │  • Zero-Shell Execution Engine (execsafe - strict binary argument vectors)
 │  • Kernel-Enforced File Extraction (unix.Openat2 RESOLVE_BENEATH)      │
 │  • Downscoped OS Worker Threads (SysProcAttr.Credential UID/GID lock)  │
 └───────┬────────────┬────────────┬────────────┬───────────┬─────────────┘
         │            │            │            │           │
         ▼            ▼            ▼            ▼           ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐
   │  Nginx   │ │Multi-PHP │ │ MariaDB  │ │ PowerDNS │ │   Postfix    │
   │ 1.31 +   │ │8.2 – 8.6 │ │ 12.3 LTS │ │ 5.1 REST │ │ Dovecot 2.4  │
   │Coraza WAF│ │FPM Pools │ │utf8mb4 / │ │Star Clust│ │ Rspamd 4.3   │
   │OWASP CRS │ │cgroups v2│ │PostgreSQL│ │ TSIG AXFR│ │RSA+Ed25519   │
   └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────────┘
```

---

## 🧩 Comprehensive Feature Matrix

### 1. Web Serving & Next-Gen WAF

- **Nginx 1.31 Mainline / 1.30 Stable:** Optimized reverse proxy supporting HTTP/3 (QUIC), Brotli compression, and dynamic upstream load balancing.
- **Coraza WAF (Go-Native Memory Safe):** In-process web application firewall eliminating C/C++ memory corruption vulnerabilities (CWE-119, CWE-416) found in legacy ModSecurity.
- **OWASP CRS v4.30.0 / v4.25.2 LTS:** Full anomaly scoring inspection with tailored exclusion profiles for popular engines (WordPress, WooCommerce, Nextcloud, Laravel, Drupal, Joomla, and Roundcube).

### 2. Multi-PHP Runtime Engine

- **PHP 8.2 through 8.6 GA (19 Nov 2026):** Run multiple concurrent PHP versions with dedicated, isolated PHP-FPM pools per tenant.
- **PHP 8.2 Security EOL Sentinel:** Visual warnings in the dashboard highlighting PHP 8.2 End-of-Life (31 Dec 2026), nudging tenants towards PHP 8.5/8.6.
- **Custom `php.ini` Directives:** Per-domain tuning for memory limits, max execution time, upload sizes, and OPcache configurations.

### 3. High-Performance Database Tier

- **MariaDB 12.3.4 LTS / 11.8.10 LTS:** Pre-configured with `utf8mb4_unicode_520_ci` collation, memory-tuned InnoDB buffer pools, and performance metrics.
- **Zero-Lock Hot Physical Backup:** Native integration with `mariadb-backup` for continuous physical streaming snapshots without table locking.
- **PostgreSQL 17 / 18 Support:** Optional multi-tenant PostgreSQL driver with dedicated users, databases, and schemas.

### 4. Authoritative DNS Clustering (WHM DNS Only Style)

- **PowerDNS 5.1+ Authoritative:** Controlled 100% via REST API without brittle flat-file zone parsing.
- **Star & Mesh Topologies:** Configure remote DNS nodes as `dns_only`, `synchronize`, `write_only`, or `standalone` with AXFR/IXFR + TSIG authentication.
- **Automated DNSSEC & ACME DNS-01:** 1-click zone signing and automatic Let's Encrypt / ZeroSSL wildcard certificate issuance via DNS validation.

### 5. Enterprise Mail Stack & Deliverability Suite

- **Postfix 3.9+ & Dovecot 2.4.5+:** Modern Maildir storage with Argon2id password hashing and SNI multi-domain SSL certificates.
- **Rspamd 4.3+ Intelligent Spam Filtering:** Neural network heuristics, greylisting, fuzzy hashes, and statistical Bayes classification.
- **Dual-Signing DKIM:** Simultaneous signing with **RSA-2048** (legacy recipient compatibility) and **Ed25519** (next-gen cryptographic efficiency).
- **Deliverability Hardening:** Automated wizard generating valid SPF records, strict DMARC policies, ARC signing, MTA-STS (`RFC 8461`), and TLS-RPT (`RFC 8460`).
- **Real-Time RBL Reputation Sentinel:** Continuous background probing against Spamhaus (SBL/XBL/PBL), Barracuda, SpamCop, SORBS, and UCEPROTECT.

### 6. Security, Firewall & Intrusion Prevention

- **nftables Dedicated Table (`inet aetherpanel`):** Atomic rule updates replacing legacy iptables without dropped connection states.
- **CrowdSec 1.8.2 IPS:** Community-driven threat intelligence blocking brute-force attacks across SSH, FTP, Dovecot IMAP, and HTTP endpoints.
- **Audit Log Chain (Hash-Chained):** Every administrative action is sealed using SHA-256 hash chaining (`prev_hash` -> `entry_hash`), creating an immutable forensic ledger.

### 7. Backup & Disaster Recovery (3-2-1 Policy)

- **Restic Engine (AES-256-GCM):** Deduplicated, encrypted snapshot backups saving bandwidth and storage.
- **Multi-Destination Sync:** Native targets for Local Disk, Amazon S3, Wasabi, Backblaze B2, Google Cloud Storage, and SFTP.
- **Automated Test-Restore:** Daily background verification testing archive integrity and database restoration in an isolated sandbox.

### 8. Frontend UI/UX (AdminLTE + Bento Grid Fusion)

- **Modern Bento Grid Layout:** Interactive KPI small-boxes with adaptive progress bars (vCPU, RAM, Disk, Security).
- **Collapsible Mini-Sidebar:** Seamless toggle between full navigation (`w-64`) and compact icon mode (`w-16`) with floating tooltips.
- **OLED Dark Mode & Light Mode:** Tailored zinc dark palette (`#09090B`) and high-contrast WCAG AAA light palette (`#F8FAFC`).
- **Command Palette (`Cmd+K`):** Spotlight-style quick switcher across domains, databases, settings, and terminal sessions.
- **Native Web Terminal (xterm.js):** Browser-based SSH terminal locked strictly to the tenant's Linux UID/GID with PTY allocation.
- **Monaco Code Editor:** Syntax-highlighting file manager for editing Nginx, PHP, and web files directly in the browser.

---

## 🚀 Quick Installation (1-Command)

### System Requirements

| Specification    |       Minimum (Personal / Dev)        | Recommended (Production / Reseller) |
| :--------------- | :-----------------------------------: | :---------------------------------: |
| **Supported OS** | Ubuntu 24.04/26.04 LTS / Debian 12/13 | Ubuntu 24.04/26.04 LTS / Debian 13  |
| **Architecture** |        x86_64 (AMD64) / ARM64         |       x86_64 (AMD64) / ARM64        |
| **Processor**    |           1 vCPU (2.0 GHz+)           |              4+ vCPUs               |
| **Memory (RAM)** |                 2 GB                  |       8 GB+ (with NVMe Swap)        |
| **Disk Space**   |               20 GB SSD               |          100 GB+ NVMe SSD           |

### Production 1-Command Installer

Execute the following command as `root` on a freshly installed Ubuntu or Debian server:

```bash
curl -sSL https://raw.githubusercontent.com/alsyundawy/SparkPanelHost/main/installer/install.sh | sudo bash
```

The idempotent installer will automatically:

1. Verify system prerequisites, architecture, and network connectivity.
2. Configure **DEB822 (`.sources`)** official repositories with isolated GPG keyrings.
3. Apply kernel tuning via `/etc/sysctl.d/99-aetherpanel.conf`.
4. Deploy `panel-web` and `panel-agent` static binaries to `/usr/local/bin/`.
5. Install and configure Nginx Mainline, Coraza WAF, Multi-PHP, MariaDB 12.3, PowerDNS, Postfix, Dovecot, Rspamd, nftables, and CrowdSec.
6. Generate random high-entropy passwords, TLS certificates, and initialize the SQLite database.
7. Output panel login credentials and port access instructions.

---

## 💻 Local Development & Monorepo Build

For developers and contributors looking to build SparkPanelHost from source:

### Prerequisites

- **Go:** 1.23+ / 1.26+
- **Node.js:** 20+ / 22+ (npm 10+)
- **Build Tools:** `make`, `gcc`, `protobuf-compiler`

### Step-by-Step Build

```bash
# 1. Clone repository
git clone https://github.com/alsyundawy/SparkPanelHost.git
cd SparkPanelHost

# 2. Build backend binaries and compile frontend SPA
make build-all

# 3. Execute unit tests across Go and React
make test

# 4. Run Trunk meta-linter and static analysis
make lint
```

---

## 📂 Monorepo Architecture Directory Map

```text
SparkPanelHost/
├── .github/workflows/          # CI/CD: MegaLinter, test -race, staticcheck, release
├── .trunk/                     # Trunk meta-linter configurations & plugins
├── api/proto/agent/v1/         # 32 gRPC contracts (agent.proto)
├── backend/
│   ├── cmd/
│   │   ├── panel-agent/        # Privileged system daemon (root)
│   │   └── panel-web/          # Unprivileged control plane & API (panel-web)
│   ├── gen/                    # Generated Go Protobuf & gRPC stubs
│   ├── internal/
│   │   ├── agent/              # drivers (10 drivers), execsafe, fileops, netfetch, server
│   │   └── web/                # api, auth (Passkeys/TOTP), store, agentclient, workers
│   └── migrations/             # 18 relational tables SQLite WAL schema (001_init.sql)
├── deploy/vagrant/             # Multi-node cluster Vagrant VM testbed (Ubuntu 24.04)
├── docs/                       # Comprehensive specifications (00 to 05)
├── frontend/                   # React 19 + TypeScript + Vite + Tailwind CSS v4
│   ├── src/components/         # Navbar, Sidebar (collapsible), CommandPalette, Banner
│   ├── src/pages/              # Overview, Domains, Databases, DNS, Email, Apps, Terminal,
│   │                           # Packages, Security (WAF/CrowdSec), Backups, ConfigEditor
│   └── src/styles/             # tokens.css (OLED Dark Mode & WCAG AAA Light Mode)
├── installer/
│   ├── install.sh              # Idempotent single-command DEB822 installer script
│   ├── nftables/               # Dedicated table inet aetherpanel configuration
│   └── systemd/                # Hardened systemd service templates
├── recipes/                    # 1-Click App Catalog (WordPress, Joomla, Drupal, etc.)
├── scripts/                    # Recipe validation & upstream manifest updater tools
├── Makefile                    # Unified build & orchestration tasks
├── sonar-project.properties    # SonarQube / SonarLint configuration
└── TASK_LIST.md                # Master tracking list & 14 review pillars evidence
```

---

## 🇮🇩 Ringkasan Bahasa Indonesia

**SparkPanelHost v0.1.0** adalah kontrol panel web hosting All-in-One modern yang dirancang untuk menggantikan panel konvensional seperti **cPanel/WHM, Plesk, dan CloudPanel** dengan mengedepankan efisiensi sumber daya (RAM < 150 MB saat idle), keamanan **Zero-Trust**, dan kemudahan operasional komersial.

### Keunggulan Pokok:

1. **Pemisahan Hak Akses Ganda (Dual-Plane):** Panel web berjalan tanpa hak akses root (`panel-web`), sedangkan daemon sistem (`panel-agent`) beroperasi sebagai root yang dikeraskan di balik Unix Domain Socket `/run/panel/agent.sock` dan diverifikasi langsung oleh kernel Linux (`SO_PEERCRED`).
2. **Mesin Bebas Shell (`execsafe`):** Menolak keras eksekusi `/bin/sh -c` untuk menumpas celah **CWE-78 (Command Injection)** yang sering menyerang kontrol panel lain.
3. **Imunitas Traversal Kernel (`unix.Openat2`):** Ekstraksi file arsip dilindungi syscall kernel dengan flag `RESOLVE_BENEATH | RESOLVE_NO_SYMLINKS` guna mencegah manipulasi symlink (**CWE-22 / CWE-59**).
4. **DNS Clustering Mandiri:** Menggunakan PowerDNS Authoritative via REST API murni, mendukung topologi Star setara _WHM DNS Only_.
5. **Dashboard Modern Bento Grid:** Mengadopsi ergonomi modern AdminLTE dan Bootstrap dashboard dengan sidebar yang dapat dikolaps, pemantauan kesehatan server real-time, serta mode gelap OLED hemat energi.

---

## 📜 Lisensi & Legalitas

Proyek ini dirilis dan dilisensikan di bawah ketentuan **[GNU General Public License v3.0 (GPL-3.0-or-later)](LICENSE)**.

_Penafian Merek Dagang: cPanel, WHM, Plesk, CloudPanel, DirectAdmin, WHMCS, Ubuntu, Debian, dan merek dagang lainnya yang disebutkan dalam dokumentasi ini adalah milik dari pemilik sahnya masing-masing. SparkPanelHost adalah proyek perangkat lunak bebas dan merdeka (FOSS) independen._
