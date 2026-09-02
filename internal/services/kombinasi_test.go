package services

import (
	"log/slog"
	"testing"
)

// kombinasiInput membangun SPH dengan 5 pekerjaan sesuai skenario Phase 7:
// Repair AMS + Repair PLC + Repair Sensor + Testing + Calibration → 1 SPH berurut 1–5.
func kombinasiInput(customerID uint) SphSaveInput {
	names := []string{"Repair AMS", "Repair PLC", "Repair Sensor", "Testing", "Calibration"}
	in := SphSaveInput{
		Header: SphHeaderInput{
			Date:        "2026-08-24",
			Sequence:    "001",
			CustomerID:  customerID,
			ProjectName: "Docking KM Bahari",
			Subject:     "Kombinasi Multi Pekerjaan",
			PicName:     "Budi",
		},
	}
	for i, n := range names {
		it := SphItemInput{Name: n, Quantity: 1, Unit: "giat", ServiceUnitPrice: int64(1_000_000 * (i + 1))}
		if i == 1 {
			it.SubItems = []SphSubItemInput{
				{Name: "Inspection", Quantity: 1, ServiceUnitPrice: 250_000},
				{Name: "Rewiring", Quantity: 2, MaterialUnitPrice: 100_000},
			}
		}
		in.Items = append(in.Items, it)
	}
	return in
}

func TestKombinasiUrutanTersimpan(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	d, err := svc.Create(kombinasiInput(cust.ID))
	if err != nil {
		t.Fatalf("create gagal: %v", err)
	}

	det, err := svc.Get(d.ID)
	if err != nil {
		t.Fatalf("get gagal: %v", err)
	}
	want := []string{"Repair AMS", "Repair PLC", "Repair Sensor", "Testing", "Calibration"}
	if len(det.Items) != len(want) {
		t.Fatalf("jumlah item = %d, mau %d", len(det.Items), len(want))
	}
	for i, w := range want {
		got := det.Items[i]
		if got.NameSnapshot != w || got.Sequence != i+1 {
			t.Errorf("urutan salah idx %d: %q seq=%d, mau %q seq=%d", i, got.NameSnapshot, got.Sequence, w, i+1)
		}
	}

	// Sequence sub point 1..M pada item kedua (Repair PLC).
	subs := det.Items[1].SubItems
	if len(subs) != 2 {
		t.Fatalf("jumlah sub = %d, mau 2", len(subs))
	}
	if subs[0].Sequence != 1 || subs[1].Sequence != 2 {
		t.Errorf("sequence sub salah: %d, %d", subs[0].Sequence, subs[1].Sequence)
	}
}

func TestKombinasiUpdateDraftReorder(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	d, err := svc.Create(kombinasiInput(cust.ID))
	if err != nil {
		t.Fatalf("create gagal: %v", err)
	}

	in := kombinasiInput(cust.ID)
	for i, j := 0, len(in.Items)-1; i < j; i, j = i+1, j-1 {
		in.Items[i], in.Items[j] = in.Items[j], in.Items[i]
	}
	if _, err := svc.UpdateDraft(d.ID, in); err != nil {
		t.Fatalf("update draft gagal: %v", err)
	}

	det, err := svc.Get(d.ID)
	if err != nil {
		t.Fatalf("get gagal: %v", err)
	}
	wantRev := []string{"Calibration", "Testing", "Repair Sensor", "Repair PLC", "Repair AMS"}
	for i, w := range wantRev {
		got := det.Items[i]
		if got.NameSnapshot != w || got.Sequence != i+1 {
			t.Errorf("setelah reorder idx %d: %q seq=%d, mau %q seq=%d", i, got.NameSnapshot, got.Sequence, w, i+1)
		}
	}
}

func TestKombinasiDuplicatePertahankanUrutan(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	d, err := svc.Create(kombinasiInput(cust.ID))
	if err != nil {
		t.Fatalf("create gagal: %v", err)
	}
	cloneView, err := svc.Duplicate(d.ID)
	if err != nil {
		t.Fatalf("duplicate gagal: %v", err)
	}

	det, err := svc.Get(cloneView.ID)
	if err != nil {
		t.Fatalf("get salinan gagal: %v", err)
	}
	want := []string{"Repair AMS", "Repair PLC", "Repair Sensor", "Testing", "Calibration"}
	if len(det.Items) != len(want) {
		t.Fatalf("jumlah item salinan = %d, mau %d", len(det.Items), len(want))
	}
	for i, w := range want {
		got := det.Items[i]
		if got.NameSnapshot != w || got.Sequence != i+1 {
			t.Errorf("salinan urutan salah idx %d: %q seq=%d, mau %q seq=%d", i, got.NameSnapshot, got.Sequence, w, i+1)
		}
	}
	if subs := det.Items[1].SubItems; len(subs) == 2 && (subs[0].Sequence != 1 || subs[1].Sequence != 2) {
		t.Errorf("salinan sequence sub salah: %d, %d", subs[0].Sequence, subs[1].Sequence)
	}
}
