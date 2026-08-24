package importers

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ColumnMapping: pemetaan kolom sheet ke field master pekerjaan (FR-IE2).
// Kolom bernilai -1 berarti tidak dipetakan.
// HeaderRows = indeks baris pertama DATA (baris-sebelumnya dianggap header).
// NameSpan = jumlah kolom yang digabung menjadi satu uraian (layout Excel
// referensi menaruh teks level berbeda di kolom berbeda dengan indentasi).
type ColumnMapping struct {
	NameCol     int `json:"nameCol"`
	NameSpan    int `json:"nameSpan"`
	QtyCol      int `json:"qtyCol"`
	UnitCol     int `json:"unitCol"`
	ServiceCol  int `json:"serviceCol"`
	MaterialCol int `json:"materialCol"`
	// UnitPriceCol: kolom "Harga Satuan" umum — dipakai hanya bila kolom
	// JASA dan MATERIAL baris tersebut kosong (lihat layout referensi yang
	// mencampur kedua konvensi).
	UnitPriceCol  int    `json:"unitPriceCol"`
	UnitPriceAs   string `json:"unitPriceAs,omitempty"` // "service" | "material"
	ServiceTotal  bool   `json:"serviceTotal"`
	MaterialTotal bool   `json:"materialTotal"`
	HeaderRows    int    `json:"headerRows"`
	Notes         string `json:"notes,omitempty"`
}

// unitPriceAsMaterial: harga satuan dianggap harga material?
// Kosong/nilai lain berarti jasa (default).
func (m ColumnMapping) unitPriceAsMaterial() bool {
	return strings.EqualFold(strings.TrimSpace(m.UnitPriceAs), "material")
}

func (m ColumnMapping) span() int {
	if m.NameSpan < 1 {
		return 1
	}
	return m.NameSpan
}

// Level klasifikasi baris (diekspor untuk dipakai service import).
const (
	LevelMain    = "main"
	LevelSub     = "sub"
	LevelUnknown = "unknown"
)

const (
	levelUnknown = LevelUnknown
	levelMain    = LevelMain
	levelSub     = LevelSub
)

// PreviewRow: satu baris hasil parse untuk ditampilkan/diputuskan pengguna.
type PreviewRow struct {
	RowIndex      int      `json:"rowIndex"`
	Suggested     string   `json:"suggested"`
	Level         string   `json:"level"`
	Marker        string   `json:"marker"`
	Name          string   `json:"name"`
	Qty           float64  `json:"qty"`
	Unit          string   `json:"unit"`
	ServicePrice  int64    `json:"servicePrice"`
	MaterialPrice int64    `json:"materialPrice"`
	Raw           string   `json:"raw"`
	Errors        []string `json:"errors,omitempty"`
}

type markerKind int

const (
	mkNone markerKind = iota
	mkRoman
	mkNumber
	mkUpper
	mkLower
)

var romanValues = map[byte]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}

func validRoman(s string) bool {
	if s == "" {
		return false
	}
	total, prev := 0, 0
	for i := len(s) - 1; i >= 0; i-- {
		v, ok := romanValues[s[i]]
		if !ok {
			return false
		}
		if v < prev {
			total -= v
		} else {
			total += v
			prev = v
		}
	}
	return total > 0 && total < 4000
}

// splitMarker memisahkan penanda hierarki di awal teks (FR-IE3).
// Penanda wajib eksplisit agar tidak salah potong nama pekerjaan:
//
//	angka : "12. " / "12) "      (wajib titik/kurung, spasi tidak sah)
//	romawi: "XII. " / "VII "     (huruf romawi valid + pemisah)
//	huruf besar : "A. " / "A "   (grup, satu huruf)
//	huruf kecil : "a. " / "b) "  (wajib titik/kurung)
func splitMarker(text string) (markerKind, string, string) {
	t := strings.TrimLeft(text, " \t ")
	// Angka: hanya titik/kurung yang dipisah.
	if i := indexDigitSep(t); i > 0 && allDigits(t[:i]) {
		return mkNumber, t[:i], strings.TrimSpace(t[i+1:])
	}
	// Romawi atau huruf besar tunggal: diikuti titik/kurung/spasi.
	j := 0
	for j < len(t) && isUpperLetter(t[j]) {
		j++
	}
	if j > 0 && j < len(t) && isSep(t[j]) {
		token := t[:j]
		if validRoman(token) {
			return mkRoman, token, strings.TrimSpace(t[j+1:])
		}
		if j == 1 {
			return mkUpper, token, strings.TrimSpace(t[j+1:])
		}
	}
	// Huruf kecil.
	k := 0
	for k < len(t) && isLowerLetter(t[k]) {
		k++
	}
	if k > 0 && k <= 3 && k < len(t) && (t[k] == '.' || t[k] == ')') {
		return mkLower, t[:k], strings.TrimSpace(t[k+1:])
	}
	return mkNone, "", t
}

