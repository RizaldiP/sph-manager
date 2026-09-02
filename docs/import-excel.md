# Import Excel — Master Pekerjaan (FR-IE1..IE4)

Fitur import daftar pekerjaan dari file Excel ke Master Pekerjaan. Implementasi mengikuti analisis struktur file referensi (`docs/excel-analysis.md`).

## Alur Wajib (FR-IE1)

Import **wajib** melewati wizard 4 langkah di halaman `/import` (`ImportPage.vue`); tidak ada jalur import tanpa pratinjau:

1. **Pilih File** — dialog hanya menerima `.xls` dan `.xlsx`.
2. **Sheet & Mapping** — pilih sheet, kategori tujuan, dan pemetaan kolom; grid pratinjau menampilkan isi sheet.
3. **Pratinjau & Klasifikasi** — seluruh baris hasil parse ditampilkan; pengguna memutuskan level tiap baris yang tidak pasti.
4. **Selesai** — ringkasan hasil (pekerjaan / sub-pekerjaan dibuat, dilewati).

Data baru **ditambahkan** ke kategori tujuan sebagai entri baru (FR-IE2) — tidak pernah menimpa data existing. Sequence melanjutkan urutan terakhir kategori; kode `PEK-`/`SUB-` digenerate otomatis.

## Format & Reader

| Ekstensi | Library | Catatan |
| --- | --- | --- |
| `.xls` (BIFF lama) | `github.com/shakinm/xlsReader` | Format file referensi SPH_KRI_OWA.xls |
| `.xlsx` | `github.com/xuri/excelize/v2` | Dipakai juga generator export Phase 9 |

Grid dinormalkan menjadi persegi (baris dipad sampai lebar maksimum) dan sel di-trim. Angka dari Excel dibaca sebagai teks bersih (tanpa notasi ilmiah); parser angka menerima gaya Indonesia/internasional: `120725000`, `1.234.567`, `1.234.567,89`, `1,234,567.89`.

## Deteksi Hierarki (FR-IE3)

Dua level saja sesuai model Master Pekerjaan (induk/sub); sub-sub diratakan menjadi sub. Penanda dikenali dari awal teks uraian **atau satu kolom di kirinya** (huruf grup/romawi sering berdiri sendiri):

| Bentuk | Contoh | Klasifikasi |
| --- | --- | --- |
| Romawi uppercase | `I`, `XII.` , `VII Laksanakan…` | Induk |
| Huruf besar tunggal | `A`, `B.` | Induk (grup) |
| Huruf kecil | `a.`, `b)` | Sub |
| Angka | `1.`, `12)` | Lihat aturan deret |
| Tanpa penanda | — | `unknown` → wajib diputuskan pengguna |

Aturan deret untuk angka ambigu: angka yang **meneruskan deret induk terakhir** (`n == lastMain+1`) adalah induk baru (`2.` setelah blok `1./a./b.` = saudara `1.`); angka lain di dalam blok sub (mis. sub-sub `1. Sensor Oli`) diratakan menjadi sub. Baris tanpa penanda **tidak pernah ditebak** — UI menandainya merah dan import diblokir sampai pengguna memilih Induk/Sub/Lewati.

Penanda angka wajib delimiter `.` atau `)` agar kolom nomor urut polos (`1`, `2`, `3`) tidak salah potong nama pekerjaan.

## Mapping Kolom (FR-IE2)

Kolom `-1` berarti tidak dipetakan. Opsi lanjutan:

