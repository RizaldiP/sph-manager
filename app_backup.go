package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/RizaldiP/sph-manager/internal/backup"
	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RestoreResult hasil alur restore (FR-B2).
type RestoreResult struct {
	Restarting bool   `json:"restarting"`
	Backup     string `json:"backup"`
	Message    string `json:"message"`
}

// BackupNow membuat backup database manual (FR-B1).
func (a *App) BackupNow() (*services.BackupInfo, error) {
	return a.backup.CreateManual()
}

// ListBackups melempar daftar backup tersimpan (terbaru dahulu).
func (a *App) ListBackups() ([]services.BackupInfo, error) {
	return a.backup.ListBackups()
}

// DeleteBackup menghapus satu file backup (permanen).
func (a *App) DeleteBackup(name string) error {
	if err := a.backup.DeleteBackup(name); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return services.NewValidationError("Backup tidak ditemukan.")
		}
		a.log.Error("gagal menghapus backup", "error", err)
		return fmt.Errorf("gagal menghapus backup")
	}
	return nil
}

// ImportBackup membuka dialog pilih file .db dari lokasi luar, memvalidasi,
// lalu menyalinnya ke folder backup. Batal memilih = tidak ada perubahan.
func (a *App) ImportBackup() (*services.BackupInfo, error) {
	src, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Backup Database",
		Filters: []runtime.FileFilter{
			{DisplayName: "Backup Database (*.db)", Pattern: "*.db"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dialog pemilih file gagal dibuka")
	}
	if strings.TrimSpace(src) == "" {
		return nil, nil
	}
	return a.backup.Import(src)
}

// OpenBackupFolder membuka folder backup di file manager.
func (a *App) OpenBackupFolder() error {
	return openPath(a.ctx, a.cfg.BackupDir)
}

// QuitApp menutup aplikasi (dipanggil setelah RestoreBackup berhasil).
func (a *App) QuitApp() {
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// RestoreBackup (FR-B2): validasi → safety backup → tutup DB → ganti file →
// catat audit → restart ulang aplikasi secara otomatis.
func (a *App) RestoreBackup(name string) (*RestoreResult, error) {
	a.restoreMu.Lock()
	defer a.restoreMu.Unlock()

	if room := a.collabMgr.CurrentRoomID(); room != "" {
		return nil, services.NewValidationError("Tutup sesi Work Together terlebih dahulu sebelum restore.")
	}
	if _, err := a.backup.Validate(name); err != nil {
		return nil, err
	}
	absPath, err := a.backup.Resolve(name)
	if err != nil {
		return nil, services.ErrNotFound
	}

	a.backup.StopAuto()
	safety, err := a.backup.CreateManual()
	if err != nil {
		a.backup.StartAuto()
		return nil, err
	}
	a.log.Info("restore dimulai", "backup", absPath, "safety", safety.Path)

	if err := database.Close(a.db); err != nil {
		a.backup.StartAuto()
		a.log.Error("gagal menutup database aktif", "error", err)
		return nil, fmt.Errorf("gagal menutup database aktif")
	}
	a.db = nil
	a.backup.MarkClosed()
	a.collabMgr.Shutdown("Restore database.")

	if err := backup.SwapMain(a.cfg.DatabasePath, absPath); err != nil {
		a.log.Error("restore gagal mengganti file database", "error", err)
		_ = backup.SwapMain(a.cfg.DatabasePath, safety.Path)
		a.quitAfterRestoreError()
		return nil, fmt.Errorf("restore dibatalkan. Aplikasi akan ditutup dan data kembali seperti semula")
	}

	writeRestoreAudit(a.cfg.DatabasePath, filepath.Base(absPath), a.log)

	exe, err := os.Executable()
	if err != nil {
		a.log.Error("tidak dapat menemukan executable aplikasi", "error", err)
		a.quitAfterRestoreError()
		return nil, fmt.Errorf("restore selesai namun gagal memulai ulang. Silakan buka kembali aplikasi")
	}
	cmd := exec.Command(exe)
	if err := cmd.Start(); err != nil {
		a.log.Error("gagal memulai ulang aplikasi", "error", err)
		a.quitAfterRestoreError()
		return nil, fmt.Errorf("restore selesai namun gagal memulai ulang. Silakan buka kembali aplikasi")
	}

	return &RestoreResult{
		Restarting: true,
		Backup:     filepath.Base(absPath),
		Message:    "Restore berhasil. Aplikasi akan ditutup dan dibuka kembali.",
	}, nil
}

// writeRestoreAudit mencatat aksi RESTORE di database hasil restore (FR-A1).
func writeRestoreAudit(dbPath, backupName string, lg *slog.Logger) {
	db, err := database.Open(dbPath, lg)
	if err != nil {
		lg.Warn("gagal membuka hasil restore untuk audit", "error", err)
		return
	}
	defer func() { _ = database.Close(db) }()
	desc := "Database dipulihkan dari backup (" + backupName + ")"
	if err := db.Create(&models.AuditLog{Action: "RESTORE", Entity: "database", Description: desc}).Error; err != nil {
		lg.Warn("gagal mencatat audit RESTORE", "error", err)
	}
}

// quitAfterRestoreError menutup aplikasi setelah restore gagal pertengahan
// karena service masih memegang pegangan database yang sudah diganti.
func (a *App) quitAfterRestoreError() {
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}