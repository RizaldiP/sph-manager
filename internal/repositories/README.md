# repositories

Layer akses data (GORM). Semua method menerima `*gorm.DB` sebagai parameter pertama agar service dapat
mengatur transaksi (`db.Transaction`) secara penuh — sesuai BR-16. Repository tidak memuat business rule.

Konvensi:
- Query default selalu mengecualikan baris soft-delete (GORM `DeletedAt` otomatis).
- Pemeriksaan unik kode bersifat *soft-delete aware* (sejalan dengan partial unique index `uq_*_live`).
- Reorder menerima daftar ID berurutan; urutan disimpan mulai dari 1.
