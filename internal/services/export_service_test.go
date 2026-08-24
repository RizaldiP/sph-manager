package services

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// seedExportDocument menyiapkan SPH contoh + mengembalikan ID-nya.
func seedExportDocument(t *testing.T, sph *SphService) uint {
	t.Helper()
	db := sph.db
	cust := &models.Customer{Code: "CUS-EXP", Name: "PT Laut Biru"}
	if err := db.Create(cust).Error; err != nil {
		t.Fatalf("seed customer gagal: %v", err)
	}
	in := sampleSphInput(cust.ID)
	view, err := sph.Create(in)
	if err != nil {
		t.Fatalf("create SPH gagal: %v", err)
	}
	return view.ID
}

func TestExportServiceDataAndFiles(t *testing.T) {
	db := serviceDB(t)
	sphSvc := NewSphService(db, slog.Default())
	setSvc := NewSettingsService(db, slog.Default())
	expSvc := NewExportService(sphSvc, setSvc, slog.Default())

	id := seedExportDocument(t, sphSvc)

	// ExportData: profil default + snapshot dokumen.
	data, err := expSvc.ExportData(id)
	if err != nil {
		t.Fatalf("ExportData gagal: %v", err)
	}
	if data.Company.Name != "PT. Ganesha Energi Indonesia" || data.Company.SignerName == "" {
		t.Errorf("profil perusahaan tidak terisi: %+v", data.Company)
	}
	if data.Document.Customer != "PT Laut Biru" {
		t.Errorf("customer tidak ikut: %+v", data.Document)
	}
	if len(data.Rows) != 3 { // 2 main + 1 sub (dari sampleSphInput)
		t.Errorf("baris harus 3, dapat %d", len(data.Rows))
	}
	if data.GrandTotal != 2*5_000_000+2*500_000+3_000_000+250_000 {
		t.Errorf("grand total salah: %d", data.GrandTotal)
	}
	if data.Document.Terbilang == "" {
		t.Error("terbilang kosong")
	}

	dir := t.TempDir()

	// Excel
	xlPath, err := expSvc.ExportExcel(id, filepath.Join(dir, "SPH_test"))
	if err != nil {
		t.Fatalf("ExportExcel gagal: %v", err)
	}
	if !strings.HasSuffix(xlPath, ".xlsx") {
		t.Errorf("ekstensi otomatis gagal: %q", xlPath)
	}
	if _, err := os.Stat(xlPath); err != nil {
		t.Fatalf("file xlsx tidak ada: %v", err)
	}
	f, err := excelize.OpenFile(xlPath)
	if err != nil {
		t.Fatalf("xlsx rusak: %v", err)
	}
	title, _ := f.GetCellValue("SPH", "B5")
	if title == "" {
		t.Error("judul Excel kosong")
	}
	_ = f.Close()

	// PDF landscape + portrait
	for _, orient := range []string{"landscape", "portrait"} {
		pdfPath, err := expSvc.ExportPDF(id, filepath.Join(dir, "SPH_"+orient), orient)
		if err != nil {
			t.Fatalf("ExportPDF(%s) gagal: %v", orient, err)
		}
		raw, err := os.ReadFile(pdfPath)
		if err != nil {
			t.Fatalf("file pdf tidak ada: %v", err)
		}
		if len(raw) < 100 || string(raw[:5]) != "%PDF-" {
			t.Errorf("pdf %s tidak valid (%d byte)", orient, len(raw))
		}
	}

	// Audit EXPORT tercatat untuk 3 file.
	var n int64
	db.Model(&models.AuditLog{}).Where("action = ? AND entity = ?", "EXPORT", "sph_document").Count(&n)
	if n != 3 {
		t.Errorf("audit export harus 3, dapat %d", n)
	}
}

func TestExportServiceValidationAndNotFound(t *testing.T) {
	db := serviceDB(t)
	sphSvc := NewSphService(db, slog.Default())
	expSvc := NewExportService(sphSvc, NewSettingsService(db, slog.Default()), slog.Default())

	if _, err := expSvc.ExportData(9999); err != ErrNotFound {
		t.Errorf("harus ErrNotFound, dapat: %v", err)
	}
	if _, err := expSvc.ExportPDF(1, "x.pdf", "diagonal"); err == nil {
		t.Error("orientasi aneh harus ditolak")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("harus ValidationError, dapat: %T", err)
	}
	missingDir := filepath.Join(os.TempDir(), "pasti-tidak-ada-xyz", "out.xlsx")
	if _, err := expSvc.ExportExcel(1, missingDir); err == nil {
		t.Error("folder hilang harus ditolak")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("harus ValidationError, dapat: %T", err)
	}
}
