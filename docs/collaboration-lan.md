# Kolaborasi Real-Time LAN — Spec Detail PHASE 10

> Dokumen ini adalah spesifikasi detail fitur **Work Together / Kolaborasi Real-Time LAN**
> yang menjadi acuan implementasi PHASE 10. Ringkasan roadmap ada di
> [`development-plan.md`](development-plan.md). Status: ⬜ Belum diimplementasikan.

## Tujuan

Tambahkan fitur **Work Together / Kolaborasi Real-Time melalui LAN**.

Fitur ini memungkinkan beberapa komputer yang berada pada jaringan lokal yang sama untuk mengerjakan satu SPH secara bersamaan dan hampir real-time.

### Constraint Utama

Fitur ini HARUS:

- bekerja tanpa internet;
- tidak membutuhkan hosting;
- tidak membutuhkan domain;
- tidak membutuhkan VPS;
- tidak membutuhkan cloud database;
- tidak membutuhkan server online;
- tetap bekerja walaupun koneksi internet dimatikan selama LAN lokal masih aktif.

Gunakan model:

```text
HOST PC
  |
  | LAN / Wi-Fi lokal
  |
  +------ CLIENT PC
  +------ CLIENT PC
  +------ CLIENT PC
```

Komputer Host menjalankan service kolaborasi lokal di dalam aplikasi **SPH Manager**.

Tidak perlu membuat aplikasi Server terpisah untuk MVP.

---

## Keputusan Implementasi (terkunci sebelum eksekusi)

| Aspek | Keputusan | Alasan |
|---|---|---|
| Scope sinkronisasi | Hanya SPH aktif dalam Room (sesuai §10.7) | Master data tidak ikut; arsitektur snapshot BR-01 membuat client tak butuh data master |
| State Room | **In-memory** di host, hilang saat app ditutup | Sesuai §10.4 (Room sementara) & §10.27; data SPH tetap aman di SQLite |
| Discovery | **UDP broadcast** (stdlib Go) | Nol dependensi baru; fallback IP manual tetap wajib |
| Mulai Room | Dari **draft SPH existing**; "buat baru" lewat wizard normal lalu kembali | Hindari duplikasi logika wizard di mode kolaborasi |
| Port default | `48765`, configurable di Pengaturan | Sesuai contoh spesifikasi |
| Dependensi | `gorilla/websocket` + `google/uuid` (sudah ada di go.mod) | Nol dependensi baru |
| Konflik edit | Host-authoritative, last accepted operation (§10.17) | Tanpa CRDT; cukup untuk form terstruktur |

---

## 10.1 — Konsep Host dan Client

Satu komputer membuat Room dan menjadi **HOST**. Komputer lain menjadi **CLIENT**.

Contoh:

```text
PC ADMIN
   ↓
Create Room
   ↓
Room: SPH KRI BAC-593
   ↓
PC ADMIN = HOST
```

Komputer lain:

```text
PC 02 → Join
PC 03 → Join
PC 04 → Join
```

Semua mengerjakan SPH yang sama.

---

## 10.2 — Arsitektur

```text
                 HOST COMPUTER
        ┌───────────────────────────┐
        │ SPH Manager               │
        │                           │
        │ Vue 3 UI                  │
        │      ↕                    │
        │ Wails                     │
        │      ↕                    │
        │ Go Application            │
        │      │                    │
        │      ├── SQLite           │
        │      │                    │
        │      └── Collaboration    │
        │           Server          │
        └────────────┬──────────────┘
                     │
                  WebSocket
                     │
          ┌──────────┼──────────┐
          ↓          ↓          ↓
       CLIENT      CLIENT     CLIENT
```

### Prinsip

- Host adalah source of truth.
- State Room aktif hidup di memori host; data SPH tersimpan di SQLite Host.
- Client tidak boleh membuka SQLite Host secara langsung.
- Client mengirim command/event ke Host.
- Host memvalidasi, menyimpan, lalu broadcast perubahan.
- Client menerima perubahan melalui WebSocket.

