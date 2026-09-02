package services

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		t.Fatalf("parse tanggal gagal: %v", err)
	}
	return d
}

func TestComposeNumber(t *testing.T) {
	cases := []struct {
		seq  string
		date string
		want string
	}{
		{"1", "2026-08-24", "001/SPH-GEI/VIII/2026"},
		{"005", "2026-08-24", "005/SPH-GEI/VIII/2026"},
		{"012", "2026-08-24", "012/SPH-GEI/VIII/2026"},
		{"1000", "2026-08-24", "1000/SPH-GEI/VIII/2026"},
		{"7", "2026-09-01", "007/SPH-GEI/IX/2026"},
	}
	for _, c := range cases {
		got, err := composeNumber(defaultSphNumberFormat, c.seq, mustDate(t, c.date))
		if err != nil {
			t.Errorf("compose(%q, %s) error: %v", c.seq, c.date, err)
			continue
		}
		if got != c.want {
			t.Errorf("compose(%q, %s) = %q, mau %q", c.seq, c.date, got, c.want)
		}
	}

	if _, err := composeNumber(defaultSphNumberFormat, "abc", mustDate(t, "2026-08-24")); err == nil {
		t.Error("seq non-angka harus ditolak")
	}
	if _, err := composeNumber("SPH/{YYYY}/tanpa-seq", "001", mustDate(t, "2026-08-24")); err == nil {
		t.Error("format tanpa {SEQ} harus ditolak")
	}
}

func seedRawSph(t *testing.T, db *gorm.DB, cust *models.Customer, number string, date time.Time) {
	t.Helper()
	d := models.SphDocument{DocumentNumber: number, Date: date, CustomerID: cust.ID, Status: models.StatusDraft}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("seed nomor %q gagal: %v", number, err)
	}
}

func TestMaxSequenceForFormatAndSuggestion(t *testing.T) {
	db := serviceDB(t)
	svc := NewSphService(db, slog.Default())
	cust := seedSphCustomer(t, db)

	aug := mustDate(t, "2026-08-24")
	sep := mustDate(t, "2026-09-01")
	seedRawSph(t, db, cust, "001/SPH-GEI/VIII/2026", aug)
	seedRawSph(t, db, cust, "012/SPH-GEI/VIII/2026", aug)
	seedRawSph(t, db, cust, "003/SPH-GEI/IX/2026", sep)
	// Nomor bergaya lama tidak boleh ikut dihitung.
	seedRawSph(t, db, cust, "SPH/GEI/VIII/2026/999", aug)

	maxAug, err := maxSequenceForFormat(db, defaultSphNumberFormat, aug)
	if err != nil {
		t.Fatalf("maxSequenceForFormat gagal: %v", err)
	}
	if maxAug != 12 {
		t.Errorf("max urut Agustus = %d, mau 12", maxAug)
	}

	got, err := svc.SuggestNumber("2026-08-24")
	if err != nil {
		t.Fatalf("SuggestNumber gagal: %v", err)
	}
	if got != "013" {
		t.Errorf("saran Agustus = %q, mau 013", got)
	}

	comp, err := svc.ComposeNumber("1", "2026-08-24")
	if err != nil {
		t.Fatalf("ComposeNumber gagal: %v", err)
	}
	if comp != "001/SPH-GEI/VIII/2026" {
		t.Errorf("compose = %q, mau 001/SPH-GEI/VIII/2026", comp)
	}
}

func TestManualDocumentNumberDuplicate(t *testing.T) {
	db := serviceDB(t)
	aug := mustDate(t, "2026-08-24")
	cust := seedSphCustomer(t, db)
	seedRawSph(t, db, cust, "002/SPH-GEI/VIII/2026", aug)

	db.Transaction(func(tx *gorm.DB) error {
		if _, err := manualDocumentNumber(tx, "002", aug); err == nil {
			t.Error("nomor duplikat harus ditolak")
		} else if !strings.Contains(err.Error(), "sudah dipakai") {
			t.Errorf("pesan duplikat tidak ramah: %v", err)
		}
		n, err := manualDocumentNumber(tx, "3", aug)
		if err != nil {
			t.Fatalf("manual seq 3 gagal: %v", err)
		}
		if n != "003/SPH-GEI/VIII/2026" {
			t.Errorf("manual = %q, mau 003/SPH-GEI/VIII/2026", n)
		}
		return nil
	})
}