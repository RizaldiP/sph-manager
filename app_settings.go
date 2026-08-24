package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/RizaldiP/sph-manager/internal/services"
)

// Binding Pengaturan aplikasi (FR-U4).

func (a *App) GetSettings() (*services.SettingsView, error) {
	return a.settings.Get()
}

func (a *App) UpdateSettings(in *services.SettingsInput) (*services.SettingsView, error) {
	return a.settings.Update(in)
}

func (a *App) PreviewSphNumber(format string) (string, error) {
	return a.settings.PreviewNumber(format)
}

// PickLogo membuka dialog pilih gambar, menyalinnya ke folder data aplikasi,
// lalu menyimpan path-nya di settings. Batal memilih = tidak ada perubahan.
func (a *App) PickLogo() (*services.SettingsView, error) {
	src, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih File Logo",
		Filters: []runtime.FileFilter{
			{DisplayName: "Gambar (*.png;*.jpg;*.jpeg)", Pattern: "*.png;*.jpg;*.jpeg"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dialog pemilih file gagal dibuka")
	}
	if strings.TrimSpace(src) == "" {
		return a.settings.Get()
	}

	ext := strings.ToLower(filepath.Ext(src))
	switch ext {
	case ".png", ".jpg", ".jpeg":
	default:
		return nil, services.NewValidationError("Format logo harus PNG atau JPG.")
	}
	if info, ierr := os.Stat(src); ierr != nil || info.IsDir() {
		return nil, services.NewValidationError("File logo tidak dapat dibaca.")
	}
	if info, _ := os.Stat(src); info.Size() > 5*1024*1024 {
		return nil, services.NewValidationError("Ukuran logo maksimal 5 MB.")
	}

	if err := os.MkdirAll(a.cfg.AssetsDir, 0755); err != nil {
		a.log.Error("gagal menyiapkan folder assets", "error", err)
		return nil, fmt.Errorf("gagal menyimpan logo")
	}
	dest := filepath.Join(a.cfg.AssetsDir, "logo"+ext)
	if err := copyFile(src, dest); err != nil {
		a.log.Error("gagal menyalin logo", "sumber", src, "error", err)
		return nil, fmt.Errorf("gagal menyimpan logo")
	}
	return a.settings.SetLogo(dest)
}

func (a *App) ClearLogo() (*services.SettingsView, error) {
	view, err := a.settings.Get()
	if err != nil {
		return nil, err
	}
	if view.LogoPath != "" {
		_ = os.Remove(view.LogoPath)
	}
	return a.settings.SetLogo("")
}

// LogoDataUrl mengembalikan logo sebagai data URL untuk pratinjau di UI;
// string kosong bila belum ada logo.
func (a *App) LogoDataUrl() (string, error) {
	view, err := a.settings.Get()
	if err != nil {
		return "", err
	}
	if view.LogoPath == "" {
		return "", nil
	}
	data, err := os.ReadFile(view.LogoPath)
	if err != nil {
		return "", nil
	}
	ext := strings.ToLower(filepath.Ext(view.LogoPath))
	mime := "image/png"
	if ext == ".jpg" || ext == ".jpeg" {
		mime = "image/jpeg"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
