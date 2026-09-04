// Package exporters menghasilkan dokumen SPH siap cetak (Excel & PDF)
// dari snapshot dokumen — FR-IE5/FR-IE6 (Phase 9).
package exporters

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// CompanyInfo profil perusahaan + penandatangan dari halaman Pengaturan.
type CompanyInfo struct {
	Name           string
	City           string
	Address        string
	LogoPath       string // kosong bila belum ada logo
	SignerName     string
	SignerPosition string
}

// DocumentInfo identitas dokumen SPH yang dicetak.
type DocumentInfo struct {
	Number       string
	Revision     int
	Date         time.Time
	ValidUntil   *time.Time
	Customer     string
	Vessel       string
	VesselNumber string
	Project      string
	Subject      string
	Reference    string
	Location     string
	PIC          string
	Notes        string
	Terbilang    string
}

// Row satu baris tabel penawaran; Level 0 = main point, 1 = sub point.
//
// Main point menampilkan nilai ROLL-UP (termasuk seluruh sub point-nya,
// sama seperti tampilan detail aplikasi); sub point menjorok dengan
// rincian nilainya sendiri.
//
// EmptyNumbers menandai baris yang kolom angkanya (JML, SAT, Harga Satuan,
// JASA, MAT, JML-total) sengaja dikosongkan dan hanya menampilkan No + Uraian:
//   - main point tanpa nilai → kolom angka dikosongkan;
//   - sub point pada blok dgn main point terisi → nilainya masuk ke uraian
//     ("(qty satuan)") dan semua kolom angka dikosongkan.
type Row struct {
	No            string  // nomor main point ("1"); kosong untuk sub point
	SubNo         string  // huruf sub point ("a"); kosong untuk main point
	Name          string  //
	Description   string  //
	WeightNote    string  // keterangan bobot utk mode PEMBOBOTAN, mis. "[bobot 25%]"
	Qty           float64 //
	Unit          string  //
	UnitPrice     int64   // harga satuan gabungan jasa+material; 0 → kolom kosong
	ServiceTotal  int64   //
	MaterialTotal int64   //
	Total         int64   //
	Bold          bool    //
	EmptyNumbers  bool    // kolom angka dikosongkan (lihat catatan di atas)
}

// ExportData sumber tunggal generator Excel dan PDF.
type ExportData struct {
	Company          CompanyInfo
	Document         DocumentInfo
	Rows             []Row
	SubtotalService  int64
	SubtotalMaterial int64
	GrandTotal       int64
}

var bulanIndonesia = [...]string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

// FormatDateID memformat tanggal gaya Indonesia: "13 Juli 2026".
func FormatDateID(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%d %s %d", t.Day(), bulanIndonesia[t.Month()-1], t.Year())
}

// TitleLine1 judul dokumen baris pertama (mengikuti pola referensi).
func (d *ExportData) TitleLine1() string {
	if s := d.Document.Subject; s != "" {
		return s
	}
	return "Surat Penawaran Harga"
}

// TitleLine2 baris kedua judul: kapal + tahun, atau nama proyek.
func (d *ExportData) TitleLine2() string {
	y := d.Document.Date.Year()
	if d.Document.Vessel != "" {
		return fmt.Sprintf("%s TA.%d", d.Document.Vessel, y)
	}
	if p := d.Document.Project; p != "" {
		return p
	}
	return strconv.Itoa(y)
}

