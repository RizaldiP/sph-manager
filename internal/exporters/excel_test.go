package exporters

import (
	"strings"
	"testing"
)

func buildFixture() *ExportData {
	return BuildData(fixtureSphDoc(), sampleCompany())
}

func TestBuildExcelStructure(t *testing.T) {
	f, err := BuildExcel(buildFixture())
	if err != nil {
		t.Fatalf("BuildExcel gagal: %v", err)
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) != 1 || sheets[0] != "SPH" {
		t.Fatalf("sheet salah: %v", sheets)
	}

	cell := func(ref string) string {
		v, _ := f.GetCellValue("SPH", ref)
		return v
	}
	if !strings.Contains(cell("B4"), "SPH/GEI/VII/2026/001") || !strings.Contains(cell("B4"), "Rev 2") {
		t.Errorf("baris nomor dokumen salah: %q", cell("B4"))
	}
	if cell("B5") != "PENAWARAN REPAIR PLC" {
		t.Errorf("judul 1 salah: %q", cell("B5"))
	}
	if cell("B6") != "KM BAHARI TA.2026" {
		t.Errorf("judul 2 salah: %q", cell("B6"))
	}
	if cell("D1") != "PT. Ganesha Energi Indonesia" {
		t.Errorf("nama perusahaan kop salah: %q", cell("D1"))
	}

	// header tabel dua baris + merge-nya
	merges := map[string]bool{}
	mc, err := f.GetMergeCells("SPH")
	if err != nil {
		t.Fatalf("GetMergeCells gagal: %v", err)
	}
	for _, m := range mc {
		merges[m.GetStartAxis()+":"+m.GetEndAxis()] = true
	}
	for _, want := range []string{"C8:E8", "B8:B9", "F8:F9", "G8:G9", "H8:H9", "I8:K8"} {
		if !merges[want] {
			t.Errorf("merge %s tidak ada (dapat %v)", want, merges)
		}
	}
	if cell("I8") != "PENAWARAN HARGA (Rp)" || cell("J9") != "MAT" {
		t.Error("header tabel salah")
	}

	// baris data pertama (r10) = main point; r11 sub pertama
	if cell("B10") != "1" || !strings.HasPrefix(cell("C10"), "Repair PLC") {
		t.Errorf("baris data 1 salah: %q | %q", cell("B10"), cell("C10"))
	}
	val := cell("K10")
	if val != "Rp11,250,000" {
		t.Errorf("total roll-up main point salah: %q", val)
	}
	if !strings.Contains(cell("C11"), "Inspection") || !strings.Contains(cell("C11"), "[bobot 40%]") {
		t.Errorf("sub point/bobot tidak tampil: %q", cell("C11"))
	}

	// rekap: cari label Sub Total Jasa / Material / GRAND TOTAL
	found := map[string]string{}
	for r := 10; r <= 30; r++ {
		lbl := cell("F" + itoa(r))
		if lbl == "" {
			continue
		}
		switch lbl {
		case "Sub Total Jasa":
			found["svc"], _ = f.GetCellValue("SPH", "I"+itoa(r))
			found["svcRow"] = itoa(r)
		case "Sub Total Material":
			found["mat"], _ = f.GetCellValue("SPH", "J"+itoa(r))
		case "TOTAL":
			found["grand"], _ = f.GetCellValue("SPH", "K"+itoa(r))
		}
	}
	if found["svc"] != "Rp10,500,000" || found["mat"] != "Rp1,000,000" || found["grand"] != "Rp11,500,000" {
		t.Errorf("rekap salah: %+v", found)
	}

	// terbilang & catatan
	okTerbilang, okNotes := false, false
	for r := 10; r <= 40; r++ {
		b := cell("B" + itoa(r))
		if strings.HasPrefix(b, "Terbilang : ") && strings.HasSuffix(b, "Rupiah.") {
			okTerbilang = true
		}
		if strings.HasPrefix(b, "Catatan: ") {
			okNotes = true
		}
	}
	if !okTerbilang || !okNotes {
		t.Errorf("terbilang/catatan tidak ditemukan (%v/%v)", okTerbilang, okNotes)
	}

	// print titles & print area (defined names bawaan Excel)
	names := f.GetDefinedName()
	hasTitles, hasArea := false, false
	for _, dn := range names {
		if dn.Name == "_xlnm.Print_Titles" && strings.Contains(dn.RefersTo, "$") && dn.Scope == "SPH" {
			hasTitles = true
		}
		if dn.Name == "_xlnm.Print_Area" && strings.Contains(dn.RefersTo, "'SPH'!$B$1:$K$") {
			hasArea = true
		}
	}
	if !hasTitles || !hasArea {
		t.Errorf("print titles/area hilang (titles=%v area=%v): %+v", hasTitles, hasArea, names)
	}

	// blok tanda tangan kanan bawah
	sigOK := false
	for r := 20; r <= 60; r++ {
		v := cell("I" + itoa(r))
		if v == "Matawai" {
			sigOK = true
			break
		}
	}
	if !sigOK {
		t.Error("nama penandatangan tidak ditemukan di kolom I:K")
	}
}

func TestBuildExcelEmptyAndNoLogo(t *testing.T) {
	d := &ExportData{Company: sampleCompany()}
	f, err := BuildExcel(d)
	if err != nil {
		t.Fatalf("dokumen kosong harus tetap bisa dibuat: %v", err)
	}
	defer func() { _ = f.Close() }()
	v, _ := f.GetCellValue("SPH", "B5")
	if v != "SURAT PENAWARAN HARGA" {
		t.Errorf("judul default salah: %q", v)
	}
}
