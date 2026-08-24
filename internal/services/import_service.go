package services

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/importers"
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/repositories"
)

// ImportResult: ringkasan hasil import master pekerjaan (FR-IE1..IE4).
type ImportResult struct {
	ItemsCreated int `json:"itemsCreated"`
	SubsCreated  int `json:"subsCreated"`
	Skipped      int `json:"skipped"`
}

// ImportService: import transaksional hasil parse Excel ke Master Pekerjaan.
// Seluruh operasi berjalan dalam SATU transaksi; gagal di tengah = rollback penuh
// (BR-16, FR-A4). Progress dipancarkan per baris untuk UI (FR-IE4).
type ImportService struct {
	db       *gorm.DB
	log      *slog.Logger
	items    *repositories.WorkItemRepository
	subs     *repositories.WorkSubItemRepository
	cats     *repositories.CategoryRepository
	progress func(done, total int) error
}

func NewImportService(db *gorm.DB, log *slog.Logger) *ImportService {
	return &ImportService{
		db:    db,
		log:   log,
		items: repositories.NewWorkItemRepository(),
		subs:  repositories.NewWorkSubItemRepository(),
		cats:  repositories.NewCategoryRepository(),
	}
}

// SetProgress memasang callback progres; callback boleh mengembalikan error
// untuk membatalkan import (rollback).
func (s *ImportService) SetProgress(fn func(done, total int) error) { s.progress = fn }

// ConfirmRow: keputusan akhir pengguna atas satu baris pratinjau.
type ConfirmRow struct {
	RowIndex int    `json:"rowIndex"`
	Level    string `json:"level"` // main | sub | skip
}

// ValidateRows menilai barisan pratinjau terhadap mapping + level final.
// Aturan: baris ber-saran unknown yang belum diklasifikasi (atau dilewati
// tanpa keputusan eksplisit) memblokir; baris terkonfirmasi yang masih
// bermasalah juga memblokir. Kosong berarti aman diimport.
func (s *ImportService) ValidateRows(rows []importers.PreviewRow, confirms []ConfirmRow) []string {
	levelOf := map[int]string{}
	for _, c := range confirms {
		levelOf[c.RowIndex] = c.Level
	}
	var blockers []string
	for i := range rows {
		r := rows[i]
		lvl, confirmed := levelOf[r.RowIndex]
		if r.Suggested == importers.LevelUnknown {
			if !confirmed {
				blockers = append(blockers, fmt.Sprintf("Baris %d: klasifikasi belum ditentukan.", r.RowIndex+1))
			}
			if !confirmed || lvl == "skip" {
				continue
			}
		} else if !confirmed {
			continue // main/sub tanpa konfirmasi berarti sengaja tidak diimport.
		}
		for _, e := range r.Errors {
			blockers = append(blockers, fmt.Sprintf("Baris %d: %s", r.RowIndex+1, e))
		}
		if strings.TrimSpace(r.Name) == "" && lvl == importers.LevelMain {
			blockers = append(blockers, fmt.Sprintf("Baris %d: nama pekerjaan kosong.", r.RowIndex+1))
		}
	}
	return blockers
}

