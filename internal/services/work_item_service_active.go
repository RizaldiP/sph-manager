package services

import (
	"fmt"

	"gorm.io/gorm"
)

// SetActive mengaktifkan / menonaktifkan pekerjaan.
func (s *WorkItemService) SetActive(id uint, active bool) error {
	if _, err := s.repo.GetByID(s.db, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status pekerjaan")
	}
	desc := "dinonaktifkan"
	if active {
		desc = "diaktifkan"
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SetActive(tx, id, active); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "work_item", id, "Pekerjaan "+desc)
	})
	if err != nil {
		s.log.Error("gagal mengubah status pekerjaan", "id", id, "active", active, "error", err)
		return fmt.Errorf("gagal mengubah status pekerjaan")
	}
	return nil
}

// Delete melakukan soft delete; ditolak bila masih memiliki sub-pekerjaan aktif.
func (s *WorkItemService) Delete(id uint) error {
	w, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus pekerjaan")
	}
	n, err := s.subRepo.CountActiveByWorkItem(s.db, id)
	if err != nil {
		s.log.Error("gagal menghitung sub-pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus pekerjaan")
	}
	if n > 0 {
		return NewConflictError("Pekerjaan tidak dapat dihapus karena masih memiliki %d sub-pekerjaan. Hapus sub-pekerjaannya terlebih dahulu.", n)
	}
	nT, err := s.tplItems.CountTemplatesUsingWorkItem(s.db, id)
	if err != nil {
		s.log.Error("gagal menghitung pemakaian pekerjaan di template", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus pekerjaan")
	}
	if nT > 0 {
		return NewConflictError("Pekerjaan tidak dapat dihapus karena masih dipakai oleh %d template. Hapus pekerjaan tersebut dari templatenya terlebih dahulu.", nT)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDelete(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "work_item", id, fmt.Sprintf("Pekerjaan \"%s\" dihapus", w.Name))
	})
	if err != nil {
		s.log.Error("gagal menghapus pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus pekerjaan")
	}
	return nil
}

// ReorderInCategory menyimpan urutan baru pekerjaan dalam satu kategori.
func (s *WorkItemService) ReorderInCategory(categoryID uint, ids []uint) error {
	current, err := s.repo.IDsInCategory(s.db, categoryID)
	if err != nil {
		s.log.Error("gagal mengambil urutan pekerjaan", "error", err)
		return fmt.Errorf("gagal menyimpan urutan pekerjaan")
	}
	if !sameIDs(current, ids) {
		return NewValidationError("Urutan pekerjaan tidak valid. Muat ulang halaman lalu coba lagi.")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.ReorderInCategory(tx, categoryID, ids); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "work_item", categoryID, "Urutan pekerjaan diubah")
	})
	if err != nil {
		s.log.Error("gagal menyimpan urutan pekerjaan", "kategori", categoryID, "error", err)
		return fmt.Errorf("gagal menyimpan urutan pekerjaan")
	}
	return nil
}
