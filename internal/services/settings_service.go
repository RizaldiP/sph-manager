package services

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

const (
	keyCompanyName     = "company_name"
	keyCompanyCity     = "company_city"
	keyCompanyAddress  = "company_address"
	keyLogoPath        = "logo_path"
	keySphNumberFormat = "sph_number_format"
	keySignerName      = "signer_name"
	keySignerPosition  = "signer_position"
	keyDefaultNotes    = "default_notes"
	keyCollabPort      = "collab_port"
	keyCollabName      = "collab_display_name"
	keyMasterDataMax   = "masterdata_max_package_size"
)

// DefaultCollabPort adalah port WebSocket Work Together bila tidak dikonfigurasi.
const DefaultCollabPort = 48765

// DefaultMasterDataMaxPackageSize adalah batas ukuran package Master Data (byte)
// bila tidak dikonfigurasi.
const DefaultMasterDataMaxPackageSize = 1_000_000

// CollabDefaults: nilai awal untuk dialog kolaborasi (host/join room).
type CollabDefaults struct {
	DeviceName  string `json:"deviceName"`
	Port        int    `json:"port"`
	DisplayName string `json:"displayName"`
}

// SettingsView: potret seluruh pengaturan aplikasi (FR-U4, kebutuhan generator dokumen FR-IE5/6).
type SettingsView struct {
	CompanyName       string `json:"companyName"`
	CompanyCity       string `json:"companyCity"`
	CompanyAddress    string `json:"companyAddress"`
	LogoPath          string `json:"logoPath"`
	SphNumberFormat   string `json:"sphNumberFormat"`
	SignerName        string `json:"signerName"`
	SignerPosition    string `json:"signerPosition"`
	DefaultNotes      string `json:"defaultNotes"`
	CollabPort        int    `json:"collabPort"`
	CollabDisplayName string `json:"collabDisplayName"`
}

// SettingsInput: payload pembaruan dari UI.
type SettingsInput struct {
	CompanyName       string `json:"companyName"`
	CompanyCity       string `json:"companyCity"`
	CompanyAddress    string `json:"companyAddress"`
	SphNumberFormat   string `json:"sphNumberFormat"`
	SignerName        string `json:"signerName"`
	SignerPosition    string `json:"signerPosition"`
	DefaultNotes      string `json:"defaultNotes"`
	CollabPort        int    `json:"collabPort"`
	CollabDisplayName string `json:"collabDisplayName"`
}

func defaultSettings() SettingsView {
	return SettingsView{
		CompanyName:     "PT. Ganesha Energi Indonesia",
		CompanyCity:     "Surabaya",
		SphNumberFormat: defaultSphNumberFormat,
		SignerName:      "Matawai",
		SignerPosition:  "Direktur",
		CollabPort:      DefaultCollabPort,
	}
}

// SettingsService: pengaturan aplikasi berbasis tabel settings (key-value).
type SettingsService struct {
	db    *gorm.DB
	log   *slog.Logger
	audit *AuditWriter
}

func NewSettingsService(db *gorm.DB, log *slog.Logger) *SettingsService {
	return &SettingsService{db: db, log: log, audit: NewAuditWriter()}
}

func (s *SettingsService) loadMap() (map[string]string, error) {
	var rows []models.Setting
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r.Key] = r.Value
	}
	return m, nil
}

