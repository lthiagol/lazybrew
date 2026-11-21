package gui

import (
	"reflect"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// inPrivateUse reports whether r is in a Unicode Private Use Area: the BMP
// block and the two supplementary planes.
func inPrivateUse(r rune) bool {
	return (r >= 0xE000 && r <= 0xF8FF) ||
		(r >= 0xF0000 && r <= 0xFFFFD) ||
		(r >= 0x100000 && r <= 0x10FFFD)
}

// setFields lists the glyph of every field of an icon set by field name.
func setFields(t *testing.T, icons style.Icons) map[string]string {
	t.Helper()
	v := reflect.ValueOf(icons)
	out := make(map[string]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		out[v.Type().Field(i).Name] = v.Field(i).String()
	}
	return out
}

func TestIconSetsFillEveryField(t *testing.T) {
	for name, icons := range map[string]style.Icons{"nerd": style.NerdIcons, "unicode": style.UnicodeIcons} {
		for field, glyph := range setFields(t, icons) {
			if glyph == "" {
				t.Errorf("%s set: field %s is empty", name, field)
			}
		}
	}
}

// TestInstalledGlyphIsBrewsCheckmark (lb25 AC-01): the installed mark is the
// ✔ brew prints, in both sets, and it is one cell wide so rows stay aligned.
func TestInstalledGlyphIsBrewsCheckmark(t *testing.T) {
	for name, icons := range map[string]style.Icons{"nerd": style.NerdIcons, "unicode": style.UnicodeIcons} {
		if icons.Installed != "✔" {
			t.Errorf("%s set: Installed = %q, want ✔ (U+2714)", name, icons.Installed)
		}
		if w := lipgloss.Width(icons.Installed); w != 1 {
			t.Errorf("%s set: Installed is %d cells wide, want 1", name, w)
		}
	}
}

// TestOKGlyphIsBrewsCheckmark (lb25 fix r2, Q-lb25-ok-uses): the success mark
// is the same ✔ as the installed mark, in both sets, one cell wide.
func TestOKGlyphIsBrewsCheckmark(t *testing.T) {
	for name, icons := range map[string]style.Icons{"nerd": style.NerdIcons, "unicode": style.UnicodeIcons} {
		if icons.OK != "✔" {
			t.Errorf("%s set: OK = %q, want ✔ (U+2714)", name, icons.OK)
		}
		if w := lipgloss.Width(icons.OK); w != 1 {
			t.Errorf("%s set: OK is %d cells wide, want 1", name, w)
		}
	}
}

func TestNerdGlyphsArePrivateUse(t *testing.T) {
	for field, glyph := range setFields(t, style.NerdIcons) {
		// Installed and OK are brew's ✔ in both sets (lb25 AC-01, r2), not Nerd glyphs.
		if field == "Installed" || field == "OK" {
			continue
		}
		for _, r := range glyph {
			if !inPrivateUse(r) {
				t.Errorf("nerd field %s = U+%04X, want a Private Use Area glyph", field, r)
			}
		}
	}
}

func TestUnicodeGlyphsAreNotPrivateUse(t *testing.T) {
	for field, glyph := range setFields(t, style.UnicodeIcons) {
		for _, r := range glyph {
			if inPrivateUse(r) {
				t.Errorf("unicode field %s = U+%04X, a Private Use Area glyph needs a Nerd Font", field, r)
			}
		}
	}
}

// TestIconSelectionMatrix covers gui.icons, LAZYBREW_ICONS and show_icons in
// every combination that changes the set a model draws with.
func TestIconSelectionMatrix(t *testing.T) {
	tests := []struct {
		name  string
		icons string
		show  bool
		env   string
		want  style.Icons
	}{
		{"default is nerd", config.IconsNerd, true, "", style.NerdIcons},
		{"config unicode", config.IconsUnicode, true, "", style.UnicodeIcons},
		{"env overrides config", config.IconsNerd, true, config.IconsUnicode, style.UnicodeIcons},
		{"env nerd overrides config unicode", config.IconsUnicode, true, config.IconsNerd, style.NerdIcons},
		{"show_icons off draws nothing", config.IconsNerd, false, "", style.Icons{}},
		{"show_icons off beats env", config.IconsNerd, false, config.IconsUnicode, style.Icons{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(config.IconsEnv, tc.env)
			cfg := testConfig()
			cfg.GUI.Icons = tc.icons
			cfg.GUI.ShowIcons = tc.show
			if got := iconSet(cfg); got != tc.want {
				t.Errorf("iconSet = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestWithIconLeavesNoStraySpace(t *testing.T) {
	if got := presentation.WithIcon("", "bash"); got != "bash" {
		t.Errorf("withIcon with no glyph = %q, want %q", got, "bash")
	}
	if got := presentation.WithIcon("◆", "bash"); got != "◆ bash" {
		t.Errorf("withIcon = %q, want %q", got, "◆ bash")
	}
}

// TestModelResolvesIconsOnce checks that New resolves the set into the model,
// so renderers read m.icons and never the environment.
func TestModelResolvesIconsOnce(t *testing.T) {
	t.Setenv(config.IconsEnv, config.IconsNerd)
	m := New(nil, testConfig())
	if m.icons != style.NerdIcons {
		t.Errorf("model icons = %+v, want the nerd set", m.icons)
	}
}
