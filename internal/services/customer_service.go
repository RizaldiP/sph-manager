package services

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// ===== View =====

type VesselView struct {
	ID           uint   `json:"id"`
	CustomerID   uint   `json:"customerId"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	VesselNumber string `json:"vesselNumber"`
	VesselType   string `json:"vesselType"`
	Notes        string `json:"notes"`
	IsActive     bool   `json:"isActive"`
}

type CustomerView struct {
	ID          uint         `json:"id"`
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Address     string       `json:"address"`
	Phone       string       `json:"phone"`
	Email       string       `json:"email"`
	PicName     string       `json:"picName"`
	PicPosition string       `json:"picPosition"`
	Notes       string       `json:"notes"`
	IsActive    bool         `json:"isActive"`
	Vessels     []VesselView `json:"vessels"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`
}

func toCustomerView(c *models.Customer) CustomerView {
	v := CustomerView{
		ID:          c.ID,
		Code:        c.Code,
		Name:        c.Name,
		Address:     c.Address,
		Phone:       c.Phone,
		Email:       c.Email,
		PicName:     c.PicName,
		PicPosition: c.PicPosition,
		Notes:       c.Notes,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04"),
		Vessels:     make([]VesselView, 0, len(c.Vessels)),
	}
	for _, ve := range c.Vessels {
		v.Vessels = append(v.Vessels, VesselView{
			ID:           ve.ID,
			CustomerID:   ve.CustomerID,
			Code:         ve.Code,
			Name:         ve.Name,
			VesselNumber: ve.VesselNumber,
			VesselType:   ve.VesselType,
			Notes:        ve.Notes,
			IsActive:     ve.IsActive,
		})
	}
	return v
}

// ===== Service =====

// CustomerService: CRUD customer & kapal (FR-M5, FR-M6, BR-02 guard pemakaian, BR-13 audit).
type CustomerService struct {
	db      *gorm.DB
	log     *slog.Logger
	repo    *repositories.CustomerRepository
	sphRepo *repositories.SphRepository
	audit   *AuditWriter
}

func NewCustomerService(db *gorm.DB, log *slog.Logger) *CustomerService {
	return &CustomerService{
		db:      db,
		log:     log,
		repo:    repositories.NewCustomerRepository(),
		sphRepo: repositories.NewSphRepository(),
		audit:   NewAuditWriter(),
	}
}

func (s *CustomerService) List(includeInactive bool, search string) ([]CustomerView, error) {
	rows, err := s.repo.List(s.db, includeInactive, search)
	if err != nil {
		s.log.Error("gagal mengambil daftar customer", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar customer")
	}
	out := make([]CustomerView, 0, len(rows))
	for i := range rows {
		out = append(out, toCustomerView(&rows[i]))
	}
	return out, nil
}

func (s *CustomerService) validate(in *models.Customer, excludeID uint) (string, string, error) {
	name := strings.TrimSpace(in.Name)
	code := strings.TrimSpace(in.Code)
	if name == "" {
		return "", "", NewValidationError("Nama customer wajib diisi.")
	}
	if len(name) > 200 {
		return "", "", NewValidationError("Nama customer maksimal 200 karakter.")
	}
	if len(code) > 50 {
		return "", "", NewValidationError("Kode customer maksimal 50 karakter.")
	}
	dup, err := s.repo.CodeExists(s.db, code, excludeID)
	if err != nil {
		return "", "", fmt.Errorf("gagal menyimpan customer")
	}
	if dup {
		return "", "", NewConflictError("Kode \"%s\" sudah digunakan customer lain.", code)
	}
	return name, code, nil
}

func (s *CustomerService) Create(in *models.Customer) (*CustomerView, error) {
	name, code, err := s.validate(in, 0)
	if err != nil {
		return nil, err
	}
	c := &models.Customer{
		Code:        code,
		Name:        name,
		Address:     strings.TrimSpace(in.Address),
		Phone:       strings.TrimSpace(in.Phone),
		Email:       strings.TrimSpace(in.Email),
		PicName:     strings.TrimSpace(in.PicName),
		PicPosition: strings.TrimSpace(in.PicPosition),
		Notes:       strings.TrimSpace(in.Notes),
		IsActive:    true,
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Create(tx, c); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "customer", c.ID, fmt.Sprintf("Customer %s dibuat", c.Name))
	})
	if err != nil {
		s.log.Error("gagal membuat customer", "error", err)
		return nil, fmt.Errorf("gagal menyimpan customer")
	}
	fresh, err := s.repo.GetByID(s.db, c.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan customer")
	}
	view := toCustomerView(fresh)
	return &view, nil
}

func (s *CustomerService) Update(id uint, in *models.Customer) (*models.Customer, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil customer", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui customer")
	}
	name, code, err := s.validate(in, id)
	if err != nil {
		return nil, err
	}
	if code != "" {
		existing.Code = code
	}
	existing.Name = name
	existing.Address = strings.TrimSpace(in.Address)
	existing.Phone = strings.TrimSpace(in.Phone)
	existing.Email = strings.TrimSpace(in.Email)
	existing.PicName = strings.TrimSpace(in.PicName)
	existing.PicPosition = strings.TrimSpace(in.PicPosition)
	existing.Notes = strings.TrimSpace(in.Notes)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(tx, existing); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "customer", existing.ID, fmt.Sprintf("Customer %s diperbarui", existing.Name))
	})
	if err != nil {
		s.log.Error("gagal memperbarui customer", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui customer")
	}
	return s.repo.GetByID(s.db, id)
}

func (s *CustomerService) SetActive(id uint, active bool) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SetActive(tx, id, active); err != nil {
			return err
		}
		state := "diaktifkan"
		if !active {
			state = "dinonaktifkan"
		}
		return s.audit.Write(tx, "UPDATE", "customer", id, fmt.Sprintf("Customer %s", state))
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengubah status customer", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status customer")
	}
	return nil
}

func (s *CustomerService) Delete(id uint) error {
	c, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil customer", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus customer")
	}
	used, err := s.sphRepo.CountByCustomer(s.db, id)
	if err != nil {
		s.log.Error("gagal menghitung pemakaian customer", "error", err)
		return fmt.Errorf("gagal menghapus customer")
	}
	if used > 0 {
		return NewConflictError("Customer masih dipakai %d dokumen SPH dan tidak dapat dihapus. Nonaktifkan saja.", used)
	}
	if len(c.Vessels) > 0 {
		return NewConflictError("Masih ada %d kapal terdaftar untuk customer ini. Hapus atau pindahkan kapalnya lebih dulu.", len(c.Vessels))
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDelete(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "customer", id, fmt.Sprintf("Customer %s dihapus", c.Name))
	})
	if err != nil {
		s.log.Error("gagal menghapus customer", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus customer")
	}
	return nil
}

// ===== kapal =====

func (s *CustomerService) validateVessel(v *models.Vessel, excludeID uint) error {
	if v.CustomerID == 0 {
		return NewValidationError("Pilih customer pemilik kapal.")
	}
	var n int64
	if err := s.db.Model(&models.Customer{}).Where("id = ?", v.CustomerID).Count(&n).Error; err != nil {
		return fmt.Errorf("gagal menyimpan kapal")
	}
	if n == 0 {
		return NewValidationError("Customer tidak ditemukan.")
	}
	if strings.TrimSpace(v.Name) == "" {
		return NewValidationError("Nama kapal wajib diisi.")
	}
	dup, err := s.repo.VesselCodeExists(s.db, strings.TrimSpace(v.Code), excludeID)
	if err != nil {
		return fmt.Errorf("gagal menyimpan kapal")
	}
	if dup {
		return NewConflictError("Kode \"%s\" sudah digunakan kapal lain.", strings.TrimSpace(v.Code))
	}
	return nil
}

func (s *CustomerService) CreateVessel(in *models.Vessel) (*models.Customer, error) {
	v := &models.Vessel{
		CustomerID:   in.CustomerID,
		Code:         strings.TrimSpace(in.Code),
		Name:         strings.TrimSpace(in.Name),
		VesselNumber: strings.TrimSpace(in.VesselNumber),
		VesselType:   strings.TrimSpace(in.VesselType),
		Notes:        strings.TrimSpace(in.Notes),
		IsActive:     true,
	}
	if err := s.validateVessel(v, 0); err != nil {
		return nil, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateVessel(tx, v); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "vessel", v.ID, fmt.Sprintf("Kapal %s dibuat", v.Name))
	})
	if err != nil {
		s.log.Error("gagal membuat kapal", "error", err)
		return nil, fmt.Errorf("gagal menyimpan kapal")
	}
	return s.repo.GetByID(s.db, v.CustomerID)
}

func (s *CustomerService) UpdateVessel(id uint, in *models.Vessel) (*models.Customer, error) {
	existing, err := s.repo.GetVesselByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil kapal", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui kapal")
	}
	existing.Code = strings.TrimSpace(in.Code)
	existing.Name = strings.TrimSpace(in.Name)
	existing.VesselNumber = strings.TrimSpace(in.VesselNumber)
	existing.VesselType = strings.TrimSpace(in.VesselType)
	existing.Notes = strings.TrimSpace(in.Notes)

	if err := s.validateVessel(existing, id); err != nil {
		return nil, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateVessel(tx, existing); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "vessel", existing.ID, fmt.Sprintf("Kapal %s diperbarui", existing.Name))
	})
	if err != nil {
		s.log.Error("gagal memperbarui kapal", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui kapal")
	}
	return s.repo.GetByID(s.db, existing.CustomerID)
}

func (s *CustomerService) SetVesselActive(id uint, active bool) error {
	v, err := s.repo.GetVesselByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil kapal", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status kapal")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SetVesselActive(tx, id, active); err != nil {
			return err
		}
		state := "diaktifkan"
		if !active {
			state = "dinonaktifkan"
		}
		return s.audit.Write(tx, "UPDATE", "vessel", id, fmt.Sprintf("Kapal %s %s", v.Name, state))
	})
	if err != nil {
		s.log.Error("gagal mengubah status kapal", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status kapal")
	}
	return nil
}

func (s *CustomerService) DeleteVessel(id uint) error {
	v, err := s.repo.GetVesselByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil kapal", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus kapal")
	}
	used, err := s.sphRepo.CountByVessel(s.db, id)
	if err != nil {
		s.log.Error("gagal menghitung pemakaian kapal", "error", err)
		return fmt.Errorf("gagal menghapus kapal")
	}
	if used > 0 {
		return NewConflictError("Kapal masih dipakai %d dokumen SPH dan tidak dapat dihapus. Nonaktifkan saja.", used)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDeleteVessel(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "vessel", id, fmt.Sprintf("Kapal %s dihapus", v.Name))
	})
	if err != nil {
		s.log.Error("gagal menghapus kapal", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus kapal")
	}
	return nil
}
