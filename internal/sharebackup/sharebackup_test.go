package sharebackup

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

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

// seedFull mencatat satu set data lengkap (termasuk SPH final) ke database.
func seedFull(t *testing.T, db *gorm.DB) {
	t.Helper()
	cat := models.Category{Code: "KAT-001", Name: "Kategori A", Sequence: 1, IsActive: true}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	wi := models.WorkItem{CategoryID: cat.ID, Code: "PEK-001", Name: "Pekerjaan 1", Sequence: 1, IsActive: true}
	if err := db.Create(&wi).Error; err != nil {
		t.Fatalf("seed work item: %v", err)
	}
	sub := models.WorkSubItem{WorkItemID: wi.ID, Code: "SUB-001", Name: "Sub 1", Sequence: 1, IsActive: true}
	if err := db.Create(&sub).Error; err != nil {
		t.Fatalf("seed sub item: %v", err)
	}
	tpl := models.Template{Code: "TPL-001", Name: "Template 1", Sequence: 1, IsActive: true}
	if err := db.Create(&tpl).Error; err != nil {
		t.Fatalf("seed template: %v", err)
	}
	if err := db.Create(&models.TemplateItem{TemplateID: tpl.ID, Sequence: 1, WorkItemID: wi.ID}).Error; err != nil {
		t.Fatalf("seed template item: %v", err)
	}
	cus := models.Customer{Code: "CUS-X", Name: "PT Maju", IsActive: true}
	if err := db.Create(&cus).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	ves := models.Vessel{CustomerID: cus.ID, Code: "KAP-1", Name: "Kapal 1", IsActive: true}
	if err := db.Create(&ves).Error; err != nil {
		t.Fatalf("seed vessel: %v", err)
	}
	if err := db.Create(&models.Material{Code: "MAT-001", Name: "Material A", IsActive: true}).Error; err != nil {
		t.Fatalf("seed material: %v", err)
	}
	if err := db.Create(&models.Material{Code: "MAT-002", Name: "Material B", IsActive: true}).Error; err != nil {
		t.Fatalf("seed material B: %v", err)
	}
	finalAt := time.Date(2026, 9, 1, 10, 0, 0, 0, time.Local)
	doc := models.SphDocument{
		DocumentNumber: "SPH-2026-001",
		Revision:       1,
		Date:           time.Date(2026, 9, 1, 8, 0, 0, 0, time.Local),
		CustomerID:     cus.ID,
		VesselID:       &ves.ID,
		Subject:        "Penawaran",
		Status:         "FINALIZED",
		GrandTotal:     1000000,
		FinalizedAt:    &finalAt,
	}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("seed sph doc: %v", err)
	}
	item := models.SphItem{
		SphDocumentID: doc.ID,
		Sequence:      1,
		WorkItemID:    &wi.ID,
		NameSnapshot:  "Pekerjaan 1",
		Quantity:      1,
		PricingMode:   "HARGA_LANGSUNG",
		Total:         1000000,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed sph item: %v", err)
	}
	if err := db.Create(&models.SphSubItem{SphItemID: item.ID, Sequence: 1, NameSnapshot: "Sub 1", Quantity: 1, Total: 1000000}).Error; err != nil {
		t.Fatalf("seed sph sub: %v", err)
	}
	created := time.Date(2026, 8, 25, 9, 0, 0, 0, time.Local)
	if err := db.Create(&models.SphRevision{SphDocumentID: doc.ID, RevisionNumber: 1, Note: "Revisi pertama", CreatedAt: created}).Error; err != nil {
		t.Fatalf("seed sph revision: %v", err)
	}
}

