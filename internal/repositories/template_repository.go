package repositories

import (
	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// TemplateRepository menyediakan akses data tabel templates.
type TemplateRepository struct{}

func NewTemplateRepository() *TemplateRepository { return &TemplateRepository{} }

// List mengembalikan template terurut sequence, lalu nama.
func (r *TemplateRepository) List(db *gorm.DB, includeInactive bool, search string) ([]models.Template, error) {
	q := db.Model(&models.Template{})
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	var out []models.Template
	err := q.Order("sequence asc, name asc").Find(&out).Error
	return out, err
}

// GetByID mengembalikan template beserta item terurut dan data pekerjaannya
// (kategori + sub-pekerjaan ikut dimuat agar siap dipakai ulang di SPH builder).
func (r *TemplateRepository) GetByID(db *gorm.DB, id uint) (*models.Template, error) {
	var t models.Template
	err := db.
		Preload("Items", func(tx *gorm.DB) *gorm.DB { return tx.Order("sequence asc") }).
		Preload("Items.WorkItem.Category").
		Preload("Items.WorkItem.SubItems", func(tx *gorm.DB) *gorm.DB { return tx.Order("sequence asc") }).
		First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CodeExists memeriksa kode duplikat di antara baris yang belum dihapus.
func (r *TemplateRepository) CodeExists(db *gorm.DB, code string, excludeID uint) (bool, error) {
	if code == "" {
		return false, nil
	}
	q := db.Model(&models.Template{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *TemplateRepository) Create(db *gorm.DB, t *models.Template) error {
	return db.Create(t).Error
}

func (r *TemplateRepository) Update(db *gorm.DB, t *models.Template) error {
	return db.Save(t).Error
}

func (r *TemplateRepository) SetActive(db *gorm.DB, id uint, active bool) error {
	return db.Model(&models.Template{}).Where("id = ?", id).Update("is_active", active).Error
}

// SoftDelete menyembunyikan template; itemnya ikut tak tampil karena menggantung pada induk.
func (r *TemplateRepository) SoftDelete(db *gorm.DB, id uint) error {
	return db.Delete(&models.Template{}, id).Error
}

// MaxSequence untuk auto urutan saat create.
func (r *TemplateRepository) MaxSequence(db *gorm.DB) (int, error) {
	var max int
	err := db.Model(&models.Template{}).Select("COALESCE(MAX(sequence), 0)").Scan(&max).Error
	return max, err
}

// Reorder menyimpan urutan baru seluruh template berdasarkan daftar ID lengkap.
func (r *TemplateRepository) Reorder(db *gorm.DB, ids []uint) error {
	for i, id := range ids {
		res := db.Model(&models.Template{}).Where("id = ?", id).Update("sequence", i+1)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

// IDs mengembalikan seluruh ID template terurut sequence.
func (r *TemplateRepository) IDs(db *gorm.DB) ([]uint, error) {
	var ids []uint
	err := db.Model(&models.Template{}).Order("sequence asc").Pluck("id", &ids).Error
	return ids, err
}
