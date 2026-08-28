# Fitur Chat & Master Data Transfer — SPH Manager

> Dokumen desain + rencana implementasi untuk dua fitur baru yang terintegrasi penuh
> dengan sistem **Work Together / Room** yang sudah ada:
>
> 1. **Real-Time Chat** (percakapan antar anggota Room melalui LAN)
> 2. **Master Data Transfer** (kirim Master Data SPH antar komputer melalui LAN)
>
> Kedua fitur bekerja **tanpa internet**, tanpa cloud, tanpa hosting, tanpat domain, dan
> tanpa mengganti/replace database. Semua komunikasi memakai infrastruktur LAN yang sudah ada.

---

## 1. Ringkasan

SPH Manager adalah aplikasi **desktop offline-first** untuk membuat Surat Penawaran Harga (SPH).
Aplikasi sudah memiliki fitur **Work Together / Kolaborasi LAN**: beberapa komputer dalam satu
jaringan lokal dapat bergabung ke satu **Room** dan mengerjakan satu dokumen SPH secara bersama.

Fitur baru ini menambahkan dua kemampuan di dalam Room:

```
                SPH MANAGER
                     │
             ┌───────┴────────┐
             │ WORK TOGETHER  │
             └───────┬────────┘
                     │
            ┌────────┴─────────┐
            │                  │
          CHAT            MASTER DATA
            │                  │
      Real-time LAN      Send / Receive
            │                  │
      Message History        Preview
            │                  │
      Notifications            │
                               ├─ Compare
                                 ├─ Install
                                 ├─ Merge
                                 └─ Rollback
```

**Target akhir:** beberapa komputer dalam satu LAN dapat bekerja bersama dalam satu Room,
berkomunikasi melalui Chat, dan saling bertukar Master Data SPH dengan aman — tanpa internet,
cloud, hosting, domain, atau database replacement.

---

## 2. Analisis Arsitektur Existing (Hasil Audit — Phase 1)

### 2.1 Teknologi

| Lapisan | Teknologi | Lokasi utama |
|---|---|---|
| Backend/desktop | Go + **Wails v2** (offline-first) | `app_*.go`, `main.go` |
| Database | **SQLite** (driver murni Go `glebarez/sqlite`) via **GORM** | `internal/database/` |
| Migrasi skema | `db.AutoMigrate` pada `database.Migrate` | `internal/database/migrate.go` |
| Frontend | Vue 3 + Pinia + Tailwind | `frontend/src/` |
| Bridge Go→UI | `runtime.EventsEmit(ctx, "collab:sync", snap)` + `EventsOn` | `app_collab.go`, `BuatSphPage.vue` |
| Binding Wails | Digenakkan dan disimpan di `frontend/wailsjs/` | `frontend/wailsjs/go/main/App.*` |

### 2.2 Work Together / Room (existing — REUSE, jangan ditulis ulang)

- Model **host-authoritative**: satu komputer membuat Room dan menjadi **HOST**; komputer lain
  menjadi **CLIENT**.
- **Transport tunggal**: setiap client membuka **satu koneksi TCP/WebSocket ke host**
  (`gorilla/websocket`). Host adalah satu-satunya hub/router semua pesan di Room.
- **Discovery**: UDP broadcast (port `48766`) + alternatif join manual via IP/port.
- **Envelope** (`internal/collaboration/messages.go`) sudah memuat field yang dibutuhkan setiap event:
  `messageId`, `roomId`, `clientId`, `timestamp`, `type`, `version`, `payload`.
- Alur komunikasi client→host: client mengirim tipe pesan (`OP_REQUEST`, `SYNC_PUSH`, dll),
  host memproses lalu **broadcast** ke semua koneksi lain via `Room.broadcastLocked`.
- **Room bersifat in-memory di host**; yang disinkronkan hanya dokumen SPH aktif. Master data
  **tidak ikut** tersinkron (tercatat di `docs/collaboration-lan.md`).
- **Activity log** room bersifat in-memory (`Room.activities`, kapasitas ~100), tidak dipersist.
- **Member/peserta** direpresentasikan oleh `Participant{ID, DisplayName, DeviceName, Role}`,
  dilacak di `r.byID` dan `r.info.Participants`. Host = `RoleHost`, client = `RoleEditor`.

### 2.3 Entity Master Data

Entity yang termasuk **Master Data** dan layak dikirim:

