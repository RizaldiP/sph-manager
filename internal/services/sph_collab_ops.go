package services

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// ===== Tipe operasi granular Work Together (Phase 10, docs/collaboration-lan.md §10.9) =====

const (
	OpHeaderUpdated   = "HEADER_UPDATED"
	OpItemAdded       = "ITEM_ADDED"
	OpItemUpdated     = "ITEM_UPDATED"
	OpItemDeleted     = "ITEM_DELETED"
	OpItemMoved       = "ITEM_MOVED"
	OpSubItemAdded    = "SUB_ITEM_ADDED"
	OpSubItemUpdated  = "SUB_ITEM_UPDATED"
	OpSubItemDeleted  = "SUB_ITEM_DELETED"
	OpSubItemMoved    = "SUB_ITEM_MOVED"
	OpPriceUpdated    = "PRICE_UPDATED"
	OpQuantityUpdated = "QUANTITY_UPDATED"
	OpWeightUpdated   = "WEIGHT_UPDATED"
)

// HeaderPatch: nilai header penuh yang dikirim editor kolaborasi.
type HeaderPatch struct {
	Date        string `json:"date"`
	CustomerID  uint   `json:"customerId"`
	VesselID    *uint  `json:"vesselId,omitempty"`
	ProjectName string `json:"projectName"`
	Subject     string `json:"subject"`
	Reference   string `json:"reference"`
	Location    string `json:"location"`
	ValidUntil  string `json:"validUntil"`
	PicName     string `json:"picName"`
	Notes       string `json:"notes"`
}

