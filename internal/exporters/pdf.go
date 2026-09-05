package exporters

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

// PDFOptions opsi generator PDF.
type PDFOptions struct {
	Landscape bool // true = A4 landscape (default), false = portrait
}

// PDFResult hasil generate: byte dokumen + jumlah halaman (untuk test FR-IE6).
type PDFResult struct {
	Bytes []byte
	Pages int
}

const (
	pdfMargin  = 12.0
	pdfLineH   = 4.3 // tinggi satu baris teks mm (Helvetica 9)
	pdfMinRowH = 7.0 //
	pdfPad     = 1.4 //
	// notesGap jarak ekstra antara baris Terbilang dan baris Catatan agar
	// catatan SPH mendapat ruang tersendiri di bawah terbilang.
	notesGap = 6.0

	// pdfFooterZone jarak tepi bawah kertas ke baris atas footer (SetY(-9)).
	pdfFooterZone = 9.0
)

type pdfGeom struct {
	wNo, wUraian, wQty, wSat, wUnit, wSvc, wMat, wJml float64
	xNo, xUraian, xQty, xSat, xUnit, xSvc, xMat, xJml float64
	totalW                                            float64
	pageW, pageH                                      float64
	usableH                                           float64
}

func newGeom(landscape bool) pdfGeom {
	var g pdfGeom
	if landscape {
		g.pageW, g.pageH = 297, 210
		g.wNo, g.wUraian, g.wQty, g.wSat = 9, 117, 13, 15
		g.wUnit, g.wSvc, g.wMat, g.wJml = 29, 30, 30, 30
	} else {
		g.pageW, g.pageH = 210, 297
		g.wNo, g.wUraian, g.wQty, g.wSat = 8, 65, 11, 13
		g.wUnit, g.wSvc, g.wMat, g.wJml = 26, 21, 21, 21
	}
	g.totalW = g.wNo + g.wUraian + g.wQty + g.wSat + g.wUnit + g.wSvc + g.wMat + g.wJml
	// posisi kolom RELATIF terhadap margin kiri; fungsi gambar menambah pdfMargin.
	x := 0.0
	for _, assign := range []struct {
		w *float64
		x *float64
	}{
		{&g.wNo, &g.xNo}, {&g.wUraian, &g.xUraian}, {&g.wQty, &g.xQty}, {&g.wSat, &g.xSat},
		{&g.wUnit, &g.xUnit}, {&g.wSvc, &g.xSvc}, {&g.wMat, &g.xMat}, {&g.wJml, &g.xJml},
	} {
		*assign.x = x
		x += *assign.w
	}
	g.usableH = g.pageH - 2*pdfMargin
	return g
}

