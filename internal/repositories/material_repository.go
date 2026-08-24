package repositories

import (
	"strings"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// MaterialRepository: akses data master material (stateless, db di-inject service).
type MaterialRepository struct{}

func NewMaterialRepository() *MaterialRepository { return &MaterialRepository{} }

func (r *MaterialRepository) List(db *gorm.DB, includeInactive bool, search string) ([]models.Material, error) {
	q := db.Model(&models.Material{})
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	if s := strings.ToLower(strings.TrimSpace(search)); s != "" {
		q = q.Where(
			"LOWER(name) LIKE ? OR LOWER(COALESCE(code,'')) LIKE ? OR LOWER(COALESCE(supplier,'')) LIKE ?",
			"%"+s+"%", "%"+s+"%", "%"+s+"%",
		)
	}
	var out []models.Material
	err := q.Order("name ASC").Find(&out).Error
	return out, err
}

func (r *MaterialRepository) GetByID(db *gorm.DB, id uint) (*models.Material, error) {
	var m models.Material
	if err := db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MaterialRepository) Create(db *gorm.DB, m *models.Material) error {
	return db.Create(m).Error
}

func (r *MaterialRepository) Update(db *gorm.DB, m *models.Material) error {
	return db.Save(m).Error
}

func (r *MaterialRepository) SetActive(db *gorm.DB, id uint, active bool) error {
	return db.Model(&models.Material{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *MaterialRepository) SoftDelete(db *gorm.DB, id uint) error {
	return db.Delete(&models.Material{}, id).Error
}

// CodeExists memeriksa kode material yang sama pada data hidup lain.
func (r *MaterialRepository) CodeExists(db *gorm.DB, code string, excludeID uint) (bool, error) {
	if strings.TrimSpace(code) == "" {
		return false, nil
	}
	q := db.Model(&models.Material{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}
