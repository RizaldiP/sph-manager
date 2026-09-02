package services

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/RizaldiP/sph-manager/internal/backup"
	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/models"
)

func TestBackupCreateManualAndAudit(t *testing.T) {
	db := serviceDB(t)
	svc := NewBackupService(db, t.TempDir(), slog.Default())

	info, err := svc.CreateManual()
	if err != nil {
		t.Fatalf("CreateManual gagal: %v", err)
	}
	if !backup.IsValidName(info.Name) || info.Path == "" {
		t.Errorf("info backup tidak lengkap: %+v", info)
	}
	if _, err := os.Stat(info.Path); err != nil {
		t.Errorf("file backup tidak ada: %v", err)
	}
	var count int64
	db.Model(&models.AuditLog{}).Where("action = ?", "BACKUP").Count(&count)
	if count != 1 {
		t.Errorf("audit BACKUP tidak tercatat, count=%d", count)
	}
}

func TestBackupEnsureDailySingle(t *testing.T) {
	db := serviceDB(t)
	svc := NewBackupService(db, t.TempDir(), slog.Default())

	if _, err := svc.CreateManual(); err != nil {
		t.Fatalf("CreateManual gagal: %v", err)
	}
	if !svc.HasFileToday() {
		t.Error("HasFileToday salah setelah backup manual")
	}
	info, err := svc.EnsureDaily()
	if err != nil {
		t.Fatalf("EnsureDaily gagal: %v", err)
	}
	if info == nil || info.Name == "" {
		t.Error("EnsureDaily tidak mengembalikan info backup hari ini")
	}
	items, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups gagal: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("EnsureDaily membuat backup kedua, tersisa %d", len(items))
	}
}

func TestBackupListDelete(t *testing.T) {
	db := serviceDB(t)
	svc := NewBackupService(db, t.TempDir(), slog.Default())

	if _, err := svc.CreateManual(); err != nil {
		t.Fatalf("CreateManual gagal: %v", err)
	}
	items, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups gagal: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("jumlah backup = %d, mau 1", len(items))
	}
	if err := svc.DeleteBackup(items[0].Name); err != nil {
		t.Fatalf("DeleteBackup gagal: %v", err)
	}
	if _, err := os.Stat(items[0].Path); err == nil {
		t.Error("file backup masih ada setelah dihapus")
	}
	if err := svc.DeleteBackup("SPH_Backup_1900-01-01_000000.db"); err != ErrNotFound {
		t.Errorf("hapus yang tidak ada harus ErrNotFound, dapat: %v", err)
	}
}

func TestBackupRetentionPrunes(t *testing.T) {
	db := serviceDB(t)
	dir := t.TempDir()
	svc := NewBackupService(db, dir, slog.Default())

	for i := 1; i <= 12; i++ {
		name := fmt.Sprintf("SPH_Backup_2026-09-02_%02d0000.db", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatalf("tulis %s gagal: %v", name, err)
		}
	}
	if _, err := svc.CreateManual(); err != nil {
		t.Fatalf("CreateManual gagal: %v", err)
	}
	items, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups gagal: %v", err)
	}
	if len(items) > BackupRetentionKeep {
		t.Errorf("retention gagal: %d backup tersisa, batas %d", len(items), BackupRetentionKeep)
	}
}

func TestBackupValidateAndResolve(t *testing.T) {
	db := serviceDB(t)
	svc := NewBackupService(db, t.TempDir(), slog.Default())

	info, err := svc.CreateManual()
	if err != nil {
		t.Fatalf("CreateManual gagal: %v", err)
	}
	if _, err := svc.Validate(info.Name); err != nil {
		t.Fatalf("Validate backup valid gagal: %v", err)
	}
	if _, err := svc.Validate("SPH_Backup_1900-01-01_000000.db"); err != ErrNotFound {
		t.Errorf("validate yang tidak ada harus ErrNotFound, dapat: %v", err)
	}
	if _, err := svc.Resolve("../../SPH_Backup_2026-01-01_000000.db"); err != ErrNotFound {
		t.Errorf("Resolve traversal harus ditolak, dapat: %v", err)
	}
}

func TestBackupImportValidExternal(t *testing.T) {
	db := serviceDB(t)
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "dari-usb.db")
	sdb, err := database.Open(srcPath, slog.Default())
	if err != nil {
		t.Fatalf("Open sumber gagal: %v", err)
	}
	if err := database.Migrate(sdb); err != nil {
		t.Fatalf("Migrate sumber gagal: %v", err)
	}
	database.Close(sdb)

	svc := NewBackupService(db, t.TempDir(), slog.Default())
	info, err := svc.Import(srcPath)
	if err != nil {
		t.Fatalf("Import gagal: %v", err)
	}
	if !backup.IsValidName(info.Name) {
		t.Errorf("nama hasil import tidak valid: %q", info.Name)
	}
	items, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups gagal: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("jumlah backup setelah import = %d, mau 1", len(items))
	}
	var count int64
	db.Model(&models.AuditLog{}).Where("action = ?", "IMPORT").Count(&count)
	if count != 1 {
		t.Errorf("audit IMPORT tidak tercatat, count=%d", count)
	}
}

func TestBackupImportRejectsInvalid(t *testing.T) {
	db := serviceDB(t)
	svc := NewBackupService(db, t.TempDir(), slog.Default())

	bad := filepath.Join(t.TempDir(), "rusak.db")
	if err := os.WriteFile(bad, []byte("ini bukan sqlite"), 0644); err != nil {
		t.Fatalf("tulis file gagal: %v", err)
	}
	if _, err := svc.Import(bad); err == nil {
		t.Fatal("file korup seharusnya ditolak")
	}
	items, _ := svc.ListBackups()
	if len(items) != 0 {
		t.Errorf("file tidak valid menyisakan file di folder backup")
	}

	txt := filepath.Join(t.TempDir(), "catatan.txt")
	if err := os.WriteFile(txt, []byte("x"), 0644); err != nil {
		t.Fatalf("tulis txt gagal: %v", err)
	}
	if _, err := svc.Import(txt); err == nil {
		t.Fatal("file non-.db seharusnya ditolak")
	}
}

func TestBackupImportExistingNoDuplicate(t *testing.T) {
	db := serviceDB(t)
	dir := t.TempDir()
	svc := NewBackupService(db, dir, slog.Default())

	info, err := svc.CreateManual()
	if err != nil {
		t.Fatalf("CreateManual gagal: %v", err)
	}
	again, err := svc.Import(info.Path)
	if err != nil {
		t.Fatalf("Import ulang gagal: %v", err)
	}
	if again.Name != info.Name {
		t.Errorf("import ulang mengembalikan %q, mau %q", again.Name, info.Name)
	}
	items, _ := svc.ListBackups()
	if len(items) != 1 {
		t.Errorf("import ulang membuat duplikat: %d backup", len(items))
	}
}

func TestBackupShutdownAfterMarkClosedSafe(t *testing.T) {
	db := serviceDB(t)
	svc := NewBackupService(db, t.TempDir(), slog.Default())
	svc.MarkClosed()
	svc.BackupOnShutdown()
	svc.StopAuto()
	svc.StopAuto()
}