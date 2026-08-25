package services

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotFound dipakai saat data yang diminta tidak ada (sudah dihapus / ID salah).
var ErrNotFound = errors.New("data tidak ditemukan")

// ValidationError membawa pesan siap-tampil ke pengguna (BR-15).
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string { return e.msg }

func NewValidationError(format string, args ...interface{}) *ValidationError {
	return &ValidationError{msg: fmt.Sprintf(format, args...)}
}

// ConflictError untuk pelanggaran aturan data (kode duplikat, hapus yang masih dipakai, dst).
type ConflictError struct {
	msg string
}

func (e *ConflictError) Error() string { return e.msg }

func NewConflictError(format string, args ...interface{}) *ConflictError {
	return &ConflictError{msg: fmt.Sprintf(format, args...)}
}

// trim memotong spasi di awal/akhir teks input.
func trim(s string) string { return strings.TrimSpace(s) }

// sameIDs membandingkan dua daftar ID berisi elemen sama persis (urutan diabaikan).
func sameIDs(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[uint]int, len(a))
	for _, id := range a {
		set[id]++
	}
	for _, id := range b {
		set[id]--
		if set[id] < 0 {
			return false
		}
	}
	return true
}

// uniqueIDs menghilangkan duplikat dan ID nol; error bila hasilnya kosong.
func uniqueIDs(ids []uint) ([]uint, error) {
	seen := make(map[uint]bool, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, errors.New("daftar id kosong")
	}
	return out, nil
}

// IsFriendly memeriksa apakah error adalah tipe yang layak tampil ke pengguna.
func IsFriendly(err error) bool {
	if err == nil {
		return false
	}
	var ve *ValidationError
	var ce *ConflictError
	return errors.As(err, &ve) || errors.As(err, &ce)
}
