package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/RizaldiP/sph-manager/internal/config"
	"github.com/RizaldiP/sph-manager/internal/database"
	"github.com/RizaldiP/sph-manager/internal/logger"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("gagal memuat konfigurasi:", err)
		os.Exit(1)
	}

	lg, closeLog, err := logger.New(cfg.LogDir)
	if err != nil {
		fmt.Println("gagal menyiapkan log:", err)
		os.Exit(1)
	}
	defer closeLog()

	db, err := database.Open(cfg.DatabasePath, lg)
	if err != nil {
		lg.Error("gagal membuka database", "error", err)
		closeLog()
		os.Exit(1)
	}
	if err := database.Migrate(db); err != nil {
		lg.Error("gagal migrasi database", "error", err)
		closeLog()
		os.Exit(1)
	}

	app := NewApp(cfg, db, lg)

	err = wails.Run(&options.App{
		Title:     "SPH Manager",
		Width:     1280,
		Height:    800,
		MinWidth:  1100,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		lg.Error("aplikasi berhenti dengan error", "error", err)
		closeLog()
		os.Exit(1)
	}
}
