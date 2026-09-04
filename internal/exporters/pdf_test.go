package exporters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-pdf/fpdf"
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

// wrapFixture membuat fpdf siap pakai untuk menguji helper wrap.
func wrapFixture(t *testing.T) (*fpdf.Fpdf, func(string) string) {
	t.Helper()
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Helvetica", "", 9)
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	return pdf, tr
}

func TestWrapTextLinesRespectsWidth(t *testing.T) {
	pdf, tr := wrapFixture(t)
	text := tr(strings.TrimSpace(strings.Repeat("kata ", 40)))
	lines := wrapTextLines(pdf, text, 40)
	if len(lines) < 2 {
		t.Fatalf("teks panjang harus ter-wrap ke beberapa baris, dapat %d", len(lines))
	}
	for i, ln := range lines {
		if strings.Contains(ln, " ") && pdf.GetStringWidth(ln) > 40 {
			t.Errorf("baris %d melebihi lebar: %q = %.1fmm", i, ln, pdf.GetStringWidth(ln))
		}
	}
	// gabungan harus tetap = jumlah kata semula
	if got := strings.Join(strings.Fields(strings.Join(lines, " ")), " "); got != text {
		t.Errorf("wrap mengubah isi teks:\ngot  %q\nwant %q", got, text)
	}
}

func TestWrapTextLinesHardBreakAndEmpty(t *testing.T) {
	pdf, tr := wrapFixture(t)
	got := wrapTextLines(pdf, tr("Baris A\nBaris B"), 200)
	if len(got) != 2 || got[0] != "Baris A" || got[1] != "Baris B" {
		t.Errorf("hard break salah: %q", got)
	}
	if e := wrapTextLines(pdf, tr(""), 200); len(e) != 1 || e[0] != "" {
		t.Errorf("teks kosong harus satu baris kosong: %q", e)
	}
}

func TestSubLetter(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "a"}, {1, "b"}, {24, "y"}, {25, "z"},
		{26, "aa"}, {27, "ab"}, {51, "az"},
	}
	for _, c := range cases {
		if got := subLetter(c.in); got != c.want {
			t.Errorf("subLetter(%d) = %q, mau %q", c.in, got, c.want)
		}
	}
	if subLetter(-1) != "" {
		t.Error("indeks negatif harus string kosong")
	}
}

func TestWrapUraianSubPrefix(t *testing.T) {
	pdf, tr := wrapFixture(t)
	g := newGeom(true)
	sub := &Row{SubNo: "a", Name: "Bongkar pasang separator oli lama", Description: strings.Repeat("deskripsi lanjutan yang cukup panjang ", 8)}
	lines := wrapUraian(pdf, tr, g, sub)
	if len(lines) == 0 {
		t.Fatalf("wrapUraian menghasilkan 0 baris")
	}
	// prefix "a." digambar terpisah oleh drawRow; wrapUraian hanya menghitung
	// baris teks tanpa prefix agar tidak terjadi overflow atau penomoran ganda.
	prefixW := pdf.GetStringWidth("a. ")
	avail := g.wUraian - 2*pdfPad - indentOf(sub) - prefixW
	for i, ln := range lines {
		if strings.Contains(ln, " ") && pdf.GetStringWidth(ln) > avail+0.01 {
			t.Errorf("baris %d melebihi lebar valid: %q = %.1fmm (maks %.1f)", i, ln, pdf.GetStringWidth(ln), avail)
		}
	}
	if !strings.Contains(lines[0], "Bongkar") {
		t.Error("isi teks sub point hilang")
	}
	if strings.TrimSpace(lines[0]) == "" {
		t.Error("nama sub point kosong")
	}
	// main point tidak boleh dikurangi lebarnya (prefix = 0)
	main := &Row{SubNo: "", Name: "Perbaikan Separator Oli"}
	if l := wrapUraian(pdf, tr, g, main); l[0] != "Perbaikan Separator Oli" {
		t.Errorf("main point berubah: %q", l)
	}
}

// TestUraianTextQtyUnitSuffix: sub point EmptyNumbers menampilkan "(qty satuan)",
// main point dan sub point normal tetap tanpa akhiran tersebut.
func TestUraianTextQtyUnitSuffix(t *testing.T) {
	sub := &Row{SubNo: "a", Name: "Bongkar pasang", EmptyNumbers: true, Qty: 1, Unit: "giat"}
	got := uraianText(sub)
	if got != "Bongkar pasang (1 giat)" {
		t.Errorf("sub EmptyNumbers uraian = %q", got)
	}

	// qty 0 -> tanpa akhiran
	zero := &Row{SubNo: "a", Name: "Bongkar pasang", EmptyNumbers: true, Qty: 0, Unit: "giat"}
	if got := uraianText(zero); got != "Bongkar pasang" {
		t.Errorf("sub qty 0 uraian = %q", got)
	}

	// sub normal (bukan EmptyNumbers) tidak boleh dapat akhiran
	normal := &Row{SubNo: "a", Name: "Bongkar pasang", Qty: 1, Unit: "giat"}
	if got := uraianText(normal); got != "Bongkar pasang" {
		t.Errorf("sub normal uraian = %q", got)
	}

	// main point tidak boleh dapat akhiran
	main := &Row{No: "1", Name: "Perbaikan", EmptyNumbers: true, Qty: 1, Unit: "unit"}
	if got := uraianText(main); got != "Perbaikan" {
		t.Errorf("main EmptyNumbers uraian = %q", got)
	}

	// deskripsi tetap menyusul di baris berikutnya
	withDesc := &Row{SubNo: "a", Name: "Bongkar", EmptyNumbers: true, Qty: 2, Unit: "ls", Description: "rincian pekerjaan"}
	if got := uraianText(withDesc); got != "Bongkar (2 ls)\nrincian pekerjaan" {
		t.Errorf("sub + desc uraian = %q", got)
	}
}

