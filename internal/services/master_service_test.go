package services

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/models"
)

func serviceDB(t *testing.T) *gorm.DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	lg := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(path, lg)
	if err != nil {
		t.Fatalf("Open gagal: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	t.Cleanup(func() {
		database.Close(db)
	})
	return db
}

func TestCategoryCRUDAndDuplicateCode(t *testing.T) {
	db := serviceDB(t)
	svc := NewCategoryService(db, slog.Default())

	cat, err := svc.Create(&models.Category{Code: "EL", Name: "  Electrical  ", Description: "Listrik"})
	if err != nil {
		t.Fatalf("create kategori gagal: %v", err)
	}
	if cat.Name != "Electrical" {
		t.Errorf("nama tidak di-trim: %q", cat.Name)
	}
	if cat.Sequence != 1 || !cat.IsActive || cat.WorkItemCount != 0 {
		t.Errorf("nilai default salah: seq=%d active=%v count=%d", cat.Sequence, cat.IsActive, cat.WorkItemCount)
	}

	if _, err := svc.Create(&models.Category{Code: "EL", Name: "Duplikat"}); err == nil {
		t.Fatal("kode duplikat seharusnya ditolak")
	} else if _, ok := err.(*ConflictError); !ok {
		t.Errorf("error yang diharapkan ConflictError, dapat: %v", err)
	}

	upd, err := svc.Update(cat.ID, &models.Category{Code: "EL", Name: "Electrical & Instrument"})
	if err != nil {
		t.Fatalf("update kategori gagal: %v", err)
	}
	if upd.Name != "Electrical & Instrument" {
		t.Errorf("update tidak tersimpan: %q", upd.Name)
	}

	if err := svc.SetActive(cat.ID, false); err != nil {
		t.Fatalf("set inactive gagal: %v", err)
	}
	list, err := svc.List(false, "")
	if err != nil {
		t.Fatalf("list gagal: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("kategori nonaktif tidak boleh muncul saat includeInactive=false")
	}
	listAll, err := svc.List(true, "")
	if err != nil {
		t.Fatalf("list(includeInactive) gagal: %v", err)
	}
	if len(listAll) != 1 || listAll[0].IsActive {
		t.Errorf("kategori nonaktif harus muncul saat includeInactive=true")
	}

	if err := svc.Delete(cat.ID); err != nil {
		t.Fatalf("hapus kategori kosong gagal: %v", err)
	}
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count != 0 {
		t.Error("soft delete tidak bekerja")
	}
}

func TestCategoryDeleteBlockedWhenHasWorkItems(t *testing.T) {
	db := serviceDB(t)
	svc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())

	cat, _ := svc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	wi, err := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Repair Control Panel", DefaultQuantity: 1})
	if err != nil {
		t.Fatalf("create pekerjaan gagal: %v", err)
	}

	err = svc.Delete(cat.ID)
	if err == nil {
		t.Fatal("hapus kategori berisi pekerjaan seharusnya ditolak")
	}
	if !strings.Contains(err.Error(), "pekerjaan") {
		t.Errorf("pesan error tidak ramah: %v", err)
	}

	views, _ := svc.List(false, "")
	if views[0].WorkItemCount != 1 {
		t.Errorf("jumlah pekerjaan salah: %d", views[0].WorkItemCount)
	}

	if err := wiSvc.Delete(wi.ID); err != nil {
		t.Fatalf("hapus pekerjaan gagal: %v", err)
	}
	if err := svc.Delete(cat.ID); err != nil {
		t.Fatalf("hapus kategori setelah kosong gagal: %v", err)
	}
}

func TestCategoryReorderPersists(t *testing.T) {
	db := serviceDB(t)
	svc := NewCategoryService(db, slog.Default())

	a, _ := svc.Create(&models.Category{Code: "A", Name: "Alpha"})
	b, _ := svc.Create(&models.Category{Code: "B", Name: "Beta"})
	c, _ := svc.Create(&models.Category{Code: "C", Name: "Gamma"})

	err := svc.Reorder([]uint{c.ID, a.ID})
	if err == nil {
		t.Fatal("reorder parsial (tidak semua id) seharusnya ditolak")
	}

	if err := svc.Reorder([]uint{c.ID, b.ID, a.ID}); err != nil {
		t.Fatalf("reorder gagal: %v", err)
	}
	list, _ := svc.List(false, "")
	want := []string{"Gamma", "Beta", "Alpha"}
	for i, v := range list {
		if v.Name != want[i] {
			t.Errorf("urutan salah index %d: dapat %s, mau %s", i, v.Name, want[i])
		}
		if v.Sequence != i+1 {
			t.Errorf("sequence salah index %d: %d", i, v.Sequence)
		}
	}

	next, err := svc.Create(&models.Category{Code: "D", Name: "Delta"})
	if err != nil {
		t.Fatalf("create setelah reorder gagal: %v", err)
	}
	if next.Sequence != 4 {
		t.Errorf("kategori baru harus dapat sequence 4, dapat %d", next.Sequence)
	}
}

