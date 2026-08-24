package services

import (
	"fmt"
	"log/slog"
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
)

// SettingsView: potret seluruh pengaturan aplikasi (FR-U4, kebutuhan generator dokumen FR-IE5/6).
type SettingsView struct {
	CompanyName     string `json:"companyName"`
	CompanyCity     string `json:"companyCity"`
	CompanyAddress  string `json:"companyAddress"`
	LogoPath        string `json:"logoPath"`
	SphNumberFormat string `json:"sphNumberFormat"`
	SignerName      string `json:"signerName"`
	SignerPosition  string `json:"signerPosition"`
	DefaultNotes    string `json:"defaultNotes"`
}

// SettingsInput: payload pembaruan dari UI.
type SettingsInput struct {
	CompanyName     string `json:"companyName"`
	CompanyCity     string `json:"companyCity"`
	CompanyAddress  string `json:"companyAddress"`
	SphNumberFormat string `json:"sphNumberFormat"`
	SignerName      string `json:"signerName"`
	SignerPosition  string `json:"signerPosition"`
	DefaultNotes    string `json:"defaultNotes"`
}

func defaultSettings() SettingsView {
	return SettingsView{
		CompanyName:     "PT. Ganesha Energi Indonesia",
		CompanyCity:     "Surabaya",
		SphNumberFormat: defaultSphNumberFormat,
		SignerName:      "Matawai",
		SignerPosition:  "Direktur",
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
		CompanyName:     pick(m, keyCompanyName, def.CompanyName),
		CompanyCity:     pick(m, keyCompanyCity, def.CompanyCity),
		CompanyAddress:  pick(m, keyCompanyAddress, ""),
		LogoPath:        pick(m, keyLogoPath, ""),
		SphNumberFormat: pick(m, keySphNumberFormat, def.SphNumberFormat),
		SignerName:      pick(m, keySignerName, def.SignerName),
		SignerPosition:  pick(m, keySignerPosition, def.SignerPosition),
		DefaultNotes:    pick(m, keyDefaultNotes, ""),
	}
	return v, nil
}

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
	return nil
}

func (s *SettingsService) Update(in *SettingsInput) (*SettingsView, error) {
	if err := s.validate(in); err != nil {
		return nil, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		pairs := []struct{ key, value string }{
			{keyCompanyName, trim(in.CompanyName)},
			{keyCompanyCity, trim(in.CompanyCity)},
			{keyCompanyAddress, trim(in.CompanyAddress)},
			{keySphNumberFormat, trim(in.SphNumberFormat)},
			{keySignerName, trim(in.SignerName)},
			{keySignerPosition, trim(in.SignerPosition)},
			{keyDefaultNotes, trim(in.DefaultNotes)},
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
