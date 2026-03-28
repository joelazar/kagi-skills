// Package config provides YAML configuration file loading for the Kagi CLI.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the Kagi CLI configuration.
type Config struct {
	APIKey       string   `yaml:"api_key"`
	SessionToken string   `yaml:"session_token"`
	Defaults     Defaults `yaml:"defaults"`
}

// Defaults holds default values for CLI flags.
type Defaults struct {
	Format string         `yaml:"format"`
	Search SearchDefaults `yaml:"search"`
}

// SearchDefaults holds default values for the search command.
type SearchDefaults struct {
	Region string `yaml:"region"`
}

// Load reads the configuration from the standard locations.
// Priority: ./.kagi.yaml > ~/.config/kagi/config.yaml
// Returns nil config (not an error) if no config file exists.
func Load() (*Config, error) {
	// Try local config first
	if cfg, err := loadFile(".kagi.yaml"); err == nil {
		return cfg, nil
	}

	// Try global config
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, nil //nolint:nilnil,nilerr // no config dir = no config, not an error
	}

	globalPath := filepath.Join(configDir, "kagi", "config.yaml")
	if cfg, err := loadFile(globalPath); err == nil {
		return cfg, nil
	}

	return nil, nil //nolint:nilnil // no config file = no config, not an error
}

func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
