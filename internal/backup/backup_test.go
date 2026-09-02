package backup

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/models"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNameFormat(t *testing.T) {
	tm := time.Date(2026, 9, 2, 15, 4, 5, 0, time.Local)
	n := Name(tm)
	if n != "SPH_Backup_2026-09-02_150405.db" {
		t.Errorf("Name = %q, mau SPH_Backup_2026-09-02_150405.db", n)
	}
	if !IsValidName(n) {
		t.Error("nama valid ditolak IsValidName")
	}
}

func TestIsValidNameRejectsBad(t *testing.T) {
	bad := []string{
		"", "backup.db", "SPH-Backup_2026-09-02_150405.db",
		"SPH_Backup_2026-09-02.db", "SPH_Backup_2026-9-2_150405.db",
		"../SPH_Backup_2026-09-02_150405.db",
		"dir/SPH_Backup_2026-09-02_150405.db",
		"SPH_Backup_2026-09-02_150405.db.exe",
		"SPH_Backup_2026-09-02_15040x.db",
	}
	for _, b := range bad {
		if IsValidName(b) {
			t.Errorf("nama tidak valid lolos: %q", b)
		}
	}
}

func TestCreateValidateRoundtrip(t *testing.T) {
	dir := t.TempDir()
	lg := testLogger()
	db, err := database.Open(filepath.Join(dir, "main.db"), lg)
	if err != nil {
		t.Fatalf("Open gagal: %v", err)
	}
	defer database.Close(db)
	if err := database.Migrate(db); err != nil {
		t.Fatalf("Migrate gagal: %v", err)
	}
	if err := db.Create(&models.Category{Code: "EL", Name: "Electrical"}).Error; err != nil {
		t.Fatalf("buat data gagal: %v", err)
	}

	bdir := filepath.Join(dir, "backups")
	p, err := Create(db, bdir)
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}
	if filepath.Base(p) != Name(time.Now()) {
		t.Errorf("nama file = %q", filepath.Base(p))
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file backup tidak ada: %v", err)
	}
	tables, err := Validate(p, lg)
	if err != nil {
		t.Fatalf("Validate gagal: %v", err)
	}
	if len(tables) == 0 {
		t.Error("Validate mengembalikan daftar tabel kosong")
	}
}

func TestValidateRejectsCorrupt(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "corrupt.db")
	if err := os.WriteFile(p, []byte("ini bukan database sqlite"), 0644); err != nil {
		t.Fatalf("tulis file gagal: %v", err)
	}
	if _, err := Validate(p, testLogger()); err == nil {
		t.Fatal("file korup tidak ditolak Validate")
	}
}

func TestValidateRejectsNonSph(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "kosong.db")
	db, err := database.Open(p, testLogger())
	if err != nil {
		t.Fatalf("Open gagal: %v", err)
	}
	database.Close(db)
	if _, err := Validate(p, testLogger()); err == nil {
		t.Fatal("database tanpa tabel SPH tidak ditolak Validate")
	}
}

func TestListSortsAndFilters(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"SPH_Backup_2026-09-01_100000.db",
		"SPH_Backup_2026-09-03_090000.db",
		"SPH_Backup_2026-09-02_120000.db",
		"catatan.txt",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0644); err != nil {
			t.Fatalf("tulis %s gagal: %v", f, err)
		}
	}
	items, err := List(dir)
	if err != nil {
		t.Fatalf("List gagal: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("hanya backup yang difilter: %d item", len(items))
	}
	want := []string{
		"SPH_Backup_2026-09-03_090000.db",
		"SPH_Backup_2026-09-02_120000.db",
		"SPH_Backup_2026-09-01_100000.db",
	}
	for i, w := range want {
		if items[i].Name != w {
			t.Errorf("urutan %d = %q, mau %q", i, items[i].Name, w)
		}
	}
}

func TestRetainKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 12; i++ {
		name := fmt.Sprintf("SPH_Backup_2026-09-02_%02d0000.db", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatalf("tulis %s gagal: %v", name, err)
		}
	}
	removed, err := Retain(dir, 10)
	if err != nil {
		t.Fatalf("Retain gagal: %v", err)
	}
	if len(removed) != 2 {
		t.Errorf("harus menghapus 2 backup, terhapus %d: %v", len(removed), removed)
	}
	items, err := List(dir)
	if err != nil {
		t.Fatalf("List gagal: %v", err)
	}
	if len(items) != 10 {
		t.Fatalf("tersisa %d backup, mau 10", len(items))
	}
	if items[len(items)-1].Name != "SPH_Backup_2026-09-02_030000.db" {
		t.Errorf("tertua tersisa = %q, mau 030000", items[len(items)-1].Name)
	}
}

func TestNextFreeNameCollision(t *testing.T) {
	dir := t.TempDir()
	tm := time.Date(2026, 9, 2, 15, 4, 5, 0, time.Local)
	first := nextFreeName(dir, tm)
	if first != filepath.Join(dir, "SPH_Backup_2026-09-02_150405.db") {
		t.Errorf("nama pertama = %q", first)
	}
	if err := os.WriteFile(first, []byte("a"), 0644); err != nil {
		t.Fatalf("tulis pertama gagal: %v", err)
	}
	second := nextFreeName(dir, tm)
	if second != filepath.Join(dir, "SPH_Backup_2026-09-02_150405_1.db") {
		t.Errorf("nama kedua = %q, mau _1", second)
	}
	if err := os.WriteFile(second, []byte("b"), 0644); err != nil {
		t.Fatalf("tulis kedua gagal: %v", err)
	}
	third := nextFreeName(dir, tm)
	if third != filepath.Join(dir, "SPH_Backup_2026-09-02_150405_2.db") {
		t.Errorf("nama ketiga = %q, mau _2", third)
	}
}

func TestCopyIntoCopiesWithValidName(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "dari-usb.db")
	if err := os.WriteFile(src, []byte("BACKUP-CONTENT"), 0644); err != nil {
		t.Fatalf("tulis sumber gagal: %v", err)
	}
	bdir := filepath.Join(dir, "backups")
	info, err := CopyInto(src, bdir)
	if err != nil {
		t.Fatalf("CopyInto gagal: %v", err)
	}
	if !IsValidName(info.Name) {
		t.Errorf("nama hasil tidak valid: %q", info.Name)
	}
	data, err := os.ReadFile(info.Path)
	if err != nil {
		t.Fatalf("baca hasil gagal: %v", err)
	}
	if string(data) != "BACKUP-CONTENT" {
		t.Errorf("isi hasil = %q, mau BACKUP-CONTENT", string(data))
	}
	if _, err := os.Stat(src); err != nil {
		t.Error("file sumber hilang setelah disalin")
	}
}

func TestCopyIntoRejectsMissing(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "tidak-ada.db")
	if _, err := CopyInto(src, filepath.Join(dir, "backups")); err == nil {
		t.Fatal("sumber yang tidak ada seharusnya ditolak")
	}
}

func TestSwapMainReplacesFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "main.db")
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("MkdirAll gagal: %v", err)
	}
	backupPath := filepath.Join(backupDir, "SPH_Backup_2020-01-01_000000.db")
	write := func(p, s string) {
		if err := os.WriteFile(p, []byte(s), 0644); err != nil {
			t.Fatalf("tulis %s gagal: %v", p, err)
		}
	}
	write(dbPath, "ORIGINAL-DATA")
	write(dbPath+"-wal", "wal")
	write(dbPath+"-shm", "shm")
	write(backupPath, "RESTORED-DATA")

	if err := SwapMain(dbPath, backupPath); err != nil {
		t.Fatalf("SwapMain gagal: %v", err)
	}
	data, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("baca hasil gagal: %v", err)
	}
	if string(data) != "RESTORED-DATA" {
		t.Errorf("isi db = %q, mau RESTORED-DATA", string(data))
	}
	for _, sidecar := range []string{dbPath + "-wal", dbPath + "-shm"} {
		if _, err := os.Stat(sidecar); err == nil {
			t.Errorf("sidecar sisa masih ada: %s", sidecar)
		}
	}
	if _, err := os.Stat(backupPath); err != nil {
		t.Error("file backup sumber hilang setelah swap")
	}
	tmp := filepath.Join(dir, ".sph_restore_tmp.db")
	if _, err := os.Stat(tmp); err == nil {
		t.Error("file sementara tidak dibersihkan")
	}
}