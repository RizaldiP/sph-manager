package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Binding export dokumen SPH (FR-IE5/FR-IE6, Phase 9).

var unsafeFilenameRe = regexp.MustCompile(`[\\/:*?"<>|]+`)

// sanitizeFilename buang karakter terlarang Windows dari nomor dokumen.
func sanitizeFilename(s string) string {
	return strings.Trim(unsafeFilenameRe.ReplaceAllString(strings.TrimSpace(s), "-"), " .")
}

// defaultExportName menyusun nama file dasar: SPH_<nomor>_rev<N>.
func defaultExportName(number string, revision int) string {
	name := "SPH_" + sanitizeFilename(number)
	if name == "" || name == "SPH_" {
		name = "SPH_Dokumen"
	}
	if revision > 0 {
		name += fmt.Sprintf("_rev%d", revision)
	}
	return name
}

// askExportDest membuka dialog Simpan; mengembalikan path ("" bila batal).
func (a *App) askExportDest(defaultName, ext, filterLabel string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "Simpan Dokumen",
		DefaultDirectory:     a.cfg.ExportDir,
		DefaultFilename:      defaultName + ext,
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: filterLabel, Pattern: "*" + ext},
		},
	})
	if err != nil {
		return "", fmt.Errorf("dialog simpan file gagal dibuka")
	}
	return strings.TrimSpace(path), nil
}

// ExportSphExcel: dialog simpan → tulis .xlsx → balik path final untuk banner UI.
func (a *App) ExportSphExcel(id uint) (string, error) {
	data, err := a.export.ExportData(id)
	if err != nil {
		return "", err
	}
	dest, err := a.askExportDest(
		defaultExportName(data.Document.Number, data.Document.Revision),
		".xlsx", "Excel (*.xlsx)",
	)
	if err != nil {
		return "", err
	}
	if dest == "" {
		return "", nil // pengguna membatalkan
	}
	return a.export.ExportExcel(id, dest)
}

// ExportSphPdf: dialog simpan → tulis .pdf sesuai orientasi.
func (a *App) ExportSphPdf(id uint, orientation string) (string, error) {
	data, err := a.export.ExportData(id)
	if err != nil {
		return "", err
	}
	suffix := ""
	if strings.EqualFold(strings.TrimSpace(orientation), "portrait") {
		suffix = "_portrait"
	}
	dest, err := a.askExportDest(
		defaultExportName(data.Document.Number, data.Document.Revision)+suffix,
		".pdf", "PDF (*.pdf)",
	)
	if err != nil {
		return "", err
	}
	if dest == "" {
		return "", nil
	}
	return a.export.ExportPDF(id, dest, orientation)
}

// OpenExportFolder membuka folder berisi hasil export di file manager.
func (a *App) OpenExportFolder(path string) {
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		runtime.BrowserOpenURL(a.ctx, dir)
	}
}
