package repositories

import (
	"strings"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// CustomerRepository: akses data customer & kapal (stateless, db di-inject service).
type CustomerRepository struct{}

func NewCustomerRepository() *CustomerRepository { return &CustomerRepository{} }

func (r *CustomerRepository) List(db *gorm.DB, includeInactive bool, search string) ([]models.Customer, error) {
	q := db.Model(&models.Customer{}).Preload("Vessels", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("name ASC")
	})
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	if s := strings.ToLower(strings.TrimSpace(search)); s != "" {
		q = q.Where("LOWER(name) LIKE ? OR LOWER(COALESCE(code,'')) LIKE ?", "%"+s+"%", "%"+s+"%")
	}
	var out []models.Customer
	err := q.Order("name ASC").Find(&out).Error
	return out, err
}

func (r *CustomerRepository) GetByID(db *gorm.DB, id uint) (*models.Customer, error) {
	var c models.Customer
	if err := db.Preload("Vessels", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("name ASC")
	}).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) Create(db *gorm.DB, c *models.Customer) error {
	return db.Create(c).Error
}

func (r *CustomerRepository) Update(db *gorm.DB, c *models.Customer) error {
	return db.Save(c).Error
}

func (r *CustomerRepository) SetActive(db *gorm.DB, id uint, active bool) error {
	return db.Model(&models.Customer{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *CustomerRepository) SoftDelete(db *gorm.DB, id uint) error {
	return db.Delete(&models.Customer{}, id).Error
}

// ===== kapal =====

func (r *CustomerRepository) GetVesselByID(db *gorm.DB, id uint) (*models.Vessel, error) {
	var v models.Vessel
	if err := db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *CustomerRepository) CreateVessel(db *gorm.DB, v *models.Vessel) error {
	return db.Create(v).Error
}

func (r *CustomerRepository) UpdateVessel(db *gorm.DB, v *models.Vessel) error {
	return db.Save(v).Error
}

func (r *CustomerRepository) SetVesselActive(db *gorm.DB, id uint, active bool) error {
	return db.Model(&models.Vessel{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *CustomerRepository) SoftDeleteVessel(db *gorm.DB, id uint) error {
	return db.Delete(&models.Vessel{}, id).Error
}

// CodeExists memeriksa kode customer yang sama pada data hidup lain.
func (r *CustomerRepository) CodeExists(db *gorm.DB, code string, excludeID uint) (bool, error) {
	if strings.TrimSpace(code) == "" {
		return false, nil
	}
	q := db.Model(&models.Customer{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}

// VesselCodeExists memeriksa kode kapal yang sama pada data hidup lain.
func (r *CustomerRepository) VesselCodeExists(db *gorm.DB, code string, excludeID uint) (bool, error) {
	if strings.TrimSpace(code) == "" {
		return false, nil
	}
	q := db.Model(&models.Vessel{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}