func TestWorkItemValidationAndCRUD(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	svc := NewWorkItemService(db, slog.Default())

	if _, err := svc.Create(&models.WorkItem{Name: "Tanpa Kategori"}); err == nil {
		t.Fatal("tanpa kategori seharusnya ditolak")
	}
	if _, err := svc.Create(&models.WorkItem{CategoryID: 9999, Name: "Kategori Hantu"}); err == nil {
		t.Fatal("kategori hantu seharusnya ditolak")
	}

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})

	if _, err := svc.Create(&models.WorkItem{CategoryID: cat.ID, Name: ""}); err == nil {
		t.Fatal("nama kosong seharusnya ditolak")
	}
	if _, err := svc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "X", DefaultQuantity: 0}); err == nil {
		t.Fatal("qty 0 seharusnya ditolak")
	}
	if _, err := svc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "X", DefaultServicePrice: -1}); err == nil {
		t.Fatal("harga negatif seharusnya ditolak")
	}

	wi, err := svc.Create(&models.WorkItem{
		CategoryID:           cat.ID,
		Code:                 "EL-001",
		Name:                 "Repair Control Panel",
		DefaultUnit:          "giat",
		DefaultQuantity:      1,
		DefaultServicePrice:  5000000,
		DefaultMaterialPrice: 1500000,
	})
	if err != nil {
		t.Fatalf("create pekerjaan valid gagal: %v", err)
	}
	if wi.Sequence != 1 || !wi.IsActive {
		t.Errorf("default salah: seq=%d active=%v", wi.Sequence, wi.IsActive)
	}

	if _, err := svc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Duplikat Kode"}); err == nil {
		t.Fatal("kode duplikat antar pekerjaan seharusnya ditolak")
	}

	upd, err := svc.Update(wi.ID, &models.WorkItem{
		CategoryID:      cat.ID,
		Code:            "EL-001",
		Name:            "Repair Main Switchboard",
		DefaultUnit:     "giat",
		DefaultQuantity: 2,
	})
	if err != nil {
		t.Fatalf("update pekerjaan gagal: %v", err)
	}
	if upd.Name != "Repair Main Switchboard" || upd.DefaultQuantity != 2 || upd.DefaultServicePrice != 0 {
		t.Errorf("update tidak sesuai: %+v", upd)
	}

	if err := svc.SetActive(wi.ID, false); err != nil {
		t.Fatalf("nonaktifkan gagal: %v", err)
	}
	got, err := svc.GetDetail(wi.ID)
	if err != nil {
		t.Fatalf("detail gagal: %v", err)
	}
	if got.IsActive {
		t.Error("status aktif tidak tersimpan")
	}

	var auditCount int64
	db.Model(&models.AuditLog{}).Where("entity IN ?", []string{"category", "work_item"}).Count(&auditCount)
	if auditCount == 0 {
		t.Error("audit log tidak terekam")
	}
}

func TestWorkItemSearchAndFilter(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	svc := NewWorkItemService(db, slog.Default())

	el, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	mc, _ := catSvc.Create(&models.Category{Code: "MC", Name: "Mechanical"})
	svc.Create(&models.WorkItem{CategoryID: el.ID, Code: "EL-001", Name: "Repair Control Panel", DefaultQuantity: 1})
	svc.Create(&models.WorkItem{CategoryID: el.ID, Code: "EL-002", Name: "Rewiring Motor", DefaultQuantity: 1})
	svc.Create(&models.WorkItem{CategoryID: mc.ID, Code: "MC-001", Name: "Overhaul Pompa", DefaultQuantity: 1})
	hidden, _ := svc.Create(&models.WorkItem{CategoryID: el.ID, Code: "EL-003", Name: "Panel Lampu", DefaultQuantity: 1})
	svc.SetActive(hidden.ID, false)

	all, _ := svc.List(0, false, "")
	if len(all) != 3 {
		t.Errorf("filter nonaktif gagal: dapat %d item, mau 3", len(all))
	}
	inactiveToo, _ := svc.List(0, true, "")
	if len(inactiveToo) != 4 {
		t.Errorf("includeInactive gagal: dapat %d item, mau 4", len(inactiveToo))
	}
	byCat, _ := svc.List(mc.ID, false, "")
	if len(byCat) != 1 || byCat[0].Code != "MC-001" {
		t.Errorf("filter kategori gagal: %+v", byCat)
	}
	found, _ := svc.List(el.ID, false, "panel")
	if len(found) != 1 || found[0].Name != "Repair Control Panel" {
		t.Errorf("search gagal: %+v", found)
	}
	byCode, _ := svc.List(0, false, "EL-002")
	if len(byCode) != 1 || byCode[0].Name != "Rewiring Motor" {
		t.Errorf("search by code gagal: %+v", byCode)
	}
}

