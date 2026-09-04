package exporters

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Layout kolom mengikuti struktur referensi (kolom B–K):
//
//	B   NO            | C–E URAIAN KEGIATAN      | F JML | G SAT
//	H   HARGA SATUAN  | I–K PENAWARAN HARGA (Rp) → JASA / MAT / JML
const (
	xlSheet     = "SPH"
	xlFirstCol  = "B"
	xlLastCol   = "K"
	rpFmt       = `"Rp"#,##0`
	uraianWidth = 66 // perkiraan karakter per baris kolom uraian (C:E)
)

type xlBorderSet []excelize.Border

func thinBorders() xlBorderSet {
	return xlBorderSet{
		{Type: "left", Style: 1, Color: "9AA5B1"},
		{Type: "right", Style: 1, Color: "9AA5B1"},
		{Type: "top", Style: 1, Color: "9AA5B1"},
		{Type: "bottom", Style: 1, Color: "9AA5B1"},
	}
}

func mediumTop() xlBorderSet {
	b := thinBorders()
	b[2] = excelize.Border{Type: "top", Style: 2, Color: "4B5563"}
	return b
}

// BuildExcel menyusun workbook SPH profesional siap cetak (FR-IE5):
// kop + judul bermarge, header tabel berulang tiap halaman cetak
// (print titles), format Rupiah, rekap, terbilang, dan blok tanda tangan;
// page setup A4 landscape fit-width dengan print area eksplisit.
func BuildExcel(d *ExportData) (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetSheetName("Sheet1", xlSheet); err != nil {
		return nil, err
	}

	st, err := buildStyles(f)
	if err != nil {
		return nil, err
	}

	setWidths(f)
	setupPage(f)

	row := writeKopAndTitle(f, d, st)
	row++ // satu baris spasi sebelum header tabel

	headTop := row
	if err := writeTableHead(f, headTop, st); err != nil {
		return nil, err
	}
	row += 2

	lastDataRow, err := writeRows(f, d, row, st)
	if err != nil {
		return nil, err
	}
	row = lastDataRow + 1

	row, err = writeRekap(f, d, row, st)
	if err != nil {
		return nil, err
	}
	writeTerbilangNotes(f, d, row+1, st)
	row += terbilangNoteRows(d)

	writeSignature(f, d, row+1, st)
	endRow := row + 8

	// Print area + repeating header ($8:$9 relatif ke area cetak).
	_ = f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Area",
		RefersTo: fmt.Sprintf("'%s'!$%s$1:$%s$%d", xlSheet, xlFirstCol, xlLastCol, endRow),
		Scope:    xlSheet,
	})
	_ = f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Titles",
		RefersTo: fmt.Sprintf("'%s'!$%d:$%d", xlSheet, headTop, headTop+1),
		Scope:    xlSheet,
	})
	_ = f.SetHeaderFooter(xlSheet, &excelize.HeaderFooterOptions{OddHeader: "&C&P"})
	return f, nil
}

func setWidths(f *excelize.File) {
	widths := map[string]float64{"B": 5, "C": 40, "D": 14, "E": 14, "F": 7, "G": 9, "H": 17, "I": 16, "J": 16, "K": 18}
	for col, w := range widths {
		_ = f.SetColWidth(xlSheet, col, col, w)
	}
}

func setupPage(f *excelize.File) {
	size, orient := 9, "landscape"
	fitW, fitH := 1, 0
	_ = f.SetPageLayout(xlSheet, &excelize.PageLayoutOptions{
		Size: &size, Orientation: &orient, FitToWidth: &fitW, FitToHeight: &fitH,
	})
	l, r, t, b := 0.394, 0.394, 0.59, 0.59
	hd, ft := 0.31, 0.31
	_ = f.SetPageMargins(xlSheet, &excelize.PageLayoutMarginsOptions{
		Left: &l, Right: &r, Top: &t, Bottom: &b, Header: &hd, Footer: &ft,
	})
	show := false
	_ = f.SetSheetView(xlSheet, 0, &excelize.ViewOptions{ShowGridLines: &show})
}

type xlStyles struct {
	title, company, cityAddr, docNo        int
	head                                   int
	cellWrap, cellCenter, mainCell         int
	mainCenter, subCell                    int
	money, moneyBold, grandLabel, grandVal int
	italic, sigName, sigLine               int
}