| Entity | Tabel | Relasi |
|---|---|---|
| `Category` (kategori pekerjaan) | `categories` | induk dari `WorkItem` |
| `WorkItem` (pekerjaan) | `work_items` | `CategoryID` → `Category` |
| `WorkSubItem` (sub-pekerjaan) | `work_sub_items` | `WorkItemID` → `WorkItem` |
| `Material` | `materials` | mandiri |

**Natural key** (sudah ada, lihat index unik parsial di `migrate.go`): `Code` (unik bila
tidak kosong & tidak terhapus) dan `Name`. Ini dipakai untuk deteksi duplikat/konflik saat merge.

> Catatan: `Template`/`TemplateItem` mereferensikan `WorkItem`, tetapi domain-nya terpisah
> ("template"). Untuk v1 ini **tidak** disertakan dalam package Master Data.

### 2.4 Simpulan Audit

1. **Reuse satu WebSocket + Envelope** — tidak perlu membuat sistem komunikasi kedua. Tambahkan
   tipe Envelope baru untuk chat & transfer, dirutekan via host.
2. **Riwayat chat** dipersist di **database host** (host adalah server/sumber kebenaran Room).
   Anggota mengambil riwayat dari host saat join. Menghindari penyimpanan ganda dan cocok dengan
   model host-authoritative.
3. **Master Data** hidup di SQLite lokal masing-masing PC. Transfer = serialize entity terpilih
   menjadi **MasterDataPackage** (JSON), dirutekan melalui host ke client tujuan; penerima
   **validate → preview → konfirmasi → import via transaksi** ke DB lokalnya. Tidak menyalin file DB,
   tidak menimpa otomatis.

---

## 3. Keputusan Desain

| Aspek | Keputusan | Alasan |
|---|---|---|
| Transport chat & transfer | **WebSocket + Envelope existing** via host | Nol protokol baru; host sudah jadi hub |
| Simpan riwayat chat | **DB local host** | Host = sumber kebenaran Room; anggota ambil saat join |
| Format transfer | **JSON object terstruktur (MasterDataPackage)** + **checksum SHA-256** | Bukan salinan DB mentah |
| Entity yang dikirim | `Category`, `WorkItem`, `WorkSubItem`, `Material` | Entity Master Data yang dipakai |
| Penerusan pesan | Semua melewati **host** (host meneruskan; host juga bisa jadi penerima) | Konsisten dengan arsitektur existing |
| Strategi konflik | **Pilihan user**, tanpa auto-merge | Keamanan data > otomatisasi (prioritas STABILITY/DATA SAFETY) |
| Transfer ukuran | **Chunk tunggal JSON**; batas `MAX_PACKAGE_SIZE` configurable | Master kecil; tanpa lapisan chunk/retry untuk v1 |
| Versioning | **Simplest**: simpan `source_version` (dan opsional `target_version`) | DB existing belum punya versi master terstruktur |
| Activity/System message | Tambah tipe `system` pada chat + event baru pada protocol | Pantau aktivitas Room tanpa membuat tabel log baru untuk v1 |

---

## 4. Perubahan Database (Migration Baru)

Semua penambahan memakai **`AutoMigrate`** di `database.Migrate` (penambahan model baru; tanpa
merubah/menghapus tabel existing). Gunakan entity GORM baru dengan `TableName()` sesuai konvensi.

### 4.1 `chat_messages` — riwayat chat Room (disimpan di host)

```sql
CREATE TABLE chat_messages (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  room_id      TEXT    NOT NULL,
  message_id   TEXT    NOT NULL,             -- uuid unik (dari Envelope.messageId)
  sender_id    TEXT    NOT NULL,             -- participant id
  sender_name  TEXT    NOT NULL,             -- displayName saat kirim
  message_type TEXT    NOT NULL DEFAULT 'text',  -- text | system | master_data
  content      TEXT    NOT NULL,             -- teks atau JSON metadata (untuk master_data: ref package)
  status       TEXT    NOT NULL DEFAULT 'SENT',   -- SENT | DELIVERED
  created_at   DATETIME NOT NULL
);
CREATE INDEX idx_chat_room_time ON chat_messages(room_id, created_at);
```

> Riwayat hanya perlu di DB **host**. Client tidak menyimpan salinan riwayat (mengambil dari host).

