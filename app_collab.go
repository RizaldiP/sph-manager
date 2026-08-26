package main

import (
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/RizaldiP/sph-manager/internal/collaboration"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// Binding Work Together (Phase 10, docs/collaboration-lan.md §10.32).
// Seluruh perubahan sesi disiarkan ke frontend lewat event `collab:sync`
// berisi collaboration.UISnapshot.

// GetCollabDefaults mengembalikan nilai awal untuk dialog Create/Join room.
func (a *App) GetCollabDefaults() services.CollabDefaults {
	device, _ := os.Hostname()
	if strings.TrimSpace(device) == "" {
		device = "PC"
	}
	return services.CollabDefaults{
		DeviceName:  device,
		Port:        a.settings.CollabPortOrDefault(),
		DisplayName: a.settings.CollabDisplayNameOrDefault(),
	}
}

// CreateCollabRoom membuat room host untuk draft SPH (§10.21).
func (a *App) CreateCollabRoom(sphDocumentID uint, roomName, displayName string) (*collaboration.RoomInfo, error) {
	a.log.Info("CreateCollabRoom dipanggil", "docID", sphDocumentID, "roomName", roomName)
	info, err := a.collabMgr.HostRoom(sphDocumentID, roomName, displayName, a.settings.CollabPortOrDefault())
	if err != nil {
		a.log.Warn("CreateCollabRoom gagal", "error", err)
	} else {
		a.log.Info("CreateCollabRoom berhasil", "roomID", info.RoomID, "port", info.Port, "hostIPs", info.HostIPs)
	}
	return info, err
}

// CloseCollabRoom menutup room yang sedang di-host (§10.27).
func (a *App) CloseCollabRoom() error { return a.collabMgr.CloseHostedRoom("Ditutup oleh host.") }

// StartDiscoveryListener mulai mendengarkan broadcast discovery untuk lobby.
func (a *App) StartDiscoveryListener() error { return a.collabMgr.StartDiscovery() }

func (a *App) StopDiscoveryListener() { a.collabMgr.StopDiscovery() }

// ListDiscoveredRooms mengembalikan room LAN yang masih hidup (§10.24).
func (a *App) ListDiscoveredRooms() []collaboration.DiscoveredRoom {
	return a.collabMgr.DiscoveredRooms()
}

// JoinCollabRoom join ke room via IP manual atau hasil discovery (§10.22–10.23).
func (a *App) JoinCollabRoom(hostIP string, port int, accessCode, roomCode, displayName string) error {
	if strings.TrimSpace(displayName) == "" {
		displayName = a.settings.CollabDisplayNameOrDefault()
	}
	a.log.Info("JoinCollabRoom dipanggil", "hostIP", hostIP, "port", port)
	err := a.collabMgr.Join(hostIP, port, displayName, accessCode, roomCode)
	if err != nil {
		a.log.Warn("JoinCollabRoom gagal", "error", err, "hostIP", hostIP, "port", port)
	} else {
		a.log.Info("JoinCollabRoom berhasil", "hostIP", hostIP, "port", port)
	}
	return err
}

// LeaveCollabRoom keluar dari room yang sedang diikuti.
func (a *App) LeaveCollabRoom() error { return a.collabMgr.LeaveClientSession() }

// SendCollabOp mengirim satu operasi edit ke sesi aktif.
func (a *App) SendCollabOp(op services.OpPayload) error {
	return a.collabMgr.SendOp(&op)
}

// GetCollabSession mengembalikan potret sesi saat ini (mis. setelah pindah halaman).
func (a *App) GetCollabSession() collaboration.UISnapshot {
	return a.collabMgr.Session()
}

// wireCollab memasang kanal notifikasi UI setelah startup Wails.
func (a *App) wireCollab() {
	a.collabMgr.SetEmit(func(snap collaboration.UISnapshot) {
		runtime.EventsEmit(a.ctx, "collab:sync", snap)
	})
}

func (a *App) AssignTurns(assignments map[string][]string) error {
	return a.collabMgr.AssignTurns(assignments)
}

func (a *App) RequestEdit(sectionID string) error {
	return a.collabMgr.RequestEdit(sectionID)
}

func (a *App) ReleaseEdit(sectionID string) error {
	return a.collabMgr.ReleaseEdit(sectionID)
}

func (a *App) SyncPush(input services.SphSaveInput) error {
	return a.collabMgr.SyncPush(&input)
}
