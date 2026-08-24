# Development Plan — SPH Manager Offline

> Rencana kerja per fase sesuai Master Specification. Setiap fase ditutup dengan: test ✓ formatter ✓ build ✓ dokumentasi ✓ lalu **BERHENTI** menunggu instruksi `LANJUT PHASE X`.

## Status Ringkas

| Fase | Nama | Status |
|---|---|---|
| 0 | Analisis & Dokumentasi | ✅ Selesai |
| 1 | Foundation (Wails + Go + Vue + SQLite) | ✅ Selesai |
| 2 | Database (migration + model) | ✅ Selesai |
| 3 | Master Pekerjaan | ✅ Selesai |
| 4 | Template | ✅ Selesai |
| 5 | SPH Builder (+ CRUD Customer & Kapal) | ✅ Selesai |
| 6 | Pembobotan & Rounding | ✅ Selesai |
| 7 | Kombinasi Multi Pekerjaan | ✅ Selesai |
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

## PHASE 1 — Foundation ✅

**Scope:** install Wails CLI; scaffold proyek Wails v2; Go backend + SQLite (GORM) + migration framework + config + structured logging; Vue 3 + TypeScript + Pinia + Vue Router + Tailwind CSS (design system dasar: biru flat + orange flat); layout Sidebar+Topbar; halaman Dashboard stub.

**Hasil:** Wails v2.15.0; GORM 1.31 + glebarez/sqlite (pure-Go, tanpa CGO); slog multi-sink; config JSON AppData; layout sidebar/topbar + 14 route; design token `brand-*`/`accent-*`; binding `Health()` teruji.

**Test:** `wails build` sukses → `SPHManager.exe` jalan (proses hidup, database WAL dibuat, log tertulis); `wails dev` sukses (vite 5173 + WebView2 environment created).

**Acceptance:** aplikasi jalan offline; lint/format bersih; build dev sukses.

## PHASE 2 — Database

**Scope:** migration seluruh tabel: `categories`, `work_items`, `work_sub_items`, `templates`, `template_items`, `customers`, `vessels`, `materials`, `sph_documents`, `sph_items`, `sph_sub_items`, `sph_revisions`, `audit_logs`, `settings` + index + FK + soft delete.

**Test:** database kosong → migration → semua tabel terbentuk → FK valid (cek pragma foreign_key_list).

**Acceptance:** migration idempotent, test integrasi DB hijau. Dokumentasi `docs/database.md`.

## PHASE 3 — Master Pekerjaan ✅

**Scope:** CRUD kategori, work item, sub work item; search; reorder (drag-and-drop); aktif/nonaktif; validasi; error handling ramah; halaman UI sesuai FR-U4.

**Hasil:** layer `internal/repositories` + `internal/services` (validasi, error ramah BR-15, audit log BR-13, transaksi BR-16); 18 binding method (`app_master.go`): kategori (list/create/update/set-active/delete/reorder), pekerjaan (+detail berisi sub), sub-pekerjaan; field `sequence` baru di `categories` & `work_items` untuk reorder; halaman **Kategori** & **Master Pekerjaan** (filter kategori + hitungan, search server-side, toggle nonaktif, drag-and-drop urutan native HTML5 via `useDragSort`, modal form, confirm dialog, empty state); store Pinia `master`; komponen `AppModal` & `ConfirmDialog`; quick action Dashboard "Tambah Pekerjaan" aktif.

**Test:** `go test ./internal/...` hijau — 15 unit/integrasi termasuk skenario acceptance: `Electrical → Repair Control Panel → Inspection…Commissioning → ubah urutan → nonaktifkan`; hapus kategori/pekerjaan yang masih berisi anak ditolak dengan pesan ramah; reorder parsial ditolak; kode duplikat ditolak soft-delete-aware. `vue-tsc --noEmit` bersih; `wails build` sukses → `SPHManager.exe`.

**Acceptance:** CRUD + reorder tersimpan ✓; test unit + integrasi hijau ✓.

## PHASE 4 — Template ✅

**Scope:** CRUD template + template_items; duplicate; reorder; aktif/nonaktif; pakai template saat membuat SPH.