---

## 10.3 — Teknologi

Gunakan:

### Transport

**WebSocket** (`gorilla/websocket`) untuk sinkronisasi real-time.

### Discovery

**UDP broadcast** pada jaringan lokal (stdlib Go, tanpa dependensi tambahan).

Host mengumumkan Room aktif via broadcast; client mendengarkan dan menampilkan daftar Room yang ditemukan.

### Fallback

Discovery otomatis WAJIB memiliki fallback:

```text
Join via IP
```

Contoh:

```text
Host IP:
192.168.1.10

Port:
48765
```

---

## 10.4 — Room

Buat konsep **Collaboration Room**.

Contoh:

```text
Room Name:
SPH KRI BAC-593

Room Code:
BAC593

Access Code:
4827

Host:
OFFICE-PC
```

Room minimal memiliki:

```text
room_id            (uuid)
sph_document_id    (draft SPH yang dikerjakan bersama)
room_code
room_name
host_client_id
host_device_name
status
created_at
updated_at
```

Status:

```text
WAITING
ACTIVE
CLOSED
```

Room bersifat sementara dan hidup di memori host (lihat Keputusan Implementasi).
Ketika Host menutup Room:

```text
ROOM CLOSED
```

Database SPH tetap tersimpan.

---

## 10.5 — Participant

Setiap komputer yang terhubung harus memiliki:

```text
participant_id
display_name
device_name
role
joined_at
last_seen
status
```

Role MVP:

```text
HOST
EDITOR
VIEWER
```

Minimal implementasikan `HOST` dan `EDITOR`.

---

## 10.6 — Identitas User Tanpa Login Online

Tidak perlu account cloud.

Saat Join:

```text
Nama:
[Rizaldi]

Nama Komputer:
[OFFICE-PC-02]
```

Gunakan identity/session ID lokal.

---

## 10.7 — SPH Collaboration State

Jangan sinkronkan seluruh database aplikasi.

Hanya sinkronkan state SPH/Room aktif:

```text
SPH metadata
Main Point
Sub Point
Order
Quantity
Unit
Jasa
Material
Price
Weight
Notes
Status
Revision
```

Master data tidak otomatis disinkronkan ke semua client — arsitektur snapshot dokumen (BR-01) membuat client cukup bekerja dengan snapshot yang ada pada SPH itu sendiri.

Hanya dokumen berstatus **DRAFT** yang dapat diedit dalam Room (BR-08).

---

## 10.8 — Host sebagai Source of Truth

```text
CLIENT
  ↓
COMMAND / EVENT
  ↓
HOST
  ↓
VALIDATION
  ↓
BUSINESS LOGIC
  ↓
SQLite Transaction
  ↓
COMMIT
  ↓
BROADCAST
  ↓
ALL CLIENTS
```

Jangan mengizinkan Client menulis langsung ke database Host.

Operasi granular (tambah item, ubah harga, dst.) dipetakan ke service layer existing agar roll-up BR-01, pembulatan BR-04, alokasi pembobotan, dan validasi BR-08 selalu dijalankan di host.

---

## 10.9 — Event Model

Minimal dukung:

```text
ROOM_CREATED
ROOM_JOINED
ROOM_LEFT
USER_CONNECTED
USER_DISCONNECTED

SPH_LOADED
SPH_UPDATED

ITEM_ADDED
ITEM_UPDATED
ITEM_DELETED
ITEM_MOVED

SUB_ITEM_ADDED
SUB_ITEM_UPDATED
SUB_ITEM_DELETED
SUB_ITEM_MOVED

PRICE_UPDATED
QUANTITY_UPDATED
WEIGHT_UPDATED

SPH_STATUS_CHANGED

SYNC_REQUEST
SYNC_RESPONSE

ERROR
PING
PONG
```