### 4.2 `master_data_inbox` — Master Data masuk (disimpan di penerima)

```sql
CREATE TABLE master_data_inbox (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  room_id       TEXT    NOT NULL,
  package_id    TEXT    NOT NULL UNIQUE,
  sender_id     TEXT    NOT NULL,
  sender_name   TEXT    NOT NULL,
  source_version INTEGER DEFAULT 0,
  payload       TEXT    NOT NULL,             -- JSON MasterDataPackage (tanpa checksum, atau utuh)
  checksum      TEXT    NOT NULL,             -- SHA-256 hex
  status        TEXT    NOT NULL DEFAULT 'PENDING',  -- PENDING | VIEWED | INSTALLED | REJECTED | FAILED
  received_at   DATETIME NOT NULL,
  installed_at  DATETIME,
  rejected_at   DATETIME
);
CREATE INDEX idx_mdi_status ON master_data_inbox(status);
CREATE INDEX idx_mdi_sender ON master_data_inbox(sender_id);
```

### 4.3 `master_data_sent` — riwayat Master Data terkirim (disimpan di pengirim)

```sql
CREATE TABLE master_data_sent (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  room_id       TEXT    NOT NULL,
  package_id    TEXT    NOT NULL UNIQUE,
  payload       TEXT    NOT NULL,
  checksum      TEXT    NOT NULL,
  source_version INTEGER DEFAULT 0,
  recipients    TEXT    NOT NULL,             -- JSON array participant id/nama tujuan
  status        TEXT    NOT NULL DEFAULT 'DELIVERED',  -- DELIVERED | (dapat di-update bila ada ACK INSTALLED)
  sent_at       DATETIME NOT NULL
);
```

> Keputusan: `master_data_inbox` & `master_data_sent` disimpan di DB **lokal masing-masing**
> (penerima/pengirim), karena inilah PC yang membutuhkannya untuk UI Inbox/Sent.

### 4.4 Opsional — konfigurasi di tabel `settings`

- `masterdata_max_package_size` (bytes, default mis. `1_000_000`). Ditambah sebagai kunci settings
  baru mengikuti pola `keyCollabPort` pada `settings_service.go`.

### 4.5 Model GORM baru (draft)

```go
type ChatMessage struct {
  ID          uint      `gorm:"primaryKey" json:"id"`
  RoomID      string    `gorm:"size:64;notNull;index" json:"roomId"`
  MessageID   string    `gorm:"size:64;notNull;index" json:"messageId"`
  SenderID    string    `gorm:"size:64;notNull" json:"senderId"`
  SenderName  string    `gorm:"size:150;notNull" json:"senderName"`
  MessageType string    `gorm:"size:20;notNull;default:text" json:"messageType"`
  Content     string    `gorm:"type:text;notNull" json:"content"`
  Status      string    `gorm:"size:20;notNull;default:SENT" json:"status"`
  CreatedAt   time.Time `json:"createdAt"`
}
func (ChatMessage) TableName() string { return "chat_messages" }
```

Model `MasterInbox` dan `MasterSent` mengikuti kolom tabel di atas.

---

## 5. Perubahan Protocol Jaringan

### 5.1 Tipe Envelope baru (`internal/collaboration/messages.go`)

Tambahkan konstanta (mengikuti konvensi `TYPE_UPPERCASE`) dan tambahkan case baru di
`Room.serveConn` (`hub.go`) serta penanganan di sisi client (`client.go`/`manager.go`).

```go
// client → host
TypeChatMessage        = "CHAT_MESSAGE"         // kirim pesan chat (text/system/master_data)
TypeChatHistoryRequest = "CHAT_HISTORY_REQUEST" // minta riwayat saat join
TypeMasterDataOffer    = "MASTER_DATA_OFFER"    // host → target client: ada transfer masuk
TypeMasterDataAck      = "MASTER_DATA_ACK"      // client → host: status (RECEIVED/REJECTED/INSTALLED/FAILED)
TypeMasterDataStatus   = "MASTER_DATA_STATUS"   // host → semua: perbarui status (untuk log/sent)

// host → client
TypeChatHistory   = "CHAT_HISTORY"      // payload daftar ChatMessage
TypeChatBroadcast = "CHAT_BROADCAST"    // broadcast pesan chat
TypeMasterData    = "MASTER_DATA_TRANSFER" // payload MasterDataPackage ke target tertentu
```

