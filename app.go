package main

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/RizaldiP/sph-manager/internal/config"
	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/services"
	"gorm.io/gorm"
)

const appVersion = "0.4.0"

type HealthInfo struct {
	Status       string `json:"status"`
	Version      string `json:"version"`
	Platform     string `json:"platform"`
	DatabasePath string `json:"databasePath"`
}

type App struct {
	ctx context.Context
	cfg *config.Config
	db  *gorm.DB
	log *slog.Logger

	categories *services.CategoryService
	workItems  *services.WorkItemService
	subItems   *services.WorkSubItemService
	templates  *services.TemplateService
	sph        *services.SphService
	customers  *services.CustomerService
	materials  *services.MaterialService
}

func NewApp(cfg *config.Config, db *gorm.DB, lg *slog.Logger) *App {
	return &App{
		cfg:        cfg,
		db:         db,
		log:        lg,
		categories: services.NewCategoryService(db, lg),
		workItems:  services.NewWorkItemService(db, lg),
		subItems:   services.NewWorkSubItemService(db, lg),
		templates:  services.NewTemplateService(db, lg),
		sph:        services.NewSphService(db, lg),
		customers:  services.NewCustomerService(db, lg),
		materials:  services.NewMaterialService(db, lg),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.log.Info("aplikasi dimulai", "versi", appVersion, "database", a.cfg.DatabasePath)
}

func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		if err := database.Close(a.db); err != nil {
			a.log.Warn("gagal menutup database", "error", err)
		}
	}
	a.log.Info("aplikasi ditutup")
}

func (a *App) Health() HealthInfo {
	return HealthInfo{
		Status:       "ok",
		Version:      appVersion,
		Platform:     runtime.GOOS + "/" + runtime.GOARCH,
		DatabasePath: a.cfg.DatabasePath,
	}
}