func buildStyles(f *excelize.File) (*xlStyles, error) {
	s := &xlStyles{}
	var err error
	boldFont := func(sz float64) *excelize.Font { return &excelize.Font{Bold: true, Size: sz} }
	borderThin := thinBorders()

	center := func(wrap bool) *excelize.Alignment {
		return &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: wrap}
	}
	leftTop := func(indent int) *excelize.Alignment {
		return &excelize.Alignment{Horizontal: "left", Vertical: "top", WrapText: true, Indent: indent}
	}

	mk := func(st *excelize.Style) (int, error) { return f.NewStyle(st) }

	if s.title, err = mk(&excelize.Style{Font: boldFont(13), Alignment: center(false)}); err != nil {
		return nil, err
	}
	if s.company, err = mk(&excelize.Style{Font: boldFont(15), Alignment: &excelize.Alignment{Vertical: "center"}}); err != nil {
		return nil, err
	}
	if s.cityAddr, err = mk(&excelize.Style{Font: &excelize.Font{Size: 9, Color: "5B6470"}, Alignment: &excelize.Alignment{Vertical: "center"}}); err != nil {
		return nil, err
	}
	if s.docNo, err = mk(&excelize.Style{Font: boldFont(10), Alignment: &excelize.Alignment{Horizontal: "right"}}); err != nil {
		return nil, err
	}
	if s.head, err = mk(&excelize.Style{
		Font: boldFont(9), Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DDEBF7"}},
		Alignment: center(true), Border: borderThin,
	}); err != nil {
		return nil, err
	}
	if s.cellWrap, err = mk(&excelize.Style{Font: &excelize.Font{Size: 9}, Alignment: leftTop(0), Border: borderThin}); err != nil {
		return nil, err
	}
	if s.cellCenter, err = mk(&excelize.Style{Font: &excelize.Font{Size: 9}, Alignment: center(false), Border: borderThin}); err != nil {
		return nil, err
	}
	if s.mainCell, err = mk(&excelize.Style{Font: boldFont(9), Alignment: leftTop(0), Border: mediumTop()}); err != nil {
		return nil, err
	}
	if s.mainCenter, err = mk(&excelize.Style{
		Font: boldFont(9), Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "top"},
		Border: mediumTop(),
	}); err != nil {
		return nil, err
	}
	if s.subCell, err = mk(&excelize.Style{Font: &excelize.Font{Size: 9}, Alignment: leftTop(2), Border: borderThin}); err != nil {
		return nil, err
	}
	if s.money, err = mk(&excelize.Style{
		Font: &excelize.Font{Size: 9}, Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "top"},
		CustomNumFmt: strPtr(rpFmt), Border: borderThin,
	}); err != nil {
		return nil, err
	}
	if s.moneyBold, err = mk(&excelize.Style{
		Font: boldFont(9), Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "top"},
		CustomNumFmt: strPtr(rpFmt), Border: mediumTop(),
	}); err != nil {
		return nil, err
	}
	if s.grandLabel, err = mk(&excelize.Style{Font: boldFont(11), Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"}, Border: borderThin}); err != nil {
		return nil, err
	}
	if s.grandVal, err = mk(&excelize.Style{
		Font: boldFont(11), Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		CustomNumFmt: strPtr(rpFmt),
		Border:       append(borderThin, excelize.Border{Type: "top", Style: 6, Color: "111827"}),
	}); err != nil {
		return nil, err
	}
	if s.italic, err = mk(&excelize.Style{Font: &excelize.Font{Italic: true, Size: 9}, Alignment: leftTop(0)}); err != nil {
		return nil, err
	}
	if s.sigName, err = mk(&excelize.Style{Font: boldFont(10), Alignment: center(false)}); err != nil {
		return nil, err
	}
	if s.sigLine, err = mk(&excelize.Style{Font: &excelize.Font{Size: 10}, Alignment: center(false)}); err != nil {
		return nil, err
	}
	return s, nil
}

func strPtr(s string) *string { return &s }

