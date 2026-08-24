package services

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// WorkSubItemService: CRUD sub-pekerjaan per pekerjaan (FR-M3, FR-M8).
type WorkSubItemService struct {
	db       *gorm.DB
	log      *slog.Logger
	repo     *repositories.WorkSubItemRepository
	itemRepo *repositories.WorkItemRepository
	audit    *AuditWriter
}

func NewWorkSubItemService(db *gorm.DB, log *slog.Logger) *WorkSubItemService {
	return &WorkSubItemService{
		db:       db,
		log:      log,
		repo:     repositories.NewWorkSubItemRepository(),
		itemRepo: repositories.NewWorkItemRepository(),
		audit:    NewAuditWriter(),
	}
}

func (s *WorkSubItemService) validate(sub *models.WorkSubItem) error {
	sub.Code = trim(sub.Code)
	sub.Name = trim(sub.Name)
	sub.Description = trim(sub.Description)
	sub.DefaultUnit = trim(sub.DefaultUnit)
	sub.Notes = trim(sub.Notes)
	if sub.WorkItemID == 0 {
		return NewValidationError("Pekerjaan induk tidak valid.")
	}
	ok, err := s.repo.WorkItemExists(s.db, sub.WorkItemID)
	if err != nil {
		return err
	}
	if !ok {
		return NewValidationError("Pekerjaan induk tidak ditemukan.")
	}
	if sub.Name == "" {
		return NewValidationError("Nama sub-pekerjaan wajib diisi.")
	}
	if len(sub.Name) > 300 {
		return NewValidationError("Nama sub-pekerjaan maksimal 300 karakter.")
	}
	if sub.DifficultyWeight < 0 || sub.DifficultyWeight > 100 {
		return NewValidationError("Bobot kesulitan harus di antara 0 dan 100.")
	}
	if sub.DefaultQuantity <= 0 {
		return NewValidationError("Qty default harus lebih besar dari 0.")
	}
	if sub.DefaultServicePrice < 0 || sub.DefaultMaterialPrice < 0 {
		return NewValidationError("Harga tidak boleh negatif.")
	}
	return nil
}

func (s *WorkSubItemService) Create(sub *models.WorkSubItem) (*models.WorkItem, error) {
	if err := s.validate(sub); err != nil {
		return nil, err
	}
	seq, err := s.repo.MaxSequenceInWorkItem(s.db, sub.WorkItemID)
	if err != nil {
		s.log.Error("gagal menghitung urutan terakhir", "error", err)
		return nil, fmt.Errorf("gagal menyimpan sub-pekerjaan")
	}
	sub.ID = 0
	sub.Sequence = seq + 1
	sub.IsActive = true

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Kode otomatis bila tidak diisi manual (FR-U3: non-teknis).
		if sub.Code == "" {
			code, err := generateCode(tx, "work_sub_items", "SUB-")
			if err != nil {
				return err
			}
			sub.Code = code
		}
		if err := s.repo.Create(tx, sub); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "work_sub_item", sub.ID, fmt.Sprintf("Sub-pekerjaan \"%s\" dibuat", sub.Name))
	})
	if err != nil {
		s.log.Error("gagal membuat sub-pekerjaan", "nama", sub.Name, "error", err)
		return nil, fmt.Errorf("gagal menyimpan sub-pekerjaan")
	}
	return s.itemRepo.GetByID(s.db, sub.WorkItemID)
}

func (s *WorkSubItemService) Update(id uint, in *models.WorkSubItem) (*models.WorkItem, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil sub-pekerjaan", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui sub-pekerjaan")
	}
	in.ID = existing.ID
	in.WorkItemID = existing.WorkItemID // sub item tidak berpindah induk
	if err := s.validate(in); err != nil {
		return nil, err
	}
	// Kode dikelola sistem; kosong berarti pertahankan kode yang sudah ada.
	if in.Code != "" {
		existing.Code = in.Code
	}
	existing.Name = in.Name
	existing.Description = in.Description
	existing.DifficultyWeight = in.DifficultyWeight
	existing.DefaultUnit = in.DefaultUnit
	existing.DefaultQuantity = in.DefaultQuantity
	existing.DefaultServicePrice = in.DefaultServicePrice
	existing.DefaultMaterialPrice = in.DefaultMaterialPrice
	existing.Notes = in.Notes

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(tx, existing); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "work_sub_item", existing.ID, fmt.Sprintf("Sub-pekerjaan \"%s\" diperbarui", existing.Name))
	})
	if err != nil {
		s.log.Error("gagal memperbarui sub-pekerjaan", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui sub-pekerjaan")
	}
	return s.itemRepo.GetByID(s.db, existing.WorkItemID)
}

