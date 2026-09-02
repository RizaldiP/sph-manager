package services

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/backup"
)

// BackupRetentionKeep adalah jumlah backup yang dipertahankan (FR-B3).
const BackupRetentionKeep = 10

// BackupInfo meta sebuah backup untuk ditampilkan di UI.
type BackupInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// BackupService menangani backup & restore database (Phase 11 / FR-B1..B3).
type BackupService struct {
	db    *gorm.DB
	dir   string
	log   *slog.Logger
	audit *AuditWriter

	mu     sync.Mutex
	closed atomic.Bool
	stopCh chan struct{}
	once   sync.Once
}

func NewBackupService(db *gorm.DB, dir string, log *slog.Logger) *BackupService {
	return &BackupService{db: db, dir: dir, log: log, audit: NewAuditWriter(), stopCh: make(chan struct{})}
}

// dbReady memastikan database masih terbuka sebelum operasi apa pun.
func (s *BackupService) dbReady() bool {
	if s.closed.Load() || s.db == nil {
		return false
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

func toBackupInfo(it backup.Info) BackupInfo {
	return BackupInfo{Name: it.Name, Path: it.Path, Size: it.Size, Modified: it.Modified}
}

// Resolve memvalidasi nama lalu mengembalikan path penuh di dalam folder backup.
func (s *BackupService) Resolve(name string) (string, error) {
	if !backup.IsValidName(name) {
		return "", ErrNotFound
	}
	absDir, _ := filepath.Abs(s.dir)
	absPath, _ := filepath.Abs(filepath.Join(s.dir, name))
	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
		return "", ErrNotFound
	}
	if _, err := os.Stat(absPath); err != nil {
		return "", ErrNotFound
	}
	return absPath, nil
}

// HasFileToday memeriksa apakah backup auto hari ini sudah ada (FR-B3 harian).
func (s *BackupService) HasFileToday() bool {
	today := time.Now().Format("2006-01-02")
	items, err := backup.List(s.dir)
	if err != nil {
		return false
	}
	for _, it := range items {
		if strings.HasPrefix(it.Name, "SPH_Backup_"+today) {
			return true
		}
	}
	return false
}

func (s *BackupService) createLocked(auto bool) (*BackupInfo, error) {
	if !s.dbReady() {
		return nil, NewValidationError("Database sedang ditutup, backup tidak dapat dibuat.")
	}
	path, err := backup.Create(s.db, s.dir)
	if err != nil {
		s.log.Error("gagal membuat backup", "error", err)
		return nil, fmt.Errorf("gagal membuat backup")
	}
	if _, err := backup.Retain(s.dir, BackupRetentionKeep); err != nil {
		s.log.Warn("gagal memberlakukan retention backup", "error", err)
	}
	desc := "Backup database dibuat manual"
	if auto {
		desc = "Backup database dibuat otomatis"
	}
	if err := s.audit.Write(s.db, "BACKUP", "database", 0, desc+" ("+filepath.Base(path)+")"); err != nil {
		s.log.Warn("gagal mencatat audit backup", "error", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		return &BackupInfo{Name: filepath.Base(path), Path: path}, nil
	}
	return &BackupInfo{
		Name:     filepath.Base(path),
		Path:     path,
		Size:     fi.Size(),
		Modified: fi.ModTime().Format("2006-01-02 15:04"),
	}, nil
}

// CreateManual membuat backup sekarang juga (FR-B1).
func (s *BackupService) CreateManual() (*BackupInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createLocked(false)
}

// EnsureDaily membuat backup bila hari ini belum ada (FR-B3); mengembalikan
// info backup yang ada/hari ini.
func (s *BackupService) EnsureDaily() (*BackupInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.HasFileToday() {
		items, _ := backup.List(s.dir)
		if len(items) > 0 {
			info := toBackupInfo(items[0])
			return &info, nil
		}
		return nil, nil
	}
	return s.createLocked(true)
}

// ListBackups meratakan backup tersimpan terurut terbaru dahulu.
func (s *BackupService) ListBackups() ([]BackupInfo, error) {
	items, err := backup.List(s.dir)
	if err != nil {
		s.log.Error("gagal membaca daftar backup", "error", err)
		return nil, fmt.Errorf("gagal membaca daftar backup")
	}
	out := make([]BackupInfo, 0, len(items))
	for _, it := range items {
		out = append(out, toBackupInfo(it))
	}
	return out, nil
}

// Delete meremukkan paksa satu backup (bersifat permanen).
func (s *BackupService) DeleteBackup(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path, err := s.Resolve(name)
	if err != nil {
		return ErrNotFound
	}
	return os.Remove(path)
}

// Validate memeriksa backup tanpa mengubah database aktif.
func (s *BackupService) Validate(name string) ([]string, error) {
	path, err := s.Resolve(name)
	if err != nil {
		return nil, ErrNotFound
	}
	tables, err := backup.Validate(path, s.log)
	if err != nil {
		return nil, NewValidationError("%s", err.Error())
	}
	return tables, nil
}

// Import menyalin backup dari lokasi luar (USB/laptop lain) ke folder backup
// setelah divalidasi; sumber yang sudah berada di folder backup tidak diduplikasi.
func (s *BackupService) Import(srcPath string) (*BackupInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fi, err := os.Stat(srcPath)
	if err != nil || fi.IsDir() {
		return nil, NewValidationError("File backup tidak dapat dibaca.")
	}
	if strings.ToLower(filepath.Ext(srcPath)) != ".db" {
		return nil, NewValidationError("File backup harus berformat .db.")
	}
	if fi.Size() > 2*1024*1024*1024 {
		return nil, NewValidationError("Ukuran file backup terlalu besar.")
	}

	absSrc, _ := filepath.Abs(srcPath)
	absDir, _ := filepath.Abs(s.dir)
	base := filepath.Base(srcPath)
	if filepath.Dir(absSrc) == absDir && backup.IsValidName(base) {
		return &BackupInfo{Name: base, Path: srcPath, Size: fi.Size(), Modified: fi.ModTime().Format("2006-01-02 15:04")}, nil
	}

	if _, verr := backup.Validate(srcPath, s.log); verr != nil {
		return nil, NewValidationError("%s", verr.Error())
	}
	it, err := backup.CopyInto(srcPath, s.dir)
	if err != nil {
		s.log.Error("gagal mengimpor backup", "sumber", srcPath, "error", err)
		return nil, fmt.Errorf("gagal mengimpor backup")
	}
	info := toBackupInfo(it)
	if s.dbReady() {
		if err := s.audit.Write(s.db, "IMPORT", "database", 0, "Backup diimpor dari file eksternal ("+base+")"); err != nil {
			s.log.Warn("gagal mencatat audit import", "error", err)
		}
	}
	return &info, nil
}

// StartAuto menjalankan penjadwal backup harian (FR-B3).
func (s *BackupService) StartAuto() {
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if _, err := s.EnsureDaily(); err != nil {
					s.log.Warn("backup harian gagal", "error", err)
				}
			case <-s.stopCh:
				return
			}
		}
	}()
}

// StopAuto menghentikan penjadwal harian (dipanggil saat tutup / restore).
func (s *BackupService) StopAuto() {
	s.once.Do(func() { close(s.stopCh) })
}

// BackupOnShutdown membuat backup otomatis saat aplikasi ditutup (FR-B3)
// bila belum ada backup hari ini; database aktif harus sudah tertutup.
func (s *BackupService) BackupOnShutdown() {
	s.StopAuto()
	if !s.dbReady() {
		return
	}
	if _, err := s.EnsureDaily(); err != nil {
		s.log.Warn("backup saat penutupan gagal", "error", err)
	}
}

// MarkClosed menandai database aktif tertutup (dipakai saat restore).
func (s *BackupService) MarkClosed() { s.closed.Store(true) }