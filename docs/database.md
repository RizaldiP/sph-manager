# Database — SPH Manager Offline

> SQLite (WAL) via GORM + driver pure-Go `glebarez/sqlite` (tanpa CGO). Lokasi: `%AppData%\sph-manager\database\sph.db`.

## Konfigurasi Koneksi

| Pragma | Nilai | Alasan |
|---|---|---|
| `foreign_keys` | ON | Integritas relasi |
| `journal_mode` | WAL | Baca/tulis cepat, aman crash |
| `busy_timeout` | 5000 ms | Hindari error lock saat transaksi bertumpuk |

## Diagram Relasi

```text
categories ──< work_items ──< work_sub_items
templates ──< template_items >── work_items
customers ──< vessels
customers ──< sph_documents >── vessels (opsional)
sph_documents ──< sph_items ──< sph_sub_items
sph_documents ──< sph_revisions        (> = referensi opsional/RESTRICT)
settings (key-value) · audit_logs (append-only)
```

## 14 Tabel

### Master Pekerjaan
| Tabel | Isi | Catatan |
|---|---|---|
| `categories` | Kategori pekerjaan | kode unik (parsial), soft delete |
| `work_items` | Master pekerjaan | harga default jasa/material integer Rupiah, qty float |
| `work_sub_items` | Sub-pekerjaan | `sequence` urutan tampil, `difficulty_weight` bobot default |

### Template & Master Data
| Tabel | Isi | Catatan |
|---|---|---|
| `templates` | Kumpulan pekerjaan siap pakai | kode unik parsial |
| `template_items` | Anggota template | urut `sequence`, FK work item RESTRICT, CASCADE dari template |
| `customers` | Customer | PIC + posisi |
| `vessels` | Kapal per customer | unik (customer_id, code) parsial |
| `materials` | Material/suku cadang | harga default, supplier |

### Dokumen SPH
| Tabel | Isi | Catatan |
|---|---|---|
| `sph_documents` | Kepala dokumen SPH | unik (document_number, revision); status DRAFT…CANCELLED; subtotal & grand total; `terbilang`; `finalized_at` |
| `sph_items` | Main point (snapshot) | `name_snapshot`/`description_snapshot`; `pricing_mode` HARGA_LANGSUNG/PEMBOBOTAN; kolom jasa/material/total int64 |
| `sph_sub_items` | Sub point (snapshot) | `weight` % + `allocated_value` hasil pembobotan; CASCADE dari item |
| `sph_revisions` | Histori revisi | `from_document_id` menunjuk sumber salinan |

### Sistem
| Tabel | Isi |
|---|---|
| `audit_logs` | action/entity/entity_id/description/timestamp (append-only) |
| `settings` | key-value (company profile, penomoran, penandatangan, dll.) |

## Aturan Skema Penting

1. **Uang = INTEGER** (`int64`) Rupiah penuh; kuantitas = REAL.
2. **Soft delete**: semua master + sph_documents memakai `is_active` + `deleted_at` (GORM `DeletedAt`).
3. **Unique index parsial** (`WHERE deleted_at IS NULL`) agar kode bisa dipakai ulang setelah soft delete:
   - `uq_categories_code_live`, `uq_work_items_code_live`, `uq_customers_code_live`, `uq_materials_code_live`, `uq_templates_code_live`
   - `uq_vessels_customer_code_live` (customer_id, code)
   - `uq_sph_documents_number_rev_live` (document_number, revision)
   - `uq_settings_key`
4. **FK ON DELETE**:
   - `RESTRICT` untuk relasi master yang terpakai (kategori↔pekerjaan, customer↔kapal/dokumen, work_item↔dokumen/template).
   - `CASCADE` untuk detail dokumen (dok→item→sub-item, dok→revision) dan template→items.

## Migration

`internal/database.Migrate(db)`:

1. `AutoMigrate` 14 model dalam urutan dependensi (idempotent).
2. Buat partial unique index (`CREATE UNIQUE INDEX IF NOT EXISTS …`).

Helper verifikasi: `ExistingTables(db)`, `ForeignKeyViolations(db)` (PRAGMA foreign_key_check).

## Test Integrasi (`internal/database/database_test.go`)

| Test | Membuktikan |
|---|---|
| `TestMigrateCreatesAllTables` | 14 tabel terbentuk dari DB kosong |
| `TestMigrateIdempotent` | migrate berulang tidak error |
| `TestForeignKeyEnforced` | FK aktif; category_id hantu ditolak; pragma check bersih |
| `TestCategoryDeleteRestrictedWhenUsed` | soft delete selalu boleh; hard delete dipakai data = ditolak |
| `TestPartialUniqueIndexAllowsReuseAfterSoftDelete` | duplikat aktif ditolak; kode bekas hapus bisa dipakai |
| `TestSphDocumentUniqueNumberRevision` | nomor+revisi unik; revisi beda boleh |
| `TestSettingsUniqueKey` | key setting unik |
| `TestCascadeDocumentToItems` | hard delete dokumen menghapus item & sub item |
| `TestWorkSubItemOrderingAndDefaults` | urutan sub item tersimpan benar |

Jalankan: `go test ./internal/...`

## Contoh Data Mengikuti Referensi Excel

```text
categories: EL / Electrical
work_items: EL-001 "Repair Control Panel" (unit giat, harga default)
work_sub_items: 01 Inspection · 02 Troubleshooting · … (sequence)
sph_documents: SPH/GEI/VIII/2026/001 Rev 0 — DRAFT
sph_items: snapshot "Repair PLC" @10.000.000
```