---

## 10.10 — Message Envelope

```json
{
  "messageId": "uuid",
  "roomId": "uuid",
  "clientId": "uuid",
  "timestamp": "2026-08-25T01:50:00+07:00",
  "type": "ITEM_UPDATED",
  "version": 42,
  "payload": {}
}
```

Aturan:

- `messageId` unik (`google/uuid`);
- `roomId` wajib untuk event Room;
- `clientId` mengidentifikasi pengirim;
- `type` menentukan event;
- `version` untuk sinkronisasi (versi Room setelah operasi diterapkan);
- `payload` berisi data event.

---

## 10.11 — Versioning State

Gunakan sequence/version Room.

```text
Version 40
Version 41
Version 42
Version 43
```

Setiap perubahan yang diterima Host menaikkan version.

Jika Client tertinggal:

```text
Client Version = 41
Host Version = 45
```

maka Client harus melakukan `SYNC`.

---

## 10.12 — Initial Synchronization

```text
CLIENT
  ↓
JOIN ROOM
  ↓
HOST
  ↓
AUTHENTICATE
  ↓
SEND CURRENT STATE
  ↓
CLIENT APPLY STATE
  ↓
CLIENT READY
```

Hanya state Room/SPH aktif yang dikirim.

---

## 10.13 — Reconnect

State connection:

```text
CONNECTED
RECONNECTING
DISCONNECTED
```

UI:

```text
🟢 Terhubung
🟡 Menghubungkan kembali...
🔴 Terputus
```

Setelah reconnect:

```text
CLIENT
  ↓
RECONNECT
  ↓
SYNC_REQUEST
  ↓
HOST
  ↓
CURRENT STATE
  ↓
CLIENT
```

---

## 10.14 — Heartbeat

Gunakan:

```text
PING
PONG
```

Interval wajar, misalnya 5–10 detik.

---

## 10.15 — Presence

Tampilkan user yang online:

```text
WORK TOGETHER

● Rizaldi
  Host

● Fajar
  Editing Repair PLC

● Admin
  Editing Repair AMS

● Andi
  Viewing
```

---

## 10.16 — Active Editing Indicator

Jika memungkinkan tampilkan area yang sedang diedit:

```text
Fajar sedang mengedit:
Repair PLC
```

MVP tidak perlu field-level lock yang kompleks.

---

## 10.17 — Conflict Handling

Untuk MVP jangan menggunakan CRDT/OT kompleks.

Gunakan:

> **Host-authoritative, last accepted operation**

Contoh:

```text
PC A → Harga Rp10.000.000
PC B → Harga Rp12.000.000
```

Host menerima event sesuai urutan yang diterima dan memproses validasi.

Simpan:

```text
last_modified_by
last_modified_at
```

---

## 10.18 — Optimistic UI

Client boleh memperlihatkan perubahan secara langsung agar responsif.

```text
USER EDIT
  ↓
LOCAL UI UPDATE
  ↓
SEND EVENT
  ↓
HOST VALIDATE
```

Jika Host menolak, rollback/apply server state.

---

## 10.19 — Activity Log

Tampilkan:

```text
Activity

01:52 Rizaldi menambahkan "Testing"
01:53 Fajar mengubah harga menjadi Rp15.000.000
01:54 Admin memindahkan "Repair PLC"
01:55 Andi menambahkan Sub-Pekerjaan
```

Aktivitas juga dicatat ke audit log host (BR-13) dengan aktor = display_name kolaborator.

---

## 10.20 — Work Together UI

Tambahkan menu:

```text
Work Together
```

Contoh:

```text
┌────────────────────────────────────────┐
│ Work Together                          │
│                                        │
│ [ + Create Room ]                      │
│                                        │
│ Available Rooms                        │
│                                        │
│ ● SPH KRI BAC-593                      │
│   Host: OFFICE-PC                      │
│   Users: 3                             │
│   [ Join ]                             │
│                                        │
│ ● Repair Control Panel                 │
│   Host: ADMIN-PC                       │
│   Users: 2                             │
│   [ Join ]                             │
│                                        │
│ [ Join via IP ]                        │
└────────────────────────────────────────┘
```