- **Gabung Kolom Uraian (nameSpan)** — beberapa kolom digabung menjadi satu teks uraian; layout referensi menaruh teks level berbeda pada kolom indentasi berbeda (kolom kosong dilewati saat penggabungan). Satu kolom kiri uraian selalu ikut digabung.
- **Baris Data Pertama (headerRows)** — indeks baris data pertama; semua baris sebelumnya dianggap header.
- **Mode Nilai Total (serviceTotal/materialTotal)** — pada file referensi, kolom JASA/MATERIAL memuat nilai TOTAL baris (= harga satuan × qty). Bila aktif, harga satuan dihitung `round(total ÷ qty)`; bila nonaktif sel dipakai apa adanya sebagai harga satuan.
- **Harga Satuan (umum) + dihitung sebagai** — layout referensi mencampur konvensi: banyak baris menaruh harga langsung di kolom HARGA SATUAN dengan JASA/MAT kosong. Kolom ini menjadi *fallback*: hanya dipakai bila hasil JASA dan MATERIAL baris tersebut keduanya kosong/nol, dan pengguna memilih apakah nilainya masuk ke Harga Jasa atau Harga Material (default Jasa). Nilai total jasa/material yang terisi tetap menang.

Saran mapping otomatis (`SuggestMapping`) memindai maksimal 15 baris pertama, membentuk *band* baris header bersebelahan yang sama-sama mengandung kata kunci (URAIAN/KEGIATAN → nama, JML/QTY/JUMLAH → qty, SAT/SATUAN/UNIT → satuan, JASA → jasa, MAT/MATERIAL → material), lalu memutuskan jenis tiap kolom dari teks gabungan band-nya — header dua baris `HARGA`/`SATUAN` dikenali sebagai kolom Harga Satuan tanpa menelan kolom SAT. Kemunculan pertama tiap jenis dalam band yang dipakai; baris indeks kolom (`1 2 3 …`) sesudah band dilewati otomatis.

## Validasi Blokir

Import ditolak (conflict error) bila masih ada:

- baris `unknown` yang belum diklasifikasi,
- baris terkonfirmasi dengan error (qty ≤ 0 / bukan angka, harga negatif, nama kosong pada level Induk),
- sub tanpa pekerjaan induk di atasnya.

Sel kosong qty **bukan** error (umum pada baris grup) — default 1.

## Transaksi, Progress, Audit

- Seluruh penulisan berjalan dalam **satu transaksi** (BR-16): kegagalan apa pun (termasuk pembatalan via callback progress) melakukan rollback penuh — diverifikasi test.
- Progress dipancarkan per baris melalui event Wails `import:progress` `{done,total}`; UI menampilkan progress bar; `import:done` menandakan selesai.
- Setiap pekerjaan/sub yang dibuat tercatat di audit log (BR-13) dalam transaksi yang sama.

## API Binding

```
PickImportFile() string                                  // dialog *.xls;*.xlsx
ListImportSheets(path) []string
PreviewImportSheet(path, sheet) SheetPreview             // grid ≤200×20 + saran mapping
ParseImportRows(path, sheet, ColumnMapping) []PreviewRow // pratinjau penuh
ValidateImportRows(path, sheet, ColumnMapping, []ConfirmRow) []string
RunWorkItemImport(categoryID, path, sheet, ColumnMapping, []ConfirmRow) ImportResult
```

`SheetPreview.Grid` dibatasi 200 baris × 20 kolom untuk render; parsing `ParseRows` tetap memproses seluruh baris.

## Test

`go test ./internal/...` mencakup:

- `internal/importers/parser_test.go` — splitMarker (angka/romawi/huruf besar-kecil/tanpa penanda), hierarki Gaya A (grup A → angka → huruf kecil → sub-sub diratakan) dan Gaya B (romawi → huruf kecil), konversi total→satuan, fallback Harga Satuan (arah jasa/material, prioritas total, nilai invalid), angka Indonesia, saran mapping (termasuk band header HARGA/SATUAN dua baris).
- `internal/importers/reader_test.go` — ListSheets/ReadSheet XLSX (fixture dibuat via excelize saat runtime), format tak didukung ditolak, dan **test integrasi opsional** terhadap `SPH_KRI_OWA.xls` asli di root repo (di-skip bila file tidak ada): memverifikasi mapping lengkap termasuk `UnitPriceCol=7` dan harga baris "Sensor Oli" = 2.200.000 masuk hasil analisis.
- `internal/services/import_service_test.go` — happy path (kode/sequence/kategori/konversi harga/audit), rollback via progress-cancel, blokir unknown & qty nol, sub tanpa induk, kategori tidak ada.