// indexDigitSep: posisi '.'/')' pertama, asalkan sebelumnya murni digit.
func indexDigitSep(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' || s[i] == ')' {
			return i
		}
		if !allDigits(s[:i+1]) {
			return -1
		}
	}
	return -1
}

func isSep(b byte) bool { return b == '.' || b == ')' || b == ' ' || b == '\t' }

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

func isUpperLetter(b byte) bool { return b >= 'A' && b <= 'Z' }
func isLowerLetter(b byte) bool { return b >= 'a' && b <= 'z' }

// mergedName menggabungkan kolom uraian menjadi satu teks.
// Dua penyesuaian layout Excel referensi:
//   - SATU kolom di kiri uraian ikut digabung: di sanalah huruf grup "A" atau
//     romawi "I" sering berdiri sendiri. Angka polos tanpa titik/kurung tetap
//     tidak akan dianggap penanda (aturan splitMarker), jadi kolom "NO"
//     berisi 1,2,3… aman.
//   - Beberapa kolom ke kanan ikut digabung (NameSpan): teks level berbeda
//     ditaruh pada indentasi kolom berbeda; sel kosong tidak ikut digabung.
func mergedName(g *Grid, ri int, m ColumnMapping) string {
	start := m.NameCol - 1
	if start < 0 {
		start = 0
	}
	end := m.NameCol + m.span()
	parts := make([]string, 0, end-start)
	for c := start; c < end && c < g.Width(); c++ {
		if v := g.Cell(ri, c); v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, " ")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ParseRows mengubah grid menjadi barisan pratinjau terklasifikasi.
// Aturan dua level (pekerjaan induk/sub) mengikuti analisis referensi:
// romawi/huruf-besar → induk; huruf kecil → sub. Untuk penanda ANGKA yang
// ambigu dipakai aturan deret: angka yang meneruskan deret induk terakhir
// (n == lastMain+1, mis. "2." setelah "1.") adalah induk baru; angka lain
// di dalam blok sub (mis. sub-sub "1. Sensor Oli") diratakan menjadi sub.
func ParseRows(g *Grid, m ColumnMapping) []PreviewRow {
	out := make([]PreviewRow, 0, g.Height())
	seenMain, seenSub, lastMainNum := false, false, 0

	for ri := m.HeaderRows; ri < g.Height(); ri++ {
		rawName := mergedName(g, ri, m)
		unit := g.Cell(ri, m.UnitCol)

		qtyStr := ""
		if m.QtyCol >= 0 {
			qtyStr = g.Cell(ri, m.QtyCol)
		}
		svcStr := ""
		if m.ServiceCol >= 0 {
			svcStr = g.Cell(ri, m.ServiceCol)
		}
		matStr := ""
		if m.MaterialCol >= 0 {
			matStr = g.Cell(ri, m.MaterialCol)
		}

		// Baris benar-benar kosong dilewati.
		if rawName == "" && strings.TrimSpace(qtyStr) == "" && strings.TrimSpace(svcStr) == "" &&
			strings.TrimSpace(matStr) == "" && unit == "" {
			continue
		}

		row := PreviewRow{RowIndex: ri, Raw: rawName, Unit: unit}
		qtyEmpty := strings.TrimSpace(qtyStr) == ""

		// Nama + penanda (penanda di kolom kiri ikut tergabung via mergedName).
		kind, marker, name := splitMarker(rawName)
		switch kind {
		case mkRoman, mkUpper:
			row.Suggested = levelMain
			seenMain, seenSub = true, false
		case mkNumber:
			n, nerr := strconv.Atoi(marker)
			switch {
			case nerr != nil || n <= 0 || n > 999:
				row.Suggested = levelUnknown
			case !seenMain || n == lastMainNum+1:
				row.Suggested = levelMain
				seenMain, seenSub, lastMainNum = true, false, n
			case seenSub:
				row.Suggested = levelSub
			default:
				row.Suggested = levelMain
				seenMain, seenSub, lastMainNum = true, false, n
			}
		case mkLower:
			row.Suggested = levelSub
			if seenMain {
				seenSub = true
			}
		default:
			row.Suggested = levelUnknown
		}
		row.Marker = marker
		row.Name = name
		if row.Name == "" {
			row.Errors = append(row.Errors, "Nama pekerjaan kosong")
		}

		// Qty: sel kosong wajar pada baris grup/induk → default 1.
		// Hanya nilai eksplisit yang tidak valid yang menjadi blokir.
		row.Qty = 1
		if m.QtyCol >= 0 && !qtyEmpty {
			if q, err := parseNumber(qtyStr); err != nil || q <= 0 {
				row.Errors = append(row.Errors, "Qty tidak valid: \""+qtyStr+"\"")
			} else {
				row.Qty = q
			}
		}

		// Harga: kolom referensi berisi NILAI TOTAL baris; dikonversi ke harga satuan.
		svc, err := parsePriceCell(svcStr, m.ServiceTotal, row.Qty)
		if err != nil {
			row.Errors = append(row.Errors, "Harga jasa tidak valid: "+err.Error())
		}
		row.ServicePrice = svc
		mat, err := parsePriceCell(matStr, m.MaterialTotal, row.Qty)
		if err != nil {
			row.Errors = append(row.Errors, "Harga material tidak valid: "+err.Error())
		}
		row.MaterialPrice = mat

		// Fallback kolom "Harga Satuan" umum: hanya dipakai bila kolom
		// JASA maupun MATERIAL baris ini tidak berisi nilai.
		if row.ServicePrice == 0 && row.MaterialPrice == 0 && m.UnitPriceCol >= 0 {
			if upStr := g.Cell(ri, m.UnitPriceCol); strings.TrimSpace(upStr) != "" {
				up, err := parsePriceCell(upStr, false, row.Qty)
				if err != nil {
					row.Errors = append(row.Errors, "Harga satuan tidak valid: "+err.Error())
				} else if up > 0 {
					if m.unitPriceAsMaterial() {
						row.MaterialPrice = up
					} else {
						row.ServicePrice = up
					}
				}
			}
		}

		row.Level = row.Suggested
		out = append(out, row)
	}
	return out
}

func parsePriceCell(s string, isTotal bool, qty float64) (int64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	v, err := parseNumber(s)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		return 0, fmt.Errorf("tidak boleh negatif")
	}
	if isTotal {
		if qty > 0 {
			v = v / qty
		} else {
			return 0, fmt.Errorf("qty nol untuk nilai total")
		}
	}
	return int64(math.Round(v)), nil
}