// ImportWorkItems menjalankan import penuh: baca ulang file → parse → terapkan
// keputusan pengguna → tulis ke database dalam satu transaksi.
func (s *ImportService) ImportWorkItems(
	categoryID uint,
	path, sheet string,
	m importers.ColumnMapping,
	confirms []ConfirmRow,
) (*ImportResult, error) {
	if categoryID == 0 {
		return nil, NewValidationError("Kategori tujuan wajib dipilih.")
	}
	ok, err := s.items.CategoryExists(s.db, categoryID)
	if err != nil {
		s.log.Error("gagal memeriksa kategori", "error", err)
		return nil, fmt.Errorf("gagal menjalankan import")
	}
	if !ok {
		return nil, ErrNotFound
	}

	grid, err := importers.ReadSheet(path, sheet)
	if err != nil {
		return nil, err
	}
	rows := importers.ParseRows(grid, m)
	blockers := s.ValidateRows(rows, confirms)
	if len(blockers) > 0 {
		return nil, NewConflictError("Import diblokir: %d masalah belum diselesaikan. Periksa pratinjau.", len(blockers))
	}

	// Susun urutan eksekusi sesuai posisi baris asli.
	levelOf := map[int]string{}
	order := make([]int, 0, len(confirms))
	for _, c := range confirms {
		if c.Level == importers.LevelMain || c.Level == importers.LevelSub {
			levelOf[c.RowIndex] = c.Level
			order = append(order, c.RowIndex)
		}
	}
	for i := 0; i < len(order); i++ {
		for j := i + 1; j < len(order); j++ {
			if order[j] < order[i] {
				order[i], order[j] = order[j], order[i]
			}
		}
	}

	res := &ImportResult{Skipped: len(rows) - len(order)}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		seqBase, err := s.items.MaxSequenceInCategory(tx, categoryID)
		if err != nil {
			return err
		}
		seq := seqBase

		var currentItem uint
		done := 0
		total := len(order)
		for _, ri := range order {
			row := findRow(rows, ri)
			if row == nil {
				continue
			}
			switch levelOf[ri] {
			case importers.LevelMain:
				code, err := generateCode(tx, "work_items", "PEK-")
				if err != nil {
					return err
				}
				w := &models.WorkItem{
					CategoryID:           categoryID,
					Code:                 code,
					Name:                 row.Name,
					DefaultUnit:          row.Unit,
					DefaultQuantity:      orOne(row.Qty),
					DefaultServicePrice:  row.ServicePrice,
					DefaultMaterialPrice: row.MaterialPrice,
					Sequence:             seq + 1,
					IsActive:             true,
				}
				if err := s.items.Create(tx, w); err != nil {
					return err
				}
				currentItem = w.ID
				seq++
				res.ItemsCreated++
				desc := fmt.Sprintf("Import Excel: pekerjaan \"%s\" (%s)", w.Name, code)
				if err := (&AuditWriter{}).Write(tx, "CREATE", "work_item", w.ID, desc); err != nil {
					return err
				}
			case importers.LevelSub:
				if currentItem == 0 {
					return fmt.Errorf("baris %d adalah sub-pekerjaan tanpa pekerjaan induk; klasifikasikan sebagai pekerjaan atau letakkan setelah induknya", ri+1)
				}
				code, err := generateCode(tx, "work_sub_items", "SUB-")
				if err != nil {
					return err
				}
				subSeq, err := s.subs.MaxSequenceInWorkItem(tx, currentItem)
				if err != nil {
					return err
				}
				sub := &models.WorkSubItem{
					WorkItemID:           currentItem,
					Code:                 code,
					Name:                 row.Name,
					DefaultUnit:          row.Unit,
					DefaultQuantity:      orOne(row.Qty),
					DefaultServicePrice:  row.ServicePrice,
					DefaultMaterialPrice: row.MaterialPrice,
					Sequence:             subSeq + 1,
					IsActive:             true,
				}
				if err := s.subs.Create(tx, sub); err != nil {
					return err
				}
				res.SubsCreated++
				if err := (&AuditWriter{}).Write(tx, "CREATE", "work_sub_item", sub.ID,
					fmt.Sprintf("Import Excel: sub-pekerjaan \"%s\" (%s)", sub.Name, code)); err != nil {
					return err
				}
			}
			done++
			if s.progress != nil {
				if perr := s.progress(done, total); perr != nil {
					return perr
				}
			}
		}
		return nil
	})
	if err != nil {
		if isFriendly(err) || strings.Contains(err.Error(), "sub-pekerjaan tanpa") {
			return nil, err
		}
		s.log.Error("gagal import excel", "error", err)
		return nil, fmt.Errorf("gagal menjalankan import; seluruh perubahan dibatalkan")
	}
	return res, nil
}

func findRow(rows []importers.PreviewRow, index int) *importers.PreviewRow {
	for i := range rows {
		if rows[i].RowIndex == index {
			return &rows[i]
		}
	}
	return nil
}

func orOne(q float64) float64 {
	if q > 0 {
		return q
	}
	return 1
}
