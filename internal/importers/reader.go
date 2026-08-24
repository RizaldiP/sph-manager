package importers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
)

// Reader workbook: mendukung XLSX (excelize) dan XLS lama/BIFF (xlsReader).

var errUnsupportedFormat = fmt.Errorf("format file tidak didukung; gunakan .xls atau .xlsx")

func isXLS(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".xls")
}

// ListSheets mengembalikan daftar nama sheet sesuai urutan di workbook.
func ListSheets(path string) ([]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("file tidak dapat dibaca")
	}
	if isXLS(path) {
		wb, err := xls.OpenFile(path)
		if err != nil {
			return nil, fmt.Errorf("gagal membuka file Excel")
		}
		names := make([]string, 0, wb.GetNumberSheets())
		for i := 0; i < wb.GetNumberSheets(); i++ {
			sh, err := wb.GetSheet(i)
			if err != nil {
				continue
			}
			names = append(names, sh.GetName())
		}
		return names, nil
	}
	if strings.EqualFold(filepath.Ext(path), ".xlsx") {
		f, err := excelize.OpenFile(path)
		if err != nil {
			return nil, fmt.Errorf("gagal membuka file Excel")
		}
		defer f.Close()
		return f.GetSheetList(), nil
	}
	return nil, errUnsupportedFormat
}

// ReadSheet membaca satu sheet menjadi Grid string.
func ReadSheet(path, sheet string) (*Grid, error) {
	if isXLS(path) {
		return readXLS(path, sheet)
	}
	if strings.EqualFold(filepath.Ext(path), ".xlsx") {
		return readXLSX(path, sheet)
	}
	return nil, errUnsupportedFormat
}

func readXLS(path, sheet string) (*Grid, error) {
	wb, err := xls.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file Excel")
	}
	var target *xls.Sheet
	for i := 0; i < wb.GetNumberSheets(); i++ {
		sh, gerr := wb.GetSheet(i)
		if gerr != nil {
			continue
		}
		if sh.GetName() == sheet {
			target = sh
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("sheet \"%s\" tidak ditemukan", sheet)
	}

	rawRows := target.GetRows()
	width := 0
	cells := make([][]string, len(rawRows))
	for ri, row := range rawRows {
		cols := row.GetCols()
		if len(cols) > width {
			width = len(cols)
		}
		cells[ri] = make([]string, len(cols))
		for ci, c := range cols {
			cells[ri][ci] = strings.TrimSpace(c.GetString())
		}
	}
	g := &Grid{Rows: cells}
	// Pad semua baris agar persegi.
	for i := range g.Rows {
		if len(g.Rows[i]) < width {
			pad := make([]string, width-len(g.Rows[i]))
			g.Rows[i] = append(g.Rows[i], pad...)
		}
	}
	return g, nil
}

func readXLSX(path, sheet string) (*Grid, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file Excel")
	}
	defer f.Close()
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("sheet \"%s\" tidak ditemukan", sheet)
	}
	width := 0
	for _, r := range rows {
		if len(r) > width {
			width = len(r)
		}
	}
	g := &Grid{Rows: make([][]string, len(rows))}
	for i, r := range rows {
		row := make([]string, width)
		copy(row, r)
		for j := range row {
			row[j] = strings.TrimSpace(row[j])
		}
		g.Rows[i] = row
	}
	return g, nil
}
