# Analisis File Excel Referensi — SPH

> Dokumen ini adalah hasil analisis menyeluruh file Excel referensi sebagai dasar desain database dan generator dokumen aplikasi SPH Manager.

---

## 1. File yang Dianalisis

| Properti | Nilai |
|---|---|
| Nama file | `SPH_KRI_OWA.xls` |
| Ukuran | 792.064 byte |
| Format | XLS lama (BIFF) |
| Metode analisis | Inspeksi COM Excel mode read-only (nilai, formula, merge cell, page setup, shapes) + ekstraksi internal paket untuk media/gambar |

---

## 2. Sheet

| Sheet | Used Range | Print Area | Konten |
|---|---|---|---|
| `SPH` | 113 baris × 13 kolom | `A1:L64` | SPH Perbaikan Sistem Kontrol CWU I & II — KRI.SBY-591 TA.2026 |
| `SPH #2` | 166 baris × 12 kolom | `A1:K117` | SPH Perbaikan AHU/LBH III, V, VI & Sistem Pendukung — KRI.OWA-591 TA.2026 |

---

## 3. Struktur Setiap Sheet

Setiap sheet terdiri atas 6 zona:

```text
(a) KOP            : nama perusahaan + kota
(b) JUDUL          : judul dokumen + kapal/tahun (merge 8 kolom)
(c) HEADER TABEL   : 3 baris header dengan merge
(d) ISI HIERARKIS  : main point → sub point (indentasi via kolom)
(e) REKAP          : Sub Total → Total → Terbilang
(f) TANDA TANGAN   : blok ttd (floating shapes, posisi setelah tabel)
```

---

## 4. Struktur Header

### 4.1 Kop

| Baris | Isi |
|---|---|
| R1 | `PT. GANESHA ENERGI INDONESIA` (kolom D, teks besar) |
| R2 | `SURABAYA` |

### 4.2 Judul Dokumen

| Barir | Isi | Merge |
|---|---|---|
| R5 | `DAFTAR PENAWARAN HARGA PERBAIKAN ...` / `SURAT PENAWARAN HARGA PERBAIKAN ...` | 8 kolom × 1 baris |
| R6 | `KRI.SBY-591 TA.2026` / `KRI.OWA-591 TA.2026` | 8 kolom × 1 baris |

### 4.3 Header Tabel (R8–R10)

| Kolom Excel | Header (R8) | Sub-header (R9) | No. cetak (R10) |
|---|---|---|---|
| B | NO (merge vertikal) | — | 1 |
| C–E | URAIAN KEGIATAN (merge 3×2) | — | 2 |
| F | JML (merge vertikal) | — | 3 |
| G | SAT (merge vertikal) | — | 4 |
| H | HARGA | SATUAN | 5 |
| I–K | PENAWARAN HARGA (Rp) | JASA / MAT / JML | 6 / 7 / 8 |

Baris R10 berisi **penomoran cetak kolom** (1–8), bukan data. Pada sheet `SPH`, baris ini **berulang di R41** sebagai header halaman kedua, diikuti baris **"Pindahan"**.

---

## 5. Struktur Main Point

Ditemukan **dua gaya penomoran** yang berbeda antar sheet:

### Gaya A (sheet `SPH`)
```text
A  Laksanakan Perbaikan Sistem Penjalan dan Ganti baru Material yang rusak   ← grup huruf besar
   1. Laksanakan Perbaikan Sistem Penjalan meliputi :                        ← main point angka (kolom C)
   2. Laksanakan Ganti Baru Sistem Penjalan :
   3. Bongkar pasang dan Setting Program          ← main point bisa langsung punya qty/harga
```

### Gaya B (sheet `SPH #2`)
```text
I    Laksanakan Perbaikan dan Service AHU III        ← main point Romawi (kolom B)
II   Laksanakan Service dan Perbaikan Drain Pan ...
...
XII  Service Kelas B Motor Blower AHU VI
```

Main point **boleh** langsung memiliki qty/unit/harga (contoh: R49 sheet 1 — "Bongkar pasang dan Setting Program", 2 Set, jasa @11.000.000).

---

## 6. Struktur Sub Point

- **Gaya A**: huruf kecil `a.`, `b.`, … di kolom D (teks di kolom E); kadang sub-sub angka `1.`, `2.`, `3.` di kolom E (teks lebih dalam lagi).
- **Gaya B**: huruf kecil `a.`–`g.` di kolom C (teks di kolom D).

Indentasi hierarki direpresentasikan lewat **kolom yang berbeda**, bukan indent teks dalam satu sel.

---

## 7. Struktur Quantity

- Kolom F (`JML`).
- Selalu **integer tanpa desimal** (nilai teramati: 1, 2, 3, 4, 5, 9).

---

## 8. Struktur Unit

- Kolom G (`SAT`), teks bebas.
- Nilai teramati: `giat`, `unit`/`Unit`, `bh`, `set`/`Set`, `roll`.
- Tidak baku — kapitalisasi campur.

---

## 9. Struktur Harga

| Kolom | Arti | Isi |
|---|---|---|
| H | HARGA SATUAN | integer |
| I | JASA | nilai total jasa baris |
| J | MAT | nilai total material baris |
| K | JML | total baris (I+J) |

Pola pengisian: item `Service/Perbaikan` diisi di kolom JASA; item `Ganti baru` diisi di kolom MAT. **Tidak 100% konsisten** (lihat §16).

---

## 10. Formula