// parseNumber menormalkan angka gaya Indonesia/internasional yang umum di Excel:
// "120725000", "1.234.567", "1.234.567,89", "1,234,567.89".
func parseNumber(s string) (float64, error) {
	t := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\u00a0' {
			return -1
		}
		return r
	}, s)
	if t == "" {
		return 0, fmt.Errorf("kosong")
	}
	switch {
	case groupPatternDots.MatchString(t):
		t = strings.ReplaceAll(t, ".", "")
	case groupPatternCommas.MatchString(t):
		t = strings.ReplaceAll(t, ",", "")
	default:
		lastComma := strings.LastIndex(t, ",")
		lastDot := strings.LastIndex(t, ".")
		if lastComma > lastDot {
			t = strings.ReplaceAll(t, ".", "")
			t = strings.Replace(t, ",", ".", 1)
		} else {
			t = strings.ReplaceAll(t, ",", "")
		}
	}
	return strconv.ParseFloat(t, 64)
}

var (
	groupPatternDots   = lazyMust(`^\d{1,3}(\.\d{3})+$`)
	groupPatternCommas = lazyMust(`^\d{1,3}(,\d{3})+$`)
)

func lazyMust(pattern string) *regexp.Regexp { return regexp.MustCompile(pattern) }

// ===== Saran mapping otomatis dari baris header =====

