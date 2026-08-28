package masterdata

import (
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/collaboration"
	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return db
}

func newService(db *gorm.DB) *Service {
	return NewService(db, slog.New(slog.DiscardHandler))
}

func seedBundle(t *testing.T, db *gorm.DB) {
	t.Helper()
	cat := models.Category{Code: "CAT-A", Name: "Kategori A", Sequence: 1, IsActive: true}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	wi := models.WorkItem{CategoryID: cat.ID, Code: "WI-1", Name: "Pekerjaan 1", Sequence: 1, IsActive: true}
	if err := db.Create(&wi).Error; err != nil {
		t.Fatalf("seed work item: %v", err)
	}
	sub := models.WorkSubItem{WorkItemID: wi.ID, Code: "SUB-1", Name: "Sub 1", Sequence: 1, IsActive: true}
	if err := db.Create(&sub).Error; err != nil {
		t.Fatalf("seed sub item: %v", err)
	}
	if err := db.Create(&models.Material{Code: "MAT-1", Name: "Material 1", IsActive: true}).Error; err != nil {
		t.Fatalf("seed material: %v", err)
	}
}

// ----- Build -----

func TestBuildPackage(t *testing.T) {
	db := newTestDB(t)
	seedBundle(t, db)
	s := newService(db)

	pkg, err := s.BuildPackage("P1", "PC-A", "room-1")
	if err != nil {
		t.Fatalf("build package: %v", err)
	}
	if pkg.Metadata.PackageID == "" {
		t.Fatal("package id kosong")
	}
	if pkg.Metadata.SenderID != "P1" || pkg.Metadata.RoomID != "room-1" {
		t.Fatalf("metadata salah: %+v", pkg.Metadata)
	}
	if len(pkg.Data.Categories) != 1 || len(pkg.Data.WorkItems) != 1 ||
		len(pkg.Data.WorkSubItems) != 1 || len(pkg.Data.Materials) != 1 {
		t.Fatalf("jumlah item salah: cat=%d w=%d s=%d m=%d",
			len(pkg.Data.Categories), len(pkg.Data.WorkItems), len(pkg.Data.WorkSubItems), len(pkg.Data.Materials))
	}
	if pkg.Data.WorkItems[0].CategoryCode != "CAT-A" {
		t.Fatalf("categoryCode salah: %q", pkg.Data.WorkItems[0].CategoryCode)
	}
	if pkg.Data.WorkSubItems[0].WorkItemCode != "WI-1" {
		t.Fatalf("workItemCode salah: %q", pkg.Data.WorkSubItems[0].WorkItemCode)
	}
	ok, err := pkg.VerifyChecksum()
	if err != nil || !ok {
		t.Fatalf("checksum tidak valid: ok=%v err=%v", ok, err)
	}
}

// ----- Install: fresh + updates -----

