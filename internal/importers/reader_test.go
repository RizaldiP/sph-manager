package importers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// writeFixtureXlsx membuat file .xlsx sementara berisi gaya A.
func writeFixtureXlsx(t *testing.T) string {
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
		{"2.", "Laksanakan Ganti Baru Sistem Penjalan :", "", "", "", "", ""},
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				t.Fatalf("set cell %s: %v", cell, err)
			}
		}
	}
	path := filepath.Join(t.TempDir(), "fixture.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("save as: %v", err)
	}
	return path
}

func TestListAndReadXLSX(t *testing.T) {
	path := writeFixtureXlsx(t)
	sheets, err := ListSheets(path)
	if err != nil {
		t.Fatalf("ListSheets gagal: %v", err)
	}
	if len(sheets) != 1 || sheets[0] != "SPH" {
		t.Fatalf("sheets = %v, mau [SPH]", sheets)
	}
	g, err := ReadSheet(path, "SPH")
	if err != nil {
		t.Fatalf("ReadSheet gagal: %v", err)
	}
	if g.Height() != 5 {
		t.Errorf("tinggi grid = %d, mau 5", g.Height())
	}
	if g.Cell(1, 1) != "Laksanakan Perbaikan Sistem Penjalan" {
		t.Errorf("isi sel salah: %q", g.Cell(1, 1))
	}
	// Angka dari XLSX harus bersih (tanpa notasi ilmiah).
	m := ColumnMapping{NameCol: 1, QtyCol: 2, UnitCol: 3, ServiceCol: 5, MaterialCol: -1,
		ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	parsed := ParseRows(g, m)
	var bearing *PreviewRow
	for i := range parsed {
		if parsed[i].Marker == "a" {
			bearing = &parsed[i]
		}
	}
	if bearing == nil {
		t.Fatal("baris sub 'a' tidak ditemukan")
	}
	if bearing.ServicePrice != 12_000_000 {
		t.Errorf("harga satuan = %d, mau 12000000", bearing.ServicePrice)
	}
	if _, err := ReadSheet(path, "Tidak Ada"); err == nil {
		t.Error("sheet tidak dikenal seharusnya error")
	}
}

func TestUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(csvPath, []byte("a,b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ListSheets(csvPath); err == nil {
		t.Error("csv seharusnya ditolak")
	}
}

// TestRealReferenceFile hanya jalan bila file referensi asli tersedia di root repo.
func TestRealReferenceFile(t *testing.T) {
	const ref = "../../SPH_KRI_OWA.xls"
	if _, err := os.Stat(ref); err != nil {
		t.Skip("file referensi SPH_KRI_OWA.xls tidak ada; lewati")
	}
	sheets, err := ListSheets(ref)
	if err != nil {
		t.Fatalf("ListSheets(.xls) gagal: %v", err)
	}
	if len(sheets) < 2 || sheets[0] != "SPH" || sheets[1] != "SPH #2" {
		t.Fatalf("sheets referensi = %v", sheets)
	}
	g, err := ReadSheet(ref, "SPH")
	if err != nil {
		t.Fatalf("ReadSheet(.xls) gagal: %v", err)
	}
	if g.Width() < 10 || g.Height() < 20 {
		t.Fatalf("grid terlalu kecil: %dx%d", g.Height(), g.Width())
	}
	mapping, notes := SuggestMapping(g)
	// Layout referensi: NO=1, URAIAN=2, JML=5, SAT=6, HARGA SATUAN=7, JASA=8, MAT=9.
	if mapping.NameCol != 2 || mapping.QtyCol != 5 || mapping.UnitCol != 6 ||
		mapping.ServiceCol != 8 || mapping.MaterialCol != 9 {
		t.Errorf("saran mapping salah: %+v (notes=%v)", mapping, notes)
	}
	if mapping.UnitPriceCol != 7 {
		t.Errorf("kolom harga satuan = %d, mau 7", mapping.UnitPriceCol)
	}
	if mapping.HeaderRows != 10 {
		t.Errorf("header rows = %d, mau 10", mapping.HeaderRows)
	}
	parsed := ParseRows(g, mapping)
	mainCount, subCount, unknownCount := CountLevels(parsed)
	if mainCount+subCount == 0 {
		t.Error("tidak ada baris yang berhasil diklasifikasi")
	}

	// Harga kolom HARGA SATUAN harus masuk hasil analisis (regresi nyata).
	priced := 0
	var sensorOli *PreviewRow
	for i := range parsed {
		if parsed[i].ServicePrice > 0 || parsed[i].MaterialPrice > 0 {
			priced++
		}
		if strings.Contains(parsed[i].Name, "Sensor Oli") {
			sensorOli = &parsed[i]
		}
	}
	if priced == 0 {
		t.Error("tidak ada baris berharga terdeteksi")
	}
	if sensorOli == nil {
		t.Fatal("baris 'Sensor Oli' tidak ditemukan")
	}
	if sensorOli.ServicePrice != 2_200_000 {
		t.Errorf("harga Sensor Oli = %d, mau 2200000", sensorOli.ServicePrice)
	}
	t.Logf("referensi: main=%d sub=%d unknown=%d berharga=%d notes=%v",
		mainCount, subCount, unknownCount, priced, notes)
}