Semua event memakai struct `Envelope` yang sudah ada (memiliki `messageId`, `roomId`,
`senderId`/`clientId`, `timestamp`, `type`, `version`, `payload`).

### 5.2 Payload

```go
// ChatPayload — isi CHAT_MESSAGE / CHAT_BROADCAST / CHAT_HISTORY
type ChatPayload struct {
  MessageID   string    `json:"messageId"`
  MessageType string    `json:"messageType"` // text | system | master_data
  Content     string    `json:"content"`
  SenderID    string    `json:"senderId,omitempty"`
  SenderName  string    `json:"senderName,omitempty"`
  CreatedAt   time.Time `json:"createdAt"`
  Status      string    `json:"status"` // SENT | DELIVERED
  RefPackage  string    `json:"refPackage,omitempty"` // package_id bila message_type=master_data
}

// UISnapshot ditambah field: Chat []ChatPayload, Unread int
```

Untuk master data, payload ditambahkan field opsional `Package` (objek `MasterDataPackage`)
saat dikirim ke target, sedangkan saat broadcast hanya `RefPackage` + metadata ringkas sebagai card.

### 5.3 Alur broadcasting

```
User A (kirim chat)
   → Envelope CHAT_MESSAGE → host
   → host persist ke chat_messages (host)
   → host broadcast CHAT_BROADCAST ke semua member (termasuk echo ke A)
   → UI masing-masing update + badge unread (kecuali pengirim/ruang chat terbuka)
```

```
User A (kirim Master Data)
   → pilih entity → build MasterDataPackage + checksum (di sisi A)
   → simpan ke master_data_sent (A)
   → Envelope MASTER_DATA_TRANSFER → host
   → host validasi (room, sender member, ukuran) → teruskan ke target client T
   → host (jika T=host) simpan ke master_data_inbox host
   → target menerima → validasi checksum → simpan ke master_data_inbox (T) → kirim MASTER_DATA_ACK
   → host broadcast MASTER_DATA_STATUS (untuk log sent + notifikasi room)
   → T dapat chat system + notification
```

---

## 6. Struktur MasterDataPackage

```jsonc
{
  "metadata": {
    "package_id":      "uuid",
    "sender_id":       "participant-id-A",
    "sender_name":     "Admin",
    "room_id":         "room-uuid",
    "created_at":      "RFC3339",
    "schema_version":  "1",
    "package_version": "1",
    "source_version":  "2"          // versi master terpilih saat dikirim (opsional)
  },
  "categories":  [ { /* Category tanpa kolom audit, tanpa relasi */ } ],
  "work_items":  [ { /* WorkItem + categoryCode (natural key induk) */ } ],
  "work_sub_items": [ { /* WorkSubItem + workItemCode (natural key induk) */ } ],
  "materials":   [ { /* Material */ } ],
  "relationships": {
    // dipertahankan via natural key Code/Name pada tiap entity, mis.:
    // workItem.categoryCode, workSubItem.workItemCode
  },
  "checksum": "sha256-hex-dari-serialisasi-bagian-data"
}
```

Aturan:

