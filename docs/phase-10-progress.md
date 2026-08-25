# Phase 10 — Kolaborasi Real-Time LAN: Progress & Rencana Lanjutan

> **Terakhir diperbarui:** 2026-08-25
> **Status:** Backend ✅ kompilasi bersih · Frontend ✅ kompilasi bersih · Tests ✅ hijau

---

## 1. Pencapaian Backend (Sesi Ini)

### File yang sudah dibuat / dimodifikasi

| File | Status | Keterangan |
|---|---|---|
| `internal/services/errors.go` | MODIFIKASI | +`IsFriendly()` (cek ValidationError/ConflictError) |
| `internal/services/settings_service.go` | MODIFIKASI | +`CollabDefaults` struct, +`CollabPort`/`CollabDisplayName` di SettingsView/Input, +`validCollabPort`, +`CollabPortOrDefault()`/`CollabDisplayNameOrDefault()` helpers |
| `internal/services/audit.go` | MODIFIKASI | +`WriteAs(db, action, entity, id, desc, actor)` (tambahan parameter `actor`) |
| `internal/services/sph_service.go` | MODIFIKASI | +`RoomGuard` interface, +`SetRoomGuard(g)`, +`docLocked(id)` helper, guard checks di `UpdateDraft`, `Delete`, `Duplicate`, `CreateRevision`, `SetStatus`; `applyDraftUpdate` diekstrak dari `UpdateDraft` |
| `internal/services/sph_collab_ops.go` | BARU | Op type constants (`OpHeaderUpdated`, dll.), `HeaderPatch`, `ItemFields`, `SubItemFields`, `OpPayload`, `CollabActivity`, `CollabOps` service — dispatch operasi granular, `Snapshot()`, `DocToSaveInput()`, helpers |
| `internal/models/models.go` | MODIFIKASI | +`Actor string` field di `AuditLog` (AutoMigrate additif) |
| `internal/collaboration/types.go` | BARU | `Participant`, `RoomInfo`, `DiscoveredRoom` structs; constants: roles, modes, connection states |
| `internal/collaboration/messages.go` | BARU | `Envelope`, `JoinRequest`, `StatePayload`, `ErrorPayload`, `ClosedPayload`; msg type constants; error code constants; `newEnvelope`/`envelopeWith` helpers |
| `internal/collaboration/helpers.go` | BARU | `GenerateRoomCode`, `GenerateAccessCode`, `sanitizeIdentity`, `statusLabelID`, `equalConstTime`, `cloneParticipants`, `cloneActivities`, `cloneRoomInfo`, `rejectUnjoined`, `sortDiscoveredByNewest` |
| `internal/collaboration/hub.go` | BARU | ~1130 baris — `Manager` (HostRoom, Join, LeaveClientSession, SendOp, Session, Shutdown, StartDiscovery, StopDiscovery, DiscoveredRooms, SetEmit, RoomGuard impl), `Room` struct (in-memory, goroutine-per-room, WS server loop, applyLocal, applyAndBroadcastLocked, heartbeat, broadcastPresence, close) |
| `internal/collaboration/server.go` | BARU | `wsServer` (HTTP→WS upgrader, port listener), `serverConn` (writePump, deliver, shutdown, rejectUnjoined) |
| `internal/collaboration/discovery.go` | BARU | `announcePacket`, `Announcer` (UDP broadcast tiap 2 detik), `Listener` (read UDP, prune TTL 10s) |
| `internal/collaboration/client.go` | BARU | `Client` (outbound WS, auto-reconnect backoff 500ms→10s, heartbeat, readLoop, `StartAndWaitReady`, `sendOp`, `stop`/`stopQuiet`); `connSession` (kematian via `sync.Once`) |
| `app_collab.go` | BARU | 10 Wails bindings: `GetCollabDefaults`, `CreateCollabRoom`, `CloseCollabRoom`, `StartDiscoveryListener`, `StopDiscoveryListener`, `ListDiscoveredRooms`, `JoinCollabRoom`, `LeaveCollabRoom`, `SendCollabOp`, `GetCollabSession` |
| `app.go` | MODIFIKASI | +`collabMgr *collaboration.Manager` field, +wiring di `NewApp` (CollabOps + Manager + RoomGuard), +`wireCollab()` di startup, +`Shutdown` di shutdown; versi → `0.10.0` |