// BuildPDF menghasilkan dokumen PDF A4 profesional (FR-IE6): kop + judul,
// tabel multi-halaman dengan header berulang dan baris Pindahan (carry
// akumulasi halaman sebelumnya), rekap Sub Total → Grand Total → Terbilang,
// catatan, blok tanda tangan, serta footer nomor halaman.
func BuildPDF(d *ExportData, opts PDFOptions) (*PDFResult, error) {
	g := newGeom(opts.Landscape)
	pdf := fpdf.New(landPort(opts.Landscape), "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AliasNbPages("{nb}")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-9)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(110, 118, 130)
		pdf.CellFormat(g.totalW, 5, tr(fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo())), "", 0, "C", false, 0, "")
	})
	pdf.SetTextColor(28, 32, 38)
	pdf.SetDrawColor(154, 165, 177)
	pdf.SetLineWidth(0.18)

	headH := tableHeadHeight()
	kopH := kopHeight(pdf, tr, g, d)
	sigH := signatureHeight(pdf, tr, g, d)
	rekapH := rekapHeight(pdf, tr, g, d)

	wrapped := make([][]string, len(d.Rows))
	heights := make([]float64, len(d.Rows))
	for i := range d.Rows {
		wrapped[i] = wrapUraian(pdf, tr, g, &d.Rows[i])
		heights[i] = rowHeightLines(len(wrapped[i]))
	}
	// Kapasitas baris: badan tabel mulai setelah kop + gap 2mm + header tabel
	// (headH), jadi avail harus mengurangi ketiganya agar baris tidak melewati
	// batas bawah pageH-margin (footer berada di pageH-9).
	availFirst := g.usableH - kopH - headH - 2
	availRest := g.usableH - headH - pindahanHeight() - 2
	pages := paginate(heights, availFirst, availRest)

	// Agar halaman rekap terakhir tidak kosong: bila rekap tak muat di halaman
	// data terakhir (sehingga ia menggeser rekap ke halaman baru), pindahkan
	// hingga beberapa baris uraian terakhir ke halaman rekap tersebut. Dengan
	// begitu halaman penutup tetap memuat uraian pekerjaan di samping rekap,
	// dan total/carry dihitung ulang supaya tetap konsisten.
	need := rekapH + sigH + terbilangNotesHeight(pdf, tr, g, d) + 11
	lastPI := len(pages) - 1
	if len(pages[lastPI]) > 0 {
		lastEndY := pageBodyStart(g, kopH, lastPI) + sumRowHeights(pages[lastPI], heights)
		if g.pageH-pdfFooterZone-lastEndY < need {
			backfillRows := 3
			pages = moveTailRows(pages, backfillRows)
		}
	}
	carry := computeCarry(pages, d.Rows)

	for pi := range pages {
		// AddPage eksplisit WAJIB termasuk halaman pertama — menggambar tanpa
		// halaman terbuka membuat fpdf menyuntik content stream sebelum
		// header %PDF sehingga file rusak.
		pdf.AddPage()
		if pi == 0 {
			drawKop(pdf, tr, g, d)
		}
		drawTableHead(pdf, g, pdf.GetY()+2)
		y := pdf.GetY()
		if pi > 0 {
			drawPindahan(pdf, tr, g, y, carry[pi])
			y += pindahanHeight()
		}
		for _, ri := range pages[pi] {
			drawRow(pdf, tr, g, &d.Rows[ri], y, heights[ri], wrapped[ri])
			y += heights[ri]
		}

		last := pi == len(pages)-1
		if !last {
			continue
		}
		// Rekap + ttd harus selesai DI ATAS zona footer (9mm dari bawah, lokasi
		// "Halaman x dari y"); include tinggi blok Terbilang/Catatan.
		if g.pageH-pdfFooterZone-y < need {
			pdf.AddPage()
			drawTableHead(pdf, g, pdf.GetY())
			y = pdfMargin + headH
			// halaman ini hanya berisi rekap → Pindahan memuat seluruh akumulasi
			drawPindahan(pdf, tr, g, y, [3]int64{d.SubtotalService, d.SubtotalMaterial, d.GrandTotal})
			y += pindahanHeight()
		}
		y = drawRekap(pdf, tr, g, d, y+3)
		y = drawTerbilangNotes(pdf, tr, g, d, y+2)
		drawSignature(pdf, tr, g, d, y+6)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return &PDFResult{Bytes: buf.Bytes(), Pages: pdf.PageNo()}, nil
}

func landPort(l bool) string {
	if l {
		return "L"
	}
	return "P"
}

// ===== paginasi =====

func paginate(heights []float64, availFirst, availRest float64) [][]int {
	pages := [][]int{}
	cur := []int{}
	used := 0.0
	avail := availFirst
	for i, h := range heights {
		if len(cur) > 0 && used+h > avail {
			pages = append(pages, cur)
			cur = []int{}
			used = 0
			avail = availRest
		}
		cur = append(cur, i)
		used += h
	}
	if len(cur) > 0 {
		pages = append(pages, cur)
	}
	if len(pages) == 0 {
		pages = [][]int{{}}
	}
	return pages
}

// pageBodyStart posisi Y (relatif ujung atas kertas) awal badan tabel baris
// pada halaman pi — konsisten dengan BuildPDF: kop (halaman 1) / "Pindahan"
// (halaman 2+) diikuti header tabel.
func pageBodyStart(g pdfGeom, kopH float64, pi int) float64 {
	y := pdfMargin + 2 + tableHeadHeight()
	if pi > 0 {
		y += pindahanHeight()
	} else {
		y += kopH
	}
	return y
}

// sumRowHeights menjumlahkan tinggi baris pada halaman (indeks baris).
func sumRowHeights(page []int, heights []float64) float64 {
	s := 0.0
	for _, ri := range page {
		s += heights[ri]
	}
	return s
}

