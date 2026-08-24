package repositories

import (
	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// WorkSubItemRepository menyediakan akses data tabel work_sub_items.
type WorkSubItemRepository struct{}

func NewWorkSubItemRepository() *WorkSubItemRepository { return &WorkSubItemRepository{} }

func (r *WorkSubItemRepository) GetByID(db *gorm.DB, id uint) (*models.WorkSubItem, error) {
	var s models.WorkSubItem
	err := db.First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *WorkSubItemRepository) WorkItemExists(db *gorm.DB, workItemID uint) (bool, error) {
	var n int64
	err := db.Model(&models.WorkItem{}).Where("id = ?", workItemID).Count(&n).Error
	return n > 0, err
}

func (r *WorkSubItemRepository) Create(db *gorm.DB, s *models.WorkSubItem) error {
	return db.Create(s).Error
}

func (r *WorkSubItemRepository) Update(db *gorm.DB, s *models.WorkSubItem) error {
	return db.Save(s).Error
}

func (r *WorkSubItemRepository) SetActive(db *gorm.DB, id uint, active bool) error {
	return db.Model(&models.WorkSubItem{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *WorkSubItemRepository) SoftDelete(db *gorm.DB, id uint) error {
	return db.Delete(&models.WorkSubItem{}, id).Error
}

// SoftDeleteByWorkItem menghapus (soft) seluruh sub-pekerjaan milik satu pekerjaan.
func (r *WorkSubItemRepository) SoftDeleteByWorkItem(db *gorm.DB, workItemID uint) error {
	return db.Where("work_item_id = ?", workItemID).Delete(&models.WorkSubItem{}).Error
}

// CountByWorkItem menghitung sub-pekerjaan yang belum dihapus milik sebuah pekerjaan.
func (r *WorkSubItemRepository) CountByWorkItem(db *gorm.DB, workItemID uint) (int64, error) {
	var n int64
	err := db.Model(&models.WorkSubItem{}).Where("work_item_id = ?", workItemID).Count(&n).Error
	return n, err
}

// CountActiveByWorkItem menghitung sub-pekerjaan aktif milik sebuah pekerjaan.
func (r *WorkSubItemRepository) CountActiveByWorkItem(db *gorm.DB, workItemID uint) (int64, error) {
	var n int64
	err := db.Model(&models.WorkSubItem{}).
		Where("work_item_id = ? AND is_active = ?", workItemID, true).
		Count(&n).Error
	return n, err
}

// MaxSequenceInWorkItem untuk auto urutan saat create.
func (r *WorkSubItemRepository) MaxSequenceInWorkItem(db *gorm.DB, workItemID uint) (int, error) {
	var max int
	err := db.Model(&models.WorkSubItem{}).Where("work_item_id = ?", workItemID).
		Select("COALESCE(MAX(sequence), 0)").Scan(&max).Error
	return max, err
}

// ReorderInWorkItem menyimpan urutan baru sub-pekerjaan dalam satu pekerjaan.
func (r *WorkSubItemRepository) ReorderInWorkItem(db *gorm.DB, workItemID uint, ids []uint) error {
	for i, id := range ids {
		res := db.Model(&models.WorkSubItem{}).
			Where("id = ? AND work_item_id = ?", id, workItemID).
			Update("sequence", i+1)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (r *WorkSubItemRepository) IDsInWorkItem(db *gorm.DB, workItemID uint) ([]uint, error) {
	var ids []uint
	err := db.Model(&models.WorkSubItem{}).Where("work_item_id = ?", workItemID).
		Order("sequence asc").Pluck("id", &ids).Error
	return ids, err
}
