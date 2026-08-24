package services

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/RizaldiP/sph-manager/internal/importers"
	"github.com/RizaldiP/sph-manager/internal/models"
)

// importFixtureXlsx: file Excel gaya A untuk test service.
func importFixtureXlsx(t *testing.T) string {
	t.Helper()
	f := excelize.NewFile()
	const sheet = "SPH"
	if _, err := f.NewSheet(sheet); err != nil {
		t.Fatalf("new sheet: %v", err)
	}
	f.DeleteSheet("Sheet1")
	rows := [][]interface{}{
		{"NO", "URAIAN KEGIATAN", "JML", "SAT", "", "JASA", "MAT"},
		{"A", "Laksanakan Perbaikan Sistem Penjalan", "", "", "", "", ""},
		{"1.", "Laksanakan Perbaikan Sistem Penjalan :", "", "", "", "", ""},
		{"", "a. Ganti Bearing Kompressor", 2, "bh", "", 24000000, ""},
		{"", "b. Ganti Seal Pump", 1, "bh", "", "", 3500000},
		{"2.", "Laksanakan Ganti Baru Sistem Penjalan :", "", "", "", "", ""},
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				t.Fatalf("set cell: %v", err)
			}
		}
	}
	path := filepath.Join(t.TempDir(), "import.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("save as: %v", err)
	}
	return path
}

