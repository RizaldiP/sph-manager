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
	return a.pickAsset("logo", "Logo", a.settings.SetLogo, ".png", ".jpg", ".jpeg")
}

// PickStamp membuka dialog pilih gambar stempel (PNG, mempertahankan
// transparansi), menyalinnya ke folder data aplikasi, lalu menyimpan path-nya.
func (a *App) PickStamp() (*services.SettingsView, error) {
	return a.pickAsset("stamp", "Stempel", a.settings.SetStamp, ".png")
}

// PickSignature membuka dialog pilih gambar tanda tangan (PNG), menyalinnya
// ke folder data aplikasi, lalu menyimpan path-nya.
func (a *App) PickSignature() (*services.SettingsView, error) {
	return a.pickAsset("signature", "Tanda tangan", a.settings.SetSignature, ".png")
}

// pickAsset alur umum upload gambar pengaturan: dialog pilih file, validasi
// ekstensi & ukuran, salin ke AssetsDir/<name><ext>, lalu simpan path-nya.
// set menyimpan path hasil upload; ekstensi / PNG dikontrol pemanggil.
func (a *App) pickAsset(name, label string, set func(string) (*services.SettingsView, error), allowedExt ...string) (*services.SettingsView, error) {
	extLabel := strings.Join(allowedExt, ";")
	src, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:  "Pilih File " + label,
		Filters: []runtime.FileFilter{
			{DisplayName: "Gambar (*" + strings.Join(allowedExt, ";*") + ")", Pattern: "*" + strings.Join(allowedExt, ";*")},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dialog pemilih file gagal dibuka")
	}
	if strings.TrimSpace(src) == "" {
		return a.settings.Get()
	}

	ext := strings.ToLower(filepath.Ext(src))
	ok := false
	for _, e := range allowedExt {
		if ext == e {
			ok = true
			break
		}
	}
	if !ok {
		return nil, services.NewValidationError("Format %s harus: %s.", label, extLabel)
	}
	if info, ierr := os.Stat(src); ierr != nil || info.IsDir() {
		return nil, services.NewValidationError("File %s tidak dapat dibaca.", label)
	}
	if info, _ := os.Stat(src); info.Size() > 5*1024*1024 {
		return nil, services.NewValidationError("Ukuran %s maksimal 5 MB.", label)
	}

	if err := os.MkdirAll(a.cfg.AssetsDir, 0755); err != nil {
		a.log.Error("gagal menyiapkan folder assets", "error", err)
		return nil, fmt.Errorf("gagal menyimpan %s", label)
	}
	dest := filepath.Join(a.cfg.AssetsDir, name+ext)
	if err := copyFile(src, dest); err != nil {
		a.log.Error("gagal menyalin file", "sumber", src, "error", err)
		return nil, fmt.Errorf("gagal menyimpan %s", label)
	}
	view, serr := set(dest)
	if serr != nil {
		return nil, serr
	}
	return view, nil
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

func (a *App) ClearStamp() (*services.SettingsView, error) {
	view, err := a.settings.Get()
	if err != nil {
		return nil, err
	}
	if view.StampPath != "" {
		_ = os.Remove(view.StampPath)
	}
	return a.settings.SetStamp("")
}

func (a *App) ClearSignature() (*services.SettingsView, error) {
	view, err := a.settings.Get()
	if err != nil {
		return nil, err
	}
	if view.SignaturePath != "" {
		_ = os.Remove(view.SignaturePath)
	}
	return a.settings.SetSignature("")
}

// SetStampPosition menyimpan posisi & ukuran stempel (fraksi 0-1) dari editor.
func (a *App) SetStampPosition(x, y, size float64) (*services.SettingsView, error) {
	return a.settings.SetStampPosition(x, y, size)
}

// SetSignaturePosition menyimpan posisi & ukuran tanda tangan (fraksi 0-1).
func (a *App) SetSignaturePosition(x, y, size float64) (*services.SettingsView, error) {
	return a.settings.SetSignaturePosition(x, y, size)
}

// LogoDataUrl mengembalikan logo sebagai data URL untuk pratinjau di UI;
// string kosong bila belum ada logo.
func (a *App) LogoDataUrl() (string, error) {
	view, err := a.settings.Get()
	if err != nil {
		return "", err
	}
	return dataURLFor(view.LogoPath)
}

// StampDataUrl mengembalikan stempel sebagai data URL untuk pratinjau di UI.
func (a *App) StampDataUrl() (string, error) {
	view, err := a.settings.Get()
	if err != nil {
		return "", err
	}
	return dataURLFor(view.StampPath)
}

// SignatureDataUrl mengembalikan tanda tangan sebagai data URL untuk pratinjau.
func (a *App) SignatureDataUrl() (string, error) {
	view, err := a.settings.Get()
	if err != nil {
		return "", err
	}
	return dataURLFor(view.SignaturePath)
}

// dataURLFor mengubah path gambar menjadi data URL; kosong bila path kosong.
func dataURLFor(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	ext := strings.ToLower(filepath.Ext(path))
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
