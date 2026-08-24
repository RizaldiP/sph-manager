package services

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// TestTemplateScenarioFromPlan adalah skenario acceptance Phase 4:
// buat template "Repair PLC" berisi 8 langkah → duplikat → nonaktifkan.
func TestTemplateScenarioFromPlan(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	svc := NewTemplateService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	stepNames := []string{
		"Inspection", "Troubleshooting", "Wiring Check", "Component Replacement",
		"Firmware Update", "Calibration", "Testing", "Commissioning",
	}
	steps := make([]*models.WorkItem, 0, len(stepNames))
	for i, n := range stepNames {
		wi, err := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-1" + string(rune('0'+i)), Name: n, DefaultQuantity: 1})
		if err != nil {
			t.Fatalf("create pekerjaan %s gagal: %v", n, err)
		}
		steps = append(steps, wi)
	}

	tpl, err := svc.Create(&models.Template{Code: "TPL-PLC", Name: "  Repair PLC  ", Description: "Paket perbaikan PLC"})
	if err != nil {
		t.Fatalf("create template gagal: %v", err)
	}
	if tpl.Name != "Repair PLC" || !tpl.IsActive || tpl.Sequence != 1 || tpl.ItemCount != 0 {
		t.Errorf("default template salah: %+v", tpl)
	}

	inputs := make([]TemplateItemInput, 0, len(steps))
	for _, s := range steps {
		inputs = append(inputs, TemplateItemInput{WorkItemID: s.ID})
	}
	detail, err := svc.SetItems(tpl.ID, inputs)
	if err != nil {
		t.Fatalf("set items gagal: %v", err)
	}
	if len(detail.Items) != len(stepNames) {
		t.Fatalf("jumlah item salah: %d", len(detail.Items))
	}
	for i, it := range detail.Items {
		if it.Sequence != i+1 || it.WorkItem.Name != stepNames[i] {
			t.Errorf("urutan item salah index %d: seq=%d name=%s", i, it.Sequence, it.WorkItem.Name)
		}
	}

	list, _ := svc.List(false, "")
	if list[0].ItemCount != int64(len(stepNames)) {
		t.Errorf("item count di daftar salah: %d", list[0].ItemCount)
	}

	// Duplikat: nama bersuffix, isi sama urutannya, kode dikosongkan.
	copyView, err := svc.Duplicate(tpl.ID)
	if err != nil {
		t.Fatalf("duplicate gagal: %v", err)
	}
	if copyView.Name != "Repair PLC - Salinan" || copyView.ItemCount != int64(len(stepNames)) {
		t.Errorf("hasil duplikat salah: %+v", copyView)
	}
	if copyView.Code == "" || copyView.Code == tpl.Code {
		t.Errorf("salinan harus dapat kode otomatis baru, dapat %q", copyView.Code)
	}
	srcDetail, _ := svc.Get(tpl.ID)
	dstDetail, err := svc.Get(copyView.ID)
	if err != nil {
		t.Fatalf("ambil salinan gagal: %v", err)
	}
	for i := range srcDetail.Items {
		if dstDetail.Items[i].WorkItemID != srcDetail.Items[i].WorkItemID || dstDetail.Items[i].Sequence != srcDetail.Items[i].Sequence {
			t.Errorf("urutan salinan beda pada index %d", i)
		}
	}
	if copyView.Sequence != 2 {
		t.Errorf("salinan harus dapat sequence 2, dapat %d", copyView.Sequence)
	}

	// Nonaktifkan aslinya; hanya salinan yang tampil di daftar default.
	if err := svc.SetActive(tpl.ID, false); err != nil {
		t.Fatalf("nonaktifkan gagal: %v", err)
	}
	activeOnly, _ := svc.List(false, "")
	if len(activeOnly) != 1 || activeOnly[0].ID != copyView.ID {
		t.Errorf("daftar aktif salah: %+v", activeOnly)
	}
	all, _ := svc.List(true, "")
	if len(all) != 2 {
		t.Errorf("includeInactive gagal: dapat %d, mau 2", len(all))
	}

	// Hapus salinan tidak masalah walau berisi banyak item.
	if err := svc.Delete(copyView.ID); err != nil {
		t.Fatalf("hapus template gagal: %v", err)
	}
	var count int64
	db.Model(&models.Template{}).Where("id = ?", copyView.ID).Count(&count)
	if count != 0 {
		t.Error("soft delete template tidak bekerja")
	}

	var auditCount int64
	db.Model(&models.AuditLog{}).Where("entity = ?", "template").Count(&auditCount)
	if auditCount == 0 {
		t.Error("audit log template tidak terekam")
	}
}

