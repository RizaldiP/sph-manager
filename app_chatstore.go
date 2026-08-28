package main

import (
	"github.com/RizaldiP/sph-manager/internal/collaboration"
	"github.com/RizaldiP/sph-manager/internal/models"
	"gorm.io/gorm"
)

// gormChatStore adalah implementasi collaboration.ChatStore memakai database
// lokal host (SQLite). Ini satu-satunya jembatan antara protokol chat LAN dan
// storage riwayat pesan.
type gormChatStore struct {
	db *gorm.DB
}

func newGormChatStore(db *gorm.DB) *gormChatStore {
	return &gormChatStore{db: db}
}

func (s *gormChatStore) SaveChat(m collaboration.ChatMessage) error {
	row := models.ChatMessage{
		RoomID:      m.RoomID,
		MessageID:   m.MessageID,
		SenderID:    m.SenderID,
		SenderName:  m.SenderName,
		MessageType: m.MessageType,
		Content:     m.Content,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
	}
	return s.db.Create(&row).Error
}

func (s *gormChatStore) History(roomID string, limit int) ([]collaboration.ChatMessage, error) {
	query := s.db.
		Where("room_id = ?", roomID).
		Order("created_at ASC, id ASC")
	var rows []models.ChatMessage
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	// Ambil `limit` pesan terbaru.
	total := len(rows)
	if limit > 0 && total > limit {
		rows = rows[total-limit:]
	}
	out := make([]collaboration.ChatMessage, 0, len(rows))
	for _, r := range rows {
		out = append(out, collaboration.ChatMessage{
			RoomID:      r.RoomID,
			MessageID:   r.MessageID,
			SenderID:    r.SenderID,
			SenderName:  r.SenderName,
			MessageType: r.MessageType,
			Content:     r.Content,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
		})
	}
	return out, nil
}
