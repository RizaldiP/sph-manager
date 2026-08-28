# Architecture Analysis — SPH Manager (Phase 1: Audit)

> Output PHASE 1 — AUDIT sebelum implementasi fitur **Real-Time Chat** dan **Master Data Transfer**.
> Dokumen ini memetakan arsitektur existing dan memberikan rekomendasi implementasi.
> Desain detail fitur ada di [`FEATURE_CHAT_MASTER_TRANSFER.md`](FEATURE_CHAT_MASTER_TRANSFER.md).

---

## 1. Ringkasan

SPH Manager adalah aplikasi **desktop offline-first** untuk membuat & mengelola Surat Penawaran
Harga (SPH). Teknologi: **Go + Wails v2** dengan database **SQLite (GORM)** dan frontend
**Vue 3 + Pinia + Tailwind**. Sudah memiliki fitur **Work Together / Kolaborasi LAN** yang
mengandalkan model **host-authoritative** dengan WebSocket + UDP discovery.

Audit ini dilakukan untuk memastikan dua fitur baru (Chat real-time & Master Data Transfer) dapat
**diintegrasikan dengan arsitektur existing** tanpa rewrite, tanpa protokol kedua, tanpa mengganti
database, dan tetap **offline-first**.

---

## 2. Arsitektur Lapisan

| Lapisan | Teknologi | Lokasi kode |
|---|---|---|
| Desktop/backend | Go + Wails v2 | `app_*.go`, `main.go`, `internal/services/` |
| Database | SQLite (driver murni Go `glebarez/sqlite`) + GORM | `internal/database/` |
| Migrasi skema | `db.AutoMigrate` | `internal/database/migrate.go` |
| Repositori (data access) | GORM dao per entity | `internal/repositories/` |
| Model/entity | Struct GORM + JSON tags | `internal/models/models.go` |
| Frontend | Vue 3 + Pinia + Tailwind | `frontend/src/` |
| Bridge Go→UI | `runtime.EventsEmit` + `EventsOn` | `app_collab.go`, `BuatSphPage.vue`, `stores/` |
| Binding Wails | Generated, disimpan di repo | `frontend/wailsjs/` |
| Komunikasi LAN (Work Together) | WebSocket (client→host) + UDP broadcast (discovery) | `internal/collaboration/` |
| Spec/docs | Markdown berbahasa Indonesia | `docs/`, root repo |

---

## 3. Backend (Go / Wails)

- Entry: `main.go`, `app.go` (struct `App` mendefinisikan semua dependency service).
- Binding Wails dipisah per domain: `app_master.go`, `app_material.go`, `app_partner.go`,
  `app_sph.go`, `app_export.go`, `app_import.go`, `app_settings.go`, `app_template.go`,
  `app_collab.go`.
- Business logic berada di `internal/services/*` (bukan di binding). Binding hanya mendelegasikan.
- Konvensi: `Service` (CRUD + validasi + transaksi + audit), `Repository` (akses data GORM),
  `AuditWriter` untuk audit log.

---

## 4. Frontend (Vue 3)

- Routing: `frontend/src/router/index.ts` (Vue Router, hash history). Halaman Work Together:
  `/work-together` → `WorkTogetherPage.vue`.
- State: Pinia stores (`stores/collaboration.ts`, `stores/master.ts`, ...).
- Tipe domain: `frontend/src/types/*.ts` (mirror JSON backend).
- Event real-time dari Go: `EventsOn('collab:sync', ...)` di `BuatSphPage.vue` / `WorkTogetherPage.vue`.
- Design system: Tailwind + komponen reusable (`AppModal.vue`, `ConfirmDialog.vue`, `EmptyState.vue`,
  `PageHeader.vue`, `SideNav.vue`, `TopBar.vue`).

---

## 5. Database

- Engine: SQLite, mode WAL, foreign key ON, busy_timeout (`internal/database/database.go`).
- Migrasi: `database.Migrate` memanggil `db.AutoMigrate` untuk semua model (`migrate.go`).
- Index unik parsial untuk natural key `Code` (unique bila non-empty & tidak terhapus) pada
  `categories`, `work_items`, `materials`, `customers`, `vessels`, `templates`.
