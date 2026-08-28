package collaboration

import "time"

// ChatMessage adalah representasi pesan chat room (ringan, tanpa koneksi model DB)
// yang dipersist oleh host. App meng-bridge ke storage (SQLite) via ChatStore.
type ChatMessage struct {
	RoomID      string    `json:"roomId"`
	MessageID   string    `json:"messageId"`
	SenderID    string    `json:"senderId"`
	SenderName  string    `json:"senderName"`
	MessageType string    `json:"messageType"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ChatStore menyimpan & memuat riwayat chat. Diimplementasikan app (host) memakai
// database lokal host sebagai sumber kebenaran riwayat Room.
type ChatStore interface {
	SaveChat(m ChatMessage) error
	History(roomID string, limit int) ([]ChatMessage, error)
}
