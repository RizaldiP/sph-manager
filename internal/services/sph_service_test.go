package services

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// seedSphCustomer membuat customer sungguhan untuk header SPH.
func seedSphCustomer(t *testing.T, db *gorm.DB) *models.Customer {
	t.Helper()
	c := &models.Customer{Code: "CUS-TST", Name: "PT Laut Biru"}
	if err := db.Create(c).Error; err != nil {
		t.Fatalf("seed customer gagal: %v", err)
	}
	return c
}

func TestTerbilangConversion(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "Nol Rupiah"},
		{1, "Satu Rupiah"},
		{11, "Sebelas Rupiah"},
		{15, "Lima Belas Rupiah"},
		{100, "Seratus Rupiah"},
		{1_000, "Seribu Rupiah"},
		{2_500, "Dua Ribu Lima Ratus Rupiah"},
		{10_000, "Sepuluh Ribu Rupiah"},
		{167_300_000, "Seratus Enam Puluh Tujuh Juta Tiga Ratus Ribu Rupiah"},
		{11_000_000, "Sebelas Juta Rupiah"},
		{1_000_000_000_000, "Satu Triliun Rupiah"},
		{-50, "Minus Lima Puluh Rupiah"},
	}
	for _, c := range cases {
		if got := Terbilang(c.in); got != c.want {
			t.Errorf("Terbilang(%d) = %q, mau %q", c.in, got, c.want)
		}
	}
}

func sampleSphInput(customerID uint) SphSaveInput {
	return SphSaveInput{
		Header: SphHeaderInput{
			Date:        time.Now().Format("2006-01-02"),
			Sequence:    "001",
			CustomerID:  customerID,
			ProjectName: "Docking KM Bahari",
			Subject:     "Penawaran Repair PLC",
			PicName:     "Budi",
		},
		Items: []SphItemInput{
			{Name: "Repair PLC", Quantity: 2, Unit: "Set", ServiceUnitPrice: 5_000_000, MaterialUnitPrice: 500_000,
				SubItems: []SphSubItemInput{{Name: "Inspection", Quantity: 1, ServiceUnitPrice: 250_000}}},
			{Name: "Wiring Check", Quantity: 1, Unit: "giat", ServiceUnitPrice: 3_000_000},
		},
	}
}

func TestSphCreateSnapshotAndNumbering(t *testing.T) {
	db := serviceDB(t)
	catSvc := NewCategoryService(db, slog.Default())
	wiSvc := NewWorkItemService(db, slog.Default())
	svc := NewSphService(db, slog.Default())

	cat, _ := catSvc.Create(&models.Category{Code: "EL", Name: "Electrical"})
	wi, err := wiSvc.Create(&models.WorkItem{CategoryID: cat.ID, Name: "Repair PLC", DefaultQuantity: 1, DefaultServicePrice: 9_999_999})
	if err != nil {
		t.Fatalf("create pekerjaan gagal: %v", err)
	}

	cust := seedSphCustomer(t, db)
	in := sampleSphInput(cust.ID)
	in.Items[0].WorkItemID = &wi.ID
	vessel := models.Vessel{CustomerID: cust.ID, Code: "KPL-TST", Name: "KM Bahari"}
	if err := db.Create(&vessel).Error; err != nil {
		t.Fatalf("seed kapal gagal: %v", err)
	}
	in.Header.VesselID = &vessel.ID

	d1, err := svc.Create(in)
	if err != nil {
		t.Fatalf("create SPH gagal: %v", err)
	}
	expectedSeq1 := fmt.Sprintf("001/SPH-GEI/%s/%04d", romanMonths[int(time.Now().Month())-1], time.Now().Year())
	if !strings.HasPrefix(d1.DocumentNumber, expectedSeq1) {
		t.Errorf("nomor tidak sesuai format BR-07: %q", d1.DocumentNumber)
	}
	if d1.Status != models.StatusDraft || d1.Revision != 0 {
		t.Errorf("status/revision salah: %+v", d1)
	}
	if d1.GrandTotal != 2*5_000_000+2*500_000+3_000_000+250_000 {
		t.Errorf("grand total salah: %d", d1.GrandTotal)
	}
	if d1.Terbilang == "" {
		t.Error("terbilang kosong")
	}

	// Roll-up: nilai pekerjaan induk mencakup sub point-nya (BR-01).
	if det, err := svc.Get(d1.ID); err != nil {
		t.Fatalf("get SPH gagal: %v", err)
	} else if len(det.Items) > 0 && det.Items[0].Total != 2*(5_000_000+500_000)+250_000 {
		t.Errorf("roll-up item induk salah: %d", det.Items[0].Total)
	}

	// Regresi halaman detail: relasi customer/kapal ter-preload + terbilang tersedia.
	det, err := svc.Get(d1.ID)
	if err != nil {
		t.Fatalf("get SPH gagal: %v", err)
	}
	if det.Customer.ID != cust.ID || det.Customer.Name != "PT Laut Biru" {
		t.Errorf("customer tidak ikut di detail: %+v", det.Customer)
	}
	if det.Vessel.ID != vessel.ID || det.Vessel.Name != "KM Bahari" {
		t.Errorf("kapal tidak ikut di detail: %+v", det.Vessel)
	}
	if det.Terbilang == "" || !strings.Contains(det.Terbilang, "Juta") {
		t.Errorf("terbilang detail salah: %q", det.Terbilang)
	}

	// Nomor urut diinput manual: dokumen kedua dengan urut berbeda pada periode sama.
	in.Header.Sequence = "005"
	d2, err := svc.Create(in)
	if err != nil {
		t.Fatalf("create SPH kedua gagal: %v", err)
	}
	expectedSeq2 := fmt.Sprintf("005/SPH-GEI/%s/%04d", romanMonths[int(time.Now().Month())-1], time.Now().Year())
	if d2.DocumentNumber != expectedSeq2 {
		t.Errorf("penomoran manual salah: %q", d2.DocumentNumber)
	}

	// Nomor urut yang sudah dipakai harus ditolak (BR-07).
	in.Header.Sequence = "001"
	if _, err := svc.Create(in); err == nil {
		t.Error("nomor urut duplikat harus ditolak")
	} else if !strings.Contains(err.Error(), "sudah dipakai") {
		t.Errorf("pesan duplikat tidak ramah: %v", err)
	}

	// BR-01: ubah master → snapshot dokumen lama tetap.
	if _, err := wiSvc.Update(wi.ID, &models.WorkItem{CategoryID: cat.ID, Name: "Repair PLC Baru", DefaultQuantity: 3, DefaultServicePrice: 15_000_000}); err != nil {
		t.Fatalf("update master gagal: %v", err)
	}
	got, _ := svc.Get(d1.ID)
	if len(got.Items) != 2 || got.Items[0].NameSnapshot != "Repair PLC" || got.Items[0].ServiceUnitPrice != 5_000_000 {
		t.Errorf("snapshot berubah saat master diubah (BR-01): %+v", got.Items[0])
	}
	if got.GrandTotal != d1.GrandTotal {
		t.Errorf("total dokumen lama berubah: %d vs %d", got.GrandTotal, d1.GrandTotal)
	}

	var auditCount int64
	db.Model(&models.AuditLog{}).Where("entity = ?", "sph_document").Count(&auditCount)
	if auditCount < 2 {
		t.Errorf("audit log kurang: %d", auditCount)
	}
}

