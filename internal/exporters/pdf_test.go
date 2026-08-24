package exporters

import (
	"bytes"
	"strings"
	"testing"
)

func TestPaginate(t *testing.T) {
	heights := []float64{10, 10, 10, 10, 10}
	pages := paginate(heights, 25, 20)
	got := make([]int, 0, len(pages))
	for _, p := range pages {
		got = append(got, len(p))
	}
	// halaman pertama muat 2 (25), sisanya 2 per halaman (20)
	want := []int{2, 2, 1}
	if len(pages) != len(want) {
		t.Fatalf("jumlah halaman = %d, mau %d (%v)", len(pages), len(want), pages)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("halaman %d = %d baris, mau %d", i+1, got[i], want[i])
		}
	}
}

func TestPaginateSingleTallRowStillPlaced(t *testing.T) {
	pages := paginate([]float64{500}, 100, 100)
	if len(pages) != 1 || len(pages[0]) != 1 {
		t.Errorf("baris raksasa harus tetap di satu halaman: %v", pages)
	}
}

func TestComputeCarry(t *testing.T) {
	rows := []Row{
		{ServiceTotal: 100, MaterialTotal: 10, Total: 110},
		{ServiceTotal: 200, MaterialTotal: 20, Total: 220},
		{ServiceTotal: 400, MaterialTotal: 40, Total: 440},
	}
	carry := computeCarry([][]int{{0, 1}, {2}}, rows)
	if carry[0] != [3]int64{0, 0, 0} {
		t.Errorf("carry halaman 1 harus nol: %v", carry[0])
	}
	if carry[1] != [3]int64{300, 30, 330} {
		t.Errorf("carry halaman 2 salah: %v (mau [300 30 330])", carry[1])
	}
}

func TestBuildPDFBasicShape(t *testing.T) {
	res, err := BuildPDF(buildFixture(), PDFOptions{Landscape: true})
	if err != nil {
		t.Fatalf("BuildPDF gagal: %v", err)
	}
	if res.Pages < 1 {
		t.Errorf("halaman = %d", res.Pages)
	}
	if !bytes.HasPrefix(res.Bytes, []byte("%PDF-")) {
		t.Errorf("bukan PDF valid: awal = %q", res.Bytes[:min(40, len(res.Bytes))])
	}
	if !bytes.Contains(res.Bytes, []byte("%%EOF")) {
		t.Error("PDF tanpa penanda akhir")
	}
}

func TestBuildPDFMultiPageBothOrientations(t *testing.T) {
	base := fixtureSphDoc()
	d := BuildData(base, sampleCompany())
	// perbanyak baris sampai pasti multi-halaman
	extra := d.Rows
	d.Rows = nil
	for i := 0; i < 60; i++ {
		d.Rows = append(d.Rows, extra...)
	}
	for _, landscape := range []bool{true, false} {
		res, err := BuildPDF(d, PDFOptions{Landscape: landscape})
		if err != nil {
			t.Fatalf("landscape=%v gagal: %v", landscape, err)
		}
		if res.Pages <= 1 {
			t.Errorf("landscape=%v harus multi-halaman, dapat %d", landscape, res.Pages)
		}
		if !strings.HasPrefix(string(res.Bytes[:5]), "%PDF-") {
			t.Error("output rusak")
		}
	}
}

func TestIDNumberEdge(t *testing.T) {
	if idNumber(7) != "7" || idNumber(70) != "70" {
		t.Error("angka kecil salah format")
	}
}
