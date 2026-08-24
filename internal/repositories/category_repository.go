package repositories

import (
	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// CategoryRepository menyediakan akses data tabel categories.
type CategoryRepository struct{}

func NewCategoryRepository() *CategoryRepository { return &CategoryRepository{} }

// List mengembalikan kategori terurut sequence, lalu nama.
func (r *CategoryRepository) List(db *gorm.DB, includeInactive bool, search string) ([]models.Category, error) {
	q := db.Model(&models.Category{})
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	var out []models.Category
	err := q.Order("sequence asc, name asc").Find(&out).Error
	return out, err
}

func (r *CategoryRepository) GetByID(db *gorm.DB, id uint) (*models.Category, error) {
	var c models.Category
	err := db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CodeExists memeriksa kode duplikat di antara baris yang belum dihapus (soft-delete aware).
func (r *CategoryRepository) CodeExists(db *gorm.DB, code string, excludeID uint) (bool, error) {
	if code == "" {
		return false, nil
	}
	q := db.Model(&models.Category{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *CategoryRepository) Create(db *gorm.DB, c *models.Category) error {
	return db.Create(c).Error
}

func (r *CategoryRepository) Update(db *gorm.DB, c *models.Category) error {
	return db.Save(c).Error
}

func (r *CategoryRepository) SetActive(db *gorm.DB, id uint, active bool) error {
	return db.Model(&models.Category{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *CategoryRepository) SoftDelete(db *gorm.DB, id uint) error {
	return db.Delete(&models.Category{}, id).Error
}

// MaxSequence untuk auto urutan saat create.
func (r *CategoryRepository) MaxSequence(db *gorm.DB) (int, error) {
	var max int
	err := db.Model(&models.Category{}).Select("COALESCE(MAX(sequence), 0)").Scan(&max).Error
	return max, err
}

// Reorder menyimpan urutan baru seluruh kategori.
func (r *CategoryRepository) Reorder(db *gorm.DB, ids []uint) error {
	for i, id := range ids {
		res := db.Model(&models.Category{}).Where("id = ?", id).Update("sequence", i+1)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (r *CategoryRepository) IDs(db *gorm.DB) ([]uint, error) {
	var ids []uint
	err := db.Model(&models.Category{}).Order("sequence asc").Pluck("id", &ids).Error
	return ids, err
}
