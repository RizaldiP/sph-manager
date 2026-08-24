package main

import (
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// ===== daftar & detail =====

func (a *App) ListSph(scope string, search string, limit int) ([]services.SphDocumentView, error) {
	return a.sph.List(scope, search, limit)
}

func (a *App) GetSph(id uint) (*models.SphDocument, error) {
	return a.sph.Get(id)
}

func (a *App) DashboardStats() (*services.DashboardStats, error) {
	return a.sph.Stats()
}

func (a *App) Terbilang(total int64) string { return services.Terbilang(total) }

// ===== create / update / delete =====

func (a *App) CreateSph(in services.SphSaveInput) (*services.SphDocumentView, error) {
	return a.sph.Create(in)
}

func (a *App) UpdateDraftSph(id uint, in services.SphSaveInput) (*models.SphDocument, error) {
	return a.sph.UpdateDraft(id, in)
}

func (a *App) DeleteSph(id uint) error { return a.sph.Delete(id) }

// ===== lifecycle & salinan =====

func (a *App) SetSphStatus(id uint, status string) error { return a.sph.SetStatus(id, status) }

func (a *App) DuplicateSph(id uint) (*services.SphDocumentView, error) {
	return a.sph.Duplicate(id)
}

func (a *App) CreateSphRevision(id uint) (*services.SphDocumentView, error) {
	return a.sph.CreateRevision(id)
}