---

## 10.21 — Create Room Flow

```text
Work Together
    ↓
Create Room
    ↓
Pilih draft SPH existing*
    ↓
Masukkan Room Name
    ↓
Generate Room Code
    ↓
Masukkan / Generate Access Code
    ↓
Start Room
```

\* Keputusan MVP: Room dibuat dari **draft SPH yang sudah ada**; tombol "Buat SPH Baru" mengarahkan ke wizard normal, setelah simpan kembali ke dialog Create Room.

---

## 10.22 — Join Room Flow

```text
Work Together
    ↓
Available Rooms
    ↓
Join
    ↓
Masukkan Access Code
    ↓
Connect
    ↓
Initial Sync
    ↓
SPH Loaded
```

---

## 10.23 — Manual IP Join

```text
Host:
192.168.1.10

Port:
48765

Access Code:
4827

[ Connect ]
```

---

## 10.24 — LAN Discovery

Host melakukan LAN discovery broadcast saat Room aktif.

Client menampilkan daftar Room lokal yang ditemukan.

Discovery hanya boleh berjalan pada jaringan lokal.

---

## 10.25 — Security LAN

Minimal gunakan:

```text
Room Code
Access Code / Session Token
```

Jangan membuat Room otomatis terbuka ke seluruh LAN tanpa proteksi.

---

## 10.26 — Windows Firewall

Ketika Host mengaktifkan Room, tampilkan pesan jika akses LAN kemungkinan diblokir firewall.

Catatan Windows: prompt resmi izin jaringan muncul otomatis dari OS saat aplikasi pertama kali membuka port — aplikasi tidak dapat menekan tombolnya sendiri (butuh UAC). Aplikasi cukup menampilkan panduan bila bind gagal:

```text
Kolaborasi LAN membutuhkan akses jaringan lokal.

Izinkan "Ganesha SPH Manager" pada jaringan privat bila Windows menanyakan,
atau tambahkan rule inbound TCP <port> secara manual.

[ Tutup ]
```

Jangan meminta akses internet bila tidak diperlukan.

---

## 10.27 — Host Disconnect

Untuk MVP jika Host keluar atau mati:

```text
ROOM DISCONNECTED
```

Client menampilkan Room disconnected/closed.

Data yang sudah committed tetap aman di SQLite Host.

Arsitektur harus memungkinkan future feature `Host Transfer`.

---

## 10.28 — Future Host Transfer

Belum perlu diimplementasikan pada Phase 10, tetapi desain harus memungkinkan:

```text
HOST OFFLINE
    ↓
Election
    ↓
CLIENT B menjadi HOST
    ↓
ROOM CONTINUE
```

---

## 10.29 — Collaboration Save Strategy

Setiap operation valid:

```text
EVENT
  ↓
VALIDATE
  ↓
BUSINESS LOGIC
  ↓
TRANSACTION
  ↓
COMMIT
  ↓
BROADCAST
```

Jangan menunggu tombol Save manual untuk perubahan collaboration — editor SPH dalam mode LIVE menyembunyikan tombol Simpan karena tiap operasi langsung commit di host.

---

## 10.30 — Database Strategy

DILARANG menggunakan shared SQLite/network share.

Jangan:

```text
Client → network share → SQLite Host
```

Gunakan:

```text
HOST
  SQLite local
  +
  Collaboration Service

CLIENT
  Local application state
  +
  WebSocket connection
```

---

## 10.31 — Service Architecture

Sesuai konvensi repo ini:

