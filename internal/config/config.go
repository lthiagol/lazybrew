package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GUI  GUIConfig  `yaml:"gui"`
	Brew BrewConfig `yaml:"brew"`
}

type GUIConfig struct {
	Theme              string `yaml:"theme"`
	SidebarWidth       int    `yaml:"sidebar_width"`
	ShowIcons          bool   `yaml:"show_icons"`
	Mouse              bool   `yaml:"mouse"`
	AutoRefreshSeconds int    `yaml:"auto_refresh_seconds"`
}

type BrewConfig struct {
	Path          string `yaml:"path"`
	UpdateOnStart bool   `yaml:"update_on_start"`
}

func Default() *Config {
	return &Config{
		GUI: GUIConfig{
			Theme:              "dark",
			SidebarWidth:       30,
			ShowIcons:          true,
			Mouse:              true,
			AutoRefreshSeconds: 0,
		},
		Brew: BrewConfig{
			Path:          "",
			UpdateOnStart: false,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, nil
		}
		path = filepath.Join(home, ".config", "lazybrew", "config.yml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}
