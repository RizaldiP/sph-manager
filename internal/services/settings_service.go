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
	keyStampPath       = "stamp_path"
	keySignaturePath   = "signature_path"
	keyStampPosX       = "stamp_pos_x"
	keyStampPosY       = "stamp_pos_y"
	keyStampSize       = "stamp_size"
	keySignPosX        = "signature_pos_x"
	keySignPosY        = "signature_pos_y"
	keySignSize        = "signature_size"
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
	StampPath         string `json:"stampPath"`
	SignaturePath     string `json:"signaturePath"`
	StampPosX         float64 `json:"stampPosX"`
	StampPosY         float64 `json:"stampPosY"`
	StampSize         float64 `json:"stampSize"`
	SignaturePosX     float64 `json:"signaturePosX"`
	SignaturePosY     float64 `json:"signaturePosY"`
	SignatureSize     float64 `json:"signatureSize"`
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

// pickPos membaca nilai posisi/ukuran (fraksi 0-1) dari string; bila bukan
// angka valid dalam rentang [0,1], dipakai nilai default.
func pickPos(s string, def float64) float64 {
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || f < 0 || f > 1 {
		return def
	}
	return f
}

func (s *SettingsService) Get() (*SettingsView, error) {
	m, err := s.loadMap()
	if err != nil {
		s.log.Error("gagal memuat pengaturan", "error", err)
		return nil, fmt.Errorf("gagal memuat pengaturan")
	}
	def := defaultSettings()
	f := pick(m, keySphNumberFormat, def.SphNumberFormat)
	if f == oldDefaultSphNumberFormat {
		f = def.SphNumberFormat
	}
	v := &SettingsView{
		CompanyName:       pick(m, keyCompanyName, def.CompanyName),
		CompanyCity:       pick(m, keyCompanyCity, def.CompanyCity),
		CompanyAddress:    pick(m, keyCompanyAddress, ""),
		LogoPath:          pick(m, keyLogoPath, ""),
		StampPath:         pick(m, keyStampPath, ""),
		SignaturePath:     pick(m, keySignaturePath, ""),
		StampPosX:         pickPos(m[keyStampPosX], 0.50),
		StampPosY:         pickPos(m[keyStampPosY], 0.40),
		StampSize:         pickPos(m[keyStampSize], 0.45),
		SignaturePosX:     pickPos(m[keySignPosX], 0.50),
		SignaturePosY:     pickPos(m[keySignPosY], 0.18),
		SignatureSize:     pickPos(m[keySignSize], 0.30),
		SphNumberFormat:   f,
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

// SetStamp menyimpan path stempel hasil upload (dipakai generator PDF).
func (s *SettingsService) SetStamp(path string) (*SettingsView, error) {
	return s.setImagePath(keyStampPath, path, "Stempel perusahaan")
}

// SetSignature menyimpan path tanda tangan hasil upload (dipakai generator PDF).
func (s *SettingsService) SetSignature(path string) (*SettingsView, error) {
	return s.setImagePath(keySignaturePath, path, "Tanda tangan")
}

// setImagePath menyimpan path gambar ke setting dengan audit; path kosong
// berarti menghapus gambar.
func (s *SettingsService) setImagePath(key, path, label string) (*SettingsView, error) {
	path = trim(path)
	if len(path) > 500 {
		return nil, NewValidationError("Path gambar terlalu panjang.")
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := repositories.SetSetting(tx, key, path); err != nil {
			return err
		}
		desc := label + " diperbarui"
		if path == "" {
			desc = label + " dihapus"
		}
		return s.audit.Write(tx, "UPDATE", "settings", 0, desc)
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal menyimpan gambar", "key", key, "error", err)
		return nil, fmt.Errorf("gagal menyimpan gambar")
	}
	return s.Get()
}

// SetStampPosition menyimpan posisi & ukuran stempel (fraksi 0-1 thd blok ttd).
func (s *SettingsService) SetStampPosition(x, y, size float64) (*SettingsView, error) {
	return s.setPosition("Stempel", keyStampPosX, keyStampPosY, keyStampSize, x, y, size)
}

// SetSignaturePosition menyimpan posisi & ukuran tanda tangan (fraksi 0-1).
func (s *SettingsService) SetSignaturePosition(x, y, size float64) (*SettingsView, error) {
	return s.setPosition("Tanda tangan", keySignPosX, keySignPosY, keySignSize, x, y, size)
}

// setPosition menyimpan tiga nilai fraksi posisi/ukuran (x, y, size) yg
// di-clamp ke rentang [0,1], beserta catatan audit.
func (s *SettingsService) setPosition(label, kx, ky, ks string, x, y, size float64) (*SettingsView, error) {
	x = clamp01(x)
	y = clamp01(y)
	size = clamp01(size)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		pairs := []struct{ key, value string }{
			{kx, strconv.FormatFloat(x, 'f', -1, 64)},
			{ky, strconv.FormatFloat(y, 'f', -1, 64)},
			{ks, strconv.FormatFloat(size, 'f', -1, 64)},
		}
		for _, p := range pairs {
			if err := repositories.SetSetting(tx, p.key, p.value); err != nil {
				return err
			}
		}
		return s.audit.Write(tx, "UPDATE", "settings", 0, label+" diatur ulang posisinya")
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal menyimpan posisi", "label", label, "error", err)
		return nil, fmt.Errorf("gagal menyimpan posisi %s", label)	}
	return s.Get()
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
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
	return composeNumber(format, "001", time.Now())
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