func TestBuildPDFLongDescriptionTolerated(t *testing.T) {
	base := fixtureSphDoc()
	d := BuildData(base, sampleCompany())
	long := strings.Repeat("deskripsi panjang tanpa spasi di dalam satu kata ", 60)
	for i := range d.Rows {
		d.Rows[i].Description = long
	}
	res, err := BuildPDF(d, PDFOptions{Landscape: true})
	if err != nil {
		t.Fatalf("BuildPDF deskripsi panjang gagal: %v", err)
	}
	if res.Pages < 1 || !bytes.HasPrefix(res.Bytes, []byte("%PDF-")) {
		t.Errorf("output rusak: %d halaman", res.Pages)
	}
}

func BenchmarkBuildPDFRows(b *testing.B) {
	d := BuildData(fixtureSphDoc(), sampleCompany())
	var rows []Row
	for i := 0; i < 250; i++ {
		rows = append(rows, d.Rows...)
	}
	d.Rows = rows
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := BuildPDF(d, PDFOptions{Landscape: true}); err != nil {
			b.Fatal(err)
		}
	}
}

// TestBuildPDFTableStaysWithinPaper: setiap halaman, baris tabel tidak boleh
// melewati batas bawah pageH-pdfMargin (zona footer di pageH-9).
func TestBuildPDFTableStaysWithinPaper(t *testing.T) {
	for _, landscape := range []bool{false, true} {
		d := BuildData(fixtureSphDoc(), sampleCompany())
		extra := d.Rows
		d.Rows = nil
		for i := 0; i < 60; i++ {
			d.Rows = append(d.Rows, extra...)
		}

		pdf := fpdf.New("P", "mm", "A4", "")
		tr := pdf.UnicodeTranslatorFromDescriptor("")
		g := newGeom(landscape)
		kopH := kopHeight(pdf, tr, g, d)
		headH := tableHeadHeight()

		availFirst := g.usableH - kopH - headH - 2
		availRest := g.usableH - headH - pindahanHeight() - 2
		heights := make([]float64, len(d.Rows))
		for i := range d.Rows {
			heights[i] = rowHeightLines(len(wrapUraian(pdf, tr, g, &d.Rows[i])))
		}
		pages := paginate(heights, availFirst, availRest)

		if len(pages) <= 1 {
			t.Fatalf("landscape=%v harus multi-halaman", landscape)
		}
		limit := g.pageH - pdfMargin
		for pi, page := range pages {
			start := pdfMargin + 2 + headH + pindahanHeight()
			if pi == 0 {
				start = pdfMargin + kopH + 2 + headH
			}
			h := 0.0
			for _, ri := range page {
				h += heights[ri]
			}
			bottom := start + h
			if bottom > limit+0.01 {
				t.Errorf("landscape=%v halaman %d bottom=%.1fmm melewati batas %.1fmm", landscape, pi+1, bottom, limit)
			}
			if pi == 0 && len(page) == 0 {
				t.Error("halaman 1 tidak memuat baris")
			}
		}
	}
}

// TestRekapMovesToNewPageWhenFull: jika baris terakhir berakhir di batas bawah
// tabel (halaman penuh), rekap+ttd harus dipindah agar tidak menyentuh footer.
func TestRekapMovesToNewPageWhenFull(t *testing.T) {
	pdf, tr := wrapFixture(t)
	d := BuildData(fixtureSphDoc(), sampleCompany())
	for _, landscape := range []bool{false, true} {
		g := newGeom(landscape)
		need := rekapHeight(pdf, tr, g, d) + signatureHeight(pdf, tr, g, d) + terbilangNotesHeight(pdf, tr, g, d) + 11
		left := g.pageH - pdfFooterZone - (g.pageH - pdfMargin)
		if left >= need {
			t.Errorf("landscape=%v rekap harus pindah halaman saat penuh: sisa %.1f need %.1f", landscape, left, need)
		}
	}
}

