package database

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	lg := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := Open(path, lg)
	if err != nil {
		t.Fatalf("Open gagal: %v", err)
	}
	t.Cleanup(func() {
		Close(db)
	})
	return db
}

func TestMigrateCreatesAllTables(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	tables, err := ExistingTables(db)
	if err != nil {
		t.Fatalf("ExistingTables gagal: %v", err)
	}
	want := []string{
		"audit_logs",
		"categories",
		"customers",
		"materials",
		"sph_documents",
		"sph_items",
		"sph_revisions",
		"sph_sub_items",
		"settings",
		"template_items",
		"templates",
		"vessels",
		"work_items",
		"work_sub_items",
	}
	got := map[string]bool{}
	for _, tb := range tables {
		got[tb] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("tabel tidak ditemukan: %s (ada: %v)", w, tables)
		}
	}
}

func TestMigrateIdempotent(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate pertama gagal: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate kedua (idempoten) gagal: %v", err)
	}
}

func TestForeignKeyEnforced(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	item := models.WorkItem{CategoryID: 9999, Code: "X-01", Name: "Tanpa Kategori"}
	if err := db.Create(&item).Error; err == nil {
		t.Fatal("FK tidak berlaku: work item dengan category_id hantu berhasil dibuat")
	}
	cat := models.Category{Code: "EL", Name: "Electrical"}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("buat kategori gagal: %v", err)
	}
	item2 := models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Repair Control Panel"}
	if err := db.Create(&item2).Error; err != nil {
		t.Fatalf("buat work item valid gagal: %v", err)
	}
	violations, err := ForeignKeyViolations(db)
	if err != nil {
		t.Fatalf("foreign_key_check gagal: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("ada pelanggaran FK: %v", violations)
	}
}

func TestCategoryDeleteRestrictedWhenUsed(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	cat := models.Category{Code: "EL", Name: "Electrical"}
	db.Create(&cat)
	item := models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Repair"}
	db.Create(&item)

	if err := db.Delete(&cat).Error; err != nil {
		t.Fatalf("soft delete kategori seharusnya selalu boleh: %v", err)
	}
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count != 0 {
		t.Error("soft delete tidak bekerja")
	}

	if err := db.Unscoped().Delete(&models.Category{}, cat.ID).Error; err == nil {
		t.Fatal("hapus permanen kategori yang masih dipakai work item seharusnya ditolak FK")
	}
}

func TestPartialUniqueIndexAllowsReuseAfterSoftDelete(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	a := models.Category{Code: "EL", Name: "Electrical"}
	if err := db.Create(&a).Error; err != nil {
		t.Fatalf("buat pertama gagal: %v", err)
	}
	dup := models.Category{Code: "EL", Name: "Duplikat"}
	if err := db.Create(&dup).Error; err == nil {
		t.Fatal("kode duplikat aktif seharusnya ditolak unique index")
	}
	db.Delete(&a)
	reuse := models.Category{Code: "EL", Name: "Electrical Baru"}
	if err := db.Create(&reuse).Error; err != nil {
		t.Fatalf("kode yang di-soft-delete seharusnya bisa dipakai lagi: %v", err)
	}
}

func TestSphDocumentUniqueNumberRevision(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	cust := models.Customer{Name: "Pelanggan A"}
	db.Create(&cust)
	doc := models.SphDocument{
		DocumentNumber: "SPH/GEI/VIII/2026/001",
		Revision:       0,
		Date:           time.Now(),
		CustomerID:     cust.ID,
		Status:         models.StatusDraft,
	}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("dokumen pertama gagal: %v", err)
	}
	sameRev := doc
	sameRev.ID = 0
	if err := db.Create(&sameRev).Error; err == nil {
		t.Fatal("nomor+revisi duplikat aktif seharusnya ditolak")
	}
	nextRev := doc
	nextRev.ID = 0
	nextRev.Revision = 1
	if err := db.Create(&nextRev).Error; err != nil {
		t.Fatalf("nomor sama revisi beda seharusnya boleh: %v", err)
	}
}

func TestSettingsUniqueKey(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	s1 := models.Setting{Key: "company_profile", Value: "{}"}
	if err := db.Create(&s1).Error; err != nil {
		t.Fatalf("setting pertama gagal: %v", err)
	}
	s2 := models.Setting{Key: "company_profile", Value: "{\"x\":1}"}
	if err := db.Create(&s2).Error; err == nil {
		t.Fatal("key setting duplikat seharusnya ditolak")
	}
}

func TestCascadeDocumentToItems(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	cust := models.Customer{Name: "Pelanggan B"}
	db.Create(&cust)
	doc := models.SphDocument{
		DocumentNumber: "SPH/GEI/VIII/2026/002",
		Date:           time.Now(),
		CustomerID:     cust.ID,
		Status:         models.StatusDraft,
	}
	db.Create(&doc)
	it := models.SphItem{
		SphDocumentID:    doc.ID,
		NameSnapshot:     "Repair PLC",
		Quantity:         1,
		PricingMode:      models.PricingModeDirect,
		ServiceUnitPrice: 10000000,
		ServiceTotal:     10000000,
		Total:            10000000,
	}
	db.Create(&it)
	sub := models.SphSubItem{
		SphItemID:    it.ID,
		NameSnapshot: "Inspection",
		Sequence:     1,
		Weight:       100,
	}
	db.Create(&sub)

	db.Unscoped().Delete(&doc)

	var itemCount, subCount int64
	db.Model(&models.SphItem{}).Where("sph_document_id = ?", doc.ID).Count(&itemCount)
	db.Model(&models.SphSubItem{}).Where("sph_item_id = ?", it.ID).Count(&subCount)
	if itemCount != 0 || subCount != 0 {
		t.Errorf("cascade gagal: item tersisa=%d, sub tersisa=%d", itemCount, subCount)
	}
}

func TestWorkSubItemOrderingAndDefaults(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	cat := models.Category{Code: "EL", Name: "Electrical"}
	db.Create(&cat)
	wi := models.WorkItem{CategoryID: cat.ID, Code: "EL-001", Name: "Repair PLC", DefaultUnit: "giat"}
	db.Create(&wi)
	names := []string{"Inspection", "Backup Program", "Mapping I/O", "Troubleshooting", "Repair", "Reprogramming", "Testing", "Commissioning"}
	for i, n := range names {
		sub := models.WorkSubItem{WorkItemID: wi.ID, Sequence: i + 1, Name: n, DifficultyWeight: 0}
		if err := db.Create(&sub).Error; err != nil {
			t.Fatalf("buat sub item %s gagal: %v", n, err)
		}
	}
	var subs []models.WorkSubItem
	db.Where("work_item_id = ?", wi.ID).Order("sequence asc").Find(&subs)
	if len(subs) != len(names) {
		t.Fatalf("jumlah sub item salah: %d", len(subs))
	}
	for i, s := range subs {
		if s.Name != names[i] {
			t.Errorf("urutan salah index %d: dapat %s, mau %s", i, s.Name, names[i])
		}
	}
}