// writeKopAndTitle menulis kop perusahaan + judul dokumen; balik baris terakhir terpakai.
func writeKopAndTitle(f *excelize.File, d *ExportData, st *xlStyles) int {
	_ = f.SetRowHeight(xlSheet, 1, 34)
	_ = f.SetRowHeight(xlSheet, 2, 15)
	_ = f.SetRowHeight(xlSheet, 3, 15)

	if d.Company.LogoPath != "" {
		opts := &excelize.GraphicOptions{ScaleX: 0.32, ScaleY: 0.32, Positioning: "oneCell"}
		if err := f.AddPicture(xlSheet, "B1", d.Company.LogoPath, opts); err != nil {
			// logo gagal digambar tidak boleh membatalkan export
			_ = f.SetCellValue(xlSheet, "B2", "")
		}
	}
	_ = f.MergeCell(xlSheet, "D1", "H1")
	_ = f.SetCellStyle(xlSheet, "D1", "H1", st.company)
	_ = f.SetCellValue(xlSheet, "D1", d.Company.Name)
	_ = f.SetCellValue(xlSheet, "D2", strings.TrimSpace(d.Company.City))
	_ = f.SetCellStyle(xlSheet, "D2", "D3", st.cityAddr)
	_ = f.SetCellValue(xlSheet, "D3", d.Company.Address)

	_ = f.MergeCell(xlSheet, "B4", "K4")
	_ = f.SetCellStyle(xlSheet, "B4", "K4", st.docNo)
	num := "Nomor: " + d.Document.Number
	if d.Document.Revision > 0 {
		num += fmt.Sprintf("  ·  Rev %d", d.Document.Revision)
	}
	_ = f.SetCellValue(xlSheet, "B4", num)

	_ = f.MergeCell(xlSheet, "B5", "K5")
	_ = f.SetCellStyle(xlSheet, "B5", "K6", st.title)
	_ = f.SetCellValue(xlSheet, "B5", strings.ToUpper(d.TitleLine1()))
	_ = f.MergeCell(xlSheet, "B6", "K6")
	_ = f.SetCellValue(xlSheet, "B6", strings.ToUpper(d.TitleLine2()))
	_ = f.SetRowHeight(xlSheet, 5, 20)
	_ = f.SetRowHeight(xlSheet, 6, 18)
	_ = f.SetRowHeight(xlSheet, 7, 6)
	return 7 // baris judul terakhir; spasi = 7+1 → header mulai baris 8
}

// writeTableHead menulis header tabel dua baris (baris ini yang diulang cetak).
func writeTableHead(f *excelize.File, top int, st *xlStyles) error {
	r2 := top + 1
	cell := func(c string, v string) error {
		if err := f.SetCellValue(xlSheet, c, v); err != nil {
			return err
		}
		return f.SetCellStyle(xlSheet, c, c, st.head)
	}
	merges := [][2]string{{"B" + itoa(top), "B" + itoa(r2)}, {"C" + itoa(top), "E" + itoa(top)},
		{"F" + itoa(top), "F" + itoa(r2)}, {"G" + itoa(top), "G" + itoa(r2)},
		{"H" + itoa(top), "H" + itoa(r2)}, {"I" + itoa(top), "K" + itoa(top)}}
	for _, m := range merges {
		if err := f.MergeCell(xlSheet, m[0], m[1]); err != nil {
			return err
		}
		if err := f.SetCellStyle(xlSheet, m[0], m[1], st.head); err != nil {
			return err
		}
	}
	for _, kv := range [][2]string{
		{"B" + itoa(top), "NO"}, {"C" + itoa(top), "URAIAN KEGIATAN"},
		{"F" + itoa(top), "JML"}, {"G" + itoa(top), "SAT"},
		{"H" + itoa(top), "HARGA SATUAN"}, {"I" + itoa(top), "PENAWARAN HARGA (Rp)"},
		{"I" + itoa(r2), "JASA"}, {"J" + itoa(r2), "MAT"}, {"K" + itoa(r2), "JML"},
	} {
		if err := cell(kv[0], kv[1]); err != nil {
			return err
		}
	}
	_ = f.SetRowHeight(xlSheet, top, 20)
	_ = f.SetRowHeight(xlSheet, r2, 16)
	return nil
}