func countTable(t *testing.T, db *gorm.DB, table string) int64 {
	t.Helper()
	var n int64
	if err := db.Table(table).Count(&n).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func TestBuildFillsNaturalKeys(t *testing.T) {
	db := newTestDB(t)
	seedFull(t, db)
	s := newService(db)

	pkg, err := s.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(pkg.Categories) != 1 || len(pkg.WorkItems) != 1 || len(pkg.WorkSubItems) != 1 {
		t.Fatalf("item counts salah: cat=%d wi=%d sub=%d", len(pkg.Categories), len(pkg.WorkItems), len(pkg.WorkSubItems))
	}
	if len(pkg.Templates) != 1 || len(pkg.Customers) != 1 || len(pkg.Materials) != 2 || len(pkg.SphDocuments) != 1 {
		t.Fatalf("counts salah: tpl=%d cus=%d mat=%d sph=%d", len(pkg.Templates), len(pkg.Customers), len(pkg.Materials), len(pkg.SphDocuments))
	}
	if pkg.WorkItems[0].CategoryName != "Kategori A" {
		t.Fatalf("categoryName salah: %q", pkg.WorkItems[0].CategoryName)
	}
	if pkg.WorkSubItems[0].WorkItemName != "Pekerjaan 1" || pkg.WorkSubItems[0].CategoryName != "Kategori A" {
		t.Fatalf("sub natural key salah: %+v", pkg.WorkSubItems[0])
	}
	if pkg.Templates[0].Items[0].WorkItemName != "Pekerjaan 1" {
		t.Fatalf("template item natural key salah: %+v", pkg.Templates[0].Items[0])
	}
	if pkg.Customers[0].Vessels[0].Name != "Kapal 1" {
		t.Fatalf("vessel salah: %+v", pkg.Customers[0].Vessels[0])
	}
	sph := pkg.SphDocuments[0]
	if sph.DocumentNumber != "SPH-2026-001" || sph.Status != "FINALIZED" || sph.CustomerName != "PT Maju" {
		t.Fatalf("sph salah: %+v", sph)
	}
	if sph.Items[0].WorkItemName != "Pekerjaan 1" || len(sph.Revisions) != 1 {
		t.Fatalf("sph items/revisions salah: %+v", sph)
	}
	ok, err := pkg.VerifyChecksum()
	if err != nil || !ok {
		t.Fatalf("checksum tidak valid: ok=%v err=%v", ok, err)
	}
}

func TestChecksumDetectsTampering(t *testing.T) {
	db := newTestDB(t)
	seedFull(t, db)
	pkg, err := newService(db).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	pkg.Checksum = ""
	pkg.WorkItems[0].Name = "Diubah"
	if ok, _ := pkg.VerifyChecksum(); ok {
		t.Fatal("checksum harus gagal setelah data diubah")
	}
}

func TestFileRoundtripAndCorrupt(t *testing.T) {
	db := newTestDB(t)
	seedFull(t, db)
	pkg, err := newService(db).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	path := filepath.Join(t.TempDir(), "bundle.sphbak")
	if err := newService(db).WriteFile(pkg, path); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got.Materials) != 2 || got.Metadata.DeviceName != pkg.Metadata.DeviceName {
		t.Fatalf("roundtrip tidak sama: %+v", got.Metadata)
	}

	raw, _ := os.ReadFile(path)
	raw = append(raw[:len(raw)-20], []byte(strings.Repeat("X", 20))...)
	corrupt := filepath.Join(t.TempDir(), "corrupt.sphbak")
	if err := os.WriteFile(corrupt, raw, 0644); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}
	if _, err := ReadFile(corrupt); err == nil {
		t.Fatal("file rusak harus ditolak")
	}
}

