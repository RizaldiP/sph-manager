package services

import (
	"fmt"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// allocateWeightedSubs memvalidasi bobot lalu mengalokasikan pool jasa &
// material main point ke snapshot sub point (BR-02, BR-03, BR-04).
// Harga satuan sub point tidak dipakai pada mode ini; nilai tersimpan adalah
// hasil alokasi. Alokasi tetap dilakukan walau Σ bobot ≠ 100 (proporsional
// terhadap Σ aktual) — penolakan Σ ≠ 100 terjadi saat finalisasi.
func allocateWeightedSubs(inputs []SphSubItemInput, svcPool, matPool int64) ([]models.SphSubItem, error) {
	if len(inputs) == 0 {
		return nil, NewValidationError("Mode pembobotan butuh minimal satu sub point.")
	}
	weights := make([]int, len(inputs))
	for j := range inputs {
		label := subLabel(inputs[j].Name, j)
		w := inputs[j].Weight
		if w < 1 || w > 100 {
			return nil, NewValidationError("Bobot %s harus di antara 1 dan 100.", label)
		}
		if inputs[j].Quantity <= 0 {
			return nil, NewValidationError("Qty sub point %s harus lebih besar dari 0.", label)
		}
		weights[j] = w
	}
	svcShares := allocateLargestRemainder(svcPool, weights)
	matShares := allocateLargestRemainder(matPool, weights)
	subs := make([]models.SphSubItem, len(inputs))
	for j := range inputs {
		in := &inputs[j]
		subs[j] = models.SphSubItem{
			NameSnapshot:        trim(in.Name),
			DescriptionSnapshot: trim(in.Description),
			Quantity:            in.Quantity,
			Unit:                trim(in.Unit),
			Weight:              in.Weight,
			AllocatedValue:      svcShares[j] + matShares[j],
			ServiceTotal:        svcShares[j],
			MaterialTotal:       matShares[j],
			Total:               svcShares[j] + matShares[j],
			Notes:               trim(in.Notes),
		}
	}
	return subs, nil
}

func subLabel(name string, idx int) string {
	if n := trim(name); n != "" {
		return fmt.Sprintf("\"%s\"", n)
	}
	return fmt.Sprintf("baris %d", idx+1)
}

// allocateLargestRemainder membagi pool uang (integer Rupiah) ke tiap bobot
// dengan metode largest remainder (BR-04): dasar = floor(pool×w/Σw), sisa sen
// diberikan satu per satu ke pecahan terbesar; tie-break urutan indeks
// (sequence baris) agar hasil deterministik. Σ hasil selalu tepat = pool.
func allocateLargestRemainder(pool int64, weights []int) []int64 {
	out := make([]int64, len(weights))
	if len(weights) == 0 {
		return out
	}
	var sumW int64
	for _, w := range weights {
		sumW += int64(w)
	}
	if sumW <= 0 {
		return out
	}
	if pool == 0 {
		for i := range out {
			out[i] = 0
		}
		return out
	}
	// pool×w aman untuk int64 pada rentang nilai SPH realistis
	// (pool ≤ ~9e16 dan w ≤ 100 → produk ≤ 9e18 < 2^63).
	type remainder struct {
		idx  int
		frac int64
	}
	var base []int64
	remains := make([]remainder, 0, len(weights))
	var allocated int64
	for i, w := range weights {
		num := pool * int64(w)
		b := num / sumW
		base = append(base, b)
		allocated += b
		remains = append(remains, remainder{idx: i, frac: num % sumW})
	}
	leftover := pool - allocated
	// insertion sort desc by frac, tie-break idx asc (stabil & deterministik).
	for i := 1; i < len(remains); i++ {
		for j := i; j > 0 && (remains[j].frac > remains[j-1].frac ||
			(remains[j].frac == remains[j-1].frac && remains[j].idx < remains[j-1].idx)); j-- {
			remains[j], remains[j-1] = remains[j-1], remains[j]
		}
	}
	for i := range out {
		out[i] = base[i]
	}
	for k := 0; k < len(remains) && leftover > 0; k++ {
		out[remains[k].idx]++
		leftover--
	}
	return out
}

// weightSum menjumlahkan bobot; dipakai validasi & pesan selisih.
func weightSum(weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	return total
}
