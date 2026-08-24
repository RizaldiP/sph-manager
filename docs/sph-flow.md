# Alur Kerja SPH (Phase 5)

Dokumen ini merangkum alur end-to-end pembuatan dan siklus hidup dokumen SPH,
beserta aturan bisnis yang dijalankan sistem (lihat `business-rules.md` untuk
definisi lengkap BR-*).

## 1. Prasyarat: Master Data

Sebelum menyusun SPH, siapkan data di halaman **Master Data → Customer & Kapal**
(FR-M5, FR-M6):

- **Customer**: nama wajib, kode opsional (unik bila diisi).
- **Kapal** milik satu customer: nama wajib, kode/nomor/jenis opsional.

Aturan penghapusan (FR-A2 / BR-02):

- Customer yang masih dipakai dokumen SPH → tidak bisa dihapus (nonaktifkan saja).
- Customer yang masih punya kapal terdaftar → tidak bisa dihapus.
- Kapal yang masih dipakai dokumen SPH → tidak bisa dihapus.
- Mengubah/menghapus master **tidak pernah** mengubah dokumen SPH lama (snapshot penuh, BR-01).

## 2. Wizard 8 Langkah (FR-S1)

Halaman **SPH → Buat SPH** (`/sph/baru`):

| # | Langkah | Isi |
|---|---------|-----|
| 1 | Info | Tanggal (wajib), customer (wajib), kapal, proyek, subjek, referensi, lokasi, masa berlaku, PIC, catatan |
| 2 | Sumber | Tab: **Master Pekerjaan**, **Template**, **SPH Lama**, **Manual** — semua nilai disalin sebagai snapshot |
| 3 | Main Point | Urutan baris (↑/↓), nama, deskripsi, qty, satuan, harga jasa & material |
| 4 | Sub Point | Rincian per main point; harganya ikut terhitung ke Jumlah main point-nya |
| 5 | Costing | Subtotal jasa/material, grand total otomatis (`qty × harga` main + sub point, dibulatkan ke Rp terdekat — BR-04) + terbilang |
| 6 | Validasi | Checklist pra-finalisasi sesuai BR-06 |
| 7 | Preview | Tampilan dokumen seperti cetak dengan nomor preview `SPH/GEI/{ROMAN}/{YYYY}/XXX` |
| 8 | Simpan | Simpan sebagai **Draft** dengan nomor otomatis |

Draft yang sudah tersimpan dapat diedit lewat `/sph/:id/edit` selama statusnya
masih Draft atau Review (BR-08).

## 3. Penomoran Dokumen (BR-07)

- Format default: `SPH/GEI/{ROMAN}/{YYYY}/{SEQ}` (contoh `SPH/GEI/VIII/2026/001`).
- Placeholder yang didukung: `{YYYY}`, `{MM}`, `{MM_ROMAN}` (angka), `{ROMAN}` (bulan romawi), `{SEQ}`.
- Format dapat diubah melalui settings key `sph_number_format` (UI Pengaturan menyusul).
- `{SEQ}` berurutan **per periode tahun-bulan** dari tanggal dokumen, bukan tanggal simpan.
- Nomor digenerate di dalam transaksi penyimpanan; unik di seluruh dokumen hidup.

## 4. Snapshot Penuh (BR-01)

Saat simpan, sistem menyalin seluruh nilai ke tabel `sph_items` /
`sph_sub_items` (name/description/qty/unit/harga/total). Konsekuensi:

- Perubahan master pekerjaan atau template setelahnya **tidak mengubah** dokumen.
- **Roll-up sub point:** kolom Jumlah sebuah main point = qty×harga main point +
  Σ (qty×harga) seluruh sub point-nya (jasa & material dihitung terpisah,
  dibulatkan per baris ke Rp terdekat — BR-04). Subtotal jasa/material dan
  grand total dokumen dengan demikian mencakup sub point.
- Grand total dihitung ulang server-side saat setiap kali draft disimpan;
  finalisasi memverifikasi konsistensi baris (termasuk roll-up sub point) vs
  total (BR-06).

## 5. Siklus Hidup Status (BR-08)

```
DRAFT ──→ REVIEW ──→ FINAL ──→ SENT ──→ ACCEPTED
   │          │                    └──→ REJECTED
   │          └──→ CANCELLED
   └──→ CANCELLED
```

- **Review/Finalisasi** menjalankan validasi BR-06: nomor terisi, tanggal terisi,
  customer dipilih, ≥1 pekerjaan, qty > 0 tiap baris, harga ≥ 0, total baris &
  grand total konsisten.
- Mencapai **FINAL** mencatat `finalized_at`.
- Isi dokumen hanya dapat diedit saat status DRAFT/REVIEW.
- REJECTED dan CANCELLED bersifat final (tidak ada transisi keluar).
- Hanya DRAFT yang dapat dihapus (soft delete).

## 6. Duplikat vs Revisi

| Aksi | Nomor | Revision | Status awal | Kapan dipakai |
|------|-------|----------|-------------|---------------|
| **Duplikat** (BR-09) | Baru (periode hari ini) | 0 | Draft | Menawarkan pekerjaan serupa ke customer/proyek lain |
| **Revisi** (BR-10) | Sama dengan asal | +1 | Draft | Koreksi setelah dokumen Final/Terkirim/Ditolak |

Keduanya menyalin penuh header + item + sub item. Revisi hanya dibuat dari
dokumen berstatus FINAL/SENT/REJECTED, dan mencatat baris histori di
`sph_revisions` (asal revisi + catatan).

## 7. Terbilang (BR-11)

Fungsi `Terbilang` di backend (`internal/services/terbilang.go`) adalah
sumber tunggal: gaya kapital tiap kata + akhiran "Rupiah", mendukung hingga
triliun, "Seribu" khusus, dan angka negatif berawalan "Minus". Dipakai di
daftar, detail, dashboard, dan preview wizard.

## 8. Dashboard (FR-U5)

`DashboardStats`: Total SPH, jumlah draft (DRAFT+REVIEW), final (FINAL+SENT),
disetujui (ACCEPTED), nilai grand total bulan berjalan, dan 5 dokumen terbaru.

## 9. Batasan Phase 5

- Mode harga **PEMBOBOTAN** sengaja ditolak backend sampai Phase 6.
- Material master (FR-M7) belum tersedia; harga material diisi manual.
- Export PDF/print menyusul di fase export.