func TestTemplateValidation(t *testing.T) {
	db := serviceDB(t)
	svc := NewTemplateService(db, slog.Default())

	if _, err := svc.Create(&models.Template{Name: "   "}); err == nil {
		t.Fatal("nama kosong seharusnya ditolak")
	}
	if _, err := svc.Create(&models.Template{Name: strings.Repeat("x", 201)}); err == nil {
		t.Fatal("nama >200 seharusnya ditolak")
	}

	first, _ := svc.Create(&models.Template{Code: "TPL-01", Name: "Repair AMS"})
	if _, err := svc.Create(&models.Template{Code: "TPL-01", Name: "Duplikat"}); err == nil {
		t.Fatal("kode duplikat seharusnya ditolak")
	} else if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}

	upd, err := svc.Update(first.ID, &models.Template{Code: "TPL-01", Name: "Repair AMS 2", Notes: "catatan"})
	if err != nil {
		t.Fatalf("update header gagal: %v", err)
	}
	if upd.Name != "Repair AMS 2" || upd.Notes != "catatan" {
		t.Errorf("update tidak tersimpan: %+v", upd)
	}

	if _, err := svc.Update(9999, &models.Template{Name: "Hantu"}); err != ErrNotFound {
		t.Errorf("template hantu seharusnya ErrNotFound, dapat: %v", err)
	}
}

func TestTemplateSetItemsValidation(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	svc := NewTemplateService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	w1, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Satu", DefaultQuantity: 1})
	w2, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-002", Name: "Dua", DefaultQuantity: 1})

	tpl, _ := svc.Create(&models.Template{Name: "Repair PLC"})
	tpl2, _ := svc.Create(&models.Template{Name: "Lain"})

	// Pekerjaan sama dua kali ditolak.
	if _, err := svc.SetItems(tpl.ID, []TemplateItemInput{{WorkItemID: w1.ID}, {WorkItemID: w1.ID}}); err == nil {
		t.Fatal("pekerjaan duplikat dalam satu template seharusnya ditolak")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("harus ValidationError, dapat: %v", err)
	}

	// ID pekerjaan hantu ditolak.
	if _, err := svc.SetItems(tpl.ID, []TemplateItemInput{{WorkItemID: 9999}}); err == nil {
		t.Fatal("pekerjaan hantu seharusnya ditolak")
	}

	// WorkItemID kosong ditolak.
	if _, err := svc.SetItems(tpl.ID, []TemplateItemInput{{WorkItemID: 0}}); err == nil {
		t.Fatal("workItemId kosong seharusnya ditolak")
	}

	// Catatan terlalu panjang ditolak.
	if _, err := svc.SetItems(tpl.ID, []TemplateItemInput{{WorkItemID: w1.ID, Notes: strings.Repeat("n", 501)}}); err == nil {
		t.Fatal("catatan >500 seharusnya ditolak")
	}

	// Set ulang isi (ganti urutan) menggantikan seluruh baris lama.
	if _, err := svc.SetItems(tpl.ID, []TemplateItemInput{{WorkItemID: w2.ID}, {WorkItemID: w1.ID}}); err != nil {
		t.Fatalf("set items valid gagal: %v", err)
	}
	detail, _ := svc.Get(tpl.ID)
	if len(detail.Items) != 2 || detail.Items[0].WorkItemID != w2.ID {
		t.Errorf("replace items salah: %+v", detail.Items)
	}

	// Mengosongkan template diperbolehkan.
	if _, err := svc.SetItems(tpl2.ID, []TemplateItemInput{}); err != nil {
		t.Fatalf("kosongkan template gagal: %v", err)
	}
	empty, _ := svc.Get(tpl2.ID)
	if len(empty.Items) != 0 {
		t.Errorf("template harus kosong, dapat %d item", len(empty.Items))
	}
}