| Kebutuhan | Formula teramati |
|---|---|
| Nilai jasa baris | `=H13*F13` (harga satuan × qty) |
| Nilai material baris | `=F26*H26` |
| Total baris | `=SUM(I13,J13)` atau `=J12+I12` |
| Subtotal jasa/material | `=SUM(I11:I33)`, `=SUM(J11:J33)` |
| Subtotal total | `=SUM(K13:K34)` |
| Pindahan (carry antar halaman) | `=I35` |
| Total akhir | `=K51` (dari subtotal) |

---

## 11. Format Angka

- **Semua sel ber-format `General`** — tanpa format Rupiah, tanpa pemisah ribuan.
- Nilai disimpan sebagai integer penuh (mis. `120725000`).

---

## 12. Format Dokumen

| Aspek | Nilai |
|---|---|
| Kertas | A4 (paper size 9) |
| Orientasi | **Landscape** (kedua sheet) |
| Skala | Fit to 1 halaman lebar × 1 halaman tinggi |
| Print area | Eksplisit (`A1:L64`, `A1:K117`) |
| Header halaman | `&P` (nomor halaman, tengah) |
| Footer | Tidak ada |

### Blok Tanda Tangan (floating shapes, dalam print area)

Terdapat TextBox + 2 gambar pada H54 (sheet 1) / H107 (sheet 2):

```text
Surabaya, 13 Juli 2026
PT. Ganesha Energi Indonesia

[gambar tanda tangan]

M a t a w i
Direktur
```

### Logo

Gambar logo bulat perusahaan ditemukan di media internal file. **Logo aplikasi diambil dari Settings (upload pengguna)** — tidak direplikasi dari file referensi.

### Terbilang

Baris `Terbilang : <jumlah dalam huruf>` — diketik **manual** di referensi (rawan salah); aplikasi akan men-generate otomatis.

---

## 13. Data yang Bisa Dijadikan Master

- Nama pekerjaan dan sub pekerjaan (uraian kegiatan).
- Satuan umum: `giat`, `unit`, `bh`, `set`, `roll`.
- Harga satuan sebagai **default awal** master pekerjaan.
- Pola kategorisasi: `Service/Perbaikan` → jasa; `Ganti baru` → material.

---

## 14. Data yang Harus Menjadi Snapshot

Semua data berikut **disalin penuh** ke dokumen SPH saat dibuat:

- Uraian main point & sub point (nama, deskripsi).
- Quantity, satuan.
- Harga satuan jasa & material, nilai total per baris.
- Judul dokumen, kapal, tanggal, terbilang.

Alasan: harga master dapat berubah kapan pun; dokumen SPH lama **tidak boleh ikut berubah**.

---

## 15. Perbedaan Format Antar-Sheet

| Aspek | Sheet `SPH` | Sheet `SPH #2` |
|---|---|---|
| Penomoran main point | Grup huruf (A) → angka (1., 2.) | Romawi (I–XII) |
| Kedalaman hierarki | 3 level (A / 1 / a / 1.) | 2 level (I / a) |
| Kolom teks utama | D / E | C / D |
| Pindahan antar halaman | Ada (R41–R42) | Tidak ada |
| Jumlah kolom | 13 | 12 |

---

## 16. Masalah / Ketidakkonsistenan

1. Penempatan **jasa vs material** tidak konsisten antar baris sejenis.
2. Gaya formula campur (`SUM(I,J)` vs `I+J`) — semantik sama.
3. Format angka `General` — tidak profesional untuk dokumen cetak.
4. Typo pada sumber (`Pebaikan`, `toutch`, `protektsi`, `penjalan`).
5. Satuan tidak baku (kapitalisasi campur).
6. **Tidak ada nomor SPH** pada dokumen — hanya judul.
7. Tidak ada kolom diskon/PPN.
8. Baris kosong tersebar; used range > print area.
9. Terbilang diketik manual.
10. Penomoran hierarki berbeda antar dokumen — parser import harus fleksibel.

---

## 17. Rekomendasi Struktur Database

Mengikuti 14 tabel pada Master Specification, dengan catatan tambahan dari analisis:

1. `sph_items` & `sph_sub_items` menyimpan **snapshot** (`name_snapshot`, `description_snapshot`) + `service_unit_price` + `material_unit_price` + hasil hitung.
2. `sph_sub_items.weight` untuk pembobotan (0–100); mode pembobotan per main point (lihat `business-rules.md` BR-02/BR-03).
3. `sph_documents`: `terbilang` digenerate otomatis; `signer_name`/`signer_position` diambil dari settings saat generate.
4. `work_sub_items.sequence` untuk urutan tampil; `difficulty_weight` untuk pembobotan default.
5. `settings`: company profile, path logo, format nomor SPH, penandatangan, catatan default dokumen.
6. `audit_logs`: semua aksi penting (create/update/delete/finalize/export/duplicate/restore).
7. Semua master memakai `is_active` + `deleted_at` (soft delete).

---

## 18. Rekomendasi Generator

### Export Excel
- Replikasi layout kolom B–K dengan merge sesuai referensi.
- Format Rupiah profesional (`Rp #,##0`) — **lebih rapi dari sumber**.
- Header tabel berulang tiap halaman cetak + baris "Pindahan".
- Terbilang otomatis + blok tanda tangan + logo dari settings.
- Page setup: A4 landscape, fit 1 halaman lebar, print area eksplisit, header `&P`.

### Export PDF
- A4 landscape; header halaman berisi nomor halaman.
- Tabel multi-halaman dengan header berulang dan carry-forward "Pindahan".
- Rekap Sub Total → Total → Terbilang → blok ttd di akhir.

### Terbilang
- Konverter angka→huruf Indonesia otomatis dari grand total (dengan unit test).
