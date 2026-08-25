package main

import (
	"context"
	"log/slog"
	"os"
	"runtime"

	"github.com/RizaldiP/sph-manager/internal/collaboration"
	"github.com/RizaldiP/sph-manager/internal/config"
	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/services"
	"gorm.io/gorm"
)

const appVersion = "0.10.0"

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
	settings   *services.SettingsService
	export     *services.ExportService
	collabMgr  *collaboration.Manager
}

func NewApp(cfg *config.Config, db *gorm.DB, lg *slog.Logger) *App {
	settingsSvc := services.NewSettingsService(db, lg)
	sphSvc := services.NewSphService(db, lg)
	collabOps := services.NewCollabOps(db, sphSvc, lg)
	deviceName, _ := os.Hostname()
	collabMgr := collaboration.NewManager(collaboration.Config{DeviceName: deviceName}, collabOps, sphSvc, lg)
	// Dokumen yang dibuka dalam room kolaborasi terkunci dari jalur solo (guard BR-08/BR-16).
	sphSvc.SetRoomGuard(collabMgr)
	return &App{
		cfg:        cfg,
		db:         db,
		log:        lg,
		categories: services.NewCategoryService(db, lg),
		workItems:  services.NewWorkItemService(db, lg),
		subItems:   services.NewWorkSubItemService(db, lg),
		templates:  services.NewTemplateService(db, lg),
		sph:        sphSvc,
		customers:  services.NewCustomerService(db, lg),
		materials:  services.NewMaterialService(db, lg),
		settings:   settingsSvc,
		export:     services.NewExportService(sphSvc, settingsSvc, lg),
		collabMgr:  collabMgr,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.wireCollab()
	a.log.Info("aplikasi dimulai", "versi", appVersion, "database", a.cfg.DatabasePath)
}

func (a *App) shutdown(ctx context.Context) {
	if a.collabMgr != nil {
		a.collabMgr.Shutdown("Aplikasi ditutup.")
	}
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
