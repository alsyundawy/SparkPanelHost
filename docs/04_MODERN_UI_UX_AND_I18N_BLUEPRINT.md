<!-- markdownlint-disable -->
# 04. Sistem Desain UI/UX Modern (v8.1), Dual Bahasa & Dark Mode

## 1. Prinsip Desain & Aksesibilitas (WCAG AAA)

1. **Modular Bento Grid Layout:** Metrik server real-time, status layanan (Nginx, Coraza WAF, PHP-FPM 8.2–8.6, MariaDB 12.3, PowerDNS 5.1, Postfix/Dovecot 2.4/Rspamd 4.3, CrowdSec 1.8), dan skor reputasi RBL email disusun dalam kartu grid proporsional dengan border tipis `1px` dan angka monospasi sejajar (`font-variant-numeric: tabular-nums`).
2. **OLED-Friendly Dark Mode (Default) & High-Contrast Light Mode:**
   - **Dark Mode:** Palet **Zinc / Obsidian** (`#09090B` kanvas, `#121215` kartu, `#18181B` elevated, `#27272A` border) untuk kenyamanan kerja sysadmin jangka panjang.
   - **Light Mode:** Palet **Slate** bersih (`#F8FAFC` kanvas, `#FFFFFF` kartu, `#E2E8F0` border) dengan kontras rasio > 7:1 (**WCAG AAA**) serta fokus cincin keyboard (`:focus-visible`) yang jelas.
3. **Tipografi Fungsional:**
   - **UI & Heading:** `Plus Jakarta Sans`
   - **Data Teknis, Hash & Terminal:** `JetBrains Mono`
4. **Command Palette (`Cmd+K` / `Ctrl+K`):** Navigasi cepat lintas halaman dan aksi instan (pindah versi PHP domain, buka File Manager, buat mailbox, blokir IP di `nftables`, atau jalankan backup Restic).
5. **Feature-List & Role-Aware Navigation:** Hook `useFeatureGate.ts` otomatis menyembunyikan menu yang dinonaktifkan pada `feature_lists` paket hosting maupun berdasarkan peran (`superadmin`, `reseller`, `tenant`), serta menampilkan banner peringatan persisten saat admin berada dalam mode **"Login as User" (Impersonation)**.
6. **Dual Bahasa Instan (`id` & `en`):** Pergantian bahasa Indonesia / Inggris seketika via `i18next` yang tersimpan pada kolom `users.locale`.

---

## 2. Komponen Utama Frontend (`frontend/src/`)

- `styles/tokens.css`: Token warna semantik, `tabular-nums`, cincin fokus aksesibilitas, dan dukungan `prefers-reduced-motion`.
- `i18n/locales/id.json` & `en.json`: Kamus lengkap untuk seluruh modul Admin, Reseller, dan Tenant.
- `hooks/useFeatureGate.ts`: Penegakan visibilitas fitur paket, RBAC role check, dan deteksi peringatan PHP 8.2 EOL.
- `components/layout/CommandPalette.tsx`: Pencarian cepat `Cmd+K`.
- `components/charts/TelemetryBento.tsx`: Grafik WebSocket CPU, RAM, Disk IOPS, serta serangan Coraza WAF & blokir CrowdSec.
- `components/filemanager/CodeEditorModal.tsx`: File Manager dengan **Monaco Editor**.
- `components/terminal/WebTerminal.tsx`: Web SSH Terminal berbasis **xterm.js** yang terhubung ke PTY terkunci UID tenant.
