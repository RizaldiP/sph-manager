package services

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func TestAllocateLargestRemainder(t *testing.T) {
	cases := []struct {
		name    string
		pool    int64
		weights []int
		want    []int64
	}{
		{"contoh br-04 tepat", 100, []int{33, 33, 34}, []int64{33, 33, 34}},
		{"sisa ke pecahan terbesar", 101, []int{34, 33, 33}, []int64{35, 33, 33}},
		{"tie-break urutan sequence", 10, []int{25, 25, 25, 25}, []int64{3, 3, 2, 2}},
		{"single 100", 7_000_000, []int{100}, []int64{7_000_000}},
		{"rencana 10/15/20/25/30", 10_000_000, []int{10, 15, 20, 25, 30}, []int64{1_000_000, 1_500_000, 2_000_000, 2_500_000, 3_000_000}},
		{"pool nol", 0, []int{60, 40}, []int64{0, 0}},
		{"tanpa bobot", 500, []int{}, []int64{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := allocateLargestRemainder(c.pool, c.weights)
			if len(got) != len(c.want) {
				t.Fatalf("panjang hasil %d, mau %d", len(got), len(c.want))
			}
			if len(c.weights) == 0 {
				return
			}
			var sum int64
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("alokasi[%d] = %d, mau %d", i, got[i], c.want[i])
				}
				sum += got[i]
			}
			if sum != c.pool {
				t.Errorf("Σ alokasi %d ≠ pool %d", sum, c.pool)
			}
		})
	}
}

// weightedSphInput menyusun satu main point mode PEMBOBOTAN dengan bobot variadik.
func weightedSphInput(customerID uint, weights ...int) SphSaveInput {
	subs := make([]SphSubItemInput, len(weights))
	for i, w := range weights {
		subs[i] = SphSubItemInput{Name: fmt.Sprintf("Tahap %d", i+1), Quantity: 1, Unit: "giat", Weight: w}
	}
	return SphSaveInput{
		Header: SphHeaderInput{
			Date:        "2026-08-24",
			Sequence:    "001",
			CustomerID:  customerID,
			ProjectName: "Overhaul KM Bahari",
			Subject:     "Penawaran Overhaul Engine",
			PicName:     "Budi",
		},
		Items: []SphItemInput{{
			Name:              "Overhaul Engine",
			Quantity:          2,
			Unit:              "Unit",
			ServiceUnitPrice:  5_000_000,
			MaterialUnitPrice: 500_000,
			PricingMode:       models.PricingModeWeight,
			SubItems:          subs,
		}},
	}
}

func TestPembobotanEndToEnd(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	d1, err := svc.Create(weightedSphInput(cust.ID, 10, 15, 20, 25, 30))
	if err != nil {
		t.Fatalf("create SPH pembobotan gagal: %v", err)
	}

	// Pool: jasa 2×5jt = 10jt, material 2×500rb = 1jt → grand 11jt.
	if d1.GrandTotal != 11_000_000 {
		t.Fatalf("grand total salah: %d", d1.GrandTotal)
	}

	det, err := svc.Get(d1.ID)
	if err != nil {
		t.Fatalf("get gagal: %v", err)
	}
	if det.SubtotalService != 10_000_000 || det.SubtotalMaterial != 1_000_000 {
		t.Fatalf("subtotal salah: svc=%d mat=%d", det.SubtotalService, det.SubtotalMaterial)
	}
	it := det.Items[0]
	if it.PricingMode != models.PricingModeWeight {
		t.Fatalf("mode salah: %q", it.PricingMode)
	}
	wantSvc := []int64{1_000_000, 1_500_000, 2_000_000, 2_500_000, 3_000_000}
	wantMat := []int64{100_000, 150_000, 200_000, 250_000, 300_000}
	var sumSvc, sumMat int64
	for j, sb := range it.SubItems {
		if sb.Weight != []int{10, 15, 20, 25, 30}[j] {
			t.Errorf("weight[%d] tersimpan %d", j, sb.Weight)
		}
		if sb.ServiceTotal != wantSvc[j] || sb.MaterialTotal != wantMat[j] {
			t.Errorf("alokasi[%d] = (%d,%d), mau (%d,%d)", j, sb.ServiceTotal, sb.MaterialTotal, wantSvc[j], wantMat[j])
		}
		if sb.AllocatedValue != wantSvc[j]+wantMat[j] {
			t.Errorf("allocated_value[%d] = %d, mau %d", j, sb.AllocatedValue, wantSvc[j]+wantMat[j])
		}
		if sb.ServiceUnitPrice != 0 || sb.MaterialUnitPrice != 0 {
			t.Errorf("harga satuan sub pembobotan harus nol [%d]", j)
		}
		sumSvc += sb.ServiceTotal
		sumMat += sb.MaterialTotal
	}
	// BR-04: Σ nilai sub TEPAT sama dengan nilai main point.
	if sumSvc != 10_000_000 || sumMat != 1_000_000 {
		t.Errorf("Σ sub ≠ main: svc=%d mat=%d", sumSvc, sumMat)
	}

	// Finalisasi dengan Σ=100 harus lolos validasi BR-06.
	if err := svc.SetStatus(d1.ID, models.StatusReview); err != nil {
		t.Fatalf("review ditolak padahal Σ=100: %v", err)
	}
}

