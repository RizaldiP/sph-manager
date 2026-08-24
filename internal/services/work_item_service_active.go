package services

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
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

// Delete melakukan soft delete pekerjaan beserta seluruh sub-pekerjaannya
// dalam satu transaksi; ditolak bila masih dipakai template hidup.
func (s *WorkItemService) Delete(id uint) error {
	w, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus pekerjaan")
	}
	if err := s.ensureNotUsedByTemplates([]uint{id}); err != nil {
		return err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		n, err := s.cascadeDeleteTx(tx, w)
		if err != nil {
			return err
		}
		desc := fmt.Sprintf("Pekerjaan \"%s\" dihapus", w.Name)
		if n > 0 {
			desc = fmt.Sprintf("Pekerjaan \"%s\" dihapus beserta %d sub-pekerjaannya", w.Name, n)
		}
		return s.audit.Write(tx, "DELETE", "work_item", id, desc)
	})
	if err != nil {
		s.log.Error("gagal menghapus pekerjaan", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus pekerjaan")
	}
	return nil
}

// DeleteMany menghapus banyak pekerjaan sekaligus (kaskade sub-pekerjaannya)
// dalam satu transaksi; seluruh batch dibatalkan bila ada yang bermasalah.
func (s *WorkItemService) DeleteMany(ids []uint) (*DeleteResult, error) {
	uniq, err := uniqueIDs(ids)
	if err != nil {
		return nil, NewValidationError("Pilih minimal satu pekerjaan untuk dihapus.")
	}
	items := make([]*models.WorkItem, 0, len(uniq))
	for _, id := range uniq {
		w, err := s.repo.GetByID(s.db, id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, NewConflictError("Sebagian pekerjaan sudah tidak tersedia (ID %d). Muat ulang halaman lalu coba lagi.", id)
			}
			s.log.Error("gagal mengambil pekerjaan", "id", id, "error", err)
			return nil, fmt.Errorf("gagal menghapus pekerjaan")
		}
		items = append(items, w)
	}
	if err := s.ensureNotUsedByTemplates(uniq); err != nil {
		return nil, err
	}
	res := &DeleteResult{}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, w := range items {
			n, err := s.cascadeDeleteTx(tx, w)
			if err != nil {
				return err
			}
			desc := fmt.Sprintf("Pekerjaan \"%s\" dihapus", w.Name)
			if n > 0 {
				desc = fmt.Sprintf("Pekerjaan \"%s\" dihapus beserta %d sub-pekerjaannya", w.Name, n)
			}
			if err := s.audit.Write(tx, "DELETE", "work_item", w.ID, desc); err != nil {
				return err
			}
			res.Items++
			res.Subs += n
		}
		return nil
	})
	if err != nil {
		s.log.Error("gagal menghapus pekerjaan massal", "jumlah", len(items), "error", err)
		return nil, fmt.Errorf("gagal menghapus pekerjaan")
	}
	return res, nil
}

// cascadeDeleteTx soft delete pekerjaan dan seluruh sub-pekerjaannya;
// mengembalikan jumlah sub-pekerjaan yang ikut terhapus.
func (s *WorkItemService) cascadeDeleteTx(tx *gorm.DB, w *models.WorkItem) (int64, error) {
	n, err := s.subRepo.CountByWorkItem(tx, w.ID)
	if err != nil {
		return 0, err
	}
	if n > 0 {
		if err := s.subRepo.SoftDeleteByWorkItem(tx, w.ID); err != nil {
			return 0, err
		}
	}
	if err := s.repo.SoftDelete(tx, w.ID); err != nil {
		return 0, err
	}
	return n, nil
}

// ensureNotUsedByTemplates menolak hapus bila salah satu pekerjaan masih
// dipakai template hidup.
func (s *WorkItemService) ensureNotUsedByTemplates(ids []uint) error {
	for _, id := range ids {
		nT, err := s.tplItems.CountTemplatesUsingWorkItem(s.db, id)
		if err != nil {
			s.log.Error("gagal menghitung pemakaian pekerjaan di template", "id", id, "error", err)
			return fmt.Errorf("gagal menghapus pekerjaan")
		}
		if nT > 0 {
			w, err := s.repo.GetByID(s.db, id)
			name := fmt.Sprintf("ID %d", id)
			if err == nil && w.Name != "" {
				name = "\"" + w.Name + "\""
			}
			return NewConflictError("Pekerjaan %s tidak dapat dihapus karena masih dipakai oleh %d template. Hapus pekerjaan tersebut dari templatenya terlebih dahulu.", name, nT)
		}
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