**Hasil:** repository + service template (validasi, error ramah BR-15, audit BR-13, transaksi BR-16); 9 binding method (`app_template.go`): list/detail (isi terurut + data pekerjaan lengkap: kategori & sub)/create/update/set-items/duplicate/set-active/delete/reorder; kolom `sequence` baru di `templates` (AutoMigrate idempotent); halaman **Template** (search server-side, toggle nonaktif, drag-and-drop urutan, editor isi template dengan pemilih pekerjaan per kategori + urutan drag + catatan per baris, duplikat sekali klik); quick action Dashboard "Tambah Template" aktif. **Kode master kini auto-generate sistem** (`KAT-/PEK-/SUB-/TPL-` + nomor berurutan; input kode dinonaktifkan di semua form; update tanpa kode mempertahankan kode lama). Perbaikan integritas: index unik parsial mengecualikan kode kosong (drop+recreate idempotent), hapus pekerjaan ditolak bila masih dipakai template hidup.

**Test:** `go test ./internal/...` hijau — skenario acceptance: template "Repair PLC" berisi 8 langkah → duplikat (nama bersuffix, isi & urutan sama, kode baru) → nonaktifkan; validasi SetItems (duplikat/hantu/kosong/catatan >500); regresi dua entitas tanpa kode hidup berdampingan; guard hapus pekerjaan terpakai template; reorder penuh wajib. `vue-tsc --noEmit` bersih; `wails build` sukses → `SPHManager.exe`.

**Acceptance:** reuse template berfungsi (detail siap dikonsumsi SPH builder Phase 5) ✓; test hijau ✓.

## PHASE 5 — SPH Builder ✅

**Scope:** wizard 8 langkah (Info → Pilih Pkerjaan → Susun Main Point → Sub Point → Costing → Validasi → Preview → Simpan/Generate); snapshot penuh saat simpan; edit dalam SPH; duplicate; revision; daftar SPH; dashboard.

**Test:** satu SPH lengkap end-to-end di UI.

**Acceptance:** SPH tersimpan sebagai snapshot; validasi & transaksi sesuai BR-06/BR-16. Dokumentasi `docs/sph-flow.md`.

**Hasil:**
- Backend: `sph_repository.go` (list/detail/penomoran/revisi/statistik), `customer_repository.go` (FR-M5/M6 + guard pemakaian SPH), `sph_service.go` (snapshot BR-01, validasi BR-06, lifecycle BR-08, duplicate BR-09, revision BR-10, terbilang BR-11, dashboard FR-U5), `customer_service.go`, `terbilang.go`, `sph_number.go` (format via settings `sph_number_format`, default `SPH/GEI/{ROMAN}/{YYYY}/{SEQ}`, urutan per periode).
- Binding baru: ListSph, GetSph, CreateSph, UpdateDraftSph, DeleteSph, SetSphStatus, DuplicateSph, CreateSphRevision, DashboardStats, Terbilang, ListCustomers, CreateCustomer, UpdateCustomer, SetCustomerActive, DeleteCustomer, CreateVessel, UpdateVessel, SetVesselActive, DeleteVessel.
- Frontend: wizard 8 langkah (`BuatSphPage.vue`, sumber Master/Template/SPH Lama/Manual), daftar 3 mode (`SphListPage.vue`: Semua/Draft/Final), detail + aksi status & revisi (`SphDetailPage.vue`), Master Data Customer & Kapal (`DataPartnerPage.vue`), statistik + 5 SPH terbaru di Dashboard.
- Test hijau: terbilang, snapshot vs perubahan master, penomoran berurutan, transisi ilegal, finalisasi + finalized_at, lock edit non-draft, duplikat nomor baru/revision rev+1 + histori, guard hapus draft/customer/kapal, scope daftar, statistik.
- Penyesuaian pasca-ulasan: harga sub point **di-roll-up** ke Jumlah main point-nya (jasa & material terpisah) sehingga ikut dalam Subtotal/Grand Total; sub point tampil sebagai baris tabel sejajar kolom Qty/Sat/Jasa/Mat./Jumlah di detail & preview.

## PHASE 6 — Pembobotan & Rounding ✅

**Scope:** mode PEMBOBOTAN per main point; input weight; alokasi nilai sub point; rounding largest remainder (BR-04); warning selisih; larang finalisasi jika ≠ 100%.

