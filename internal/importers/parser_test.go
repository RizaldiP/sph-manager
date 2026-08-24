package importers

import (
	"strings"
	"testing"
)

func TestSplitMarkerKinds(t *testing.T) {
	cases := []struct {
		text string
		kind markerKind
		name string
	}{
		{"12. Laksanakan Perbaikan", mkNumber, "Laksanakan Perbaikan"},
		{"3) Bongkar Pasang", mkNumber, "Bongkar Pasang"},
		{"I   Laksanakan Service AHU", mkRoman, "Laksanakan Service AHU"},
		{"XII. Service Kelas B Motor", mkRoman, "Service Kelas B Motor"},
		{"A  Laksanakan Perbaikan Sistem Penjalan", mkUpper, "Laksanakan Perbaikan Sistem Penjalan"},
		{"a. Ganti Bearing", mkLower, "Ganti Bearing"},
		{"b) Cleaning Filter", mkLower, "Cleaning Filter"},
		{"Service Kelas B Motor Blower", mkNone, "Service Kelas B Motor Blower"},
		{"Bongkar pasang dan Setting Program", mkNone, "Bongkar pasang dan Setting Program"},
	}
	for _, c := range cases {
		kind, _, name := splitMarker(c.text)
		if kind != c.kind {
			t.Errorf("splitMarker(%q) kind = %v, mau %v", c.text, kind, c.kind)
		}
		if name != c.name {
			t.Errorf("splitMarker(%q) name = %q, mau %q", c.text, name, c.name)
		}
	}
}

func grid(rows ...[]string) *Grid { return &Grid{Rows: rows} }

