package exporters

import (
	"strings"
	"testing"
	"time"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// fixtureSphDoc menyusun dokumen contoh: 1 main point + 2 sub + 1 main polos.
func fixtureSphDoc() *models.SphDocument {
	date := time.Date(2026, 7, 13, 0, 0, 0, 0, time.Local)
	valid := date.AddDate(0, 1, 0)
	doc := &models.SphDocument{
		DocumentNumber:   "SPH/GEI/VII/2026/001",
		Revision:         2,
		Date:             date,
		ValidUntil:       &valid,
		ProjectName:      "Docking KM Bahari",
		Subject:          "Penawaran Repair PLC",
		Reference:        "Email PO 12/07",
		Location:         "Surabaya",
		PicName:          "Budi",
		Notes:            "Harga berlaku 30 hari.",
		Terbilang:        "Sebelas Juta Dua Ratus Lima Puluh Ribu Rupiah",
		SubtotalService:  10_500_000,
		SubtotalMaterial: 1_000_000,
		GrandTotal:       11_500_000,
		Customer:         models.Customer{Code: "CUS-1", Name: "PT Laut Biru"},
		Vessel:           &models.Vessel{Code: "KPL-1", Name: "KM Bahari", VesselNumber: "123/IV"},
	}
	doc.Items = []models.SphItem{
		{
			Sequence: 1, NameSnapshot: "Repair PLC", DescriptionSnapshot: "Panel utama",
			Quantity: 2, Unit: "Set",
			ServiceUnitPrice: 5_000_000, MaterialUnitPrice: 500_000,
			ServiceTotal: 10_250_000, MaterialTotal: 1_000_000, Total: 11_250_000,
			PricingMode: models.PricingModeWeight,
			SubItems: []models.SphSubItem{
				{
					NameSnapshot: "Inspection", Quantity: 1, Unit: "ls",
					ServiceUnitPrice: 250_000, ServiceTotal: 250_000, Total: 250_000, Weight: 40,
				},
				{
					NameSnapshot: "Rewiring", Quantity: 1, Unit: "ls",
					MaterialTotal: 1_000_000, Total: 1_000_000, Weight: 60,
				},
			},
		},
		{
			Sequence: 2, NameSnapshot: "Wiring Check", Quantity: 1, Unit: "giat",
			ServiceTotal: 250_000, Total: 250_000,
		},
	}
	return doc
}

func sampleCompany() CompanyInfo {
	return CompanyInfo{
		Name: "PT. Ganesha Energi Indonesia", City: "Surabaya",
		Address: "Jl. Contoh No. 1", SignerName: "Matawai", SignerPosition: "Direktur",
	}
}

func TestBuildDataFlattensRows(t *testing.T) {
	d := BuildData(fixtureSphDoc(), sampleCompany())
	if len(d.Rows) != 4 { // main + 2 sub + main
		t.Fatalf("baris harus 4, dapat %d", len(d.Rows))
	}
	main := d.Rows[0]
	if !main.Bold || main.No != "1" {
		t.Errorf("main point salah: %+v", main)
	}
	if main.Total != 11_250_000 || main.ServiceTotal != 10_250_000 {
		t.Errorf("roll-up main point salah: %+v", main)
	}
	sub := d.Rows[1]
	if sub.Bold || sub.No != "" || sub.SubNo != "a" {
		t.Errorf("sub point tidak boleh tebal/bernomor, harus huruf 'a': %+v", sub)
	}
	if d.Rows[2].SubNo != "b" {
		t.Errorf("sub point kedua harus huruf 'b': %+v", d.Rows[2])
	}
	if d.Rows[3].SubNo != "" {
		t.Errorf("main point tidak boleh punya huruf sub: %+v", d.Rows[3])
	}
	if sub.WeightNote != "[bobot 40%]" {
		t.Errorf("catatan bobot salah: %q", sub.WeightNote)
	}
	if d.Document.Customer != "PT Laut Biru" || d.Document.VesselNumber != "123/IV" {
		t.Errorf("info dokumen kurang lengkap: %+v", d.Document)
	}
	if d.GrandTotal != 11_500_000 || d.SubtotalMaterial != 1_000_000 {
		t.Errorf("rekap salah: %+v", d)
	}
}

func TestFormatDateID(t *testing.T) {
	ts := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	if got := FormatDateID(ts); got != "13 Juli 2026" {
		t.Errorf("FormatDateID = %q", got)
	}
	if FormatDateID(time.Time{}) != "" {
		t.Error("tanggal nol harus string kosong")
	}
}

func TestTitleLines(t *testing.T) {
	d := BuildData(fixtureSphDoc(), sampleCompany())
	if l1 := d.TitleLine1(); l1 != "Penawaran Repair PLC" {
		t.Errorf("judul 1 = %q", l1)
	}
	if l2 := d.TitleLine2(); l2 != "KM Bahari TA.2026" {
		t.Errorf("judul 2 = %q", l2)
	}
}

func TestWrappedLinesAndNumbers(t *testing.T) {
	if n := wrappedLines("", 10); n != 0 {
		t.Errorf("teks kosong = %d baris", n)
	}
	if n := wrappedLines(strings.Repeat("a", 25), 10); n != 3 {
		t.Errorf("25 karakter @10 = %d baris", n)
	}
	cases := map[int64]string{
		120725000:   "120.725.000",
		1000:        "1.000",
		999:         "999",
		-1500:       "-1.500",
		11000000:    "11.000.000",
		0:           "0",
		12345678901: "12.345.678.901",
	}
	for in, want := range cases {
		if got := idNumber(in); got != want {
			t.Errorf("idNumber(%d) = %q, mau %q", in, got, want)
		}
	}
	if qtyText(2) != "2" || qtyText(1.5) != "1,5" || qtyText(0.25) != "0,25" || qtyText(2.50) != "2,5" {
		t.Error("qtyText salah")
	}
}

// TestQtyUnitSuffix memeriksa format akhiran "(qty satuan)" sub point gaya uraian.
func TestQtyUnitSuffix(t *testing.T) {
	cases := []struct {
		q    float64
		unit string
		want string
	}{
		{1, "giat", " (1 giat)"},
		{2, "set", " (2 set)"},
		{1, "", " (1)"},
		{1.5, "ls", " (1,5 ls)"},
		{0, "giat", ""},
		{-1, "giat", ""},
	}
	for _, c := range cases {
		if got := qtyUnitSuffix(c.q, c.unit); got != c.want {
			t.Errorf("qtyUnitSuffix(%v,%q) = %q, mau %q", c.q, c.unit, got, c.want)
		}
	}
}

// TestMarkEmptyNumberRows memeriksa aturan per-blok: main point terisi ->
// sub point EmptyNumbers; main point tanpa nilai -> main point EmptyNumbers.
func TestMarkEmptyNumberRows(t *testing.T) {
	// blok 1: main terisi (qty>0 & price>0) → sub point EmptyNumbers
	// blok 2: main tanpa nilai → main EmptyNumbers, sub normal
	rows := []Row{
		{No: "1", Name: "Main Terisi", Bold: true, Qty: 1, UnitPrice: 100},
		{SubNo: "a", Name: "sub1", Bold: false, Qty: 1, UnitPrice: 100},
		{SubNo: "b", Name: "sub2", Bold: false, Qty: 2, UnitPrice: 50},
		{No: "2", Name: "Main Kosong", Bold: true, Qty: 1, UnitPrice: 0},
		{SubNo: "a", Name: "sub3", Bold: false, Qty: 1, UnitPrice: 300},
	}
	markEmptyNumberRows(rows)

	if rows[0].EmptyNumbers {
		t.Error("main point terisi tidak boleh EmptyNumbers")
	}
	if !rows[1].EmptyNumbers || !rows[2].EmptyNumbers {
		t.Errorf("sub point pada main terisi harus EmptyNumbers: %+v %+v", rows[1], rows[2])
	}
	if !rows[3].EmptyNumbers {
		t.Error("main point tanpa nilai harus EmptyNumbers")
	}
	if rows[4].EmptyNumbers {
		t.Error("sub point pada main tanpa nilai tidak boleh EmptyNumbers")
	}
}

// TestEmptyNumbersAppliedViaBuildData: fixture memicu kedua blok terisi/tdk.
func TestEmptyNumbersAppliedViaBuildData(t *testing.T) {
	d := BuildData(fixtureSphDoc(), sampleCompany())
	// main 1: qty=2 & unitprice=5.500.000 → terisi; sub a & b jadi EmptyNumbers
	if d.Rows[0].EmptyNumbers {
		t.Error("main 1 (terisi) tidak boleh EmptyNumbers")
	}
	if !d.Rows[1].EmptyNumbers || !d.Rows[2].EmptyNumbers {
		t.Errorf("sub a/b main1 harus EmptyNumbers: %+v %+v", d.Rows[1], d.Rows[2])
	}
	// main 2: Wiring Check qty=1 unitprice=0 → tidak terisi → EmptyNumbers
	if !d.Rows[3].EmptyNumbers {
		t.Error("main 2 (tanpa harga) harus EmptyNumbers")
	}
}
