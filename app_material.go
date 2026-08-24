package main

import (
	"github.com/RizaldiP/sph-manager/internal/models"
)

// Binding Master Data Material (FR-M7).
func (a *App) ListMaterials(includeInactive bool, search string) ([]models.Material, error) {
	return a.materials.List(includeInactive, search)
}

func (a *App) CreateMaterial(in *models.Material) (*models.Material, error) {
	return a.materials.Create(in)
}

func (a *App) UpdateMaterial(id uint, in *models.Material) (*models.Material, error) {
	return a.materials.Update(id, in)
}

func (a *App) SetMaterialActive(id uint, active bool) error { return a.materials.SetActive(id, active) }

func (a *App) DeleteMaterial(id uint) error { return a.materials.Delete(id) }
