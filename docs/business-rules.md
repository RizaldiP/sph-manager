# Business Rules — SPH Manager Offline

> Aturan bisnis yang mengikat implementasi backend. Semua perhitungan & validasi berada di backend (Go), bukan di Vue component.

---

## BR-01 — SNAPSHOT (aturan terpenting)

Saat master dipakai membuat SPH, data **disalin penuh** ke dokumen:

```text
MASTER → COPY/SNAPSHOT → SPH
```

- Perubahan master **tidak pernah** mengubah SPH yang sudah ada.
- Contoh: Master `Repair PLC = Rp10.000.000` → SPH dibuat `Rp10.000.000` → master diubah `Rp15.000.000` → SPH lama **tetap Rp10.000.000**; SPH baru memakai `Rp15.000.000`.
- Wajib automated test untuk aturan ini.

## BR-02 — MODE PENENTUAN HARGA MAIN POINT

Per main point berlaku salah satu dari dua mode:

1. **HARGA_LANGSUNG** — main point punya qty × harga sendiri (sesuai referensi R49 sheet 1: "Bongkar pasang dan Setting Program", 2 Set @11.000.000).
2. **PEMBOBOTAN** — main point punya nilai total; sub point dialokasikan via weight (BR-03).

Formula dasar per baris (jika baris punya harga sendiri):

```text
service_total  = quantity × service_unit_price
material_total = quantity × material_unit_price
total          = service_total + material_total
```

Total main point pada mode PEMBOBOTAN = Σ total sub point-nya (harus identik setelah rounding).

## BR-03 — PEMBOBOTAN (WEIGHT)

- Weight per sub point: `0 ≤ weight ≤ 100`, tipe data aman (basis poin, bukan float murni).
- Σ weight per main point **wajib = 100**.
- Nilai sub point = `main_point_value × weight / 100`.
- Jika Σ weight ≠ 100: tampilkan warning + total saat ini + selisih, dan **larang finalisasi**.
- Main point tanpa sub point tidak memakai weight.

## BR-04 — ROUNDING DETERMINISTIK

- Satuan uang: **integer Rupiah**. Floating point bukan sumber kebenaran.
- Alokasi pembobotan memakai metode **largest remainder** agar Σ nilai sub point **tepat sama** dengan nilai main point.
- Contoh: Rp100 dibagi 3 bagian @33,33% → 33 + 33 + 34 (sisa diberikan ke baris dengan pecahan terbesar, tie-breaker: urutan sequence).
- Wajib unit test: total sub-item harus sama dengan Main Point.

## BR-05 — SOFT DELETE

- Master data memakai `is_active` + `deleted_at`.
- Data yang pernah dipakai oleh SPH **tidak boleh dihapus permanen** (melarang histori rusak).
- Delete = set `is_active=false` + `deleted_at`; data nonaktif tidak muncul di pemilihan default tetap bisa dicari untuk referensi.

## BR-06 — VALIDASI FINALISASI

Finalisasi (DRAFT → REVIEW/FINAL) hanya boleh jika:

1. Nomor SPH terisi & unik;
2. Tanggal terisi;
3. Customer terisi;
4. Minimal satu pekerjaan;
5. Semua quantity valid (> 0);
6. Semua harga valid (≥ 0);
7. Jika ada main point mode PEMBOBOTAN → Σ weight = 100%;
8. Grand total = Σ semua baris (verifikasi konsistensi).

## BR-07 — PENOMORAN SPH

- Format configurable: contoh `SPH/GEI/VIII/2026/001` (prefix, bulan Romawi, tahun, sequence, separator).
- Sequence per template format (per tahun/bulan sesuai konfigurasi).
- Anti-duplikat: unique constraint di database + generator transaksional.

## BR-08 — STATUS LIFECYCLE

```text
DRAFT → REVIEW → FINAL → SENT → ACCEPTED
                             └→ REJECTED
DRAFT/REVIEW → CANCELLED
```

- Transisi hanya mengikuti diagram di atas.
- Dokumen berstatus FINAL ke atas **tidak bisa diedit isinya** (hanya status & catatan administratif).
- Finalisasi mencatat `finalized_at`.

## BR-09 — DUPLICATE SPH

- Duplicate menghasilkan snapshot **baru penuh** (semua item & sub item disalin ulang).
- Source tidak berubah sedikit pun.
- Nomor baru digenerate; status dokumen baru = DRAFT.

## BR-10 — REVISION

- `SPH-001 Rev 0`, `Rev 1`, `Rev 2` …
- Revision baru = salinan snapshot dari revision terakhir, lalu boleh diedit.
- Revision lama **tetap dapat dilihat**; tidak ada overwrite histori.

## BR-11 — TERBILANG

- Jumlah dalam huruf digenerate **otomatis** dari grand total (Bahasa Indonesia): "Seratus Enam Puluh Tujuh Juta Tiga Ratus Ribu Rupiah".
- Wajib unit test konversi.

## BR-12 — CARRY-FORWARD (PINDAHAN)

- Baris "Pindahan" antar halaman adalah **urusan presentasi generator** (Excel/PDF), bukan data tersimpan.
- Halaman lanjutan menampilkan subtotal halaman sebelumnya sebagai "Pindahan", sesuai referensi R41–R42.

## BR-13 — AUDIT LOG

- Catat aksi: CREATE, UPDATE, DELETE, FINALIZE, EXPORT, DUPLICATE, RESTORE.
- Field: timestamp, action, entity, entity_id, description.

## BR-14 — BACKUP & RESTORE

- Backup manual + auto backup (saat tutup aplikasi & harian) + retention (default 10 file terakhir).
- Restore wajib melalui: konfirmasi → backup kondisi sekarang → restore → validate → reload.
- Restore dalam transaksi; gagal = rollback, database lama tetap utuh.

## BR-15 — ERROR HANDLING

- Jangan tampilkan error database mentah (`SQLITE_CONSTRAINT_FOREIGNKEY`).
- Contoh pesan yang benar: "Data tidak dapat dihapus karena masih digunakan oleh dokumen SPH."
- Detail teknis masuk structured log.

## BR-16 — TRANSACTION

Wajib transaction untuk: create SPH, duplicate SPH, revision, finalisasi, import, restore. Gagal di tengah = ROLLBACK penuh.
