# Issues, Ambiguity & Assumptions

> Semua requirement yang ambigu didokumentasikan di sini dengan format: **Question / Assumption / Reason / Impact**. Prioritas sumber kebenaran (Master Specification §76): 1) Excel referensi, 2) Master Specification, 3) Existing project, 4) Business best practice, 5) Assumption.

---

## ISS-01 — Pembobotan tidak ada di Excel referensi

- **Question:** Bagaimana format pembobotan yang benar?
- **Assumption:** Ikuti Master Specification §22: sub point value = main point value × weight/100, Σ weight = 100%.
- **Reason:** Excel referensi tidak memuat kolom weight sama sekali.
- **Impact:** UI perlu mode per main point (harga langsung vs pembobotan); validasi finalisasi mengecek Σ weight.

## ISS-02 — Logo tidak direplikasi dari file referensi

- **Question:** Apakah logo perusahaan diambil dari file Excel?
- **Assumption:** Tidak. Logo di-upload pengguna melalui Settings (path disimpan di settings), sesuai instruksi pengguna ("skip logo").
- **Reason:** Master Specification mewajibkan logo jadi bagian settings, bukan hard-code.
- **Impact:** Generator dokumen membaca logo dari settings; jika kosong, dokumen tanpa logo (tidak error).

## ISS-03 — Penandatangan dokumen

- **Question:** Siapa penandatangan default?
- **Assumption:** Default "Matawi / Direktur" sesuai referensi, **dapat diubah** di Settings (Signer Name + Signer Position).
- **Reason:** Referensi menampilkan ttd "M a t a w i — Direktur"; Master Specification mewajibkan signer configurable.
- **Impact:** Blok ttd generator membaca settings; tanggal ttd = tanggal generate dokumen.

## ISS-04 — Hierarki penomoran tidak seragam

- **Question:** Gaya penomoran mana yang jadi standar aplikasi?
- **Assumption:** Aplikasi mendukung beberapa gaya (angka `1.`, huruf kecil `a.`, huruf besar `A`, Romawi `I.`); gaya tampilan dokumen default mengikuti gaya B (Romawi + huruf) karena lebih ringkas, dan dapat dikonfigurasi.
- **Reason:** Kedua sheet referensi memakai gaya berbeda; importer wajib fleksibel.
- **Impact:** Parser import tidak boleh menebak — klasifikasi ambigu ditampilkan ke pengguna.

## ISS-05 — Penempatan Jasa vs Material tidak konsisten di referensi

- **Question:** Kapan harga masuk kolom JASA vs MAT?
- **Assumption:** Aplikasi menyediakan **dua kolom eksplisit** (service_unit_price & material_unit_price); item boleh mengisi keduanya. Panduan: pekerjaan/jasa → JASA; suku cadang/"ganti baru" → MATERIAL.
- **Reason:** Referensi tidak konsisten; kolom eksplisit menghilangkan ambiguitas.
- **Impact:** Total baris = jasa + material (BR-02).

## ISS-06 — Terbilang manual di referensi

- **Question:** Terbilang diketik manual atau digenerate?
- **Assumption:** Digenerate otomatis (BR-11) dengan unit test.
- **Reason:** Manual rawan salah; aplikasi harus akurat.
- **Impact:** Perlu konverter angka→huruf Indonesia; grand total 0 → "Nol Rupiah".

## ISS-07 — Nomor SPH tidak ada di referensi

- **Question:** Format nomor SPH perusahaan?
- **Assumption:** Format configurable default `SPH/GEI/{ROMAWI_BULAN}/{TAHUN}/{SEQ}` (contoh `SPH/GEI/VIII/2026/001`).
- **Reason:** Master Specification memberi contoh format tersebut; referensi Excel tidak memuat nomor.
- **Impact:** Generator penomoran transaksional + unique constraint (BR-07).

## ISS-08 — Data customer/kapal minim di referensi

- **Question:** Dari mana data customer & kapal?
- **Assumption:** Diinput manual via Master Data; nama kapal bebas teks (contoh referensi: `KRI.SBY-591 TA.2026`).
- **Reason:** Referensi hanya menyebut nama kapal di judul.
- **Impact:** Tidak ada import awal customer/kapal; validasi customer wajib saat finalisasi.

## ISS-09 — Satuan tidak baku

- **Question:** Apakah satuan (unit) memakai daftar tetap?
- **Assumption:** Teks bebas dengan daftar saran (giat, unit, bh, set, roll, dll.) yang tumbuh dari penggunaan.
- **Reason:** Referensi memakai kapitalisasi campur; memaksa daftar tertutup berisiko memblokir data nyata.
- **Impact:** Tidak ada FK ke tabel satuan; hanya validasi non-kosong.

## ISS-10 — Mata uang & presisi

- **Question:** Apakah perlu desimal (sen)?
- **Assumption:** Hanya Rupiah integer, tanpa desimal.
- **Reason:** Semua nilai referensi integer; Master Specification mensyaratkan integer Rupiah aman.
- **Impact:** Kolom harga bertipe integer (basis poin); rounding largest remainder (BR-04).

## ISS-11 — Arah visual UI

- **Question:** Tema visual apa yang dipakai?
- **Assumption:** Simpel, rapi, tertata; warna dasar **biru flat** (primer) + **orange flat** (aksen/status perhatian), tanpa gradasi.
- **Reason:** Instruksi langsung pengguna.
- **Impact:** Design token Tailwind didefinisikan di Phase 1; semua komponen memakai token yang sama.

## ISS-12 — Kolom diskon/PPN

- **Question:** Apakah SPH perlu diskon/PPN?
- **Assumption:** Tidak ada di MVP (referensi tidak memuatnya); arsitektur siap ditambah kemudian.
- **Reason:** Master Specification tidak memintanya di MVP.
- **Impact:** Grand total = Σ baris tanpa pajak; penambahan PPN nanti hanya menyentuh lapisan dokumen.

## ISS-13 — Waktu & zona waktu

- **Question:** Zona waktu untuk tanggal dokumen & log?
- **Assumption:** Waktu lokal komputer pengguna (Asia/Jakarta umumnya); disimpan UTC di database, ditampilkan lokal.
- **Reason:** Praktik aman untuk backup/audit.
- **Impact:** Format tanggal tampil `dd/mm/yyyy` sesuai kebiasaan Indonesia.

## ISS-14 — Ukuran tim pengguna & auth

- **Question:** Perlu login/multi-user?
- **Assumption:** Tidak ada auth di MVP (single user, offline).
- **Reason:** Master Specification: offline, auth online dilarang; multi-user = FUTURE.
- **Impact:** Arsitektur service layer tetap dipisah agar auth mudah ditambah nanti.
