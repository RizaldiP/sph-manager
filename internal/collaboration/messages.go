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

	// client → host (chat & master data transfer)
	TypeChatMessage        = "CHAT_MESSAGE"         // kirim pesan chat (text/system/master_data)
	TypeChatHistoryRequest = "CHAT_HISTORY_REQUEST" // minta riwayat chat saat join/reconnect
	TypeMasterDataAck      = "MASTER_DATA_ACK"      // client → host: status (RECEIVED/REJECTED/INSTALLED/FAILED)

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

	// host → client (chat & master data transfer)
	TypeChatHistory      = "CHAT_HISTORY"         // payload daftar ChatMessage
	TypeChatBroadcast    = "CHAT_BROADCAST"       // broadcast pesan chat ke semua member
	TypeMasterDataOffer  = "MASTER_DATA_OFFER"    // notifikasi ada transfer masuk (ringkas)
	TypeMasterData       = "MASTER_DATA_TRANSFER" // isi MasterDataPackage ke target tertentu
	TypeMasterDataStatus = "MASTER_DATA_STATUS"   // broadcast status transfer (untuk log sent)
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

// Status pesan chat.
const (
	ChatStatusSent      = "SENT"
	ChatStatusDelivered = "DELIVERED"
)

// Tipe pesan chat.
const (
	ChatTypeText       = "text"
	ChatTypeSystem     = "system"
	ChatTypeMasterData = "master_data"
)

// ChatPayload isi CHAT_MESSAGE / CHAT_BROADCAST / CHAT_HISTORY.
// Untuk message_type=master_data, RefPackage memuat package_id; metadata ringkas
// disimpan di RefMeta agar UI bisa menampilkan card tanpa memuat seluruh package.
type ChatPayload struct {
	MessageID   string    `json:"messageId"`
	RoomID      string    `json:"roomId,omitempty"`
	SenderID    string    `json:"senderId,omitempty"`
	SenderName  string    `json:"senderName,omitempty"`
	MessageType string    `json:"messageType"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	RefPackage  string    `json:"refPackage,omitempty"`
	RefMeta     string    `json:"refMeta,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ChatHistoryPayload isi CHAT_HISTORY (daftar riwayat dari host).
type ChatHistoryPayload struct {
	Messages []ChatPayload `json:"messages"`
}

// Master data install/status values (digunakan untuk ACK & status broadcast).
const (
	MasterStatusReceived     = "RECEIVED"
	MasterStatusAckInstalled = "INSTALLED"
	MasterStatusAckRejected  = "REJECTED"
	MasterStatusAckFailed    = "FAILED"
)

// MasterDataOfferPayload isi MASTER_DATA_OFFER (notifikasi ringkas transfer masuk).
type MasterDataOfferPayload struct {
	PackageID     string    `json:"packageId"`
	SenderID      string    `json:"senderId"`
	SenderName    string    `json:"senderName"`
	SourceVersion int       `json:"sourceVersion"`
	Title         string    `json:"title"`
	Summary       string    `json:"summary"`
	CreatedAt     time.Time `json:"createdAt"`
}

// MasterDataAckPayload isi MASTER_DATA_ACK (client → host: status terima/install/reject/fail).
type MasterDataAckPayload struct {
	PackageID  string    `json:"packageId"`
	Status     string    `json:"status"` // RECEIVED | INSTALLED | REJECTED | FAILED
	Message    string    `json:"message,omitempty"`
	AckedAt    time.Time `json:"ackedAt"`
	TargetID   string    `json:"targetId,omitempty"`
	TargetName string    `json:"targetName,omitempty"`
	Title      string    `json:"title,omitempty"`
}

// MasterDataStatusPayload isi MASTER_DATA_STATUS (broadcast status transfer untuk log sent).
type MasterDataStatusPayload struct {
	PackageID  string    `json:"packageId"`
	TargetID   string    `json:"targetId,omitempty"`
	TargetName string    `json:"targetName,omitempty"`
	Status     string    `json:"status"`
	SenderName string    `json:"senderName,omitempty"`
	Title      string    `json:"title,omitempty"`
	At         time.Time `json:"at"`
}

// MasterStatusEntry adalah status per-penerima untuk satu package Master Data,
// dipakai untuk menampilkan status (Terpasang/Ditolak) pada card chat master_data.
type MasterStatusEntry struct {
	PackageID  string    `json:"packageId"`
	TargetID   string    `json:"targetId,omitempty"`
	TargetName string    `json:"targetName,omitempty"`
	Status     string    `json:"status"`
	At         time.Time `json:"at"`
}

// MasterDataTransferPayload isi MASTER_DATA_TRANSFER: paket penuh + metadata
// relay. Saat client mengirim, field Targets diisi agar host tahu tujuannya;
// saat host meneruskan ke target, Targets dikosongkan.
type MasterDataTransferPayload struct {
	PackageID     string          `json:"packageId"`
	SenderID      string          `json:"senderId,omitempty"`
	SenderName    string          `json:"senderName,omitempty"`
	SourceVersion int             `json:"sourceVersion"`
	Checksum      string          `json:"checksum,omitempty"`
	Title         string          `json:"title,omitempty"`
	Summary       string          `json:"summary,omitempty"`
	Targets       []string        `json:"targets,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}