func TestRestoreAddOnlyAllSections(t *testing.T) {
	src := newTestDB(t)
	seedFull(t, src)
	pkg, err := newService(src).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	dst := newTestDB(t)
	// Data lokal yang bertabrakan nama: kategori, pekerjaan, sub-pekerjaan, dan
	// material yang sama, plus satu material ekstra khas lokal.
	catDst := models.Category{Code: "KAT-001", Name: "Kategori A", Sequence: 1, IsActive: true}
	if err := dst.Create(&catDst).Error; err != nil {
		t.Fatalf("seed dst: %v", err)
	}
	wiDst := models.WorkItem{CategoryID: catDst.ID, Code: "PEK-001", Name: "Pekerjaan 1", Sequence: 1, IsActive: true}
	if err := dst.Create(&wiDst).Error; err != nil {
		t.Fatalf("seed dst: %v", err)
	}
	if err := dst.Create(&models.WorkSubItem{WorkItemID: wiDst.ID, Code: "SUB-001", Name: "Sub 1", Sequence: 1, IsActive: true}).Error; err != nil {
		t.Fatalf("seed dst: %v", err)
	}
	if err := dst.Create(&models.Material{Code: "MAT-001", Name: "Material A", IsActive: true}).Error; err != nil {
		t.Fatalf("seed dst: %v", err)
	}
	if err := dst.Create(&models.Material{Code: "MAT-Z", Name: "Material Z", IsActive: true}).Error; err != nil {
		t.Fatalf("seed dst: %v", err)
	}

	sum, err := newService(dst).Restore(pkg, InstallSelectedSections(allKeys()))
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	if sum.Categories.Skipped != 1 || sum.WorkItems.Skipped != 1 || sum.SubItems.Skipped != 1 {
		t.Fatalf("skip pekerjaan salah: %+v", sum)
	}
	if sum.Templates.Added != 1 || sum.TemplateItemsAdded != 1 {
		t.Fatalf("template tidak diimpor: %+v", sum)
	}
	if sum.Customers.Added != 1 || sum.Vessels.Added != 1 {
		t.Fatalf("customer/vessel tidak diimpor: %+v", sum)
	}
	if sum.Materials.Skipped != 1 || sum.Materials.Added != 1 {
		t.Fatalf("material salah: %+v", sum)
	}
	if sum.Sph.Added != 1 {
		t.Fatalf("sph tidak diimpor: %+v", sum)
	}

	if got := countTable(t, dst, "categories"); got != 1 {
		t.Fatalf("kategori duplikat: %d", got)
	}
	if got := countTable(t, dst, "work_items"); got != 1 {
		t.Fatalf("pekerjaan duplikat: %d", got)
	}
	if got := countTable(t, dst, "work_sub_items"); got != 1 {
		t.Fatalf("sub duplikat: %d", got)
	}
	if got := countTable(t, dst, "templates"); got != 1 {
		t.Fatalf("template duplikat: %d", got)
	}
	if got := countTable(t, dst, "customers"); got != 1 {
		t.Fatalf("customer duplikat: %d", got)
	}
	if got := countTable(t, dst, "vessels"); got != 1 {
		t.Fatalf("vessel duplikat: %d", got)
	}
	if got := countTable(t, dst, "materials"); got != 3 {
		t.Fatalf("material jumlah salah: %d", got)
	}
	if got := countTable(t, dst, "sph_documents"); got != 1 {
		t.Fatalf("sph duplikat: %d", got)
	}

	// SPH dipertahankan apa adanya (status + snapshot + revisi tanpa FromDocumentID).
	var doc models.SphDocument
	if err := dst.Where("document_number = ?", "SPH-2026-001").First(&doc).Error; err != nil {
		t.Fatalf("doc tidak ditemukan: %v", err)
	}
	if doc.Status != "FINALIZED" || doc.FinalizedAt == nil {
		t.Fatalf("status/finalized tidak dipertahankan: %+v", doc)
	}
	var item models.SphItem
	if err := dst.Where("sph_document_id = ?", doc.ID).First(&item).Error; err != nil {
		t.Fatalf("sph item tidak ada: %v", err)
	}
	if item.WorkItemID == nil {
		t.Fatal("sph item seharusnya tertaut ke pekerjaan yang cocok")
	}
	var rev models.SphRevision
	if err := dst.Where("sph_document_id = ?", doc.ID).First(&rev).Error; err != nil {
		t.Fatalf("revisi tidak ada: %v", err)
	}
	if rev.FromDocumentID != nil {
		t.Fatal("revisi impor harus tanpa FromDocumentID")
	}

	// Restore kedua: semuanya dilewati, tanpa duplikasi.
	sum2, err := newService(dst).Restore(pkg, InstallSelectedSections(allKeys()))
	if err != nil {
		t.Fatalf("restore kedua: %v", err)
	}
	if sum2.Sph.Added != 0 || sum2.Materials.Added != 0 || sum2.Templates.Added != 0 {
		t.Fatalf("restore kedua tidak idempoten: %+v", sum2)
	}
	if got := countTable(t, dst, "sph_documents"); got != 1 {
		t.Fatalf("duplikasi sph saat restore kedua: %d", got)
	}
}