### Build status

```
go build ./...   ✅ (clean)
go vet ./...     ✅ (clean)
gofmt -l         ✅ (formatted)
go test ./internal/... ✅ (semua hijau)
```

### Arsitektur backend (alur data)

```
Frontend (Vue)
  │
  ├─ SendCollabOp(OpPayload)  ──── app_collab.go ──→ Manager.SendOp()
  │                                                   │
  │                                                   ├─ Client.sendOp() [CLIENT mode]
  │                                                   │   └── WS → host room
  │                                                   │
  │                                                   └─ Room.applyAndBroadcastLocked() [HOST mode]
  │                                                       ├─ CollabOps.Apply(docID, actor, op)
  │                                                       │   └─ dispatch → rebuild SphSaveInput
  │                                                       ├─ applyDraftUpdate(docID, saveInput, actor)
  │                                                       │   └─ UpdateDraft → DB
  │                                                       ├─ broadcastLocked(StatePayload{doc, participants, activity})
  │                                                       └─ touchParticipant() → presence
  │
  └─ EventsOn("collab:sync")  ←──── Manager.emitFn ── runtime.EventsEmit
```

### Guard hook (BR-08/BR-16 dokumen solo)

```go
// sph_service.go
type RoomGuard interface {
    DocLocked(id uint) bool
}

// App implementasi: Manager.DocLocked(id) → true bila ada room hidup dengan docID ini
// → UpdateDraft/Delete/Duplicate/CreateRevision/SetStatus ditolak dengan pesan:
//   "Dokumen ini sedang dikolaborasi secara real-time. Tutup room terlebih dahulu."
```

---

## 2. Rencana Frontend (Belum Dikerjakan)

### Struktur file yang akan dibuat/diubah

| # | File | Aksi | Keterangan |
|---|---|---|---|
| 1 | `src/types/collaboration.ts` | BUAT | TS types: `CollabSnapshot`, `RoomInfo`, `Participant`, `DiscoveredRoom`, `CollabActivity`, `OpPayload`, `CollabDefaults` |
| 2 | `src/stores/collaboration.ts` | BUAT | Pinia setup store — state `snapshot`, getters (`isLive`/`isHost`/`modeLabel`), actions (`createRoom`/`joinRoom`/`leaveRoom`/`sendOp`/`listDiscovered`/dll.) |
| 3 | `src/composables/useCollabSync.ts` | BUAT | Wire `EventsOn('collab:sync')` → store.applySnapshot; `sendItemOp()`/`sendHeaderOp()` helpers; rollback on ERROR |
| 4 | `src/pages/WorkTogetherPage.vue` | BUAT | Lobby: mode badge, discovered rooms list, Create Room, Join via IP; saat LIVE tampilkan CollabToolbar |
| 5 | `src/components/collaboration/CollabToolbar.vue` | BUAT | Mode badge `🟢 LIVE`, connection status dot, presence list, activity log, tombol Keluar |
| 6 | `src/components/collaboration/CreateRoomDialog.vue` | BUAT | Wizard AppModal: pick draft SPH → room name → access code → Start Room |
| 7 | `src/components/collaboration/JoinDialog.vue` | BUAT | AppModal: display name, device name, access code → Gabung |
| 8 | `src/components/collaboration/JoinManualDialog.vue` | BUAT | AppModal: host IP, port, display name, device name, access code → Hubungkan |
| 9 | `src/router/index.ts` | MODIFIKASI | +route `/work-together` → WorkTogetherPage |
| 10 | `src/components/SideNav.vue` | MODIFIKASI | +"Work Together" di section "Lainnya" dengan icon `users` |
| 11 | `src/components/TopBar.vue` | MODIFIKASI | Saat `collaborationStore.isLive` tampilkan `🟢 LIVE · RoomName · N users` menggantikan badge "Terhubung" |
| 12 | `wailsjs/go/main/App.d.ts` + `App.js` | REGENERATE | Jalankan `wails generate module` — 10 binding collab baru muncul |

