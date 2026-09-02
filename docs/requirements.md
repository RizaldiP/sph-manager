# Kebutuhan Aplikasi — SPH Manager Offline

> PT. Ganesha Energi Indonesia — perusahaan repair & maintenance kapal.
> Sumber: Master Specification + analisis `SPH_KRI_OWA.xls` (lihat `excel-analysis.md`).
>
> Konvensi prioritas: **[MVP]** wajib ada di rilis pertama · **[FUTURE]** tidak dikerjakan di MVP, arsitektur harus siap.

---

## 1. Platform & Offline

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-P1 | Aplikasi desktop Windows, 100% offline (Wi-Fi/internet mati tetap jalan) | MVP |
| FR-P2 | Tanpa CDN, API eksternal, Google Fonts online, cloud storage, online auth — semua asset di-bundle | MVP |
| FR-P3 | Data tersimpan di folder AppData: `database/`, `backups/`, `exports/`, `logs/`, `templates/` (path OS-aware) | MVP |
| FR-P4 | Multi-user, cloud sync, approval, quotation tracking, PO, invoice, work order | FUTURE |

## 2. Master Data

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-M1 | CRUD Kategori pekerjaan (Electrical, Automation, Instrumentation, Mechanical, HVAC, Navigation, Communication, PLC, Control System, Testing & Commissioning, Other) | MVP |
| FR-M2 | CRUD Master Pekerjaan (kode, nama, deskripsi, satuan default, qty default, harga jasa/material default, catatan) | MVP |
| FR-M3 | CRUD Sub-Pekerjaan per pekerjaan (urutan bisa diubah, bobot kesulitan default, harga default) | MVP |
| FR-M4 | CRUD Template (kumpulan pekerjaan yang sering dipakai ulang): create, edit, duplicate, reorder, aktif/nonaktif | MVP |
| FR-M5 | CRUD Customer (kode, nama, alamat, telp, email, PIC, posisi PIC) | MVP |
| FR-M6 | CRUD Kapal per customer (kode, nama, nomor kapal, jenis) | MVP |
| FR-M7 | CRUD Material (kode, nama, satuan, harga default, supplier) | MVP |
| FR-M8 | Semua master data: search, aktif/nonaktif, soft delete (`is_active` + `deleted_at`) | MVP |

## 3. SPH

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-S1 | SPH Builder wizard 8 langkah: Info → Pilih Pekerjaan → Susun Main Point → Susun Sub Point → Costing → Validasi → Preview → Simpan/Generate | MVP |
| FR-S2 | Pilih banyak pekerjaan sekaligus (search + filter kategori + dari template + dari SPH lama + manual) → semua jadi Main Point | MVP |
| FR-S3 | Drag-and-drop urutan Main Point & Sub Point; urutan tersimpan | MVP |
| FR-S4 | Edit di dalam SPH: nama, deskripsi, qty, unit, jasa, material, notes, sub-item, weight — hanya berlaku untuk SPH tersebut (master tidak berubah) | MVP |
| FR-S5 | Snapshot: SPH tersimpan sebagai salinan data saat dibuat (lihat BR-01) | MVP |
| FR-S6 | Status dokumen: DRAFT, REVIEW, FINAL, SENT, ACCEPTED, REJECTED, CANCELLED | MVP |
| FR-S7 | Daftar SPH: kolom Nomor, Revision, Tanggal, Customer, Kapal, Project, Total, Status; fitur search, filter, sort, preview, edit draft, duplicate, export, archive | MVP |
| FR-S8 | Duplicate SPH → snapshot baru penuh, source tidak berubah | MVP |
| FR-S9 | Revision: SPH-001 Rev 0/1/2 — histori lama tetap bisa dilihat, tidak dioverwrite | MVP |
| FR-S10 | Informasi SPH: nomor, tanggal, customer, kapal, project, subject/perihal, reference, lokasi, masa berlaku, PIC, catatan — field wajib ditandai jelas | MVP |
| FR-S11 | Gabung banyak pekerjaan menjadi satu SPH | MVP |
| FR-S12 | Finalisasi hanya jika semua validasi lulus (lihat BR-06) | MVP |

## 4. Costing, Pembobotan & Rounding

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-C1 | Dukung Jasa, Material, Total per baris; formula default lihat BR-02 | MVP |
| FR-C2 | Pembobotan: Main Point punya nilai; Sub Point dialokasikan via weight %; total weight wajib 100% | MVP |
| FR-C3 | Validasi weight: 0 ≤ weight ≤ 100; jika belum 100% tampilkan warning + total + selisih, larang finalisasi | MVP |
| FR-C4 | Rounding deterministik di backend; integer Rupiah; floating point bukan sumber kebenaran (lihat BR-04) | MVP |

