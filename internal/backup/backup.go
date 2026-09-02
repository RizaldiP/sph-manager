// Package backup menangani pembuatan, daftar, validasi, dan penggantian file
// database backup (.db) untuk fitur Backup & Restore (Phase 11 / FR-B1..B3).
package backup

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/database"
)

// Prefix nama file backup sesuai FR-B1: SPH_Backup_YYYY-MM-DD_HHMMSS.db
const filePrefix = "SPH_Backup_"

// nameRe mengizinkan akhiran _N (mis. SPH_Backup_2026-09-02_150405_1.db)
// agar dua import di detik yang sama tidak bertabrakan.
var nameRe = regexp.MustCompile(`^SPH_Backup_\d{4}-\d{2}-\d{2}_\d{6}(?:_\d+)?\.db$`)

// requiredTables dipakai saat validasi: backup harus berisi tabel-tabel inti SPH.
var requiredTables = []string{
	"categories", "work_items", "work_sub_items",
	"templates", "template_items",
	"customers", "vessels", "materials",
	"sph_documents", "sph_items", "sph_sub_items",
	"audit_logs", "settings",
}

// Info meta sebuah berkas backup.
type Info struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// Name menyusun nama file backup untuk waktu t (FR-B1).
func Name(t time.Time) string {
	return filePrefix + t.Format("2006-01-02_150405") + ".db"
}

// IsValidName memastikan nama hanya file backup sah (cegah traversal/path jelek).
func IsValidName(name string) bool {
	return nameRe.MatchString(name)
}

// Create membuat snapshot konsisten database ke dir lewat VACUUM INTO;
// mengembalikan path file backup yang terbentuk.
func Create(db *gorm.DB, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat folder backup: %w", err)
	}
	dest := filepath.Join(dir, Name(time.Now()))
	if err := db.Exec("VACUUM INTO '" + strings.ReplaceAll(dest, "'", "''") + "'").Error; err != nil {
		return "", fmt.Errorf("gagal membuat backup: %w", err)
	}
	if _, err := os.Stat(dest); err != nil {
		return "", fmt.Errorf("file backup tidak ditemukan setelah dibuat: %w", err)
	}
	return dest, nil
}

// List meratakan isi folder backup terurut dari yang terbaru (FR-B1 daftar).
func List(dir string) ([]Info, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]Info, 0, len(entries))
	for _, e := range entries {
		if !IsValidName(e.Name()) {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		p := filepath.Join(dir, e.Name())
		out = append(out, Info{
			Name:     e.Name(),
			Path:     p,
			Size:     fi.Size(),
			Modified: fi.ModTime().Format("2006-01-02 15:04"),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

// Retain menghapus backup tertua bila jumlah melewati keep (FR-B3 retention).
func Retain(dir string, keep int) ([]string, error) {
	if keep < 1 {
		keep = 1
	}
	items, err := List(dir)
	if err != nil {
		return nil, err
	}
	var removed []string
	for i := keep; i < len(items); i++ {
		if err := os.Remove(items[i].Path); err != nil {
			return removed, err
		}
		removed = append(removed, items[i].Name)
	}
	return removed, nil
}

// Validate membuka path sebagai database, menjalankan integrity_check dan
// foreign_key_check, lalu memastikan tabel inti ada; mengembalikan daftar tabel.
func Validate(path string, lg *slog.Logger) ([]string, error) {
	db, err := database.Open(path, lg)
	if err != nil {
		return nil, fmt.Errorf("bukan file backup yang valid")
	}
	defer func() {
		_ = database.Close(db)
		_ = os.Remove(path + "-wal")
		_ = os.Remove(path + "-shm")
	}()

	var integrity string
	if err := db.Raw(`PRAGMA integrity_check`).Scan(&integrity).Error; err != nil {
		return nil, fmt.Errorf("file backup rusak: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(integrity), "ok") {
		return nil, fmt.Errorf("file backup rusak (integrity_check = %q)", integrity)
	}
	var fk []map[string]interface{}
	_ = db.Raw(`PRAGMA foreign_key_check`).Find(&fk).Error
	if len(fk) > 0 {
		return nil, fmt.Errorf("file backup mengandung pelanggaran foreign key (%d baris)", len(fk))
	}
	tables, err := database.ExistingTables(db)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca struktur backup: %w", err)
	}
	have := map[string]bool{}
	for _, t := range tables {
		have[t] = true
	}
	for _, need := range requiredTables {
		if !have[need] {
			return nil, fmt.Errorf("file backup bukan database SPH (tabel %q tidak ditemukan)", need)
		}
	}
	return tables, nil
}

// SwapMain menggantikan dbFile dengan backupPath (asumsi handle DB sudah
// ditutup): salin ke file sementara lalu pindah atomik, bersihkan sisa WAL.
func SwapMain(dbFile, backupPath string) error {
	dir := filepath.Dir(dbFile)
	tmp := filepath.Join(dir, ".sph_restore_tmp.db")
	if err := copyFile(backupPath, tmp); err != nil {
		return fmt.Errorf("gagal menyalin backup: %w", err)
	}
	for _, sidecar := range []string{dbFile + "-wal", dbFile + "-shm"} {
		if _, err := os.Stat(sidecar); err == nil {
			_ = os.Remove(sidecar)
		}
	}
	if err := os.Rename(tmp, dbFile); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("gagal mengganti file database: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// nextFreeName memilih nama backup di dir yang belum dipakai untuk waktu t;
// bila nama dasar sudah ada, diberi akhiran _1, _2, dst.
func nextFreeName(dir string, t time.Time) string {
	base := Name(t)
	dest := filepath.Join(dir, base)
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return dest
	}
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for n := 1; ; n++ {
		cand := filepath.Join(dir, fmt.Sprintf("%s_%d%s", stem, n, ext))
		if _, err := os.Stat(cand); os.IsNotExist(err) {
			return cand
		}
	}
}

// CopyInto menyalin file backup dari lokasi mana pun (USB/laptop lain) ke dir
// dengan nama timestamp; mengembalikan Info backup yang tersalin.
func CopyInto(src, dir string) (Info, error) {
	fi, err := os.Stat(src)
	if err != nil {
		return Info{}, fmt.Errorf("file backup tidak dapat dibaca: %w", err)
	}
	if fi.IsDir() {
		return Info{}, fmt.Errorf("yang dipilih adalah folder, bukan file backup")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Info{}, fmt.Errorf("gagal menyiapkan folder backup: %w", err)
	}
	dest := nextFreeName(dir, time.Now())
	if err := copyFile(src, dest); err != nil {
		return Info{}, fmt.Errorf("gagal menyalin backup: %w", err)
	}
	return Info{
		Name:     filepath.Base(dest),
		Path:     dest,
		Size:     fi.Size(),
		Modified: fi.ModTime().Format("2006-01-02 15:04"),
	}, nil
}