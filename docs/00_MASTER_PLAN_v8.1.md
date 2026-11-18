<!-- markdownlint-disable -->
# AetherPanel — Blueprint Arsitektur All-in-One Linux Hosting Control Panel (v8.1)
**Target OS:** Ubuntu 26.04 LTS (Resolute) / 24.04 LTS (Noble) & Debian 13 (*Trixie*) / 12 (*Bookworm*)  
**Status:** Production-Ready Master Blueprint (Personal, Reseller & Commercial Hosting)  
**Tanggal:** 18 November 2026

---

## 1. Sintesis Deep Research & Stack Infrastruktur Server (v8.1)

Seluruh komponen layanan sistem menggunakan repositori paket resmi dengan format **DEB822 (`.sources`)** dan GPG keyring terisolasi (`Signed-By`), dipadukan dengan dua biner Go statis (`panel-web` dan `panel-agent`) agar waktu instalasi bersih di bawah **5 menit**, hemat RAM pada VPS 2 GB sekalipun, dan menerima patch CVE sistem operasi secara otomatis via `apt upgrade` / `unattended-upgrades`.

| Komponen | Target Stack v8.1 | Sumber Repositori & Keputusan Arsitektur |
| :--- | :--- | :--- |
| **Sistem Operasi & Kernel** | Ubuntu 26.04 LTS / 24.04 LTS<br>Debian 13 (Trixie) / 12 (Bookworm) | **Mitigasi `tmpfs` `/tmp` Debian 13:** seluruh upload staging Nginx, kompresi File Manager, dan backup diarahkan ke `/var/panel/tmp/<username>` di partisi disk utama berkuota. Kernel dikeraskan via `/etc/sysctl.d/99-aetherpanel.conf` (TCP BBR, `fs.protected_symlinks=1`, `yama.ptrace_scope=2`). |
| **Web Server & WAF** | Nginx 1.31.8+ (Mainline) / 1.30.6+ (Stable)<br>**Coraza WAF** (Go-native) + **OWASP CRS 4.30.0 / 4.25.2 LTS** | Template vhost terisolasi per domain + linter konfigurasi otomatis anti-CVE Nginx Oktober 2026 (`alias`+`rewrite`, map regex captures, SSI+AIO, HTTP/3 body injection, TLS session ticket). **Coraza** (memory-safe) + CRS v4 dengan profil pengecualian per-CMS. |
| **Multi-PHP Runtime** | PHP 8.2 (*EOL 31 Des 2026*), 8.3, 8.4, **8.5 (Default)**, **8.6 (GA 19 Nov 2026)** | `packages.sury.org/php` (Debian) & `ppa:ondrej/php` (Ubuntu) dalam format DEB822. Pool PHP-FPM terpisah per tenant, `open_basedir`, `disable_functions`, secure session defaults (PHP 8.6), serta peringatan migrasi otomatis untuk PHP 8.2. |
| **Database Engine** | **MariaDB 12.3.4 LTS** (preferred) / 11.8.10 LTS<br>+ Opsional **PostgreSQL 17/18** | Repositori resmi MariaDB (`mirror.mariadb.org`). Autentikasi `unix_socket` + `PARSEC`, `local_infile=OFF`, backup fisik konsisten tanpa lock via `mariadb-backup`. Dukungan PostgreSQL opsional. |
| **DNS Server & Clustering** | **PowerDNS Authoritative 5.1.4+**<br>+ **DNS Cluster (WHM DNS Only style)** | Repositori `repo.powerdns.com` (backend MariaDB/gmysql). Dikelola 100% melalui PowerDNS REST API. Mendukung DNSSEC 1-click, AXFR/IXFR + TSIG, serta **DNS Clustering** (lihat Bagian 2). |
| **Enterprise Mail Stack** | Postfix + **Dovecot 2.4.5+** + **Rspamd 4.3.0+** + **MTA-STS** + Roundcube | Maildir terisolasi. Rspamd 4.3.0+ (fix worker crash, UAF milter, DMARC/ARC bypass). Dual-signing DKIM RSA-2048 + Ed25519, ARC, DMARC, outbound rate-limit. Otomasi SPF/DKIM/DMARC/MTA-STS + **RBL Blacklist Monitor**. |
| **Firewall & IDS/IPS** | **nftables** (`table inet aetherpanel`) + **CrowdSec 1.8.2+** | Reload aturan atomik. Integrasi `crowdsec-firewall-bouncer-nftables` + AppSec/WAF bouncer (Coraza-based) + GeoIP. Opsi CSF fork GPLv3. |
| **Backup & DR Engine** | **Restic** + `rclone` (S3/B2/Wasabi/SFTP) + `mariadb-backup` | Snapshot inkremental terenkripsi AES-256, kebijakan retensi 3-2-1, Automated Test-Restore. |
| **Core Control Panel** | Go (`panel-agent` & `panel-web`) + React 19 TypeScript | Static binary (`CGO_ENABLED=0`). Pemisahan mutlak Web Control Plane non-root dan Privileged System Agent via gRPC Unix Socket (`SO_PEERCRED`). |

