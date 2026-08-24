package main

// app_template.go — binding Wails untuk Template Pekerjaan (Phase 4).
// Semua method mendelegasikan ke service layer; tidak ada business rule di sini.

import (
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

func (a *App) ListTemplates(includeInactive bool, search string) ([]services.TemplateView, error) {
	return a.templates.List(includeInactive, search)
}

// GetTemplateDetail mengembalikan template lengkap beserta isian pekerjaannya.
func (a *App) GetTemplateDetail(id uint) (*models.Template, error) {
	return a.templates.Get(id)
}

func (a *App) CreateTemplate(in *models.Template) (*services.TemplateView, error) {
	return a.templates.Create(in)
}

func (a *App) UpdateTemplate(id uint, in *models.Template) (*services.TemplateView, error) {
	return a.templates.Update(id, in)
}

// SetTemplateItems mengganti seluruh isi template; urutan mengikuti urutan input.
func (a *App) SetTemplateItems(id uint, items []services.TemplateItemInput) (*models.Template, error) {
	return a.templates.SetItems(id, items)
}

func (a *App) DuplicateTemplate(id uint) (*services.TemplateView, error) {
	return a.templates.Duplicate(id)
}

func (a *App) SetTemplateActive(id uint, active bool) error {
	return a.templates.SetActive(id, active)
}

func (a *App) DeleteTemplate(id uint) error {
	return a.templates.Delete(id)
}

// ReorderTemplates menyimpan urutan baru seluruh template.
func (a *App) ReorderTemplates(ids []uint) error {
	return a.templates.Reorder(ids)
}