// TestWorkItemDeleteBlockedByTemplate memastikan pekerjaan yang masih dipakai
// template hidup tidak bisa dihapus (integritas FK RESTRICT).
func TestWorkItemDeleteBlockedByTemplate(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	svc := NewTemplateService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	wi, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Inspection", DefaultQuantity: 1})
	tpl, _ := svc.Create(&models.Template{Name: "Repair PLC"})
	if _, err := svc.SetItems(tpl.ID, []TemplateItemInput{{WorkItemID: wi.ID}}); err != nil {
		t.Fatalf("set items gagal: %v", err)
	}

	err := wiSvc.Delete(wi.ID)
	if err == nil {
		t.Fatal("hapus pekerjaan yang dipakai template seharusnya ditolak")
	}
	if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}

	// Setelah dilepas dari semua template, hapus boleh.
	if _, err := svc.SetItems(tpl.ID, []TemplateItemInput{}); err != nil {
		t.Fatalf("kosongkan template gagal: %v", err)
	}
	if err := wiSvc.Delete(wi.ID); err != nil {
		t.Fatalf("hapus pekerjaan setelah lepas dari template gagal: %v", err)
	}
}

// TestTemplatesWithoutCodeCoexist adalah regresi untuk index unik parsial:
// beberapa template tanpa kode (kode kosong) harus bisa hidup berdampingan.
func TestTemplatesWithoutCodeCoexist(t *testing.T) {
	db := serviceDB(t)
	svc := NewTemplateService(db, slog.Default())

	a, err := svc.Create(&models.Template{Name: "Tanpa Kode Satu"})
	if err != nil {
		t.Fatalf("template pertama tanpa kode gagal: %v", err)
	}
	b, err := svc.Create(&models.Template{Name: "Tanpa Kode Dua"})
	if err != nil {
		t.Fatalf("dua template tanpa kode harus boleh: %v", err)
	}
	if a.ID == b.ID {
		t.Fatal("ID tidak boleh sama")
	}
}

