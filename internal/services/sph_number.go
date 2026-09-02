package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// Konfigurasi penomoran SPH (BR-07). Placeholder yang didukung:
//
//	{YYYY} tahun 4 digit, {MM} bulan 2 digit, {ROMAN} bulan Romawi,
//	{SEQ} nomor urut 3 digit (diisi manual oleh pengguna saat membuat SPH).
const defaultSphNumberFormat = "{SEQ}/SPH-GEI/{ROMAN}/{YYYY}"

// oldDefaultSphNumberFormat adalah format bawaan versi lama (SEQ di akhir).
// Bila nilai tersimpan masih memakai ini, diperlakukan sebagai kosong agar
// instalasi lama ikut format baru tanpa perlu diubah manual di Pengaturan.
const oldDefaultSphNumberFormat = "SPH/GEI/{ROMAN}/{YYYY}/{SEQ}"

var romanMonths = []string{
	"I", "II", "III", "IV", "V", "VI",
	"VII", "VIII", "IX", "X", "XI", "XII",
}

func settingSphNumberFormat(db *gorm.DB) (string, error) {
	v, err := repositories.SettingValue(db, "sph_number_format")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(v) == "" || strings.TrimSpace(v) == oldDefaultSphNumberFormat {
		return defaultSphNumberFormat, nil
	}
	return strings.TrimSpace(v), nil
}

// renderStatic menggantikan placeholder bertanggal {YYYY}/{MM}/{ROMAN}.
func renderStatic(format string, t time.Time) string {
	out := strings.ReplaceAll(format, "{YYYY}", strconv.Itoa(t.Year()))
	out = strings.ReplaceAll(out, "{MM}", fmt.Sprintf("%02d", int(t.Month())))
	return strings.ReplaceAll(out, "{ROMAN}", romanMonths[int(t.Month())-1])
}

// padSeq memastikan nomor urut setidaknya 3 digit (001, 012, 1000 tetap utuh).
func padSeq(seq string) string {
	seq = strings.TrimSpace(seq)
	if len(seq) >= 3 {
		return seq
	}
	return strings.Repeat("0", 3-len(seq)) + seq
}

// composeNumber merender format dengan {SEQ} di posisi apa pun.
// seq dinormalisasi (trim + pad 3 digit) sebelum disusun.
func composeNumber(format, seq string, t time.Time) (string, error) {
	if !strings.Contains(format, "{SEQ}") {
		return "", NewValidationError("Format nomor SPH harus memuat placeholder {SEQ}.")
	}
	seq = padSeq(seq)
	if !regexp.MustCompile(`^\d{1,9}$`).MatchString(seq) {
		return "", NewValidationError("Nomor urut SPH harus berupa angka 1–9 digit.")
	}
	return strings.ReplaceAll(renderStatic(format, t), "{SEQ}", seq), nil
}

// splitPrefixSuffix membagi format menjadi bagian kiri/kanan {SEQ} setelah
// placeholder tanggal dirender.
func splitPrefixSuffix(format string, t time.Time) (prefix, suffix string) {
	rendered := renderStatic(format, t)
	idx := strings.Index(rendered, "{SEQ}")
	if idx < 0 {
		return rendered, ""
	}
	return rendered[:idx], rendered[idx+len("{SEQ}"):]
}

// maxSequenceForFormat mencari nomor urut tertinggi dari dokumen tersimpan
// yang cocok dengan format & periode (tanggal). Bekerja walau {SEQ} berada
// di depan, tengah, maupun akhir format.
func maxSequenceForFormat(db *gorm.DB, format string, t time.Time) (int, error) {
	prefix, suffix := splitPrefixSuffix(format, t)
	if !strings.Contains(format, "{SEQ}") {
		return 0, nil
	}
	return repositories.NewSphRepository().MaxSequenceInNumber(db, prefix, suffix)
}

// generateDocumentNumber menyusun nomor dokumen baru berurutan otomatis
// (dipakai Duplicate): seq = urut tertinggi periode +1. Dipanggil dalam
// transaksi agar bebas tabrakan (BR-07, BR-16).
func generateDocumentNumber(tx *gorm.DB, t time.Time) (string, error) {
	format, err := settingSphNumberFormat(tx)
	if err != nil {
		return "", err
	}
	maxSeq, err := maxSequenceForFormat(tx, format, t)
	if err != nil {
		return "", err
	}
	return composeAndCheck(tx, format, fmt.Sprintf("%03d", maxSeq+1), t)
}

// manualDocumentNumber menyusun nomor dari nomor urut yang diinput pengguna
// (dipakai Create). Nomor urut wajib angka; nomor yang sudah dipakai ditolak.
func manualDocumentNumber(tx *gorm.DB, rawSeq string, t time.Time) (string, error) {
	format, err := settingSphNumberFormat(tx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(rawSeq) == "" {
		return "", NewValidationError("Nomor urut SPH wajib diisi.")
	}
	return composeAndCheck(tx, format, rawSeq, t)
}

// composeAndCheck menyusun nomor dan menolak bila sudah dipakai baris hidup.
func composeAndCheck(tx *gorm.DB, format, seq string, t time.Time) (string, error) {
	number, err := composeNumber(format, seq, t)
	if err != nil {
		return "", err
	}
	exists, err := repositories.NewSphRepository().NumberExists(tx, number, 0)
	if err != nil {
		return "", err
	}
	if exists {
		return "", NewValidationError("Nomor dokumen %s sudah dipakai. Gunakan nomor urut lain.", number)
	}
	return number, nil
}