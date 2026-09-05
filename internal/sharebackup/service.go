package sharebackup

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Service membangun, menulis, membaca, dan me-restore paket backup yang
// dapat dibagikan antar perangkat.
type Service struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewService(db *gorm.DB, log *slog.Logger) *Service {
	return &Service{db: db, log: log}
}

func now() time.Time {
	return time.Now()
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("pkg-%d", now().UnixNano())
	}
	return "pkg-" + hex.EncodeToString(b)
}

// nextCode membuat kode otomatis berurutan untuk sebuah tabel, mis. "TPL-007".
// Salinan logika internal/services.generateCode (tidak dapat diimpor lintas paket).
func nextCode(db *gorm.DB, table, prefix string) (string, error) {
	var codes []string
	if err := db.Raw(
		"SELECT code FROM "+table+" WHERE code LIKE ? AND deleted_at IS NULL",
		prefix+"%",
	).Scan(&codes).Error; err != nil {
		return "", err
	}
	max := 0
	for _, c := range codes {
		n, err := strconv.Atoi(strings.TrimPrefix(c, prefix))
		if err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("%s%03d", prefix, max+1), nil
}

// WriteFile menulis paket ke file, memastikan checksum terbaru sebelum disimpan.
func (s *Service) WriteFile(pkg *ShareBackupPackage, path string) error {
	sum, err := pkg.ComputeChecksum()
	if err != nil {
		return err
	}
	pkg.Checksum = sum
	raw, err := pkg.Serialize()
	if err != nil {
		return fmt.Errorf("gagal menyusun paket backup: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0644); err != nil {
		return fmt.Errorf("gagal menulis file backup: %w", err)
	}
	return nil
}

// ReadFile membaca dan memvalidasi file paket.
func ReadFile(path string) (*ShareBackupPackage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file backup: %w", err)
	}
	pkg, err := Deserialize(raw)
	if err != nil {
		return nil, fmt.Errorf("file backup rusak atau tidak valid: %w", err)
	}
	if pkg.SchemaVersion != PackageSchemaVersion {
		return nil, fmt.Errorf("versi file backup tidak didukung (ditemukan %q)", pkg.SchemaVersion)
	}
	ok, err := pkg.VerifyChecksum()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("file backup tidak valid (checksum tidak cocok)")
	}
	return pkg, nil
}