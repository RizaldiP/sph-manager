package importers

// Grid adalah potret nilai sel sebuah sheet: baris & kolom 0-based,
// semua sel dinormalkan menjadi string (sel kosong = "").
type Grid struct {
	Rows [][]string
}

func (g *Grid) Height() int { return len(g.Rows) }

func (g *Grid) Width() int {
	w := 0
	for _, r := range g.Rows {
		if len(r) > w {
			w = len(r)
		}
	}
	return w
}

// Cell mengambil nilai satu sel dengan aman ("" bila di luar grid).
func (g *Grid) Cell(row, col int) string {
	if row < 0 || row >= len(g.Rows) || col < 0 || col >= len(g.Rows[row]) {
		return ""
	}
	return g.Rows[row][col]
}
