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
	return a.collabMgr.HostRoom(sphDocumentID, roomName, displayName, a.settings.CollabPortOrDefault())
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
	return a.collabMgr.Join(hostIP, port, displayName, accessCode, roomCode)
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
