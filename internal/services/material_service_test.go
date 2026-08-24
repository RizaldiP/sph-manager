package services

import (
	"log/slog"
	"testing"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func TestMaterialCrudAndAutoCode(t *testing.T) {
	db := serviceDB(t)
	svc := NewMaterialService(db, slog.Default())

	m1, err := svc.Create(&models.Material{Name: "Oli Sistem", Unit: "drum", DefaultPrice: 12_500_000, Supplier: "PT Sumber Teknik"})
	if err != nil {
		t.Fatalf("create material gagal: %v", err)
	}
	if m1.Code != "MAT-001" {
		t.Errorf("kode otomatis salah: %q", m1.Code)
	}
	m2, _ := svc.Create(&models.Material{Name: "Filter Udara", DefaultPrice: 750_000})
	if m2.Code != "MAT-002" {
		t.Errorf("kode otomatis berurutan salah: %q", m2.Code)
	}

	// Kode manual dipertahankan; duplikat ditolak.
	m3, err := svc.Create(&models.Material{Code: "MAT-OLI-X", Name: "Oli Hidrolika"})
	if err != nil || m3.Code != "MAT-OLI-X" {
		t.Fatalf("kode manual gagal: %v %q", err, m3.Code)
	}
	if _, err := svc.Create(&models.Material{Code: "MAT-OLI-X", Name: "Duplikat"}); err == nil {
		t.Error("kode duplikat harus ditolak")
	}

	// Validasi ramah.
	if _, err := svc.Create(&models.Material{Name: "   "}); err == nil {
		t.Error("nama kosong harus ditolak")
	}
	if _, err := svc.Create(&models.Material{Name: "Negatif", DefaultPrice: -1}); err == nil {
		t.Error("harga negatif harus ditolak")
	}

	// Update tanpa kode mempertahankan kode lama.
	upd, err := svc.Update(m1.ID, &models.Material{Name: "Oli Sistem 15W40", DefaultPrice: 13_000_000, Supplier: "PT Sumber Teknik Jaya", Unit: "drum"})
	if err != nil {
		t.Fatalf("update gagal: %v", err)
	}
	if upd.Code != "MAT-001" || upd.DefaultPrice != 13_000_000 || upd.Name != "Oli Sistem 15W40" {
		t.Errorf("hasil update salah: %+v", upd)
	}

	// Toggle aktif + pencarian.
	if err := svc.SetActive(m2.ID, false); err != nil {
		t.Fatalf("set active gagal: %v", err)
	}
	activeList, _ := svc.List(false, "")
	if len(activeList) != 2 {
		t.Errorf("daftar aktif = %d, mau 2", len(activeList))
	}
	all, _ := svc.List(true, "")
	if len(all) != 3 {
		t.Errorf("daftar lengkap = %d, mau 3", len(all))
	}
	hit, _ := svc.List(true, "sumber teknik")
	if len(hit) != 1 || hit[0].ID != m1.ID {
		t.Errorf("pencarian supplier salah: %d hasil", len(hit))
	}

	// Soft delete.
	if err := svc.Delete(m3.ID); err != nil {
		t.Fatalf("hapus gagal: %v", err)
	}
	after, _ := svc.List(true, "")
	if len(after) != 2 {
		t.Errorf("setelah hapus sisa %d, mau 2", len(after))
	}
}
