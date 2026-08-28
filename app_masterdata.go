package main

import (
	"time"

	"github.com/RizaldiP/sph-manager/internal/collaboration"
	"github.com/RizaldiP/sph-manager/internal/masterdata"
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// gormMasterDataStore adalah implementasi collaboration.MasterDataStore yang
// meneruskan ke masterdata.Service (database lokal), serupa gormChatStore.
type gormMasterDataStore struct {
	svc *masterdata.Service
}

func newGormMasterDataStore(svc *masterdata.Service) *gormMasterDataStore {
	return &gormMasterDataStore{svc: svc}
}

func (s *gormMasterDataStore) SaveInbox(m *models.MasterInbox) error { return s.svc.SaveInbox(m) }
func (s *gormMasterDataStore) SetInboxStatus(packageID, status string, at time.Time) error {
	return s.svc.SetInboxStatus(packageID, status, at)
}
func (s *gormMasterDataStore) SaveSent(m *models.MasterSent) error { return s.svc.SaveSent(m) }
func (s *gormMasterDataStore) SetSentStatus(packageID, status string) error {
	return s.svc.UpdateSentStatus(packageID, status)
}

// BuildMasterDataPackage menyusun package Master Data dari seluruh data lokal
// (kategori, pekerjaan, sub-pekerjaan, material) untuk dikirim ke member room.
func (a *App) BuildMasterDataPackage() (*collaboration.MasterDataPackage, error) {
	senderID, senderName := a.collabMgr.CurrentIdentity()
	roomID := a.currentRoomID()
	return a.masterSvc.BuildPackage(senderID, senderName, roomID)
}

// SendMasterData mengirim package Master Data ke target (participant IDs).
func (a *App) SendMasterData(pkg *collaboration.MasterDataPackage, targets []string) error {
	return a.collabMgr.SendMasterData(pkg, targets)
}

// ListMasterInbox menampilkan daftar paket Master Data yang diterima.
func (a *App) ListMasterInbox() ([]masterdata.InboxItem, error) {
	return a.masterSvc.InboxList()
}

// GetMasterInbox menampilkan satu paket masuk.
func (a *App) GetMasterInbox(packageID string) (*masterdata.InboxItem, error) {
	return a.masterSvc.InboxGet(packageID)
}

// GetMasterInboxPayload memuat ulang package mentah dari inbox.
func (a *App) GetMasterInboxPayload(packageID string) (*collaboration.MasterDataPackage, error) {
	return a.masterSvc.GetInboxPayload(packageID)
}

// PreviewMasterData membandingkan package masuk dengan data lokal (pratinjau).
func (a *App) PreviewMasterData(packageID string) ([]masterdata.DiffItem, error) {
	pkg, err := a.masterSvc.GetInboxPayload(packageID)
	if err != nil {
		return nil, err
	}
	return a.masterSvc.Compare(pkg)
}

// InstallMasterData memasang paket masuk ke database lokal, lalu melaporkan
// status INSTALLED ke pengirim. strategy: PROMPT|USE_LOCAL|USE_INCOMING|SKIP.
func (a *App) InstallMasterData(packageID, strategy string, decisions map[string]string) (*masterdata.InstallSummary, error) {
	if packageID == "" {
		return nil, services.NewValidationError("Package ID wajib diisi.")
	}
	pkg, err := a.masterSvc.GetInboxPayload(packageID)
	if err != nil {
		return nil, err
	}
	sum, err := a.masterSvc.Install(pkg, strategy, decisions)
	if err != nil {
		return nil, err
	}
	_ = a.masterSvc.SetInboxStatus(packageID, models.MasterStatusInstalled, time.Now().UTC())
	_ = a.collabMgr.AcknowledgeMasterData(packageID, collaboration.MasterStatusAckInstalled, "")
	return sum, nil
}

// RejectMasterData menandai paket masuk sebagai ditolak dan melaporkannya ke pengirim.
func (a *App) RejectMasterData(packageID string) error {
	if packageID == "" {
		return services.NewValidationError("Package ID wajib diisi.")
	}
	if err := a.masterSvc.SetInboxStatus(packageID, models.MasterStatusRejected, time.Now().UTC()); err != nil {
		return err
	}
	return a.collabMgr.AcknowledgeMasterData(packageID, collaboration.MasterStatusAckRejected, "")
}

// MarkMasterInboxViewed menandai paket masuk telah dibuka pengguna (VIEWED).
func (a *App) MarkMasterInboxViewed(packageID string) error {
	return a.masterSvc.SetInboxStatus(packageID, models.MasterStatusViewed, time.Now())
}

// ListMasterSent menampilkan daftar paket Master Data yang telah dikirim PC ini.
func (a *App) ListMasterSent() ([]masterdata.SentItem, error) {
	return a.masterSvc.SentList()
}

// currentRoomID mengembalikan RoomID sesi aktif (host atau client).
func (a *App) currentRoomID() string {
	return a.collabMgr.CurrentRoomID()
}
