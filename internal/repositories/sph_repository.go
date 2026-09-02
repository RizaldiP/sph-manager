package repositories

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// SphRepository menyediakan akses data dokumen SPH beserta item & sub-itemnya.
type SphRepository struct{}

func NewSphRepository() *SphRepository { return &SphRepository{} }

// List mengembalikan daftar dokumen terurut terbaru.
// statuses kosong berarti semua status. search mencari di nomor/project/subject.
func (r *SphRepository) List(db *gorm.DB, statuses []string, search string, limit int) ([]models.SphDocument, error) {
	q := db.Model(&models.SphDocument{})
	if len(statuses) > 0 {
		q = q.Where("status IN ?", statuses)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("document_number LIKE ? OR project_name LIKE ? OR subject LIKE ?", like, like, like)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var out []models.SphDocument
	err := q.Preload("Customer").Preload("Vessel").
		Order("date desc, id desc").Find(&out).Error
	return out, err
}

// GetByID mengembalikan dokumen lengkap: customer/kapal, item terurut beserta sub-itemnya.
func (r *SphRepository) GetByID(db *gorm.DB, id uint) (*models.SphDocument, error) {
	var d models.SphDocument
	err := db.
		Preload("Customer").
		Preload("Vessel").
		Preload("Items", func(tx *gorm.DB) *gorm.DB { return tx.Order("sequence asc") }).
		Preload("Items.SubItems", func(tx *gorm.DB) *gorm.DB { return tx.Order("sequence asc") }).
		Preload("Revisions", func(tx *gorm.DB) *gorm.DB { return tx.Order("revision_number asc") }).
		First(&d, id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// NumberExists memeriksa pemakaian nomor pada baris hidup (untuk validasi BR-07).
func (r *SphRepository) NumberExists(db *gorm.DB, number string, excludeID uint) (bool, error) {
	if number == "" {
		return false, nil
	}
	q := db.Model(&models.SphDocument{}).Where("document_number = ?", number)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// MaxSequenceInNumber mengembalikan nomor urut terbesar dari dokumen tersimpan
// yang berawalan prefix dan berakhiran suffix (bagian format di sekitar {SEQ},
// salah satunya boleh kosong). Berfungsi untuk penempatan {SEQ} di depan,
// tengah, maupun akhir. Dipakai generator penomoran transaksional (BR-07).
func (r *SphRepository) MaxSequenceInNumber(db *gorm.DB, prefix, suffix string) (int, error) {
	q := db.Model(&models.SphDocument{})
	if prefix != "" {
		q = q.Where("document_number LIKE ?", prefix+"%")
	}
	if suffix != "" {
		q = q.Where("document_number LIKE ?", "%"+suffix)
	}
	var numbers []string
	if err := q.Pluck("document_number", &numbers).Error; err != nil {
		return 0, err
	}
	max := 0
	for _, n := range numbers {
		if !strings.HasPrefix(n, prefix) || !strings.HasSuffix(n, suffix) {
			continue
		}
		middle := n[len(prefix):]
		if suffix != "" {
			middle = n[len(prefix) : len(n)-len(suffix)]
		}
		v := 0
		valid := true
		if len(middle) == 0 {
			continue
		}
		for _, ch := range middle {
			if ch < '0' || ch > '9' {
				valid = false
				break
			}
			v = v*10 + int(ch-'0')
		}
		if valid && v > max {
			max = v
		}
	}
	return max, nil
}

// MaxRevision mengembalikan nomor revisi tertinggi untuk sebuah nomor dokumen.
func (r *SphRepository) MaxRevision(db *gorm.DB, number string) (int, error) {
	var max int
	err := db.Model(&models.SphDocument{}).
		Where("document_number = ?", number).
		Select("COALESCE(MAX(revision), -1)").Scan(&max).Error
	return max, err
}

func (r *SphRepository) Create(db *gorm.DB, d *models.SphDocument) error {
	return db.Create(d).Error
}

func (r *SphRepository) Update(db *gorm.DB, d *models.SphDocument) error {
	return db.Save(d).Error
}

func (r *SphRepository) SetStatus(db *gorm.DB, id uint, status string) error {
	return db.Model(&models.SphDocument{}).Where("id = ?", id).Update("status", status).Error
}

// Finalize menyimpan status final beserta waktu finalisasi.
func (r *SphRepository) Finalize(db *gorm.DB, id uint, status string, finalizedAt time.Time) error {
	return db.Model(&models.SphDocument{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "finalized_at": finalizedAt}).Error
}

func (r *SphRepository) SoftDelete(db *gorm.DB, id uint) error {
	return db.Delete(&models.SphDocument{}, id).Error
}

// ReplaceItems mengganti seluruh item & sub item dokumen dengan snapshot baru.
// Harus dipanggil dalam transaksi bersama operasi lain (BR-16).
func (r *SphRepository) ReplaceItems(db *gorm.DB, documentID uint, items []models.SphItem) error {
	if err := db.Where("sph_document_id = ?", documentID).Delete(&models.SphItem{}).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].ID = 0
		items[i].SphDocumentID = documentID
		items[i].Sequence = i + 1
		subs := items[i].SubItems
		items[i].SubItems = nil
		if err := db.Create(&items[i]).Error; err != nil {
			return err
		}
		for j := range subs {
			subs[j].ID = 0
			subs[j].SphItemID = items[i].ID
			subs[j].Sequence = j + 1
			if err := db.Create(&subs[j]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// CountItems menghitung jumlah main point sebuah dokumen.
func (r *SphRepository) CountItems(db *gorm.DB, documentID uint) (int64, error) {
	var n int64
	err := db.Model(&models.SphItem{}).Where("sph_document_id = ?", documentID).Count(&n).Error
	return n, err
}

// CountByCustomer menghitung dokumen SPH yang memakai customer (guard hapus master).
func (r *SphRepository) CountByCustomer(db *gorm.DB, customerID uint) (int64, error) {
	var n int64
	err := db.Model(&models.SphDocument{}).Where("customer_id = ?", customerID).Count(&n).Error
	return n, err
}

// CountByVessel menghitung dokumen SPH yang memakai kapal (guard hapus master).
func (r *SphRepository) CountByVessel(db *gorm.DB, vesselID uint) (int64, error) {
	var n int64
	err := db.Model(&models.SphDocument{}).Where("vessel_id = ?", vesselID).Count(&n).Error
	return n, err
}

// ===== statistik dashboard =====

type SphStatsRow struct {
	Status string
	Count  int64
	Sum    int64
}

// StatsByStatus menjumlahkan jumlah & nilai dokumen per status (dokumen hidup).
func (r *SphRepository) StatsByStatus(db *gorm.DB) ([]SphStatsRow, error) {
	var rows []SphStatsRow
	err := db.Model(&models.SphDocument{}).
		Select("status, COUNT(*) as count, COALESCE(SUM(grand_total), 0) as sum").
		Group("status").Scan(&rows).Error
	return rows, err
}

// SumGrandTotalBetween menjumlahkan nilai dokumen pada rentang tanggal (inklusif).
func (r *SphRepository) SumGrandTotalBetween(db *gorm.DB, from, to string) (int64, error) {
	var sum int64
	err := db.Model(&models.SphDocument{}).
		Where("date >= ? AND date <= ?", from, to).
		Select("COALESCE(SUM(grand_total), 0)").Scan(&sum).Error
	return sum, err
}

// Setting mengambil satu nilai konfigurasi; kosong bila belum ada.
func SettingValue(db *gorm.DB, key string) (string, error) {
	var s models.Setting
	err := db.Where("`key` = ?", key).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

// SetSetting menyimpan nilai konfigurasi (upsert).
func SetSetting(db *gorm.DB, key, value string) error {
	return db.Save(&models.Setting{Key: key, Value: value}).Error
}
