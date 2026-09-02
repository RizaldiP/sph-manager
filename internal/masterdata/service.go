package masterdata

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/collaboration"
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// Strategi konflik saat memasang Master Data.
const (
	StrategyPrompt     = "PROMPT"      // pengguna memutuskan per item (default; konflik dilewati bila tanpa keputusan)
	StrategyUseLocal   = "USE_LOCAL"   // pertahankan data lokal saat konflik
	StrategyUseIncoming = "USE_INCOMING" // timpa data lokal dengan data masuk
	StrategySkip       = "SKIP"        // lewati item yang konflik
)

const (
	DiffNew      = "NEW"
	DiffUpdated  = "UPDATED"
	DiffUnchanged = "UNCHANGED"
	DiffConflict = "CONFLICT"
)

// Service melakukan build/compare/install Master Data antar PC dalam Room,
// plus penyimpanan inbox (diterima) & sent (dikirim). Semua operasi tulis
// transaksional dengan rollback penuh bila gagal (FR-A4 / BR-16).
type Service struct {
	db   *gorm.DB
	log  *slog.Logger
	cats *repositories.CategoryRepository
	items *repositories.WorkItemRepository
	subs  *repositories.WorkSubItemRepository
	mats  *repositories.MaterialRepository
}

func NewService(db *gorm.DB, log *slog.Logger) *Service {
	return &Service{
		db: db, log: log,
		cats:  repositories.NewCategoryRepository(),
		items: repositories.NewWorkItemRepository(),
		subs:  repositories.NewWorkSubItemRepository(),
		mats:  repositories.NewMaterialRepository(),
	}
}

// FilterSelection berisi kode-kode item yang dipilih untuk dikirim.
// Jika SendAll true, semua data akan disertakan tanpa filter.
type FilterSelection struct {
	CategoryCodes  []string `json:"categoryCodes"`
	WorkItemCodes  []string `json:"workItemCodes"`
	SubItemKeys    []string `json:"subItemKeys"`   // format: "workItemCode:subCode" atau "workItemCode::subName"
	MaterialCodes  []string `json:"materialCodes"`
	SendAll        bool     `json:"sendAll"`
}

// MasterDataList adalah daftar semua Master Data untuk UI selection.
type MasterDataList struct {
	Categories   []collaboration.PackageCategory    `json:"categories"`
	WorkItems    []collaboration.PackageWorkItem    `json:"workItems"`
	WorkSubItems []collaboration.PackageWorkSubItem `json:"workSubItems"`
	Materials    []collaboration.PackageMaterial    `json:"materials"`
}

// ===== Build =====

