# MASTER PROMPT — APLIKASI SPH MANAGER OFFLINE
## PT. Ganesha Energi Indonesia

> Dokumen ini adalah instruksi utama untuk OpenCode AI. Gunakan dokumen ini sebagai source of truth selama pengembangan aplikasi.
>
> **PENTING:** Jangan langsung membuat seluruh aplikasi. Kerjakan berdasarkan fase, lakukan pemeriksaan, testing, build, dan dokumentasi pada setiap fase. Setelah satu fase selesai, berhenti dan tunggu instruksi `LANJUT PHASE X`.

---

# 1. PERAN

Bertindak sebagai:

- Senior Software Architect
- Senior Golang Developer
- Senior Vue 3 Developer
- Desktop Application Developer
- Database Designer
- UI/UX Designer
- QA Engineer
- Business Analyst untuk perusahaan repair & maintenance kapal

Jangan bertindak hanya sebagai code generator.

Sebelum membuat kode, pahami:

1. kebutuhan bisnis,
2. struktur data,
3. hubungan antar-data,
4. alur pengguna,
5. aturan perhitungan,
6. kebutuhan dokumen SPH,
7. kebutuhan offline,
8. kebutuhan pengembangan masa depan.

Jika menemukan requirement ambigu, jangan diam-diam membuat asumsi yang berisiko. Dokumentasikan asumsi tersebut.

---

# 2. TUJUAN APLIKASI

Buat aplikasi desktop Windows **SPH Manager Offline** untuk membantu pengguna:

- menyimpan database pekerjaan,
- menyimpan sub-pekerjaan,
- membuat template pekerjaan,
- menyimpan customer,
- menyimpan data kapal,
- menyusun costing,
- membuat SPH,
- menggabungkan banyak pekerjaan menjadi satu SPH,
- menggunakan kembali pekerjaan lama,
- melakukan pembobotan pekerjaan,
- melakukan import data Excel,
- menghasilkan PDF/Excel,
- melakukan backup dan restore.

Aplikasi bukan sekadar “generator SPH”.

Konsep utama aplikasi:

> **Database pekerjaan perusahaan + template pekerjaan + costing + SPH builder + document generator.**

---

# 3. MASALAH BISNIS YANG HARUS DISELESAIKAN

Pengguna sering mengalami:

1. lupa detail/sub-list pekerjaan;
2. menulis pekerjaan yang sama berulang kali;
3. memiliki banyak pekerjaan yang harus digabung menjadi satu SPH;
4. kesulitan membagi harga pekerjaan ke sub-pekerjaan;
5. memiliki banyak file Excel dengan format pekerjaan yang berbeda;
6. kesulitan mencari pekerjaan lama;
7. ingin membuat SPH secara cepat namun tetap mengikuti format perusahaan.

Aplikasi harus membuat proses:

```text
Cari pekerjaan
    ↓
pilih pekerjaan
    ↓
sub-pekerjaan muncul
    ↓
aturl pekerjaan
    ↓
masukkan costing
    ↓
gabungkan beberapa pekerjaan
    ↓
validasi
    ↓
preview
    ↓
generate SPH
```

menjadi sederhana.

---

# 4. PLATFORM

Target utama:

- Windows Desktop
- 100% offline
- tidak membutuhkan internet saat digunakan
- database lokal
- dokumen lokal
- backup lokal

Aplikasi harus tetap berjalan ketika:

- Wi-Fi dimatikan;
- internet tidak tersedia;
- komputer tidak terhubung ke server.

---

# 5. STACK TEKNOLOGI

Gunakan:

## Backend

- Golang
- SQLite
- GORM atau data access layer yang rapi
- Migration
- Transaction
- Validation
- Structured logging

## Frontend

- Vue 3
- Composition API
- TypeScript
- Pinia
- Vue Router
- Tailwind CSS

## Desktop

Prioritaskan:

- Wails

Karena kombinasi:

```text
Go + Vue + Desktop
```

cocok untuk aplikasi offline.

Jika project yang tersedia sudah menggunakan framework lain, jangan menggantinya tanpa analisis terlebih dahulu.

---

# 6. ARSITEKTUR