// SuggestMapping memindai baris header (maksimal 15 baris pertama) dan
// mengusulkan pemetaan kolom. Header Excel sering membentang beberapa baris
// (judul gabungan + sub-judul, mis. "PENAWARAN HARGA" / "JASA | MAT"), jadi
// baris bersebelahan yang sama-sama mengandung kata kunci digabung menjadi
// satu band; tiap jenis kolom diambil dari kemunculan pertama dalam band.
func SuggestMapping(g *Grid) (ColumnMapping, []string) {
	m := ColumnMapping{NameCol: -1, NameSpan: 3, QtyCol: -1, UnitCol: -1, ServiceCol: -1, MaterialCol: -1,
		UnitPriceCol: -1, UnitPriceAs: "service",
		ServiceTotal: true, MaterialTotal: true, HeaderRows: 1}
	var notes []string

	limit := min(15, g.Height())
	width := g.Width()

	// Pilih band baris header bersebelahan dengan skor kumulatif tertinggi.
	bestStart, bestEnd, bestScore := -1, -1, 0
	curStart, curScore := 0, 0
	for r := 0; r <= limit; r++ {
		score := 0
		if r < limit {
			for c := 0; c < width; c++ {
				if headerKind(g.Cell(r, c)) != hkNone {
					score++
				}
			}
		}
		if r < limit && score > 0 {
			if curScore == 0 {
				curStart = r
			}
			curScore += score
			continue
		}
		// Band berakhir di r-1.
		if curScore > bestScore {
			bestScore, bestStart, bestEnd = curScore, curStart, r-1
		}
		curScore = 0
	}
	if bestStart < 0 {
		notes = append(notes, "Baris header tidak dikenali otomatis; petakan kolom secara manual.")
		return m, notes
	}

	band := make([]string, 0, bestEnd-bestStart+1)
	for c := 0; c < width; c++ {
		band = band[:0]
		for r := bestStart; r <= bestEnd; r++ {
			if v := g.Cell(r, c); v != "" {
				band = append(band, v)
			}
		}
		switch bandColumnKind(band) {
		case hkName:
			if m.NameCol < 0 {
				m.NameCol = c
			}
		case hkQty:
			if m.QtyCol < 0 {
				m.QtyCol = c
			}
		case hkUnit:
			if m.UnitCol < 0 {
				m.UnitCol = c
			}
		case hkUnitPrice:
			if m.UnitPriceCol < 0 {
				m.UnitPriceCol = c
			}
		case hkService:
			if m.ServiceCol < 0 {
				m.ServiceCol = c
			}
		case hkMaterial:
			if m.MaterialCol < 0 {
				m.MaterialCol = c
			}
		}
	}
	m.HeaderRows = bestEnd + 1

	// Lewati baris indeks kolom ("1 2 3 4 …") sesudah band header.
	for m.HeaderRows < g.Height() && isIndexRow(g, m.HeaderRows) {
		m.HeaderRows++
	}

	if m.NameCol < 0 {
		notes = append(notes, "Kolom URAIAN KEGIATAN tidak ditemukan; pilih manual.")
	}
	return m, notes
}

// isIndexRow: semua sel berisi angka polos pendek (penomoran kolom).
func isIndexRow(g *Grid, r int) bool {
	found := false
	for c := 0; c < g.Width(); c++ {
		v := g.Cell(r, c)
		if v == "" {
			continue
		}
		if len(v) > 2 || !allDigits(v) {
			return false
		}
		found = true
	}
	return found
}

// bandColumnKind menentukan jenis kolom dari seluruh sel band header-nya.
// Teks gabungan dipakai untuk membedakan "HARGA"+"SATUAN" (harga satuan
// umum, sering terbentang dua baris) dari kolom SATUAN biasa maupun dari
// kolom JASA/MATERIAL; tanpa itu, jenis diambil dari sel ber-kata kunci pertama.
func bandColumnKind(cells []string) headerKindT {
	joined := strings.ToUpper(strings.Join(cells, " "))
	if strings.Contains(joined, "HARGA") && strings.Contains(joined, "SATUAN") {
		return hkUnitPrice
	}
	for _, v := range cells {
		if k := headerKind(v); k != hkNone {
			return k
		}
	}
	return hkNone
}

type headerKindT int

const (
	hkNone headerKindT = iota
	hkName
	hkQty
	hkUnit
	hkUnitPrice
	hkService
	hkMaterial
)

func headerKind(s string) headerKindT {
	u := strings.ToUpper(strings.TrimSpace(s))
	if u == "" {
		return hkNone
	}
	switch {
	case strings.Contains(u, "URAIAN"), strings.Contains(u, "KEGIATAN"), u == "PEKERJAAN":
		return hkName
	case u == "JML", u == "QTY", strings.Contains(u, "JUMLAH"):
		return hkQty
	case u == "SAT", u == "UNIT", strings.Contains(u, "SATUAN"):
		return hkUnit
	case strings.Contains(u, "JASA"):
		return hkService
	case strings.Contains(u, "MATERIAL"), u == "MAT":
		return hkMaterial
	}
	return hkNone
}

// CountLevels menghitung ringkasan level saran untuk UI.
func CountLevels(rows []PreviewRow) (mainCount, subCount, unknownCount int) {
	for _, r := range rows {
		switch r.Suggested {
		case levelMain:
			mainCount++
		case levelSub:
			subCount++
		default:
			unknownCount++
		}
	}
	return
}