func TestPembobotanSigmaNot100BlocksFinalization(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	for i, tc := range []struct {
		name            string
		weights         []int
		wantSumText     string
		wantSelisihText string
	}{
		{"kurang dari 100", []int{40, 40}, "= 80%", "(selisih +20%)"},
		{"lebih dari 100", []int{60, 50}, "= 110%", "(selisih -10%)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := weightedSphInput(cust.ID, tc.weights...)
			in.Header.Sequence = fmt.Sprintf("%03d", i+1)
			d1, err := svc.Create(in)
			if err != nil {
				t.Fatalf("draft dengan Σ≠100 harus boleh disimpan: %v", err)
			}
			// Alokasi proporsional: Σ tetap pas meski Σ bobot belum 100.
			if d1.GrandTotal != 11_000_000 {
				t.Errorf("grand total harus tetap 11jt, dapat %d", d1.GrandTotal)
			}
			err = svc.SetStatus(d1.ID, models.StatusReview)
			if err == nil {
				t.Fatal("review harus ditolak saat Σ≠100")
			}
			if _, ok := err.(*ValidationError); !ok {
				t.Fatalf("harus ValidationError, dapat: %v", err)
			}
			if !strings.Contains(err.Error(), tc.wantSumText) || !strings.Contains(err.Error(), tc.wantSelisihText) {
				t.Errorf("pesan selisih salah: %q", err.Error())
			}
		})
	}
}

func TestPembobotanValidationErrors(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	cases := []struct {
		name     string
		weights  []int
		contains string
	}{
		{"tanpa sub point", nil, "minimal satu sub point"},
		{"bobot nol", []int{0, 100}, "harus di antara 1 dan 100"},
		{"bobot melebihi 100", []int{50, 101}, "harus di antara 1 dan 100"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(weightedSphInput(cust.ID, tc.weights...))
			if err == nil {
				t.Fatal("harus ditolak")
			}
			if !strings.Contains(err.Error(), tc.contains) {
				t.Errorf("pesan salah: %q", err.Error())
			}
		})
	}
}

func TestPembobotanTamperedAllocationRejected(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	d1, err := svc.Create(weightedSphInput(cust.ID, 60, 40))
	if err != nil {
		t.Fatalf("create gagal: %v", err)
	}
	det, _ := svc.Get(d1.ID)
	first := det.Items[0].SubItems[0]
	if err := db.Model(&models.SphSubItem{}).Where("id = ?", first.ID).
		Update("service_total", first.ServiceTotal+1).Error; err != nil {
		t.Fatalf("tamper gagal: %v", err)
	}

	err = svc.SetStatus(d1.ID, models.StatusReview)
	if err == nil {
		t.Fatal("finalisasi harus menolak alokasi yang diubah manual")
	}
	if !strings.Contains(err.Error(), "tidak konsisten") {
		t.Errorf("pesan konsistensi salah: %q", err.Error())
	}
}