## 5. Import

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-IE1 | Import Excel: pilih file → baca workbook → pilih sheet → preview → mapping kolom → deteksi hierarki → validasi → preview final → import. **Tidak boleh import tanpa preview** | MVP |
| FR-IE2 | Mapping kolom dapat disesuaikan (URAIAN KEGIATAN→Name, JML→Quantity, SAT→Unit, JASA→Service, MATERIAL→Material) | MVP |
| FR-IE3 | Deteksi hierarki fleksibel: `1./a.`, `I./a.`, `A/1/a/1.` (lihat excel-analysis §5–6); jika parser tidak yakin → minta user klasifikasi, **jangan menebak** | MVP |
| FR-IE4 | Import dengan progress; import dalam 1 transaksi (rollback jika gagal) | MVP |
| FR-IE5 | Export Excel profesional: nama perusahaan, logo, alamat, customer, kapal, nomor SPH, tanggal, project, subject, pekerjaan, sub-pekerjaan, qty, unit, jasa, material, total, subtotal, grand total, notes, tanda tangan; merge, border, alignment, format Rupiah, page setup, print area, repeating header | MVP |
| FR-IE6 | Export PDF: A4 Portrait & Landscape (default landscape sesuai referensi); rapi, tabel tidak terpotong, multi-page, header/footer, page number, total, signature, Rupiah, terbilang, baris Pindahan antar halaman | MVP |
| FR-IE7 | Export DOCX | FUTURE |

## 6. Backup & Restore

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-B1 | Backup database manual (`SPH_Backup_YYYY-MM-DD_HHMMSS.db`) | MVP |
| FR-B2 | Restore: konfirmasi → backup dulu kondisi sekarang → restore → validate → reload | MVP |
| FR-B3 | Auto backup: saat aplikasi ditutup + harian; retention (default simpan 10 backup terakhir) | MVP |

## 7. Settings & Penomoran

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-ST1 | Company settings: Company Name, Address, Phone, Email, Website, NPWP, Logo, Signer Name, Signer Position — tidak di-hard-code | MVP |
| FR-ST2 | Format nomor SPH configurable (prefix, tahun, bulan, sequence, separator), contoh `SPH/GEI/VIII/2026/001`; generator anti-duplikat | MVP |

## 8. Audit, Keamanan Data & Error

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-A1 | Audit log: timestamp, action (CREATE/UPDATE/DELETE/FINALIZE/EXPORT/DUPLICATE/RESTORE), entity, entity_id, description | MVP |
| FR-A2 | Soft delete master data; data yang pernah dipakai SPH tidak boleh hilang | MVP |
| FR-A3 | Error ramah pengguna (bukan error database mentah); detail teknis ke log terstruktur | MVP |
| FR-A4 | Transaction untuk: create SPH, duplicate, revision, finalisasi, import, restore — gagal = rollback | MVP |

## 9. UI/UX

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-U1 | Desain desktop modern: Sidebar, Topbar, Breadcrumb, Table, Card, Modal, Drawer, Tabs, Toast, Confirmation Dialog, Empty State, Loading State, Error State | MVP |
| FR-U2 | **Tampilan simpel, rapi, tertata** — warna dasar **biru flat** + **orange flat** (aksen), komponen seragam | MVP |
| FR-U3 | Bahasa UI: Indonesia, non-teknis (Pekerjaan, Sub-Pekerjaan, Template, Harga, Detail, Dokumen — bukan Entity/Foreign Key/Snapshot ID) | MVP |
| FR-U4 | Halaman: Dashboard; SPH (Semua/Draft/Final/Buat); Pekerjaan (Master/Kategori/Template); Master Data (Customer/Kapal/Material); Import; Backup; Settings | MVP |
| FR-U5 | Dashboard: Total SPH, Draft, Final, Accepted, Nilai SPH bulan ini; quick action (Buat SPH, Tambah Pekerjaan, Tambah Template, Import Excel); Recent SPH | MVP |
| FR-U6 | Search di semua data utama: Work Item, Template, Customer, Vessel, SPH, Material | MVP |
| FR-U7 | Keyboard shortcut: Ctrl+N (SPH baru), Ctrl+S (simpan), Ctrl+F (search), Ctrl+P (preview/print), Esc (tutup modal) | MVP |
| FR-U8 | Performance: startup cepat, search cepat, tabel tidak lag, import ada progress, export tidak membekukan UI | MVP |

## 10. Testing (Wajib)

| ID | Kebutuhan | Prioritas |
|---|---|---|
| FR-T1 | Unit test: pricing (qty × harga), weight (100% / <100% / >100%), rounding (Σ sub = main point), snapshot, duplicate, revision, import (valid/invalid/hierarchy/mapping), backup-restore | MVP |
| FR-T2 | End-to-end scenario lengkap: kategori → pekerjaan → sub → template → customer → kapal → SPH → costing → weight → export → duplicate → ubah master → verifikasi SPH lama tak berubah → backup → restore → verifikasi | MVP |

## 11. Dokumentasi (Wajib)

`README.md`, `docs/requirements.md`, `docs/excel-analysis.md`, `docs/architecture.md`, `docs/database.md`, `docs/business-rules.md`, `docs/sph-flow.md`, `docs/import-excel.md`, `docs/export-document.md`, `docs/backup-restore.md`, `docs/development-plan.md`, `docs/issues.md`.

> `architecture.md`, `database.md`, `sph-flow.md`, `import-excel.md`, `export-document.md`, `backup-restore.md` dibuat pada fase implementasinya masing-masing.
