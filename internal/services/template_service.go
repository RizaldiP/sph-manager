package services

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// TemplateView adalah template plus jumlah pekerjaan untuk tampilan daftar.
type TemplateView struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	Sequence    int    `json:"sequence"`
	IsActive    bool   `json:"isActive"`
	ItemCount   int64  `json:"itemCount"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// TemplateItemInput adalah satu baris pekerjaan pada editor template.
// Struct terpisah (tanpa field waktu) agar aman dikirim dari frontend ke binding Wails.
type TemplateItemInput struct {
	WorkItemID uint   `json:"workItemId"`
	Notes      string `json:"notes"`
}

// TemplateService: CRUD template pekerjaan (FR-M4, FR-U6, BR-05, BR-13, BR-15).
type TemplateService struct {
	db       *gorm.DB
	log      *slog.Logger
	repo     *repositories.TemplateRepository
	itemRepo *repositories.TemplateItemRepository
	audit    *AuditWriter
}

func NewTemplateService(db *gorm.DB, log *slog.Logger) *TemplateService {
	return &TemplateService{
		db:       db,
		log:      log,
		repo:     repositories.NewTemplateRepository(),
		itemRepo: repositories.NewTemplateItemRepository(),
		audit:    NewAuditWriter(),
	}
}

func toTemplateViews(db *gorm.DB, templates []models.Template) ([]TemplateView, error) {
	itemRepo := repositories.NewTemplateItemRepository()
	out := make([]TemplateView, 0, len(templates))
	for _, t := range templates {
		n, err := itemRepo.CountByTemplate(db, t.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, TemplateView{
			ID:          t.ID,
			Code:        t.Code,
			Name:        t.Name,
			Description: t.Description,
			Notes:       t.Notes,
			Sequence:    t.Sequence,
			IsActive:    t.IsActive,
			ItemCount:   n,
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:   t.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	return out, nil
}

func toTemplateView(db *gorm.DB, t *models.Template) (*TemplateView, error) {
	views, err := toTemplateViews(db, []models.Template{*t})
	if err != nil {
		return nil, err
	}
	return &views[0], nil
}

func (s *TemplateService) List(includeInactive bool, search string) ([]TemplateView, error) {
	templates, err := s.repo.List(s.db, includeInactive, trim(search))
	if err != nil {
		s.log.Error("gagal mengambil daftar template", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar template")
	}
	views, err := toTemplateViews(s.db, templates)
	if err != nil {
		s.log.Error("gagal menghitung isi template", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar template")
	}
	return views, nil
}

// Get mengembalikan template lengkap dengan isian pekerjaannya (terurut).
func (s *TemplateService) Get(id uint) (*models.Template, error) {
	t, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil template", "id", id, "error", err)
		return nil, fmt.Errorf("gagal mengambil template")
	}
	return t, nil
}

func (s *TemplateService) validate(t *models.Template, excludeID uint) error {
	t.Code = trim(t.Code)
	t.Name = trim(t.Name)
	t.Description = trim(t.Description)
	t.Notes = trim(t.Notes)
	if t.Name == "" {
		return NewValidationError("Nama template wajib diisi.")
	}
	if len(t.Name) > 200 {
		return NewValidationError("Nama template maksimal 200 karakter.")
	}
	if len(t.Code) > 50 {
		return NewValidationError("Kode template maksimal 50 karakter.")
	}
	if len(t.Description) > 1000 {
		return NewValidationError("Deskripsi maksimal 1000 karakter.")
	}
	if len(t.Notes) > 1000 {
		return NewValidationError("Catatan maksimal 1000 karakter.")
	}
	dup, err := s.repo.CodeExists(s.db, t.Code, excludeID)
	if err != nil {
		return err
	}
	if dup {
		return NewConflictError("Kode \"%s\" sudah digunakan template lain.", t.Code)
	}
	return nil
}

func (s *TemplateService) Create(in *models.Template) (*TemplateView, error) {
	if err := s.validate(in, 0); err != nil {
		return nil, err
	}
	seq, err := s.repo.MaxSequence(s.db)
	if err != nil {
		s.log.Error("gagal menghitung urutan terakhir", "error", err)
		return nil, fmt.Errorf("gagal menyimpan template")
	}
	in.ID = 0
	in.Sequence = seq + 1
	in.IsActive = true

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Kode otomatis bila tidak diisi manual (FR-U3: non-teknis).
		if in.Code == "" {
			code, err := generateCode(tx, "templates", "TPL-")
			if err != nil {
				return err
			}
			in.Code = code
		}
		if err := s.repo.Create(tx, in); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "template", in.ID, fmt.Sprintf("Template \"%s\" dibuat", in.Name))
	})
	if err != nil {
		s.log.Error("gagal membuat template", "nama", in.Name, "error", err)
		return nil, fmt.Errorf("gagal menyimpan template")
	}
	return toTemplateView(s.db, in)
}

// Update hanya mengubah data header; isi pekerjaan diatur lewat SetItems.
func (s *TemplateService) Update(id uint, in *models.Template) (*TemplateView, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil template", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui template")
	}
	in.ID = existing.ID
	if err := s.validate(in, id); err != nil {
		return nil, err
	}
	// Kode dikelola sistem; kosong berarti pertahankan kode yang sudah ada.
	if in.Code != "" {
		existing.Code = in.Code
	}
	existing.Name = in.Name
	existing.Description = in.Description
	existing.Notes = in.Notes

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(tx, existing); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "template", existing.ID, fmt.Sprintf("Template \"%s\" diperbarui", existing.Name))
	})
	if err != nil {
		s.log.Error("gagal memperbarui template", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui template")
	}
	return toTemplateView(s.db, existing)
}

// SetItems mengganti seluruh isi template; urutan = urutan input.
// Daftar kosong berarti mengosongkan template.
func (s *TemplateService) SetItems(id uint, inputs []TemplateItemInput) (*models.Template, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil template", "id", id, "error", err)
		return nil, fmt.Errorf("gagal menyimpan isi template")
	}

	items := make([]models.TemplateItem, 0, len(inputs))
	seen := make(map[uint]bool, len(inputs))
	for i := range inputs {
		inputs[i].Notes = trim(inputs[i].Notes)
		if inputs[i].WorkItemID == 0 {
			return nil, NewValidationError("Pekerjaan pada baris %d belum dipilih.", i+1)
		}
		if len(inputs[i].Notes) > 500 {
			return nil, NewValidationError("Catatan pada baris %d maksimal 500 karakter.", i+1)
		}
		if seen[inputs[i].WorkItemID] {
			return nil, NewValidationError("Ada pekerjaan yang sama dimasukkan lebih dari sekali ke template.")
		}
		seen[inputs[i].WorkItemID] = true
		items = append(items, models.TemplateItem{
			WorkItemID: inputs[i].WorkItemID,
			Notes:      inputs[i].Notes,
		})
	}

	missing, err := s.itemRepo.MissingWorkItemIDs(s.db, seenToSlice(seen))
	if err != nil {
		s.log.Error("gagal memeriksa pekerjaan", "error", err)
		return nil, fmt.Errorf("gagal menyimpan isi template")
	}
	if len(missing) > 0 {
		return nil, NewValidationError("%d pekerjaan tidak ditemukan (mungkin sudah dihapus). Muat ulang halaman lalu coba lagi.", len(missing))
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.itemRepo.ReplaceAll(tx, id, items); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "template", id,
			fmt.Sprintf("Isi template \"%s\" diperbarui (%d pekerjaan)", existing.Name, len(items)))
	})
	if err != nil {
		s.log.Error("gagal menyimpan isi template", "id", id, "error", err)
		return nil, fmt.Errorf("gagal menyimpan isi template")
	}
	return s.repo.GetByID(s.db, id)
}

// Duplicate menyalin template beserta seluruh isinya dengan urutan yang sama.
// Salinan selalu mendapat kode otomatis baru agar tidak bentrok dengan asalnya.
func (s *TemplateService) Duplicate(id uint) (*TemplateView, error) {
	src, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil template", "id", id, "error", err)
		return nil, fmt.Errorf("gagal menduplikasi template")
	}

	seq, err := s.repo.MaxSequence(s.db)
	if err != nil {
		s.log.Error("gagal menghitung urutan terakhir", "error", err)
		return nil, fmt.Errorf("gagal menduplikasi template")
	}

	name := trim(src.Name) + " - Salinan"
	if len(name) > 200 {
		name = name[:200]
	}
	clone := &models.Template{
		Code:        "",
		Name:        name,
		Description: src.Description,
		Notes:       src.Notes,
		IsActive:    true,
		Sequence:    seq + 1,
	}
	items := make([]models.TemplateItem, 0, len(src.Items))
	for _, it := range src.Items {
		items = append(items, models.TemplateItem{WorkItemID: it.WorkItemID, Notes: it.Notes})
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Salinan selalu dapat kode baru agar tidak bentrok dengan asalnya.
		code, err := generateCode(tx, "templates", "TPL-")
		if err != nil {
			return err
		}
		clone.Code = code
		if err := s.repo.Create(tx, clone); err != nil {
			return err
		}
		if err := s.itemRepo.ReplaceAll(tx, clone.ID, items); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "template", clone.ID,
			fmt.Sprintf("Diduplikat dari template \"%s\"", src.Name))
	})
	if err != nil {
		s.log.Error("gagal menduplikasi template", "id", id, "error", err)
		return nil, fmt.Errorf("gagal menduplikasi template")
	}
	return toTemplateView(s.db, clone)
}

func (s *TemplateService) SetActive(id uint, active bool) error {
	if _, err := s.repo.GetByID(s.db, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil template", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status template")
	}
	desc := "dinonaktifkan"
	if active {
		desc = "diaktifkan"
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SetActive(tx, id, active); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "template", id, "Template "+desc)
	})
	if err != nil {
		s.log.Error("gagal mengubah status template", "id", id, "active", active, "error", err)
		return fmt.Errorf("gagal mengubah status template")
	}
	return nil
}

// Delete melakukan soft delete; isi template ikut tak tampil bersama induknya.
func (s *TemplateService) Delete(id uint) error {
	t, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil template", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus template")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDelete(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "template", id, fmt.Sprintf("Template \"%s\" dihapus", t.Name))
	})
	if err != nil {
		s.log.Error("gagal menghapus template", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus template")
	}
	return nil
}

// Reorder menyimpan urutan baru seluruh template (harus berisi semua ID).
func (s *TemplateService) Reorder(ids []uint) error {
	current, err := s.repo.IDs(s.db)
	if err != nil {
		s.log.Error("gagal mengambil urutan template", "error", err)
		return fmt.Errorf("gagal menyimpan urutan template")
	}
	if !sameIDs(current, ids) {
		return NewValidationError("Urutan template tidak valid. Muat ulang halaman lalu coba lagi.")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Reorder(tx, ids); err != nil {
			return err
		}
		return s.audit.Write(tx, "UPDATE", "template", 0, "Urutan template diubah")
	})
	if err != nil {
		s.log.Error("gagal menyimpan urutan template", "error", err)
		return fmt.Errorf("gagal menyimpan urutan template")
	}
	return nil
}

// seenToSlice mengonversi set ID menjadi slice.
func seenToSlice(set map[uint]bool) []uint {
	out := make([]uint, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}