// TestPageBodyStart memastikan posisi awal badan tabel konsisten dgn BuildPDF.
func TestPageBodyStart(t *testing.T) {
	g := newGeom(true)
	kopH := 40.0
	headH := tableHeadHeight()
	if got := pageBodyStart(g, kopH, 0); got != pdfMargin+2+headH+kopH {
		t.Errorf("halaman 1 body start = %.1f", got)
	}
	if got := pageBodyStart(g, kopH, 1); got != pdfMargin+2+headH+pindahanHeight() {
		t.Errorf("halaman 2+ body start = %.1f", got)
	}
}

// TestMoveTailRows memindahkan hingga n baris terakhir ke halaman penutup baru.
func TestMoveTailRows(t *testing.T) {
	pages := [][]int{{0, 1, 2, 3, 4, 5}}
	moved := moveTailRows(pages, 3)
	if len(moved) != 2 {
		t.Fatalf("harus jadi 2 halaman, dapat %d: %v", len(moved), moved)
	}
	if len(moved[0]) != 3 || len(moved[1]) != 3 {
		t.Fatalf("pembagian salah: %v", moved)
	}
	if moved[1][0] != 3 || moved[1][2] != 5 {
		t.Errorf("baris pindah salah: %v", moved[1])
	}

	// bila isi halaman terakhir < n dan habis dipindah, halaman kosong dibuang
	small := [][]int{{0, 1, 2}}
	r := moveTailRows(small, 3)
	if len(r) != 1 || len(r[0]) != 3 {
		t.Fatalf("halaman kosong harus dibuang: %v", r)
	}

	more := [][]int{{0, 1}, {2, 3, 4, 5}}
	r2 := moveTailRows(more, 3)
	if len(r2) != 3 || len(r2[1]) != 1 || len(r2[2]) != 3 {
		t.Fatalf("pindah dari halaman tengah salah: %v", r2)
	}
	if r2[2][0] != 3 || r2[2][2] != 5 {
		t.Errorf("isi halaman penutup salah: %v", r2[2])
	}
}

// TestBackfillKeepsLastPageNonEmpty: bila rekap terpaksa pindah halaman karena
// halaman data terakhir penuh, backfill harus memindahkan beberapa baris uraian
// ke halaman penutup agar lembar terakhir tidak kosong.
func TestBackfillKeepsLastPageNonEmpty(t *testing.T) {
	for _, landscape := range []bool{false, true} {
		d := BuildData(fixtureSphDoc(), sampleCompany())
		// deskripsi panjang agar baris tinggi -> halaman cepat penuh sehingga
		// rekap tak muat di halaman data terakhir (memicu backfill).
		for i := range d.Rows {
			d.Rows[i].Description = strings.Repeat("deskripsi panjang kata demi kata yang cukup banyak ", 12)
		}
		extra := d.Rows
		d.Rows = nil
		for i := 0; i < 40; i++ {
			d.Rows = append(d.Rows, extra...)
		}

		pdf := fpdf.New("P", "mm", "A4", "")
		tr := pdf.UnicodeTranslatorFromDescriptor("")
		g := newGeom(landscape)
		kopH := kopHeight(pdf, tr, g, d)
		headH := tableHeadHeight()
		availFirst := g.usableH - kopH - headH - 2
		availRest := g.usableH - headH - pindahanHeight() - 2
		heights := make([]float64, len(d.Rows))
		for i := range d.Rows {
			heights[i] = rowHeightLines(len(wrapUraian(pdf, tr, g, &d.Rows[i])))
		}
		pages := paginate(heights, availFirst, availRest)

		need := rekapHeight(pdf, tr, g, d) + signatureHeight(pdf, tr, g, d) + terbilangNotesHeight(pdf, tr, g, d) + 11
		lastPI := len(pages) - 1
		lastEndY := pageBodyStart(g, kopH, lastPI) + sumRowHeights(pages[lastPI], heights)
		if g.pageH-pdfFooterZone-lastEndY >= need {
			t.Logf("landscape=%v rekap muat di halaman data, lewati", landscape)
			continue
		}

		before := len(pages[lastPI])
		pages = moveTailRows(pages, 3)
		afterLast := len(pages[len(pages)-1])
		want := 3
		if before < 3 {
			want = before
		}
		if afterLast != want {
			t.Errorf("landscape=%v halaman penutup harus %d baris, dapat %d", landscape, want, afterLast)
		}

		// total baris harus tetap (tidak hilang/doppel)
		total := 0
		for _, p := range pages {
			total += len(p)
		}
		if total != len(heights) {
			t.Errorf("landscape=%v total baris berubah: %d != %d", landscape, total, len(heights))
		}

		// BuildPDF dgn data yg sama tetap valid & lintas halaman
		res, err := BuildPDF(d, PDFOptions{Landscape: landscape})
		if err != nil {
			t.Fatalf("landscape=%v BuildPDF gagal: %v", landscape, err)
		}
		if res.Pages <= 1 || !strings.HasPrefix(string(res.Bytes[:5]), "%PDF-") {
			t.Errorf("landscape=%v output tidak valid: %d halaman", landscape, res.Pages)
		}
	}
}
