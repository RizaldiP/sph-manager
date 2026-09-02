package services

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// ===== View & Input (struct input bebas field waktu agar aman binding Wails) =====

type SphDocumentView struct {
	ID             uint   `json:"id"`
	DocumentNumber string `json:"documentNumber"`
	Revision       int    `json:"revision"`
	Date           string `json:"date"`
	CustomerID     uint   `json:"customerId"`
	CustomerName   string `json:"customerName"`
	VesselID       *uint  `json:"vesselId,omitempty"`
	VesselName     string `json:"vesselName"`
	ProjectName    string `json:"projectName"`
	Subject        string `json:"subject"`
	Status         string `json:"status"`
	ItemCount      int64  `json:"itemCount"`
	GrandTotal     int64  `json:"grandTotal"`
	Terbilang      string `json:"terbilang"`
	FinalizedAt    string `json:"finalizedAt,omitempty"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type DashboardStats struct {
	TotalSph      int64             `json:"totalSph"`
	DraftCount    int64             `json:"draftCount"`
	FinalCount    int64             `json:"finalCount"`
	AcceptedCount int64             `json:"acceptedCount"`
	MonthValue    int64             `json:"monthValue"`
	Recent        []SphDocumentView `json:"recent"`
}

type SphSubItemInput struct {
	ID                uint    `json:"id,omitempty"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Quantity          float64 `json:"quantity"`
	Unit              string  `json:"unit"`
	ServiceUnitPrice  int64   `json:"serviceUnitPrice"`
	MaterialUnitPrice int64   `json:"materialUnitPrice"`
	Weight            int     `json:"weight"` // % bobot (mode PEMBOBOTAN)
	Notes             string  `json:"notes"`
}