Gunakan pemisahan yang jelas:

```text
Desktop
│
├── Frontend
│   ├── Pages
│   ├── Components
│   ├── Layouts
│   ├── Stores
│   ├── Services
│   ├── Types
│   └── Utils
│
├── Desktop Bridge
│
└── Backend
    ├── Models
    ├── Repositories
    ├── Services
    ├── Validators
    ├── Documents
    ├── Importers
    ├── Backup
    ├── Settings
    └── Migrations
             │
             └── SQLite
```

Business logic tidak boleh ditaruh sembarangan di Vue component.

Frontend hanya bertanggung jawab pada:

- UI;
- input;
- state;
- presentation;
- komunikasi dengan backend.

Backend bertanggung jawab pada:

- business rule;
- perhitungan;
- validasi;
- database;
- import;
- export;
- backup;
- snapshot;
- revision.

---

# 7. WAJIB ANALISIS FILE EXCEL REFERENSI

Jika terdapat file:

```text
SPH_KRI_OWA.xls
```

atau file SPH/Excel lain di workspace, **WAJIB dianalisis terlebih dahulu sebelum membuat database dan generator dokumen.**

Jangan hanya membaca beberapa cell.

Analisis:

- semua sheet;
- header;
- sub-header;
- merge cell;
- hierarchy;
- nomor pekerjaan;
- Main Point;
- Sub Point;
- quantity;
- unit;
- harga jasa;
- harga material;
- subtotal;
- grand total;
- formula;
- format Rupiah;
- catatan;
- tanda tangan;
- footer;
- page layout;
- print area;
- ukuran halaman;
- orientasi halaman.

Buat file:

```text
docs/excel-analysis.md
```

Isi:

```text
1. File yang dianalisis
2. Sheet
3. Struktur setiap sheet
4. Struktur header
5. Struktur Main Point
6. Struktur Sub Point
7. Struktur quantity
8. Struktur unit
9. Struktur harga
10. Formula
11. Format angka
12. Format dokumen
13. Data yang bisa dijadikan master
14. Data yang harus menjadi snapshot
15. Perbedaan format
16. Masalah/ketidakkonsistenan
17. Rekomendasi struktur database
18. Rekomendasi generator
```

Jika ada perbedaan format antar-file, jangan menghapus informasi. Buat sistem yang cukup fleksibel untuk menanganinya.

---

# 8. KONSEP DATA

Konsep dasar:

```text
CATEGORY
    ↓
WORK ITEM
    ↓
SUB WORK ITEM
    ↓
TEMPLATE
    ↓
SPH
    ↓
SPH ITEM
    ↓
SPH SUB ITEM
```

Contoh:

```text
Electrical
    ↓
Repair Control Panel
    ↓
    ├── Inspection
    ├── Troubleshooting
    ├── Wiring Check
    ├── Component Replacement
    ├── Testing
    └── Commissioning
```

---

# 9. DATABASE

Gunakan SQLite.

Database minimal harus memiliki:

```text
categories
work_items
work_sub_items
templates
template_items
customers
vessels
materials
sph_documents
sph_items
sph_sub_items
sph_revisions
audit_logs
settings
```

---

# 10. CATEGORY

Tabel:

```text
categories
```

Field:

```text
id
code
name
description
is_active
created_at
updated_at
deleted_at
```

Contoh:

```text
Electrical
Automation
Instrumentation
Mechanical
HVAC
Navigation
Communication
PLC
Control System
Testing & Commissioning
Other
```

---

# 11. MASTER PEKERJAAN

Tabel:

```text
work_items
```

Field:

```text
id
category_id
code
name
description
default_unit
default_quantity
default_service_price
default_material_price
notes
is_active
created_at
updated_at
deleted_at
```

Contoh:

```text
EL-001
Repair Control Panel
```

---

# 12. SUB-PEKERJAAN

Tabel:

```text
work_sub_items
```

Field:

```text
id
work_item_id
code
sequence
name
description
difficulty_weight
default_unit
default_quantity
default_service_price
default_material_price
notes
is_active
created_at
updated_at
deleted_at
```

Contoh:

```text
Repair PLC

01 Inspection
02 Backup Program
03 Mapping I/O
04 Troubleshooting
05 Repair
06 Reprogramming
07 Testing
08 Commissioning
```

Urutan harus bisa diubah.

---

# 13. TEMPLATE

Template adalah kumpulan pekerjaan yang sering digunakan kembali.

Contoh:

```text
Template: Repair PLC
```

Isi:

```text
Inspection
Backup Program
Mapping I/O
Troubleshooting
Repair
Reprogramming
Testing
Commissioning
```

Template dapat:

- dibuat;
- diedit;
- diduplikasi;
- diaktifkan;
- dinonaktifkan;
- diurutkan.

---

# 14. SNAPSHOT

Ini adalah business rule penting.

Ketika master digunakan untuk membuat SPH:

```text
MASTER
    ↓
COPY/SNAPSHOT
    ↓
SPH
```

Jangan membuat SPH lama bergantung pada harga master saat ini.

Contoh:

Master:

```text
Repair PLC = Rp10.000.000
```

SPH dibuat:

```text
Rp10.000.000
```

Kemudian master berubah:

```text
Repair PLC = Rp15.000.000
```

SPH lama **tetap Rp10.000.000**.

SPH baru:

```text
Rp15.000.000
```

Wajib dibuat automated test untuk ini.

---

# 15. CUSTOMER

Tabel:

```text
customers
```

Field:

```text
id
code
name
address
phone
email
pic_name
pic_position
notes
is_active
created_at
updated_at
deleted_at
```

---

# 16. KAPAL

Tabel:

```text
vessels
```

Field:

```text
id
customer_id
code
name
vessel_number
vessel_type
notes
is_active
created_at
updated_at
deleted_at
```

Relasi:

```text
Customer
    ↓
Vessel
```

---

# 17. MATERIAL

Tabel:

```text
materials
```

Field:

```text
id
code
name
description
unit
default_price
supplier
notes
is_active
created_at
updated_at
deleted_at
```

---

# 18. SPH

Tabel:

```text
sph_documents
```

Field minimal:

```text
id
document_number
revision
date
customer_id
vessel_id
project_name
subject
reference
location
valid_until
status
subtotal_service
subtotal_material
grand_total
notes
created_at
updated_at
finalized_at
deleted_at
```

Status:

```text
DRAFT
REVIEW
FINAL
SENT
ACCEPTED
REJECTED
CANCELLED
```

---

# 19. SPH ITEM

Tabel:

```text
sph_items
```

Field:

```text
id
sph_id
sequence
work_item_id
name_snapshot
description_snapshot
quantity
unit
service_unit_price
material_unit_price
service_total
material_total
total
notes
created_at
updated_at
```

---

# 20. SPH SUB ITEM

Tabel:

```text
sph_sub_items
```

Field:

```text
id
sph_item_id
sequence
name_snapshot
description_snapshot
quantity
unit
weight
service_unit_price
material_unit_price
service_total
material_total
total
notes
created_at
updated_at
```

---

# 21. MAIN POINT DAN SUB POINT

Aplikasi harus mendukung:

```text
MAIN POINT
    ↓
SUB POINT
```

Contoh:

```text
1. Service dan Perbaikan Transmitter Sensor MPK I dan MPK II

    1.1 Service kabel, konektor dan housing sensor
    1.2 Identifikasi transmitter
    1.3 Service line supply
    1.4 Pembersihan housing transmitter
    1.5 Kalibrasi sensor
    1.6 Service Engine RPM Tacho Generator
    1.7 Service Turbo Charge Tacho
    1.8 Service Fuel Index Sensor
```

Struktur harus dapat digunakan untuk dokumen SPH dan costing.

---

# 22. PEMBOBOTAN

Jika Main Point memiliki nilai:

```text
Rp30.625.000
```

dan Sub Point:

```text
10%
15%
20%
25%
30%
```

maka:

```text
Sub Point Value =
Main Point Value × Weight / 100
```

Total weight wajib:

```text
100%
```

Validasi:

```text
0 <= weight <= 100
```

Jika belum 100%:

- tampilkan warning;
- tampilkan total;
- tampilkan selisih;
- jangan izinkan finalisasi.