// subLetter mengubah urutan sub point menjadi huruf gaya penomoran alfabetis:
// 0→"a", 25→"z", 26→"aa", 51→"az", dst (basis-26 bijective seperti kolom Excel).
func subLetter(idx int) string {
	if idx < 0 {
		return ""
	}
	n := idx + 1
	var rev []byte
	for n > 0 {
		n--
		rev = append(rev, byte('a'+n%26))
		n /= 26
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return string(rev)
}

// BuildData meratakan snapshot dokumen menjadi baris cetak siap digambar.
func BuildData(doc *models.SphDocument, company CompanyInfo) *ExportData {
	d := &ExportData{
		Company: company,
		Document: DocumentInfo{
			Number:     doc.DocumentNumber,
			Revision:   doc.Revision,
			Date:       doc.Date,
			ValidUntil: doc.ValidUntil,
			Project:    doc.ProjectName,
			Subject:    doc.Subject,
			Reference:  doc.Reference,
			Location:   doc.Location,
			PIC:        doc.PicName,
			Notes:      doc.Notes,
			Terbilang:  doc.Terbilang,
		},
		Rows:             make([]Row, 0, len(doc.Items)),
		SubtotalService:  doc.SubtotalService,
		SubtotalMaterial: doc.SubtotalMaterial,
		GrandTotal:       doc.GrandTotal,
	}
	if doc.Customer.Name != "" {
		d.Document.Customer = doc.Customer.Name
	}
	if doc.Vessel != nil {
		d.Document.Vessel = doc.Vessel.Name
		d.Document.VesselNumber = doc.Vessel.VesselNumber
	}

	for i := range doc.Items {
		it := &doc.Items[i]
		d.Rows = append(d.Rows, Row{
			No:            strconv.Itoa(it.Sequence),
			Name:          it.NameSnapshot,
			Description:   it.DescriptionSnapshot,
			Qty:           it.Quantity,
			Unit:          it.Unit,
			UnitPrice:     it.ServiceUnitPrice + it.MaterialUnitPrice,
			ServiceTotal:  it.ServiceTotal,
			MaterialTotal: it.MaterialTotal,
			Total:         it.Total,
			Bold:          true,
		})
		weighted := it.PricingMode == models.PricingModeWeight
		for j := range it.SubItems {
			sb := &it.SubItems[j]
			r := Row{
				SubNo:         subLetter(j),
				Name:          sb.NameSnapshot,
				Description:   sb.DescriptionSnapshot,
				Qty:           sb.Quantity,
				Unit:          sb.Unit,
				UnitPrice:     sb.ServiceUnitPrice + sb.MaterialUnitPrice,
				ServiceTotal:  sb.ServiceTotal,
				MaterialTotal: sb.MaterialTotal,
				Total:         sb.Total,
			}
			if weighted && sb.Weight > 0 {
				r.WeightNote = fmt.Sprintf("[bobot %d%%]", sb.Weight)
			}
			d.Rows = append(d.Rows, r)
		}
	}
	markEmptyNumberRows(d.Rows)
	return d
}

// markEmptyNumberRows mengoleskan flag EmptyNumbers pada baris sesuai aturan
// tampilan per blok (main point + sub point-nya yang berurutan):
//
//   - Main point terisi bila Qty>0 DAN UnitPrice>0. Saat terisi, seluruh sub
//     point-nya memakai gaya "rincian masuk uraian" → kolom angka dikosongkan.
//   - Bila main point tidak terisi, kolom angka main point itu sendiri yang
//     dikosongkan (hanya No + Uraian tampil); sub point tampil normal di kolom.
func markEmptyNumberRows(rows []Row) {
	i := 0
	for i < len(rows) {
		if !rows[i].Bold {
			i++
			continue
		}
		main := &rows[i]
		filled := main.Qty > 0 && main.UnitPrice > 0
		if !filled {
			main.EmptyNumbers = true
		}
		i++
		for i < len(rows) && !rows[i].Bold {
			if filled {
				rows[i].EmptyNumbers = true
			}
			i++
		}
	}
}

// qtyUnitSuffix membentuk akhiran "(qty satuan)" untuk sub point yang memakai
// gaya uraian (main point terisi). Bila qty <= 0 akhiran kosong; bila unit
// tidak diisi menjadi "(qty)".
func qtyUnitSuffix(q float64, unit string) string {
	if q <= 0 {
		return ""
	}
	s := qtyText(q)
	if u := strings.TrimSpace(unit); u != "" {
		s += " " + u
	}
	return " (" + s + ")"
}
