package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/RizaldiP/sph-manager/internal/importers"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// Binding Import Excel (FR-IE1..IE4). Alur wajib:
// pilih file → pilih sheet → mapping kolom → pratinjau & klasifikasi → import.
func (a *App) PickImportFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih File Excel",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel (*.xls;*.xlsx)", Pattern: "*.xls;*.xlsx"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("dialog pemilih file gagal dibuka")
	}
	return strings.TrimSpace(path), nil
}

func (a *App) ListImportSheets(path string) ([]string, error) {
	return importers.ListSheets(path)
}

func (a *App) PreviewImportSheet(path, sheet string) (*importers.SheetPreview, error) {
	return importers.BuildSheetPreview(path, sheet)
}

func (a *App) ParseImportRows(path, sheet string, m importers.ColumnMapping) ([]importers.PreviewRow, error) {
	g, err := importers.ReadSheet(path, sheet)
	if err != nil {
		return nil, err
	}
	return importers.ParseRows(g, m), nil
}

func (a *App) ValidateImportRows(path, sheet string, m importers.ColumnMapping, confirms []services.ConfirmRow) ([]string, error) {
	g, err := importers.ReadSheet(path, sheet)
	if err != nil {
		return nil, err
	}
	rows := importers.ParseRows(g, m)
	svc := services.NewImportService(a.db, a.log)
	return svc.ValidateRows(rows, confirms), nil
}

func (a *App) RunWorkItemImport(categoryID uint, path, sheet string, m importers.ColumnMapping, confirms []services.ConfirmRow) (*services.ImportResult, error) {
	if strings.TrimSpace(filepath.Ext(path)) == "" {
		return nil, fmt.Errorf("file belum dipilih")
	}
	svc := services.NewImportService(a.db, a.log)
	svc.SetProgress(func(done, total int) error {
		runtime.EventsEmit(a.ctx, "import:progress", map[string]int{"done": done, "total": total})
		return nil
	})
	res, err := svc.ImportWorkItems(categoryID, path, sheet, m, confirms)
	runtime.EventsEmit(a.ctx, "import:done", err == nil)
	return res, err
}