// ItemFields: nilai baris main point penuh (bukan partial patch agar idempotent).
type ItemFields struct {
	WorkItemID        *uint   `json:"workItemId,omitempty"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Quantity          float64 `json:"quantity"`
	Unit              string  `json:"unit"`
	ServiceUnitPrice  int64   `json:"serviceUnitPrice"`
	MaterialUnitPrice int64   `json:"materialUnitPrice"`
	PricingMode       string  `json:"pricingMode"`
	Notes             string  `json:"notes"`
}

// SubItemFields: nilai baris sub point penuh. Pada induk PEMBOBOTAN hanya Weight
// yang berpengaruh (harga satuan sub dinolkan oleh alokasi BR-02..BR-04).
type SubItemFields struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Quantity          float64 `json:"quantity"`
	Unit              string  `json:"unit"`
	Weight            int     `json:"weight"`
	ServiceUnitPrice  int64   `json:"serviceUnitPrice"`
	MaterialUnitPrice int64   `json:"materialUnitPrice"`
	Notes             string  `json:"notes"`
}

// OpPayload: satu operasi edit dari peserta room (host maupun client).
// Untuk ITEM_MOVED/SUB_ITEM_MOVED, ToIndex adalah posisi akhir SETELAH baris
// dikeluarkan dari posisi lamanya (dipatok ke rentang valid).
type OpPayload struct {
	Type      string         `json:"type"`
	ItemID    uint           `json:"itemId,omitempty"`
	SubItemID uint           `json:"subItemId,omitempty"`
	ToIndex   *int           `json:"toIndex,omitempty"`
	Header    *HeaderPatch   `json:"header,omitempty"`
	Item      *ItemFields    `json:"item,omitempty"`
	SubItem   *SubItemFields `json:"subItem,omitempty"`
}

// CollabActivity: ringkasan aktivitas siap tampil untuk activity log (§10.19).
type CollabActivity struct {
	Actor   string `json:"actor"`
	Action  string `json:"action"`
	Summary string `json:"summary"`
}

// CollabOps menjalankan operasi granular kolaborasi di host dengan memetakan setiap
// op ke jalur UpdateDraft yang sudah teruji (validasi BR-06, roll-up BR-01, alokasi
// BR-02..04, transaksi BR-16, audit BR-13) sehingga tidak ada rumus yang diduplikasi.
type CollabOps struct {
	db    *gorm.DB
	log   *slog.Logger
	sph   *SphService
	repo  *repositories.SphRepository
	audit *AuditWriter
}

func NewCollabOps(db *gorm.DB, sph *SphService, log *slog.Logger) *CollabOps {
	return &CollabOps{
		db:    db,
		log:   log,
		sph:   sph,
		repo:  repositories.NewSphRepository(),
		audit: NewAuditWriter(),
	}
}

// Snapshot mengambil state dokumen terkini untuk dikirim sebagai initial sync / broadcast.
func (c *CollabOps) Snapshot(docID uint) (*models.SphDocument, error) {
	d, err := c.repo.GetByID(c.db, docID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		c.log.Error("gagal mengambil snapshot SPH", "id", docID, "error", err)
		return nil, fmt.Errorf("gagal memuat dokumen")
	}
	return d, nil
}

// Apply menerapkan satu operasi atas dokumen DRAFT dan mengembalikan state dokumen
// terbaru beserta ringkasan aktivitasnya. Aman dipanggil serial dari room host.
func (c *CollabOps) Apply(docID uint, actor string, op *OpPayload) (*models.SphDocument, *CollabActivity, error) {
	if op == nil || trim(op.Type) == "" {
		return nil, nil, NewValidationError("Operasi tidak dikenal.")
	}

	doc, err := c.repo.GetByID(c.db, docID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, ErrNotFound
		}
		c.log.Error("gagal mengambil SPH untuk kolaborasi", "id", docID, "error", err)
		return nil, nil, fmt.Errorf("gagal memuat dokumen")
	}
	if doc.Status != models.StatusDraft {
		return nil, nil, NewConflictError(
			"Hanya dokumen berstatus Draft yang dapat diedit dalam Room (status sekarang: %s).",
			statusLabel(doc.Status))
	}

	in := DocToSaveInput(doc)
	act, err := c.mutateInput(doc, in, op)
	if err != nil {
		return nil, nil, err
	}
	act.Actor = actor

	fresh, err := c.sph.applyDraftUpdate(docID, in, actor)
	if err != nil {
		return nil, nil, err
	}
	return fresh, act, nil
}

// mutateInput menerapkan delta operasi pada salinan input; seluruh perhitungan
// tetap dikerjakan applyDraftUpdate. Nama untuk ringkasan diambil sebelum mutasi.
func (c *CollabOps) mutateInput(doc *models.SphDocument, in SphSaveInput, op *OpPayload) (*CollabActivity, error) {
	actorLabel := "Peserta"
	itemName := func(idx int) string {
		if idx >= 0 && idx < len(in.Items) {
			return in.Items[idx].Name
		}
		return ""
	}
	subName := func(itemIdx, subIdx int) string {
		if itemIdx >= 0 && itemIdx < len(in.Items) && subIdx >= 0 && subIdx < len(in.Items[itemIdx].SubItems) {
			return in.Items[itemIdx].SubItems[subIdx].Name
		}
		return ""
	}

	switch op.Type {
	case OpHeaderUpdated:
		if op.Header == nil {
			return nil, NewValidationError("Data header tidak lengkap.")
		}
		in.Header = SphHeaderInput{
			Date:        op.Header.Date,
			CustomerID:  op.Header.CustomerID,
			VesselID:    op.Header.VesselID,
			ProjectName: op.Header.ProjectName,
			Subject:     op.Header.Subject,
			Reference:   op.Header.Reference,
			Location:    op.Header.Location,
			ValidUntil:  op.Header.ValidUntil,
			PicName:     op.Header.PicName,
			Notes:       op.Header.Notes,
		}
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s memperbarui info dokumen", actorLabel)}, nil

	case OpItemAdded:
		if op.Item == nil {
			return nil, NewValidationError("Data pekerjaan baru tidak lengkap.")
		}
		row := itemFieldsToInput(op.Item, nil)
		pos := len(in.Items)
		if op.ToIndex != nil {
			pos = clampIndex(*op.ToIndex, len(in.Items)+1)
		}
		in.Items = insertAt(in.Items, pos, row)
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s menambahkan pekerjaan \"%s\"", actorLabel, row.Name)}, nil

	case OpItemUpdated, OpPriceUpdated, OpQuantityUpdated:
		idx := indexOfItemID(doc, op.ItemID)
		if idx < 0 {
			return nil, NewValidationError("Pekerjaan tidak ditemukan (mungkin sudah dihapus peserta lain).")
		}
		if op.Type == OpItemUpdated {
			if op.Item == nil {
				return nil, NewValidationError("Data pekerjaan tidak lengkap.")
			}
			keep := in.Items[idx].WorkItemID
			subs := in.Items[idx].SubItems // pembaruan baris tidak menyentuh sub point
			in.Items[idx] = itemFieldsToInput(op.Item, keep)
			in.Items[idx].SubItems = subs
			return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s mengubah pekerjaan \"%s\"", actorLabel, op.Item.Name)}, nil
		}
		if op.Type == OpPriceUpdated {
			if op.Item == nil {
				return nil, NewValidationError("Data harga tidak lengkap.")
			}
			in.Items[idx].ServiceUnitPrice = op.Item.ServiceUnitPrice
			in.Items[idx].MaterialUnitPrice = op.Item.MaterialUnitPrice
			return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s mengubah harga \"%s\"", actorLabel, itemName(idx))}, nil
		}
		if op.Item == nil || op.Item.Quantity <= 0 {
			return nil, NewValidationError("Qty harus lebih besar dari 0.")
		}
		in.Items[idx].Quantity = op.Item.Quantity
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s mengubah qty \"%s\" menjadi %s", actorLabel, itemName(idx), formatQty(op.Item.Quantity))}, nil

	case OpItemDeleted:
		idx := indexOfItemID(doc, op.ItemID)
		if idx < 0 {
			return nil, NewValidationError("Pekerjaan tidak ditemukan (mungkin sudah dihapus peserta lain).")
		}
		name := itemName(idx)
		in.Items = append(in.Items[:idx], in.Items[idx+1:]...)
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s menghapus pekerjaan \"%s\"", actorLabel, name)}, nil

	case OpItemMoved:
		idx := indexOfItemID(doc, op.ItemID)
		if idx < 0 {
			return nil, NewValidationError("Pekerjaan tidak ditemukan (mungkin sudah dihapus peserta lain).")
		}
		if op.ToIndex == nil {
			return nil, NewValidationError("Posisi tujuan tidak lengkap.")
		}
		name := itemName(idx)
		row := in.Items[idx]
		in.Items = append(in.Items[:idx], in.Items[idx+1:]...)
		pos := clampIndex(*op.ToIndex, len(in.Items))
		in.Items = insertAt(in.Items, pos, row)
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s memindahkan pekerjaan \"%s\"", actorLabel, name)}, nil

	case OpSubItemAdded:
		itemIdx := indexOfItemID(doc, op.ItemID)
		if itemIdx < 0 {
			return nil, NewValidationError("Pekerjaan induk tidak ditemukan.")
		}
		if op.SubItem == nil {
			return nil, NewValidationError("Data sub point baru tidak lengkap.")
		}
		row := subFieldsToInput(op.SubItem)
		subs := in.Items[itemIdx].SubItems
		pos := len(subs)
		if op.ToIndex != nil {
			pos = clampIndex(*op.ToIndex, len(subs)+1)
		}
		in.Items[itemIdx].SubItems = insertAt(subs, pos, row)
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s menambahkan sub point \"%s\"", actorLabel, row.Name)}, nil

	case OpSubItemUpdated, OpWeightUpdated:
		itemIdx, subIdx := locateSub(doc, op.SubItemID)
		if itemIdx < 0 {
			return nil, NewValidationError("Sub point tidak ditemukan (mungkin sudah dihapus peserta lain).")
		}
		if op.Type == OpWeightUpdated {
			if op.SubItem == nil || op.SubItem.Weight <= 0 {
				return nil, NewValidationError("Bobot harus lebih besar dari 0.")
			}
			name := subName(itemIdx, subIdx)
			in.Items[itemIdx].SubItems[subIdx].Weight = op.SubItem.Weight
			return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s mengubah bobot sub point \"%s\"", actorLabel, name)}, nil
		}
		if op.SubItem == nil {
			return nil, NewValidationError("Data sub point tidak lengkap.")
		}
		name := op.SubItem.Name
		old := in.Items[itemIdx].SubItems[subIdx]
		in.Items[itemIdx].SubItems[subIdx] = subFieldsToInput(op.SubItem)
		in.Items[itemIdx].SubItems[subIdx].Weight = old.Weight // bobot hanya lewat WEIGHT_UPDATED
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s mengubah sub point \"%s\"", actorLabel, name)}, nil

	case OpSubItemDeleted:
		itemIdx, subIdx := locateSub(doc, op.SubItemID)
		if itemIdx < 0 {
			return nil, NewValidationError("Sub point tidak ditemukan (mungkin sudah dihapus peserta lain).")
		}
		name := subName(itemIdx, subIdx)
		subs := in.Items[itemIdx].SubItems
		in.Items[itemIdx].SubItems = append(subs[:subIdx], subs[subIdx+1:]...)
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s menghapus sub point \"%s\"", actorLabel, name)}, nil

	case OpSubItemMoved:
		itemIdx, subIdx := locateSub(doc, op.SubItemID)
		if itemIdx < 0 {
			return nil, NewValidationError("Sub point tidak ditemukan (mungkin sudah dihapus peserta lain).")
		}
		if op.ToIndex == nil {
			return nil, NewValidationError("Posisi tujuan tidak lengkap.")
		}
		name := subName(itemIdx, subIdx)
		row := in.Items[itemIdx].SubItems[subIdx]
		subs := append(in.Items[itemIdx].SubItems[:subIdx], in.Items[itemIdx].SubItems[subIdx+1:]...)
		pos := clampIndex(*op.ToIndex, len(subs))
		in.Items[itemIdx].SubItems = insertAt(subs, pos, row)
		return &CollabActivity{Action: op.Type, Summary: fmt.Sprintf("%s memindahkan sub point \"%s\"", actorLabel, name)}, nil

	default:
		return nil, NewValidationError("Operasi \"%s\" tidak dikenal.", op.Type)
	}
}

// ===== helper pemetaan dokumen <-> input =====

// DocToSaveInput mengubah snapshot dokumen menjadi payload simpan (urutan & data persis sama),
// dipakai operasi kolaborasi sebagai titik awal sebelum delta diterapkan.
func DocToSaveInput(d *models.SphDocument) SphSaveInput {
	h := SphHeaderInput{
		Date:        d.Date.Format("2006-01-02"),
		CustomerID:  d.CustomerID,
		VesselID:    d.VesselID,
		ProjectName: d.ProjectName,
		Subject:     d.Subject,
		Reference:   d.Reference,
		Location:    d.Location,
		PicName:     d.PicName,
		Notes:       d.Notes,
	}
	if d.ValidUntil != nil {
		h.ValidUntil = d.ValidUntil.Format("2006-01-02")
	}
	items := make([]SphItemInput, 0, len(d.Items))
	for _, it := range d.Items {
		row := SphItemInput{
			ID:                it.ID,
			WorkItemID:        it.WorkItemID,
			Name:              it.NameSnapshot,
			Description:       it.DescriptionSnapshot,
			Quantity:          it.Quantity,
			Unit:              it.Unit,
			ServiceUnitPrice:  it.ServiceUnitPrice,
			MaterialUnitPrice: it.MaterialUnitPrice,
			PricingMode:       it.PricingMode,
			Notes:             it.Notes,
			SubItems:          make([]SphSubItemInput, 0, len(it.SubItems)),
		}
		for _, sb := range it.SubItems {
			row.SubItems = append(row.SubItems, SphSubItemInput{
				ID:                sb.ID,
				Name:              sb.NameSnapshot,
				Description:       sb.DescriptionSnapshot,
				Quantity:          sb.Quantity,
				Unit:              sb.Unit,
				ServiceUnitPrice:  sb.ServiceUnitPrice,
				MaterialUnitPrice: sb.MaterialUnitPrice,
				Weight:            sb.Weight,
				Notes:             sb.Notes,
			})
		}
		items = append(items, row)
	}
	return SphSaveInput{Header: h, Items: items}
}

func itemFieldsToInput(f *ItemFields, keepWorkItem *uint) SphItemInput {
	wid := f.WorkItemID
	if wid == nil {
		wid = keepWorkItem
	}
	mode := f.PricingMode
	if trim(mode) == "" {
		mode = models.PricingModeDirect
	}
	return SphItemInput{
		WorkItemID:        wid,
		Name:              f.Name,
		Description:       f.Description,
		Quantity:          f.Quantity,
		Unit:              f.Unit,
		ServiceUnitPrice:  f.ServiceUnitPrice,
		MaterialUnitPrice: f.MaterialUnitPrice,
		PricingMode:       mode,
		Notes:             f.Notes,
	}
}

func subFieldsToInput(f *SubItemFields) SphSubItemInput {
	return SphSubItemInput{
		Name:              f.Name,
		Description:       f.Description,
		Quantity:          f.Quantity,
		Unit:              f.Unit,
		Weight:            f.Weight,
		ServiceUnitPrice:  f.ServiceUnitPrice,
		MaterialUnitPrice: f.MaterialUnitPrice,
		Notes:             f.Notes,
	}
}

// indexOfItemID memetakan ID item snapshot ke indeks pada urutan input.
func indexOfItemID(d *models.SphDocument, id uint) int {
	if id == 0 {
		return -1
	}
	for i := range d.Items {
		if d.Items[i].ID == id {
			return i
		}
	}
	return -1
}

// locateSub mencari sub point berdasar ID di seluruh item.
func locateSub(d *models.SphDocument, subID uint) (itemIdx, subIdx int) {
	itemIdx, subIdx = -1, -1
	if subID == 0 {
		return
	}
	for i := range d.Items {
		for j := range d.Items[i].SubItems {
			if d.Items[i].SubItems[j].ID == subID {
				return i, j
			}
		}
	}
	return
}

func clampIndex(v, length int) int {
	if v < 0 {
		return 0
	}
	if v > length {
		return length
	}
	return v
}

func insertAt[T any](slice []T, pos int, val T) []T {
	out := make([]T, 0, len(slice)+1)
	out = append(out, slice[:pos]...)
	out = append(out, val)
	out = append(out, slice[pos:]...)
	return out
}

// formatQty merapikan tampilan qty untuk ringkasan aktivitas.
func formatQty(q float64) string {
	if q == float64(int64(q)) {
		return fmt.Sprintf("%d", int64(q))
	}
	return fmt.Sprintf("%g", q)
}