---

## 2. DNS Clustering — Arsitektur WHM DNS Only Style (v8.1)

AetherPanel mendukung **DNS Cluster** untuk memisahkan authoritative nameserver dari web/mail nodes, mirip cPanel DNSOnly / WHM DNS Cluster. Tujuan: DNS tetap online meskipun web node down, distribusi query, dan isolasi beban.

### Topologi yang Didukung

1. **Star Topology (Recommended)**  
   - Web/Hosting nodes berstatus **Write-Only**.  
   - Perubahan zone (create/update/delete) di-push ke DNS-Only nodes via API.  
   - DNS-Only nodes **tidak** menerima perubahan balik ke web nodes.  
   - Minimal 2 DNS-Only nodes (ns1 + ns2).

2. **Primary / Secondary (AXFR/IXFR + TSIG)**  
   - PowerDNS native primary/secondary dengan TSIG.  
   - Cocok untuk cluster kecil atau hybrid dengan BIND eksternal.

3. **Database-Level Replication (MariaDB / Galera)**  
   - Semua PowerDNS node membaca dari MariaDB yang direplikasi (async replica atau Galera multi-master).  
   - Tulisan diarahkan ke primary (atau Galera node) melalui panel-agent.  
   - Lebih cepat untuk volume zone tinggi, tanpa per-zone NOTIFY plumbing.

### Peran Node (DNS Role)

| Role | Perilaku | Contoh Penggunaan |
| :--- | :--- | :--- |
| **Write-Only** | Push zone ke remote, tidak menerima update balik | Web/Hosting node utama |
| **Synchronize** | Two-way sync (dengan reverse trust) | DNS-Only node yang juga bisa diedit |
| **Standalone** | Hanya menjawab query, tidak sync | Edge anycast node read-only |
| **DNS-Only** | Mode instalasi minimal: hanya PowerDNS + agent ringan | Dedicated nameserver |

### Mekanisme Keamanan & Trust

- Setiap node memiliki **API Token** scoped (`dns:sync`, `dns:read`) dengan optional IP allowlist.  
- Reverse Trust relationship wajib untuk mode Synchronize.  
- Komunikasi antar node via HTTPS + mTLS atau token + IP restriction.  
- DNSSEC signing dilakukan di primary; key material disinkronkan secara aman.  
- Validasi zone sebelum commit (`pdnsutil check-zone` / API equivalent).

### Skema Database Tambahan (lihat `02_DATABASE_SCHEMA_AND_RBAC.md`)

- `dns_clusters` — definisi cluster  
- `dns_cluster_nodes` — membership + role + token reference  
- `dns_zone_sync_status` — last serial, last sync, error

### RPC Baru (lihat `agent.proto`)

- `RegisterDNSClusterNode`  
- `SyncDNSZoneToCluster`  
- `GetDNSClusterStatus`  
- `PromoteDNSPrimary`

---

## 3. Arsitektur Keamanan Zero-Trust (v8.1)