func TestParseRowsGayaA(t *testing.T) {
	g := grid(
		[]string{"NO", "URAIAN KEGIATAN", "JML", "SAT", "", "JASA", "MAT"},
		[]string{"A", "Laksanakan Perbaikan Sistem Penjalan", "", "", "", "", ""},
		[]string{"1.", "Laksanakan Perbaikan Sistem Penjalan meliputi :", "", "", "", "", ""},
		[]string{"", "a. Ganti Bearing Kompressor", "2", "bh", "", "24000000", ""},
		[]string{"", "b. Ganti Seal Pump", "1", "bh", "", "", "3500000"},
		[]string{"2.", "Laksanakan Ganti Baru Sistem Penjalan :", "", "", "", "", ""},
		[]string{"", "a. Pipa Inlet", "9", "m", "", "4500000", ""},
		[]string{"1.", "Pipa Outlet (sub-sub diratakan jadi sub)", "3", "unit", "", "6000000", ""},
	)
	m := ColumnMapping{NameCol: 1, QtyCol: 2, UnitCol: 3, ServiceCol: 5, MaterialCol: 6,
		ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	rows := ParseRows(g, m)
	mainCount, subCount, unknownCount := CountLevels(rows)
	if mainCount != 3 || subCount != 4 || unknownCount != 0 {
		t.Fatalf("level salah: main=%d sub=%d unknown=%d", mainCount, subCount, unknownCount)
	}

	// Urutan & konteks.
	wantLevels := []string{LevelMain, LevelMain, LevelSub, LevelSub, LevelMain, LevelSub, LevelSub}
	for i, w := range wantLevels {
		if rows[i].Suggested != w {
			t.Errorf("baris %d level = %q, mau %q (%q)", i, rows[i].Suggested, w, rows[i].Name)
		}
	}

	// Konversi nilai total ke harga satuan: 24.000.000 / 2 = 12.000.000.
	sub1 := rows[2]
	if sub1.ServicePrice != 12_000_000 || sub1.MaterialPrice != 0 || sub1.Qty != 2 {
		t.Errorf("konversi harga salah: %+v", sub1)
	}
	sub2 := rows[3]
	if sub2.MaterialPrice != 3_500_000 {
		t.Errorf("material satuan salah: %d", sub2.MaterialPrice)
	}
	if !strings.HasSuffix(sub1.Name, "Kompressor") || sub1.Marker != "a" {
		t.Errorf("marker/nama salah: marker=%q name=%q", sub1.Marker, sub1.Name)
	}
}

func TestParseRowsGayaB(t *testing.T) {
	g := grid(
		[]string{"NO", "URAIAN KEGIATAN", "JML", "SAT", "", "JASA", "MAT"},
		[]string{"I", "Laksanakan Perbaikan dan Service AHU III", "", "", "", "", ""},
		[]string{"", "a. Cleaning Fan Blade", "1", "giat", "", "1500000", ""},
		[]string{"II", "Laksanakan Service dan Perbaikan Drain Pan", "", "", "", "", ""},
		[]string{"", "a. Bersihkan Drain Pan", "1", "giat", "", "750000", ""},
		[]string{"", "b. Ganti Selang Pembuangan", "2", "bh", "", "", "600000"},
	)
	m := ColumnMapping{NameCol: 1, QtyCol: 2, UnitCol: 3, ServiceCol: 5, MaterialCol: 6,
		ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	rows := ParseRows(g, m)
	mainCount, subCount, unknownCount := CountLevels(rows)
	if mainCount != 2 || subCount != 3 || unknownCount != 0 {
		t.Fatalf("level gaya B salah: main=%d sub=%d unknown=%d", mainCount, subCount, unknownCount)
	}
	if rows[0].Marker != "I" || rows[2].Marker != "II" {
		t.Errorf("marker romawi salah: %q %q", rows[0].Marker, rows[2].Marker)
	}
}

func TestParseRowsUnknownDanValidasi(t *testing.T) {
	g := grid(
		[]string{"URAIAN KEGIATAN", "JML"},
		[]string{"Tanpa penanda sekaligus tanpa harga", "", ""},
		[]string{"Qty Nol", "0", ""},
	)
	m := ColumnMapping{NameCol: 0, QtyCol: 1, UnitCol: -1, ServiceCol: -1, MaterialCol: -1,
		ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	rows := ParseRows(g, m)
	if len(rows) != 2 {
		t.Fatalf("jumlah baris = %d, mau 2", len(rows))
	}
	if rows[0].Suggested != LevelUnknown {
		t.Errorf("baris tanpa penanda harus unknown, dapat %q", rows[0].Suggested)
	}
	if len(rows[1].Errors) == 0 {
		t.Error("qty nol harus menghasilkan error")
	}
}

// Fallback kolom Harga Satuan: dipakai hanya bila JASA & MATERIAL kosong;
// prioritas tetap pada nilai total jasa/material; arah jasa/material bisa diatur.
func TestParseRowsUnitPriceFallback(t *testing.T) {
	g := grid(
		[]string{"URAIAN KEGIATAN", "JML", "SAT", "", "HARGA SATUAN", "JASA"},
		[]string{"a. Item Satuan Murni", "2", "bh", "", "7750000", ""},
		[]string{"b. Total Jasa Menang", "3", "bh", "", "9999999", "3000000"},
	)
	base := ColumnMapping{NameCol: 0, NameSpan: 1, QtyCol: 1, UnitCol: 2, ServiceCol: 5, MaterialCol: -1,
		ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}

	svcMap := base
	svcMap.UnitPriceCol = 4
	svcMap.UnitPriceAs = "service"
	rows := ParseRows(g, svcMap)
	if rows[0].ServicePrice != 7_750_000 || rows[0].MaterialPrice != 0 {
		t.Errorf("fallback service salah: %+v", rows[0])
	}
	// Nilai total jasa (3.000.000/3 = 1.000.000) menang atas harga satuan.
	if rows[1].ServicePrice != 1_000_000 {
		t.Errorf("prioritas total jasa salah: %+v", rows[1])
	}

	matMap := base
	matMap.UnitPriceCol = 4
	matMap.UnitPriceAs = "material"
	rows = ParseRows(g, matMap)
	if rows[0].MaterialPrice != 7_750_000 || rows[0].ServicePrice != 0 {
		t.Errorf("fallback material salah: %+v", rows[0])
	}

	// Harga satuan tidak valid → error baris.
	bad := grid(
		[]string{"URAIAN KEGIATAN", "HARGA SATUAN"},
		[]string{"a. Harga Rusak", "abc"},
	)
	badMap := base
	badMap.QtyCol = -1
	badMap.UnitPriceCol = 1
	rows = ParseRows(bad, badMap)
	found := false
	for _, e := range rows[0].Errors {
		if strings.Contains(e, "Harga satuan tidak valid") {
			found = true
		}
	}
	if !found {
		t.Errorf("harga satuan invalid harus error: %+v", rows[0].Errors)
	}
}

func TestParseNumber(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"120725000", 120725000},
		{"1.234.567", 1234567},
		{"1,234,567", 1234567},
		{"1.234.567,89", 1234567.89},
		{"1,234,567.89", 1234567.89},
		{" 250000 ", 250000},
	}
	for _, c := range cases {
		got, err := parseNumber(c.in)
		if err != nil || got != c.want {
			t.Errorf("parseNumber(%q) = %v,%v mau %v", c.in, got, err, c.want)
		}
	}
}

func TestSuggestMappingReferensi(t *testing.T) {
	g := grid(
		[]string{"PT. GANESHA ENERGI INDONESIA"},
		[]string{"NO", "URAIAN KEGIATAN", "JML", "SAT", "HARGA SATUAN", "JASA", "MAT"},
		[]string{"1", "Perbaikan sistem kontrol", "1", "giat", "11000000", "11000000", ""},
	)
	m, notes := SuggestMapping(g)
	if m.NameCol != 1 || m.QtyCol != 2 || m.UnitCol != 3 || m.ServiceCol != 5 || m.MaterialCol != 6 {
		t.Errorf("saran mapping salah: %+v (notes %v)", m, notes)
	}
	if m.UnitPriceCol != 4 {
		t.Errorf("kolom harga satuan = %d, mau 4", m.UnitPriceCol)
	}
	if m.HeaderRows != 2 {
		t.Errorf("header rows = %d, mau 2", m.HeaderRows)
	}
	if !m.ServiceTotal || !m.MaterialTotal {
		t.Error("default mode nilai total harus aktif")
	}
}

// Header dua baris ala file referensi: "HARGA" terbentang di atas "SATUAN",
// sedangkan SAT tetap kolom satuan. Pastikan keduanya tidak tertukar.
func TestSuggestMappingBandHargaSatuan(t *testing.T) {
	g := grid(
		[]string{"NO", "URAIAN KEGIATAN", "", "", "JML", "SAT", "HARGA", "PENAWARAN HARGA (RP)", "", ""},
		[]string{"", "", "", "", "", "", "SATUAN", "JASA", "MAT", "JML"},
	)
	m, notes := SuggestMapping(g)
	if m.NameCol != 1 || m.QtyCol != 4 || m.ServiceCol != 7 || m.MaterialCol != 8 {
		t.Errorf("mapping band salah: %+v (notes %v)", m, notes)
	}
	if m.UnitCol != 5 {
		t.Errorf("kolom satuan = %d, mau 5", m.UnitCol)
	}
	if m.UnitPriceCol != 6 {
		t.Errorf("kolom harga satuan = %d, mau 6", m.UnitPriceCol)
	}
}