**Test:** `10 + 15 + 20 + 25 + 30 = 100`; kasus <100 dan >100; kasus pembulatan (mis. 3 × 33,33%).

**Acceptance:** Σ sub = main point tepat; unit test hijau.

**Hasil:**
- Desain terkunci (keputusan user): satu **bobot gabungan %** per sub point; bobot mengalokasikan pool jasa dan pool material secara terpisah, masing-masing dengan largest remainder sehingga Σ alokasi = pool tepat.
- Backend `internal/services/pembobotan.go`: `allocateLargestRemainder` (aritmetika integer, dasar floor, sisa ke pecahan terbesar, tie-break urutan baris) + `allocateWeightedSubs` (validasi ≥1 sub, bobot 1–100, qty>0; set `weight`/`allocated_value`/total per sub; harga satuan sub dinolkan).
- `buildItems` menerima PEMBOBOTAN (tidak lagi ditolak); draft **boleh disimpan dengan Σ≠100** (alokasi proporsional thd Σ aktual), tapi finalisasi REVIEW/FINAL ditolak backend sampai Σ=100 — pesan memuat persentase & selisih. Verifikasi ulang alokasi tersimpan saat finalisasi (deteksi data diubah manual).
- Frontend wizard: dropdown "Mode Harga" per main point (langkah 3); langkah 4 menampilkan input Bobot % (+pratinjau nilai alokasi) untuk mode pembobotan dengan badge Σ bobot live (hijau saat 100%, amber saat selisih); prefill bobot dari master `difficulty_weight`; preview langkah 7 menampilkan badge % dan jumlah hasil alokasi; langkah 6 memisahkan checks blokir vs warning pembobotan non-blokir.
- Detail SPH: badge % pada baris sub mode pembobotan.
- Test hijau: tabel alokasi (kasus BR-04, tie-break, 10/15/20/25/30), end-to-end simpan+finalisasi, penolakan Σ<100/>100, validasi bobot, deteksi tamper `allocated_value`.
- Pasca-ulasan (FR-M7 dipercepat): master **Material** diaktifkan — repository/service/binding CRUD (`ListMaterials/CreateMaterial/UpdateMaterial/SetMaterialActive/DeleteMaterial`), kode auto-generate sistem `MAT-xxx`, search nama/kode/supplier, soft delete, audit log; halaman **Material** menggantikan placeholder `/data/material` (search debounced, toggle nonaktif, modal form tanpa input kode, konfirmasi hapus). Wizard SPH mendapat **pemilih material**: tombol ⌕ di kolom Mat. (langkah 3 & sub point harga langsung langkah 4) mengisi nama, satuan, dan harga material baris dari master sehingga ikut perhitungan; tanpa tautan permanen ke master (snapshot nilai).

## PHASE 7 — Kombinasi Multi Pekerjaan ✅

**Scope:** pilih banyak pekerjaan → satu SPH; drag-and-drop urutan final; penggabungan dari master + template + SPH lama.

**Test:** Repair AMS + Repair PLC + Repair Sensor + Testing + Calibration → 1 SPH berurut 1–5.

**Acceptance:** urutan tersimpan; snapshot benar.
**Hasil:**
- Perbaikan: kolom `sph_items.sequence`/`sph_sub_items.sequence` kini terisi oleh `buildItems` (sebelumnya selalu 0 padahal detail mengurutkan `sequence asc`). Update draft me-renumber otomatis via ReplaceItems; duplicate/revisi menyalin sequence (`cloneItems`).
- Wizard langkah 3: drag-and-drop urutan main point (HTML5 DnD, handle di header kartu, reuse `useDragSort` yang digeneralisasi dengan parameter `keyOf` agar kompatibel halaman master/template); tombol ↑↓ tetap tersedia.
- Penggabungan sumber: **SPH lama kini merge** (append baris unik per `workItemId`, tidak lagi menimpa pilihan existing), template juga merge; umpan balik "N baris digabungkan, M dilewati" di langkah 2; label tombol "+ Gabungkan".
- Test hijau (`kombinasi_test.go`): 5 pekerjaan tersimpan berurut seq 1–5 + sub 1..M, reorder update draft → sequence ter-renumber, duplicate mempertahankan urutan & sequence.
- Pasca-fase (melengkapi placeholder "Phase 7"): **halaman Pengaturan aktif** — service `settings_service.go` (get/update/preview nomor/logo, validasi `{SEQ}`, audit BR-13), binding `app_settings.go` (`GetSettings/UpdateSettings/PreviewSphNumber/PickLogo/ClearLogo/LogoDataUrl`; dialog file via Wails runtime, logo disalin ke `%AppData%\sph-manager\assets`), halaman `PengaturanPage.vue` (profil perusahaan, logo + pratinjau, format penomoran SPH dengan chips placeholder & contoh nomor live, penandatangan, catatan default); config field baru `assets_dir`. Test hijau: default & roundtrip, validasi format, preview, logo persist, audit log.

