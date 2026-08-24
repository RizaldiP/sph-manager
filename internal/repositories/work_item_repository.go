package repositories

import (
	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// WorkItemRepository menyediakan akses data tabel work_items.
type WorkItemRepository struct{}

func NewWorkItemRepository() *WorkItemRepository { return &WorkItemRepository{} }

// List mengembalikan pekerjaan terurut sequence, lalu nama.
// categoryID == 0 berarti semua kategori.
func (r *WorkItemRepository) List(db *gorm.DB, categoryID uint, includeInactive bool, search string) ([]models.WorkItem, error) {
	q := db.Model(&models.WorkItem{})
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	var out []models.WorkItem
	err := q.Order("sequence asc, name asc").Find(&out).Error
	return out, err
}

func (r *WorkItemRepository) GetByID(db *gorm.DB, id uint) (*models.WorkItem, error) {
	var w models.WorkItem
	err := db.Preload("SubItems", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sequence asc")
	}).Preload("Category").First(&w, id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// CodeExists memeriksa kode duplikat di antara baris yang belum dihapus.
func (r *WorkItemRepository) CodeExists(db *gorm.DB, code string, excludeID uint) (bool, error) {
	if code == "" {
		return false, nil
	}
	q := db.Model(&models.WorkItem{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *WorkItemRepository) CategoryExists(db *gorm.DB, categoryID uint) (bool, error) {
	var n int64
	err := db.Model(&models.Category{}).Where("id = ?", categoryID).Count(&n).Error
	return n > 0, err
}

func (r *WorkItemRepository) Create(db *gorm.DB, w *models.WorkItem) error {
	return db.Create(w).Error
}

func (r *WorkItemRepository) Update(db *gorm.DB, w *models.WorkItem) error {
	return db.Save(w).Error
}

func (r *WorkItemRepository) SetActive(db *gorm.DB, id uint, active bool) error {
	return db.Model(&models.WorkItem{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *WorkItemRepository) SoftDelete(db *gorm.DB, id uint) error {
	return db.Delete(&models.WorkItem{}, id).Error
}

// CountByCategory menghitung pekerjaan milik sebuah kategori.
func (r *WorkItemRepository) CountByCategory(db *gorm.DB, categoryID uint, onlyLive bool) (int64, error) {
	q := db.Model(&models.WorkItem{}).Where("category_id = ?", categoryID)
	if onlyLive {
		q = q.Where("deleted_at IS NULL")
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

// MaxSequenceInCategory untuk auto urutan saat create.
func (r *WorkItemRepository) MaxSequenceInCategory(db *gorm.DB, categoryID uint) (int, error) {
	var max int
	err := db.Model(&models.WorkItem{}).Where("category_id = ?", categoryID).
		Select("COALESCE(MAX(sequence), 0)").Scan(&max).Error
	return max, err
}

// ReorderInCategory menyimpan urutan baru pekerjaan dalam satu kategori.
func (r *WorkItemRepository) ReorderInCategory(db *gorm.DB, categoryID uint, ids []uint) error {
	for i, id := range ids {
		res := db.Model(&models.WorkItem{}).
			Where("id = ? AND category_id = ?", id, categoryID).
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

func (r *WorkItemRepository) IDsInCategory(db *gorm.DB, categoryID uint) ([]uint, error) {
	var ids []uint
	err := db.Model(&models.WorkItem{}).Where("category_id = ?", categoryID).
		Order("sequence asc").Pluck("id", &ids).Error
	return ids, err
}
