package services

import "strings"

var (
	angkaKata = []string{"", "Satu", "Dua", "Tiga", "Empat", "Lima", "Enam", "Tujuh", "Delapan", "Sembilan", "Sepuluh", "Sebelas"}
	skala     = []struct {
		nilai int64
		kata  string
	}{
		{1_000_000_000_000, "Triliun"},
		{1_000_000_000, "Miliar"},
		{1_000_000, "Juta"},
		{1_000, "Ribu"},
	}
)

// angkaKeKata mengonversi 0..999 menjadi kalimat Indonesia.
func angkaKeKata(n int64) string {
	switch {
	case n < 0 || n > 999:
		return ""
	case n <= 11:
		return angkaKata[n]
	case n < 20:
		return angkaKata[n-10] + " Belas"
	case n < 100:
		puluhan := n / 10
		sisa := n % 10
		out := angkaKata[puluhan] + " Puluh"
		if sisa > 0 {
			out += " " + angkaKata[sisa]
		}
		return out
	default:
		ratusan := n / 100
		sisa := n % 100
		kata := angkaKata[ratusan] + " Ratus"
		if ratusan == 1 {
			kata = "Seratus"
		}
		if sisa > 0 {
			kata += " " + angkaKeKata(sisa)
		}
		return kata
	}
}

// Terbilang mengubah jumlah uang integer Rupiah menjadi kalimat Bahasa Indonesia
// dengan gaya kapital tiap kata, mis. 167300000 → "Seratus Enam Puluh Tujuh Juta Tiga Ratus Ribu Rupiah" (BR-11).
func Terbilang(total int64) string {
	if total < 0 {
		return "Minus " + Terbilang(-total)
	}
	if total == 0 {
		return "Nol Rupiah"
	}

	var bagian []string
	sisa := total
	for _, s := range skala {
		if sisa >= s.nilai {
			n := sisa / s.nilai
			sisa %= s.nilai
			kata := angkaKeKata(n)
			if n == 1 && s.kata == "Ribu" {
				kata = "Se"
			}
			bagian = append(bagian, kata+" "+s.kata)
		}
	}
	if sisa > 0 {
		bagian = append(bagian, angkaKeKata(sisa))
	}

	gabung := strings.TrimSpace(strings.Join(bagian, " "))
	for strings.Contains(gabung, "Se Ribu") {
		gabung = strings.ReplaceAll(gabung, "Se Ribu", "Seribu")
	}
	return gabung + " Rupiah"
}