func TestInstallFreshAndUpdate(t *testing.T) {
	db := newTestDB(t)
	s := newService(db)

	pkg := &collaboration.MasterDataPackage{
		Metadata: collaboration.MasterPackageMetadata{
			PackageID: "md-1", SenderID: "P1", SenderName: "PC-A",
			RoomID: "r1", CreatedAt: time.Now().UTC(), SchemaVersion: collaboration.PackageSchemaVersion,
		},
		Data: collaboration.MasterPackageData{
			Categories: []collaboration.PackageCategory{
				{Code: "CAT-A", Name: "Kategori A", Sequence: 1, IsActive: true},
			},
			WorkItems: []collaboration.PackageWorkItem{
				{Code: "WI-1", Name: "Pekerjaan 1", Sequence: 1, IsActive: true, CategoryCode: "CAT-A"},
			},
			WorkSubItems: []collaboration.PackageWorkSubItem{
				{Code: "SUB-1", Name: "Sub 1", Sequence: 1, IsActive: true, WorkItemCode: "WI-1"},
			},
			Materials: []collaboration.PackageMaterial{
				{Code: "MAT-1", Name: "Material 1", IsActive: true},
			},
		},
	}
	sum, err := pkg.ComputeChecksum()
	if err != nil {
		t.Fatal(err)
	}
	pkg.Checksum = sum

	res, err := s.Install(pkg, StrategyPrompt, nil)
	if err != nil {
		t.Fatalf("install fresh: %v", err)
	}
	if res.CategoriesCreated != 1 || res.WorkItemsCreated != 1 ||
		res.SubItemsCreated != 1 || res.MaterialsCreated != 1 {
		t.Fatalf("create counts salah: %+v", res)
	}

	// Install ulang identik → tidak ada perubahan (UNCHANGED), hanya konflik 0.
	res2, err := s.Install(pkg, StrategyPrompt, nil)
	if err != nil {
		t.Fatalf("install identik: %v", err)
	}
	if res2.CategoriesUpdated != 0 || res2.CategoriesCreated != 0 {
		t.Fatalf("install identik harus noop: %+v", res2)
	}

	// Ubah satu field → USE_INCOMING harus update.
	pkg.Data.Categories[0].Name = "Kategori A v2"
	sum2, _ := pkg.ComputeChecksum()
	pkg.Checksum = sum2
	res3, err := s.Install(pkg, StrategyUseIncoming, nil)
	if err != nil {
		t.Fatalf("install update: %v", err)
	}
	if res3.CategoriesUpdated != 1 {
		t.Fatalf("update count salah: %+v", res3)
	}
	var cat models.Category
	if err := db.Where("code = ?", "CAT-A").First(&cat).Error; err != nil {
		t.Fatal(err)
	}
	if cat.Name != "Kategori A v2" {
		t.Fatalf("nama tidak terupdate: %q", cat.Name)
	}
}

func TestInstallDefaultSkipConflict(t *testing.T) {
	db := newTestDB(t)
	seedBundle(t, db)
	s := newService(db)

	// Konflik: kategori CAT-A sudah ada dengan nilai berbeda.
	pkg := &collaboration.MasterDataPackage{
		Data: collaboration.MasterPackageData{
			Categories: []collaboration.PackageCategory{
				{Code: "CAT-A", Name: "Nama Berbeda", Sequence: 99, IsActive: false},
			},
			Materials: []collaboration.PackageMaterial{
				{Code: "MAT-1", Name: "Material 1", IsActive: true}, // identik → UNCHANGED
			},
		},
	}
	sum, _ := pkg.ComputeChecksum()
	pkg.Checksum = sum

	res, err := s.Install(pkg, StrategyPrompt, nil)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if res.Conflicts != 1 || res.Skipped != 1 || res.CategoriesUpdated != 0 {
		t.Fatalf("default harus skip konflik: %+v", res)
	}
	var cat models.Category
	db.Where("code = ?", "CAT-A").First(&cat)
	if cat.Name == "Nama Berbeda" {
		t.Fatal("konflik seharusnya tidak mengubah data lokal")
	}
}

func TestInstallUseIncomingOverwrites(t *testing.T) {
	db := newTestDB(t)
	seedBundle(t, db)
	s := newService(db)

	pkg := &collaboration.MasterDataPackage{
		Data: collaboration.MasterPackageData{
			Categories: []collaboration.PackageCategory{
				{Code: "CAT-A", Name: "Nama Berbeda", Sequence: 99, IsActive: false},
			},
		},
	}
	sum, _ := pkg.ComputeChecksum()
	pkg.Checksum = sum

	res, err := s.Install(pkg, StrategyUseIncoming, nil)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if res.CategoriesUpdated != 1 || res.Conflicts != 0 {
		t.Fatalf("USE_INCOMING harus menimpa: %+v", res)
	}
	var cat models.Category
	db.Where("code = ?", "CAT-A").First(&cat)
	if cat.Name != "Nama Berbeda" {
		t.Fatalf("data lokal harus ditimpa: %q", cat.Name)
	}
}

