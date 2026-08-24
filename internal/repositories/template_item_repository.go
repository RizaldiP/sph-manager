package repositories

import (
	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// TemplateItemRepository menyediakan akses data tabel template_items.
type TemplateItemRepository struct{}

func NewTemplateItemRepository() *TemplateItemRepository { return &TemplateItemRepository{} }

// ReplaceAll mengganti seluruh isi template dengan daftar baru (urutan = posisi di slice).
// Harus dipanggil di dalam transaksi bersama operasi lain.
func (r *TemplateItemRepository) ReplaceAll(db *gorm.DB, templateID uint, items []models.TemplateItem) error {
	if err := db.Where("template_id = ?", templateID).Delete(&models.TemplateItem{}).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].ID = 0
		items[i].TemplateID = templateID
		items[i].Sequence = i + 1
		if err := db.Create(&items[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// CountByTemplate menghitung jumlah item sebuah template.
func (r *TemplateItemRepository) CountByTemplate(db *gorm.DB, templateID uint) (int64, error) {
	var n int64
	err := db.Model(&models.TemplateItem{}).Where("template_id = ?", templateID).Count(&n).Error
	return n, err
}

// MissingWorkItemIDs mengembalikan ID pekerjaan yang tidak ada / sudah dihapus.
func (r *TemplateItemRepository) MissingWorkItemIDs(db *gorm.DB, ids []uint) ([]uint, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var found []uint
	if err := db.Model(&models.WorkItem{}).Where("id IN ?", ids).Pluck("id", &found).Error; err != nil {
		return nil, err
	}
	exists := make(map[uint]bool, len(found))
	for _, id := range found {
		exists[id] = true
	}
	var missing []uint
	for _, id := range ids {
		if !exists[id] {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

// CountTemplatesUsingWorkItem menghitung template yang masih hidup (belum dihapus)
// dan memuat pekerjaan tertentu — dipakai untuk mencegah hapus pekerjaan terpakai.
func (r *TemplateItemRepository) CountTemplatesUsingWorkItem(db *gorm.DB, workItemID uint) (int64, error) {
	var n int64
	err := db.Model(&models.TemplateItem{}).
		Joins("JOIN templates ON templates.id = template_items.template_id AND templates.deleted_at IS NULL").
		Where("template_items.work_item_id = ?", workItemID).
		Count(&n).Error
	return n, err
}