func pick(m map[string]string, key, def string) string {
	if v, ok := m[key]; ok && strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func (s *SettingsService) Get() (*SettingsView, error) {
	m, err := s.loadMap()
	if err != nil {
		s.log.Error("gagal memuat pengaturan", "error", err)
		return nil, fmt.Errorf("gagal memuat pengaturan")
	}
	def := defaultSettings()
	v := &SettingsView{
		CompanyName:       pick(m, keyCompanyName, def.CompanyName),
		CompanyCity:       pick(m, keyCompanyCity, def.CompanyCity),
		CompanyAddress:    pick(m, keyCompanyAddress, ""),
		LogoPath:          pick(m, keyLogoPath, ""),
		SphNumberFormat:   pick(m, keySphNumberFormat, def.SphNumberFormat),
		SignerName:        pick(m, keySignerName, def.SignerName),
		SignerPosition:    pick(m, keySignerPosition, def.SignerPosition),
		DefaultNotes:      pick(m, keyDefaultNotes, ""),
		CollabPort:        def.CollabPort,
		CollabDisplayName: pick(m, keyCollabName, ""),
	}
	if p, err := strconv.Atoi(strings.TrimSpace(m[keyCollabPort])); err == nil && validCollabPort(p) {
		v.CollabPort = p
	}
	return v, nil
}

func validCollabPort(p int) bool { return p >= 1024 && p <= 65535 }

func (s *SettingsService) validate(in *SettingsInput) error {
	if trim(in.CompanyName) == "" {
		return NewValidationError("Nama perusahaan wajib diisi.")
	}
	if len(trim(in.CompanyName)) > 200 {
		return NewValidationError("Nama perusahaan maksimal 200 karakter.")
	}
	if len(trim(in.CompanyCity)) > 100 {
		return NewValidationError("Kota maksimal 100 karakter.")
	}
	if len(trim(in.CompanyAddress)) > 500 {
		return NewValidationError("Alamat maksimal 500 karakter.")
	}
	if len(trim(in.SignerName)) > 100 {
		return NewValidationError("Nama penandatangan maksimal 100 karakter.")
	}
	if len(trim(in.SignerPosition)) > 100 {
		return NewValidationError("Jabatan penandatangan maksimal 100 karakter.")
	}
	if len(trim(in.DefaultNotes)) > 1000 {
		return NewValidationError("Catatan default maksimal 1000 karakter.")
	}
	format := trim(in.SphNumberFormat)
	if len(format) > 100 {
		return NewValidationError("Format nomor SPH maksimal 100 karakter.")
	}
	if !strings.Contains(format, "{SEQ}") {
		return NewValidationError("Format nomor SPH harus memuat placeholder {SEQ}.")
	}
	if in.CollabPort != 0 && !validCollabPort(in.CollabPort) {
		return NewValidationError("Port kolaborasi harus di antara 1024 dan 65535.")
	}
	if len(trim(in.CollabDisplayName)) > 100 {
		return NewValidationError("Nama tampilan kolaborasi maksimal 100 karakter.")
	}
	return nil
}

func (s *SettingsService) Update(in *SettingsInput) (*SettingsView, error) {
	if err := s.validate(in); err != nil {
		return nil, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		port := in.CollabPort
		if port == 0 {
			port = DefaultCollabPort
		}
		pairs := []struct{ key, value string }{
			{keyCompanyName, trim(in.CompanyName)},
			{keyCompanyCity, trim(in.CompanyCity)},
			{keyCompanyAddress, trim(in.CompanyAddress)},
			{keySphNumberFormat, trim(in.SphNumberFormat)},
			{keySignerName, trim(in.SignerName)},
			{keySignerPosition, trim(in.SignerPosition)},
			{keyDefaultNotes, trim(in.DefaultNotes)},
			{keyCollabPort, strconv.Itoa(port)},
			{keyCollabName, trim(in.CollabDisplayName)},
		}
		for _, p := range pairs {
			if err := repositories.SetSetting(tx, p.key, p.value); err != nil {
				return err
			}
		}
		return s.audit.Write(tx, "UPDATE", "settings", 0, "Pengaturan aplikasi diperbarui")
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal menyimpan pengaturan", "error", err)
		return nil, fmt.Errorf("gagal menyimpan pengaturan")
	}
	return s.Get()
}

// SetLogo menyimpan path logo hasil upload (dipakai generator dokumen Phase 9).
func (s *SettingsService) SetLogo(path string) (*SettingsView, error) {
	path = trim(path)
	if len(path) > 500 {
		return nil, NewValidationError("Path logo terlalu panjang.")
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := repositories.SetSetting(tx, keyLogoPath, path); err != nil {
			return err
		}
		desc := "Logo perusahaan diperbarui"
		if path == "" {
			desc = "Logo perusahaan dihapus"
		}
		return s.audit.Write(tx, "UPDATE", "settings", 0, desc)
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal menyimpan logo", "error", err)
		return nil, fmt.Errorf("gagal menyimpan logo")
	}
	return s.Get()
}

// PreviewNumber merender format dengan tanggal hari ini dan contoh urut 001.
func (s *SettingsService) PreviewNumber(format string) (string, error) {
	format = trim(format)
	if format == "" {
		format = defaultSphNumberFormat
	}
	if !strings.Contains(format, "{SEQ}") {
		return "", NewValidationError("Format nomor SPH harus memuat placeholder {SEQ}.")
	}
	prefix := buildNumberPrefix(format, time.Now())
	if prefix == "" {
		return "", NewValidationError("Format nomor SPH tidak valid.")
	}
	return prefix + fmt.Sprintf("%03d", 1), nil
}

// CollabPortOrDefault membaca port WebSocket Work Together dari settings
// dengan fallback ke default bila tidak valid.
func (s *SettingsService) CollabPortOrDefault() int {
	v, err := repositories.SettingValue(s.db, keyCollabPort)
	if err != nil {
		return DefaultCollabPort
	}
	if p, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && validCollabPort(p) {
		return p
	}
	return DefaultCollabPort
}

// CollabDisplayNameOrDefault mengembalikan nama tampilan tersimpan untuk kolaborasi.
func (s *SettingsService) CollabDisplayNameOrDefault() string {
	v, err := repositories.SettingValue(s.db, keyCollabName)
	if err != nil || trim(v) == "" {
		return ""
	}
	return v
}

// MasterDataMaxPackageSizeOrDefault membaca batas ukuran package Master Data (byte)
// dari settings dengan fallback ke default bila tidak valid.
func (s *SettingsService) MasterDataMaxPackageSizeOrDefault() int {
	v, err := repositories.SettingValue(s.db, keyMasterDataMax)
	if err != nil {
		return DefaultMasterDataMaxPackageSize
	}
	if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
		return n
	}
	return DefaultMasterDataMaxPackageSize
}
