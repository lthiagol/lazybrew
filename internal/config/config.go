package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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

// BrewConfig mirrors the `brew:` block in lazybrew config.yml.
//
// OutdatedTTL bounds how long `brew outdated` results are reused before
// the next shell invocation. M9 default is 30 minutes; set to 0 to fall
// back to the cache default (30s).
type BrewConfig struct {
	Path          string        `yaml:"path"`
	UpdateOnStart bool          `yaml:"update_on_start"`
	OutdatedTTL   time.Duration `yaml:"outdated_ttl"`
}

const DefaultOutdatedTTL = 30 * time.Minute

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
			OutdatedTTL:   DefaultOutdatedTTL,
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

	if cfg.Brew.OutdatedTTL <= 0 {
		cfg.Brew.OutdatedTTL = DefaultOutdatedTTL
	}

	return cfg, nil
}
