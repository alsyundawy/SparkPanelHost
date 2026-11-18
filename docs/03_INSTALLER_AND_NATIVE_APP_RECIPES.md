<!-- markdownlint-disable -->

# 03. Arsitektur Installer Server & Native App Recipes (v8.1)

## 1. Catatan Penting Ekosistem OS (Ubuntu 26.04/24.04 & Debian 13/12)

1. **Debian 13 (_Trixie_) `tmpfs` `/tmp` & Pencegahan OOM:**
   Debian 13 me-mount `/tmp` ke RAM (`tmpfs`) secara bawaan. Jika pengguna mengunggah arsip besar via Nginx atau mengekstrak backup di `/tmp`, sistem berisiko mengalami _Out-Of-Memory (OOM)_.
   - **Solusi v8.1:** `install.sh` membuat `/var/panel/tmp` di partisi disk utama, menulis konfigurasi `/etc/nginx/conf.d/00-aetherpanel-temp.conf` (`client_body_temp_path /var/panel/tmp/nginx; proxy_temp_path /var/panel/tmp/nginx_proxy;`), dan daemon mengarahkan ekstraksi sementara ke `/var/panel/tmp/<username>` di bawah kuota disk tenant.
2. **Format Repositori Modern DEB822 (`.sources`):**
   Baik Debian 13 (_Trixie_) maupun Ubuntu 24.04/26.04 mengadopsi format `.sources` dengan keyring terisolasi `/usr/share/keyrings/*.gpg` (`Signed-By`), menghilangkan ketergantungan pada `apt-key` yang sudah dihapus.
3. **Siklus Hidup Multi-PHP (8.2 – 8.6):**
   - **PHP 8.2** mencapai _Security End-of-Life_ pada **31 Desember 2026**. Panel menandainya dengan peringatan migrasi `Legacy / EOL` pada UI serta menyediakan alat uji kompatibilitas sebelum beralih ke **PHP 8.5** (default rekomendasi) atau **PHP 8.6**.
4. **Coraza WAF + OWASP CRS v4 & CrowdSec 1.8 + nftables:**
   - Menggunakan **Coraza WAF** (Go-native, memory-safe) yang terhubung ke Nginx melalui SPOE (`coraza-spoa`), dipadukan dengan **OWASP CRS v4** dan **CrowdSec 1.8 + `crowdsec-firewall-bouncer-nftables`**.

---

## 2. Alur Eksekusi `installer/install.sh` (v8.1)

1. **Preflight Check:** Memverifikasi OS (`Ubuntu 24.04/26.04` atau `Debian 12/13`), arsitektur (`x86_64` / `aarch64`), hak `root`, dan minimal 2 GB RAM.
2. **Kernel Hardening (`sysctl`) & Disk Quota:** Memasang `/etc/sysctl.d/99-aetherpanel.conf` (TCP BBR, `fs.protected_symlinks=1`, `fs.protected_hardlinks=1`, `kernel.yama.ptrace_scope=2`), membuat user `panel-web` dan `vmail` (`UID 5000`), serta menyiapkan `/var/panel/tmp`.
3. **Repositori Upstream Resmi (DEB822 `.sources` + GPG Keyrings):**
   - **Nginx** resmi (`nginx.org/packages/mainline`) + modul SPOE & **Coraza WAF + OWASP CRS v4**.
   - **Multi-PHP 8.2 – 8.6** via `packages.sury.org/php` (Debian) / `ppa:ondrej/php` (Ubuntu).
   - **MariaDB 12.3 LTS** (`mirror.mariadb.org`) dengan hardening awal (`local_infile=OFF`, `symbolic-links=0`).
   - **PowerDNS Authoritative 5.1** (`repo.powerdns.com`) + `pdns-backend-mysql`, lalu mengonfigurasi REST API lokal pada `127.0.0.1:8081` dengan API key acak kriptografis di `/etc/aetherpanel/secrets.env` (`0600`).
   - **Enterprise Mail Stack:** Postfix, **Dovecot 2.4**, **Rspamd 4.3** (`rspamd.com/apt-stable`), serta paket **Roundcube** terverifikasi SHA-256.
   - **Firewall & IDS/IPS:** Memasang aturan atomik `/etc/nftables.d/aetherpanel.nft` (`table inet aetherpanel`) + **CrowdSec 1.8** (`crowdsec` & `crowdsec-firewall-bouncer-nftables`).
   - **Backup Engine:** **Restic** & **rclone** siap pakai.

---

## 3. Mesin Resep App Installer, Staging & Importer (v8.1)

| Kategori              | Daftar Aplikasi                               | Alur Eksekusi Aman (`openat2` + SHA-256)                                                                                                                                                                                                            |
| :-------------------- | :-------------------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **CMS**               | WordPress, Joomla, Drupal, Grav               | Unduh via dialer anti-SSRF -> Verifikasi SHA-256 -> Ekstrak via pekerja unprivileged (`openat2`) -> Buat DB MariaDB 12.3 otomatis -> Render konfigurasi dengan salt acak -> Aktifkan profil pengecualian Coraza WAF CMS.                            |
| **Forum**             | Flarum, phpBB, MyBB, SMF                      | Unduh bundle rilis resmi -> Verifikasi SHA-256 -> Ekstrak via `openat2` -> Atur izin `storage/` & `assets/` -> Jalankan migrasi skema otomatis.                                                                                                     |
| **Marketplace**       | WooCommerce, PrestaShop, OpenCart, Bagisto    | Unduh arsip `.tar.gz` / `.zip` resmi -> Verifikasi SHA-256 -> Ekstrak via `openat2` (anti-Zip-Bomb) -> Provisioning DB `utf8mb4` -> Pasang Cron Job indeks & cache otomatis.                                                                        |
| **Staging & Migrasi** | 1-Click Staging & cPanel/DirectAdmin Importer | Kloning situs produksi ke subdomain staging terisolasi + _Push-to-Live_ (file + DB URL search-replace). Importer memvalidasi arsip `cpmove-*.tar.gz` atau backup DirectAdmin menggunakan `openat2` sebelum memetakan domain, database, dan mailbox. |
