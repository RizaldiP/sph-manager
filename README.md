# SPH Manager Offline

> Aplikasi desktop Windows untuk manajemen Surat Penawaran Harga (SPH) — **PT. Ganesha Energi Indonesia** (perusahaan repair & maintenance kapal).
>
> **Database pekerjaan perusahaan + template pekerjaan + costing + SPH builder + document generator.** 100% offline.

## Status Proyek

**PHASE 0–3: SELESAI** ✅ (Analisis → Foundation → Database → Master Pekerjaan) · Menunggu instruksi `LANJUT PHASE 4` (Template).

Rencana lengkap: [docs/development-plan.md](docs/development-plan.md).

## Fitur Utama (MVP)

- Database pekerjaan: kategori → pekerjaan → sub-pekerjaan (bisa dipakai ulang)
- Template pekerjaan yang sering digunakan
- Master data: customer, kapal, material
- SPH Builder wizard 8 langkah: gabung banyak pekerjaan → costing → pembobotan → validasi → preview → generate
- Snapshot harga: dokumen lama tidak berubah walau harga master berubah
- Revision (Rev 0/1/2) & duplicate SPH
- Import Excel (preview wajib, mapping fleksibel, deteksi hierarki)
- Export Excel & PDF profesional (A4 landscape, terbilang otomatis, tanda tangan)
- Backup & restore + auto backup + retention
- Audit log & soft delete

## Stack Teknologi

| Lapisan | Teknologi |
|---|---|
| Desktop | Wails v2 |
| Backend | Go 1.26 · SQLite · GORM · migration · structured logging |
| Frontend | Vue 3 (Composition API) · TypeScript · Pinia · Vue Router · Tailwind CSS |
| Desain | Simpel, rapi, tertata — **biru flat** (primer) + **orange flat** (aksen), Bahasa Indonesia |

## Struktur Dokumentasi

| Dokumen | Isi |
|---|---|
| [docs/requirements.md](docs/requirements.md) | Kebutuhan fungsional & non-fungsional (ID + prioritas) |
| [docs/excel-analysis.md](docs/excel-analysis.md) | Analisis menyeluruh file Excel referensi SPH |
| [docs/business-rules.md](docs/business-rules.md) | 16 aturan bisnis mengikat (snapshot, pembobotan, rounding, dll.) |
| [docs/development-plan.md](docs/development-plan.md) | 12 fase pengembangan + acceptance criteria + status |
| [docs/issues.md](docs/issues.md) | Ambiguity, asumsi & dampaknya |

> Dokumen tambahan (`architecture.md`, `database.md`, `sph-flow.md`, `import-excel.md`, `export-document.md`, `backup-restore.md`) dibuat pada fase implementasinya masing-masing.

## Menjalankan (setelah Phase 1)

```powershell
wails dev      # mode pengembangan
wails build    # build produksi Windows
```

> Catatan: file referensi `SPH_KRI_OWA.xls` sengaja **tidak** di-commit (memuat harga asli perusahaan — lihat `.gitignore`).

## Alur Bisnis

```text
DATABASE PEKERJAAN → MASTER PEKERJAAN → SUB-PEKERJAAN → TEMPLATE
        ↓
PILIH BEBERAPA PEKERJAAN → SPH BUILDER → COSTING → PEMBOBOTAN
        ↓
VALIDASI → SNAPSHOT → PREVIEW → PDF / EXCEL
```

## Aturan Kerja Pengembangan

Setiap sesi: baca dokumentasi → tentukan fase aktif → kerjakan **hanya** fase aktif → test ✓ formatter ✓ build ✓ dokumentasi ✓ → **berhenti** menunggu instruksi `LANJUT PHASE X`.
