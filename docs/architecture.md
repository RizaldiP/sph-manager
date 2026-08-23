# Arsitektur — SPH Manager Offline

> Desktop app: **Wails v2** (Go backend + Vue 3 frontend) + **SQLite**. 100% offline, semua asset di-bundle.

## Gambaran Umum

```text
┌─────────────────────────────────────────────┐
│                Windows Desktop              │
│                                             │
│  ┌──────────────┐   Wails Bridge   ┌──────┐ │
│  │    Vue 3     │◄────────────────►│  Go  │ │
│  │  (WebView2)  │  bindings (JSON) │      │ │
│  └──────────────┘                  └──┬───┘ │
│                                       │     │
│                              ┌────────▼────┐│
│                              │   SQLite    ││
│                              │ (WAL + FK)  ││
│                              └─────────────┘│
└─────────────────────────────────────────────┘
```

## Struktur Proyek

```text
sph-manager/
├── main.go                      # entrypoint: config → logger → database → wails.Run
├── app.go                       # struct App (bound ke frontend): Health(), lifecycle
├── internal/
│   ├── config/                  # konfigurasi JSON + path AppData (OS-aware)
│   ├── logger/                  # slog → file (logs/app.log) + stdout (aman windowsgui)
│   ├── database/                # GORM + SQLite pure-Go (tanpa CGO), WAL, foreign_keys
│   ├── models/                  # (Phase 2) model GORM semua tabel
│   ├── repositories/            # (Phase 3+) akses data
│   ├── services/                # (Phase 3+) business rule, perhitungan, validasi
│   ├── validators/              # (Phase 5+) validasi input & dokumen
│   ├── documents/               # (Phase 9) generator Excel/PDF
│   ├── importers/               # (Phase 8) reader & parser Excel
│   ├── backup/                  # (Phase 10) backup/restore/auto-backup
│   └── migrations/              # (Phase 2) migrasi skema
├── frontend/
│   ├── src/
│   │   ├── layouts/MainLayout.vue   # Sidebar + Topbar + konten
│   │   ├── components/              # SideNav, TopBar, PageHeader, EmptyState, …
│   │   ├── pages/                   # DashboardPage, PlaceholderPage, …
│   │   ├── stores/                  # Pinia (app store: health)
│   │   ├── router/                  # Vue Router (hash history)
│   │   ├── services/                # (nanti) wrapper pemanggilan binding Go
│   │   ├── types/                   # (nanti) tipe domain bersama
│   │   └── style.css                # Tailwind v4 + design token biru/orange flat
│   └── wailsjs/                     # binding otomatis hasil generate Wails
├── build/                          # ikon, manifest, installer template
└── docs/
```

## Keputusan Teknis

| Aspek | Keputusan | Alasan |
|---|---|---|
| Driver SQLite | `github.com/glebarez/sqlite` (pure-Go) | Tanpa CGO → build Windows sederhana & portabel |
| Database mode | WAL + `foreign_keys=ON` + `busy_timeout=5000` | Aman, cepat, integritas relasi |
| Logging | `log/slog` teks → file + stdout (multi-sink aman) | Structured log; stdout GUI bisa gagal → tiap sink independen |
| Konfigurasi | JSON di `%AppData%\sph-manager\config.json`, auto-create | OS-aware, sesuai FR-P3 |
| Router | hash history (`createWebHashHistory`) | Wajib untuk asset protocol WebView |
| Styling | Tailwind CSS v4 (`@tailwindcss/vite`) dengan token `brand-*` (biru flat) & `accent-*` (orange flat) | Konsisten, offline-bundled, tanpa CDN/font online |
| Font | System stack ("Segoe UI", system-ui) | Tanpa font online; native look Windows |

## Design Token (UI)

- `brand-600 #2563EB` — warna primer (sidebar aktif, tombol utama)
- `accent-500 #F97316` — aksen (aksi cepat "Buat SPH", highlight)
- Netral: `slate-50 … slate-800` untuk latar/teks
- Semua komponen memakai token ini — tidak ada warna hard-code di component

## Alur Data

1. Frontend memanggil binding (`wailsjs/go/main/App.*`) → Go method.
2. Go method → service layer → repository → GORM → SQLite.
3. Hasil (struct bertag JSON) → kembali ke frontend sebagai Promise.
4. Business rule **tidak pernah** di Vue component (aturan Master Specification §6).

## Lokasi Data Runtime

```text
%AppData%\sph-manager\
├── config.json      # konfigurasi aplikasi
├── database\sph.db  # SQLite (+ .wal/.shm saat berjalan)
├── backups\         # hasil backup (Phase 10)
├── exports\         # hasil export dokumen (Phase 9)
├── logs\app.log     # structured log
└── templates\       # aset dokumen (Phase 9)
```

## Build

```powershell
wails dev      # pengembangan (vite hot-reload + window)
wails build    # produksi → build/bin/SPHManager.exe
```
