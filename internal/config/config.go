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
	// AutoRefreshPause is the window after a key event during which
	// the auto-refresh tick defers its work so typing/navigation is
	// not janked by a refresh storm (M12 AC-03). 0 disables pause.
	AutoRefreshPause time.Duration `yaml:"auto_refresh_pause"`
}

const DefaultAutoRefreshPause = 500 * time.Millisecond

// BrewConfig mirrors the `brew:` block in lazybrew config.yml.
//
// Per-class TTLs (M12) bound how long `brew …` results are reused before
// the next shell invocation. Each field defaults to a sensible class
// cadence (formulae/casks/taps change rarely; services/doctor can be
// more frequent). Setting any TTL to <= 0 keeps the cache default (30s),
// preserving pre-M12 behavior for that class.
type BrewConfig struct {
	Path          string        `yaml:"path"`
	UpdateOnStart bool          `yaml:"update_on_start"`
	FormulaeTTL   time.Duration `yaml:"formulae_ttl"`
	CasksTTL      time.Duration `yaml:"casks_ttl"`
	OutdatedTTL   time.Duration `yaml:"outdated_ttl"`
	TapsTTL       time.Duration `yaml:"taps_ttl"`
	ServicesTTL   time.Duration `yaml:"services_ttl"`
	DoctorTTL     time.Duration `yaml:"doctor_ttl"`
}

const (
	DefaultFormulaeTTL = 5 * time.Minute
	DefaultCasksTTL    = 5 * time.Minute
	DefaultOutdatedTTL = 30 * time.Minute
	DefaultTapsTTL     = 1 * time.Hour
	DefaultServicesTTL = 30 * time.Second
	DefaultDoctorTTL   = 1 * time.Hour
)

func Default() *Config {
	return &Config{
		GUI: GUIConfig{
			Theme:              "dark",
			SidebarWidth:       30,
			ShowIcons:          true,
			Mouse:              true,
			AutoRefreshSeconds: 0,
			AutoRefreshPause:   DefaultAutoRefreshPause,
		},
		Brew: BrewConfig{
			Path:          "",
			UpdateOnStart: false,
			FormulaeTTL:   DefaultFormulaeTTL,
			CasksTTL:      DefaultCasksTTL,
			OutdatedTTL:   DefaultOutdatedTTL,
			TapsTTL:       DefaultTapsTTL,
			ServicesTTL:   DefaultServicesTTL,
			DoctorTTL:     DefaultDoctorTTL,
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
	if cfg.Brew.FormulaeTTL <= 0 {
		cfg.Brew.FormulaeTTL = DefaultFormulaeTTL
	}
	if cfg.Brew.CasksTTL <= 0 {
		cfg.Brew.CasksTTL = DefaultCasksTTL
	}
	if cfg.Brew.TapsTTL <= 0 {
		cfg.Brew.TapsTTL = DefaultTapsTTL
	}
	if cfg.Brew.ServicesTTL <= 0 {
		cfg.Brew.ServicesTTL = DefaultServicesTTL
	}
	if cfg.Brew.DoctorTTL <= 0 {
		cfg.Brew.DoctorTTL = DefaultDoctorTTL
	}
	if cfg.GUI.AutoRefreshPause < 0 {
		cfg.GUI.AutoRefreshPause = DefaultAutoRefreshPause
	}

	return cfg, nil
}
