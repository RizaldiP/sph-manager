package services

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func TestSettingsDefaultsAndGetUpdate(t *testing.T) {
	db := serviceDB(t)
	svc := NewSettingsService(db, slog.Default())

	def, err := svc.Get()
	if err != nil {
		t.Fatalf("get default gagal: %v", err)
	}
	if def.CompanyName != "PT. Ganesha Energi Indonesia" {
		t.Errorf("nama perusahaan default salah: %q", def.CompanyName)
	}
	if def.SphNumberFormat != "SPH/GEI/{ROMAN}/{YYYY}/{SEQ}" {
		t.Errorf("format default salah: %q", def.SphNumberFormat)
	}
	if def.SignerName != "Matawai" || def.SignerPosition != "Direktur" {
		t.Errorf("penandatangan default salah: %q / %q", def.SignerName, def.SignerPosition)
	}

	upd, err := svc.Update(&SettingsInput{
		CompanyName:     "  PT. Uji Coba  ",
		CompanyCity:     "Jakarta",
		CompanyAddress:  "Jl. Merdeka No. 1",
		SphNumberFormat: "SPH/{YYYY}/{SEQ}",
		SignerName:      "Budi",
		SignerPosition:  "Manajer",
		DefaultNotes:    "Harga berlaku 30 hari",
	})
	if err != nil {
		t.Fatalf("update gagal: %v", err)
	}
	if upd.CompanyName != "PT. Uji Coba" || upd.CompanyCity != "Jakarta" || upd.DefaultNotes != "Harga berlaku 30 hari" {
		t.Errorf("hasil update salah: %+v", upd)
	}

	got, _ := svc.Get()
	if got.CompanyName != "PT. Uji Coba" {
		t.Errorf("nilai tidak persisten: %+v", got)
	}
	if got.LogoPath != "" {
		t.Errorf("logo harus tetap kosong via Update: %q", got.LogoPath)
	}
}

func TestSettingsValidation(t *testing.T) {
	db := serviceDB(t)
	svc := NewSettingsService(db, slog.Default())

	if _, err := svc.Update(&SettingsInput{CompanyName: "   ", SphNumberFormat: "SPH/{YYYY}/{SEQ}"}); err == nil {
		t.Error("nama perusahaan kosong harus ditolak")
	}
	if _, err := svc.Update(&SettingsInput{CompanyName: "PT X", SphNumberFormat: "SPH/{YYYY}/tanpa-seq"}); err == nil {
		t.Error("format tanpa {SEQ} harus ditolak")
	}
	long := strings.Repeat("x", 201)
	if _, err := svc.Update(&SettingsInput{CompanyName: long, SphNumberFormat: "SPH/{SEQ}"}); err == nil {
		t.Error("nama >200 karakter harus ditolak")
	}

	// PreviewNumber.
	pv, err := svc.PreviewNumber("")
	if err != nil {
		t.Fatalf("preview format kosong gagal: %v", err)
	}
	if !strings.Contains(pv, "/2026/") && !strings.Contains(pv, "/") {
		t.Errorf("preview aneh: %q", pv)
	}
	if !strings.HasSuffix(pv, "001") {
		t.Errorf("preview harus berakhiran 001: %q", pv)
	}
	if _, err := svc.PreviewNumber("TANPA-SEQ"); err == nil {
		t.Error("preview tanpa {SEQ} harus ditolak")
	}
}

func TestSettingsLogoAndAudit(t *testing.T) {
	db := serviceDB(t)
	svc := NewSettingsService(db, slog.Default())

	logo := filepath.Join(t.TempDir(), "logo.png")
	if err := os.WriteFile(logo, []byte("png"), 0644); err != nil {
		t.Fatalf("fixture logo gagal: %v", err)
	}

	if _, err := svc.SetLogo(logo); err != nil {
		t.Fatalf("set logo gagal: %v", err)
	}
	view, _ := svc.Get()
	if view.LogoPath != logo {
		t.Errorf("path logo tidak tersimpan: %q", view.LogoPath)
	}

	cleared, err := svc.SetLogo("")
	if err != nil {
		t.Fatalf("clear logo gagal: %v", err)
	}
	if cleared.LogoPath != "" {
		t.Errorf("logo harus kosong: %q", cleared.LogoPath)
	}

	var count int64
	if err := db.Model(&models.AuditLog{}).Where("entity = ?", "settings").Count(&count).Error; err != nil {
		t.Fatalf("hitung audit gagal: %v", err)
	}
	// set logo (1) + clear logo (1).
	if count != 2 {
		t.Errorf("audit log settings = %d, mau 2", count)
	}
}
