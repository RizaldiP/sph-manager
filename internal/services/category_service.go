package services

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// CategoryView adalah kategori plus jumlah pekerjaan untuk tampilan daftar.
type CategoryView struct {
	ID            uint   `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Sequence      int    `json:"sequence"`
	IsActive      bool   `json:"isActive"`
	WorkItemCount int64  `json:"workItemCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// CategoryService: CRUD kategori pekerjaan (FR-M1, FR-M8, BR-05, BR-13, BR-15).
type CategoryService struct {
	db        *gorm.DB
	log       *slog.Logger
	repo      *repositories.CategoryRepository
	workItems *repositories.WorkItemRepository
	audit     *AuditWriter
}

func NewCategoryService(db *gorm.DB, log *slog.Logger) *CategoryService {
	return &CategoryService{
		db:        db,
		log:       log,
		repo:      repositories.NewCategoryRepository(),
		workItems: repositories.NewWorkItemRepository(),
		audit:     NewAuditWriter(),
	}
}

func toCategoryViews(db *gorm.DB, cats []models.Category) ([]CategoryView, error) {
	out := make([]CategoryView, 0, len(cats))
	for _, c := range cats {
		n, err := repositories.NewWorkItemRepository().CountByCategory(db, c.ID, true)
		if err != nil {
			return nil, err
		}
		out = append(out, CategoryView{
			ID:            c.ID,
			Code:          c.Code,
			Name:          c.Name,
			Description:   c.Description,
			Sequence:      c.Sequence,
			IsActive:      c.IsActive,
			WorkItemCount: n,
			CreatedAt:     c.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:     c.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	return out, nil
}

func (s *CategoryService) List(includeInactive bool, search string) ([]CategoryView, error) {
	cats, err := s.repo.List(s.db, includeInactive, trim(search))
	if err != nil {
		s.log.Error("gagal mengambil daftar kategori", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar kategori")
	}
	views, err := toCategoryViews(s.db, cats)
	if err != nil {
		s.log.Error("gagal menghitung jumlah pekerjaan per kategori", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar kategori")
	}
	return views, nil
}

func (s *CategoryService) Get(id uint) (*models.Category, error) {
	c, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil kategori", "id", id, "error", err)
		return nil, fmt.Errorf("gagal mengambil kategori")
	}
	return c, nil
}

func (s *CategoryService) validate(c *models.Category, excludeID uint) error {
	c.Code = trim(c.Code)
	c.Name = trim(c.Name)
	c.Description = trim(c.Description)
	if c.Name == "" {
		return NewValidationError("Nama kategori wajib diisi.")
	}
	if len(c.Name) > 150 {
		return NewValidationError("Nama kategori maksimal 150 karakter.")
	}
	if c.Code != "" && len(c.Code) > 50 {
		return NewValidationError("Kode kategori maksimal 50 karakter.")
	}
	dup, err := s.repo.CodeExists(s.db, c.Code, excludeID)
	if err != nil {
		return err
	}
	if dup {
		return NewConflictError("Kode \"%s\" sudah digunakan kategori lain.", c.Code)
	}
	return nil
}

func (s *CategoryService) Create(c *models.Category) (*CategoryView, error) {
	if err := s.validate(c, 0); err != nil {
		return nil, err
	}
	seq, err := s.repo.MaxSequence(s.db)
	if err != nil {
		s.log.Error("gagal menghitung urutan terakhir", "error", err)
		return nil, fmt.Errorf("gagal menyimpan kategori")
	}
	c.ID = 0
	c.Sequence = seq + 1
	c.IsActive = true

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Create(tx, c); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "category", c.ID, fmt.Sprintf("Kategori \"%s\" dibuat", c.Name))
	})
	if err != nil {
		s.log.Error("gagal membuat kategori", "nama", c.Name, "error", err)
		return nil, fmt.Errorf("gagal menyimpan kategori")
	}
	view, err := toCategoryViews(s.db, []models.Category{*c})
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan kategori")
	}
	return &view[0], nil
}

func (s *CategoryService) Update(id uint, in *models.Category) (*CategoryView, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil kategori", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui kategori")
	}
	in.ID = existing.ID
	if err := s.validate(in, id); err != nil {
		return nil, err
	}
	existing.Code = in.Code
	existing.Name = in.Name
	existing.Description = in.Description

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(tx, existing); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "category", existing.ID, fmt.Sprintf("Kategori \"%s\" diperbarui", existing.Name))
	})
	if err != nil {
		s.log.Error("gagal memperbarui kategori", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui kategori")
	}
	view, err := toCategoryViews(s.db, []models.Category{*existing})
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui kategori")
	}
	return &view[0], nil
}

// SetActive mengaktifkan / menonaktifkan kategori tanpa menghapus datanya.
func (s *CategoryService) SetActive(id uint, active bool) error {
	if _, err := s.repo.GetByID(s.db, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil kategori", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status kategori")
	}
	desc := "dinonaktifkan"
	if active {
		desc = "diaktifkan"
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SetActive(tx, id, active); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "category", id, "Kategori "+desc)
	})
	if err != nil {
		s.log.Error("gagal mengubah status kategori", "id", id, "active", active, "error", err)
		return fmt.Errorf("gagal mengubah status kategori")
	}
	return nil
}

// Delete melakukan soft delete; ditolak bila masih memiliki pekerjaan (BR-05, BR-15).
func (s *CategoryService) Delete(id uint) error {
	c, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil kategori", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus kategori")
	}
	n, err := s.workItems.CountByCategory(s.db, id, true)
	if err != nil {
		s.log.Error("gagal menghitung pekerjaan kategori", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus kategori")
	}
	if n > 0 {
		return NewConflictError("Kategori tidak dapat dihapus karena masih memiliki %d pekerjaan. Hapus atau pindahkan pekerjaannya terlebih dahulu.", n)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDelete(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "category", id, fmt.Sprintf("Kategori \"%s\" dihapus", c.Name))
	})
	if err != nil {
		s.log.Error("gagal menghapus kategori", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus kategori")
	}
	return nil
}

// Reorder menyimpan urutan baru seluruh kategori (harus berisi semua ID).
func (s *CategoryService) Reorder(ids []uint) error {
	current, err := s.repo.IDs(s.db)
	if err != nil {
		s.log.Error("gagal mengambil urutan kategori", "error", err)
		return fmt.Errorf("gagal menyimpan urutan kategori")
	}
	if !sameIDs(current, ids) {
		return NewValidationError("Urutan kategori tidak valid. Muat ulang halaman lalu coba lagi.")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Reorder(tx, ids); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "category", 0, "Urutan kategori diubah")
	})
	if err != nil {
		s.log.Error("gagal menyimpan urutan kategori", "error", err)
		return fmt.Errorf("gagal menyimpan urutan kategori")
	}
	return nil
}
