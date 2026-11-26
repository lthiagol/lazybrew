package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	Icons              string `yaml:"icons"` // nerd or unicode
	Mouse              bool   `yaml:"mouse"`
	Animations         bool   `yaml:"animations"` // motion; LAZYBREW_ANIMATIONS overrides
	Preload            bool   `yaml:"preload"`    // background panel loads; LAZYBREW_PRELOAD overrides
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
// more frequent). Setting any TTL to <= 0 keeps that class's default (the
// Default*TTL constants), preserving pre-M12 behavior for that class. A TTL
// above 0 but under 1s is an error: 300ns parses, but it is no cache TTL. A TTL
// must also be written as a duration string: a bare number is an error.
type BrewConfig struct {
	Path          string     `yaml:"path"`
	UpdateOnStart UpdateMode `yaml:"update_on_start"`
	// UpdateStaleHours is how old brew's last update may be before auto
	// mode runs `brew update` at launch.
	UpdateStaleHours int           `yaml:"update_stale_hours"`
	FormulaeTTL      time.Duration `yaml:"formulae_ttl"`
	CasksTTL         time.Duration `yaml:"casks_ttl"`
	OutdatedTTL      time.Duration `yaml:"outdated_ttl"`
	TapsTTL          time.Duration `yaml:"taps_ttl"`
	ServicesTTL      time.Duration `yaml:"services_ttl"`
	DoctorTTL        time.Duration `yaml:"doctor_ttl"`
}

// ttlField is one brew TTL: its config key, the field it fills and its class default.
type ttlField struct {
	key   string
	value *time.Duration
	def   time.Duration
}

// ttls lists the six brew TTLs. loadFile checks their values; UnmarshalYAML
// checks their forms.
func (b *BrewConfig) ttls() []ttlField {
	return []ttlField{
		{"brew.outdated_ttl", &b.OutdatedTTL, DefaultOutdatedTTL},
		{"brew.formulae_ttl", &b.FormulaeTTL, DefaultFormulaeTTL},
		{"brew.casks_ttl", &b.CasksTTL, DefaultCasksTTL},
		{"brew.taps_ttl", &b.TapsTTL, DefaultTapsTTL},
		{"brew.services_ttl", &b.ServicesTTL, DefaultServicesTTL},
		{"brew.doctor_ttl", &b.DoctorTTL, DefaultDoctorTTL},
	}
}

// UnmarshalYAML reads the brew block. A TTL is checked for its form before the
// block is decoded: yaml's own message for a bare number ("cannot unmarshal
// !!int") does not name the key.
func (b *BrewConfig) UnmarshalYAML(node *yaml.Node) error {
	for i := 0; i+1 < len(node.Content); i += 2 {
		name := "brew." + node.Content[i].Value
		for _, ttl := range b.ttls() {
			if ttl.key == name {
				if err := checkTTLForm(ttl, node.Content[i+1]); err != nil {
					return err
				}
			}
		}
	}
	type plain BrewConfig
	return node.Decode((*plain)(b))
}

// checkTTLForm refuses a TTL that yaml cannot read as a duration, such as 300 or
// abc. yaml decides: an alias (*d) or a local tag (!custom 5m) reads as it does
// on its own. An empty value keeps the class default, as it always did.
func checkTTLForm(ttl ttlField, val *yaml.Node) error {
	var d time.Duration
	if err := val.Decode(&d); err == nil {
		return nil
	}
	shown := val.Value
	if val.Kind == yaml.AliasNode {
		shown = val.Alias.Value
	}
	return fmt.Errorf("%s: %q is not a duration; write a duration with a unit, e.g. %s",
		ttl.key, shown, sampleDuration(ttl.def))
}

// UpdateMode is brew.update_on_start. never (false) leaves brew's metadata as it
// is; always (true, the default) runs `brew update` at every launch; auto runs
// it at launch only when brew's last update is older than
// brew.update_stale_hours.
type UpdateMode string

const (
	UpdateNever  UpdateMode = "false"
	UpdateAlways UpdateMode = "true"
	UpdateAuto   UpdateMode = "auto"
)

// DefaultUpdateStaleHours is how old brew's last update may be before auto mode
// runs `brew update` at launch.
const DefaultUpdateStaleHours = 6

// UnmarshalYAML reads true, false or auto. A YAML bool decodes as it always
// did, and so does its quoted form ("true", "false"), which YAML reads as a
// string.
func (u *UpdateMode) UnmarshalYAML(node *yaml.Node) error {
	var on bool
	if err := node.Decode(&on); err == nil {
		*u = boolUpdateMode(on)
		return nil
	}
	switch node.Value {
	case string(UpdateAuto):
		*u = UpdateAuto
		return nil
	case "true", "false":
		*u = boolUpdateMode(node.Value == "true")
		return nil
	}
	return fmt.Errorf("brew.update_on_start: unknown value %q (want true, false or auto)", node.Value)
}

