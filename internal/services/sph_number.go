package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// Konfigurasi penomoran SPH (BR-07). Placeholder yang didukung:
//
//	{YYYY} tahun 4 digit, {MM} bulan 2 digit, {ROMAN} bulan Romawi,
//	{SEQ} nomor urut 3 digit per periode (tahun+bulan).
const defaultSphNumberFormat = "SPH/GEI/{ROMAN}/{YYYY}/{SEQ}"

var romanMonths = []string{
	"I", "II", "III", "IV", "V", "VI",
	"VII", "VIII", "IX", "X", "XI", "XII",
}

func settingSphNumberFormat(db *gorm.DB) (string, error) {
	v, err := repositories.SettingValue(db, "sph_number_format")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(v) == "" {
		return defaultSphNumberFormat, nil
	}
	return v, nil
}

// buildNumberPrefix menyusun bagian statis nomor dokumen (tanpa {SEQ}).
// Mengembalikan string kosong bila format tidak mengandung {SEQ}.
func buildNumberPrefix(format string, t time.Time) string {
	if !strings.Contains(format, "{SEQ}") {
		return ""
	}
	prefix := strings.ReplaceAll(format, "{YYYY}", strconv.Itoa(t.Year()))
	prefix = strings.ReplaceAll(prefix, "{MM}", fmt.Sprintf("%02d", int(t.Month())))
	prefix = strings.ReplaceAll(prefix, "{ROMAN}", romanMonths[int(t.Month())-1])
	return strings.ReplaceAll(prefix, "{SEQ}", "")
}

// generateDocumentNumber menyusun nomor dokumen baru berurutan per periode.
// Harus dipanggil di dalam transaksi agar bebas tabrakan (BR-07, BR-16).
func generateDocumentNumber(tx *gorm.DB, t time.Time) (string, error) {
	format, err := settingSphNumberFormat(tx)
	if err != nil {
		return "", err
	}
	prefix := buildNumberPrefix(format, t)
	if prefix == "" {
		return "", fmt.Errorf("format penomoran tidak valid: harus memuat {SEQ}")
	}
	maxSeq, err := repositories.NewSphRepository().MaxSequenceInNumber(tx, prefix)
	if err != nil {
		return "", err
	}
	number := prefix + fmt.Sprintf("%03d", maxSeq+1)

	exists, err := repositories.NewSphRepository().NumberExists(tx, number, 0)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("nomor dokumen %s sudah dipakai", number)
	}
	return number, nil
}
