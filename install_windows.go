//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// installUpdate menulis skrip pengganti executable, menjalankannya di balik
// layar, lalu menyuruh aplikasi menutup diri agar skrip dapat mengganti &
// memulai ulang aplikasi dengan versi baru.
func installUpdate(exePath, newExePath string) error {
	exePath, err := filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("tidak dapat menentukan lokasi aplikasi")
	}
	appName := filepath.Base(exePath)
	batPath := filepath.Join(os.TempDir(), fmt.Sprintf("SPHManager_update_%d.bat", time.Now().UnixNano()))
	script := "@echo off\r\n" +
		"setlocal\r\n" +
		"set \"OLD=%~1\"\r\n" +
		"set \"NEW=%~2\"\r\n" +
		"set \"NAME=" + appName + "\"\r\n" +
		":wait\r\n" +
		"tasklist /FI \"IMAGENAME eq %NAME%\" | find /I \"%NAME%\" >nul\r\n" +
		"if %errorlevel%==0 (\r\n" +
		"  timeout /t 2 /nobreak >nul\r\n" +
		"  goto wait\r\n" +
		")\r\n" +
		"move /Y \"%NEW%\" \"%OLD%\" >nul 2>&1\r\n" +
		"start \"\" \"%OLD%\"\r\n" +
		"del \"%~f0\" >nul 2>&1\r\n"
	if err := os.WriteFile(batPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("gagal menyiapkan installer")
	}
	cmd := exec.Command(batPath, exePath, newExePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal menjalankan installer")
	}
	return nil
}