package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	ShowMiniMonth bool `json:"show_mini_month"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ShowMiniMonth: true,
	}
}

// configPath returns the path to the config file
func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".simple-tui-cal")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(configDir, "config.json"), nil
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), nil
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist yet, return defaults
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}
	
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}
	
	return &cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}