package main

// app_master.go — binding Wails untuk Master Pekerjaan (Phase 3).
// Semua method mendelegasikan ke service layer; tidak ada business rule di sini.

import (
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// ===== Kategori =====

func (a *App) ListCategories(includeInactive bool, search string) ([]services.CategoryView, error) {
	return a.categories.List(includeInactive, search)
}

func (a *App) CreateCategory(in *models.Category) (*services.CategoryView, error) {
	return a.categories.Create(in)
}

func (a *App) UpdateCategory(id uint, in *models.Category) (*services.CategoryView, error) {
	return a.categories.Update(id, in)
}

func (a *App) SetCategoryActive(id uint, active bool) error {
	return a.categories.SetActive(id, active)
}

func (a *App) DeleteCategory(id uint) error {
	return a.categories.Delete(id)
}

// ReorderCategories menyimpan urutan baru seluruh kategori.
func (a *App) ReorderCategories(ids []uint) error {
	return a.categories.Reorder(ids)
}

// ===== Pekerjaan =====

func (a *App) ListWorkItems(categoryID uint, includeInactive bool, search string) ([]services.WorkItemView, error) {
	return a.workItems.List(categoryID, includeInactive, search)
}

func (a *App) GetWorkItemDetail(id uint) (*models.WorkItem, error) {
	return a.workItems.GetDetail(id)
}

func (a *App) CreateWorkItem(in *models.WorkItem) (*models.WorkItem, error) {
	return a.workItems.Create(in)
}

func (a *App) UpdateWorkItem(id uint, in *models.WorkItem) (*models.WorkItem, error) {
	return a.workItems.Update(id, in)
}

func (a *App) SetWorkItemActive(id uint, active bool) error {
	return a.workItems.SetActive(id, active)
}

func (a *App) DeleteWorkItem(id uint) error {
	return a.workItems.Delete(id)
}

// ReorderWorkItems menyimpan urutan baru pekerjaan dalam satu kategori.
func (a *App) ReorderWorkItems(categoryID uint, ids []uint) error {
	if categoryID == 0 {
		return services.NewValidationError("Pilih satu kategori terlebih dahulu untuk mengatur urutan pekerjaan.")
	}
	return a.workItems.ReorderInCategory(categoryID, ids)
}

// ===== Sub-Pekerjaan =====

func (a *App) CreateSubItem(in *models.WorkSubItem) (*models.WorkItem, error) {
	return a.subItems.Create(in)
}

func (a *App) UpdateSubItem(id uint, in *models.WorkSubItem) (*models.WorkItem, error) {
	return a.subItems.Update(id, in)
}

func (a *App) SetSubItemActive(id uint, active bool) error {
	return a.subItems.SetActive(id, active)
}

func (a *App) DeleteSubItem(id uint) error {
	return a.subItems.Delete(id)
}

// ReorderSubItems menyimpan urutan baru sub-pekerjaan dalam satu pekerjaan.
func (a *App) ReorderSubItems(workItemID uint, ids []uint) error {
	return a.subItems.ReorderInWorkItem(workItemID, ids)
}