// ListAllMasterData mengembalikan seluruh Master Data aktif untuk UI selection.
func (s *Service) ListAllMasterData() (*MasterDataList, error) {
	var cats []models.Category
	if err := s.db.Where("deleted_at IS NULL").Order("sequence ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	var items []models.WorkItem
	if err := s.db.Where("deleted_at IS NULL").Order("sequence ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	var subs []models.WorkSubItem
	if err := s.db.Where("deleted_at IS NULL").Order("sequence ASC, id ASC").Find(&subs).Error; err != nil {
		return nil, err
	}
	var mats []models.Material
	if err := s.db.Where("deleted_at IS NULL").Order("id ASC").Find(&mats).Error; err != nil {
		return nil, err
	}

	catByID := map[uint]models.Category{}
	for _, c := range cats {
		catByID[c.ID] = c
	}
	itemByID := map[uint]models.WorkItem{}
	for _, w := range items {
		itemByID[w.ID] = w
	}

	list := &MasterDataList{
		Categories:   make([]collaboration.PackageCategory, 0, len(cats)),
		WorkItems:    make([]collaboration.PackageWorkItem, 0, len(items)),
		WorkSubItems: make([]collaboration.PackageWorkSubItem, 0, len(subs)),
		Materials:    make([]collaboration.PackageMaterial, 0, len(mats)),
	}
	for _, c := range cats {
		list.Categories = append(list.Categories, collaboration.PackageCategory{
			Code: c.Code, Name: c.Name, Description: c.Description,
			Sequence: c.Sequence, IsActive: c.IsActive,
		})
	}
	for _, w := range items {
		cc := catByID[w.CategoryID].Code
		list.WorkItems = append(list.WorkItems, collaboration.PackageWorkItem{
			Code: w.Code, Name: w.Name, Description: w.Description,
			DefaultUnit: w.DefaultUnit, DefaultQuantity: w.DefaultQuantity,
			DefaultServicePrice: w.DefaultServicePrice, DefaultMaterialPrice: w.DefaultMaterialPrice,
			Notes: w.Notes, Sequence: w.Sequence, IsActive: w.IsActive, CategoryCode: cc,
		})
	}
	for _, sb := range subs {
		wc := itemByID[sb.WorkItemID].Code
		list.WorkSubItems = append(list.WorkSubItems, collaboration.PackageWorkSubItem{
			Code: sb.Code, Sequence: sb.Sequence, Name: sb.Name, Description: sb.Description,
			DifficultyWeight: sb.DifficultyWeight, DefaultUnit: sb.DefaultUnit,
			DefaultQuantity: sb.DefaultQuantity, DefaultServicePrice: sb.DefaultServicePrice,
			DefaultMaterialPrice: sb.DefaultMaterialPrice, Notes: sb.Notes, IsActive: sb.IsActive,
			WorkItemCode: wc,
		})
	}
	for _, m := range mats {
		list.Materials = append(list.Materials, collaboration.PackageMaterial{
			Code: m.Code, Name: m.Name, Description: m.Description, Unit: m.Unit,
			DefaultPrice: m.DefaultPrice, Supplier: m.Supplier, Notes: m.Notes, IsActive: m.IsActive,
		})
	}
	return list, nil
}

// BuildPackageFiltered menyusun MasterDataPackage berdasarkan filter selection.
func (s *Service) BuildPackageFiltered(senderID, senderName, roomID string, sel FilterSelection) (*collaboration.MasterDataPackage, error) {
	if sel.SendAll {
		return s.BuildPackage(senderID, senderName, roomID)
	}

	list, err := s.ListAllMasterData()
	if err != nil {
		return nil, err
	}

	selCat := toSet(sel.CategoryCodes)
	selItem := toSet(sel.WorkItemCodes)
	selMat := toSet(sel.MaterialCodes)
	selSub := toSet(sel.SubItemKeys)

	pkg := &collaboration.MasterDataPackage{
		Metadata: collaboration.MasterPackageMetadata{
			PackageID:      newPackageID(),
			SenderID:       senderID,
			SenderName:     senderName,
			RoomID:         roomID,
			CreatedAt:      time.Now().UTC(),
			SchemaVersion:  collaboration.PackageSchemaVersion,
			PackageVersion: "1",
		},
		Data: collaboration.MasterPackageData{
			Categories:   make([]collaboration.PackageCategory, 0),
			WorkItems:    make([]collaboration.PackageWorkItem, 0),
			WorkSubItems: make([]collaboration.PackageWorkSubItem, 0),
			Materials:    make([]collaboration.PackageMaterial, 0),
		},
	}

	for _, c := range list.Categories {
		if len(selCat) > 0 && !selCat[c.Code] {
			continue
		}
		pkg.Data.Categories = append(pkg.Data.Categories, c)
	}

	for _, w := range list.WorkItems {
		if len(selItem) > 0 && !selItem[w.Code] {
			continue
		}
		pkg.Data.WorkItems = append(pkg.Data.WorkItems, w)
	}

	for _, sb := range list.WorkSubItems {
		if len(selSub) > 0 {
			key := sb.WorkItemCode + ":" + sb.Code
			if sb.Code == "" {
				key = sb.WorkItemCode + "::" + sb.Name
			}
			if !selSub[key] {
				continue
			}
		}
		pkg.Data.WorkSubItems = append(pkg.Data.WorkSubItems, sb)
	}

	for _, m := range list.Materials {
		if len(selMat) > 0 && !selMat[m.Code] {
			continue
		}
		pkg.Data.Materials = append(pkg.Data.Materials, m)
	}

	sum, err := pkg.ComputeChecksum()
	if err != nil {
		return nil, err
	}
	pkg.Checksum = sum
	return pkg, nil
}

func toSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	m := make(map[string]bool, len(items))
	for _, v := range items {
		m[v] = true
	}
	return m
}

// BuildPackage mengambil seluruh Master Data lokal (kategori, pekerjaan,
// sub-pekerjaan, material) dan menyusun MasterDataPackage untuk dikirim.
func (s *Service) BuildPackage(senderID, senderName, roomID string) (*collaboration.MasterDataPackage, error) {
	var cats []models.Category
	if err := s.db.Where("deleted_at IS NULL").Order("sequence ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	var items []models.WorkItem
	if err := s.db.Where("deleted_at IS NULL").Order("sequence ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	var subs []models.WorkSubItem
	if err := s.db.Where("deleted_at IS NULL").Order("sequence ASC, id ASC").Find(&subs).Error; err != nil {
		return nil, err
	}
	var mats []models.Material
	if err := s.db.Where("deleted_at IS NULL").Order("id ASC").Find(&mats).Error; err != nil {
		return nil, err
	}

	catByID := map[uint]models.Category{}
	for _, c := range cats {
		catByID[c.ID] = c
	}
	itemByID := map[uint]models.WorkItem{}
	for _, w := range items {
		itemByID[w.ID] = w
	}

	pkg := &collaboration.MasterDataPackage{
		Metadata: collaboration.MasterPackageMetadata{
			PackageID:      newPackageID(),
			SenderID:       senderID,
			SenderName:     senderName,
			RoomID:         roomID,
			CreatedAt:      time.Now().UTC(),
			SchemaVersion:  collaboration.PackageSchemaVersion,
			PackageVersion: "1",
		},
		Data: collaboration.MasterPackageData{
			Categories:   make([]collaboration.PackageCategory, 0, len(cats)),
			WorkItems:    make([]collaboration.PackageWorkItem, 0, len(items)),
			WorkSubItems: make([]collaboration.PackageWorkSubItem, 0, len(subs)),
			Materials:    make([]collaboration.PackageMaterial, 0, len(mats)),
		},
	}
	for _, c := range cats {
		pkg.Data.Categories = append(pkg.Data.Categories, collaboration.PackageCategory{
			Code: c.Code, Name: c.Name, Description: c.Description,
			Sequence: c.Sequence, IsActive: c.IsActive,
		})
	}
	for _, w := range items {
		cc := catByID[w.CategoryID].Code
		pkg.Data.WorkItems = append(pkg.Data.WorkItems, collaboration.PackageWorkItem{
			Code: w.Code, Name: w.Name, Description: w.Description,
			DefaultUnit: w.DefaultUnit, DefaultQuantity: w.DefaultQuantity,
			DefaultServicePrice: w.DefaultServicePrice, DefaultMaterialPrice: w.DefaultMaterialPrice,
			Notes: w.Notes, Sequence: w.Sequence, IsActive: w.IsActive, CategoryCode: cc,
		})
	}
	for _, sb := range subs {
		wc := itemByID[sb.WorkItemID].Code
		pkg.Data.WorkSubItems = append(pkg.Data.WorkSubItems, collaboration.PackageWorkSubItem{
			Code: sb.Code, Sequence: sb.Sequence, Name: sb.Name, Description: sb.Description,
			DifficultyWeight: sb.DifficultyWeight, DefaultUnit: sb.DefaultUnit,
			DefaultQuantity: sb.DefaultQuantity, DefaultServicePrice: sb.DefaultServicePrice,
			DefaultMaterialPrice: sb.DefaultMaterialPrice, Notes: sb.Notes, IsActive: sb.IsActive,
			WorkItemCode: wc,
		})
	}
	for _, m := range mats {
		pkg.Data.Materials = append(pkg.Data.Materials, collaboration.PackageMaterial{
			Code: m.Code, Name: m.Name, Description: m.Description, Unit: m.Unit,
			DefaultPrice: m.DefaultPrice, Supplier: m.Supplier, Notes: m.Notes, IsActive: m.IsActive,
		})
	}
	sum, err := pkg.ComputeChecksum()
	if err != nil {
		return nil, err
	}
	pkg.Checksum = sum
	return pkg, nil
}

// ===== Compare =====

// DiffItem menggambarkan satu perubahan Master Data hasil perbandingan
// terhadap data lokal (untuk pratinjau sebelum memasang).
type DiffItem struct {
	Kind    string `json:"kind"`
	Entity  string `json:"entity"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// Compare membandingkan package terhadap data lokal dan menghasilkan daftar
// perubahan (NEW/UPDATED/UNCHANGED/CONFLICT) untuk pratinjau.
func (s *Service) Compare(pkg *collaboration.MasterDataPackage) ([]DiffItem, error) {
	var out []DiffItem

	liveCats, err := s.liveCategories()
	if err != nil {
		return nil, err
	}
	for _, pc := range pkg.Data.Categories {
		local, ok := liveCats[pc.Code]
		if !ok {
			out = append(out, DiffItem{Kind: DiffNew, Entity: "category", Code: pc.Code, Name: pc.Name, Summary: "Kategori baru"})
			continue
		}
		kind := diffKind(categoryFields(local), categoryFieldsFromPkg(pc))
		out = append(out, DiffItem{Kind: kind, Entity: "category", Code: pc.Code, Name: pc.Name, Summary: kindLabel(kind)})
	}

	liveItems, err := s.liveWorkItems()
	if err != nil {
		return nil, err
	}
	for _, pw := range pkg.Data.WorkItems {
		local, ok := liveItems[pw.Code]
		if !ok {
			out = append(out, DiffItem{Kind: DiffNew, Entity: "work_item", Code: pw.Code, Name: pw.Name, Summary: "Pekerjaan baru"})
			continue
		}
		kind := diffKind(workItemFields(local), workItemFieldsFromPkg(pw))
		out = append(out, DiffItem{Kind: kind, Entity: "work_item", Code: pw.Code, Name: pw.Name, Summary: kindLabel(kind)})
	}

	for _, ps := range pkg.Data.WorkSubItems {
		out = append(out, DiffItem{Kind: DiffNew, Entity: "work_sub_item", Code: ps.Code, Name: ps.Name, Summary: "Sub-pekerjaan (dibandingkan per induk)"})
	}

	liveMats, err := s.liveMaterials()
	if err != nil {
		return nil, err
	}
	for _, pm := range pkg.Data.Materials {
		local, ok := liveMats[pm.Code]
		if !ok {
			out = append(out, DiffItem{Kind: DiffNew, Entity: "material", Code: pm.Code, Name: pm.Name, Summary: "Material baru"})
			continue
		}
		kind := diffKind(materialFields(local), materialFieldsFromPkg(pm))
		out = append(out, DiffItem{Kind: kind, Entity: "material", Code: pm.Code, Name: pm.Name, Summary: kindLabel(kind)})
	}
	return out, nil
}

// ===== Install =====

// InstallSummary adalah ringkasan hasil pemasangan Master Data.
type InstallSummary struct {
	CategoriesCreated int `json:"categoriesCreated"`
	CategoriesUpdated int `json:"categoriesUpdated"`
	WorkItemsCreated  int `json:"workItemsCreated"`
	WorkItemsUpdated  int `json:"workItemsUpdated"`
	SubItemsCreated   int `json:"subItemsCreated"`
	SubItemsUpdated   int `json:"subItemsUpdated"`
	MaterialsCreated  int `json:"materialsCreated"`
	MaterialsUpdated  int `json:"materialsUpdated"`
	Skipped           int `json:"skipped"`
	Conflicts         int `json:"conflicts"`
}

// Install memasang package ke database lokal dalam satu transaksi. Jika satu
// langkah gagal, seluruh perubahan di-rollback penuh. `decisions` berisi
// keputusan per item ber-format "entity:code" → strategi; strategi tingkat
// package dipakai sebagai default.
func (s *Service) Install(pkg *collaboration.MasterDataPackage, strategy string, decisions map[string]string) (*InstallSummary, error) {
	if pkg == nil {
		return nil, services.NewValidationError("Package kosong.")
	}
	ok, err := pkg.VerifyChecksum()
	if err != nil {
		return nil, services.NewConflictError("Gagal memverifikasi checksum package.")
	}
	if !ok {
		return nil, services.NewConflictError("Package rusak: checksum tidak cocok. Transfer dibatalkan.")
	}
	strategy = normalizeStrategy(strategy)
	if strategy == "" {
		strategy = StrategyPrompt
	}

	sum := &InstallSummary{}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		catID := map[string]uint{}
		itemID := map[string]uint{}
		// 1) Kategori (buat dulu; tidak ada induk).
		for _, pc := range pkg.Data.Categories {
			decision := decide(decisions, "category", pc.Code, strategy)
			existing, found, err := s.findCategory(tx, pc.Code)
			if err != nil {
				return err
			}
			if found {
				catID[pc.Code] = existing.ID
				if reflect.DeepEqual(categoryFields(existing), categoryFieldsFromPkg(pc)) {
					continue // tidak berubah
				}
				switch decision {
				case StrategyUseIncoming:
					if err := s.cats.Update(tx, applyCategory(existing, pc)); err != nil {
						return err
					}
					sum.CategoriesUpdated++
				case StrategyUseLocal:
					sum.Skipped++
				default: // PROMPT / SKIP → lewati konflik
					sum.Conflicts++
					sum.Skipped++
				}
				continue
			}
			nc := &models.Category{Code: pc.Code, Name: pc.Name, Description: pc.Description,
				Sequence: pc.Sequence, IsActive: pc.IsActive}
			if err := s.cats.Create(tx, nc); err != nil {
				return err
			}
			catID[pc.Code] = nc.ID
			sum.CategoriesCreated++
			_ = (&services.AuditWriter{}).Write(tx, "CREATE", "category", nc.ID,
				fmt.Sprintf("Master Data: kategori \"%s\" (%s)", nc.Name, nc.Code))
		}

		// 2) Pekerjaan (butuh kategori dari natural key CategoryCode).
		for _, pw := range pkg.Data.WorkItems {
			decision := decide(decisions, "work_item", pw.Code, strategy)
			cid, okCID := catID[pw.CategoryCode]
			if !okCID {
				// kategori induk tidak ikut package; cari yang sudah ada
				ex, found, err := s.findCategory(tx, pw.CategoryCode)
				if err != nil {
					return err
				}
				if !found {
					sum.Skipped++
					continue
				}
				cid = ex.ID
			}
			existing, found, err := s.findWorkItemByCode(tx, pw.Code)
			if err != nil {
				return err
			}
			if found {
				itemID[pw.Code] = existing.ID
				if reflect.DeepEqual(workItemFields(existing), workItemFieldsFromPkg(pw)) {
					continue
				}
				switch decision {
				case StrategyUseIncoming:
					if err := s.items.Update(tx, applyWorkItem(existing, pw, cid)); err != nil {
						return err
					}
					sum.WorkItemsUpdated++
				case StrategyUseLocal:
					sum.Skipped++
				default:
					sum.Conflicts++
					sum.Skipped++
				}
				continue
			}
			nw := &models.WorkItem{CategoryID: cid, Code: pw.Code, Name: pw.Name, Description: pw.Description,
				DefaultUnit: pw.DefaultUnit, DefaultQuantity: pw.DefaultQuantity,
				DefaultServicePrice: pw.DefaultServicePrice, DefaultMaterialPrice: pw.DefaultMaterialPrice,
				Notes: pw.Notes, Sequence: pw.Sequence, IsActive: pw.IsActive}
			if err := s.items.Create(tx, nw); err != nil {
				return err
			}
			itemID[pw.Code] = nw.ID
			sum.WorkItemsCreated++
			_ = (&services.AuditWriter{}).Write(tx, "CREATE", "work_item", nw.ID,
				fmt.Sprintf("Master Data: pekerjaan \"%s\" (%s)", nw.Name, nw.Code))
		}

		// 3) Sub-pekerjaan (butuh pekerjaan dari natural key WorkItemCode).
		for _, ps := range pkg.Data.WorkSubItems {
			decision := decide(decisions, "work_sub_item", ps.Code, strategy)
			wiID, okWI := itemID[ps.WorkItemCode]
			if !okWI {
				ex, found, err := s.findWorkItemByCode(tx, ps.WorkItemCode)
				if err != nil {
					return err
				}
				if !found {
					sum.Skipped++
					continue
				}
				wiID = ex.ID
			}
			existing, found, err := s.findSubByParent(tx, wiID, ps.Code, ps.Name)
			if err != nil {
				return err
			}
			if found {
				if reflect.DeepEqual(workSubFields(existing), workSubFieldsFromPkg(ps)) {
					continue
				}
				switch decision {
				case StrategyUseIncoming:
					if err := s.subs.Update(tx, applyWorkSub(existing, ps)); err != nil {
						return err
					}
					sum.SubItemsUpdated++
				case StrategyUseLocal:
					sum.Skipped++
				default:
					sum.Conflicts++
					sum.Skipped++
				}
				continue
			}
			ns := &models.WorkSubItem{WorkItemID: wiID, Code: ps.Code, Sequence: ps.Sequence, Name: ps.Name,
				Description: ps.Description, DifficultyWeight: ps.DifficultyWeight,
				DefaultUnit: ps.DefaultUnit, DefaultQuantity: ps.DefaultQuantity,
				DefaultServicePrice: ps.DefaultServicePrice, DefaultMaterialPrice: ps.DefaultMaterialPrice,
				Notes: ps.Notes, IsActive: ps.IsActive}
			if err := s.subs.Create(tx, ns); err != nil {
				return err
			}
			sum.SubItemsCreated++
			_ = (&services.AuditWriter{}).Write(tx, "CREATE", "work_sub_item", ns.ID,
				fmt.Sprintf("Master Data: sub-pekerjaan \"%s\" (%s)", ns.Name, ns.Code))
		}

		// 4) Material (mandiri).
		for _, pm := range pkg.Data.Materials {
			decision := decide(decisions, "material", pm.Code, strategy)
			existing, found, err := s.findMaterial(tx, pm.Code)
			if err != nil {
				return err
			}
			if found {
				if reflect.DeepEqual(materialFields(existing), materialFieldsFromPkg(pm)) {
					continue
				}
				switch decision {
				case StrategyUseIncoming:
					if err := s.mats.Update(tx, applyMaterial(existing, pm)); err != nil {
						return err
					}
					sum.MaterialsUpdated++
				case StrategyUseLocal:
					sum.Skipped++
				default:
					sum.Conflicts++
					sum.Skipped++
				}
				continue
			}
			nm := &models.Material{Code: pm.Code, Name: pm.Name, Description: pm.Description, Unit: pm.Unit,
				DefaultPrice: pm.DefaultPrice, Supplier: pm.Supplier, Notes: pm.Notes, IsActive: pm.IsActive}
			if err := s.mats.Create(tx, nm); err != nil {
				return err
			}
			sum.MaterialsCreated++
			_ = (&services.AuditWriter{}).Write(tx, "CREATE", "material", nm.ID,
				fmt.Sprintf("Master Data: material \"%s\" (%s)", nm.Name, nm.Code))
		}
		return nil
	})
	if err != nil {
		if services.IsFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal memasang master data", "error", err)
		return nil, fmt.Errorf("gagal memasang Master Data; seluruh perubahan dibatalkan")
	}
	return sum, nil
}

// ===== Inbox & Sent =====

// InboxItem adalah tampilan ringkas satu paket masuk.
type InboxItem struct {
	PackageID     string     `json:"packageId"`
	SenderID      string     `json:"senderId"`
	SenderName    string     `json:"senderName"`
	RoomID        string     `json:"roomId"`
	SourceVersion int        `json:"sourceVersion"`
	Title         string     `json:"title"`
	Summary       string     `json:"summary"`
	ItemCount     int        `json:"itemCount"`
	Status        string     `json:"status"`
	ReceivedAt    time.Time  `json:"receivedAt"`
	InstalledAt   *time.Time `json:"installedAt,omitempty"`
	RejectedAt    *time.Time `json:"rejectedAt,omitempty"`
}

// SaveInbox menyimpan paket masuk (dedup by PackageID).
func (s *Service) SaveInbox(m *models.MasterInbox) error {
	var existing models.MasterInbox
	err := s.db.Where("package_id = ?", m.PackageID).First(&existing).Error
	if err == nil {
		return nil // sudah ada (duplikat)
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return s.db.Create(m).Error
}

// InboxList menampilkan daftar paket masuk (terbaru dulu).
func (s *Service) InboxList() ([]InboxItem, error) {
	var rows []models.MasterInbox
	if err := s.db.Order("received_at DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]InboxItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.inboxView(r))
	}
	return out, nil
}

// InboxGet mengembalikan tampilan satu paket masuk.
func (s *Service) InboxGet(packageID string) (*InboxItem, error) {
	r, err := s.getInbox(packageID)
	if err != nil {
		return nil, err
	}
	v := s.inboxView(*r)
	return &v, nil
}

// GetInboxPayload memuat ulang package mentah (untuk pratinjau / pemasangan).
func (s *Service) GetInboxPayload(packageID string) (*collaboration.MasterDataPackage, error) {
	r, err := s.getInbox(packageID)
	if err != nil {
		return nil, err
	}
	return collaboration.DeserializeMasterDataPackage([]byte(r.Payload))
}

// SetInboxStatus memperbarui status paket masuk.
func (s *Service) SetInboxStatus(packageID, status string, at time.Time) error {
	updates := map[string]interface{}{"status": status}
	if status == models.MasterStatusInstalled {
		updates["installed_at"] = at
	} else if status == models.MasterStatusRejected {
		updates["rejected_at"] = at
	}
	return s.db.Model(&models.MasterInbox{}).Where("package_id = ?", packageID).Updates(updates).Error
}

// SentItem adalah tampilan ringkas satu paket terkirim.
type SentItem struct {
	PackageID     string    `json:"packageId"`
	RoomID        string    `json:"roomId"`
	SourceVersion int       `json:"sourceVersion"`
	Title         string    `json:"title"`
	ItemCount     int       `json:"itemCount"`
	Status        string    `json:"status"`
	Recipients    string    `json:"recipients"`
	SentAt        time.Time `json:"sentAt"`
}

// SentList menampilkan daftar paket terkirim.
func (s *Service) SentList() ([]SentItem, error) {
	var rows []models.MasterSent
	if err := s.db.Order("sent_at DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]SentItem, 0, len(rows))
	for _, r := range rows {
		title, count := describePayload([]byte(r.Payload))
		out = append(out, SentItem{
			PackageID: r.PackageID, RoomID: r.RoomID, SourceVersion: r.SourceVersion,
			Title: title, ItemCount: count, Status: r.Status, Recipients: r.Recipients, SentAt: r.SentAt,
		})
	}
	return out, nil
}

// MapInboxToSent menyalin paket masuk sebagai paket terkirim (saat meneruskan).
func (s *Service) SaveSent(m *models.MasterSent) error {
	return s.db.Create(m).Error
}

// UpdateSentStatus memperbarui status pengiriman tingkat paket.
func (s *Service) UpdateSentStatus(packageID, status string) error {
	return s.db.Model(&models.MasterSent{}).Where("package_id = ?", packageID).
		Update("status", status).Error
}

func (s *Service) getInbox(packageID string) (*models.MasterInbox, error) {
	var r models.MasterInbox
	err := s.db.Where("package_id = ?", packageID).First(&r).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, services.ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *Service) inboxView(r models.MasterInbox) InboxItem {
	title, count := describePayload([]byte(r.Payload))
	return InboxItem{
		PackageID: r.PackageID, SenderID: r.SenderID, SenderName: r.SenderName, RoomID: r.RoomID,
		SourceVersion: r.SourceVersion, Title: title, Summary: title, ItemCount: count,
		Status: r.Status, ReceivedAt: r.ReceivedAt, InstalledAt: r.InstalledAt, RejectedAt: r.RejectedAt,
	}
}

// ===== helpers =====

func newPackageID() string {
	return "md-" + time.Now().UTC().Format("20060102150405") + "-" + randomSuffix()
}

func (s *Service) liveCategories() (map[string]models.Category, error) {
	var rows []models.Category
	if err := s.db.Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]models.Category{}
	for _, r := range rows {
		if r.Code != "" {
			out[r.Code] = r
		}
	}
	return out, nil
}

func (s *Service) liveWorkItems() (map[string]models.WorkItem, error) {
	var rows []models.WorkItem
	if err := s.db.Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]models.WorkItem{}
	for _, r := range rows {
		if r.Code != "" {
			out[r.Code] = r
		}
	}
	return out, nil
}

func (s *Service) liveMaterials() (map[string]models.Material, error) {
	var rows []models.Material
	if err := s.db.Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]models.Material{}
	for _, r := range rows {
		if r.Code != "" {
			out[r.Code] = r
		}
	}
	return out, nil
}

func (s *Service) findCategory(tx *gorm.DB, code string) (models.Category, bool, error) {
	if code == "" {
		return models.Category{}, false, nil
	}
	var r models.Category
	err := tx.Where("code = ? AND deleted_at IS NULL", code).First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return models.Category{}, false, nil
	}
	if err != nil {
		return models.Category{}, false, err
	}
	return r, true, nil
}

func (s *Service) findWorkItemByCode(tx *gorm.DB, code string) (models.WorkItem, bool, error) {
	if code == "" {
		return models.WorkItem{}, false, nil
	}
	var r models.WorkItem
	err := tx.Where("code = ? AND deleted_at IS NULL", code).First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return models.WorkItem{}, false, nil
	}
	if err != nil {
		return models.WorkItem{}, false, err
	}
	return r, true, nil
}

func (s *Service) findSubByParent(tx *gorm.DB, workItemID uint, code, name string) (models.WorkSubItem, bool, error) {
	var r models.WorkSubItem
	q := tx.Where("work_item_id = ? AND deleted_at IS NULL", workItemID)
	if code != "" {
		q = q.Where("code = ?", code)
	} else {
		q = q.Where("name = ?", name)
	}
	err := q.First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return models.WorkSubItem{}, false, nil
	}
	if err != nil {
		return models.WorkSubItem{}, false, err
	}
	return r, true, nil
}

func (s *Service) findMaterial(tx *gorm.DB, code string) (models.Material, bool, error) {
	if code == "" {
		return models.Material{}, false, nil
	}
	var r models.Material
	err := tx.Where("code = ? AND deleted_at IS NULL", code).First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return models.Material{}, false, nil
	}
	if err != nil {
		return models.Material{}, false, err
	}
	return r, true, nil
}

func decide(decisions map[string]string, entity, code, fallback string) string {
	if decisions != nil {
		if v, ok := decisions[entity+":"+code]; ok && v != "" {
			return normalizeStrategy(v)
		}
	}
	return fallback
}

func normalizeStrategy(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case StrategyUseLocal, StrategyUseIncoming, StrategySkip, StrategyPrompt:
		return strings.ToUpper(strings.TrimSpace(s))
	}
	return ""
}

// applyCategory/applyWorkItem/... mengisi field model dari data package.

func applyCategory(dst models.Category, pc collaboration.PackageCategory) *models.Category {
	dst.Name = pc.Name
	dst.Description = pc.Description
	dst.Sequence = pc.Sequence
	dst.IsActive = pc.IsActive
	return &dst
}

func applyWorkItem(dst models.WorkItem, pw collaboration.PackageWorkItem, categoryID uint) *models.WorkItem {
	dst.CategoryID = categoryID
	dst.Name = pw.Name
	dst.Description = pw.Description
	dst.DefaultUnit = pw.DefaultUnit
	dst.DefaultQuantity = pw.DefaultQuantity
	dst.DefaultServicePrice = pw.DefaultServicePrice
	dst.DefaultMaterialPrice = pw.DefaultMaterialPrice
	dst.Notes = pw.Notes
	dst.Sequence = pw.Sequence
	dst.IsActive = pw.IsActive
	return &dst
}

func applyWorkSub(dst models.WorkSubItem, ps collaboration.PackageWorkSubItem) *models.WorkSubItem {
	dst.Sequence = ps.Sequence
	dst.Name = ps.Name
	dst.Description = ps.Description
	dst.DifficultyWeight = ps.DifficultyWeight
	dst.DefaultUnit = ps.DefaultUnit
	dst.DefaultQuantity = ps.DefaultQuantity
	dst.DefaultServicePrice = ps.DefaultServicePrice
	dst.DefaultMaterialPrice = ps.DefaultMaterialPrice
	dst.Notes = ps.Notes
	dst.IsActive = ps.IsActive
	return &dst
}

func applyMaterial(dst models.Material, pm collaboration.PackageMaterial) *models.Material {
	dst.Name = pm.Name
	dst.Description = pm.Description
	dst.Unit = pm.Unit
	dst.DefaultPrice = pm.DefaultPrice
	dst.Supplier = pm.Supplier
	dst.Notes = pm.Notes
	dst.IsActive = pm.IsActive
	return &dst
}

// diffKind mengembalikan kategori perbedaan dua snapshot field.
func diffKind(local, incoming []interface{}) string {
	if reflect.DeepEqual(local, incoming) {
		return DiffUnchanged
	}
	return DiffConflict // data lokal berbeda dengan data masuk → perlu strategi
}

func kindLabel(k string) string {
	switch k {
	case DiffNew:
		return "Baru"
	case DiffUpdated:
		return "Diperbarui"
	case DiffUnchanged:
		return "Tidak berubah"
	case DiffConflict:
		return "Konflik (data lokal berbeda)"
	}
	return k
}

func categoryFields(c models.Category) []interface{} {
	return []interface{}{c.Name, c.Description, c.Sequence, c.IsActive}
}
func categoryFieldsFromPkg(pc collaboration.PackageCategory) []interface{} {
	return []interface{}{pc.Name, pc.Description, pc.Sequence, pc.IsActive}
}
func workItemFields(w models.WorkItem) []interface{} {
	return []interface{}{w.Name, w.Description, w.DefaultUnit, w.DefaultQuantity,
		w.DefaultServicePrice, w.DefaultMaterialPrice, w.Notes, w.Sequence, w.IsActive}
}
func workItemFieldsFromPkg(pw collaboration.PackageWorkItem) []interface{} {
	return []interface{}{pw.Name, pw.Description, pw.DefaultUnit, pw.DefaultQuantity,
		pw.DefaultServicePrice, pw.DefaultMaterialPrice, pw.Notes, pw.Sequence, pw.IsActive}
}
func workSubFields(sb models.WorkSubItem) []interface{} {
	return []interface{}{sb.Sequence, sb.Name, sb.Description, sb.DifficultyWeight,
		sb.DefaultUnit, sb.DefaultQuantity, sb.DefaultServicePrice, sb.DefaultMaterialPrice,
		sb.Notes, sb.IsActive}
}
func workSubFieldsFromPkg(ps collaboration.PackageWorkSubItem) []interface{} {
	return []interface{}{ps.Sequence, ps.Name, ps.Description, ps.DifficultyWeight,
		ps.DefaultUnit, ps.DefaultQuantity, ps.DefaultServicePrice, ps.DefaultMaterialPrice,
		ps.Notes, ps.IsActive}
}
func materialFields(m models.Material) []interface{} {
	return []interface{}{m.Name, m.Description, m.Unit, m.DefaultPrice, m.Supplier, m.Notes, m.IsActive}
}
func materialFieldsFromPkg(pm collaboration.PackageMaterial) []interface{} {
	return []interface{}{pm.Name, pm.Description, pm.Unit, pm.DefaultPrice, pm.Supplier, pm.Notes, pm.IsActive}
}

// describePayload mengekstrak judul & jumlah item dari paket mentah.
func describePayload(raw []byte) (string, int) {
	var md collaboration.MasterDataPackage
	if err := json.Unmarshal(raw, &md); err != nil {
		return "Master Data", 0
	}
	total := len(md.Data.Categories) + len(md.Data.WorkItems) + len(md.Data.WorkSubItems) + len(md.Data.Materials)
	title := "Master Data (" + itoa(total) + " item)"
	return title, total
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func randomSuffix() string {
	// cukup unik untuk PackageID; tidak butuh kriptografi.
	b := make([]byte, 4)
	seed := time.Now().UnixNano()
	for i := range b {
		seed = seed*6364136223846793005 + 1442695040888963407
		b[i] = byte((seed>>28)%26) + 'a'
	}
	return string(b)
}

// SortDiff mengurutkan daftar diff agar terprediksi (entity, lalu code).
func SortDiff(diffs []DiffItem) {
	sort.SliceStable(diffs, func(i, j int) bool {
		if diffs[i].Entity != diffs[j].Entity {
			return diffs[i].Entity < diffs[j].Entity
		}
		return diffs[i].Code < diffs[j].Code
	})
}

