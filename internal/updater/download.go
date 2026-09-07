package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ProgressFunc menerima jumlah byte terunduh dan total (0 bila tidak diketahui).
type ProgressFunc func(done, total int64)

// countingWriter meneruskan byte ke writer sambil melaporkan progress.
type countingWriter struct {
	w        io.Writer
	done     int64
	total    int64
	progress ProgressFunc
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.done += int64(n)
	if c.progress != nil {
		c.progress(c.done, c.total)
	}
	return n, err
}

// DownloadFile mengunduh url (bisa link share Google Drive) ke path sementara
// sambil melaporkan progress via callback.
func DownloadFile(rawURL, dest string, progress ProgressFunc) error {
	return DownloadFileTimeout(rawURL, dest, 60*time.Second, progress)
}

// DownloadFileTimeout sama seperti DownloadFile namun memakai client dengan
// batas waktu tersendiri (berguna untuk file kandidat yang besar).
func DownloadFileTimeout(rawURL, dest string, timeout time.Duration, progress ProgressFunc) error {
	url, err := ResolveDriveLink(rawURL)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("tidak dapat terhubung ke server unduhan")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server mengembalikan status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("tidak dapat menyiapkan file sementara")
	}
	defer out.Close()

	pw := &countingWriter{w: out, total: resp.ContentLength, progress: progress}
	if _, err := io.CopyBuffer(pw, resp.Body, make([]byte, 64*1024)); err != nil {
		return fmt.Errorf("unduhan terputus")
	}
	if err := out.Sync(); err != nil {
		return fmt.Errorf("gagal menyimpan file unduhan")
	}
	return nil
}