```text
┌──────────────────────────────────────────────────────────────────────────┐
│  FRONTEND (SPA): React 19 + TypeScript + Tailwind v4 + Shadcn UI         │
│  (Admin WHM-View, Reseller View & Tenant cPanel-View | id/en | OLED Dark)│
└────────────────────────────────────┬─────────────────────────────────────┘
                                     │ HTTPS / REST + WebSocket (Stats/PTY)
┌────────────────────────────────────▼─────────────────────────────────────┐
│  CONTROL PLANE API (Unprivileged User: `panel-web` | Zero Capabilities)  │
│  Go (Chi) — Berjalan dalam sandbox systemd ketat tanpa akses root!       │
│  • Auth: Argon2id, Passkeys/WebAuthn, TOTP 2FA, RBAC + Feature Lists     │
│  • Billing & Automation API: Scoped API Tokens (WHMCS / Blesta ready)    │
│  • State DB: SQLite WAL (Single-Node) / PostgreSQL 17+ (Multi-Node)      │
└────────────────────────────────────┬─────────────────────────────────────┘
                                     │ gRPC / Protobuf via Unix Domain Socket
                                     │ (`/run/panel/agent.sock` chmod 0660)
┌────────────────────────────────────▼─────────────────────────────────────┐
│  PRIVILEGED SYSTEM AGENT (User: `root` + Hardened Systemd Unit)          │
│  Go Static Binary — ZERO SHELL (`execve` allowlist + sanitized `Env`)    │
│  • SO_PEERCRED Caller UID Check + Hash-Chained Immutable Audit Log       │
│  • Unprivileged Worker Spawn + `openat2(RESOLVE_BENEATH|RESOLVE_NO_SYMLINKS)`
│  • cgroups v2 (`CPUQuota`, `MemoryMax`, `TasksMax`) + Linux Disk Quota   │
└────┬───────────────┬───────────────┬───────────────┬─────────────────────┘
     │               │               │               │
     ▼               ▼               ▼               ▼
┌─────────┐    ┌───────────┐   ┌───────────┐   ┌────────────┐
│  Nginx  │    │  PHP-FPM  │   │ MariaDB   │   │ PowerDNS   │
│ +Coraza │    │  8.2–8.6  │   │ + Mail    │   │ + CrowdSec │
└─────────┘    └───────────┘   └───────────┘   └────────────┘
```

---

## 3. Fitur Utama Control Panel (v8.1 — Personal, Reseller & Commercial)

### A. Admin Control Plane (WHM-Equivalent — Port `2087`)
- **Telemetri & Health Bento Grid:** Streaming metrik real-time CPU, RAM, Disk IOPS, Network throughput, antrian Postfix, dan serangan WAF terblokir via WebSocket.
- **Manajemen Akun, Paket & Feature Lists:** Kontrol kuota disk/inode, bandwidth, `cgroups v2` (`CPUQuota`, `MemoryMax`, `IOWeight`, `TasksMax`), Suspend/Unsuspend instan, **Feature Lists** (mengaktifkan/menyembunyikan modul per paket), serta **"Login as User" (Impersonation)** dengan token sementara 30 menit yang tercatat di *immutable audit log*.
- **Sistem Reseller Lengkap:** Alokasi sumber daya agregat reseller (`max_accounts`, `total_disk_mb`, `total_bandwidth_mb`), sakelar kontrol *overselling*, private nameserver (`ns1`/`ns2`), dan kustomisasi logo/nama brand.
- **Billing & Provisioning API:** Manajemen `api_tokens` dengan granular *scopes* dan pembatasan IP untuk integrasi modul WHMCS, Blesta, HostBill, atau skrip internal.
- **Matrix Layanan & Editor Konfigurasi Aman:** Pengaturan PHP 8.2–8.6 global, editor `nginx.conf` / `php.ini` dengan validasi sintaks otomatis (`nginx -t`, `php-fpm -t`) + *1-click rollback*.
- **Security Advisor, Coraza WAF & RBL Monitor:** Log OWASP CRS v4 real-time, integrasi CrowdSec 1.8 + `nftables`, audit hardening OS, serta pengecekan berkala IP server terhadap daftar RBL/DNSBL email global.
- **Backup & Disaster Recovery Center:** Jadwal backup Restic terenkripsi (Lokal, S3, Backblaze B2, SFTP), pemantauan status 3-2-1, dan *automated test-restore*.

### B. Tenant Hosting Panel (cPanel-Equivalent — Port `2083`)
- **Domain, Multi-PHP & SSL Otomatis:** Manajemen domain utama, addon domain, subdomain, alias; *1-click PHP switcher* (8.2–8.6); SSL Let's Encrypt (`HTTP-01` & Wildcard `DNS-01` dengan dukungan ACME ARI); HSTS, Brotli, & FastCGI Cache toggle.
- **Native 1-Click App Installer, Staging & Importer:** Instalasi instan WordPress, Joomla, Drupal, Grav, Flarum, phpBB, MyBB, SMF, WooCommerce, PrestaShop, OpenCart, Bagisto dengan verifikasi SHA-256 wajib. Dilengkapi **1-Click Staging & Push-to-Live** (sinkronisasi file + search-replace URL database) serta **Account Importer** dari arsip backup cPanel (`cpmove`) dan DirectAdmin.
- **Enterprise Email Hosting & Deliverability Wizard:** Mailbox, alias, forwarder, auto-responder Sieve, kuota per kotak surat, outbound rate-limit, Roundcube 1-Click SSO, serta panduan DNS SPF, DKIM (RSA + Ed25519), DMARC, dan MTA-STS.
- **File Manager, Web SSH Terminal, Database & DNS:** File Manager dengan **Monaco Editor** & ekstraksi arsip aman; **Web Terminal (`xterm.js`)** terkunci pada UID tenant; Database Manager (MariaDB 12.3 & PostgreSQL 17) + phpMyAdmin/Adminer SSO; DNS Zone Editor + *1-click DNSSEC*.
- **Modern App Runtimes & DevOps:** Manajemen aplikasi **Node.js** dan **Python (`venv` + `gunicorn`/`uvicorn`)** sebagai systemd user service terisolasi, **Git Deployment** otomatis via webhook, Visual Cron Job Manager, dan akun FTP/SFTP.