- Tabel `settings` (key-value) dipakai untuk konfigurasi (mis. `collab_port`, `collab_display_name`).
- Model master: `Category`, `WorkItem`, `WorkSubItem`, `Material` (lihat §7).

### Hasil audit database

- Pola migrasi menggunakan **AutoMigrate** — aman untuk menambah model baru tanpa menyentuh tabel lama.
- Ada `createPartialUniqueIndexes` untuk index unik parsial — pola yang bisa diperluas bila perlu.
- Tidak ada sistem migrasi versi numerik; AutoMigrate cukup untuk fase berikutnya.

---

## 6. Work Together / Kolaborasi LAN (existing)

### 6.1 Bentuk komunikasi

- **Transport utama**: WebSocket (`gorilla/websocket`). Setiap client membuka **satu koneksi ke host**.
- **Discovery**: UDP broadcast (port `DefaultDiscoveryPort = 48766`) + join manual via IP/port.
- **Model**: host-authoritative. Room dibuat di host (in-memory); client join sebagai anggota.

### 6.2 Envelope / protokol

`internal/collaboration/messages.go` — `Envelope` memiliki:
`messageId`, `roomId`, `clientId`, `timestamp`, `type`, `version`, `payload` (json.RawMessage).

Tipe existing (contoh):
- `JOIN_REQUEST`, `LEAVE`, `OP_REQUEST`, `SYNC_REQUEST`, `SYNC_PUSH`, `PING`,
  `ASSIGN_TURNS`, `REQUEST_EDIT`, `RELEASE_EDIT` (client→host)
- `ROOM_JOINED`, `SYNC_RESPONSE`, `SPH_UPDATED`, `TURNS_UPDATED`, `USER_CONNECTED`,
  `USER_DISCONNECTED`, `ROOM_CLOSED`, `PONG`, `ERROR` (host→client)

### 6.3 Hub & Role

- `internal/collaboration/hub.go` — `Manager` mengelola satu sesi aktif per aplikasi
  (host ATAU client, tidak keduanya). Menyediakan `HostRoom`, `Join`, `Leave`, `SendOp`.
- `Room` — state in-memory host: `participants`, `byID`, `conns`, `activities`, `turn`.
- `Room.serveConn` membaca envelope dari client & memproses per tipe; `Room.broadcastLocked`
  menyiarkan envelope ke semua koneksi remote.
- `internal/collaboration/server.go` — wrapper HTTP+WebSocket (host) + `serverConn` (antrian kirim).
- `internal/collaboration/client.go` — sesi client: dial + auto-reconnect + heartbeat.

### 6.4 UI Sync mechanism

- Perubahan sesi → `Manager.emit(UISnapshot)` → `app_collab.go.wireCollab` →
  `runtime.EventsEmit(ctx, "collab:sync", snap)` → store `collaboration.ts` menerapkan snapshot.
- `UISnapshot` berisi `mode, room, doc, participants, activities, turn, version, error, notice`.

### 6.5 Kesimpulan audit kolaborasi

- **Host adalah hub tunggal** — semua pesan antar-anggota sudah melewati host.
  Cocok untuk menjadi router chat & transfer master data.
- **Envelope sudah membawa field event yang dibutuhkan** (id, room, sender, timestamp, type, payload).
- **Pesan masuk dipilah di `Room.serveConn` via switch `e.Type`** — perlu menambah case baru
  untuk tipe chat/master data.
- **Riwayat/activity room bersifat in-memory** dan tidak dipersist — untuk riwayat chat yang
  bertahan perlu persistensi (diusulkan di host, lihat dokumen fitur).
- **Tidak ada sistem callback target per-member** selain broadcast; untuk transfer ke member
  tertentu perlu mekanisme rute (host → `r.conns[participantID]`).

---

## 7. Master Data (existing)

Entity Master Data (untuk ditransfer):

