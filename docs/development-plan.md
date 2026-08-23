# Development Plan — SPH Manager Offline

> Rencana kerja per fase sesuai Master Specification. Setiap fase ditutup dengan: test ✓ formatter ✓ build ✓ dokumentasi ✓ lalu **BERHENTI** menunggu instruksi `LANJUT PHASE X`.

## Status Ringkas

| Fase | Nama | Status |
|---|---|---|
| 0 | Analisis & Dokumentasi | ✅ Selesai |
| 1 | Foundation (Wails + Go + Vue + SQLite) | ⬜ Belum |
| 2 | Database (migration + model) | ⬜ Belum |
| 3 | Master Pekerjaan | ⬜ Belum |
| 4 | Template | ⬜ Belum |
| 5 | SPH Builder | ⬜ Belum |
| 6 | Pembobotan & Rounding | ⬜ Belum |
| 7 | Kombinasi Multi Pekerjaan | ⬜ Belum |
| 8 | Import Excel | ⬜ Belum |
| 9 | Export (Excel + PDF) | ⬜ Belum |
| 10 | Backup & Restore | ⬜ Belum |
| 11 | Polish UI/UX | ⬜ Belum |
| 12 | Release (build + push GitHub) | ⬜ Belum |

---

## PHASE 0 — Analisis ✅

**Scope:** inspect repository, analisis file Excel referensi menyeluruh, requirements, business rules, database design (rekomendasi), development plan, dokumentasi ambiguity.

**Output:** `docs/excel-analysis.md`, `docs/requirements.md`, `docs/business-rules.md`, `docs/development-plan.md`, `docs/issues.md`, `README.md`, `.gitignore`.

**Acceptance:** semua dokumen terbuat dan konsisten; file referensi di-exclude dari git.

## PHASE 1 — Foundation

**Scope:** install Wails CLI; scaffold proyek Wails v2; Go backend + SQLite (GORM) + migration framework + config + structured logging; Vue 3 + TypeScript + Pinia + Vue Router + Tailwind CSS (design system dasar: biru flat + orange flat); layout Sidebar+Topbar; halaman Dashboard stub.

**Test:** `wails dev` — aplikasi terbuka, navigasi antar halaman stub berfungsi.

**Acceptance:** aplikasi jalan offline; lint/format bersih; build dev sukses.

## PHASE 2 — Database

**Scope:** migration seluruh tabel: `categories`, `work_items`, `work_sub_items`, `templates`, `template_items`, `customers`, `vessels`, `materials`, `sph_documents`, `sph_items`, `sph_sub_items`, `sph_revisions`, `audit_logs`, `settings` + index + FK + soft delete.

**Test:** database kosong → migration → semua tabel terbentuk → FK valid (cek pragma foreign_key_list).

**Acceptance:** migration idempotent, test integrasi DB hijau. Dokumentasi `docs/database.md`.

## PHASE 3 — Master Pekerjaan

**Scope:** CRUD kategori, work item, sub work item; search; reorder (drag-and-drop); aktif/nonaktif; validasi; error handling ramah; halaman UI sesuai FR-U4.

**Test:** buat `Electrical → Repair Control Panel → Inspection, Troubleshooting, Wiring Check, Component Replacement, Testing, Commissioning`; ubah urutan; nonaktifkan.

**Acceptance:** CRUD + reorder tersimpan; test unit + integrasi hijau.

## PHASE 4 — Template

**Scope:** CRUD template + template_items; duplicate; reorder; aktif/nonaktif; pakai template saat membuat SPH.

**Test:** buat template "Repair PLC" berisi 8 sub langkah; duplikat; nonaktifkan.

**Acceptance:** reuse template berfungsi; test hijau.

## PHASE 5 — SPH Builder

**Scope:** wizard 8 langkah (Info → Pilih Pkerjaan → Susun Main Point → Sub Point → Costing → Validasi → Preview → Simpan/Generate); snapshot penuh saat simpan; edit dalam SPH; duplicate; revision; daftar SPH; dashboard.

**Test:** satu SPH lengkap end-to-end di UI.

**Acceptance:** SPH tersimpan sebagai snapshot; validasi & transaksi sesuai BR-06/BR-16. Dokumentasi `docs/sph-flow.md`.

## PHASE 6 — Pembobotan & Rounding

**Scope:** mode PEMBOBOTAN per main point; input weight; alokasi nilai sub point; rounding largest remainder (BR-04); warning selisih; larang finalisasi jika ≠ 100%.

**Test:** `10 + 15 + 20 + 25 + 30 = 100`; kasus <100 dan >100; kasus pembulatan (mis. 3 × 33,33%).

**Acceptance:** Σ sub = main point tepat; unit test hijau.

## PHASE 7 — Kombinasi Multi Pekerjaan

**Scope:** pilih banyak pekerjaan → satu SPH; drag-and-drop urutan final; penggabungan dari master + template + SPH lama.

**Test:** Repair AMS + Repair PLC + Repair Sensor + Testing + Calibration → 1 SPH berurut 1–5.

**Acceptance:** urutan tersimpan; snapshot benar.

## PHASE 8 — Import Excel

**Scope:** reader XLS/XLSX; pilih sheet; preview wajib; mapping kolom adjustable; deteksi hierarki fleksibel (angka/huruf/Romawi); validasi; import transaksional dengan progress.

**Test:** import referensi `SPH_KRI_OWA.xls` (fixture test pakai file tiruan di `testdata/`); kasus valid, invalid, hierarchy, mapping.

**Acceptance:** tidak ada import tanpa preview; rollback bekerja. Dokumentasi `docs/import-excel.md`.

## PHASE 9 — Export

**Scope:** generator Excel (layout kolom B–K, merge, border, format Rupiah, repeating header, Pindahan, Terbilang, ttd, page setup A4 landscape) + generator PDF (A4 portrait/landscape, multi-page, header/footer, page number, terbilang, ttd).

**Test:** hasil export dibandingkan terhadap struktur referensi; multi-page tidak memotong tabel.

**Acceptance:** dokumen profesional & konsisten dengan format perusahaan. Dokumentasi `docs/export-document.md`.

## PHASE 10 — Backup & Restore

**Scope:** backup manual, restore (konfirmasi → backup dulu → restore → validate → reload), auto backup (saat tutup + harian), retention 10, validasi file backup.

**Test:** backup → modify → restore → verify.

**Acceptance:** restore aman (rollback jika gagal). Dokumentasi `docs/backup-restore.md`.

## PHASE 11 — Polish

**Scope:** UI/UX final (loading, empty, error state), keyboard shortcut (Ctrl+N/S/F/P, Esc), performance, accessibility, konsistensi design system biru flat + orange flat.

**Acceptance:** semua halaman konsisten; shortcut berfungsi; tidak ada regresi test.

## PHASE 12 — Release

**Scope:** build Windows executable + installer + portable build; versioning; installation guide; test di environment bersih; **push seluruh riwayat ke `https://github.com/RizaldiP/sph-manager.git`**.

**Acceptance:** aplikasi terinstall & jalan di Windows bersih tanpa internet; repo ter-push lengkap.

---

## Aturan Kerja per Sesi (dari Master Specification §74)

1. Baca `README.md`, `docs/development-plan.md`, `docs/requirements.md`, `docs/architecture.md`, `docs/business-rules.md`
2. Tentukan fase aktif; inspect implementasi existing
3. Kerjakan **hanya** fase aktif
4. Run tests → formatter → build → update dokumentasi
5. Berhenti; tunggu instruksi
