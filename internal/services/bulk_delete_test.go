package services

import (
	"log/slog"
	"testing"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func TestDeleteWorkItemCascadesSubItems(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	subSvc := NewWorkSubItemService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	wi, err := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Repair Control Panel", DefaultQuantity: 1})
	if err != nil {
		t.Fatalf("create pekerjaan gagal: %v", err)
	}
	s1, err := subSvc.Create(&models.WorkSubItem{WorkItemID: wi.ID, Name: "Inspeksi awal", DefaultQuantity: 1})
	if err != nil {
		t.Fatalf("create sub 1 gagal: %v", err)
	}
	if _, err := subSvc.Create(&models.WorkSubItem{WorkItemID: wi.ID, Name: "Penggantian part", DefaultQuantity: 1}); err != nil {
		t.Fatalf("create sub 2 gagal: %v", err)
	}
	// sub nonaktif juga harus ikut terhapus kaskade.
	var s2 models.WorkSubItem
	if err := db.First(&s2, "id <> ?", s1.ID).Error; err != nil {
		t.Fatalf("ambil sub 2 gagal: %v", err)
	}
	if err := subSvc.SetActive(s2.ID, false); err != nil {
		t.Fatalf("set inactive sub gagal: %v", err)
	}

	if err := wiSvc.Delete(wi.ID); err != nil {
		t.Fatalf("hapus pekerjaan dengan sub seharusnya kaskade, gagal: %v", err)
	}

	var nItems, nSubs int64
	db.Model(&models.WorkItem{}).Where("id = ?", wi.ID).Count(&nItems)
	db.Model(&models.WorkSubItem{}).Where("work_item_id = ?", wi.ID).Count(&nSubs)
	if nItems != 0 || nSubs != 0 {
		t.Errorf("kaskade tidak lengkap: items=%d subs=%d", nItems, nSubs)
	}
}

func TestDeleteWorkItemsBatch(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	subSvc := NewWorkSubItemService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	a, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Alpha", DefaultQuantity: 1})
	b, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Beta", DefaultQuantity: 1})
	c, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Gamma", DefaultQuantity: 1})
	for _, n := range []string{"B-1", "B-2"} {
		if _, err := subSvc.Create(&models.WorkSubItem{WorkItemID: b.ID, Name: n, DefaultQuantity: 1}); err != nil {
			t.Fatalf("create sub gagal: %v", err)
		}
	}

	res, err := wiSvc.DeleteMany([]uint{a.ID, b.ID, c.ID, a.ID}) // duplikat diabaikan
	if err != nil {
		t.Fatalf("hapus massal gagal: %v", err)
	}
	if res.Items != 3 || res.Subs != 2 {
		t.Errorf("ringkasan salah: %+v (harus 3 pekerjaan, 2 sub)", res)
	}
	var nItems, nSubs int64
	db.Model(&models.WorkItem{}).Where("category_id = ?", cat.ID).Count(&nItems)
	db.Model(&models.WorkSubItem{}).Count(&nSubs)
	if nItems != 0 || nSubs != 0 {
		t.Errorf("masih ada data hidup: items=%d subs=%d", nItems, nSubs)
	}
}

func TestDeleteWorkItemsBatchRejectedOnMissing(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	a, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Alpha", DefaultQuantity: 1})

	res, err := wiSvc.DeleteMany([]uint{a.ID, 99999})
	if err == nil {
		t.Fatal("batch berisi ID hilang seharusnya ditolak")
	}
	if res != nil {
		t.Errorf("hasil harus nil saat batch ditolak, dapat: %+v", res)
	}
	if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}
	var n int64
	db.Model(&models.WorkItem{}).Where("id = ?", a.ID).Count(&n)
	if n != 1 {
		t.Error("batch ditolak tapi pekerjaan lain ikut terhapus (transaksi bocor)")
	}
}

func TestDeleteWorkItemsEmptySelection(t *testing.T) {
	db := serviceDB(t)
	wiSvc := NewWorkItemService(db, slog.Default())
	subSvc := NewWorkSubItemService(db, slog.Default())

	if _, err := wiSvc.DeleteMany(nil); err == nil {
		t.Fatal("hapus massal tanpa pilihan seharusnya ditolak")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("harus ValidationError, dapat: %v", err)
	}
	if _, err := subSvc.DeleteMany([]uint{}); err == nil {
		t.Fatal("hapus massal sub tanpa pilihan seharusnya ditolak")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("harus ValidationError, dapat: %v", err)
	}
}

func TestDeleteWorkItemsBatchBlockedByTemplateRollback(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	tplSvc := NewTemplateService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	a, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Dipakai Template", DefaultQuantity: 1})
	b, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Bebas", DefaultQuantity: 1})

	tpl, _ := tplSvc.Create(&models.Template{Name: "Repair PLC"})
	if _, err := tplSvc.SetItems(tpl.ID, []TemplateItemInput{{WorkItemID: a.ID}}); err != nil {
		t.Fatalf("set items gagal: %v", err)
	}

	res, err := wiSvc.DeleteMany([]uint{a.ID, b.ID})
	if err == nil {
		t.Fatal("batch yang menyentuh pekerjaan terpakai template seharusnya ditolak")
	}
	if res != nil {
		t.Errorf("hasil harus nil saat batch ditolak, dapat: %+v", res)
	}
	if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}
	var nA, nB int64
	db.Model(&models.WorkItem{}).Where("id = ?", a.ID).Count(&nA)
	db.Model(&models.WorkItem{}).Where("id = ?", b.ID).Count(&nB)
	if nA != 1 || nB != 1 {
		t.Errorf("rollback tidak penuh: dipakai=%d bebas=%d", nA, nB)
	}
}

func TestDeleteSubItemsBatch(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	subSvc := NewWorkSubItemService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	wi, _ := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Alpha", DefaultQuantity: 1})
	for _, n := range []string{"S-1", "S-2", "S-3"} {
		if _, err := subSvc.Create(&models.WorkSubItem{WorkItemID: wi.ID, Name: n, DefaultQuantity: 1}); err != nil {
			t.Fatalf("create sub gagal: %v", err)
		}
	}
	var ids []uint
	db.Model(&models.WorkSubItem{}).Where("work_item_id = ?", wi.ID).Order("sequence asc").Pluck("id", &ids)

	res, err := subSvc.DeleteMany(ids[:2])
	if err != nil {
		t.Fatalf("hapus massal sub gagal: %v", err)
	}
	if res.Subs != 2 {
		t.Errorf("jumlah terhapus salah: %d", res.Subs)
	}
	var n int64
	db.Model(&models.WorkSubItem{}).Where("work_item_id = ?", wi.ID).Count(&n)
	if n != 1 {
		t.Errorf("sub tersisa salah: %d", n)
	}

	// ID basi (sudah terhapus) menolak seluruh batch.
	if _, err := subSvc.DeleteMany([]uint{ids[2], ids[0]}); err == nil {
		t.Fatal("batch dengan sub yang sudah terhapus seharusnya ditolak")
	} else if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}
}