---

# 23. ROUNDING

Nilai uang harus menggunakan integer Rupiah atau decimal yang aman.

Jangan menjadikan floating point sebagai sumber kebenaran.

Jika terjadi:

```text
Rp33.333
Rp33.333
Rp33.334
```

total harus tepat.

Selisih pembulatan harus ditangani secara deterministik.

Business rule rounding harus berada di backend.

---

# 24. JASA DAN MATERIAL

Dukung minimal:

```text
Jasa
Material
Total
```

Formula default:

```text
service_total =
quantity × service_unit_price

material_total =
quantity × material_unit_price

total =
service_total + material_total
```

Jika Excel referensi menggunakan formula berbeda, ikuti format bisnis yang ditemukan dari Excel.

---

# 25. SPH BUILDER

Buat wizard:

```text
STEP 1
Informasi SPH

STEP 2
Pilih Pekerjaan

STEP 3
Susun Main Point

STEP 4
Susun Sub Point

STEP 5
Costing

STEP 6
Validasi

STEP 7
Preview

STEP 8
Simpan / Generate
```

---

# 26. INFORMASI SPH

Input:

- Nomor SPH
- Tanggal
- Customer
- Kapal
- Project
- Subject/Perihal
- Reference
- Lokasi
- Masa berlaku
- PIC
- Catatan

Field wajib harus jelas.

---

# 27. MEMILIH PEKERJAAN

User harus bisa:

- search;
- filter kategori;
- memilih master;
- memilih template;
- mengambil dari SPH lama;
- menambah manual.

Contoh UI:

```text
[ Cari pekerjaan... ]

Kategori:
[ Semua ▼ ]

--------------------------------
☐ Repair AMS
☐ Repair PLC
☐ Repair Sensor
☐ Testing Control Panel
--------------------------------
```

---

# 28. MULTI SELECTION

User dapat memilih banyak pekerjaan:

```text
☑ Repair AMS
☑ Repair PLC
☑ Repair Sensor
☑ Testing Control Panel
```

Kemudian:

```text
[ Tambahkan ke SPH ]
```

Semua menjadi Main Point.

---

# 29. DRAG AND DROP

Main Point dapat diurutkan.

Contoh:

```text
1 Repair AMS
2 Repair PLC
3 Repair Sensor
4 Testing
```

menjadi:

```text
1 Repair PLC
2 Repair AMS
3 Testing
4 Repair Sensor
```

Urutan harus tersimpan.

---

# 30. EDIT DI DALAM SPH

User dapat mengubah:

- nama;
- deskripsi;
- quantity;
- unit;
- jasa;
- material;
- notes;
- sub-item;
- weight.

Perubahan hanya untuk SPH tersebut.

Master tidak berubah.

---

# 31. DUPLICATE SPH

Fitur:

```text
Duplicate SPH
```

Contoh:

```text
SPH-001 KRI A
```

menjadi:

```text
SPH-002 KRI B
```

Semua data menjadi snapshot baru.

Source tidak boleh berubah.

---

# 32. REVISION

Dukung:

```text
SPH-001 Rev 0
SPH-001 Rev 1
SPH-001 Rev 2
```

Revision lama tetap dapat dilihat.

Jangan overwrite histori.

---

# 33. DAFTAR SPH

Halaman:

```text
Semua SPH
```

Kolom:

```text
Nomor
Revision
Tanggal
Customer
Kapal
Project
Total
Status
```

Fitur:

- search;
- filter;
- sort;
- preview;
- edit draft;
- duplicate;
- export;
- archive.

---

# 34. DASHBOARD

Tampilkan:

```text
Total SPH
SPH Draft
SPH Final
SPH Accepted
Nilai SPH Bulan Ini
```

Quick action:

```text
+ Buat SPH
+ Tambah Pekerjaan
+ Tambah Template
+ Import Excel
```

Recent SPH:

```text
Nomor
Customer
Kapal
Total
Status
```

---

# 35. IMPORT EXCEL

Flow:

```text
Pilih File
    ↓
Baca Workbook
    ↓
Pilih Sheet
    ↓
Preview
    ↓
Mapping Kolom
    ↓
Deteksi Hierarki
    ↓
Validasi
    ↓
Preview Final
    ↓
Import
```

Tidak boleh import langsung tanpa preview.

---

# 36. EXCEL MAPPING

Contoh:

```text
URAIAN KEGIATAN → Name
JML             → Quantity
SAT             → Unit
JASA            → Service Price
MATERIAL        → Material Price
```

Mapping harus dapat disesuaikan.

---

# 37. HIERARCHY IMPORT

Importer harus mampu mengenali struktur seperti:

```text
1. Main Point
    1.1 Sub Point
    1.2 Sub Point
```

atau:

```text
I. Main Point
    a. Sub Point
    b. Sub Point
```

Jika parser tidak yakin:

**jangan menebak.**

Tampilkan preview dan biarkan user melakukan klasifikasi/mapping.

---

# 38. EXPORT EXCEL

Hasil harus profesional.

Minimal:

- nama perusahaan;
- logo;
- alamat;
- customer;
- kapal;
- nomor SPH;
- tanggal;
- project;
- subject;
- pekerjaan;
- sub-pekerjaan;
- quantity;
- unit;
- jasa;
- material;
- total;
- subtotal;
- grand total;
- notes;
- signature.

Perhatikan:

- merge;
- border;
- alignment;
- format Rupiah;
- page setup;
- print area;
- repeating header.

---

# 39. EXPORT PDF

Gunakan ukuran sesuai template hasil analisis Excel.

Minimal dukung:

```text
A4 Portrait
A4 Landscape
```

Harus:

- rapi;
- tabel tidak terpotong;
- multi-page;
- header;
- footer;
- page number;
- total;
- signature;
- Rupiah.

---

# 40. COMPANY SETTINGS

Settings:

```text
Company Name
Address
Phone
Email
Website
NPWP
Logo
Signer Name
Signer Position
```

Jangan hard-code.

---

# 41. NOMOR SPH

Format configurable.

Contoh:

```text
SPH/GEI/VIII/2026/001
```

Settings:

```text
Prefix
Year
Month
Sequence
Separator
```

Generator harus mencegah nomor duplikat.

---

# 42. BACKUP

WAJIB.

Fitur:

```text
Backup Database
Restore Database
Backup Settings
```

Contoh:

```text
SPH_Backup_2026-08-23_231500.db
```

Restore:

```text
1. Konfirmasi
2. Backup database sekarang
3. Restore
4. Validate
5. Reload aplikasi
```

---

# 43. AUTO BACKUP

Dukung:

- backup saat aplikasi ditutup;
- backup harian;
- retention.

Contoh:

```text
Simpan 10 backup terakhir.
```

---

# 44. AUDIT LOG

Catat:

```text
CREATE
UPDATE
DELETE
FINALIZE
EXPORT
DUPLICATE
RESTORE
```

Field:

```text
timestamp
action
entity
entity_id
description
```

---

# 45. SOFT DELETE

Untuk master data gunakan:

```text
is_active
deleted_at
```

Data yang pernah digunakan oleh SPH tidak boleh hilang sehingga merusak histori.

---

# 46. ERROR HANDLING

Jangan tampilkan error database mentah.

Buruk:

```text
SQLITE_CONSTRAINT_FOREIGNKEY
```

Bagus:

```text
Data tidak dapat dihapus karena masih digunakan oleh dokumen SPH.
```

Detail teknis masuk log.

---

# 47. UI/UX

Gunakan desain desktop modern:

- Sidebar
- Topbar
- Breadcrumb
- Table
- Card
- Modal
- Drawer
- Tabs
- Toast
- Confirmation Dialog
- Empty State
- Loading State
- Error State

Bahasa UI:

**Bahasa Indonesia.**

---

# 48. HALAMAN

Minimal:

```text
Dashboard

SPH
├── Semua SPH
├── Draft
├── Final
└── Buat SPH

Pekerjaan
├── Master Pekerjaan
├── Kategori
└── Template

Master Data
├── Customer
├── Kapal
└── Material

Import / Export

Backup

Settings
```

---

# 49. UX UNTUK PENGGUNA NON-TEKNIS

