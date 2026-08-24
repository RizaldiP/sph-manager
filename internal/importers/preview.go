package importers

// SheetPreview: data untuk langkah pilih sheet & mapping kolom di UI.
type SheetPreview struct {
	Grid             [][]string    `json:"grid"`
	TotalRows        int           `json:"totalRows"`
	TotalCols        int           `json:"totalCols"`
	SuggestedMapping ColumnMapping `json:"suggestedMapping"`
	Notes            []string      `json:"notes,omitempty"`
	MainCount        int           `json:"mainCount"`
	SubCount         int           `json:"subCount"`
	UnknownCount     int           `json:"unknownCount"`
}

const (
	previewMaxRows = 200
	previewMaxCols = 20
)

// BuildSheetPreview membaca sheet lalu menyusun pratinjau + saran mapping.
func BuildSheetPreview(path, sheet string) (*SheetPreview, error) {
	g, err := ReadSheet(path, sheet)
	if err != nil {
		return nil, err
	}
	mapping, notes := SuggestMapping(g)
	mainCount, subCount, unknownCount := CountLevels(ParseRows(g, mapping))

	cols := g.Width()
	if cols > previewMaxCols {
		cols = previewMaxCols
	}
	rows := g.Height()
	if rows > previewMaxRows {
		rows = previewMaxRows
		notes = append(notes, "Pratinjau dibatasi 200 baris pertama; parsing tetap memproses seluruh baris.")
	}
	grid := make([][]string, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]string, cols)
		for c := 0; c < cols; c++ {
			grid[r][c] = g.Cell(r, c)
		}
	}
	return &SheetPreview{
		Grid:             grid,
		TotalRows:        g.Height(),
		TotalCols:        g.Width(),
		SuggestedMapping: mapping,
		Notes:            notes,
		MainCount:        mainCount,
		SubCount:         subCount,
		UnknownCount:     unknownCount,
	}, nil
}