## PHASE 8 — Import Excel ✅

**Scope:** reader XLS/XLSX; pilih sheet; preview wajib; mapping kolom adjustable; deteksi hierarki fleksibel (angka/huruf/Romawi); validasi; import transaksional dengan progress.

**Test:** import referensi `SPH_KRI_OWA.xls` (fixture test pakai file tiruan di `testdata/`); kasus valid, invalid, hierarchy, mapping.

**Acceptance:** tidak ada import tanpa preview; rollback bekerja. Dokumentasi `docs/import-excel.md`.
**Hasil:**
- Package `internal/importers`: `reader.go` (xlsReader untuk .xls BIFF + excelize untuk .xlsx, grid dinormalkan persegi), `parser.go` (splitMarker romawi/huruf/angka dengan aturan delimiter ketat; state machine dua level dengan **aturan deret induk** — angka `lastMain+1` = induk baru, angka lain di blok sub diratakan jadi sub; baris tanpa penanda = unknown yang wajib diputuskan pengguna; konversi nilai total → harga satuan via qty; parser angka Indonesia/internasional), `preview.go` (SheetPreview ≤200×20 + saran mapping).
- `SuggestMapping`: band header multi-baris (judul gabungan + sub-judul) dari kata kunci URAIAN/JML/SAT/JASA/MAT, skip baris indeks kolom; layout referensi SPH_KRI_OWA.xls (NO=1, URAIAN=2, JML=5, SAT=6, HARGA SATUAN=7, JASA=8, MAT=9, data mulai r10) terdeteksi otomatis.
- `services/import_service.go`: ImportWorkItems satu transaksi — append ke kategori tujuan, sequence lanjut max+1, kode PEK-/SUB- auto, sub tanpa induk ditolak, audit BR-13 per entri, callback progress membatalkan → rollback penuh (BR-16); ValidateRows mengeblokir unknown belum diputuskan & baris bermasalah.
- Binding `app_import.go` (PickImportFile/ListImportSheets/PreviewImportSheet/ParseImportRows/ValidateImportRows/RunWorkItemImport) + event progress Wails `import:progress`/`import:done`.
- Halaman `ImportPage.vue` wizard 4 langkah menggantikan placeholder `/impor-ekspor`: file dialog → sheet & mapping editable (select kolom A..Z+, nameSpan, headerRows, mode nilai total, kategori tujuan) → pratinjau grid dengan penanda baris data & klasifikasi manual baris merah → ringkasan hasil; progress bar realtime.
- Test hijau: parser Gaya A/B & splitMarker & angka Indonesia, reader XLSX fixture excelize runtime + integrasi opsional file referensi asli (skip bila absen), service happy path/rollback/blokir/sub-tanpa-induk. Fixture dibuat via excelize saat runtime (tanpa binary testdata); dokumentasi `docs/import-excel.md`.
- Pasca-ulasan (harga tidak muncul): layout referensi ternyata mencampur konvensi — banyak baris menaruh harga di kolom **HARGA SATUAN** dengan JASA/MAT kosong. Ditambahkan `UnitPriceCol` + `UnitPriceAs` pada `ColumnMapping`: fallback bila JASA & MATERIAL keduanya kosong, arah Jasa/Material dipilih di wizard; `SuggestMapping` mengenali header dua baris HARGA/SATUAN via teks gabungan band-kolom (referensi: UnitPriceCol=7 tanpa menelan kolom SAT=6). Test file asli kini mengaserti harga baris terdeteksi (Sensor Oli = 2.200.000) sebagai regresi permanen.

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
