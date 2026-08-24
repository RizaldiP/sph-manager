package services

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/RizaldiP/sph-manager/internal/exporters"
)

// ExportService menyatukan dokumen SPH + profil perusahaan lalu
// mendelegasikan penggambaran ke package exporters (FR-IE5/FR-IE6).
type ExportService struct {
	sph      *SphService
	settings *SettingsService
	log      *slog.Logger
	audit    *AuditWriter
}

func NewExportService(sph *SphService, settings *SettingsService, log *slog.Logger) *ExportService {
	return &ExportService{sph: sph, settings: settings, log: log, audit: NewAuditWriter()}
}

// ExportData membangun data netral (tanpa detail DB) untuk generator Excel/PDF.
func (s *ExportService) ExportData(id uint) (*exporters.ExportData, error) {
	doc, err := s.sph.Get(id)
	if err != nil {
		return nil, err
	}
	view, err := s.settings.Get()
	if err != nil {
		return nil, fmt.Errorf("gagal memuat profil perusahaan: %w", err)
	}
	data := exporters.BuildData(doc, exporters.CompanyInfo{
		Name:           view.CompanyName,
		City:           view.CompanyCity,
		Address:        view.CompanyAddress,
		LogoPath:       view.LogoPath,
		SignerName:     view.SignerName,
		SignerPosition: view.SignerPosition,
	})
	return data, nil
}

// ExportExcel menulis file .xlsx ke dest; mengembalikan path final.
func (s *ExportService) ExportExcel(id uint, dest string) (string, error) {
	dest, err := normalizeDest(dest, ".xlsx")
	if err != nil {
		return "", err
	}
	data, err := s.ExportData(id)
	if err != nil {
		return "", err
	}
	f, err := exporters.BuildExcel(data)
	if err != nil {
		s.log.Error("gagal menyusun Excel SPH", "id", id, "error", err)
		return "", fmt.Errorf("gagal menyusun file Excel")
	}
	defer func() { _ = f.Close() }()
	if err := f.SaveAs(dest); err != nil {
		s.log.Error("gagal menyimpan Excel SPH", "dest", dest, "error", err)
		return "", fmt.Errorf("gagal menyimpan file Excel")
	}
	s.record(id, "Excel", dest, len(data.Rows))
	return dest, nil
}

var pdfOrientations = map[string]bool{"portrait": true, "landscape": true}

// ExportPDF menulis file .pdf ke dest dengan orientasi pilihan; path final dikembalikan.
func (s *ExportService) ExportPDF(id uint, dest, orientation string) (string, error) {
	orientation = strings.ToLower(strings.TrimSpace(orientation))
	if orientation == "" {
		orientation = "landscape"
	}
	if !pdfOrientations[orientation] {
		return "", NewValidationError("Orientasi harus landscape atau portrait.")
	}
	dest, err := normalizeDest(dest, ".pdf")
	if err != nil {
		return "", err
	}
	data, err := s.ExportData(id)
	if err != nil {
		return "", err
	}
	res, err := exporters.BuildPDF(data, exporters.PDFOptions{Landscape: orientation == "landscape"})
	if err != nil {
		s.log.Error("gagal menyusun PDF SPH", "id", id, "error", err)
		return "", fmt.Errorf("gagal menyusun file PDF")
	}
	if err := os.WriteFile(dest, res.Bytes, 0o644); err != nil {
		s.log.Error("gagal menyimpan PDF SPH", "dest", dest, "error", err)
		return "", fmt.Errorf("gagal menyimpan file PDF")
	}
	s.log.Info("PDF SPH diekspor", "id", id, "halaman", res.Pages, "dest", dest)
	s.record(id, "PDF", dest, len(data.Rows))
	return dest, nil
}

// record mencatat audit export (BR-13); kegagalan audit tidak membatalkan export.
func (s *ExportService) record(id uint, kind, dest string, rows int) {
	db := s.sph.db
	desc := fmt.Sprintf("Ekspor %s (%d baris) ke %s", kind, rows, filepath.Base(dest))
	if err := s.audit.Write(db, "EXPORT", "sph_document", id, desc); err != nil {
		s.log.Warn("gagal mencatat audit export", "id", id, "error", err)
	}
}

// normalizeDest memastikan folder tujuan ada dan ekstensi sesuai format.
func normalizeDest(dest, ext string) (string, error) {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return "", NewValidationError("Lokasi penyimpanan belum dipilih.")
	}
	dir := filepath.Dir(dest)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", NewValidationError("Folder tujuan tidak ditemukan: %s.", dir)
	}
	if strings.ToLower(filepath.Ext(dest)) != ext {
		dest += ext
	}
	return dest, nil
}
