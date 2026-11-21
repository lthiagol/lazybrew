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
	"gui.theme":                "dark or light",
	"gui.sidebar_width":        "percent of the terminal width, 15-50",
	"gui.show_icons":           "true or false; false draws no icons",
	"gui.icons":                "nerd or unicode (no Nerd Font needed)",
	"gui.mouse":                "true or false",
	"gui.animations":           "true or false; motion, short and never delays a key",
	"gui.preload":              "true or false; load every panel in the background",
	"gui.auto_refresh_seconds": "0 is off; N refreshes the data every N seconds",
	"gui.auto_refresh_pause":   "a duration, e.g. 500ms; 0 turns the pause off",
	"brew":                     "how lazybrew runs brew",
	"brew.path":                "the brew binary; empty auto-detects it",
	"brew.update_on_start":     "auto, true or false: when lazybrew runs brew update",
	"brew.update_stale_hours":  "hours; auto updates when brew's last update is older",
	"brew.formulae_ttl":        "a duration, e.g. 5m; how long formula results are reused",
	"brew.casks_ttl":           "a duration, e.g. 5m; how long cask results are reused",
	"brew.outdated_ttl":        "a duration, e.g. 30m; how long the outdated list is reused",
	"brew.taps_ttl":            "a duration, e.g. 1h; how long tap results are reused",
	"brew.services_ttl":        "a duration, e.g. 30s; how long service results are reused",
	"brew.doctor_ttl":          "a duration, e.g. 1h; how long doctor output is reused",
}

var durationType = reflect.TypeOf(time.Duration(0))

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
#   %s=nerd|unicode   gui.icons
#   %s=0|1            gui.animations
#   %s=0|1            gui.preload

`, IconsEnv, AnimationsEnv, PreloadEnv)
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
	}
}

// sampleValue writes a leaf value the way it reads in a config file.
func sampleValue(v reflect.Value) string {
	if v.Type() == durationType {
		return sampleDuration(time.Duration(v.Int()))
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
