package services

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// MaterialService: CRUD master material (FR-M7, BR-05 soft delete,
// BR-13 audit, BR-15 error ramah, BR-16 transaksi).
type MaterialService struct {
	db    *gorm.DB
	log   *slog.Logger
	repo  *repositories.MaterialRepository
	audit *AuditWriter
}

func NewMaterialService(db *gorm.DB, log *slog.Logger) *MaterialService {
	return &MaterialService{
		db:    db,
		log:   log,
		repo:  repositories.NewMaterialRepository(),
		audit: NewAuditWriter(),
	}
}

func (s *MaterialService) List(includeInactive bool, search string) ([]models.Material, error) {
	rows, err := s.repo.List(s.db, includeInactive, search)
	if err != nil {
		s.log.Error("gagal mengambil daftar material", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar material")
	}
	return rows, nil
}

func (s *MaterialService) validate(in *models.Material) (string, int64, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return "", 0, NewValidationError("Nama material wajib diisi.")
	}
	if len(name) > 300 {
		return "", 0, NewValidationError("Nama material maksimal 300 karakter.")
	}
	if in.DefaultPrice < 0 {
		return "", 0, NewValidationError("Harga default material tidak boleh negatif.")
	}
	if len(strings.TrimSpace(in.Code)) > 50 {
		return "", 0, NewValidationError("Kode material maksimal 50 karakter.")
	}
	return name, in.DefaultPrice, nil
}

func (s *MaterialService) Create(in *models.Material) (*models.Material, error) {
	name, _, err := s.validate(in)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(in.Code)
	dup, err := s.repo.CodeExists(s.db, code, 0)
	if err != nil {
		s.log.Error("gagal memeriksa kode material", "error", err)
		return nil, fmt.Errorf("gagal menyimpan material")
	}
	if dup {
		return nil, NewConflictError("Kode \"%s\" sudah digunakan material lain.", code)
	}

	m := &models.Material{
		Name:         name,
		Description:  strings.TrimSpace(in.Description),
		Unit:         strings.TrimSpace(in.Unit),
		DefaultPrice: in.DefaultPrice,
		Supplier:     strings.TrimSpace(in.Supplier),
		Notes:        strings.TrimSpace(in.Notes),
		IsActive:     true,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Kode otomatis bila tidak diisi manual (konsisten master lain).
		if code != "" {
			m.Code = code
		} else if c, err := generateCode(tx, "materials", "MAT-"); err != nil {
			return err
		} else {
			m.Code = c
		}
		if err := s.repo.Create(tx, m); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "material", m.ID, fmt.Sprintf("Material \"%s\" dibuat", m.Name))
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal membuat material", "nama", name, "error", err)
		return nil, fmt.Errorf("gagal menyimpan material")
	}
	return s.repo.GetByID(s.db, m.ID)
}

func (s *MaterialService) Update(id uint, in *models.Material) (*models.Material, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil material", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui material")
	}
	name, _, err := s.validate(in)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(in.Code)
	dup, err := s.repo.CodeExists(s.db, code, id)
	if err != nil {
		s.log.Error("gagal memeriksa kode material", "error", err)
		return nil, fmt.Errorf("gagal memperbarui material")
	}
	if dup {
		return nil, NewConflictError("Kode \"%s\" sudah digunakan material lain.", code)
	}

	existing.Name = name
	existing.Description = strings.TrimSpace(in.Description)
	existing.Unit = strings.TrimSpace(in.Unit)
	existing.DefaultPrice = in.DefaultPrice
	existing.Supplier = strings.TrimSpace(in.Supplier)
	existing.Notes = strings.TrimSpace(in.Notes)
	// Kode tidak boleh kosong; input tanpa kode mempertahankan kode lama.
	if code != "" {
		existing.Code = code
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(tx, existing); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "material", existing.ID, fmt.Sprintf("Material \"%s\" diperbarui", existing.Name))
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal memperbarui material", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui material")
	}
	return s.repo.GetByID(s.db, id)
}

func (s *MaterialService) SetActive(id uint, active bool) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SetActive(tx, id, active); err != nil {
			return err
		}
		label := "diaktifkan"
		if !active {
			label = "dinonaktifkan"
		}
		return s.audit.Write(tx, "UPDATE", "material", id, fmt.Sprintf("Material %s", label))
	})
	if err != nil {
		s.log.Error("gagal mengubah status material", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status material")
	}
	return nil
}

func (s *MaterialService) Delete(id uint) error {
	// Material tidak direferensikan tabel lain (snapshot SPH menyimpan nilai),
	// jadi soft delete langsung tanpa guard pemakaian.
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDelete(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "material", id, "Material dihapus")
	})
	if err != nil {
		s.log.Error("gagal menghapus material", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus material")
	}
	return nil
}
