//go:build !windows

package main

import "fmt"

func installUpdate(exePath, newExePath string) error {
	return fmt.Errorf("pembaruan otomatis hanya tersedia di Windows")
}