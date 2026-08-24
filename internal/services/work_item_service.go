package services

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// WorkItemView adalah pekerjaan plus jumlah sub-pekerjaan untuk tampilan daftar.
type WorkItemView struct {
	ID                   uint    `json:"id"`
	CategoryID           uint    `json:"categoryId"`
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	DefaultUnit          string  `json:"defaultUnit"`
	DefaultQuantity      float64 `json:"defaultQuantity"`
	DefaultServicePrice  int64   `json:"defaultServicePrice"`
	DefaultMaterialPrice int64   `json:"defaultMaterialPrice"`
	Notes                string  `json:"notes"`
	Sequence             int     `json:"sequence"`
	IsActive             bool    `json:"isActive"`
	SubItemCount         int64   `json:"subItemCount"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

// WorkItemService: CRUD master pekerjaan (FR-M2, FR-M8, BR-05, BR-13, BR-15).
type WorkItemService struct {
	db       *gorm.DB
	log      *slog.Logger
	repo     *repositories.WorkItemRepository
	subRepo  *repositories.WorkSubItemRepository
	tplItems *repositories.TemplateItemRepository
	audit    *AuditWriter
}

func NewWorkItemService(db *gorm.DB, log *slog.Logger) *WorkItemService {
	return &WorkItemService{
		db:       db,
		log:      log,
		repo:     repositories.NewWorkItemRepository(),
		subRepo:  repositories.NewWorkSubItemRepository(),
		tplItems: repositories.NewTemplateItemRepository(),
		audit:    NewAuditWriter(),
	}
}

func toWorkItemViews(db *gorm.DB, items []models.WorkItem) ([]WorkItemView, error) {
	subRepo := repositories.NewWorkSubItemRepository()
	out := make([]WorkItemView, 0, len(items))
	for _, w := range items {
		n, err := subRepo.CountByWorkItem(db, w.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, WorkItemView{
			ID:                   w.ID,
			CategoryID:           w.CategoryID,
			Code:                 w.Code,
			Name:                 w.Name,
			Description:          w.Description,
			DefaultUnit:          w.DefaultUnit,
			DefaultQuantity:      w.DefaultQuantity,
			DefaultServicePrice:  w.DefaultServicePrice,
			DefaultMaterialPrice: w.DefaultMaterialPrice,
			Notes:                w.Notes,
			Sequence:             w.Sequence,
			IsActive:             w.IsActive,
			SubItemCount:         n,
			CreatedAt:            w.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:            w.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	return out, nil
}

func (s *WorkItemService) List(categoryID uint, includeInactive bool, search string) ([]WorkItemView, error) {
	items, err := s.repo.List(s.db, categoryID, includeInactive, trim(search))
	if err != nil {
		s.log.Error("gagal mengambil daftar pekerjaan", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar pekerjaan")
	}
	views, err := toWorkItemViews(s.db, items)
	if err != nil {
		s.log.Error("gagal menghitung jumlah sub-pekerjaan", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar pekerjaan")
	}
	return views, nil
}

// GetDetail mengembalikan pekerjaan lengkap dengan sub-pekerjaannya (terurut).
func (s *WorkItemService) GetDetail(id uint) (*models.WorkItem, error) {
	w, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil pekerjaan", "id", id, "error", err)
		return nil, fmt.Errorf("gagal mengambil pekerjaan")
	}
	return w, nil
}

func (s *WorkItemService) validate(w *models.WorkItem, excludeID uint) error {
	w.Code = trim(w.Code)
	w.Name = trim(w.Name)
	w.Description = trim(w.Description)
	w.DefaultUnit = trim(w.DefaultUnit)
	w.Notes = trim(w.Notes)
	if w.CategoryID == 0 {
		return NewValidationError("Kategori wajib dipilih.")
	}
	ok, err := s.repo.CategoryExists(s.db, w.CategoryID)
	if err != nil {
		return err
	}
	if !ok {
		return NewValidationError("Kategori tidak ditemukan. Pilih kategori yang tersedia.")
	}
	if w.Name == "" {
		return NewValidationError("Nama pekerjaan wajib diisi.")
	}
	if len(w.Name) > 300 {
		return NewValidationError("Nama pekerjaan maksimal 300 karakter.")
	}
	if w.Code != "" && len(w.Code) > 50 {
		return NewValidationError("Kode pekerjaan maksimal 50 karakter.")
	}
	if w.DefaultQuantity <= 0 {
		return NewValidationError("Qty default harus lebih besar dari 0.")
	}
	if w.DefaultServicePrice < 0 || w.DefaultMaterialPrice < 0 {
		return NewValidationError("Harga tidak boleh negatif.")
	}
	dup, err := s.repo.CodeExists(s.db, w.Code, excludeID)
	if err != nil {
		return err
	}
	if dup {
		return NewConflictError("Kode \"%s\" sudah digunakan pekerjaan lain.", w.Code)
	}
	return nil
}

func (s *WorkItemService) Create(w *models.WorkItem) (*models.WorkItem, error) {
	if err := s.validate(w, 0); err != nil {
		return nil, err
	}
	seq, err := s.repo.MaxSequenceInCategory(s.db, w.CategoryID)
	if err != nil {
		s.log.Error("gagal menghitung urutan terakhir", "error", err)
		return nil, fmt.Errorf("gagal menyimpan pekerjaan")
	}
	w.ID = 0
	w.Sequence = seq + 1
	w.IsActive = true

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Kode otomatis bila tidak diisi manual (FR-U3: non-teknis).
		if w.Code == "" {
			code, err := generateCode(tx, "work_items", "PEK-")
			if err != nil {
				return err
			}
			w.Code = code
		}
		if err := s.repo.Create(tx, w); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "work_item", w.ID, fmt.Sprintf("Pekerjaan \"%s\" dibuat", w.Name))
	})
	if err != nil {
		s.log.Error("gagal membuat pekerjaan", "nama", w.Name, "error", err)
		return nil, fmt.Errorf("gagal menyimpan pekerjaan")
	}
	return s.GetDetail(w.ID)
}

func (s *WorkItemService) Update(id uint, in *models.WorkItem) (*models.WorkItem, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil pekerjaan", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui pekerjaan")
	}
	in.ID = existing.ID
	if err := s.validate(in, id); err != nil {
		return nil, err
	}
	existing.CategoryID = in.CategoryID
	// Kode dikelola sistem; kosong berarti pertahankan kode yang sudah ada.
	if in.Code != "" {
		existing.Code = in.Code
	}
	existing.Name = in.Name
	existing.Description = in.Description
	existing.DefaultUnit = in.DefaultUnit
	existing.DefaultQuantity = in.DefaultQuantity
	existing.DefaultServicePrice = in.DefaultServicePrice
	existing.DefaultMaterialPrice = in.DefaultMaterialPrice
	existing.Notes = in.Notes

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(tx, existing); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "work_item", existing.ID, fmt.Sprintf("Pekerjaan \"%s\" diperbarui", existing.Name))
	})
	if err != nil {
		s.log.Error("gagal memperbarui pekerjaan", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui pekerjaan")
	}
	return s.GetDetail(existing.ID)
}
