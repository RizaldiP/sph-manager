package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/RizaldiP/sph-manager/internal/services"
	"github.com/RizaldiP/sph-manager/internal/updater"
)

// updateCandidateTimeout dipakai saat mengunduh file kandidat utuh (~22MB)
// yang lebih lambat daripada feed.
const updateCandidateTimeout = 120 * time.Second

// defaultUpdateSourceURL adalah link Google Drive ke file aplikasi
// (SPHManager.exe) yang dibakar ke binary. Tiap rilis, file itu diganti
// isinya via "Manage versions -> Upload new version" sehingga link tetap.
const defaultUpdateSourceURL = ""

// UpdateStatus potret hasil pemeriksaan versi terbaru dari Google Drive.
type UpdateStatus struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	Notes          string `json:"notes"`
}

// CurrentVersion mengembalikan versi aplikasi yang sedang berjalan.
func (a *App) CurrentVersion() string {
	return appVersion
}

// updateSource memakai link pengaturan bila diisi, selain itu memakai link
// default yang dibakar ke binary.
func (a *App) updateSource() string {
	view, err := a.settings.Get()
	if err == nil {
		if s := strings.TrimSpace(view.UpdateSourceURL); s != "" {
			return s
		}
	}
	return defaultUpdateSourceURL
}

// candidatePath lokasi sementara file kandidat update yang diunduh.
func (a *App) candidatePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("tidak dapat menentukan lokasi aplikasi")
	}
	return filepath.Join(os.TempDir(), filepath.Base(exe)+".candidate"), nil
}

// CheckForUpdate mengunduh exe kandidat, membaca versinya dari trailer file,
// lalu membandingkannya dengan versi lokal. Tanpa link dianggap tidak ada
// update. File kandidat disimpan untuk dipakai saat pemasangan.
func (a *App) CheckForUpdate() (*UpdateStatus, error) {
	source := a.updateSource()
	if source == "" {
		return &UpdateStatus{CurrentVersion: appVersion, HasUpdate: false}, nil
	}
	dest, err := a.candidatePath()
	if err != nil {
		return &UpdateStatus{CurrentVersion: appVersion}, err
	}
	meta, err := updater.FetchCandidate(source, dest, updateCandidateTimeout, func(done, total int64) {
		runtime.EventsEmit(a.ctx, "update:progress", map[string]int64{"done": done, "total": total})
	})
	if err != nil {
		return &UpdateStatus{CurrentVersion: appVersion}, err
	}
	cmp, err := updater.CompareVersions(appVersion, meta.Version)
	if err != nil {
		return &UpdateStatus{CurrentVersion: appVersion, LatestVersion: meta.Version}, fmt.Errorf("versi kandidat tidak valid: %s", err.Error())
	}
	return &UpdateStatus{
		CurrentVersion: appVersion,
		LatestVersion:  meta.Version,
		HasUpdate:      cmp < 0,
		Notes:          meta.Notes,
	}, nil
}

// BackupForUpdate membuat backup manual sebagai syarat menjalankan update.
// Tanpa langkah ini, unduhan & pemasangan ditolak oleh backend.
func (a *App) BackupForUpdate() (*services.BackupInfo, error) {
	info, err := a.backup.CreateManual()
	if err != nil {
		a.updMu.Lock()
		a.updBackupReady = false
		a.updMu.Unlock()
		return nil, err
	}
	a.updMu.Lock()
	a.updBackupReady = true
	a.updMu.Unlock()
	return info, nil
}

// DownloadAndApplyUpdate mengunduh exe kandidat (atau memakai file sisa
// pemeriksaan), mengganti aplikasi yang berjalan, lalu memulai ulang. Wajib
// menjalankan BackupForUpdate lebih dulu.
func (a *App) DownloadAndApplyUpdate() error {
	a.updMu.Lock()
	ready := a.updBackupReady
	a.updMu.Unlock()
	if !ready {
		return services.NewValidationError("Backup data wajib dibuat terlebih dahulu sebelum update.")
	}

	source := a.updateSource()
	if source == "" {
		return services.NewValidationError("Link Google Drive update belum diatur. Atur di halaman Pengaturan terlebih dahulu.")
	}
	newExe, err := a.candidatePath()
	if err != nil {
		return err
	}
	meta, err := updater.FetchCandidate(source, newExe, updateCandidateTimeout, func(done, total int64) {
		runtime.EventsEmit(a.ctx, "update:progress", map[string]int64{"done": done, "total": total})
	})
	if err != nil {
		_ = os.Remove(newExe)
		return err
	}
	if cmp, cerr := updater.CompareVersions(appVersion, meta.Version); cerr != nil || cmp >= 0 {
		_ = os.Remove(newExe)
		return services.NewValidationError("Aplikasi sudah versi terbaru.")
	}
	runtime.EventsEmit(a.ctx, "update:progress", map[string]int64{"done": -1, "total": -1})

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("tidak dapat menemukan executable aplikasi")
	}
	if err := installUpdate(exe, newExe); err != nil {
		_ = os.Remove(newExe)
		return err
	}
	a.log.Info("update siap diterapkan, aplikasi akan ditutup", "versi", meta.Version)
	runtime.Quit(a.ctx)
	return nil
}