- Kolom **audit/relasi/**`created_at`/`updated_at`/`id` lokal **di-strip** — penerima membuat ID lokal sendiri.
- Relasi disimpan memakai **natural key** (`Code`), bukan ID integer, agar aman di PC yang berbeda.
- Hanya entity yang benar-benar terpilih oleh user yang dimasukkan (bukan seluruh DB).
- `checksum` adalah SHA-256 dari JSON bagian data (`categories`..`relationships`), dihitung saat kirim
  dan diverifikasi saat terima. Bila tidak cocok → **TRANSFER FAILED / data rusak, kirim ulang**.
- `MAX_PACKAGE_SIZE` diterapkan; bila melebihi → tolak dengan pesan ramah.

---

## 7. Fitur Chat

### 7.1 Data model

- `message_type` minimal: `text`, `system`, `master_data` (opsional `file` — tidak dibuat untuk v1).
- `status` minimal: `SENT`, `DELIVERED`. (Untuk v1 tanpa `READ` agar tidak mengorbankan stabilitas.)
- `master_data` tampil sebagai **card** (tautan `refPackage`) — bukan JSON mentah.

### 7.2 Alur real-time

```
JOIN ROOM → LOAD CHAT HISTORY (host) → CONNECT REAL-TIME → RECEIVE NEW MESSAGE
```

- Saat `ROOM_JOINED`/`SYNC_RESPONSE`, client meminta `CHAT_HISTORY_REQUEST`; host membalas
  `CHAT_HISTORY` berisi daftar pesan terakhir.
- Pesan baru masuk via `CHAT_BROADCAST`.

### 7.3 Penyimpanan riwayat

- **Host** menyimpan ke `chat_messages` (host-authoritative). Client tidak menyimpan duplikat.
- Bila host-tutup room dan tak ada host = tidak ada riwayat yang bisa diakses (bukan kegagalan;
  Room bersifat sementara). Ini catat sebagai keterbatasan di §15.

### 7.4 Status & notifikasi

- Pesan pengirim → `SENT`; saat host meneruskan → `DELIVERED`.
- Badge unread di item navigasi **Work Together / Chat**: `Chat 🔴 n`. Saat chat dibuka, unread → 0.
- Notifikasi tidak popup berlebihan; cukup badge + sorotan pesan baru + (opsional) notifikasi sistem.

### 7.5 System message

Jenis `system` untuk aktivitas penting:
- `<Nama> mengirim Master Data "..."`
- `<Nama> memasang Master Data "..."`
- `<Nama> menolak Master Data "..."`

---

## 8. Fitur Master Data Transfer

### 8.1 Alur lengkap

```
USER A
  → Pilih Master Data (kategori/pekerjaan [+sub]/material, bisa banyak)
  → Pilih penerima (satu / beberapa / semua)
  → Generate MasterDataPackage (+ checksum)
  → Validasi package
  → Kirim via LAN (lewat host)
  → simpan master_data_sent (A)
  → USER B menerima
  → Notification + masuk master_data_inbox (status PENDING)
  → Preview (detail + perubahan)
  → Compare dengan Master Data lokal (NEW/UPDATED/UNCHANGED/CONFLICT)
  → User pilih "Pasang"
  → Validasi ulang
  → Backup prapasang / transaction
  → Import / Merge
  → COMMIT; update status → INSTALLED; broadcast MASTER_DATA_STATUS
  → Selesai
```

### 8.2 UI pemilihan & penerima

- Modal **Kirim Master Data**: pencarian pekerjaan (checkbox, bisa banyak), pilihan penerima
  (Semua Member / per member), tombol `[Batal] [Kirim]`.
- Mendukung: satu/banyak master; satu/banyak member; semua member.

### 8.3 Inbox & Sent

- **📥 Master Data Masuk** (dari `master_data_inbox`): card `📦 Nama · vX · Dari: Nama · waktu`
  dengan tombol `[Lihat] [Pasang]`. Status: `PENDING → VIEWED → INSTALLED/REJECTED/FAILED`.
- **📤 Master Data Terkirim** (dari `master_data_sent`): card paket + daftar penerima + status
  (Delivered / Installed).

### 8.4 Preview & Compare

- Sebelum pasang, user melihat **preview** ringkas (dikirim oleh, versi, jumlah +/−/≈ baris).
- **Compare** dengan data lokal menghasilkan kategori:
  - `NEW` (+): belum ada (natural key tidak ditemukan)
  - `UPDATED` (~): ada di lokal, beda isi
  - `UNCHANGED` (=): identik
  - `CONFLICT` (!): tabrakan — mis. sama natural key tapi beda isi dan rawan hilang data

### 8.5 Import / Merge (aman)

- Seluruh operasi pemasangan memakai **database transaction**:

```
BEGIN TRANSACTION
  validate ulang
  backup/prepare rollback (snapshot baris terdampak / tulis ulang-batal)
  import entity baru (dengan natural-key dedup → EXISTING vs NEW)
  update entity
  buat/betulkan relasi (Category → WorkItem → WorkSubItem)
  validasi relasi
COMMIT
```
Bila ada error → `ROLLBACK`; DB tidak pernah dalam kondisi setengah ter-import.

- Duplicate handling **berbasis natural key**, bukan ID: `Code` (bila bukan "") atau `Name`.

### 8.6 Strategi Konflik

**Jangan** menimpa otomatis. Saat `CONFLICT`:

```
CONFLICT TERDETEKSI
[Gunakan Lokal] [Gunakan Incoming] [Bandingkan] [Batal]
```

Untuk v1, bila terlalu kompleks → **minta user memilih** (Gunakan Lokal / Incoming), tanpa merge otomatis.

### 8.7 Versi

- Simpan `source_version` (opsional `target_version`) pada package & inbox/sent.
- Tidak membuat sistem versioning baru di tabel master existing (DB belum mendukung).
- Versi diwakili oleh nilai sederhana (mis. integer) yang dapat dinaikkan user saat perubahan besar.

---

## 9. Keamanan, Offline-First, dan Disconnect

### 9.1 Keamanan (LAN tetapi tidak percaya payload)

Validasi wajib saat menerima:
- `room_id` cocok dengan Room aktif
- `sender_id` adalah member Room
- checksum paket cocok
- `schema_version` didukung
- struktur payload sesuai (field wajib, tipe benar)
- referensi entity valid (induk Code ada/akan dibuat)

**Dilarang**: eksekusi SQL arbitrary / eksekusi command / kode dari network payload. Import hanya
menyentuh tabel yang memang ditentukan importer (kategori, pekerjaan, sub, material).

### 9.2 Offline-first

- Bekerja dengan **Internet OFF, LAN ON**.
- Tanpa Firebase/Supabase/cloud/domain/external API. Hanya WebSocket+UDP existing.

### 9.3 Disconnect handling

- **Client disconnect saat terima**: `TRANSFER INTERRUPTED`; data tak lengkap tidak masuk DB.
  Untuk v1: package dikirim **chunk tunggal**; bila koneksi putus saat transfer, penerima menolak
  dan meminta kirim ulang (log event `FAILED`). Tidak membuat lapisan chunk/retry (ukuran kecil).
- Reconnect tetap jalan seperti existing (auto-reconnect client). Usulan untuk menghindari pesan
  tak lengkap: kirim paket hanya bila status koneksi penerima `CONNECTED` (via host presence).

### 9.4 Batas ukuran

- `MAX_PACKAGE_SIZE` (default ~1 MB, configurable di Pengaturan / settings key
  `masterdata_max_package_size`). Lewat batas → `Master Data terlalu besar untuk dikirim.`

---

## 10. Perubahan Binding Wails (API)

Tambahan di `app_collab.go` (atau file binding baru di `app_*.go`):

- `SendChat(messageType, content) error` — kirim chat room.
- `GetChatHistory(limit) ([]ChatMessage, error)` — ambil riwayat (dari host).
- `GetChatUnread() int` — badge unread (lokal client).
- `SelectMasterData(ids) / ListSelectableMaster(search) ([]WorkItemView, error)` — pemilihan untuk kirim.
- `SendMasterData(items []SelectedMaster, recipients []ParticipantID, all bool) (packageID, error)`
  — generate package + kirim.
- `ListMasterInbox() ([]MasterInboxView, error)` — daftar masuk.
- `GetMasterPackage(packageID) (*MasterDataPackage, error)` — detail untuk preview.
- `CompareMasterPackage(packageID) ([]CompareItem, error)` — hasil NEW/UPDATED/UNCHANGED/CONFLICT.
- `ImportMasterPackage(packageID) (result, error)` — konfirmasi pasang + transaction.
- `RejectMasterPackage(packageID) error`.
- `ListMasterSent() ([]MasterSentView, error)`.

Event UI baru (via `runtime.EventsEmit`):
- `collab:chat` (snapshot chat + unread)
- `masterdata:inbox` (notifikasi masuk baru internat)
- `collab:sync` diperpanjang dengan `Chat`/`Unread`.

---

## 11. Perubahan UI/UX

- **Panel Chat** pada halaman Room (host & client): daftar member + area chat + input `[Kirim]`.
- **Message bubble**: nama pengirim, waktu, tipe (teks / card master data / system).
- **Card Master Data** di chat: `📦 Nama · versi · ringkasan` + tombol `[Lihat] [Pasang]` — bukan JSON mentah.
- **Badge unread** pada navigasi Work Together (`Chat 🔴 n`), reset saat chat dibuka.
- **Modal Kirim Master Data** (pilih master + penerima).
- **Halaman/panel 📥 Master Data Masuk** dan **📤 Master Data Terkirim** (dapat diakses dari Work Together).
- **Modal Preview/Compare/Conflict** sesuai strategi §8.6 (dapat di-scroll; layout aman untuk data panjang).
- Gunakan design system existing (typography, button, modal, notification, card, icon). Dukung **dark mode**.
- Desktop: panel chat dapat di-resize; chat tidak menutupi pekerjaan utama.

---

## 12. Rencana Implementasi (Fase 1–14)

| Fase | Isi | Output & verifikasi |
|---|---|---|
| 1 | **Audit** — analisis arsitektur (dokumen ini) | `ARCHITECTURE`/desain disetujui |
| 2 | **Database** — model + migrasi `chat_messages`, `master_data_inbox`, `master_data_sent`, settings baru | Migrate jalan; unit test |
| 3 | **Protocol** — tipe Envelope + routing host | Test komunikasi |
| 4 | **Chat backend** — send/receive/persist/history/unread/status | Unit test |
| 5 | **Chat UI** — panel, bubble, badge, input | Uji manual |
| 6 | **Serializer** — `MasterDataPackage` serialize/deserialize/checksum/validasi | Unit test |
| 7 | **Transfer** — pilih master, penerima, send/receive/ACK/error | Unit test |
| 8 | **Inbox** — daftar masuk, notifikasi, preview, status | Uji |
| 9 | **Comparison** — NEW/UPDATED/UNCHANGED/CONFLICT | Unit test |
| 10 | **Import/Merge** — validasi, transaksi, import, relasi, rollback | Uji agresif (rollback test) |
| 11 | **Chat attachment** — sambung Chat → card → preview → install | Uji |
| 12 | **Activity/system message** — `CHAT_SENT`, `MASTER_DATA_SENT/RECEIVED/INSTALLED/REJECTED` | Uji |
| 13 | **UI polish** — loading, empty, error, notifikasi, dark mode, responsive | Uji manual |
| 14 | **Testing** — seluruh unit test + manual 1 host/3 client | Lihat §13 |

---

## 12b. Status Pengerjaan (live)

| Fase | Status | Catatan |
|---|---|---|
| 1 | ✅ Selesai | `ARCHITECTURE_ANALYSIS.md` |
| 2 | ✅ Selesai | Model/migrasi + settings `masterdata_max_package_size`; unit test lolos |
| 3 | ✅ Selesai | Tipe Envelope chat/master + routing host; `master_package.go`, `chat_store.go`, `hub.go`, `client.go` |
| 4 | ✅ Selesai | Chat backend: send/handle/persist (SQLite via `gormChatStore`), history, unread host+client, binding `SendChatMessage`/`ClearChatUnread`/`GetChatUnread`; unit test `handleIncomingChat` |
| 5 | ✅ Selesai | UI: `CollabChatPanel.vue` (bubble, badge unread, input) di `CollabToolbar.vue` + `WorkTogetherPage.vue`; vue-tsc & vite build lolos |
| 6–7 | ✅ Selesai | Routing Master Data: envelope `TypeMasterData`/`Ack`, `MasterDataStore` interface (anti cycle), `Manager.SendMasterData`/`AcknowledgeMasterData`/`CurrentIdentity`/`CurrentRoomID`, host reload ke semua client, broadcast `ChatTypeMasterData`, `clientIncoming` + `UISnapshot.Incoming`; routing unit test (`fakeMasterDataStore`) |
| 8 | ✅ Selesai | `internal/masterdata.Service`: `BuildPackage` (semua baris aktif, urutan sequence, parent natural key, checksum SHA-256) |
| 9 | ✅ Selesai | `Compare(pkg)` → `[]DiffItem` (NEW/UPDATED/UNCHANGED/CONFLICT) + `PreviewMasterData` binding |
| 10 | ✅ Selesai | `Install(pkg,strategy,decisions)`: verifikasi checksum, satu `db.Transaction`, create berurutan kategori→pekerjaan→sub→material, resolusi parent via natural key, skip induk hilang, dedup, audit, rollback; konflik strategi PROMPT/USE_LOCAL/USE_INCOMING/SKIP |
| 11 | ✅ Selesai | Inbox/Sent persist: `SaveInbox` (dedup PackageID), `InboxList/Get/Payload`, `SetInboxStatus`, `SaveSent`, `UpdateSentStatus`; status ACK ↔ DB dialihkan lewat `MasterDataStore` |
| 12 | ✅ Selesai | App layer: `gormMasterDataStore` bridge + binding `BuildMasterDataPackage`/`SendMasterData`/`ListMasterInbox`/`GetMasterInbox`/`GetMasterInboxPayload`/`PreviewMasterData`/`InstallMasterData`/`RejectMasterData`/`MarkMasterInboxViewed`; wails module di-regenerate |
| 13 | ✅ Selesai | UI: `MasterDataPanel.vue` (kirim dialog rakit+target, inbox list+pratinjau diff, install strategi, tolak) + store action master data, badge pending; vue-tsc & vite build lolos |
| — | ✅ Test | `go build/vet/test ./...` + `npm run build` hijau; unit test `internal/masterdata/service_test.go` (build, install fresh/update, skip-konflik, use-incoming, corrupt, compare, inbox dedup/status)

Build/verifikasi hijau: `go build ./...`, `go vet ./...`, `go test ./...`, `npm run build` (frontend typecheck+vite).

---

## 13. Testing

### 13.1 Unit test (wajib)

**Chat**
- send message, receive message, persistence, room isolation, disconnect, reconnect, unread count.

**Master Data**
- package generation, serialization, checksum, validation, duplicate detection,
  comparison, conflict, import, rollback, rejected package, corrupted package,
  wrong room, unauthorized sender.

**LAN (host)**
```
Host + Client A + Client B + Client C
  A→B chat; A→C chat
  A→B Master Data; A→semua Master Data
  B install; C reject
```

**Databes safety khusus (rollback)**
```
BEGIN IMPORT → ERROR → ROLLBACK
```
Pastikan DB tidak berubah (tidak 50/50 data masuk).

### 13.2 Manual test (Phase 14)

1 HOST + 3 CLIENT:
- Chat: Host→Client1, Client1→Host, Client2→Client3.
- Master: Host→Client1, Host→Client2, Host→All.
- Install: Client1 Install, Client2 Reject.
- Conflict: Client1 memiliki data berbeda → jalankan strategi konflik.

---

## 14. Cara Menjalankan & Testing LAN

1. Pastikan semua PC di WiFi/LAN yang sama.
2. Jalankan build/run seperti existing (Wails): `wails dev` untuk pengembangan,
   atau build installer. (Generator binding Wails otomatis saat build/dev.)
3. **Host**: buka SPH DRAFT → **Work Together → + Mulai Room Baru** (buka halaman
   `work-together`). Catat kode ruang/akses + IP + port.
4. **Client**: **Gabung via IP** (atau join dari daftar discovery) dengan access code.
5. Di dalam Room: gunakan **panel Chat** untuk mengirim pesan.
6. **Kirim Master Data**: tombol `📦 Kirim Master Data` → pilih pekerjaan → pilih penerima → Kirim.
7. **Terima Master Data**: Client buka **📥 Master Data Masuk** → `[Lihat]` → `[Pasang]`, atau
   pasang langsung dari card di Chat.
8. Pantau **📤 Master Data Terkirim** di pengirim untuk status.
9. Uji mati internet → fitur tetap jalan selama LAN aktif.

---

## 15. Keterbatasan yang Diketahui (v1)

- Riwayat chat hanya selama Room hidup & disimpan di DB host; saat host tutup room,
  riwayat tidak tersedia bagi client.
- Tanpa status `READ` (hanya `SENT`/`DELIVERED`).
- Tanpa file transfer umum (kecuali `master_data`-card).
- Transfer chunk tunggal, tanpa resume/retry otomatis untuk paket yang terputus
  (`TRANSFER INTERRUPTED` → kirim ulang).
- Tidak ada merge otomatis untuk `CONFLICT` (pilihan user).
- `Template`/`TemplateItem` tidak disertakan dalam package Master Data.
- Versioning master disederhanakan (`source_version` integer), bukan sistem revisi penuh.

---

## 16. Acceptance Criteria (ringkasan)

**CHAT**
- [ ] Chat dalam Room, real-time, nama & waktu terlihat, history tersedia, unread badge,
      disconnect/reconnect aman, room isolation.

**MASTER DATA**
- [ ] Pilih master, pilih penerima, kirim via LAN, notifikasi, package tervalidasi, checksum,
      preview, compare, duplicate & conflict ditangani, konfirmasi sebelum install,
      import transaction + rollback, history & status transfer tersedia.

**INTEGRASI**
- [ ] Terintegrasi dengan Room/member existing, reuse LAN communication, tanpa internet/cloud,
      tidak merusak Work Together existing.