// moveTailRows memindahkan hingga n baris TERAKHIR dari halaman terakhir ke
// halaman baru (penutup) yang di-append di belakang. Halaman penutup inilah
// yang memuat baris uraian sisa sekaligus rekap, sehingga tidak kosong.
// Bila seluruh isi halaman terakhir turut dipindah, halaman kosong itu
// dibuang agar tidak menyisakan halaman tanpa baris di tengah dokumen.
func moveTailRows(pages [][]int, n int) [][]int {
	if len(pages) == 0 {
		return pages
	}
	last := len(pages) - 1
	cur := pages[last]
	if len(cur) == 0 {
		return pages
	}
	back := n
	if back > len(cur) {
		back = len(cur)
	}
	moved := make([]int, back)
	copy(moved, cur[len(cur)-back:])
	rest := cur[:len(cur)-back]
	if len(rest) == 0 {
		pages = pages[:last]
	} else {
		pages[last] = rest
	}
	pages = append(pages, moved)
	return pages
}

// computeCarry: carry[pi] = akumulasi (jasa, material, total) seluruh baris
// HALAMAN SEBELUM halaman pi — angka yang tampil di baris "Pindahan".
func computeCarry(pages [][]int, rows []Row) [][3]int64 {
	carry := make([][3]int64, len(pages))
	var svc, mat, jml int64
	for p := range pages {
		carry[p] = [3]int64{svc, mat, jml}
		for _, ri := range pages[p] {
			svc += rows[ri].ServiceTotal
			mat += rows[ri].MaterialTotal
			jml += rows[ri].Total
		}
	}
	return carry
}

// ===== pengukuran tinggi =====

// rowHeightLines menghitung tinggi baris dari jumlah baris hasil wrap.
func rowHeightLines(n int) float64 {
	h := float64(n)*pdfLineH + 2*pdfPad
	if h < pdfMinRowH {
		h = pdfMinRowH
	}
	return h
}

func uraianText(r *Row) string {
	parts := []string{r.Name}
	if r.EmptyNumbers && r.SubNo != "" {
		if s := qtyUnitSuffix(r.Qty, r.Unit); s != "" {
			parts[0] = r.Name + s
		}
	}
	if d := joinDesc(r.Description, r.WeightNote); d != "" {
		parts = append(parts, d)
	}
	return strings.Join(parts, "\n")
}

// wrapTextLines membungkus teks menjadi baris selebar width. Lebar baris
// diakumulasi inkremental per kata (satu GetStringWidth per kata) sehingga
// biayanya O(panjang teks) — berbeda dari SplitText/MultiCell fpdf yang
// mengukur ulang kalimat menumpuk dan menjadi lambat pada deskripsi panjang.
func wrapTextLines(pdf *fpdf.Fpdf, text string, width float64) []string {
	var out []string
	sep := pdf.GetStringWidth(" ")
	for _, seg := range strings.Split(text, "\n") {
		words := strings.Fields(strings.TrimSpace(seg))
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		line := words[0]
		w := pdf.GetStringWidth(words[0])
		for _, word := range words[1:] {
			ww := pdf.GetStringWidth(word)
			if w+sep+ww <= width {
				line += " " + word
				w += sep + ww
			} else {
				out = append(out, line)
				line = word
				w = ww
			}
		}
		out = append(out, line)
	}
	return out
}

// subPrefixOf awalan huruf sub point ("a.") atau kosong untuk main point.
func subPrefixOf(r *Row) string {
	if r.SubNo == "" {
		return ""
	}
	return r.SubNo + "."
}