// writeRows menulis seluruh baris main/sub point; balik baris data terakhir.
func writeRows(f *excelize.File, d *ExportData, start int, st *xlStyles) (int, error) {
	r := start
	for i := range d.Rows {
		it := &d.Rows[i]
		nameCell := it.Name
		if it.SubNo != "" {
			nameCell = it.SubNo + ". " + nameCell
			// sub point pada blok dgn main point terisi: rincian masuk uraian
			if it.EmptyNumbers && it.Qty > 0 {
				nameCell += qtyUnitSuffix(it.Qty, it.Unit)
			}
		}
		nameLines := wrappedLines(nameCell, uraianWidth)
		descText := joinDesc(it.Description, it.WeightNote)
		descLines := wrappedLines(descText, uraianWidth)
		lines := nameLines + descLines
		if lines < 1 {
			lines = 1
		}
		_ = f.SetRowHeight(xlSheet, r, float64(lines*12+8))

		bodyStyle, textStyle := st.cellCenter, st.subCell
		if it.Bold {
			textStyle = st.mainCell
			bodyStyle = st.mainCenter
		}

		_ = f.MergeCell(xlSheet, "C"+itoa(r), "E"+itoa(r))
		uraian := nameCell
		if descText != "" {
			uraian += "\n" + descText
		}
		_ = f.SetCellValue(xlSheet, "B"+itoa(r), it.No)
		_ = f.SetCellValue(xlSheet, "C"+itoa(r), uraian)

		_ = f.SetCellStyle(xlSheet, "B"+itoa(r), "B"+itoa(r), bodyStyle)
		_ = f.SetCellStyle(xlSheet, "C"+itoa(r), "C"+itoa(r), textStyle)

		// kolom angka dikosongkan untuk baris bertipe EmptyNumbers
		if it.EmptyNumbers {
			_ = f.SetCellStyle(xlSheet, "F"+itoa(r), "G"+itoa(r), bodyStyle)
			_ = f.SetCellStyle(xlSheet, "H"+itoa(r), "H"+itoa(r), bodyStyle)
			_ = f.SetCellStyle(xlSheet, "I"+itoa(r), "K"+itoa(r), bodyStyle)
			r++
			continue
		}

		_ = f.SetCellValue(xlSheet, "F"+itoa(r), qtyValue(it.Qty))
		_ = f.SetCellValue(xlSheet, "G"+itoa(r), it.Unit)
		_ = f.SetCellStyle(xlSheet, "F"+itoa(r), "G"+itoa(r), bodyStyle)
		if it.UnitPrice > 0 {
			_ = f.SetCellValue(xlSheet, "H"+itoa(r), it.UnitPrice)
			_ = f.SetCellStyle(xlSheet, "H"+itoa(r), "H"+itoa(r), moneyStyle(textStyle, st))
		} else {
			_ = f.SetCellStyle(xlSheet, "H"+itoa(r), "H"+itoa(r), bodyStyle)
		}
		writeMoney(f, r, it.ServiceTotal, it.MaterialTotal, it.Total, textStyle, st)
		r++
	}
	return r - 1, nil
}

// moneyStyle memilih gaya angka Rupiah biasa atau tebal (untuk main point).
func moneyStyle(base int, st *xlStyles) int {
	if base == st.mainCell {
		return st.moneyBold
	}
	return st.money
}

func writeMoney(f *excelize.File, r int, svc, mat, total int64, base int, st *xlStyles) {
	style := moneyStyle(base, st)
	for col, v := range map[string]int64{"I": svc, "J": mat, "K": total} {
		if v != 0 {
			_ = f.SetCellValue(xlSheet, col+itoa(r), v)
		}
		_ = f.SetCellStyle(xlSheet, col+itoa(r), col+itoa(r), style)
	}
}

// writeRekap menulis Sub Total Jasa/Material + Grand Total; balik baris terakhir rekap.
func writeRekap(f *excelize.File, d *ExportData, start int, st *xlStyles) (int, error) {
	r := start
	rows := []struct {
		label string
		col   string
		val   int64
		valSt int
	}{
		{"Sub Total Jasa", "I", d.SubtotalService, st.money},
		{"Sub Total Material", "J", d.SubtotalMaterial, st.money},
		{"GRAND TOTAL", "K", d.GrandTotal, st.grandVal},
	}
	for _, rw := range rows {
		_ = f.MergeCell(xlSheet, "F"+itoa(r), "H"+itoa(r))
		lbl := rw.label
		if lbl == "GRAND TOTAL" {
			lbl = "TOTAL"
		}
		_ = f.SetCellValue(xlSheet, "F"+itoa(r), lbl)
		_ = f.SetCellStyle(xlSheet, "F"+itoa(r), "H"+itoa(r), st.grandLabel)
		_ = f.SetCellValue(xlSheet, rw.col+itoa(r), rw.val)
		for _, c := range []string{"I", "J", "K"} {
			vs := st.money
			if c == rw.col {
				vs = rw.valSt
			}
			_ = f.SetCellStyle(xlSheet, c+itoa(r), c+itoa(r), vs)
		}
		_ = f.SetRowHeight(xlSheet, r, 18)
		r++
	}
	return r - 1, nil
}