func TestSphValidationLifecycleAndRules(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	// Header wajib.
	if _, err := svc.Create(SphSaveInput{Header: SphHeaderInput{}}); err == nil {
		t.Fatal("SPH tanpa tanggal/customer seharusnya ditolak")
	}
	bad := sampleSphInput(9999)
	if _, err := svc.Create(bad); err == nil {
		t.Fatal("customer hantu seharusnya ditolak")
	}

	doc, err := svc.Create(sampleSphInput(cust.ID))
	if err != nil {
		t.Fatalf("create gagal: %v", err)
	}

	// Finalisasi dengan qty tidak valid ditolak (BR-06).
	in := sampleSphInput(cust.ID)
	in.Items[0].Quantity = 0
	upd, _ := svc.Get(doc.ID)
	_ = upd
	if _, err := svc.UpdateDraft(doc.ID, in); err == nil {
		t.Fatal("qty 0 seharusnya ditolak")
	}

	// Transisi ilegal ditolak (BR-08).
	if err := svc.SetStatus(doc.ID, models.StatusSent); err == nil {
		t.Fatal("DRAFT ke SENT seharusnya ditolak")
	} else if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}

	// DRAFT → REVIEW → FINAL (finalisasi mencatat finalized_at).
	if err := svc.SetStatus(doc.ID, models.StatusReview); err != nil {
		t.Fatalf("ke review gagal: %v", err)
	}
	if err := svc.SetStatus(doc.ID, models.StatusFinal); err != nil {
		t.Fatalf("finalisasi gagal: %v", err)
	}
	got, _ := svc.Get(doc.ID)
	if got.Status != models.StatusFinal || got.FinalizedAt == nil {
		t.Errorf("finalisasi tidak tercatat: %+v", got)
	}

	// Dokumen final tidak bisa diedit isinya (BR-08).
	lockIn := sampleSphInput(cust.ID)
	lockIn.Header.ProjectName = "Diubah"
	if _, err := svc.UpdateDraft(doc.ID, lockIn); err == nil {
		t.Fatal("edit dokumen final seharusnya ditolak")
	}

	// FINAL → SENT → ACCEPTED.
	if err := svc.SetStatus(doc.ID, models.StatusSent); err != nil {
		t.Fatalf("ke sent gagal: %v", err)
	}
	if err := svc.SetStatus(doc.ID, models.StatusAccepted); err != nil {
		t.Fatalf("ke accepted gagal: %v", err)
	}

	// Hanya draft boleh dihapus.
	if err := svc.Delete(doc.ID); err == nil {
		t.Fatal("hapus dokumen accepted seharusnya ditolak")
	}
}

