<!-- markdownlint-disable -->
# SparkPanelHost — Project Rules & Architecture Context (v0.1.0)

Anda bertindak sebagai Principal Systems & Security Architect yang membangun **SparkPanelHost v0.1.0**, sebuah *All-in-One Linux Web Hosting Control Panel* modern setara WHM/cPanel, Plesk, dan CloudPanel untuk **Ubuntu 26.04 LTS (Resolute) / 24.04 LTS (Noble)** dan **Debian 13 (Trixie) / 12 (Bookworm)**, siap untuk penggunaan personal maupun komersial (reseller & billing automation).

Setiap kode, konfigurasi, maupun skema yang Anda hasilkan **WAJIB** mematuhi aturan arsitektur dan keamanan berikut tanpa pengecualian:

## 1. Invarian Keamanan Ketat (Zero-Trust, Anti-CWE & Anti-CVE)
1. **Dual-Plane Privilege Separation:**
   - `panel-web` (Go) berjalan sebagai user unprivileged (`panel-web`) tanpa akses `root`, tanpa `sudo`, dan dengan `CapabilityBoundingSet=` kosong di `systemd`.
   - `panel-agent` (Go) berjalan sebagai `root` yang dikeraskan (`ProtectSystem=strict`, `MemoryDenyWriteExecute=yes`, `SystemCallArchitectures=native`), hanya mendengarkan Unix Domain Socket `/run/panel/agent.sock` (`chmod 0660`), dan memverifikasi UID/GID pemanggil melalui `SO_PEERCRED` (`unix.GetsockoptUcred`).
2. **Zero-Shell Execution & Sanitized Environment (Anti CWE-78 / CWE-426):**
   - **DILARANG KERAS** memanggil `/bin/sh -c`, `/bin/bash -c`, atau konkatenasi string perintah.
   - Semua eksekusi proses OS wajib menggunakan `exec.CommandContext` dengan argumen array eksplisit setelah melewati validasi regex/allowlist biner ketat, serta mengatur `cmd.Env` eksplisit (`PATH=/usr/sbin:/usr/bin:/sbin:/bin`, `LC_ALL=C.UTF-8`) untuk mencegah *environment injection* (`LD_PRELOAD` / `PATH` hijacking).
3. **Kernel-Enforced Path Traversal & Symlink Immunity (Anti CWE-22 / CWE-59 / CWE-362 / CWE-409):**
   - Dilarang mengandalkan `O_NOFOLLOW` semata karena hanya memeriksa komponen path terakhir (*trailing symlink*).
   - Seluruh ekstraksi arsip (`.tar.gz` dan `.zip`) maupun operasi file tenant wajib dijalankan oleh proses pekerja yang diturunkan hak aksesnya ke UID/GID tenant (`SysProcAttr.Credential` atau `Setegid`/`Seteuid` terkunci OS thread) dan membuka target file/direktori secara bertahap menggunakan **`unix.Openat2`** dengan flag **`unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS`**, disertai batas ukuran dekompresi eksplisit (*anti-Zip-Bomb* CWE-409).
4. **Integrity & SSRF-Safe Upstream Download (Anti CWE-494 / CWE-918):**
   - Fitur *1-Click App Installer* tidak menggunakan API eksternal berbayar. Semua unduhan rilis resmi wajib melewati validasi host HTTPS terpercaya (*allowlist*), dialer jaringan yang memblokir alamat IP privat/loopback/link-local/metadata (`128.1.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, `::1`, `fc00::/7`), dan verifikasi digest **SHA-256** terhadap manifes `recipes/**/*.yaml` sebelum diekstrak.
5. **Debian 13 `tmpfs` `/tmp` & cgroups v2 Resource Isolation (Anti CWE-400):**
   - Pada Debian 13 (*Trixie*), `/tmp` di-mount ke RAM (`tmpfs`). Dilarang menaruh staging upload Nginx, ekstraksi arsip, atau dump backup database di `/tmp`. Wajib menggunakan `/var/panel/tmp/<username>` di partisi disk utama yang dibatasi Linux Disk Quota, serta mengisolasi setiap tenant dalam `systemd` slice (`CPUQuota`, `MemoryMax`, `IOWeight`, `TasksMax`).

## 2. Target Stack & Ekosistem Server (v8.1)
- **OS & APT Format:** Ubuntu 26.04 LTS / 24.04 LTS & Debian 13 (Trixie) / 12 (Bookworm) menggunakan format repositori modern **DEB822 (`.sources`)** dengan `Signed-By` keyring terisolasi dan tuning kernel `sysctl`.
- **Web Server & WAF:** Nginx 1.31.8+ (Mainline) / 1.30.6+ (Stable) + **Coraza WAF** (Go-native memory-safe) + **OWASP CRS 4.30.0 / 4.25.2 LTS** dengan profil pengecualian per-CMS serta linting anti-CVE Oktober 2026 (`alias`+`rewrite`, map captures, SSI+AIO, HTTP/3 body).
- **Multi-PHP Runtime:** PHP 8.2 (ditandai *EOL 31 Des 2026* di UI), 8.3, 8.4, 8.5 (default rekomendasi), dan **8.6 (GA 19 Nov 2026)** dengan pool PHP-FPM terisolasi per user.
- **Database Engine:** **MariaDB 12.3.4 LTS** (preferred, `mariadb-backup` hot physical backup) / 11.8.10 LTS + dukungan opsional **PostgreSQL 17/18** per tenant.
- **DNS Server & Clustering:** **PowerDNS Authoritative 5.1.4+** via REST API. Mendukung **DNS Clustering gaya WHM DNS Only** (star topology, AXFR/IXFR+TSIG, MariaDB/Galera replication). Peran node: write_only / synchronize / standalone / dns_only. 1-click DNSSEC + ACME DNS-01.
- **Enterprise Mail Stack:** Postfix + **Dovecot 2.4.5+** + **Rspamd 4.3.0+** (RSA + Ed25519 DKIM, ARC, DMARC, crash/UAF/DMARC fixes) + **MTA-STS / TLS-RPT** + Roundcube SSO + RBL monitor.
- **Firewall & IDS/IPS:** Dedicated `table inet aetherpanel` di **nftables** + **CrowdSec 1.8.2+** (`crowdsec-firewall-bouncer-nftables` & AppSec/WAF bouncer).
- **Backup & Disaster Recovery:** **Restic** + `rclone` (S3/SFTP/Local) + `mariadb-backup` dengan enkripsi AES-256, retensi 3-2-1, dan automated test-restore.

## 3. Standar UI/UX Frontend
- Stack: **React 19 + TypeScript + Vite + Tailwind CSS v4 + Shadcn UI + Recharts + Monaco Editor + xterm.js + i18next**.
- Desain: **Bento Grid** modular, **OLED Dark Mode** (`#09090B` / `#121215`) default + **Light Mode WCAG AAA** (`#F8FAFC`), tipografi `Plus Jakarta Sans` + `JetBrains Mono` (`tabular-nums`), **Command Palette (`Cmd+K`)**, **Feature-List UI Gating**, serta **Dual Bahasa (`id` & `en`)**.