type SphItemInput struct {
	ID                uint              `json:"id,omitempty"`
	WorkItemID        *uint             `json:"workItemId,omitempty"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Quantity          float64           `json:"quantity"`
	Unit              string            `json:"unit"`
	ServiceUnitPrice  int64             `json:"serviceUnitPrice"`
	MaterialUnitPrice int64             `json:"materialUnitPrice"`
	PricingMode       string            `json:"pricingMode"`
	Notes             string            `json:"notes"`
	SubItems          []SphSubItemInput `json:"subItems"`
}

type SphHeaderInput struct {
	Date        string `json:"date"` // "2006-01-02"
	Sequence    string `json:"sequence"` // nomor urut SPH diinput manual (opsional saat update)
	CustomerID  uint   `json:"customerId"`
	VesselID    *uint  `json:"vesselId,omitempty"`
	ProjectName string `json:"projectName"`
	Subject     string `json:"subject"`
	Reference   string `json:"reference"`
	Location    string `json:"location"`
	ValidUntil  string `json:"validUntil"` // opsional
	PicName     string `json:"picName"`
	Notes       string `json:"notes"`
}

// SphSaveInput adalah payload simpan SPH dari wizard (create maupun update draft).
type SphSaveInput struct {
	Header SphHeaderInput `json:"header"`
	Items  []SphItemInput `json:"items"`
}

// ===== Service =====

// RoomGuard dilaporkan oleh layer kolaborasi (Work Together): dokumen yang sedang
// dibuka dalam Room aktif tidak boleh dimutasi dari jalur solo.
type RoomGuard interface {
	IsDocLocked(sphDocumentID uint) bool
}

// SphService: builder & lifecycle dokumen SPH (FR-S*, BR-01, BR-06..BR-11, BR-13, BR-15, BR-16).
type SphService struct {
	db    *gorm.DB
	log   *slog.Logger
	repo  *repositories.SphRepository
	audit *AuditWriter

	guardMu   sync.Mutex
	roomGuard RoomGuard
}

func NewSphService(db *gorm.DB, log *slog.Logger) *SphService {
	return &SphService{
		db:    db,
		log:   log,
		repo:  repositories.NewSphRepository(),
		audit: NewAuditWriter(),
	}
}

// SetRoomGuard memasang hook kolaborasi (dipasang sekali saat wiring aplikasi).
func (s *SphService) SetRoomGuard(g RoomGuard) {
	s.guardMu.Lock()
	s.roomGuard = g
	s.guardMu.Unlock()
}

func (s *SphService) docLocked(id uint) bool {
	s.guardMu.Lock()
	g := s.roomGuard
	s.guardMu.Unlock()
	return g != nil && g.IsDocLocked(id)
}

const errDocLockedFmt = "Dokumen ini sedang dibuka dalam Room Work Together. Tutup Room terlebih dulu sebelum mengubahnya dari luar."

func toSphViews(db *gorm.DB, docs []models.SphDocument) ([]SphDocumentView, error) {
	out := make([]SphDocumentView, 0, len(docs))
	for i := range docs {
		d := &docs[i]
		n, err := repositories.NewSphRepository().CountItems(db, d.ID)
		if err != nil {
			return nil, err
		}
		v := SphDocumentView{
			ID:             d.ID,
			DocumentNumber: d.DocumentNumber,
			Revision:       d.Revision,
			Date:           d.Date.Format("2006-01-02"),
			CustomerID:     d.CustomerID,
			VesselID:       d.VesselID,
			ProjectName:    d.ProjectName,
			Subject:        d.Subject,
			Status:         d.Status,
			ItemCount:      n,
			GrandTotal:     d.GrandTotal,
			Terbilang:      Terbilang(d.GrandTotal),
			FinalizedAt:    formatTimePtr(d.FinalizedAt),
			CreatedAt:      d.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:      d.UpdatedAt.Format("2006-01-02 15:04"),
		}
		if d.Customer.Name != "" || d.Customer.ID != 0 {
			v.CustomerName = d.Customer.Name
		}
		if d.Vessel != nil {
			v.VesselName = d.Vessel.Name
		}
		out = append(out, v)
	}
	return out, nil
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

// scopeStatuses memetakan tab daftar ke kumpulan status.
func scopeStatuses(scope string) []string {
	switch scope {
	case "open":
		return []string{models.StatusDraft, models.StatusReview}
	case "final":
		return []string{models.StatusFinal, models.StatusSent, models.StatusAccepted, models.StatusRejected, models.StatusCancelled}
	default:
		return nil
	}
}

func (s *SphService) List(scope string, search string, limit int) ([]SphDocumentView, error) {
	docs, err := s.repo.List(s.db, scopeStatuses(scope), trim(search), limit)
	if err != nil {
		s.log.Error("gagal mengambil daftar SPH", "error", err)
		return nil, fmt.Errorf("gagal memuat daftar SPH")
	}
	return toSphViews(s.db, docs)
}

// SuggestNumber mengembalikan saran nomor urut (3 digit) untuk periode tanggal
// tersebut, dipakai UI mengisi field "Nomor Urut" saat membuat SPH.
func (s *SphService) SuggestNumber(date string) (string, error) {
	t, err := time.ParseInLocation("2006-01-02", trim(date), time.Local)
	if err != nil {
		return "", NewValidationError("Format tanggal tidak valid.")
	}
	format, err := settingSphNumberFormat(s.db)
	if err != nil {
		return "", err
	}
	maxSeq, err := maxSequenceForFormat(s.db, format, t)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%03d", maxSeq+1), nil
}

// ComposeNumber merender nomor lengkap dari nomor urut manual + tanggal,
// dipakai preview live di wizard (validasi tanpa menyimpan).
func (s *SphService) ComposeNumber(seq, date string) (string, error) {
	t, err := time.ParseInLocation("2006-01-02", trim(date), time.Local)
	if err != nil {
		return "", NewValidationError("Format tanggal tidak valid.")
	}
	format, err := settingSphNumberFormat(s.db)
	if err != nil {
		return "", err
	}
	return composeNumber(format, seq, t)
}

// Get mengembalikan dokumen lengkap (customer/kapal, item terurut, histori revisi)
// beserta terbilang yang selalu sinkron dengan grand total.
func (s *SphService) Get(id uint) (*models.SphDocument, error) {
	d, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil SPH", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memuat SPH")
	}
	d.Terbilang = Terbilang(d.GrandTotal)
	return d, nil
}

// parseHeader memvalidasi & menormalisasi header dokumen.
func (s *SphService) parseHeader(in SphHeaderInput) (*models.SphDocument, error) {
	h := &models.SphDocument{
		ProjectName: trim(in.ProjectName),
		Subject:     trim(in.Subject),
		Reference:   trim(in.Reference),
		Location:    trim(in.Location),
		PicName:     trim(in.PicName),
		Notes:       trim(in.Notes),
		CustomerID:  in.CustomerID,
		VesselID:    in.VesselID,
	}
	dateStr := trim(in.Date)
	if dateStr == "" {
		return nil, NewValidationError("Tanggal SPH wajib diisi.")
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return nil, NewValidationError("Format tanggal tidak valid.")
	}
	h.Date = t

	if in.ValidUntil != "" && strings.TrimSpace(in.ValidUntil) != "" {
		vu, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(in.ValidUntil), time.Local)
		if err != nil {
			return nil, NewValidationError("Format masa berlaku tidak valid.")
		}
		h.ValidUntil = &vu
	}

	if h.CustomerID == 0 {
		return nil, NewValidationError("Customer wajib dipilih.")
	}
	var custCount int64
	if err := s.db.Model(&models.Customer{}).Where("id = ?", h.CustomerID).Count(&custCount).Error; err != nil {
		s.log.Error("gagal mengambil customer", "id", h.CustomerID, "error", err)
		return nil, fmt.Errorf("gagal menyimpan SPH")
	}
	if custCount == 0 {
		return nil, NewValidationError("Customer tidak ditemukan. Pilih customer yang tersedia.")
	}
	if h.VesselID != nil && *h.VesselID > 0 {
		var v models.Vessel
		if err := s.db.First(&v, *h.VesselID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, NewValidationError("Kapal tidak ditemukan. Pilih kapal yang tersedia.")
			}
			s.log.Error("gagal mengambil kapal", "id", *h.VesselID, "error", err)
			return nil, fmt.Errorf("gagal menyimpan SPH")
		}
		if v.CustomerID != h.CustomerID {
			return nil, NewValidationError("Kapal tidak termasuk milik customer yang dipilih.")
		}
	} else {
		h.VesselID = nil
	}
	return h, nil
}

// lineTotal menghitung qty × harga satuan dengan pembulatan ke Rupiah terdekat (BR-04).
func lineTotal(qty float64, unitPrice int64) int64 {
	return int64(math.Round(qty * float64(unitPrice)))
}

// buildItems menyusun snapshot item & sub item dari input sambil menghitung total (BR-01, BR-02).
func buildItems(inputs []SphItemInput) ([]models.SphItem, int64, int64, error) {
	items := make([]models.SphItem, 0, len(inputs))
	for i := range inputs {
		in := &inputs[i]
		name := trim(in.Name)
		if in.Quantity <= 0 {
			return nil, 0, 0, NewValidationError("Qty pada baris %d harus lebih besar dari 0.", i+1)
		}
		if in.ServiceUnitPrice < 0 || in.MaterialUnitPrice < 0 {
			return nil, 0, 0, NewValidationError("Harga pada baris %d tidak boleh negatif.", i+1)
		}
		mode := models.PricingModeDirect
		if strings.TrimSpace(in.PricingMode) == models.PricingModeWeight {
			mode = models.PricingModeWeight
		}
		svcTot := lineTotal(in.Quantity, in.ServiceUnitPrice)
		matTot := lineTotal(in.Quantity, in.MaterialUnitPrice)

		item := models.SphItem{
			Sequence:            i + 1,
			WorkItemID:          in.WorkItemID,
			NameSnapshot:        name,
			DescriptionSnapshot: trim(in.Description),
			Quantity:            in.Quantity,
			Unit:                trim(in.Unit),
			ServiceUnitPrice:    in.ServiceUnitPrice,
			MaterialUnitPrice:   in.MaterialUnitPrice,
			PricingMode:         mode,
			Notes:               trim(in.Notes),
		}

		if mode == models.PricingModeWeight {
			subs, err := allocateWeightedSubs(in.SubItems, svcTot, matTot)
			if err != nil {
				return nil, 0, 0, err
			}
			for j := range subs {
				subs[j].Sequence = j + 1
			}
			item.SubItems = subs
			item.ServiceTotal = svcTot
			item.MaterialTotal = matTot
			item.Total = svcTot + matTot
			items = append(items, item)
			continue
		}

		for j := range in.SubItems {
			sub := &in.SubItems[j]
			if sub.Quantity <= 0 {
				return nil, 0, 0, NewValidationError("Qty sub point baris %d harus lebih besar dari 0.", i+1)
			}
			if sub.ServiceUnitPrice < 0 || sub.MaterialUnitPrice < 0 {
				return nil, 0, 0, NewValidationError("Harga sub point baris %d tidak boleh negatif.", i+1)
			}
			ssvc := lineTotal(sub.Quantity, sub.ServiceUnitPrice)
			smat := lineTotal(sub.Quantity, sub.MaterialUnitPrice)
			svcTot += ssvc
			matTot += smat
			item.SubItems = append(item.SubItems, models.SphSubItem{
				Sequence:            j + 1,
				NameSnapshot:        trim(sub.Name),
				DescriptionSnapshot: trim(sub.Description),
				Quantity:            sub.Quantity,
				Unit:                trim(sub.Unit),
				ServiceUnitPrice:    sub.ServiceUnitPrice,
				MaterialUnitPrice:   sub.MaterialUnitPrice,
				ServiceTotal:        ssvc,
				MaterialTotal:       smat,
				Total:               ssvc + smat,
				Notes:               trim(sub.Notes),
			})
		}
		// Roll-up: nilai pekerjaan induk mencakup seluruh sub point-nya.
		item.ServiceTotal = svcTot
		item.MaterialTotal = matTot
		item.Total = svcTot + matTot
		items = append(items, item)
	}

	var serviceTotal, materialTotal int64
	for i := range items {
		serviceTotal += items[i].ServiceTotal
		materialTotal += items[i].MaterialTotal
	}
	return items, serviceTotal, materialTotal, nil
}

func (s *SphService) Create(in SphSaveInput) (*SphDocumentView, error) {
	header, err := s.parseHeader(in.Header)
	if err != nil {
		return nil, err
	}
	items, svcTot, matTot, err := buildItems(in.Items)
	if err != nil {
		return nil, err
	}

	doc := &models.SphDocument{
		DocumentNumber:   "", // digenerate dalam transaksi
		Revision:         0,
		Date:             header.Date,
		CustomerID:       header.CustomerID,
		VesselID:         header.VesselID,
		ProjectName:      header.ProjectName,
		Subject:          header.Subject,
		Reference:        header.Reference,
		Location:         header.Location,
		ValidUntil:       header.ValidUntil,
		PicName:          header.PicName,
		Status:           models.StatusDraft,
		SubtotalService:  svcTot,
		SubtotalMaterial: matTot,
		GrandTotal:       svcTot + matTot,
		Notes:            header.Notes,
		Items:            items,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		number, err := manualDocumentNumber(tx, in.Header.Sequence, doc.Date)
		if err != nil {
			return err
		}
		doc.DocumentNumber = number
		if err := s.repo.Create(tx, doc); err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "sph_document", doc.ID,
			fmt.Sprintf("SPH %s dibuat (%d pekerjaan)", number, len(items)))
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal membuat SPH", "error", err)
		return nil, fmt.Errorf("gagal menyimpan SPH")
	}
	views, err := toSphViews(s.db, []models.SphDocument{*doc})
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan SPH")
	}
	return &views[0], nil
}

// UpdateDraft mengubah draft: header + snapshot item baru. Ditolak bila sudah FINAL ke atas (BR-08)
// atau bila dokumen sedang dibuka dalam Room kolaborasi.
func (s *SphService) UpdateDraft(id uint, in SphSaveInput) (*models.SphDocument, error) {
	if s.docLocked(id) {
		return nil, NewConflictError(errDocLockedFmt)
	}
	return s.applyDraftUpdate(id, in, "")
}

// ApplySave applies a full document state to the database. Used by collaboration SyncPush.
func (s *SphService) ApplySave(id uint, in *SphSaveInput, actor string) (*models.SphDocument, error) {
	return s.applyDraftUpdate(id, *in, actor)
}

// applyDraftUpdate inti pembaruan draft tanpa guard room — dipakai jalur solo maupun
// operasi kolaborasi. Actor tidak kosong berarti perubahan dari kolaborator (audit BR-13).
func (s *SphService) applyDraftUpdate(id uint, in SphSaveInput, actor string) (*models.SphDocument, error) {
	existing, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil SPH", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui SPH")
	}
	if existing.Status != models.StatusDraft && existing.Status != models.StatusReview {
		return nil, NewConflictError("Dokumen berstatus %s tidak dapat diedit isinya. Buat revisi untuk mengubahnya.", statusLabel(existing.Status))
	}

	header, err := s.parseHeader(in.Header)
	if err != nil {
		return nil, err
	}
	items, svcTot, matTot, err := buildItems(in.Items)
	if err != nil {
		return nil, err
	}

	existing.Date = header.Date
	existing.CustomerID = header.CustomerID
	existing.VesselID = header.VesselID
	existing.ProjectName = header.ProjectName
	existing.Subject = header.Subject
	existing.Reference = header.Reference
	existing.Location = header.Location
	existing.ValidUntil = header.ValidUntil
	existing.PicName = header.PicName
	existing.Notes = header.Notes
	existing.SubtotalService = svcTot
	existing.SubtotalMaterial = matTot
	existing.GrandTotal = svcTot + matTot

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(tx, existing); err != nil {
			return err
		}
		if err := s.repo.ReplaceItems(tx, existing.ID, items); err != nil {
			return err
		}
		desc := fmt.Sprintf("SPH %s Rev %d diperbarui", existing.DocumentNumber, existing.Revision)
		if actor != "" {
			desc = fmt.Sprintf("SPH %s Rev %d diperbarui via kolaborasi oleh %s", existing.DocumentNumber, existing.Revision, actor)
		}
		return s.audit.WriteAs(tx, "UPDATE", "sph_document", existing.ID, desc, actor)
	})
	if err != nil {
		if isFriendly(err) {
			return nil, err
		}
		s.log.Error("gagal memperbarui SPH", "id", id, "error", err)
		return nil, fmt.Errorf("gagal memperbarui SPH")
	}
	return s.repo.GetByID(s.db, id)
}

// validateForFinalization memeriksa seluruh syarat BR-06.
func (s *SphService) validateForFinalization(d *models.SphDocument) error {
	if strings.TrimSpace(d.DocumentNumber) == "" {
		return NewValidationError("Nomor SPH wajib terisi sebelum finalisasi.")
	}
	if d.Date.IsZero() {
		return NewValidationError("Tanggal SPH wajib diisi.")
	}
	if d.CustomerID == 0 {
		return NewValidationError("Customer wajib dipilih.")
	}
	if len(d.Items) == 0 {
		return NewValidationError("Minimal satu pekerjaan harus ada sebelum finalisasi.")
	}
	var grand int64
	for i := range d.Items {
		it := &d.Items[i]
		if it.Quantity <= 0 {
			return NewValidationError("Qty \"%s\" harus lebih besar dari 0.", it.NameSnapshot)
		}
		if it.ServiceUnitPrice < 0 || it.MaterialUnitPrice < 0 {
			return NewValidationError("Harga \"%s\" tidak boleh negatif.", it.NameSnapshot)
		}
		mainSvc := lineTotal(it.Quantity, it.ServiceUnitPrice)
		mainMat := lineTotal(it.Quantity, it.MaterialUnitPrice)
		if it.PricingMode == models.PricingModeWeight {
			ws := make([]int, len(it.SubItems))
			for j := range it.SubItems {
				ws[j] = it.SubItems[j].Weight
			}
			sumW := weightSum(ws)
			if sumW != 100 {
				return NewValidationError("Total bobot \"%s\" = %d%%, harus tepat 100%% (selisih %+d%%). Perbaiki bobot sebelum finalisasi.", it.NameSnapshot, sumW, 100-sumW)
			}
			expSvc := allocateLargestRemainder(mainSvc, ws)
			expMat := allocateLargestRemainder(mainMat, ws)
			for j := range it.SubItems {
				sb := it.SubItems[j]
				if sb.ServiceTotal != expSvc[j] || sb.MaterialTotal != expMat[j] ||
					sb.Total != expSvc[j]+expMat[j] || sb.AllocatedValue != expSvc[j]+expMat[j] {
					return NewValidationError("Alokasi bobot \"%s\" tidak konsisten. Muat ulang dokumen lalu coba lagi.", it.NameSnapshot)
				}
			}
			if it.ServiceTotal != mainSvc || it.MaterialTotal != mainMat || it.Total != mainSvc+mainMat {
				return NewValidationError("Perhitungan nilai \"%s\" tidak konsisten. Muat ulang dokumen lalu coba lagi.", it.NameSnapshot)
			}
			grand += it.Total
			continue
		}
		expectSvc := mainSvc
		expectMat := mainMat
		for _, sb := range it.SubItems {
			expectSvc += lineTotal(sb.Quantity, sb.ServiceUnitPrice)
			expectMat += lineTotal(sb.Quantity, sb.MaterialUnitPrice)
		}
		if it.ServiceTotal != expectSvc || it.MaterialTotal != expectMat || it.Total != expectSvc+expectMat {
			return NewValidationError("Perhitungan nilai \"%s\" tidak konsisten. Muat ulang dokumen lalu coba lagi.", it.NameSnapshot)
		}
		grand += it.Total
	}
	if d.GrandTotal != grand {
		return NewValidationError("Grand total tidak sesuai penjumlahan baris. Muat ulang dokumen lalu coba lagi.")
	}
	return nil
}

// SetStatus menjalankan transisi status sesuai diagram BR-08; masuk REVIEW/FINAL divalidasi (BR-06).
// Dokumen dalam Room kolaborasi tidak boleh berubah status dari luar room.
func (s *SphService) SetStatus(id uint, target string) error {
	if s.docLocked(id) {
		return NewConflictError(errDocLockedFmt)
	}
	d, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil SPH", "id", id, "error", err)
		return fmt.Errorf("gagal mengubah status SPH")
	}

	allowed := map[string][]string{
		models.StatusDraft:     {models.StatusReview, models.StatusCancelled},
		models.StatusReview:    {models.StatusFinal, models.StatusCancelled},
		models.StatusFinal:     {models.StatusSent},
		models.StatusSent:      {models.StatusAccepted, models.StatusRejected},
		models.StatusAccepted:  {},
		models.StatusRejected:  {},
		models.StatusCancelled: {},
	}[d.Status]

	ok := false
	for _, a := range allowed {
		if a == target {
			ok = true
			break
		}
	}
	if !ok {
		return NewConflictError("Perubahan status dari %s ke %s tidak diizinkan.", statusLabel(d.Status), statusLabel(target))
	}

	enteringGate := target == models.StatusReview || target == models.StatusFinal
	if enteringGate {
		if err := s.validateForFinalization(d); err != nil {
			return err
		}
	}

	desc := fmt.Sprintf("Status SPH %s diubah menjadi %s", d.DocumentNumber, statusLabel(target))
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if target == models.StatusFinal {
			if err := s.repo.Finalize(tx, id, target, time.Now()); err != nil {
				return err
			}
		} else if err := s.repo.SetStatus(tx, id, target); err != nil {
			return err
		}
		action := "UPDATE"
		if target == models.StatusReview || target == models.StatusFinal {
			action = "FINALIZE"
		}
		return s.audit.Write(tx, action, "sph_document", id, desc)
	})
	if err != nil {
		if isFriendly(err) {
			return err
		}
		s.log.Error("gagal mengubah status SPH", "id", id, "target", target, "error", err)
		return fmt.Errorf("gagal mengubah status SPH")
	}
	return nil
}

// Duplicate menyalin penuh dokumen sebagai draft baru dengan nomor baru (BR-09).
func (s *SphService) Duplicate(id uint) (*SphDocumentView, error) {
	if s.docLocked(id) {
		return nil, NewConflictError(errDocLockedFmt)
	}
	src, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil SPH", "id", id, "error", err)
		return nil, fmt.Errorf("gagal menduplikasi SPH")
	}

	clone := &models.SphDocument{
		Date:             time.Now(),
		CustomerID:       src.CustomerID,
		VesselID:         src.VesselID,
		ProjectName:      src.ProjectName,
		Subject:          src.Subject,
		Reference:        src.Reference,
		Location:         src.Location,
		ValidUntil:       src.ValidUntil,
		PicName:          src.PicName,
		Status:           models.StatusDraft,
		Revision:         0,
		SubtotalService:  src.SubtotalService,
		SubtotalMaterial: src.SubtotalMaterial,
		GrandTotal:       src.GrandTotal,
		Notes:            src.Notes,
		Items:            cloneItems(src.Items),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		number, err := generateDocumentNumber(tx, clone.Date)
		if err != nil {
			return err
		}
		clone.DocumentNumber = number
		if err := s.repo.Create(tx, clone); err != nil {
			return err
		}
		return s.audit.Write(tx, "DUPLICATE", "sph_document", clone.ID,
			fmt.Sprintf("Diduplikat dari SPH %s Rev %d", src.DocumentNumber, src.Revision))
	})
	if err != nil {
		s.log.Error("gagal menduplikasi SPH", "id", id, "error", err)
		return nil, fmt.Errorf("gagal menduplikasi SPH")
	}
	views, err := toSphViews(s.db, []models.SphDocument{*clone})
	if err != nil {
		return nil, fmt.Errorf("gagal menduplikasi SPH")
	}
	return &views[0], nil
}

// CreateRevision membuat revisi baru (BR-10): nomor sama, revision+1, salinan penuh, status DRAFT.
func (s *SphService) CreateRevision(id uint) (*SphDocumentView, error) {
	if s.docLocked(id) {
		return nil, NewConflictError(errDocLockedFmt)
	}
	src, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		s.log.Error("gagal mengambil SPH", "id", id, "error", err)
		return nil, fmt.Errorf("gagal membuat revisi")
	}
	switch src.Status {
	case models.StatusFinal, models.StatusSent, models.StatusRejected:
	default:
		return nil, NewConflictError("Revisi hanya dibuat untuk dokumen yang sudah final/dikirim/ditolak.")
	}

	nextRev, err := s.repo.MaxRevision(s.db, src.DocumentNumber)
	if err != nil {
		s.log.Error("gagal menghitung revisi terakhir", "error", err)
		return nil, fmt.Errorf("gagal membuat revisi")
	}
	nextRev++

	clone := &models.SphDocument{
		DocumentNumber:   src.DocumentNumber,
		Revision:         nextRev,
		Date:             time.Now(),
		CustomerID:       src.CustomerID,
		VesselID:         src.VesselID,
		ProjectName:      src.ProjectName,
		Subject:          src.Subject,
		Reference:        src.Reference,
		Location:         src.Location,
		ValidUntil:       src.ValidUntil,
		PicName:          src.PicName,
		Status:           models.StatusDraft,
		SubtotalService:  src.SubtotalService,
		SubtotalMaterial: src.SubtotalMaterial,
		GrandTotal:       src.GrandTotal,
		Notes:            src.Notes,
		Items:            cloneItems(src.Items),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Create(tx, clone); err != nil {
			return err
		}
		hist := models.SphRevision{
			SphDocumentID:  clone.ID,
			FromDocumentID: &src.ID,
			RevisionNumber: nextRev,
			Note:           fmt.Sprintf("Revisi dari Rev %d", src.Revision),
		}
		if err := tx.Create(&hist).Error; err != nil {
			return err
		}
		return s.audit.Write(tx, "CREATE", "sph_document", clone.ID,
			fmt.Sprintf("Revisi %d dibuat dari SPH %s Rev %d", nextRev, src.DocumentNumber, src.Revision))
	})
	if err != nil {
		s.log.Error("gagal membuat revisi", "id", id, "error", err)
		return nil, fmt.Errorf("gagal membuat revisi")
	}
	views, err := toSphViews(s.db, []models.SphDocument{*clone})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat revisi")
	}
	return &views[0], nil
}

// Delete menghapus draft (soft). Dokumen final ke atas tidak dapat dihapus.
func (s *SphService) Delete(id uint) error {
	if s.docLocked(id) {
		return NewConflictError(errDocLockedFmt)
	}
	d, err := s.repo.GetByID(s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		s.log.Error("gagal mengambil SPH", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus SPH")
	}
	if d.Status != models.StatusDraft {
		return NewConflictError("Hanya dokumen berstatus Draft yang dapat dihapus.")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.SoftDelete(tx, id); err != nil {
			return err
		}
		return s.audit.Write(tx, "DELETE", "sph_document", id,
			fmt.Sprintf("SPH %s Rev %d dihapus", d.DocumentNumber, d.Revision))
	})
	if err != nil {
		s.log.Error("gagal menghapus SPH", "id", id, "error", err)
		return fmt.Errorf("gagal menghapus SPH")
	}
	return nil
}

// Stats merangkum angka dashboard (FR-U5).
func (s *SphService) Stats() (*DashboardStats, error) {
	rows, err := s.repo.StatsByStatus(s.db)
	if err != nil {
		s.log.Error("gagal menghitung statistik SPH", "error", err)
		return nil, fmt.Errorf("gagal memuat statistik")
	}
	st := &DashboardStats{}
	for _, r := range rows {
		st.TotalSph += r.Count
		switch r.Status {
		case models.StatusDraft, models.StatusReview:
			st.DraftCount += r.Count
		case models.StatusFinal, models.StatusSent:
			st.FinalCount += r.Count
		case models.StatusAccepted:
			st.AcceptedCount += r.Count
		}
	}
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, -1).Format("2006-01-02")
	mv, err := s.repo.SumGrandTotalBetween(s.db, from.Format("2006-01-02"), to)
	if err != nil {
		s.log.Error("gagal menjumlahkan nilai bulan ini", "error", err)
		return nil, fmt.Errorf("gagal memuat statistik")
	}
	st.MonthValue = mv

	recent, err := s.List("", "", 5)
	if err != nil {
		return nil, err
	}
	st.Recent = recent
	return st, nil
}

// ===== helper =====

// cloneItems menyalin penuh item & sub item untuk duplicate/revision (BR-09, BR-10).
func cloneItems(src []models.SphItem) []models.SphItem {
	items := make([]models.SphItem, 0, len(src))
	for _, it := range src {
		c := it
		c.ID = 0
		c.SphDocumentID = 0
		subs := make([]models.SphSubItem, 0, len(it.SubItems))
		for _, sb := range it.SubItems {
			sb2 := sb
			sb2.ID = 0
			sb2.SphItemID = 0
			subs = append(subs, sb2)
		}
		c.SubItems = subs
		items = append(items, c)
	}
	return items
}

// isFriendly melaporkan error validasi/konflik agar tidak dibungkus generik.
func isFriendly(err error) bool {
	switch err.(type) {
	case *ValidationError, *ConflictError:
		return true
	}
	return false
}

// statusLabel menerjemahkan kode status ke label ramah (BR-15).
func statusLabel(status string) string {
	switch status {
	case models.StatusDraft:
		return "Draft"
	case models.StatusReview:
		return "Review"
	case models.StatusFinal:
		return "Final"
	case models.StatusSent:
		return "Terkirim"
	case models.StatusAccepted:
		return "Disetujui"
	case models.StatusRejected:
		return "Ditolak"
	case models.StatusCancelled:
		return "Dibatalkan"
	default:
		return status
	}
}