```text
internal/
└── collaboration/
    ├── room_service.go
    ├── session_service.go
    ├── websocket_server.go
    ├── websocket_client.go
    ├── event_service.go
    ├── sync_service.go
    ├── presence_service.go
    ├── discovery_service.go
    ├── messages.go
    └── types.go

internal/services/
└── sph_collab_ops.go       # operasi granular SPH utk kolaborasi

app_collab.go               # binding Wails
```

Frontend (mengikuti konvensi single-file store/page):

```text
frontend/src/
├── components/collaboration/
├── pages/WorkTogetherPage.vue
├── stores/collaboration.ts
└── router/index.ts         # route baru /work-together
```

Jangan meletakkan seluruh logic collaboration dalam `main.go`.

---

## 10.32 — Internal Collaboration API

Binding internal minimal:

```go
CreateRoom()
JoinRoom()
LeaveRoom()
CloseRoom()
BroadcastEvent()
SyncState()
GetRoom()
GetParticipants()
GetConnectionStatus()
HandleReconnect()
SendCollabOp()
ListDiscoveredRooms()
```

Sesuaikan dengan arsitektur project existing.

---

## 10.33 — Collaboration State Store

Frontend gunakan store khusus:

```text
collaborationStore
```

Minimal:

```text
room
connectionStatus
participants
currentVersion
currentHost
activities
lastSync
```

---

## 10.34 — SPH Store Integration

Solo mode:

```text
SPH Store
   ↓
Local Backend
```

Collaboration mode:

```text
SPH Store
   ↓
Collaboration Store
   ↓
Host
```

Business rule tetap berada di backend.

---

## 10.35 — Mode Indicator

Solo:

```text
OFFLINE / SOLO
```

Collaboration:

```text
🟢 LIVE
Room: SPH KRI BAC-593
Users: 4
```

User harus selalu tahu mode aktif.

---

## 10.36 — Offline Internet Test

Wajib test:

```text
Internet OFF
LAN ON
```

Kemudian:

```text
Host Create Room
Client Join
Edit SPH
Client lain menerima perubahan
```

Semua harus tetap bekerja.

---

## 10.37 — Multi-PC Test

Minimal:

```text
PC A = Host
PC B = Client
PC C = Client
```

Test:

```text
PC A Create Room
PC B Join
PC C Join

PC B tambah item
PC A menerima
PC C menerima

PC C edit harga
PC A menerima
PC B menerima
```

---

## 10.38 — Reconnect Test

```text
PC B Join
↓
Putuskan koneksi LAN sementara
↓
Status Reconnecting
↓
Aktifkan LAN
↓
Reconnect
↓
SYNC
↓
State sama dengan Host
```

---

## 10.39 — Conflict Test

Buat dua Client mengedit data yang sama hampir bersamaan.

Pastikan:

- Host menentukan hasil;
- semua Client akhirnya konsisten;
- version meningkat;
- tidak ada database corruption.

---

## 10.40 — Room Closure Test

Host menutup Room.

Semua Client harus menerima:

```text
Room Closed
```

SPH tetap tersimpan.

---

## 10.41 — Crash Test

Host crash setelah ada beberapa operation.

Setelah restart:

```text
Database valid
SPH valid
Data terakhir yang sudah committed tetap ada
```

---

## 10.42 — Acceptance Criteria

Phase 10 dianggap selesai jika:

```text
[ ] Create Room
[ ] Join Room
[ ] LAN Discovery
[ ] Manual IP Join
[ ] Access Code
[ ] Host
[ ] Client
[ ] WebSocket
[ ] Initial Sync
[ ] Real-Time Update
[ ] Presence
[ ] Activity Log
[ ] Reconnect
[ ] Versioning
[ ] Host Authority
[ ] SQLite Persistence
[ ] Offline Internet Test
[ ] Multi-PC Test
[ ] Error Handling
[ ] Unit Tests
[ ] Integration Tests
[ ] Documentation
```

Jangan menandai item `PASS` tanpa benar-benar mengujinya.

---

