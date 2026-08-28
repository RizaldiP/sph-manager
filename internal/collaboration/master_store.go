package collaboration

import (
	"time"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// MasterDataStore menyimpan & memperbarui status Master Data di database lokal
// (inbox untuk yang diterima, sent untuk yang dikirim PC ini). Diimplementasikan
// app memakai database lokal (SQLite), serupa dengan ChatStore.
type MasterDataStore interface {
	// SaveInbox menyimpan paket masuk (dedup by PackageID).
	SaveInbox(m *models.MasterInbox) error
	// SetInboxStatus memperbarui status paket masuk (VIEWED/INSTALLED/REJECTED/FAILED).
	SetInboxStatus(packageID, status string, at time.Time) error
	// SaveSent mencatat paket yang dikirim PC ini (dedup by PackageID).
	SaveSent(m *models.MasterSent) error
	// SetSentStatus memperbarui status pengiriman tingkat paket.
	SetSentStatus(packageID, status string) error
}
