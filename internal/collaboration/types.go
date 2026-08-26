// Package collaboration menyediakan fitur Work Together (Phase 10):
// co-edit satu draft SPH oleh beberapa PC pada jaringan lokal yang sama,
// tanpa internet/hosting/domain. Model host-authoritative: room in-memory
// di host, sinkronisasi via WebSocket, discovery via UDP broadcast.
// Acuan spesifikasi: docs/collaboration-lan.md §10.1–10.48.
package collaboration

import "time"

const (
	RoleHost   = "HOST"
	RoleEditor = "EDITOR"
)

const (
	RoomStatusActive = "ACTIVE"
	RoomStatusClosed = "CLOSED"
)

// Mode sesi aplikasi lokal (indikator §10.35).
const (
	ModeNone   = ""
	ModeHost   = "HOST"
	ModeClient = "CLIENT"
)

// Status koneksi client ke host (§10.13).
const (
	ConnConnected    = "CONNECTED"
	ConnReconnecting = "RECONNECTING"
	ConnDisconnected = "DISCONNECTED"
)

// DefaultDiscoveryPort adalah port UDP broadcast discovery room.
const DefaultDiscoveryPort = 48766

// Participant adalah identitas peserta room (§10.5).
type Participant struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	DeviceName  string    `json:"deviceName"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
	LastSeen    time.Time `json:"lastSeen"`
}

// RoomInfo adalah potret room untuk UI host maupun client.
// AccessCode hanya diisi pada sisi host (tidak pernah dikirim ke client).
type RoomInfo struct {
	RoomID           string        `json:"roomId"`
	SphDocumentID    uint          `json:"sphDocumentId"`
	DocumentNumber   string        `json:"documentNumber"`
	ProjectName      string        `json:"projectName"`
	RoomCode         string        `json:"roomCode"`
	RoomName         string        `json:"roomName"`
	AccessCode       string        `json:"accessCode,omitempty"`
	HostName         string        `json:"hostName"`
	HostDevice       string        `json:"hostDevice"`
	HostIPs          []string      `json:"hostIPs,omitempty"`
	Port             int           `json:"port"`
	Status           string        `json:"status"`
	Version          uint64        `json:"version"`
	Participants     []Participant `json:"participants,omitempty"`
	FirewallWarning  string        `json:"firewallWarning,omitempty"`
	CreatedAt        time.Time     `json:"createdAt"`
}

// Sanitized menyalin info room tanpa access code agar aman dikirim ke client.
func (r RoomInfo) Sanitized() RoomInfo {
	r.AccessCode = ""
	r.Participants = nil
	return r
}

// DiscoveredRoom: entri hasil UDP broadcast yang tampil pada lobby (§10.24).
type DiscoveredRoom struct {
	RoomID         string    `json:"roomId"`
	RoomName       string    `json:"roomName"`
	DocumentNumber string    `json:"documentNumber"`
	ProjectName    string    `json:"projectName"`
	HostIP         string    `json:"hostIP"`
	HostName       string    `json:"hostName"`
	Port           int       `json:"port"`
	Users          int       `json:"users"`
	LastSeen       time.Time `json:"lastSeen"`
}

// TurnState tracks turn-based edit assignments for the room.
// Section IDs: "header", "items" (all items group), "subitems" (all sub-items group).
type TurnState struct {
	Assignments map[string][]string `json:"assignments"` // participantID → []sectionID
	ActiveEdits map[string]string   `json:"activeEdits"` // sectionID → participantID
}