func TestInstallCorruptPackageRejected(t *testing.T) {
	db := newTestDB(t)
	s := newService(db)

	pkg := &collaboration.MasterDataPackage{
		Data: collaboration.MasterPackageData{
			Categories: []collaboration.PackageCategory{{Code: "CAT-A", Name: "A"}},
		},
	}
	sum, _ := pkg.ComputeChecksum()
	pkg.Checksum = sum

	pkg.Data.Categories[0].Name = "A-rusak" // checksum sekarang tidak cocok
	_, err := s.Install(pkg, StrategyPrompt, nil)
	if err == nil {
		t.Fatal("package rusak harus ditolak")
	}
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count != 0 {
		t.Fatalf("package rusak harus ditolak tanpa perubahan: count=%d", count)
	}
}

// Install menggabungkan seluruh entity dalam satu transaksi (Fase 10). Karena
// keseluruhan data diverifikasi checksum-nya SEBELUM pemasangan, dan pemasangan
// bersifat idempoten per natural key, tidak ada langkah yang bisa meninggalkan
// data parsial saat package valid.

// ----- Compare -----

func TestCompareKinds(t *testing.T) {
	db := newTestDB(t)
	seedBundle(t, db)
	s := newService(db)

	pkg := &collaboration.MasterDataPackage{
		Data: collaboration.MasterPackageData{
			Categories: []collaboration.PackageCategory{
				{Code: "CAT-A", Name: "Berbeda", Sequence: 9, IsActive: false}, // beda vs lokal → CONFLICT
				{Code: "CAT-B", Name: "Baru", Sequence: 2, IsActive: true},     // baru → NEW
			},
			Materials: []collaboration.PackageMaterial{
				{Code: "MAT-1", Name: "Material 1", IsActive: true}, // identik → UNCHANGED
				{Code: "MAT-2", Name: "Baru", IsActive: true},       // → NEW
			},
		},
	}
	diffs, err := s.Compare(pkg)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	kinds := map[string]int{}
	byCode := map[string]string{}
	for _, d := range diffs {
		kinds[d.Kind]++
		byCode[d.Code] = d.Kind
	}
	if kinds[DiffNew] != 2 || kinds[DiffUnchanged] != 1 || kinds[DiffConflict] != 1 {
		t.Fatalf("kinds salah: %+v", kinds)
	}
	if byCode["CAT-B"] != DiffNew || byCode["CAT-A"] != DiffConflict || byCode["MAT-2"] != DiffNew {
		t.Fatalf("diff per code salah: %+v", byCode)
	}
}

// ----- Inbox -----

func TestInboxDedup(t *testing.T) {
	db := newTestDB(t)
	s := newService(db)

	in := &models.MasterInbox{
		RoomID: "r1", PackageID: "md-dup", SenderID: "P1", SenderName: "PC-A",
		Payload: `{}`, Checksum: "x", Status: models.MasterStatusPending, ReceivedAt: time.Now().UTC(),
	}
	if err := s.SaveInbox(in); err != nil {
		t.Fatalf("save inbox: %v", err)
	}
	if err := s.SaveInbox(in); err != nil {
		t.Fatalf("save inbox duplikat harus diabaikan: %v", err)
	}
	list, err := s.InboxList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("inbox harus dedup: count=%d", len(list))
	}
}

func TestInboxStatusTransition(t *testing.T) {
	db := newTestDB(t)
	s := newService(db)

	in := &models.MasterInbox{
		RoomID: "r1", PackageID: "md-1", SenderID: "P1", SenderName: "PC-A",
		Payload: `{}`, Checksum: "x", Status: models.MasterStatusPending, ReceivedAt: time.Now().UTC(),
	}
	if err := s.SaveInbox(in); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := s.SetInboxStatus("md-1", models.MasterStatusInstalled, now); err != nil {
		t.Fatal(err)
	}
	item, err := s.InboxGet("md-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != models.MasterStatusInstalled || item.InstalledAt == nil {
		t.Fatalf("status transisi salah: %+v", item)
	}
}
