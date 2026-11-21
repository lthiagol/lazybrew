package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// sampleComments is the prose in the sample: one line per key, keyed by its
// dotted path. It names what the key does and the values it accepts. A key
// added to Config without a line here fails TestSampleCommentsCoverEveryKey.
var sampleComments = map[string]string{
	"gui":                      "the interface",
	"gui.theme":                "dark, light, catppuccin-mocha, catppuccin-macchiato, catppuccin-frappe, catppuccin-latte, dracula or nord",
	"gui.sidebar_width":        "percent of the terminal width, 15-50",
	"gui.show_icons":           "true or false; false draws no icons",
	"gui.icons":                "nerd or unicode (no Nerd Font needed)",
	"gui.mouse":                "true or false; true blocks terminal text selection: hold Option/Shift, or set false",
	"gui.animations":           "true or false; motion, short and never delays a key",
	"gui.preload":              "true or false; load every panel in the background",
	"gui.auto_refresh_seconds": "0 is off; N refreshes the data every N seconds",
	"gui.auto_refresh_pause":   "a duration, e.g. 500ms; 0 turns the pause off",
	"brew":                     "how lazybrew runs brew",
	"brew.path":                "the brew binary; empty auto-detects it",
	"brew.update_on_start":     "true: brew update at every launch; the other two values are below",
	"brew.update_stale_hours":  "1 or more hours; auto updates when brew's last update is older",
	"brew.formulae_ttl":        "a duration, e.g. 5m; how long formula results are reused; 0 or less keeps the default",
	"brew.casks_ttl":           "a duration, e.g. 5m; how long cask results are reused; 0 or less keeps the default",
	"brew.outdated_ttl":        "a duration, e.g. 30m; how long the outdated list is reused; 0 or less keeps the default",
	"brew.taps_ttl":            "a duration, e.g. 1h; how long tap results are reused; 0 or less keeps the default",
	"brew.services_ttl":        "a duration, e.g. 30s; how long service results are reused; 0 or less keeps the default",
	"brew.doctor_ttl":          "a duration, e.g. 1h; how long doctor output is reused; 0 or less keeps the default",
}

// sampleAlternatives are the values a key takes besides the one the sample
// sets. Each is written as a commented line under the key, with what it does.
var sampleAlternatives = map[string][]sampleAlternative{
	"brew.update_on_start": {
		{"auto", "brew update at launch only when brew's last update is older than update_stale_hours"},
		{"false", "never run brew update at launch"},
	},
}

type sampleAlternative struct {
	value   string
	comment string
}

var (
	durationType   = reflect.TypeOf(time.Duration(0))
	updateModeType = reflect.TypeOf(UpdateMode(""))
)

// Sample returns the documented config file that --config-sample prints. Its
// uncommented values are Default(), so the sample cannot drift from the defaults.
func Sample() string {
	return sample(Default())
}

// sample writes cfg in the order of Config's fields, with each key's prose from
// sampleComments. Sample is sample(Default()); the split lets a test change a
// default and see the sample follow.
func sample(cfg *Config) string {
	var b strings.Builder
	b.WriteString(sampleHeader())
	writeSampleSection(&b, reflect.ValueOf(cfg).Elem(), "", "")
	return b.String()
}

// sampleHeader names the file's rules and the environment overrides. The
// overrides come from the constants, so a renamed one changes the header too.
func sampleHeader() string {
	return fmt.Sprintf(`# lazybrew config. The file is optional: delete a line to follow
# lazybrew's default for that key.
#
# For one run, these environment variables override the file:
#   %-28s gui.icons
#   %-28s gui.animations
#   %-28s gui.preload

`, IconsEnv+"=nerd|unicode", AnimationsEnv+"=0|1", PreloadEnv+"=0|1")
}

// writeSampleSection writes the keys of the struct v. A nested struct is a
// section: its key on a line of its own, its keys indented beneath it.
func writeSampleSection(b *strings.Builder, v reflect.Value, prefix, indent string) {
	t := v.Type()
	for i := range t.NumField() {
		key := t.Field(i).Tag.Get("yaml")
		if key == "" {
			continue
		}
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		fv := v.Field(i)
		if fv.Kind() == reflect.Struct {
			fmt.Fprintf(b, "%-32s # %s\n", indent+key+":", sampleComments[path])
			writeSampleSection(b, fv, path, indent+"  ")
			continue
		}
		line := indent + key + ": " + sampleValue(fv)
		fmt.Fprintf(b, "%-32s # %s\n", line, sampleComments[path])
		for _, alt := range sampleAlternatives[path] {
			commented := indent + "# " + key + ": " + alt.value
			fmt.Fprintf(b, "%-32s # %s\n", commented, alt.comment)
		}
	}
}

// sampleValue writes a leaf value the way it reads in a config file.
func sampleValue(v reflect.Value) string {
	if v.Type() == durationType {
		return sampleDuration(time.Duration(v.Int()))
	}
	if v.Type() == updateModeType {
		// The two bool modes are written as the bools a person types, true and
		// false, not as the quoted strings they are held as.
		switch UpdateMode(v.String()) {
		case UpdateAlways:
			return "true"
		case UpdateNever:
			return "false"
		}
	}
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int:
		return fmt.Sprint(v.Int())
	case reflect.String:
		// yaml quotes a value it would otherwise misread, such as "".
		out, _ := yaml.Marshal(v.String())
		return strings.TrimSpace(string(out))
	}
	panic("config: no sample form for " + v.Type().String())
}

// sampleDuration writes d the way a person would type it: 5m, 1h, 30s, 500ms.
// Duration.String pads whole units with a zero ("5m0s", "1h0m0s"); the zeros go.
func sampleDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}
