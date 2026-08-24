package main

import (
	"github.com/RizaldiP/sph-manager/internal/models"
	"github.com/RizaldiP/sph-manager/internal/services"
)

// ===== customer =====

func (a *App) ListCustomers(includeInactive bool, search string) ([]services.CustomerView, error) {
	return a.customers.List(includeInactive, search)
}

func (a *App) CreateCustomer(in *models.Customer) (*services.CustomerView, error) {
	return a.customers.Create(in)
}

func (a *App) UpdateCustomer(id uint, in *models.Customer) (*models.Customer, error) {
	return a.customers.Update(id, in)
}

func (a *App) SetCustomerActive(id uint, active bool) error { return a.customers.SetActive(id, active) }

func (a *App) DeleteCustomer(id uint) error { return a.customers.Delete(id) }

// ===== kapal =====

func (a *App) CreateVessel(in *models.Vessel) (*models.Customer, error) {
	return a.customers.CreateVessel(in)
}

func (a *App) UpdateVessel(id uint, in *models.Vessel) (*models.Customer, error) {
	return a.customers.UpdateVessel(id, in)
}

func (a *App) SetVesselActive(id uint, active bool) error {
	return a.customers.SetVesselActive(id, active)
}

func (a *App) DeleteVessel(id uint) error { return a.customers.DeleteVessel(id) }