// SetActive mengaktifkan / menonaktifkan sub-pekerjaan.
func (s *WorkSubItemService) SetActive(id uint, active bool) error {
	if _, err := s.repo.GetByID(s.db, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil sub-pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status sub-pekerjaan")
	}
	desc := "dinonaktifkan"
	if active {
		desc = "diaktifkan"
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SetActive(tx, id, active); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "work_sub_item", id, "Sub-pekerjaan "+desc)
	})
	if err != nil {
		s.log.Error("gagal mengubah status sub-pekerjaan", "id", id, "active", active, "error", err)
		return fmt.Errorf("gagal mengubah status sub-pekerjaan")
	}
	return nil
}

// Delete melakukan soft delete sub-pekerjaan.
func (s *WorkSubItemService) Delete(id uint) error {
	sub, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil sub-pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus sub-pekerjaan")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDelete(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "work_sub_item", id, fmt.Sprintf("Sub-pekerjaan \"%s\" dihapus", sub.Name))
	})
	if err != nil {
		s.log.Error("gagal menghapus sub-pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus sub-pekerjaan")
	}
	return nil
}

// DeleteMany menghapus banyak sub-pekerjaan sekaligus dalam satu transaksi;
// seluruh batch dibatalkan bila ada ID yang sudah tidak tersedia.
func (s *WorkSubItemService) DeleteMany(ids []uint) (*DeleteResult, error) {
	uniq, err := uniqueIDs(ids)
	if err != nil {
		return nil, NewValidationError("Pilih minimal satu sub-pekerjaan untuk dihapus.")
	}
	subs := make([]*models.WorkSubItem, 0, len(uniq))
	for _, id := range uniq {
		sub, err := s.repo.GetByID(s.db, id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, NewConflictError("Sebagian sub-pekerjaan sudah tidak tersedia (ID %d). Muat ulang halaman lalu coba lagi.", id)
			}
			s.log.Error("gagal mengambil sub-pekerjaan", "id", id, "error", err)
			return nil, fmt.Errorf("gagal menghapus sub-pekerjaan")
		}
		subs = append(subs, sub)
	}
	res := &DeleteResult{}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, sub := range subs {
			if err := s.repo.SoftDelete(tx, sub.ID); err != nil {
				return err
			}
			if err := s.audit.Write(tx, "DELETE", "work_sub_item", sub.ID, fmt.Sprintf("Sub-pekerjaan \"%s\" dihapus", sub.Name)); err != nil {
				return err
			}
			res.Subs++
		}
		return nil
	})
	if err != nil {
		s.log.Error("gagal menghapus sub-pekerjaan massal", "jumlah", len(subs), "error", err)
		return nil, fmt.Errorf("gagal menghapus sub-pekerjaan")
	}
	return res, nil
}

// ReorderInWorkItem menyimpan urutan baru sub-pekerjaan dalam satu pekerjaan.
func (s *WorkSubItemService) ReorderInWorkItem(workItemID uint, ids []uint) error {
	current, err := s.repo.IDsInWorkItem(s.db, workItemID)
	if err != nil {
		s.log.Error("gagal mengambil urutan sub-pekerjaan", "error", err)
		return fmt.Errorf("gagal menyimpan urutan sub-pekerjaan")
	}
	if !sameIDs(current, ids) {
		return NewValidationError("Urutan sub-pekerjaan tidak valid. Muat ulang halaman lalu coba lagi.")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.ReorderInWorkItem(tx, workItemID, ids); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "work_sub_item", workItemID, "Urutan sub-pekerjaan diubah")
	})
	if err != nil {
		s.log.Error("gagal menyimpan urutan sub-pekerjaan", "pekerjaan", workItemID, "error", err)
		return fmt.Errorf("gagal menyimpan urutan sub-pekerjaan")
	}
	return nil
}