func TestRestoreCodeCollisionGeneratesNewCode(t *testing.T) {
	src := newTestDB(t)
	seedFull(t, src)
	pkg, err := newService(src).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	dst := newTestDB(t)
	// Kode KAT-001 sudah dipakai kategori lokal dengan NAMA berbeda.
	if err := dst.Create(&models.Category{Code: "KAT-001", Name: "Kategori Lokal", Sequence: 5, IsActive: true}).Error; err != nil {
		t.Fatalf("seed dst: %v", err)
	}

	sum, err := newService(dst).Restore(pkg, InstallSelectedSections([]string{"categories"}))
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if sum.Categories.Added != 1 || sum.Categories.CodeGenerated != 1 {
		t.Fatalf("kode harus digenerate baru: %+v", sum.Categories)
	}
	var lok models.Category
	if err := dst.Where("name = ?", "Kategori Lokal").First(&lok).Error; err != nil {
		t.Fatalf("kategori lokal hilang: %v", err)
	}
	if lok.Code != "KAT-001" {
		t.Fatalf("kategori lokal harus tetap memegang kodenya: %s", lok.Code)
	}
	var baru models.Category
	if err := dst.Where("name = ?", "Kategori A").First(&baru).Error; err != nil {
		t.Fatalf("kategori impor tidak ada: %v", err)
	}
	if baru.Code == "KAT-001" {
		t.Fatalf("kode kategori impor menabrak milik lokal")
	}
}

func TestRestorePartialSelection(t *testing.T) {
	src := newTestDB(t)
	seedFull(t, src)
	pkg, err := newService(src).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	dst := newTestDB(t)
	sum, err := newService(dst).Restore(pkg, InstallSelectedSections([]string{"materials"}))
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if sum.Materials.Added != 2 {
		t.Fatalf("material harus diimpor: %+v", sum)
	}
	if sum.Sph.Added != 0 || sum.WorkItems.Added != 0 || sum.Categories.Added != 0 || sum.Templates.Added != 0 {
		t.Fatalf("seksi lain ikut terimpor: %+v", sum)
	}
	if got := countTable(t, dst, "sph_documents"); got != 0 {
		t.Fatalf("sph tidak boleh terimpor: %d", got)
	}
	if got := countTable(t, dst, "categories"); got != 0 {
		t.Fatalf("kategori tidak boleh terimpor: %d", got)
	}
}

func TestRestoreSphOnlyAutoCreatesParents(t *testing.T) {
	src := newTestDB(t)
	seedFull(t, src)
	pkg, err := newService(src).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	dst := newTestDB(t)
	sum, err := newService(dst).Restore(pkg, InstallSelectedSections([]string{"sph"}))
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if sum.Sph.Added != 1 {
		t.Fatalf("sph harus terimpor: %+v", sum)
	}
	if sum.Customers.Added != 1 || sum.Vessels.Added != 1 {
		t.Fatalf("customer/vessel induk harus dibuat otomatis: %+v", sum)
	}
	if sum.WorkItems.Added != 0 {
		t.Fatalf("master pekerjaan tidak boleh terimpor: %+v", sum)
	}

	var cus models.Customer
	if err := dst.Where("name = ?", "PT Maju").First(&cus).Error; err != nil {
		t.Fatalf("customer otomatis tidak ada: %v", err)
	}
	var doc models.SphDocument
	if err := dst.Where("document_number = ?", "SPH-2026-001").First(&doc).Error; err != nil {
		t.Fatalf("sph otomatis tidak ada: %v", err)
	}
	if doc.CustomerID != cus.ID {
		t.Fatalf("sph harus menunjuk customer lokal: %d != %d", doc.CustomerID, cus.ID)
	}
	if doc.CustomerID == 0 {
		t.Fatal("customer id tidak valid")
	}
}

func TestRestoreTemplateItemMissedWhenWorkItemAbsent(t *testing.T) {
	pkg := &ShareBackupPackage{
		SchemaVersion: PackageSchemaVersion,
		Metadata:      ShareBackupMetadata{PackageID: "pkg-test", DeviceName: "mesin", CreatedAt: time.Now()},
		Templates: []PackageTemplate{{
			Code: "TPL-X", Name: "Template Tanpa Pekerjaan", Sequence: 1, IsActive: true,
			Items: []PackageTemplateItem{{WorkItemName: "Pekerjaan Tidak Ada"}},
		}},
	}
	sum, err := newService(newTestDB(t)).Restore(pkg, InstallSelectedSections([]string{"templates"}))
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if sum.Templates.Added != 1 || sum.TemplateItemsMissed != 1 || sum.TemplateItemsAdded != 0 {
		t.Fatalf("item template yang induknya tak ada harus dilewati: %+v", sum)
	}
}

func allKeys() []string {
	keys := make([]string, 0, len(AllSectionKeys()))
	for _, k := range AllSectionKeys() {
		keys = append(keys, string(k))
	}
	return keys
}