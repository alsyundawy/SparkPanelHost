<!-- markdownlint-disable -->
# 05. Struktur Monorepo, Otomasi CI/CD & Panduan Maintenance (v8.1)

## 1. Struktur Monorepo Lengkap (`aetherpanel/`)

```text
aetherpanel/
├── .github/workflows/
│   ├── ci-security.yml               # SAST (gosec, govulncheck), buf lint, trivy & race tests
│   └── auto-update-manifests.yml     # Cron harian auto-update versi & SHA-256 app installer
├── api/proto/agent/v1/
│   ├── agent.proto                   # Kontrak gRPC v7 lengkap
│   ├── buf.yaml
│   └── buf.gen.yaml
├── backend/
│   ├── go.mod
│   ├── cmd/
│   │   ├── panel-agent/main.go       # Privileged Daemon (Unix Socket + SO_PEERCRED)
│   │   └── panel-web/main.go         # Unprivileged Control Plane API & Web Server
│   └── internal/
│       ├── agent/
│       │   ├── server/               # Socket auth (SO_PEERCRED) & audit interceptor
│       │   ├── execsafe/             # Zero-shell execve runner + cleanEnv (anti CWE-78)
│       │   ├── fileops/              # openat2(RESOLVE_BENEATH|RESOLVE_NO_SYMLINKS) extractor
│       │   ├── netfetch/             # SSRF-safe HTTPS downloader (anti CWE-918)
│       │   └── drivers/              # nginx, coraza, phpfpm, mariadb, postgres, pdns, mail, nftables, restic, cgroups
│       └── web/
│           ├── auth/                 # Argon2id, Passkeys/WebAuthn, TOTP 2FA, API Tokens, Impersonation
│           ├── rbac/                 # Isolasi Superadmin, Reseller & Tenant + Feature Lists
│           ├── store/                # Migrasi SQLite WAL / PostgreSQL 17+ & hash-chained audit logs
│           ├── appinstaller/         # Parser manifes YAML + SHA-256 + Staging + cPanel Importer
│           ├── telemetry/            # WebSocket live metrics, RBL monitor & WAF log streamer
│           └── terminal/             # Web SSH PTY bridge terkunci di UID tenant
├── frontend/
│   ├── package.json
│   └── src/
│       ├── hooks/useFeatureGate.ts   # Penegakan Feature Lists & peringatan PHP EOL
│       ├── i18n/locales/             # Dual bahasa (id.json & en.json)
│       ├── styles/tokens.css         # OLED Dark Mode & Light Mode WCAG AAA tokens
│       ├── components/               # BentoGrid, CommandPalette, Charts, Monaco, xterm.js
│       └── pages/                    # Admin (WHM), Reseller & Tenant (cPanel) views
├── recipes/
│   ├── schema/recipe.schema.json     # JSON Schema wajib untuk seluruh resep
│   ├── cms/                          # wordpress.yaml, joomla.yaml, drupal.yaml
│   ├── forum/                        # flarum.yaml, phpbb.yaml
│   ├── marketplace/                  # prestashop.yaml, opencart.yaml
│   └── webmail/                      # roundcube.yaml
├── installer/
│   ├── install.sh                    # Installer DEB822 Ubuntu 26.04/24.04 & Debian 13/12
│   ├── nftables/aetherpanel.nft      # Aturan firewall atomik inet aetherpanel
│   ├── sysctl/99-aetherpanel.conf    # Hardening & performa kernel Linux
│   └── systemd/                      # panel-agent.service & panel-web.service
├── scripts/
│   ├── update_manifests.py           # Skrip pengecek rilis upstream & penghitung SHA-256
│   └── verify_recipes.py             # Validator JSON Schema & integritas resep
└── Makefile
```

---

## 2. Empat Pilar Pemeliharaan Jangka Panjang

1. **Pembaruan Manifes App Installer Otomatis (Harian via GitHub Actions):**
   `.github/workflows/auto-update-manifests.yml` menjalankan `scripts/update_manifests.py` setiap hari pukul `02:00 UTC`, mengambil rilis stabil terbaru dari upstream resmi dengan *host allowlist*, menghitung digest **SHA-256**, memvalidasi terhadap `recipe.schema.json`, dan membuat Pull Request otomatis.
2. **Siklus Hidup OS & Multi-PHP (6 Bulanan – Tahunan):**
   Penambahan versi PHP baru setiap rilis November ke driver `phpfpm` dan penandaan versi EOL (seperti **PHP 8.2 pada 31 Desember 2026**) melalui `useFeatureGate.ts`.
3. **Tuning Profil OWASP CRS v4 & Coraza WAF:**
   Pembaruan aturan OWASP CRS v4 serta pengujian pengecualian *false-positive* untuk CMS/E-Commerce modern (WordPress Gutenberg, PrestaShop, Flarum).
4. **Gate Keamanan CI/CD Berkelanjutan:**
   Setiap Pull Request wajib lolos `govulncheck`, `gosec`, `buf lint`, `trivy fs`, `go test -race`, dan `verify_recipes.py`.
