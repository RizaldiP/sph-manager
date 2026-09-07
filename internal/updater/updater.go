package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// driveDownloadURL membentuk URL unduhan langsung untuk sebuah file ID.
func driveDownloadURL(fileID string) string {
	return "https://drive.usercontent.google.com/download?id=" + url.QueryEscape(fileID) + "&export=download&confirm=t"
}

// ResolveDriveLink mengubah link share Google Drive menjadi URL unduhan
// langsung. Mendukung format:
//   - https://drive.google.com/file/d/<ID>/view?usp=sharing
//   - https://drive.google.com/uc?export=download&id=<ID>
//   - https://drive.usercontent.google.com/download?id=<ID>...
//
// URL non-Drive dikembalikan apa adanya sehingga feed dapat memakai link
// dari server mana pun.
func ResolveDriveLink(link string) (string, error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return "", fmt.Errorf("link kosong")
	}
	u, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("link tidak valid")
	}
	host := strings.ToLower(u.Host)
	if host != "drive.google.com" && host != "drive.usercontent.google.com" {
		return link, nil
	}

	if strings.HasPrefix(u.Path, "/file/d/") {
		parts := strings.Split(strings.TrimPrefix(u.Path, "/file/d/"), "/")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return driveDownloadURL(parts[0]), nil
		}
	}
	if id := strings.TrimSpace(u.Query().Get("id")); id != "" {
		return driveDownloadURL(id), nil
	}
	return "", fmt.Errorf("tidak dapat menemukan ID file pada link Google Drive")
}

// Version membandingkan dua versi semver-style (major.minor.patch). Bagian
// yang tidak ada diperlakukan sebagai nol. Mengembalikan negatif bila a < b,
// nol bila sama, positif bila a > b.
type Version [3]int

// ParseVersion mengurai string "x.y.z" yang juga menerima "x" dan "x.y";
// komponen yang tidak disebut dianggap 0.
func ParseVersion(v string) (Version, error) {
	var ver Version
	parts := strings.Split(strings.TrimSpace(v), ".")
	if len(parts) > 3 {
		return ver, fmt.Errorf("format versi tidak valid")
	}
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return ver, fmt.Errorf("format versi tidak valid")
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return ver, fmt.Errorf("format versi tidak valid")
		}
		ver[i] = n
	}
	return ver, nil
}

// CompareVersions membandingkan dua string versi; mengembalikan -1, 0, atau 1.
func CompareVersions(a, b string) (int, error) {
	va, err := ParseVersion(a)
	if err != nil {
		return 0, err
	}
	vb, err := ParseVersion(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < 3; i++ {
		if va[i] < vb[i] {
			return -1, nil
		}
		if va[i] > vb[i] {
			return 1, nil
		}
	}
	return 0, nil
}

// Metadata update (trailer) yang disisipkan ke ekor file instalasi saat rilis.
const (
	metaBegin = "SPHMANAGER-META-BEGIN"
	metaEnd   = "SPHMANAGER-META-END"
)

// ExeMeta adalah metadata versi yang dibaca dari trailer file kandidat.
type ExeMeta struct {
	Version string `json:"version"`
	Notes   string `json:"notes,omitempty"`
}

// appendMetaBlock menulis blok trailer ke w (termasuk pembatas baris).
func appendMetaBlock(w io.Writer, meta ExeMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	block := "\n" + metaBegin + "\n" + string(data) + "\n" + metaEnd + "\n"
	if _, err := io.WriteString(w, block); err != nil {
		return fmt.Errorf("gagal menyisipkan metadata update")
	}
	return nil
}

// AppendMeta menyisipkan metadata versi ke ekor file kandidat (dipakai skrip
// rilis). Data mengikuti data biner sehingga file tetap dapat dijalankan.
func AppendMeta(path string, meta ExeMeta) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("tidak dapat membuka file untuk metadata update")
	}
	defer f.Close()
	return appendMetaBlock(f, meta)
}

// ReadExeMeta membaca metadata versi dari trailer file kandidat update.
// Mengembalikan error bila penanda tidak ditemukan atau isinya rusak.
func ReadExeMeta(path string) (*ExeMeta, error) {
	const maxTail = 4096
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("tidak dapat membaca file kandidat update")
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("tidak dapat membaca file kandidat update")
	}
	readLen := stat.Size()
	offset := int64(0)
	if readLen > maxTail {
		readLen = maxTail
		offset = stat.Size() - maxTail
	}
	buf := make([]byte, readLen)
	if _, err := f.ReadAt(buf, offset); err != nil && err != io.EOF {
		return nil, fmt.Errorf("tidak dapat membaca file kandidat update")
	}
	tail := string(buf)
	i := strings.Index(tail, metaBegin)
	if i < 0 {
		return nil, fmt.Errorf("bukan berkas update yang valid (metadata tidak ditemukan)")
	}
	rest := tail[i+len(metaBegin):]
	j := strings.Index(rest, metaEnd)
	if j < 0 {
		return nil, fmt.Errorf("metadata update rusak")
	}
	raw := strings.TrimSpace(rest[:j])
	var meta ExeMeta
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil, fmt.Errorf("metadata update rusak")
	}
	if strings.TrimSpace(meta.Version) == "" {
		return nil, fmt.Errorf("metadata update tidak memuat versi")
	}
	return &meta, nil
}

// FetchCandidate mengunduh calon versi terbaru ke dest lalu membaca versinya.
// Memakai client dengan timeout tersendiri karena file kandidat cukup besar.
func FetchCandidate(rawURL, dest string, timeout time.Duration, progress ProgressFunc) (*ExeMeta, error) {
	if err := DownloadFileTimeout(rawURL, dest, timeout, progress); err != nil {
		return nil, err
	}
	meta, err := ReadExeMeta(dest)
	if err != nil {
		_ = os.Remove(dest)
		return nil, err
	}
	return meta, nil
}