// TestAutoGenerateCodes memastikan kode dibuat sistem bila tidak diisi manual,
// berlaku untuk keempat entitas master (FR-U3), dan update tanpa kode tidak menghapus kode lama.
func TestAutoGenerateCodes(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	subSvc := NewWorkSubItemService(db, slog.Default())
	tplSvc := NewTemplateService(db, slog.Default())

	c1, err := catSvc.Create(&models.Category{Name: "Electrical"})
	if err != nil {
		t.Fatalf("kategori tanpa kode gagal: %v", err)
	}
	c2, _ := catSvc.Create(&models.Category{Name: "Mechanical"})
	if c1.Code != "KAT-001" || c2.Code != "KAT-002" {
		t.Errorf("kode kategori otomatis salah: %q, %q", c1.Code, c2.Code)
	}

	w1, err := wiSvc.Create(&models.WorkItem{CategoryID: c1.ID, Name: "Repair Control Panel", DefaultQuantity: 1})
	if err != nil {
		t.Fatalf("pekerjaan tanpa kode gagal: %v", err)
	}
	w2, _ := wiSvc.Create(&models.WorkItem{CategoryID: c1.ID, Name: "Rewiring Motor", DefaultQuantity: 1})
	if w1.Code != "PEK-001" || w2.Code != "PEK-002" {
		t.Errorf("kode pekerjaan otomatis salah: %q, %q", w1.Code, w2.Code)
	}

	if _, err := subSvc.Create(&models.WorkSubItem{WorkItemID: w1.ID, Name: "Inspection", DefaultQuantity: 1}); err != nil {
		t.Fatalf("create sub gagal: %v", err)
	}
	if _, err := subSvc.Create(&models.WorkSubItem{WorkItemID: w1.ID, Name: "Testing", DefaultQuantity: 1}); err != nil {
		t.Fatalf("create sub kedua gagal: %v", err)
	}
	detail, _ := wiSvc.GetDetail(w1.ID)
	if len(detail.SubItems) != 2 {
		t.Fatalf("jumlah sub salah: %d", len(detail.SubItems))
	}
	if detail.SubItems[0].Code != "SUB-001" || detail.SubItems[1].Code != "SUB-002" {
		t.Errorf("kode sub otomatis salah: %q, %q", detail.SubItems[0].Code, detail.SubItems[1].Code)
	}

	t1, err := tplSvc.Create(&models.Template{Name: "Repair PLC"})
	if err != nil {
		t.Fatalf("template tanpa kode gagal: %v", err)
	}
	t2, _ := tplSvc.Create(&models.Template{Name: "Repair AMS"})
	if t1.Code != "TPL-001" || t2.Code != "TPL-002" {
		t.Errorf("kode template otomatis salah: %q, %q", t1.Code, t2.Code)
	}

	// Update tanpa kode harus mempertahankan kode lama.
	upd, err := wiSvc.Update(w1.ID, &models.WorkItem{CategoryID: c1.ID, Name: "Repair Main Panel", DefaultQuantity: 2})
	if err != nil {
		t.Fatalf("update tanpa kode gagal: %v", err)
	}
	if upd.Code != "PEK-001" {
		t.Errorf("kode lama hilang saat update: %q", upd.Code)
	}

	// Kode manual tetap boleh dan unik.
	m, err := catSvc.Create(&models.Category{Code: "ELX", Name: "Electrical X"})
	if err != nil {
		t.Fatalf("kode manual gagal: %v", err)
	}
	if m.Code != "ELX" {
		t.Errorf("kode manual berubah: %q", m.Code)
	}
	if _, err := catSvc.Create(&models.Category{Code: "ELX", Name: "Dup"}); err == nil {
		t.Fatal("kode manual duplikat seharusnya ditolak")
	}

	// Nomor berlanjut dari suffix numerik terbesar, mengabaikan kode non-numerik.
	x, _ := catSvc.Create(&models.Category{Code: "KAT-ZZ", Name: "Non Numerik"})
	n, err := catSvc.Create(&models.Category{Name: "Berikutnya"})
	if err != nil {
		t.Fatalf("create setelah kode non-numerik gagal: %v", err)
	}
	if n.Code != "KAT-003" {
		t.Errorf("nomor lanjutan salah: %q (setelah %q dan %q)", n.Code, x.Code, c2.Code)
	}

	var auditCount int64
	db.Model(&models.AuditLog{}).Count(&auditCount)
	if auditCount == 0 {
		t.Error("audit log tidak terekam")
	}
}

func TestTemplateReorderPersists(t *testing.T) {
	db := serviceDB(t)
	svc := NewTemplateService(db, slog.Default())

	a, _ := svc.Create(&models.Template{Code: "A", Name: "Alpha"})
	b, _ := svc.Create(&models.Template{Code: "B", Name: "Beta"})
	c, _ := svc.Create(&models.Template{Code: "C", Name: "Gamma"})

	if err := svc.Reorder([]uint{c.ID, a.ID}); err == nil {
		t.Fatal("reorder parsial seharusnya ditolak")
	}
	if err := svc.Reorder([]uint{c.ID, b.ID, a.ID}); err != nil {
		t.Fatalf("reorder gagal: %v", err)
	}
	list, _ := svc.List(false, "")
	want := []string{"Gamma", "Beta", "Alpha"}
	for i, v := range list {
		if v.Name != want[i] || v.Sequence != i+1 {
			t.Errorf("urutan salah index %d: %s seq=%d", i, v.Name, v.Sequence)
		}
	}
	next, _ := svc.Create(&models.Template{Code: "D", Name: "Delta"})
	if next.Sequence != 4 {
		t.Errorf("template baru harus sequence 4, dapat %d", next.Sequence)
	}
}
