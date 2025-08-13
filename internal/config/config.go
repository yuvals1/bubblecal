package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Category represents a calendar category with color
type Category struct {
	Name  string `json:"name"`
	Color string `json:"color"` // Lipgloss color string (e.g., "#FF5733" or "205")
}

// Config holds application configuration
type Config struct {
	ShowMiniMonth bool       `json:"show_mini_month"`
	AgendaBottom  bool       `json:"agenda_bottom"`
	Theme         int        `json:"theme"`
	Categories    []Category `json:"categories"`
}

// DefaultCategories returns the default set of categories
func DefaultCategories() []Category {
	return []Category{
		{Name: "Work", Color: "#4287f5"},      // Blue
		{Name: "Personal", Color: "#42f554"},  // Green
		{Name: "Health", Color: "#f54242"},    // Red
		{Name: "Meeting", Color: "#f5a442"},   // Orange
		{Name: "Important", Color: "#f542e0"}, // Magenta
		{Name: "Travel", Color: "#42f5f5"},    // Cyan
		{Name: "Family", Color: "#f5f542"},    // Yellow
		{Name: "Project", Color: "#8b42f5"},   // Purple
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ShowMiniMonth: true,
		AgendaBottom:  false,             // Default to right side
		Theme:         0,                 // Default theme
		Categories:    DefaultCategories(),
	}
}

// configPath returns the path to the config file
func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".bubblecal")
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
	
	// If categories are empty, use defaults
	if len(cfg.Categories) == 0 {
		cfg.Categories = DefaultCategories()
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

// GetCategoryColor returns the color for a given category name
func (c *Config) GetCategoryColor(categoryName string) string {
	for _, cat := range c.Categories {
		if cat.Name == categoryName {
			return cat.Color
		}
	}
	// Default color if category not found
	return "#808080" // Gray
}