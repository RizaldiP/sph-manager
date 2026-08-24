# Export Dokumen SPH (Phase 9)

Setiap dokumen SPH yang sudah tersimpan dapat diekspor menjadi **Excel (.xlsx)** atau **PDF** siap kirim ke customer, dari halaman **Detail SPH → tombol Export ▾** (Excel / PDF Landscape / PDF Portrait). Kedua format digambar dari sumber data yang sama sehingga selalu identik.

## Sumber Data

- Snapshot dokumen lengkap via `SphService.Get` — customer, kapal, item + sub point terurut, subtotal & grand total, terbilang otomatis (`BR-01`, `BR-04`).
- Profil perusahaan & penandatangan dari halaman **Pengaturan** (nama, kota, alamat, logo, penandatangan).
- Main point tampil **tebal** dengan nilai roll-up (termasuk seluruh sub point-nya); sub point menjorok di bawahnya dengan rincian nilainya sendiri. Pada mode **PEMBOBOTAN**, sub point diberi keterangan `[bobot N%]`.

## Excel (.xlsx)

Struktur mengikuti layout referensi perusahaan (`docs/excel-analysis.md`):

| Bagian | Isi |
| --- | --- |
| Kop | Logo + nama/kota/alamat perusahaan; nomor dokumen (+ `Rev N`) rata kanan |
| Judul | Subjek dokumen + kapal/tahun, bermarge B:K, huruf kapital |
| Header tabel | Baris 8–9 dua tingkat: NO · URAIAN KEGIATAN · JML · SAT · HARGA SATUAN · PENAWARAN HARGA (Rp): JASA/MAT/JML |
| Data | Mulai baris 10; uraian merge C:E dengan wrap dan tinggi baris menyesuaikan panjang teks |
| Rekap | Sub Total Jasa (kolom JASA), Sub Total Material (MAT), TOTAL bergaya double-rule di kolom JML |
| Penutup | Terbilang italic, catatan dokumen, blok tanda tangan kanan bawah |

Page setup: A4 landscape fit-width 1 halaman lebar, print area eksplisit, **header tabel berulang** tiap halaman cetak melalui print titles (`_xlnm.Print_Titles`), nomor halaman di header cetak (`&P`). Gridlines disembunyikan agar tampil bersih.

Angka memakai format Rupiah `"Rp"#,##0`; sel kosong berarti nol.

## PDF

- **Landscape (default)** untuk tabel lebar; **Portrait** untuk lampiran hemat kertas. Lebar kolom menyesuaikan orientasi.
- Multi-halaman aman: header tabel digambar ulang setiap halaman, dan baris **Pindahan** pertama di tiap halaman lanjutan menampilkan akumulasi jasa/material/jumlah dari halaman-halaman sebelumnya (khas dokumen penawaran cetak).
- Footer setiap halaman: `Halaman X dari Y`.
- Blok rekap, terbilang, catatan, dan tanda tangan otomatis pindah ke halaman baru bila ruang tidak cukup — tidak pernah terpotong.
- PDF Landscape/Portrait dibuat langsung di memori lalu ditulis ke lokasi pilihan pengguna.

## Alur Simpan

1. Klik Export ▾ → pilih format.
2. Dialog **Simpan** terbuka dengan nama default `SPH_<nomor>_rev<N>[_portrait].xlsx/.pdf` dan folder awal `%AppData%\sph-manager\exports`. Ekstensi ditambahkan otomatis bila kurang.
3. Membatalkan dialog tidak melakukan apa-apa.
4. Setelah berhasil muncul banner hijau berisi path hasil + tombol **Buka Folder**.
5. Aktivitas tercatat di audit log sebagai aksi `EXPORT` pada entitas `sph_document` (`BR-13`).

Logo yang rusak/gagal digambar tidak membatalkan export — dokumen tetap dibuat tanpa logo.

## Implementasi

| Berkas | Peran |
| --- | --- |
| `internal/exporters/data.go` | `ExportData` netral + `BuildData` (snapshot → baris cetak) |
| `internal/exporters/excel.go` | Renderer Excel (excelize) |
| `internal/exporters/pdf.go` | Renderer PDF (go-pdf/fpdf) + paginator & Pindahan |
| `internal/services/export_service.go` | Orkestrasi, validasi tujuan, audit |
| `app_export.go` | Binding Wails + dialog simpan + buka folder |

Dokumen DOCX (FR-IE7) sengaja ditunda ke fase FUTURE sesuai requirements.
