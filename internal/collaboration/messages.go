package collaboration

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/RizaldiP/sph-manager/internal/services"
)

// Tipe pesan (envelope) sesuai §10.9–10.10.
const (
	// client → host
	TypeJoinRequest = "JOIN_REQUEST"
	TypeLeave       = "LEAVE"
	TypeOpRequest   = "OP_REQUEST"
	TypeSyncRequest = "SYNC_REQUEST"
	TypePing        = "PING"
	TypeAssignTurns = "ASSIGN_TURNS"
	TypeRequestEdit = "REQUEST_EDIT"
	TypeReleaseEdit = "RELEASE_EDIT"
	TypeSyncPush    = "SYNC_PUSH"

	// host → client
	TypeRoomJoined      = "ROOM_JOINED"   // jawaban join/reconnect berisi initial sync
	TypeSyncResponse    = "SYNC_RESPONSE" // jawaban SYNC_REQUEST (bentuk payload sama)
	TypeSphUpdated      = "SPH_UPDATED"   // hasil operasi diterapkan (broadcast)
	TypeTurnsUpdated    = "TURNS_UPDATED"
	TypeUserConnected   = "USER_CONNECTED"
	TypeUserDisonnected = "USER_DISCONNECTED"
	TypeRoomClosed      = "ROOM_CLOSED"
	TypePong            = "PONG"
	TypeError           = "ERROR"
)

// Kode error untuk payload ERROR.
const (
	errCodeAuth      = "AUTH_FAILED"
	errCodeClosed    = "ROOM_CLOSED"
	errCodeNotJoined = "NOT_JOINED"
	errCodeOp        = "OP_REJECTED"
	errCodeInternal  = "INTERNAL"
)

// Envelope: amplop pesan WebSocket (§10.10).
type Envelope struct {
	MessageID string          `json:"messageId"`
	RoomID    string          `json:"roomId,omitempty"`
	ClientID  string          `json:"clientId,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Type      string          `json:"type"`
	Version   uint64          `json:"version,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

func newEnvelope(t string) *Envelope {
	return &Envelope{
		MessageID: uuid.NewString(),
		Timestamp: time.Now().UTC(),
		Type:      t,
	}
}

func envelopeWith(t, roomID, clientID string, version uint64, payload interface{}) *Envelope {
	e := newEnvelope(t)
	e.RoomID = roomID
	e.ClientID = clientID
	e.Version = version
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			e.Payload = b
		}
	}
	return e
}

// JoinRequest adalah pesan pertama dari client (auth + identitas §10.6).
// ClientID terisi berarti reconnect dengan identitas lama.
type JoinRequest struct {
	ClientID    string `json:"clientId,omitempty"`
	DisplayName string `json:"displayName"`
	DeviceName  string `json:"deviceName"`
	AccessCode  string `json:"accessCode"`
	RoomCode    string `json:"roomCode,omitempty"`
}

// StatePayload: isi ROOM_JOINED/SYNC_RESPONSE/SPH_UPDATED/USER_* — selalu membawa
// state dokumen penuh sehingga client cukup mengganti state lokal secara keseluruhan.
type StatePayload struct {
	Room         *RoomInfo                `json:"room,omitempty"`
	State        json.RawMessage          `json:"state,omitempty"`
	Activity     *services.CollabActivity `json:"activity,omitempty"`
	Participants []Participant            `json:"participants,omitempty"`
	Turn         *TurnState               `json:"turn,omitempty"`
}

// ErrorPayload isi pesan ERROR.
type ErrorPayload struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// ClosedPayload isi pesan ROOM_CLOSED.
type ClosedPayload struct {
	Reason string `json:"reason"`
}