// boolUpdateMode is the mode a true or false update_on_start means.
func boolUpdateMode(on bool) UpdateMode {
	if on {
		return UpdateAlways
	}
	return UpdateNever
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
			SidebarWidth:       35,
			ShowIcons:          true,
			Icons:              IconsNerd,
			Mouse:              true,
			Animations:         true,
			Preload:            true,
			AutoRefreshSeconds: 0,
			AutoRefreshPause:   DefaultAutoRefreshPause,
		},
		Brew: BrewConfig{
			Path:             "",
			UpdateOnStart:    UpdateAlways,
			UpdateStaleHours: DefaultUpdateStaleHours,
			FormulaeTTL:      DefaultFormulaeTTL,
			CasksTTL:         DefaultCasksTTL,
			OutdatedTTL:      DefaultOutdatedTTL,
			TapsTTL:          DefaultTapsTTL,
			ServicesTTL:      DefaultServicesTTL,
			DoctorTTL:        DefaultDoctorTTL,
		},
	}
}

// IconsEnv overrides gui.icons for one run: LAZYBREW_ICONS=unicode.
const IconsEnv = "LAZYBREW_ICONS"

// AnimationsEnv overrides gui.animations for one run: LAZYBREW_ANIMATIONS=0
// turns the motion off, LAZYBREW_ANIMATIONS=1 turns it on.
const AnimationsEnv = "LAZYBREW_ANIMATIONS"

// PreloadEnv overrides gui.preload for one run: LAZYBREW_PRELOAD=0 turns the
// background loading off, LAZYBREW_PRELOAD=1 turns it on.
const PreloadEnv = "LAZYBREW_PRELOAD"

// The icon set names gui.icons and LAZYBREW_ICONS accept.
const (
	IconsNerd    = "nerd"
	IconsUnicode = "unicode"
)

// themeNames are the names gui.theme accepts, in the order the errors list
// them. The GUI's registry must match: a test in the style package checks it,
// so config does not import the GUI.
var themeNames = []string{
	"dark", "light",
	"catppuccin-mocha", "catppuccin-macchiato", "catppuccin-frappe", "catppuccin-latte",
	"dracula", "nord",
}

// ThemeNames lists the names gui.theme accepts, in display order.
func ThemeNames() []string {
	return append([]string(nil), themeNames...)
}

// Load reads the config file and then applies LAZYBREW_ICONS, LAZYBREW_ANIMATIONS
// and LAZYBREW_PRELOAD. An unknown icon set or switch is an error, from the file
// or from the environment.
func Load(path string) (*Config, error) {
	cfg, err := loadFile(path)
	if err != nil {
		return nil, err
	}
	if !validIcons(cfg.GUI.Icons) {
		return nil, fmt.Errorf("gui.icons: unknown value %q (want %s or %s)",
			cfg.GUI.Icons, IconsNerd, IconsUnicode)
	}
	if !validTheme(cfg.GUI.Theme) {
		return nil, fmt.Errorf("gui.theme: unknown value %q (want one of %s)",
			cfg.GUI.Theme, strings.Join(ThemeNames(), ", "))
	}
	if v := os.Getenv(IconsEnv); v != "" {
		if !validIcons(v) {
			return nil, fmt.Errorf("%s: unknown value %q (want %s or %s)",
				IconsEnv, v, IconsNerd, IconsUnicode)
		}
		cfg.GUI.Icons = v
	}
	if err := boolEnv(AnimationsEnv, &cfg.GUI.Animations); err != nil {
		return nil, err
	}
	if err := boolEnv(PreloadEnv, &cfg.GUI.Preload); err != nil {
		return nil, err
	}
	return cfg, nil
}

// boolEnv sets dst from a 0/1 environment switch. An unset variable leaves dst
// as it is; any other value is an error.
func boolEnv(name string, dst *bool) error {
	switch v := os.Getenv(name); v {
	case "":
	case "0":
		*dst = false
	case "1":
		*dst = true
	default:
		return fmt.Errorf("%s: unknown value %q (want 0 or 1)", name, v)
	}
	return nil
}

func validIcons(name string) bool {
	return name == IconsNerd || name == IconsUnicode
}

// validTheme reports whether name is a theme gui.theme accepts, in any case.
func validTheme(name string) bool {
	for _, t := range themeNames {
		if strings.EqualFold(t, name) {
			return true
		}
	}
	return false
}

// DefaultPath is the config file read when no --config path is given.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lazybrew", "config.yml"), nil
}

func loadFile(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		var err error
		if path, err = DefaultPath(); err != nil {
			return cfg, nil
		}
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

	// A TTL of 0 or less keeps the class's default; one above 0 but under 1s is
	// an error. This is the one place the check is made; the brew readers take
	// the TTLs as they are given.
	for _, ttl := range cfg.Brew.ttls() {
		if *ttl.value <= 0 {
			*ttl.value = ttl.def
			continue
		}
		if *ttl.value < time.Second {
			return nil, fmt.Errorf("%s: %s is under 1s; write a duration with a unit, e.g. %s",
				ttl.key, *ttl.value, sampleDuration(ttl.def))
		}
	}
	if cfg.Brew.UpdateStaleHours < 1 {
		return nil, fmt.Errorf("brew.update_stale_hours: %d is below 1", cfg.Brew.UpdateStaleHours)
	}
	if cfg.GUI.AutoRefreshPause < 0 {
		cfg.GUI.AutoRefreshPause = DefaultAutoRefreshPause
	}

	return cfg, nil
}
