# Backup & Restore (Phase 11)

Database aplikasi (SQLite, `sph.db`) dilindungi dengan fitur **backup & restore**:
backup manual, backup otomatis harian + saat penutupan, retention 10 backup
terakhir, dan pemulihan yang aman dengan restart otomatis. Halaman **Backup**
tersedia di menu samping (`/backup`).

## Strategi Snapshot: `VACUUM INTO`

Backup dibuat dengan perintah SQLite `VACUUM INTO '<path>'` di atas database
yang sedang terbuka (mode **WAL**):

- Menghasilkan file `.db` baru yang **konsisten** tanpa perlu mengunci aplikasi
  atau menghentikan aktivitas; semua transaksi WAL ikut terangkum.
- Berkas output dikompaksi penuh (tanpa ruang kosong), sehingga ukurannya kecil.
- Tidak pernah menyentuh file database aktif.

Penamaan file mengikuti FR-B1: `SPH_Backup_YYYY-MM-DD_HHMMSS.db`

## Lokasi

| Item | Path |
| --- | --- |
| Database aktif | `%AppData%\sph-manager\database\sph.db` |
| Folder backup | `%AppData%\sph-manager\backups\` |

## Backup Manual (FR-B1)

Halaman **Backup → Buat Backup Sekarang**:

1. `BackupNow()` memanggil `BackupService.CreateManual`.
2. Snapshot `VACUUM INTO` dibuat dengan nama ber-timestamp.
3. Retention diberlakukan (hapus backup tertua bila sudah melebihi 10).
4. Aksi dicatat ke audit log (`BACKUP` / database).

## Backup Otomatis (FR-B3)

- **Saat aplikasi ditutup:** `App.shutdown` memanggil `BackupOnShutdown` sebelum
  database ditutup.
- **Harian:** saat aplikasi mulai, `EnsureDaily` langsung memeriksa; bila backup
  hari ini belum ada, langsung dibuat. Timer 24 jam memeriksa ulang selama
  aplikasi menyala.
- Aturan **paling banyak satu backup otomatis per hari** (apakah itu dibuat
  manual atau otomatis). Backup manual tidak dibatasi.
- **Retention:** hanya 10 snapshot terbaru yang disimpan (FR-B3). Lampaui batas =
  backup tertua terhapus.
- Backup ditangguhkan otomatis bila database sudah tertutup (kasus pasca-restore,
  agar tidak membuat file basi).

## Restore (FR-B2)

Alur di halaman **Backup → tombol Pulihkan**:

1. **Blokir bila sesi Work Together masih aktif** — tutup ruang kolaborasi dulu.
2. **Validasi file backup:** `PRAGMA integrity_check` (harus `ok`),
   `foreign_key_check` (tidak boleh ada yang melanggar), dan cek keberadaan
   tabel-tabel inti SPH (`categories`, `work_items`, `work_sub_items`,
   `templates`, `template_items`, `customers`, `vessels`, `materials`,
   `sph_documents`, `sph_items`, `sph_sub_items`, `audit_logs`, `settings`).
3. **Backup kondisi sekarang** dibuat otomatis sebagai pengaman (safety backup).
4. Database aktif **ditutup** (file `-wal`/`-shm` dibersihkan), lalu file hasil
   backup **disalin ke file sementara dan dipindahkan secara atomik** menggantikan
   `sph.db`.
5. Aksi **`RESTORE`** dicatat ke audit log di database yang baru dipulihkan.
6. **Aplikasi dimulai ulang otomatis** memakai executable yang sama — proses baru
   membuka database hasil restore dan menjalankan migrasi idempotent bila skema
   backup lebih tua. Proses lama ditutup lewat binding `QuitApp()`.

> Konsekuensi desain: restart dipilih daripada "swap in-process" karena seluruh
> service berbagi satu koneksi database; memulai ulang proses menjamin tidak ada
> service yang masih memegang database lama (reload paling aman).

### Bila restore gagal di tengah (rollback)

Gagal pada langkah pertukaran file = **file backup yang lama diretur** (dari
safety backup) lalu aplikasi ditutup. Saat dibuka kembali, data kembali ke
kondisi sebelum restore. Error dijelaskan pada banner UI.

## Mengimpor Backup dari Lokasi Lain (perangkat/USB)

Backup yang ada di luar folder backup aplikasi (misalnya dari laptop lain atau
flashdisk) bisa dibawa masuk tanpa menyalin manual:

1. Halaman **Backup → tombol Import Backup…** membuka dialog pilih file (filter `*.db`).
2. File dipilih **divalidasi dulu** (integrity_check + tabel inti SPH). Bila bukan
   database SPH yang sehat, ditolak dengan pesan jelas dan tidak ada file tersalin.
3. Bila valid, file **disalin** ke folder backup dengan nama `SPH_Backup_<timestamp>.db`
   (dua import di detik yang sama diberi akhiran `_1`, `_2`, …) lalu aksi tercatat
   sebagai audit `IMPORT`.
4. Daftar backup di-refresh → file siap dipulihkan lewat tombol **Pulihkan**.

Membatalkan dialog = tidak ada perubahan. Memilih file yang memang sudah berada
di folder backup tidak membuat duplikat. File `sph.db` aktif **tidak** disarankan
dipilih langsung karena jumlah versi/riwayatnya berbeda dari file backup.

## Operasi Lain

- **Hapus backup:** permanen; dikonfirmasi lewat dialog. File yang tidak
  mengikuti pola penamaan (mis. `catatan.txt`, path traversal `../…`) ditolak
  (fungsi `Resolve`).
- **Buka Folder Backup:** membuka folder `backups\` di file manager.

## Keamanan File

Nama backup divalidasi dengan regex ketat (`^SPH_Backup_\d{4}-\d{2}-\d{2}_\d{6}\.db$`).
Semua path berada di dalam folder backup (cek prefix absolut), mencegah
pembacaan/penghapusan/penimpaan file di luar folder.

## Arsitektur Kode

| Komponen | Peran |
| --- | --- |
| `internal/backup` | Util murni: `Name`, `Create`, `List`, `Retain`, `Validate`, `SwapMain` |
| `internal/services/backup_service.go` | Orkestrasi + audit + jadwal auto + penyimpanan pegangan DB |
| `app_backup.go` | Binding Wails: `BackupNow`, `ListBackups`, `DeleteBackup`, `OpenBackupFolder`, `RestoreBackup`, `QuitApp`; alur restart |
| `app.go` | Wiring service; `EnsureDaily` + ticker saat startup; `BackupOnShutdown` saat tutup |
| `frontend/src/pages/BackupPage.vue` | UI halaman Backup |
| `docs/backup-restore.md` | Dokumen ini |

## Status Implementasi

Sesuai rencana development-plan **Phase 11** — selesai dan sudah melewati
`go vet`, `go test ./internal/backup/... ./internal/services/...`, serta build
aplikasi (`wails build`).