Jangan menggunakan istilah teknis database.

Gunakan:

```text
Pekerjaan
Sub-Pekerjaan
Template
Harga
Detail
Dokumen
```

Bukan:

```text
Entity
Foreign Key
Snapshot ID
Relation
```

---

# 50. SEARCH

Semua data utama harus dapat dicari:

```text
Work Item
Template
Customer
Vessel
SPH
Material
```

---

# 51. KEYBOARD SHORTCUT

Minimal:

```text
Ctrl + N = SPH Baru
Ctrl + S = Simpan
Ctrl + F = Search
Ctrl + P = Preview / Print
Esc = Tutup modal
```

---

# 52. VALIDASI

SPH:

- nomor wajib;
- tanggal wajib;
- customer wajib;
- minimal satu pekerjaan;
- quantity valid;
- harga valid.

Weight:

```text
0 <= weight <= 100
TOTAL = 100%
```

Finalisasi hanya boleh dilakukan jika seluruh validasi berhasil.

---

# 53. TRANSACTION

Gunakan database transaction untuk:

- create SPH;
- duplicate SPH;
- revision;
- finalisasi;
- import;
- restore.

Jika gagal:

```text
ROLLBACK
```

---

# 54. OFFLINE

Runtime tidak boleh bergantung pada:

- CDN;
- API eksternal;
- Google Fonts online;
- cloud storage;
- online authentication.

Semua asset dibundle.

---

# 55. LOKASI DATA WINDOWS

Jangan menyimpan database di folder executable jika tidak diperlukan.

Gunakan folder aplikasi data:

```text
AppData/
    database/
    backups/
    exports/
    logs/
    templates/
```

Gunakan path OS-aware.

---

# 56. PERFORMANCE

Target:

- startup cepat;
- search cepat;
- table tidak lag;
- import memiliki progress;
- export tidak membekukan UI;
- database transaction aman.

---

# 57. DOKUMENTASI

Wajib membuat:

```text
README.md
docs/requirements.md
docs/excel-analysis.md
docs/architecture.md
docs/database.md
docs/business-rules.md
docs/sph-flow.md
docs/import-excel.md
docs/export-document.md
docs/backup-restore.md
docs/development-plan.md
docs/issues.md
```

---

# 58. PHASE DEVELOPMENT

## PHASE 0 — ANALISIS

Kerjakan:

1. Inspect repository.
2. Cari file Excel.
3. Analisis file SPH.
4. Analisis struktur pekerjaan.
5. Analisis hierarchy.
6. Analisis formula.
7. Analisis format dokumen.
8. Buat requirements.
9. Buat business rules.
10. Buat database design.
11. Buat development plan.
12. Dokumentasikan ambiguity.

Output:

```text
docs/excel-analysis.md
docs/requirements.md
docs/business-rules.md
docs/development-plan.md
```

**JANGAN CODING FITUR UTAMA PADA PHASE 0.**

Setelah selesai:

```text
PHASE 0 COMPLETE

Files Analyzed:
...

Excel Structure:
...

Business Rules:
...

Recommended Database:
...

Recommended Architecture:
...

Important Assumptions:
...

Potential Issues:
...

Documentation Created:
...

Next:
LANJUT PHASE 1
```

---

# 59. PHASE 1 — FOUNDATION

Implement:

- Wails;
- Go;
- Vue;
- TypeScript;
- Tailwind;
- SQLite;
- migration;
- config;
- logging.

Test:

```text
wails dev
```

Aplikasi harus bisa terbuka.

Setelah selesai:

```text
PHASE 1 COMPLETE
```

Berhenti.

---

# 60. PHASE 2 — DATABASE

Implement seluruh migration dan model.

Test:

```text
Database kosong
→ migration
→ semua table terbentuk
→ foreign key valid
```

Berhenti setelah selesai.

---

# 61. PHASE 3 — MASTER PEKERJAAN

Implement:

- category;
- work item;
- sub work item;
- search;
- CRUD;
- reorder;
- active/inactive.

Test:

```text
Repair PLC
    ├── Inspection
    ├── Backup
    ├── Mapping
    ├── Troubleshooting
    ├── Repair
    └── Testing
```

