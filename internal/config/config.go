package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Theme           string `toml:"theme"`
	DefaultSort     string `toml:"default_sort"`
	LiveRefreshRate int    `toml:"live_refresh_rate_seconds"`
	AnimationSpeed  string `toml:"animation_speed"`
}

func Default() *Config {
	return &Config{
		Theme:           "default",
		DefaultSort:     "duration",
		LiveRefreshRate: 2,
		AnimationSpeed:  "normal",
	}
}

func ConfigPath() (string, error) {
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		xdgHome = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgHome, "wlocks", "config.toml"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Default(), nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil
	}

	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return Default(), nil
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