---

## 4. Matriks Pertahanan CWE & CVE (v8.1)

| Kode CWE / CVE | Ancaman & Pelajaran Industri (2024–2026) | Mekanisme Pertahanan di AetherPanel v8.1 |
| :--- | :--- | :--- |
| **CWE-78 / CWE-426** | OS Command & Environment Injection (penyebab utama RCE pada panel hosting lama) | gRPC Unix Socket + `exec.CommandContext` allowlist biner + `cmd.Env` bersih tanpa shell. |
| **CWE-22 / CWE-59 / CWE-362** | Path Traversal, Symlink Race Condition (TOCTOU) & Zip-Slip saat ekstraksi/backup | Pekerja unprivileged UID/GID tenant + **`unix.Openat2(RESOLVE_BENEATH \| RESOLVE_NO_SYMLINKS)`** pada setiap komponen path. |
| **CWE-409** | Zip Bomb / Decompression Exhaustion pada upload & App Installer | Pembatasan ukuran stream via `io.LimitReader` pada setiap entri arsip maupun total ekstraksi. |
| **CWE-494 / CWE-918** | Unduhan kode tidak terverifikasi & SSRF ke jaringan internal / cloud metadata | Allowlist domain HTTPS upstream + pemblokiran IP privat/link-local pada socket dialer + wajib SHA-256. |
| **CWE-284 / CWE-639** | Broken Access Control / IDOR antar tenant maupun reseller | Middleware `tenant_id` & `reseller_id` + verifikasi `feature_lists` sebelum RPC dikirim ke agent. |
| **CWE-352 / CWE-79** | CSRF & Stored XSS melalui log WAF/akses | Cookie `HttpOnly; Secure; SameSite=Strict` + header `X-CSRF-Token` + Strict CSP + sanitasi output log. |
| **CWE-307 / CWE-400** | Brute-force login & Fork-bomb / OOM `/tmp` | Rate-limit + WebAuthn Passkeys/TOTP + CrowdSec 1.8 + `cgroups v2` (`TasksMax`/`MemoryMax`) + `/var/panel/tmp` disk staging. |
| **Nginx CVE Hardening** | Alias traversal, regex `map` capture overflow, HTTP/3 header/body abuse | Linter konfigurasi Nginx terintegrasi sebelum `nginx -t` dan reload. |

---

## 5. Roadmap Implementasi 16 Minggu (v8.1)

- **Minggu 1–4 (Core Agent, `openat2` FileOps & Hardened Installer):**
  Implementasi kontrak `agent.proto` v7, daemon Go `panel-agent` (`SO_PEERCRED`, `execsafe`, `openat2` extractor `.tar.gz`/`.zip`, `cgroups v2` slice), serta pengujian otomatis `install.sh` (DEB822) pada VM Ubuntu 26.04 LTS & Debian 13.
- **Minggu 5–8 (Control Plane API, Reseller, Feature Lists & Restic Backup):**
  Autentikasi Argon2id + Passkeys/WebAuthn + TOTP + API Tokens, isolasi RBAC Superadmin/Reseller/Tenant, Virtual Host Nginx + Coraza WAF + Multi-PHP 8.2–8.6 + ACME ARI SSL, serta mesin backup Restic 3-2-1.
- **Minggu 9–12 (PowerDNS 5.1, Enterprise Mail Stack & CrowdSec Firewall):**
  Integrasi PowerDNS REST API + 1-click DNSSEC, orkestrasi Postfix + Dovecot 2.4 + Rspamd 4.3 + MTA-STS + monitor RBL, serta dashboard firewall `nftables` + CrowdSec 1.8.
- **Minggu 13–16 (App Installer, Staging, Importer, Modern Runtimes & Audit Keamanan):**
  Mesin resep SHA-256 + Staging Push-to-Live + Importer cPanel/DirectAdmin, File Manager (Monaco) + Web SSH (`xterm.js`), Node.js/Python/Git Deploy, serta audit penetrasi & fuzzing sebelum rilis GA.