Berhenti.

---

# 62. PHASE 4 — TEMPLATE

Implement:

- create;
- edit;
- duplicate;
- reorder;
- activate/deactivate.

Test reuse.

Berhenti.

---

# 63. PHASE 5 — SPH BUILDER

Implement:

```text
Info
→ Select Work
→ Organize
→ Sub Work
→ Costing
→ Validation
→ Preview
→ Save
```

Test satu SPH lengkap.

Berhenti.

---

# 64. PHASE 6 — PEMBOBOTAN

Implement:

- weight;
- allocation;
- rounding;
- validation.

Test:

```text
10 + 15 + 20 + 25 + 30 = 100
```

Berhenti.

---

# 65. PHASE 7 — MULTI WORK COMBINATION

Test:

```text
Repair AMS
Repair PLC
Repair Sensor
Testing
Calibration
```

menjadi:

```text
1 Repair AMS
2 Repair PLC
3 Repair Sensor
4 Testing
5 Calibration
```

dalam satu SPH.

Berhenti.

---

# 66. PHASE 8 — IMPORT EXCEL

Implement:

- XLS/XLSX reader;
- sheet selection;
- preview;
- column mapping;
- hierarchy detection;
- validation;
- import transaction.

Berhenti.

---

# 67. PHASE 9 — EXPORT

Implement:

- Excel;
- PDF;
- optional DOCX.

Hasil harus mengikuti template referensi.

Berhenti.

---

# 68. PHASE 10 — BACKUP

Implement:

- manual backup;
- restore;
- auto backup;
- retention;
- validation.

Berhenti.

---

# 69. PHASE 11 — POLISH

Perbaiki:

- UI;
- UX;
- loading;
- empty states;
- error states;
- shortcut;
- performance;
- accessibility.

Berhenti.

---

# 70. PHASE 12 — RELEASE

Buat:

- Windows executable;
- installer;
- portable build jika memungkinkan;
- versioning;
- installation guide.

Test pada environment bersih.

---

# 71. TESTING WAJIB

Buat unit test untuk:

## Pricing

```text
quantity × unit price
```

## Weight

```text
100%
<100%
>100%
```

## Rounding

Total sub-item harus sama dengan Main Point.

## Snapshot

Master berubah tidak boleh mengubah SPH lama.

## Duplicate

Duplicate tidak mengubah source.

## Revision

Revision lama tetap tersedia.

## Import

- valid;
- invalid;
- hierarchy;
- mapping.

## Backup

Backup → modify → restore → verify.

---

# 72. END-TO-END TEST

Harus ada skenario:

```text
Create Category
↓
Create Work Item
↓
Create 5 Sub Items
↓
Create Template
↓
Create Customer
↓
Create Vessel
↓
Create SPH
↓
Select Template
↓
Add Work Item lain
↓
Reorder Main Point
↓
Edit Sub Item
↓
Set Price
↓
Set Weight
↓
Validate
↓
Save
↓
Preview
↓
Export PDF
↓
Export Excel
↓
Duplicate SPH
↓
Change Master Price
↓
Verify Old SPH Unchanged
↓
Backup
↓
Restore
↓
Verify Data
```

---

# 73. FUTURE READY

Arsitektur harus memungkinkan:

- multi-user;
- cloud sync;
- approval;
- quotation tracking;
- purchase order;
- invoice;
- work order;
- project management.

Namun jangan implementasikan pada MVP.

---

# 74. ATURAN KERJA OPENCODE

Setiap sesi:

```text
1. Baca README.md
2. Baca docs/development-plan.md
3. Baca docs/requirements.md
4. Baca docs/architecture.md
5. Baca docs/business-rules.md
6. Tentukan phase aktif
7. Inspect implementasi existing
8. Kerjakan hanya phase aktif
9. Run tests
10. Run formatter
11. Run build
12. Update documentation
```

Jangan mengerjakan phase berikutnya tanpa instruksi.

---

# 75. JIKA TERJADI ERROR

Jangan menggunakan solusi seperti:

```text
hapus database
hapus migration
hapus fitur
disable validation
comment code
```

