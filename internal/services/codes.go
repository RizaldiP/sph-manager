package services

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// generateCode membuat kode otomatis berurutan untuk sebuah tabel, mis. "TPL-007".
// Nomor diambil dari suffix numerik terbesar yang sudah terpakai di antara baris
// yang belum dihapus, lalu dinaikkan satu angka (padding nol minimal 3 digit).
// Kode isi-an manual dengan prefix sama tapi bukan angka akan diabaikan.
func generateCode(db *gorm.DB, table, prefix string) (string, error) {
	var codes []string
	err := db.Raw(
		"SELECT code FROM "+table+" WHERE code LIKE ? AND deleted_at IS NULL",
		prefix+"%",
	).Scan(&codes).Error
	if err != nil {
		return "", err
	}
	max := 0
	for _, c := range codes {
		n, err := strconv.Atoi(strings.TrimPrefix(c, prefix))
		if err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("%s%03d", prefix, max+1), nil
}