func writeTerbilangNotes(f *excelize.File, d *ExportData, start int, st *xlStyles) {
	r := start
	if d.Document.Terbilang != "" {
		_ = f.MergeCell(xlSheet, "B"+itoa(r), "K"+itoa(r))
		_ = f.SetCellStyle(xlSheet, "B"+itoa(r), "K"+itoa(r), st.italic)
		_ = f.SetCellValue(xlSheet, "B"+itoa(r), "Terbilang : "+d.Document.Terbilang+".")
		_ = f.SetRowHeight(xlSheet, r, float64(wrappedLines("Terbilang : "+d.Document.Terbilang+".", 100)*12+8))
		r++
		// Baris kosong sebagai ruang tersendiri sebelum catatan SPH.
		if d.Document.Notes != "" {
			_ = f.SetRowHeight(xlSheet, r, 12)
			r++
		}
	}
	if n := d.Document.Notes; n != "" {
		_ = f.MergeCell(xlSheet, "B"+itoa(r), "K"+itoa(r))
		_ = f.SetCellStyle(xlSheet, "B"+itoa(r), "K"+itoa(r), st.cellWrap)
		_ = f.SetCellValue(xlSheet, "B"+itoa(r), "Catatan: "+n)
		_ = f.SetRowHeight(xlSheet, r, float64(wrappedLines("Catatan: "+n, 100)*12+8))
	}
}

func terbilangNoteRows(d *ExportData) int {
	n := 0
	if d.Document.Terbilang != "" {
		n++
		if d.Document.Notes != "" {
			n++
		}
	}
	if d.Document.Notes != "" {
		n++
	}
	return n
}

// writeSignature blok ttd kanan bawah: kota/tanggal, perusahaan, ruang, nama, jabatan.
func writeSignature(f *excelize.File, d *ExportData, start int, st *xlStyles) {
	dateLine := FormatDateID(d.Document.Date)
	if d.Company.City != "" {
		dateLine = d.Company.City + ", " + dateLine
	}
	lines := []struct {
		text  string
		style int
	}{
		{dateLine, st.sigLine},
		{d.Company.Name, st.sigLine},
		{"", st.sigLine}, {"", st.sigLine}, {"", st.sigLine},
		{d.Company.SignerName, st.sigName},
		{d.Company.SignerPosition, st.sigLine},
	}
	r := start
	for _, ln := range lines {
		_ = f.MergeCell(xlSheet, "I"+itoa(r), "K"+itoa(r))
		_ = f.SetCellStyle(xlSheet, "I"+itoa(r), "K"+itoa(r), ln.style)
		if ln.text != "" {
			_ = f.SetCellValue(xlSheet, "I"+itoa(r), ln.text)
		}
		_ = f.SetRowHeight(xlSheet, r, 15)
		r++
	}
}

// ===== util =====

func itoa(n int) string { return fmt.Sprintf("%d", n) }

func qtyValue(q float64) interface{} {
	if q == float64(int64(q)) {
		return int64(q)
	}
	return q
}

func joinDesc(desc, weight string) string {
	parts := make([]string, 0, 2)
	if desc != "" {
		parts = append(parts, desc)
	}
	if weight != "" {
		parts = append(parts, weight)
	}
	return strings.Join(parts, " ")
}

// wrappedLines estimasi jumlah baris teks pada lebar kolom tertentu.
func wrappedLines(text string, width int) int {
	if text == "" {
		return 0
	}
	total := 0
	for _, ln := range strings.Split(text, "\n") {
		n := len([]rune(ln))
		if n == 0 {
			total++
			continue
		}
		total += (n + width - 1) / width
	}
	if total < 1 {
		total = 1
	}
	return total
}
