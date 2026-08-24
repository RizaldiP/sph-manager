package services

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// AuditWriter mencatat aksi pengguna ke tabel audit_logs (BR-13).
// Dipanggil di dalam transaksi yang sama dengan operasinya agar konsisten.
type AuditWriter struct{}

func NewAuditWriter() *AuditWriter { return &AuditWriter{} }

func (w *AuditWriter) Write(db *gorm.DB, action, entity string, entityID uint, description string) error {
	entry := models.AuditLog{
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
		Description: description,
	}
	if err := db.Create(&entry).Error; err != nil {
		return fmt.Errorf("gagal menulis audit log: %w", err)
	}
	return nil
}