func importMapping() importers.ColumnMapping {
	return importers.ColumnMapping{NameCol: 1, QtyCol: 2, UnitCol: 3, ServiceCol: 5, MaterialCol: 6,
		UnitPriceCol: -1, ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
}

func confirmAll(rows []importers.PreviewRow) []ConfirmRow {
	out := make([]ConfirmRow, 0, len(rows))
	for _, r := range rows {
		lvl := r.Suggested
		if lvl == importers.LevelUnknown {
			continue
		}
		out = append(out, ConfirmRow{RowIndex: r.RowIndex, Level: lvl})
	}
	return out
}

func TestImportWorkItemsHappyPath(t *testing.T) {
	db := serviceDB(t)
	cats := NewCategoryService(db, slog.Default())
	cat, err := cats.Create(&models.Category{Code: "ME", Name: "Mechanical"})
	if err != nil {
		t.Fatalf("create kategori: %v", err)
	}
	// Item yang sudah ada → sequence harus melanjutkan (BR-11).
	items := NewWorkItemService(db, slog.Default())
	existing, err := items.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Sudah Ada", Sequence: 5, DefaultQuantity: 1})
	if err != nil {
		t.Fatalf("create pekerjaan awal: %v", err)
	}

	path := importFixtureXlsx(t)
	svc := NewImportService(db, slog.Default())
	g, _ := importers.ReadSheet(path, "SPH")
	rows := importers.ParseRows(g, importMapping())
	res, err := svc.ImportWorkItems(cat.ID, path, "SPH", importMapping(), confirmAll(rows))
	if err != nil {
		t.Fatalf("ImportWorkItems gagal: %v", err)
	}
	if res.ItemsCreated != 3 || res.SubsCreated != 2 || res.Skipped != 0 {
		t.Errorf("hasil salah: %+v", res)
	}

	list, err := items.List(cat.ID, true, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 4 {
		t.Fatalf("jumlah pekerjaan = %d, mau 4", len(list))
	}
	var imported []models.WorkItem
	if err := db.Where("category_id = ? AND id <> ?", cat.ID, existing.ID).
		Order("sequence").Find(&imported).Error; err != nil {
		t.Fatalf("query import: %v", err)
	}
	if len(imported) != 3 {
		t.Fatalf("pekerjaan hasil import = %d, mau 3", len(imported))
	}
	// Item awal otomatis mendapat sequence 1; hasil import melanjutkan 2,3,4.
	for i, w := range imported {
		if w.Sequence != 2+i {
			t.Errorf("sequence item %d = %d, mau %d", i, w.Sequence, 2+i)
		}
		if !strings.HasPrefix(w.Code, "PEK-") {
			t.Errorf("kode = %q, harus PEK-", w.Code)
		}
		if w.CategoryID != cat.ID {
			t.Errorf("kategori salah: %d", w.CategoryID)
		}
	}

	// Sub-pekerjaan + konversi total→satuan.
	var bearing *models.WorkSubItem
	ids := make([]uint, 0, len(imported))
	for _, w := range imported {
		ids = append(ids, w.ID)
	}
	var subs []models.WorkSubItem
	if err := db.Where("work_item_id IN ?", ids).Order("sequence").Find(&subs).Error; err != nil {
		t.Fatalf("query subs: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("jumlah sub = %d, mau 2", len(subs))
	}
	for i := range subs {
		if strings.Contains(subs[i].Name, "Bearing") {
			bearing = &subs[i]
		}
		if !strings.HasPrefix(subs[i].Code, "SUB-") {
			t.Errorf("kode sub = %q, harus SUB-", subs[i].Code)
		}
	}
	if bearing == nil {
		t.Fatal("sub Ganti Bearing tidak ditemukan")
	}
	if bearing.DefaultServicePrice != 12_000_000 || bearing.DefaultQuantity != 2 || bearing.DefaultUnit != "bh" {
		t.Errorf("nilai sub salah: %+v", bearing)
	}

	// Audit tercatat.
	var auditCount int64
	db.Model(&models.AuditLog{}).Where("action = ? AND entity IN ?", "CREATE", []string{"work_item", "work_sub_item"}).Count(&auditCount)
	if auditCount < 5 {
		t.Errorf("audit log kurang: %d", auditCount)
	}
}

func TestImportRollbackOnError(t *testing.T) {
	db := serviceDB(t)
	cats := NewCategoryService(db, slog.Default())
	cat, _ := cats.Create(&models.Category{Code: "EL", Name: "Electrical"})

	path := importFixtureXlsx(t)
	svc := NewImportService(db, slog.Default())
	svc.SetProgress(func(done, total int) error {
		return fmt.Errorf("dibatalkan pengguna")
	})
	g, _ := importers.ReadSheet(path, "SPH")
	rows := importers.ParseRows(g, importMapping())
	res, err := svc.ImportWorkItems(cat.ID, path, "SPH", importMapping(), confirmAll(rows))
	if err == nil {
		t.Fatal("progress error seharusnya membatalkan import")
	}
	if res != nil {
		t.Error("hasil harus nil saat rollback")
	}
	var count int64
	db.Model(&models.WorkItem{}).Where("category_id = ?", cat.ID).Count(&count)
	if count != 0 {
		t.Errorf("rollback gagal: masih ada %d pekerjaan", count)
	}
}

func TestImportBlockedByUnknownAndErrors(t *testing.T) {
	f := excelize.NewFile()
	const sheet = "Data"
	if _, err := f.NewSheet(sheet); err != nil {
		t.Fatal(err)
	}
	f.DeleteSheet("Sheet1")
	cells := map[string]interface{}{
		"A1": "URAIAN KEGIATAN", "B1": "JML",
		"A2": "Baris tanpa penanda dan tanpa harga", "B2": 1,
		"A3": "Qty Nol", "B3": 0,
	}
	for cell, v := range cells {
		if err := f.SetCellValue(sheet, cell, v); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "blokir.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}

	db := serviceDB(t)
	cats := NewCategoryService(db, slog.Default())
	cat, _ := cats.Create(&models.Category{Code: "IN", Name: "Instrumen"})

	m := importers.ColumnMapping{NameCol: 0, QtyCol: 1, UnitCol: -1, ServiceCol: -1, MaterialCol: -1,
		UnitPriceCol: -1, ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	svc := NewImportService(db, slog.Default())

	g, _ := importers.ReadSheet(path, sheet)
	rows := importers.ParseRows(g, m)

	// Semua unknown dibiarkan belum diputuskan → blokir.
	blockers := svc.ValidateRows(rows, nil)
	if len(blockers) == 0 {
		t.Fatal("harus ada blokir saat baris unknown belum diklasifikasi")
	}
	res, err := svc.ImportWorkItems(cat.ID, path, sheet, m, nil)
	if err == nil {
		t.Fatal("import harus diblokir")
	}
	if res != nil {
		t.Error("hasil harus nil saat diblokir")
	}

	// Setelah diklasifikasi manual, qty nol tetap ditolak.
	confirms := []ConfirmRow{{RowIndex: 1, Level: importers.LevelMain}, {RowIndex: 2, Level: importers.LevelSub}}
	blockers = svc.ValidateRows(rows, confirms)
	found := false
	for _, b := range blockers {
		if strings.Contains(b, "Qty tidak valid") {
			found = true
		}
	}
	if !found {
		t.Errorf("blokir qty nol hilang: %v", blockers)
	}
	if _, err := svc.ImportWorkItems(cat.ID, path, sheet, m, confirms); err == nil {
		t.Error("import dengan qty nol seharusnya gagal")
	}
	var count int64
	db.Model(&models.WorkItem{}).Where("category_id = ?", cat.ID).Count(&count)
	if count != 0 {
		t.Errorf("tidak boleh ada data tersimpan: %d", count)
	}
}

func TestImportSubTanpaIndukGagal(t *testing.T) {
	f := excelize.NewFile()
	const sheet = "Data"
	if _, err := f.NewSheet(sheet); err != nil {
		t.Fatal(err)
	}
	f.DeleteSheet("Sheet1")
	cells := map[string]interface{}{
		"A1": "URAIAN KEGIATAN",
		"A2": "a. Sub langsung tanpa induk di atasnya",
	}
	for cell, v := range cells {
		if err := f.SetCellValue(sheet, cell, v); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "tanpa-induk.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}

	db := serviceDB(t)
	cats := NewCategoryService(db, slog.Default())
	cat, _ := cats.Create(&models.Category{Code: "CV", Name: "Civil"})

	m := importers.ColumnMapping{NameCol: 0, QtyCol: -1, UnitCol: -1, ServiceCol: -1, MaterialCol: -1,
		UnitPriceCol: -1, ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	svc := NewImportService(db, slog.Default())
	g, _ := importers.ReadSheet(path, sheet)
	rows := importers.ParseRows(g, m)
	if rows[0].Suggested != importers.LevelSub {
		t.Fatalf("saran level = %q, mau sub", rows[0].Suggested)
	}
	_, err := svc.ImportWorkItems(cat.ID, path, sheet, m, confirmAll(rows))
	if err == nil {
		t.Fatal("sub tanpa induk harus gagal")
	}
	if !strings.Contains(err.Error(), "induk") {
		t.Errorf("pesan tidak menjelaskan masalah: %v", err)
	}
	var count int64
	db.Model(&models.WorkItem{}).Where("category_id = ?", cat.ID).Count(&count)
	if count != 0 {
		t.Errorf("rollback gagal: %d pekerjaan tersimpan", count)
	}
}

func TestImportKategoriTidakAda(t *testing.T) {
	db := serviceDB(t)
	svc := NewImportService(db, slog.Default())
	path := importFixtureXlsx(t)
	if _, err := svc.ImportWorkItems(9999, path, "SPH", importMapping(), nil); err == nil {
		t.Fatal("kategori tak dikenal seharusnya ditolak")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("fixture hilang: %v", err)
	}
}

// Harga dari kolom "Harga Satuan" umum harus mengalir sampai ke database.
func TestImportUnitPriceFallbackRoundtrip(t *testing.T) {
	f := excelize.NewFile()
	const sheet = "Data"
	if _, err := f.NewSheet(sheet); err != nil {
		t.Fatal(err)
	}
	f.DeleteSheet("Sheet1")
	rows := [][]interface{}{
		{"URAIAN KEGIATAN", "JML", "SAT", "", "HARGA SATUAN"},
		{"1. Kalibrasi Alat Ukur", 2, "giat", "", 7750000},
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				t.Fatal(err)
			}
		}
	}
	path := filepath.Join(t.TempDir(), "satuan.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}

	db := serviceDB(t)
	cats := NewCategoryService(db, slog.Default())
	cat, _ := cats.Create(&models.Category{Code: "KA", Name: "Kalibrasi"})

	m := importers.ColumnMapping{NameCol: 0, NameSpan: 1, QtyCol: 1, UnitCol: 2,
		ServiceCol: -1, MaterialCol: -1, UnitPriceCol: 4, UnitPriceAs: "service",
		ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	g, _ := importers.ReadSheet(path, sheet)
	parsed := importers.ParseRows(g, m)
	if len(parsed) != 1 || parsed[0].ServicePrice != 7_750_000 {
		t.Fatalf("pratinjau harga satuan salah: %+v", parsed)
	}

	svc := NewImportService(db, slog.Default())
	confirms := []ConfirmRow{{RowIndex: parsed[0].RowIndex, Level: importers.LevelMain}}
	if _, err := svc.ImportWorkItems(cat.ID, path, sheet, m, confirms); err != nil {
		t.Fatalf("import gagal: %v", err)
	}
	var item models.WorkItem
	if err := db.Where("category_id = ?", cat.ID).First(&item).Error; err != nil {
		t.Fatalf("query item: %v", err)
	}
	if item.DefaultServicePrice != 7_750_000 || item.DefaultQuantity != 2 {
		t.Errorf("nilai tersimpan salah: %+v", item)
	}
}