// wrapUraian menghitung baris tampil kolom Uraian satu baris data; hasilnya
// dipakai bersama untuk pengukuran tinggi dan penggambaran (satu kali wrap).
// Awalan sub point ("a.") digambar terpisah oleh drawRow; fungsi ini hanya
// menghitung baris teks di lebar tersisa (sudah dikurangi prefix width).
//
// Font dipilih mengikuti gaya baris (bold utk main point, regular utk sub
// point) agar lebar terukur konsisten dengan font saat menggambar — teks bold
// lebih lebar dan bila diukur dgn regular akan meluber ke kolom di sebelahnya.
func wrapUraian(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, r *Row) []string {
	fontStyle := ""
	if r.Bold {
		fontStyle = "B"
	}
	pdf.SetFont("Helvetica", fontStyle, 9)
	avail := g.wUraian - 2*pdfPad - indentOf(r)
	prefix := subPrefixOf(r)
	if prefix != "" {
		avail -= pdf.GetStringWidth(prefix + " ")
	}
	lines := wrapTextLines(pdf, tr(uraianText(r)), avail)
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

func indentOf(r *Row) float64 {
	if r.Bold {
		return 0
	}
	return 4.5
}

const (
	tableHeadBandH = 5.0
	tableHeadRowH  = 7.5
	pindahanH      = 6.5
)

func tableHeadHeight() float64 { return tableHeadBandH + tableHeadRowH }
func pindahanHeight() float64  { return pindahanH }

// ===== gambar bagian =====

func drawTableHead(pdf *fpdf.Fpdf, g pdfGeom, y float64) {
	pdf.SetFillColor(221, 235, 247)
	pdf.SetFont("Helvetica", "B", 8)
	// band grup kolom harga
	bandX := pdfMargin + g.xSvc
	bandW := g.wSvc + g.wMat + g.wJml
	pdf.Rect(bandX, y, bandW, tableHeadBandH, "FD")
	leftW := bandX - pdfMargin
	if leftW > 0 {
		pdf.Rect(pdfMargin, y, leftW, tableHeadBandH, "FD")
	}
	pdf.SetXY(bandX, y)
	pdf.CellFormat(bandW, tableHeadBandH, "PENAWARAN HARGA (Rp)", "1", 0, "C", false, 0, "")

	y2 := y + tableHeadBandH
	cells := []struct {
		x, w   float64
		label  string
		align  string
		fillBg bool
	}{
		{g.xNo, g.wNo, "NO", "C", false},
		{g.xUraian, g.wUraian, "URAIAN KEGIATAN", "C", false},
		{g.xQty, g.wQty, "JML", "C", false},
		{g.xSat, g.wSat, "SAT", "C", false},
		{g.xUnit, g.wUnit, "HARGA SATUAN", "C", false},
		{g.xSvc, g.wSvc, "JASA", "C", false},
		{g.xMat, g.wMat, "MAT", "C", false},
		{g.xJml, g.wJml, "JML", "C", false},
	}
	pdf.Rect(pdfMargin, y2, g.totalW, tableHeadRowH, "FD")
	for _, c := range cells {
		pdf.SetXY(pdfMargin+c.x, y2)
		pdf.CellFormat(c.w, tableHeadRowH, c.label, "", 0, c.align, false, 0, "")
	}
	// garis vertikal pemisah kolom pada kedua baris header
	vlines(pdf, g, y, tableHeadHeight())
	// garis horizontal antara band dan baris label
	pdf.Line(pdfMargin, y2, pdfMargin+g.totalW, y2)
	pdf.SetY(y + tableHeadHeight())
}

func vlines(pdf *fpdf.Fpdf, g pdfGeom, y, h float64) {
	for _, xv := range []float64{g.xNo, g.xUraian, g.xQty, g.xSat, g.xUnit, g.xSvc, g.xMat, g.xJml} {
		pdf.Line(pdfMargin+xv, y, pdfMargin+xv, y+h)
	}
	pdf.Line(pdfMargin+g.totalW, y, pdfMargin+g.totalW, y+h)
}

func drawPindahan(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, y float64, carry [3]int64) {
	h := pindahanH
	pdf.Rect(pdfMargin, y, g.totalW, h, "D")
	vlines(pdf, g, y, h)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(90, 99, 112)
	pdf.SetXY(pdfMargin+g.xUraian+pdfPad, y+1.1)
	pdf.CellFormat(g.wUraian-2*pdfPad, h-2, tr("Pindahan"), "", 0, "L", false, 0, "")
	writeMoneyCells(pdf, tr, g, y, h, carry[0], carry[1], carry[2])
	pdf.SetTextColor(28, 32, 38)
	pdf.SetY(y + h)
}

func drawRow(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, r *Row, y, h float64, lines []string) {
	pdf.Rect(pdfMargin, y, g.totalW, h, "D")
	vlines(pdf, g, y, h)

	fontStyle := ""
	if r.Bold {
		fontStyle = "B"
	}
	pdf.SetFont("Helvetica", fontStyle, 9)

	// NO
	putCell(pdf, tr, g.xNo, g.wNo, y, h, r.No, "C", false)
	// URAIAN (multi-line; baris sudah dihitung sekali di BuildPDF)
	indent := indentOf(r)
	x := pdfMargin + g.xUraian + pdfPad + indent
	prefix := subPrefixOf(r)
	prefixW := 0.0
	if prefix != "" {
		prefixW = pdf.GetStringWidth(prefix + " ")
		pdf.SetXY(x, y+pdfPad)
		pdf.CellFormat(prefixW, pdfLineH, tr(prefix), "", 0, "L", false, 0, "")
	}
	w := g.wUraian - indent - 2*pdfPad
	for i, ln := range lines {
		pdf.SetXY(x+prefixW, y+pdfPad+float64(i)*pdfLineH)
		pdf.CellFormat(w-prefixW, pdfLineH, ln, "", 0, "L", false, 0, "")
	}
	// angka — dikosongkan semua bila baris bertipe EmptyNumbers (main point
	// tanpa nilai, atau sub point pada blok dgn main point terisi).
	if !r.EmptyNumbers {
		putCell(pdf, tr, g.xQty, g.wQty, y, h, qtyText(r.Qty), "C", false)
		putCell(pdf, tr, g.xSat, g.wSat, y, h, tr(strings.TrimSpace(r.Unit)), "C", false)
		up := ""
		if r.UnitPrice > 0 {
			up = idNumber(r.UnitPrice)
		}
		putCell(pdf, tr, g.xUnit, g.wUnit, y, h, up, "R", false)
		writeMoneyCells(pdf, tr, g, y, h, r.ServiceTotal, r.MaterialTotal, r.Total)
	}
	pdf.SetFont("Helvetica", "", 9)
}

func putCell(pdf *fpdf.Fpdf, _ func(string) string, x, w, y, h float64, text, align string, fill bool) {
	th := pdfLineH
	if h > th {
		th = h - 2*pdfPad
	}
	pdf.SetXY(pdfMargin+x+pdfPad, y+(h-th)/2)
	pdf.CellFormat(w-2*pdfPad, th, text, "", 0, align, fill, 0, "")
}

func writeMoneyCells(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, y, h float64, svc, mat, jml int64) {
	cells := []struct {
		x, w float64
		v    int64
	}{
		{g.xSvc, g.wSvc, svc},
		{g.xMat, g.wMat, mat},
		{g.xJml, g.wJml, jml},
	}
	for _, c := range cells {
		txt := ""
		if c.v != 0 {
			txt = idNumber(c.v)
		}
		putCell(pdf, tr, c.x, c.w, y, h, txt, "R", false)
	}
}

// drawKop kop + judul + blok info dokumen di halaman pertama.
func drawKop(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, d *ExportData) {
	const logoBox = 22.0
	y := pdfMargin

	if d.Company.LogoPath != "" {
		func() {
			defer func() { _ = recover() }() // gambar rusak tidak boleh membatalkan export
			pdf.Image(d.Company.LogoPath, pdfMargin, y, logoBox, logoBox, false, "", 0, "")
		}()
	}
	tx := pdfMargin + logoBox + 4
	pdf.SetXY(tx, y+1)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(g.totalW-logoBox-4, 6.5, tr(d.Company.Name), "", 2, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(90, 99, 112)
	pdf.CellFormat(g.totalW-logoBox-4, 4.5, tr(d.Company.City), "", 2, "L", false, 0, "")
	pdf.CellFormat(g.totalW-logoBox-4, 4.5, tr(d.Company.Address), "", 2, "L", false, 0, "")
	pdf.SetTextColor(28, 32, 38)

	y += 26
	pdf.SetXY(pdfMargin, y)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(g.totalW, 7, tr(strings.ToUpper(d.TitleLine1())), "", 2, "C", false, 0, "")
	pdf.CellFormat(g.totalW, 6.5, tr(strings.ToUpper(d.TitleLine2())), "", 2, "C", false, 0, "")

	y += 16.5
	info := docInfoPairs(d)
	pdf.SetFont("Helvetica", "", 9)
	colW := (g.totalW - 8) / 2
	labelW := 30.0
	for i := 0; i < len(info); i += 2 {
		maxLines := 1
		rowY := y
		for k := 0; k < 2 && i+k < len(info); k++ {
			it := info[i+k]
			x := pdfMargin + float64(k)*(colW+8)
			lines := drawInfoPair(pdf, tr, it.label, it.value, x, rowY, labelW, colW-labelW)
			if lines > maxLines {
				maxLines = lines
			}
		}
		y = rowY + float64(maxLines)*pdfLineH + 0.8
	}
	pdf.SetY(y)
}

type kv struct{ label, value string }

func docInfoPairs(d *ExportData) []kv {
	doc := d.Document
	date := FormatDateID(doc.Date)
	valid := ""
	if doc.ValidUntil != nil {
		valid = FormatDateID(*doc.ValidUntil)
	}
	pairs := []kv{
		{"Nomor", doc.Number},
		{"Tanggal", date},
	}
	if doc.Revision > 0 {
		pairs = append(pairs, kv{"Revisi", fmt.Sprintf("%d", doc.Revision)})
	}
	pairs = append(pairs,
		kv{"Kepada", doc.Customer},
		kv{"Kapal", doc.Vessel},
		kv{"Proyek", doc.Project},
		kv{"Subjek", doc.Subject},
		kv{"Referensi", doc.Reference},
		kv{"Lokasi", doc.Location},
	)
	if valid != "" {
		pairs = append(pairs, kv{"Berlaku s.d.", valid})
	}
	if pic := strings.TrimSpace(doc.PIC); pic != "" {
		pairs = append(pairs, kv{"PIC", pic})
	}
	out := make([]kv, 0, len(pairs))
	for _, p := range pairs {
		if strings.TrimSpace(p.value) != "" {
			out = append(out, p)
		}
	}
	return out
}

// drawInfoPair menulis satu pasangan label:value; balik jumlah baris terpakai.
func drawInfoPair(pdf *fpdf.Fpdf, tr func(string) string, label, value string, x, y, labelW, valueW float64) int {
	pdf.SetXY(x, y)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(labelW, pdfLineH, tr(label+" :"), "", 0, "L", false, 0, "")
	pdf.SetXY(x+labelW, y)
	pdf.MultiCell(valueW, pdfLineH, tr(value), "", "L", false)
	return len(pdf.SplitText(value, valueW))
}

// kopHeight menghitung tinggi blok kop+judul+info (konsisten dgn drawKop).
func kopHeight(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, d *ExportData) float64 {
	h := 26.0 // logo/perusahaan
	h += 16.5 // judul dua baris
	info := docInfoPairs(d)
	valueW := ((g.totalW - 8) / 2) - 30.0
	pdf.SetFont("Helvetica", "", 9)
	for i := 0; i < len(info); i += 2 {
		lines := 1
		for k := 0; k < 2 && i+k < len(info); k++ {
			if n := len(pdf.SplitText(tr(info[i+k].value), valueW)); n > lines {
				lines = n
			}
		}
		h += float64(lines)*pdfLineH + 0.8
	}
	return h
}

func drawRekap(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, d *ExportData, y float64) float64 {
	rh := 6.4
	labelX := pdfMargin + g.xUraian
	labelW := g.xSvc - g.xUraian

	line := func(label string, colIdx int, v int64, bold bool) {
		style := ""
		size := 9.5
		if bold {
			style = "B"
			size = 10.5
		}
		pdf.SetFont("Helvetica", style, size)
		pdf.SetXY(labelX, y+1.1)
		pdf.CellFormat(labelW, rh-2.2, tr(label), "", 0, "R", false, 0, "")
		cols := [][2]float64{{g.xSvc, g.wSvc}, {g.xMat, g.wMat}, {g.xJml, g.wJml}}
		for ci, c := range cols {
			txt := ""
			if ci == colIdx {
				txt = idNumber(v)
			}
			putCell(pdf, tr, c[0], c[1], y, rh, txt, "R", false)
		}
		y += rh
	}
	line("Sub Total Jasa", 0, d.SubtotalService, false)
	line("Sub Total Material", 1, d.SubtotalMaterial, false)
	// GRAND TOTAL dibingkai tebal
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetXY(labelX, y+1.3)
	pdf.CellFormat(labelW, rh, tr("GRAND TOTAL"), "", 0, "R", false, 0, "")
	pdf.SetLineWidth(0.5)
	pdf.Rect(pdfMargin+g.xSvc, y, g.wSvc+g.wMat+g.wJml, rh, "D")
	pdf.Line(pdfMargin+g.xMat, y, pdfMargin+g.xMat, y+rh)
	pdf.SetLineWidth(0.18)
	pdf.SetXY(pdfMargin+g.xJml, y+1.3)
	pdf.CellFormat(g.wJml-2*pdfPad, rh, idNumber(d.GrandTotal), "", 0, "R", false, 0, "")
	pdf.SetY(y + rh + 1)
	return pdf.GetY()
}

// rekapHeight tinggi blok rekap (konsisten dgn drawRekap).
func rekapHeight(_ *fpdf.Fpdf, _ func(string) string, _ pdfGeom, _ *ExportData) float64 {
	return 3*6.4 + 1
}

func drawTerbilangNotes(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, d *ExportData, y float64) float64 {
	pdf.SetFont("Helvetica", "I", 9)
	if tb := d.Document.Terbilang; tb != "" {
		text := tr("Terbilang : " + tb + ".")
		pdf.SetXY(pdfMargin, y)
		pdf.MultiCell(g.totalW, pdfLineH, text, "", "L", false)
		y += float64(len(pdf.SplitText(text, g.totalW)))*pdfLineH + 1
		if d.Document.Notes != "" {
			y += notesGap
		}
	}
	if n := d.Document.Notes; n != "" {
		pdf.SetFont("Helvetica", "", 8.5)
		text := tr("Catatan: " + n)
		pdf.SetXY(pdfMargin, y)
		pdf.MultiCell(g.totalW, pdfLineH, text, "", "L", false)
		y += float64(len(pdf.SplitText(text, g.totalW)))*pdfLineH + 1
		pdf.SetFont("Helvetica", "", 9)
	}
	return y
}

// terbilangNotesHeight mengukur tinggi blok Terbilang + Catatan (konsisten
// dengan drawTerbilangNotes) untuk keputusan pemindahan rekap ke halaman baru.
func terbilangNotesHeight(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, d *ExportData) float64 {
	h := 0.0
	if tb := d.Document.Terbilang; tb != "" {
		pdf.SetFont("Helvetica", "I", 9)
		text := tr("Terbilang : " + tb + ".")
		h += float64(len(pdf.SplitText(text, g.totalW)))*pdfLineH + 1
		if d.Document.Notes != "" {
			h += notesGap
		}
	}
	if n := d.Document.Notes; n != "" {
		pdf.SetFont("Helvetica", "", 8.5)
		text := tr("Catatan: " + n)
		h += float64(len(pdf.SplitText(text, g.totalW)))*pdfLineH + 1
	}
	return h
}

func signatureHeight(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, d *ExportData) float64 {
	_ = pdf
	_ = tr
	_ = g
	_ = d
	return 42.0
}

// drawSignature menggambar blok ttd kanan bawah (kota/tanggal, perusahaan,
// ruang, nama, jabatan) plus stempel & tanda tangan (PNG transparan) pada
// posisi hasil editor; bila belum diatur, dipakai posisi otomatis.
func drawSignature(pdf *fpdf.Fpdf, tr func(string) string, g pdfGeom, d *ExportData, y float64) {
	const blockW = 78.0
	x := pdfMargin + g.totalW - blockW

	// Gambar tanda tangan & stempel ditempel di zona ttd. Gambar rusak/gagal
	// dibaca tidak boleh membatalkan export (pola logo).
	func() {
		defer func() { _ = recover() }()
		drawSignatureImage(pdf, x, y, d)
		drawStampImage(pdf, x, y, d)
	}()

	dateLine := FormatDateID(d.Document.Date)
	if d.Company.City != "" {
		dateLine = d.Company.City + ", " + dateLine
	}
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(x, y)
	pdf.CellFormat(blockW, 5.5, tr(dateLine), "", 2, "C", false, 0, "")
	pdf.CellFormat(blockW, 5.5, tr(d.Company.Name), "", 2, "C", false, 0, "")
	pdf.SetFont("Helvetica", "BU", 10)
	pdf.SetXY(x, y+22)
	pdf.CellFormat(blockW, 5.5, tr(d.Company.SignerName), "", 2, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(blockW, 5.5, tr(d.Company.SignerPosition), "", 2, "C", false, 0, "")
	pdf.SetY(y + 34)
}

// effectiveSignPos mengembalikan posisi default (otomatis) bila belum diatur.
func effectiveSignPos(c *CompanyInfo) (x, y, size float64) {
	if c.SignatureSize > 0 {
		return clampFrac(c.SignaturePosX), clampFrac(c.SignaturePosY), clampFrac(c.SignatureSize)
	}
	// default: terpusat, tepat di atas baris nama (mirip perilaku lama)
	return 0.5, 0.19, 0.30
}

// effectiveStampPos serupa utk stempel (menimpa area tanda tangan).
func effectiveStampPos(c *CompanyInfo) (x, y, size float64) {
	if c.StampSize > 0 {
		return clampFrac(c.StampPosX), clampFrac(c.StampPosY), clampFrac(c.StampSize)
	}
	return 0.5, 0.21, 0.32
}

func clampFrac(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// drawSignatureImage menggambar tanda tangan (PNG) pada posisi manual atau
// default; proporsi dipertahankan dan dibatasi lebar blok.
func drawSignatureImage(pdf *fpdf.Fpdf, x0, y float64, d *ExportData) {
	if d.Company.SignaturePath == "" {
		return
	}
	fx, fy, fw := effectiveSignPos(&d.Company)
	placeImage(pdf, d.Company.SignaturePath, x0, y, fx, fy, fw)
}

// drawStampImage menggambar stempel (PNG) pada posisi manual atau default.
func drawStampImage(pdf *fpdf.Fpdf, x0, y float64, d *ExportData) {
	if d.Company.StampPath == "" {
		return
	}
	fx, fy, fw := effectiveStampPos(&d.Company)
	placeImage(pdf, d.Company.StampPath, x0, y, fx, fy, fw)
}

// placeImage menggambar PNG pada (x0+fx*W, y+fy*H) dgn lebar fw*W (tinggi
// mengikuti rasio aspek), dibatasi agar tidak melewati kotak blok.
func placeImage(pdf *fpdf.Fpdf, path string, x0, y, fx, fy, fw float64) {
	if !fileExists(path) {
		return
	}
	defer func() { _ = recover() }()
	info := pdf.RegisterImageOptions(path, fpdf.ImageOptions{ReadDpi: false, ImageType: "png"})
	if info == nil {
		return
	}
	iw, ih := info.Width(), info.Height()
	if iw <= 0 || ih <= 0 {
		return
	}
	blockW, blockH := SignatureBlockW, SignatureBlockH
	fw = clampFrac(fw)
	if fw <= 0 {
		fw = 0.30
	}
	w := fw * blockW
	h := w * ih / iw
	// jaga agar gambar tak lebih tinggi dari blok
	if h > blockH {
		h = blockH
		w = h * iw / ih
	}
	ix := x0 + clampFrac(fx)*blockW
	iy := y + clampFrac(fy)*blockH
	// pastikan masih dalam kotak
	if ix < x0 {
		ix = x0
	}
	if ix+w > x0+blockW {
		ix = x0 + blockW - w
	}
	if iy < y {
		iy = y
	}
	if iy+h > y+blockH {
		iy = y + blockH - h
	}
	pdf.Image(path, ix, iy, w, h, false, "", 0, "")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// ===== util format =====

// qtyText format qty gaya Indonesia: "2", "1,5", "0,25" (tanpa nol berlebih).
func qtyText(q float64) string {
	s := strconv.FormatFloat(q, 'f', -1, 64)
	return strings.Replace(s, ".", ",", 1)
}

// idNumber memformat integer gaya Indonesia: 120725000 → "120.725.000".
func idNumber(v int64) string {
	s := fmt.Sprintf("%d", v)
	neg := ""
	if strings.HasPrefix(s, "-") {
		neg = "-"
		s = s[1:]
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return neg + strings.Join(parts, ".")
}
