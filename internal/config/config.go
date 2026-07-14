package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds persistent user configuration.
type Config struct {
	Theme string `toml:"theme"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Theme: "default",
	}
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "wlocks", "config.toml"), nil
}

// Load reads the config file. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Default(), nil
	}

	// If file doesn't exist, return default
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

// Save writes the config file.
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to TOML
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