| Entity | Tabel | Relasi / natural key |
|---|---|---|
| `Category` | `categories` | induk WorkItem; key `Code`/`Name` |
| `WorkItem` | `work_items` | `CategoryID → Category`; key `Code`/`Name` |
| `WorkSubItem` | `work_sub_items` | `WorkItemID → WorkItem`; key `Code`/`Name` |
| `Material` | `materials` | mandiri; key `Code`/`Name` |

- Metadata: `createdAt`, `updatedAt`, `deletedAt` (soft delete via `gorm.DeletedAt`),
  `isActive`, `sequence`.
- Natural key `Code` dijamin unik parsial (bila non-empty & live) oleh `createPartialUniqueIndexes`.
- Repository/service: GORM; operasi tulis melalui service layer.

### Kesimpulan audit master data

- Transfer harus memakai **natural key (`Code`/`Name`)** untuk relasi & dedup, bukan ID integer
  (karena ID lokal antar PC bisa berbeda).
- Kolom audit/timestamp/ID lokal harus di-strip dari package; penerima membuat ID lokal sendiri.

---

## 8. Pembelajaran Penting (untuk desain fitur)

1. **Reuse satu WebSocket + Envelope + host hub** — tidak perlu protokol/transport kedua.
2. **Riwayat chat** dipersist di DB host (sumber kebenaran Room); client ambil saat join.
3. **Master Data** setiap PC di DB lokal; transfer = serialize entity terpilih → JSON package
   + checksum → rute via host → validate → preview → confirm → import transaksional.
4. **Tidak ada salinan DB mentah**, tidak ada overwrite otomatis, tidak ada SQL/command arbitrary.
5. **Natural key (Code/Name)** untuk dedup/compare/conflict & relasi package.
6. **Konfigurasi** via tabel `settings` (mis. `masterdata_max_package_size`).

---

## 9. Rekomendasi Implementasi

| Area | Rekomendasi |
|---|---|
| Protocol | Tambah tipe Envelope: `CHAT_MESSAGE`, `CHAT_HISTORY_REQUEST`, `CHAT_HISTORY`, `CHAT_BROADCAST`, `MASTER_DATA_OFFER`, `MASTER_DATA_TRANSFER`, `MASTER_DATA_ACK`, `MASTER_DATA_STATUS` |
| Database | Migrasi/AutoMigrate model baru: `chat_messages`, `master_data_inbox`, `master_data_sent`; kunci settings baru |
| Routing | Host terima chat → persist + broadcast; host terima transfer → teruskan ke target tertentu (via `r.conns[participantID]`) atau broadcast |
| Chat history | Simpan di DB host; load saat join (`CHAT_HISTORY_REQUEST`) |
| Transfer | Serializer `MasterDataPackage` (JSON + SHA-256) di service; import transaksional di service master |
| UI | Panel Chat + badge unread di Work Together; modal Kirim Master Data; halaman/panel Inbox & Sent; preview/compare/conflict |
| Wails binding | Tambah method baru (chat, history, unread, pilih master, kirim, inbox, detail, compare, import, reject, sent) + event `collab:chat`, `masterdata:inbox` |

---

## 10. Batasan / Risiko

- Riwayat chat hanya bertahan selama Room hidup di host (room host-authoritative & sementara).
- Tanpa status `READ`, tanpa file transfer umum untuk v1.
- Transfer chunk tunggal tanpa resume/retry otomatis (ukuran master kecil).
- Konflik: pilihan user, tanpa auto-merge.
- `Template`/`TemplateItem` tidak disertakan dalam package Master Data v1.

---

## 11. Lanjut ke Phase 2 (Database)

Dari sini, implementasi dapat dilanjutkan ke **Phase 2 — Database**: menambahkan model GORM baru
(`ChatMessage`, `MasterInbox`, `MasterSent`) ke `database.Migrate` dan memastikan `AutoMigrate`
berjalan; lalu uji migrasi. Detail skema ada di `FEATURE_CHAT_MASTER_TRANSFER.md` (bagian 4).