// TestSubItemScenarioFromPlan adalah skenario acceptance Phase 3:
// Electrical → Repair Control Panel → Inspection, Troubleshooting, Wiring Check,
// Component Replacement, Testing, Commissioning → ubah urutan → nonaktifkan.
func TestSubItemScenarioFromPlan(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	subSvc := NewWorkSubItemService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	wi, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Repair Control Panel", DefaultQuantity: 1})

	names := []string{"Inspection", "Troubleshooting", "Wiring Check", "Component Replacement", "Testing", "Commissioning"}
	for _, n := range names {
		if _, err := subSvc.Create(&models.WorkSubItem{WorkItemID: wi.ID, Name: n, DefaultQuantity: 1}); err != nil {
			t.Fatalf("create sub %s gagal: %v", n, err)
		}
	}

	detail, err := wiSvc.GetDetail(wi.ID)
	if err != nil {
		t.Fatalf("detail gagal: %v", err)
	}
	if len(detail.SubItems) != len(names) {
		t.Fatalf("jumlah sub salah: %d", len(detail.SubItems))
	}
	for i, s := range detail.SubItems {
		if s.Name != names[i] || s.Sequence != i+1 {
			t.Errorf("urutan awal salah index %d: %s (seq=%d)", i, s.Name, s.Sequence)
		}
	}

	// Ubah urutan: Commissioning pindah ke depan.
	ids := []uint{detail.SubItems[5].ID, detail.SubItems[0].ID, detail.SubItems[1].ID,
		detail.SubItems[2].ID, detail.SubItems[3].ID, detail.SubItems[4].ID}
	if err := subSvc.ReorderInWorkItem(wi.ID, ids); err != nil {
		t.Fatalf("reorder gagal: %v", err)
	}
	detail2, _ := wiSvc.GetDetail(wi.ID)
	wantOrder := []string{"Commissioning", "Inspection", "Troubleshooting", "Wiring Check", "Component Replacement", "Testing"}
	for i, s := range detail2.SubItems {
		if s.Name != wantOrder[i] || s.Sequence != i+1 {
			t.Errorf("urutan baru salah index %d: %s (seq=%d), mau %s", i, s.Name, s.Sequence, wantOrder[i])
		}
	}

	// Nonaktifkan satu sub.
	target := detail2.SubItems[3]
	if err := subSvc.SetActive(target.ID, false); err != nil {
		t.Fatalf("nonaktifkan sub gagal: %v", err)
	}
	detail3, _ := wiSvc.GetDetail(wi.ID)
	activeCount := 0
	for _, s := range detail3.SubItems {
		if s.ID == target.ID && s.IsActive {
			t.Error("sub target masih aktif")
		}
		if s.IsActive {
			activeCount++
		}
	}
	if activeCount != 5 {
		t.Errorf("jumlah sub aktif salah: %d", activeCount)
	}

	// Validasi bobot kesulitan out-of-range.
	if _, err := subSvc.Update(target.ID, &models.WorkSubItem{
		WorkItemID: wi.ID, Name: target.Name, DifficultyWeight: 150,
	}); err == nil {
		t.Fatal("bobot >100 seharusnya ditolak")
	}

	// Sub item tidak bisa pindah induk lewat update.
	otherWi, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Code: "EL-002", Name: "Lain", DefaultQuantity: 1})
	moved, err := subSvc.Update(target.ID, &models.WorkSubItem{
		WorkItemID: otherWi.ID, Name: target.Name, DefaultQuantity: 1,
	})
	if err != nil {
		t.Fatalf("update sub gagal: %v", err)
	}
	found := false
	for _, s := range moved.SubItems {
		if s.ID == target.ID {
			found = true
			if s.WorkItemID != wi.ID {
				t.Errorf("sub berpindah induk secara ilegal: workItemID=%d", s.WorkItemID)
			}
		}
	}
	if !found {
		t.Fatal("sub hilang dari induk asal")
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatal("tak terduga")
	}

	// Hapus pekerjaan induk kini kaskade: seluruh sub-pekerjaannya ikut terhapus.
	if err := wiSvc.Delete(wi.ID); err != nil {
		t.Fatalf("hapus pekerjaan dengan sub harus kaskade, gagal: %v", err)
	}
	var nItems, nSubs int64
	db.Model(&models.WorkItem{}).Where("id = ?", wi.ID).Count(&nItems)
	db.Model(&models.WorkSubItem{}).Where("work_item_id = ?", wi.ID).Count(&nSubs)
	if nItems != 0 || nSubs != 0 {
		t.Errorf("kaskade tidak lengkap: items=%d subs=%d", nItems, nSubs)
	}
}
