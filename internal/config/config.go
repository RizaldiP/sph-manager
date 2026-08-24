package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const appName = "sph-manager"

type Config struct {
	DatabasePath string `json:"database_path"`
	BackupDir    string `json:"backup_dir"`
	ExportDir    string `json:"export_dir"`
	LogDir       string `json:"log_dir"`
	TemplateDir  string `json:"template_dir"`
	AssetsDir    string `json:"assets_dir"`
}

func RootDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

func Load() (*Config, error) {
	root, err := RootDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(root, "config.json")
	cfg := defaults(root)
	data, err := os.ReadFile(path)
	if err == nil {
		if jerr := json.Unmarshal(data, cfg); jerr != nil {
			return nil, jerr
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	dirs := []string{
		filepath.Dir(cfg.DatabasePath),
		cfg.BackupDir,
		cfg.ExportDir,
		cfg.LogDir,
		cfg.TemplateDir,
		cfg.AssetsDir,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, err
		}
	}
	if werr := save(path, cfg); werr != nil {
		return nil, werr
	}
	return cfg, nil
}

func defaults(root string) *Config {
	return &Config{
		DatabasePath: filepath.Join(root, "database", "sph.db"),
		BackupDir:    filepath.Join(root, "backups"),
		ExportDir:    filepath.Join(root, "exports"),
		LogDir:       filepath.Join(root, "logs"),
		TemplateDir:  filepath.Join(root, "templates"),
		AssetsDir:    filepath.Join(root, "assets"),
	}
}

func save(path string, c *Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
