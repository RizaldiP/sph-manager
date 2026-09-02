package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// openPath membuka folder di file manager sistem. Di Windows memakai
// explorer.exe agar folder selalu terbuka di Explorer; platform lain
// jatuh ke handler default via Wails runtime.
func openPath(ctx context.Context, path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("path tidak valid")
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("folder tidak ditemukan: %s", abs)
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command("explorer.exe", abs)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("gagal membuka folder: %s", abs)
		}
		return nil
	}
	wailsruntime.BrowserOpenURL(ctx, "file://"+abs)
	return nil
}