hanya untuk membuat aplikasi berhasil build.

Cari root cause.

Urutan:

```text
Error
↓
Reproduce
↓
Root Cause
↓
Fix
↓
Test
↓
Build
↓
Document
```

---

# 76. JIKA REQUIREMENT AMBIGU

Prioritas:

```text
1. Excel referensi
2. Dokumen Master Specification
3. Existing project
4. Business best practice
5. Assumption
```

Jika harus membuat asumsi penting, tulis ke:

```text
docs/issues.md
```

Format:

```text
Question:
Assumption:
Reason:
Impact:
```

---

# 77. DEFINITION OF DONE

Fitur hanya dianggap selesai jika:

- UI selesai;
- backend selesai;
- database selesai;
- validation selesai;
- error handling selesai;
- test selesai;
- build berhasil;
- documentation selesai.

---

# 78. MVP ACCEPTANCE CRITERIA

Pengguna harus dapat:

1. membuka aplikasi tanpa internet;
2. membuat kategori;
3. membuat pekerjaan;
4. membuat sub-pekerjaan;
5. membuat template;
6. membuat customer;
7. membuat kapal;
8. membuat SPH;
9. memilih banyak pekerjaan;
10. menggabungkannya menjadi satu SPH;
11. mengubah urutan;
12. mengedit sub-pekerjaan;
13. memasukkan jasa/material;
14. melakukan pembobotan;
15. melihat total;
16. menyimpan snapshot;
17. duplicate SPH;
18. membuat revision;
19. melihat histori;
20. preview;
21. export PDF;
22. export Excel;
23. backup;
24. restore.

---

# 79. VISI AKHIR

Aplikasi harus terasa seperti:

> “Saya tidak perlu mengingat semua detail pekerjaan yang pernah saya kerjakan. Saya cukup mencari pekerjaan yang saya butuhkan, memilihnya, menggabungkan beberapa pekerjaan, mengatur harga, dan SPH langsung jadi.”

Alur final:

```text
DATABASE PEKERJAAN
        ↓
MASTER PEKERJAAN
        ↓
SUB-PEKERJAAN
        ↓
TEMPLATE
        ↓
PILIH BEBERAPA PEKERJAAN
        ↓
SPH BUILDER
        ↓
COSTING
        ↓
PEMBOBOTAN
        ↓
VALIDASI
        ↓
SNAPSHOT
        ↓
PREVIEW
        ↓
PDF / EXCEL
```

Prioritas:

1. Akurasi data.
2. Akurasi perhitungan.
3. Kemudahan penggunaan.
4. Reusable work library.
5. Keamanan data offline.
6. Kesamaan hasil dengan format SPH perusahaan.
7. Maintainability.
8. Kemudahan pengembangan masa depan.

---

# 80. INSTRUKSI PERTAMA UNTUK OPENCODE

Mulai sekarang.

Jangan langsung coding.

Kerjakan **PHASE 0 — ANALISIS**.

Langkah pertama:

```text
1. Inspect seluruh repository.
2. Cari semua file:
   - .xls
   - .xlsx
   - .csv
   - .pdf
   - .docx
   - .md
3. Cari file yang berkaitan dengan SPH.
4. Analisis file SPH secara menyeluruh.
5. Identifikasi pola Main Point dan Sub Point.
6. Identifikasi costing.
7. Identifikasi formula.
8. Identifikasi struktur dokumen.
9. Buat database recommendation.
10. Buat business rules.
11. Buat development plan.
```

Jangan mengarang struktur SPH jika belum ditemukan pada file referensi.

Jika file referensi tersedia, jadikan file tersebut sebagai sumber utama untuk menentukan format output.

Setelah Phase 0 selesai:

**BERHENTI.**

Tampilkan:

```text
PHASE 0 COMPLETE

Files Analyzed:
...

Excel Structure:
...

Business Rules:
...

Recommended Database:
...

Recommended Architecture:
...

Important Assumptions:
...

Potential Issues:
...

Documentation Created:
...

Next:
LANJUT PHASE 1
```

Jangan lanjut otomatis ke Phase 1.

# END OF MASTER PROMPT
