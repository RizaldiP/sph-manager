package services

import (
	"log/slog"
	"testing"
	"time"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func TestCustomerCRUDAndGuards(t *testing.T) {
	db := serviceDB(t)
	svc := NewCustomerService(db, slog.Default())

	// Validasi dasar.
	if _, err := svc.Create(&models.Customer{Name: "   "}); err == nil {
		t.Fatal("customer tanpa nama seharusnya ditolak")
	}

	c1, err := svc.Create(&models.Customer{Code: "PTL", Name: "PT Laut Biru", Phone: "0812"})
	if err != nil {
		t.Fatalf("create customer gagal: %v", err)
	}
	if c1.ID == 0 || len(c1.Vessels) != 0 {
		t.Errorf("hasil create salah: %+v", c1)
	}

	// Kode ganda ditolak.
	if _, err := svc.Create(&models.Customer{Code: "PTL", Name: "Lain"}); err == nil {
		t.Fatal("kode duplikat seharusnya ditolak")
	} else if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}

	// Update: kode kosong dipertahankan (pola master).
	upd, err := svc.Update(c1.ID, &models.Customer{Code: "", Name: "PT Laut Biru Jaya"})
	if err != nil {
		t.Fatalf("update gagal: %v", err)
	}
	if upd.Code != "PTL" || upd.Name != "PT Laut Biru Jaya" {
		t.Errorf("hasil update salah: code=%q name=%q", upd.Code, upd.Name)
	}

	// Kapal.
	if _, err := svc.CreateVessel(&models.Vessel{CustomerID: 9999, Name: "KM X"}); err == nil {
		t.Fatal("kapal untuk customer hantu seharusnya ditolak")
	}
	if _, err := svc.CreateVessel(&models.Vessel{CustomerID: c1.ID, Name: ""}); err == nil {
		t.Fatal("kapal tanpa nama seharusnya ditolak")
	}
	v1, err := svc.CreateVessel(&models.Vessel{CustomerID: c1.ID, Code: "KMB-1", Name: "KM Bahari"})
	if err != nil {
		t.Fatalf("create vessel gagal: %v", err)
	}
	if len(v1.Vessels) != 1 || v1.Vessels[0].Name != "KM Bahari" {
		t.Errorf("vessel tidak tersimpan di customer: %+v", v1.Vessels)
	}

	// Hapus customer dengan kapal → ditolak.
	if err := svc.Delete(c1.ID); err == nil {
		t.Fatal("hapus customer ber-kapal seharusnya ditolak")
	}

	// SPH memakai kapal → hapus kapal ditolak; hapus customer juga ditolak karena kapal masih ada.
	doc := &models.SphDocument{
		DocumentNumber: "SPH/GEI/I/2026/001",
		Date:           time.Now(),
		CustomerID:     c1.ID,
		VesselID:       &v1.Vessels[0].ID,
		Status:         models.StatusDraft,
	}
	if err := db.Create(doc).Error; err != nil {
		t.Fatalf("seed sph gagal: %v", err)
	}
	if err := svc.DeleteVessel(v1.Vessels[0].ID); err == nil {
		t.Fatal("hapus kapal terpakai seharusnya ditolak")
	} else if _, ok := err.(*ConflictError); !ok {
		t.Errorf("harus ConflictError, dapat: %v", err)
	}

	// Setelah dokumen SPH dihapus (hard), kapal boleh dihapus; customer tetap terhalang kapal lain jika ada.
	if err := db.Unscoped().Delete(doc).Error; err != nil {
		t.Fatalf("bersih-bersih gagal: %v", err)
	}
	if err := svc.DeleteVessel(v1.Vessels[0].ID); err != nil {
		t.Fatalf("hapus kapal bebas gagal: %v", err)
	}

	// Nonaktifkan lalu hapus bersih.
	if err := svc.SetActive(c1.ID, false); err != nil {
		t.Fatalf("nonaktif gagal: %v", err)
	}
	list, _ := svc.List(false, "")
	if len(list) != 0 {
		t.Errorf("customer nonaktif tidak boleh muncul saat includeInactive=false: %d", len(list))
	}
	listAll, _ := svc.List(true, "laut")
	if len(listAll) != 1 || len(listAll[0].Vessels) != 0 {
		t.Errorf("list lengkap salah: %+v", listAll)
	}
	if err := svc.Delete(c1.ID); err != nil {
		t.Fatalf("hapus customer bersih gagal: %v", err)
	}
	var count int64
	db.Model(&models.Customer{}).Count(&count)
	if count != 0 {
		t.Errorf("soft delete tidak bekerja: %d", count)
	}
}
