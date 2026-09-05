package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
	"github.com/RizaldiP/sph-manager/internal/sharebackup"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Binding Backup yang Dapat Dibagikan ("Shareable Backup", .sphbak).
// File berisi 6 seksi (SPH, master pekerjaan, kategori, template, customer,
// material); penerima memilih seksi yang ingin dipulihkan.

// ShareBackupCreateResult hasil pembuatan file .sphbak.
type ShareBackupCreateResult struct {
	Path  string `json:"path"`
	Items int    `json:"items"`
}

// ShareBackupPreview pratinjau file .sphbak sebelum dipulihkan.
type ShareBackupPreview struct {
	Path       string                    `json:"path"`
	DeviceName string                    `json:"deviceName"`
	CreatedAt  string                    `json:"createdAt"`
	Counts     sharebackup.SectionCounts `json:"counts"`
}

// CreateShareableBackup: susun paket → dialog simpan → tulis .sphbak.
func (a *App) CreateShareableBackup() (*ShareBackupCreateResult, error) {
	pkg, err := a.share.Build()
	if err != nil {
		return nil, err
	}
	base := "ShareBackup_" + sanitizeFilename(pkg.Metadata.DeviceName) +
		"-" + time.Now().Format("20060102-150405")
	dest, err := a.askSharebackupDest(base)
	if err != nil {
		return nil, err
	}
	if dest == "" {
		return nil, nil // pengguna membatalkan
	}
	if !strings.EqualFold(filepath.Ext(dest), ".sphbak") {
		dest += ".sphbak"
	}
	if err := a.share.WriteFile(pkg, dest); err != nil {
		return nil, err
	}
	c := pkg.Counts()
	return &ShareBackupCreateResult{
		Path:  dest,
		Items: c.Sph + c.WorkItems + c.Categories + c.Templates + c.Customers + c.Materials,
	}, nil
}

// askSharebackupDest membuka dialog simpan untuk file .sphbak.
func (a *App) askSharebackupDest(defaultName string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "Simpan Backup yang Dapat Dibagikan",
		DefaultDirectory:     a.cfg.BackupDir,
		DefaultFilename:      defaultName + ".sphbak",
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "Backup Share SPH (*.sphbak)", Pattern: "*.sphbak"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("dialog simpan file gagal dibuka")
	}
	return strings.TrimSpace(path), nil
}

// OpenShareableBackup: pilih & validasi file .sphbak, ingat path untuk restore.
func (a *App) OpenShareableBackup() (*ShareBackupPreview, error) {
	src, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih File Backup yang Dapat Dibagikan",
		Filters: []runtime.FileFilter{
			{DisplayName: "Backup Share SPH (*.sphbak)", Pattern: "*.sphbak"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dialog pemilih file gagal dibuka")
	}
	if strings.TrimSpace(src) == "" {
		return nil, nil // pengguna membatalkan
	}
	pkg, err := sharebackup.ReadFile(src)
	if err != nil {
		return nil, err
	}
	a.sharePath = src
	counts := pkg.Counts()
	preview := &ShareBackupPreview{
		Path:       src,
		DeviceName: pkg.Metadata.DeviceName,
		CreatedAt:  pkg.Metadata.CreatedAt.Format("02 Jan 2006 15:04"),
		Counts:     counts,
	}
	return preview, nil
}

// RestoreShareableBackup: restore hanya menambah data baru sesuai seksi yang
// dipilih penerima (data yang sudah ada dilewati, tidak ditimpa/dihapus).
func (a *App) RestoreShareableBackup(sections []string) (*sharebackup.InstallSummary, error) {
	a.restoreMu.Lock()
	defer a.restoreMu.Unlock()

	if room := a.collabMgr.CurrentRoomID(); room != "" {
		return nil, services.NewValidationError("Tutup sesi Work Together terlebih dahulu sebelum memulihkan backup.")
	}
	if strings.TrimSpace(a.sharePath) == "" {
		return nil, services.NewValidationError("Pilih file backup terlebih dahulu.")
	}
	if len(sections) == 0 {
		return nil, services.NewValidationError("Pilih minimal satu seksi data yang ingin dipulihkan.")
	}
	pkg, err := sharebackup.ReadFile(a.sharePath)
	if err != nil {
		return nil, err
	}
	opts := sharebackup.InstallSelectedSections(sections)
	sum, err := a.share.Restore(pkg, opts)
	if err != nil {
		return nil, err
	}
	a.writeShareRestoreAudit(pkg.Metadata.DeviceName)
	return sum, nil
}

func (a *App) writeShareRestoreAudit(sourceDevice string) {
	desc := "Data dipulihkan dari backup yang dibagikan"
	if sourceDevice != "" {
		desc += " (" + sourceDevice + ")"
	}
	if err := a.db.Create(&models.AuditLog{Action: "RESTORE", Entity: "backup", Description: desc}).Error; err != nil {
		a.log.Warn("gagal mencatat audit RESTORE backup share", "error", err)
	}
}