## 10.43 — Hal yang Dilarang

Jangan gunakan:

```text
Firebase
Supabase
Hosted WebSocket
Cloud Database
VPS
Domain
Online Authentication
Network Shared SQLite
```

Jangan:

- membuka SQLite Host melalui network share;
- membuat Client menulis database Host secara langsung;
- menyinkronkan seluruh database perusahaan;
- menggunakan internet untuk collaboration;
- membuat Collaboration Service menjadi aplikasi server terpisah untuk MVP.

---

## 10.44 — Future Roadmap

Setelah Phase 10 stabil, pertimbangkan:

```text
Phase 10.1 — Host Transfer
Phase 10.2 — Advanced Conflict Resolution
Phase 10.3 — Offline Operation Queue
Phase 10.4 — Undo / Redo Collaboration
Phase 10.5 — File Collaboration
Phase 10.6 — Local Asset Transfer
Phase 10.7 — Secure LAN Pairing
```

Jangan implementasikan semuanya pada MVP Phase 10.

---

## 10.45 — Checklist Eksekusi

Saat mulai Phase 10:

1. Baca dokumentasi Phase 0–9.
2. Audit arsitektur existing.
3. Pastikan Solo Mode tetap bekerja.
4. Tambahkan Collaboration Service secara modular.
5. Implementasikan Room.
6. Implementasikan Host/Client.
7. Implementasikan WebSocket.
8. Implementasikan LAN discovery (UDP broadcast).
9. Implementasikan manual IP fallback.
10. Implementasikan access token/access code.
11. Implementasikan initial sync.
12. Implementasikan event broadcasting.
13. Implementasikan persistence (data SPH via service layer).
14. Implementasikan presence.
15. Implementasikan activity log.
16. Implementasikan reconnect.
17. Implementasikan versioning.
18. Implementasikan error handling.
19. Buat unit test.
20. Buat integration test.
21. Test minimal dua komputer.
22. Test dengan internet dimatikan.
23. Test firewall behavior.
24. Update dokumentasi.
25. Build aplikasi.

Jangan merusak fitur Phase sebelumnya.

---

## 10.46 — Final Report Phase 10

Jika seluruh test benar-benar lulus, tampilkan:

```text
PHASE 10 COMPLETE

Room: PASS
LAN Discovery: PASS
Manual IP: PASS
Multi Client: PASS
Real-Time Sync: PASS
Presence: PASS
Activity Log: PASS
Reconnect: PASS
Versioning: PASS
Offline Internet: PASS
SQLite Persistence: PASS
Unit Tests: PASS
Integration Tests: PASS
Build: PASS
Documentation: PASS

Known Issues:
...

Files Changed:
...

Next Recommended Phase:
...
```

Jangan menyatakan `PASS` jika test belum dijalankan.

---

## 10.47 — Hasil Akhir yang Diharapkan

User harus dapat melakukan:

```text
PC ADMIN
    ↓
SPH Manager
    ↓
Work Together
    ↓
Create Room
    ↓
SPH KRI BAC-593
    ↓
PC ADMIN menjadi HOST

                 LAN

PC 02 → Join
PC 03 → Join
PC 04 → Join
```

Kemudian:

```text
PC 02:
Tambah pekerjaan
        ↓
HOST:
Validasi + Commit
        ↓
Broadcast
        ↓
PC 03:
Pekerjaan muncul
        ↓
PC 04:
Pekerjaan muncul
```

Semuanya berjalan melalui jaringan lokal dan **tanpa internet**.

---

## 10.48 — Prinsip Utama Phase 10

> **Kolaborasi harus terasa real-time bagi pengguna, tetapi tetap sederhana, aman, dan sepenuhnya lokal.**

Arsitektur inti:

```text
Wails
 +
Vue
 +
Go
 +
SQLite
 +
LAN WebSocket
 +
Host-Authoritative State
```

Tidak diperlukan infrastruktur online.