### Conventions frontend yang harus diikuti

- Semua komponen pakai `<script setup lang="ts">`
- Store pakai Composition API `defineStore('name', () => { ... })`
- Text: `text-[13px]` untuk body, `text-sm font-semibold` untuk heading
- Container: `rounded-xl border border-slate-200 bg-white`
- Primary btn: `bg-brand-600 hover:bg-brand-700 text-white`
- Secondary btn: `border border-slate-200 text-slate-600 hover:bg-slate-50`
- Modal: re-use `AppModal.vue` + `ConfirmDialog.vue`
- Errors: `rounded-lg border border-red-200 bg-red-50 px-4 py-2.5 text-[13px] text-red-700`
- Wails events: `EventsOn('event', cb)` di composable, `EventsOff('event')` di `onUnmounted`
- Icon SVG inline Heroicons, path string di object `icons`

### UX flow utama

**Create Room (Host):**
```
Work Together → [+ Mulai Room Baru]
  → Step 1: Pick draft SPH (list DRAFT saja)
  → Step 2: Room Name + Access Code
  → "Mulai Room"
  → Navigate ke /sph/{id} (LIVE mode, Save button hidden)
```

**Join Room (Client - Auto Discovery):**
```
Work Together → [Join] pada room card
  → Dialog: Display Name + Device Name + Access Code
  → "Gabung"
  → Navigate ke /sph/{id} (LIVE mode)
```

**Join Room (Client - Manual IP):**
```
Work Together → [Gabung via IP...]
  → Dialog: Host IP + Port + Name + Device + Access Code
  → "Hubungkan"
  → Navigate ke /sph/{id} (LIVE mode)
```

**Real-time editing:**
```
User edits di /sph/{id}
  → useCollabSync.sendItemOp(type, itemId, itemFields)
  → collaborationStore.sendOp(OpPayload) → Go backend
  → Host validates + saves + broadcasts
  → EventsOn('collab:sync') → store.applySnapshot → UI updates
```

### Go types → TypeScript mapping

```typescript
// types/collaboration.ts

interface CollabSnapshot {
  mode: '' | 'HOST' | 'CLIENT'
  connection: '' | 'CONNECTED' | 'RECONNECTING' | 'DISCONNECTED'
  room?: RoomInfo
  doc?: SphSaveInput              // full SPH state dari host
  participants?: Participant[]
  activities?: CollabActivity[]
  version?: number
  error?: string
  notice?: string
}

interface RoomInfo {
  roomId: string
  sphDocumentId: number
  documentNumber: string
  projectName: string
  roomCode: string
  roomName: string
  accessCode?: string             // host-only
  hostName: string
  hostDevice: string
  port: number
  status: string                  // 'ACTIVE' | 'CLOSED'
  version: number
  participants?: Participant[]
  createdAt: string
}

interface Participant {
  id: string
  displayName: string
  deviceName: string
  role: string                    // 'HOST' | 'EDITOR'
  joinedAt: string
  lastSeen: string
}

interface DiscoveredRoom {
  roomId: string
  roomName: string
  documentNumber: string
  projectName: string
  hostName: string
  port: number
  users: number
  lastSeen: string
}

interface CollabActivity {
  actor: string
  action: string
  summary: string
}

interface OpPayload {
  type: string                    // 'header_updated' | 'item_added' | 'item_updated' | dll.
  itemId?: number
  subItemId?: number
  toIndex?: number
  header?: HeaderPatch
  item?: ItemFields
  subItem?: SubItemFields
}

interface HeaderPatch {
  date: string
  customerId: number
  vesselId?: number
  projectName: string
  subject: string
  reference: string
  location: string
  validUntil: string
  picName: string
  notes: string
}

interface ItemFields {
  workItemId?: number
  name: string
  description: string
  quantity: number
  unit: string
  serviceUnitPrice: number
  materialUnitPrice: number
  pricingMode: string
  notes: string
}

interface SubItemFields {
  name: string
  description: string
  quantity: number
  unit: string
  weight: number
  serviceUnitPrice: number
  materialUnitPrice: number
  notes: string
}

interface CollabDefaults {
  deviceName: string
  port: number
  displayName: string
}
```