func TestSphDuplicateAndRevision(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	src, err := svc.Create(sampleSphInput(cust.ID))
	if err != nil {
		t.Fatalf("create gagal: %v", err)
	}

	// Revisi hanya untuk dokumen final ke atas (BR-10).
	if _, err := svc.CreateRevision(src.ID); err == nil {
		t.Fatal("revisi dari draft seharusnya ditolak")
	}

	if err := svc.SetStatus(src.ID, models.StatusReview); err != nil {
		t.Fatalf("review gagal: %v", err)
	}
	if err := svc.SetStatus(src.ID, models.StatusFinal); err != nil {
		t.Fatalf("final gagal: %v", err)
	}

	// Duplicate → nomor baru, status draft, isi identik (BR-09).
	cp, err := svc.Duplicate(src.ID)
	if err != nil {
		t.Fatalf("duplicate gagal: %v", err)
	}
	if cp.ID == src.ID || cp.DocumentNumber == src.DocumentNumber || cp.Status != models.StatusDraft || cp.Revision != 0 {
		t.Errorf("hasil duplicate salah: %+v", cp)
	}
	if cp.GrandTotal != src.GrandTotal || cp.ItemCount != src.ItemCount {
		t.Errorf("isi duplikat beda: %+v vs %+v", cp, src)
	}
	cpDetail, _ := svc.Get(cp.ID)
	if len(cpDetail.Items[0].SubItems) != 1 || cpDetail.Items[0].SubItems[0].NameSnapshot != "Inspection" {
		t.Errorf("sub item tidak ikut tersalin: %+v", cpDetail.Items[0].SubItems)
	}

	// Revision → nomor sama, rev+1, salinan penuh, histori tercatat (BR-10).
	rev, err := svc.CreateRevision(src.ID)
	if err != nil {
		t.Fatalf("revision gagal: %v", err)
	}
	if rev.DocumentNumber != src.DocumentNumber || rev.Revision != 1 || rev.Status != models.StatusDraft {
		t.Errorf("revisi salah: %+v", rev)
	}
	hist, _ := svc.Get(rev.ID)
	if len(hist.Revisions) != 1 || hist.Revisions[0].FromDocumentID == nil || *hist.Revisions[0].FromDocumentID != src.ID {
		t.Errorf("histori revisi salah: %+v", hist.Revisions)
	}

	// Dokumen asal tak berubah sedikit pun setelah duplicate & revision.
	after, _ := svc.Get(src.ID)
	if after.Status != models.StatusFinal || len(after.Items) != 2 || after.GrandTotal != src.GrandTotal {
		t.Errorf("dokumen sumber berubah: %+v", after)
	}
}

func TestSphScopesStatsAndDeleteDraft(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	a, err := svc.Create(sampleSphInput(cust.ID)) // draft
	if err != nil {
		t.Fatalf("create draft gagal: %v", err)
	}
	inB := sampleSphInput(cust.ID)
	inB.Header.Sequence = "002"
	b, err := svc.Create(inB)
	if err != nil {
		t.Fatalf("create kedua gagal: %v", err)
	}
	svc.SetStatus(b.ID, models.StatusReview)
	svc.SetStatus(b.ID, models.StatusFinal)

	openList, _ := svc.List("open", "", 0)
	if len(openList) != 1 || openList[0].ID != a.ID {
		t.Errorf("scope open salah: %+v", openList)
	}
	finalList, _ := svc.List("final", "", 0)
	if len(finalList) != 1 || finalList[0].ID != b.ID {
		t.Errorf("scope final salah: %+v", finalList)
	}
	allList, _ := svc.List("", "", 0)
	if len(allList) != 2 {
		t.Errorf("scope semua salah: %d", len(allList))
	}
	found, _ := svc.List("", "PLC", 0)
	if len(found) != 2 {
		t.Errorf("search project gagal: %d", len(found))
	}

	st, err := svc.Stats()
	if err != nil {
		t.Fatalf("stats gagal: %v", err)
	}
	if st.TotalSph != 2 || st.DraftCount != 1 || st.FinalCount != 1 || st.AcceptedCount != 0 {
		t.Errorf("statistik salah: %+v", st)
	}
	if st.MonthValue <= 0 {
		t.Error("nilai bulan ini harus positif")
	}
	if len(st.Recent) != 2 {
		t.Errorf("recent salah: %d", len(st.Recent))
	}

	// Draft bisa dihapus; setelah itu statistik ikut turun.
	if err := svc.Delete(a.ID); err != nil {
		t.Fatalf("hapus draft gagal: %v", err)
	}
	allAfter, _ := svc.List("", "", 0)
	if len(allAfter) != 1 {
		t.Errorf("draft tidak hilang dari daftar: %d", len(allAfter))
	}
}