---

## 3. Testing Plan

### Backend tests (belum ditulis)

Buat file `internal/collaboration/hub_test.go`:

1. **Unit — helpers**: `GenerateRoomCode` (6 chars, unique), `GenerateAccessCode` (6 digits), `sanitizeIdentity` (trim, max 100), `cloneParticipants` (independent copy)
2. **Unit — messages**: `envelopeWith` produces valid UUID, timestamp, type; `StatePayload` marshal/unmarshal roundtrip
3. **Unit — CollabOps**: `Apply` header_updated, item_added, item_updated, item_deleted, sub_item_added, sub_item_updated, sub_item_deleted, item_moved, sub_item_moved
4. **Integration — in-process**: host localhost + 2 WS clients → join → initial sync → send op → convergence → reconnect after drop → close room → clients notified
5. **Integration — RoomGuard**: buka room untuk doc X → UpdateDraft(docX) ditolak → tutup room → UpdateDraft(docX) berhasil
6. **Integration — discovery**: announcer + listener same port → room muncul di `Rooms()`

### Frontend tests (belum ditulis)

1. Unit: `stores/collaboration.ts` — snapshot handling, getters, action delegates
2. Unit: `composables/useCollabSync.ts` — OpPayload building
3. E2E manual: Create room → join dari 2 PC → edit bersama → tutup room

---

## 4. Urutan Eksekusi Lanjutan

### Sesi berikutnya (Frontend)

1. ✅ Jalankan `wails generate module` → regenerate `wailsjs/go/main/App.d.ts`
2. ✅ Buat `frontend/src/types/collaboration.ts`
3. ✅ Buat `frontend/src/stores/collaboration.ts`
4. ✅ Buat `frontend/src/composables/useCollabSync.ts`
5. ✅ Buat `frontend/src/components/collaboration/CreateRoomDialog.vue`
6. ✅ Buat `frontend/src/components/collaboration/JoinDialog.vue`
7. ✅ Buat `frontend/src/components/collaboration/JoinManualDialog.vue`
8. ✅ Buat `frontend/src/components/collaboration/CollabToolbar.vue`
9. ✅ Buat `frontend/src/pages/WorkTogetherPage.vue`
10. ✅ Modifikasi `src/router/index.ts` (tambah route)
11. ✅ Modifikasi `src/components/SideNav.vue` (tambah nav item)
12. ✅ Modifikasi `src/components/TopBar.vue` (LIVE badge)
13. ✅ `cd frontend && npm run build` — fix TypeScript errors
14. ✅ Test manual

### Sesi sesudahnya (Tests)

15. ✅ Buat `internal/collaboration/hub_test.go`
16. ✅ `go test ./internal/collaboration/ -v` — 15 test hijau
17. ✅ `go test ./internal/... -count=1` — semua hijau
18. ✅ `go vet ./...` + `gofmt` — bersih
19. ✅ Update `docs/development-plan.md` status Phase 10
20. ✅ Update `docs/collaboration-lan.md` (centang §10.38–10.48 checklist)

---

## 5. Referensi Cepat

- **Spesifikasi lengkap**: `docs/collaboration-lan.md` (§10.1–10.48, 1231 baris)
- **Development plan**: `docs/development-plan.md` (Phase 10 deskripsi)
- **Business rules**: `docs/business-rules.md`
- **Architecture**: `docs/architecture.md`
- **Port default**: 48765 (WS) + 48766 (UDP broadcast)
- **Event Wails**: `collab:sync` → `UISnapshot`
- **Kedudukan di sidebar**: section "Lainnya", setelah "Pengaturan"
- **Kedudukan di router**: `/work-together`
