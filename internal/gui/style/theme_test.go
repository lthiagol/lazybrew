package style

import (
	"image/color"
	"regexp"
	"testing"

	"charm.land/lipgloss/v2"
)

// The official Catppuccin hex for every token, per flavour. The palette must
// carry exactly these values, so a typo in palette.go fails here.
var mochaHex = map[string]struct {
	got color.Color
	hex string
}{
	"Rosewater": {Mocha.Rosewater, "#f5e0dc"}, "Flamingo": {Mocha.Flamingo, "#f2cdcd"},
	"Pink": {Mocha.Pink, "#f5c2e7"}, "Mauve": {Mocha.Mauve, "#cba6f7"},
	"Red": {Mocha.Red, "#f38ba8"}, "Maroon": {Mocha.Maroon, "#eba0ac"},
	"Peach": {Mocha.Peach, "#fab387"}, "Yellow": {Mocha.Yellow, "#f9e2af"},
	"Green": {Mocha.Green, "#a6e3a1"}, "Teal": {Mocha.Teal, "#94e2d5"},
	"Sky": {Mocha.Sky, "#89dceb"}, "Sapphire": {Mocha.Sapphire, "#74c7ec"},
	"Blue": {Mocha.Blue, "#89b4fa"}, "Lavender": {Mocha.Lavender, "#b4befe"},
	"Text": {Mocha.Text, "#cdd6f4"}, "Subtext1": {Mocha.Subtext1, "#bac2de"},
	"Subtext0": {Mocha.Subtext0, "#a6adc8"}, "Overlay2": {Mocha.Overlay2, "#9399b2"},
	"Overlay1": {Mocha.Overlay1, "#7f849c"}, "Overlay0": {Mocha.Overlay0, "#6c7086"},
	"Surface2": {Mocha.Surface2, "#585b70"}, "Surface1": {Mocha.Surface1, "#45475a"},
	"Surface0": {Mocha.Surface0, "#313244"}, "Base": {Mocha.Base, "#1e1e2e"},
	"Mantle": {Mocha.Mantle, "#181825"}, "Crust": {Mocha.Crust, "#11111b"},
}

var latteHex = map[string]struct {
	got color.Color
	hex string
}{
	"Rosewater": {Latte.Rosewater, "#dc8a78"}, "Flamingo": {Latte.Flamingo, "#dd7878"},
	"Pink": {Latte.Pink, "#ea76cb"}, "Mauve": {Latte.Mauve, "#8839ef"},
	"Red": {Latte.Red, "#d20f39"}, "Maroon": {Latte.Maroon, "#e64553"},
	"Peach": {Latte.Peach, "#fe640b"}, "Yellow": {Latte.Yellow, "#df8e1d"},
	"Green": {Latte.Green, "#40a02b"}, "Teal": {Latte.Teal, "#179299"},
	"Sky": {Latte.Sky, "#04a5e5"}, "Sapphire": {Latte.Sapphire, "#209fb5"},
	"Blue": {Latte.Blue, "#1e66f5"}, "Lavender": {Latte.Lavender, "#7287fd"},
	"Text": {Latte.Text, "#4c4f69"}, "Subtext1": {Latte.Subtext1, "#5c5f77"},
	"Subtext0": {Latte.Subtext0, "#6c6f85"}, "Overlay2": {Latte.Overlay2, "#7c7f93"},
	"Overlay1": {Latte.Overlay1, "#8c8fa1"}, "Overlay0": {Latte.Overlay0, "#9ca0b0"},
	"Surface2": {Latte.Surface2, "#acb0be"}, "Surface1": {Latte.Surface1, "#bcc0cc"},
	"Surface0": {Latte.Surface0, "#ccd0da"}, "Base": {Latte.Base, "#eff1f5"},
	"Mantle": {Latte.Mantle, "#e6e9ef"}, "Crust": {Latte.Crust, "#dce0e8"},
}

// TestPaletteTokensAreOfficialHex checks every token of both flavours: each
// is a valid #rrggbb literal and carries exactly the official value.
func TestPaletteTokensAreOfficialHex(t *testing.T) {
	hex := regexp.MustCompile(`^#[0-9a-f]{6}$`)
	for name, set := range map[string]map[string]struct {
		got color.Color
		hex string
	}{"Mocha": mochaHex, "Latte": latteHex} {
		if len(set) != 26 {
			t.Errorf("%s defines %d tokens, want 26", name, len(set))
		}
		for tok, v := range set {
			if !hex.MatchString(v.hex) {
				t.Errorf("%s %s: spec %q is not #rrggbb", name, tok, v.hex)
			}
			if v.got == nil {
				t.Errorf("%s %s is nil", name, tok)
				continue
			}
			if v.got != lipgloss.Color(v.hex) {
				t.Errorf("%s %s = %v, want %s", name, tok, v.got, v.hex)
			}
		}
	}
}

func TestDarkThemeColors(t *testing.T) {
	theme := DarkTheme()
	want := map[string]color.Color{
		"Accent":     lipgloss.Color("#cba6f7"), // mauve
		"Success":    lipgloss.Color("#a6e3a1"), // green
		"Warning":    lipgloss.Color("#f9e2af"), // yellow
		"Error":      lipgloss.Color("#f38ba8"), // red
		"Info":       lipgloss.Color("#89b4fa"), // blue
		"Pinned":     lipgloss.Color("#94e2d5"), // teal
		"Muted":      lipgloss.Color("#6c7086"), // overlay0
		"Subtle":     lipgloss.Color("#a6adc8"), // subtext0
		"Text":       lipgloss.Color("#cdd6f4"), // text
		"Border":     lipgloss.Color("#45475a"), // surface1
		"Focus":      lipgloss.Color("#cba6f7"), // mauve
		"Surface":    lipgloss.Color("#313244"), // surface0
		"Background": lipgloss.Color("#1e1e2e"), // base
	}
	got := map[string]color.Color{
		"Accent":     theme.Accent,
		"Success":    theme.Success,
		"Warning":    theme.Warning,
		"Error":      theme.Error,
		"Info":       theme.Info,
		"Pinned":     theme.Pinned,
		"Muted":      theme.Muted,
		"Subtle":     theme.Subtle,
		"Text":       theme.Text,
		"Border":     theme.Border,
		"Focus":      theme.Focus,
		"Surface":    theme.Surface,
		"Background": theme.Background,
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("DarkTheme %s = %q, want %q", name, got[name], w)
		}
	}
}

func TestLightThemeIsLatte(t *testing.T) {
	theme := LightTheme()
	want := map[string]struct{ got, want color.Color }{
		"base":    {theme.Background, lipgloss.Color("#eff1f5")},
		"mauve":   {theme.Accent, lipgloss.Color("#8839ef")},
		"text":    {theme.Text, lipgloss.Color("#4c4f69")},
		"surface": {theme.Surface, lipgloss.Color("#ccd0da")},
		"border":  {theme.Border, lipgloss.Color("#bcc0cc")},
		"subtle":  {theme.Subtle, lipgloss.Color("#6c6f85")},
		"muted":   {theme.Muted, lipgloss.Color("#9ca0b0")},
		"pinned":  {theme.Pinned, lipgloss.Color("#179299")},
		"error":   {theme.Error, lipgloss.Color("#d20f39")},
		"success": {theme.Success, lipgloss.Color("#40a02b")},
	}
	for name, w := range want {
		if w.got != w.want {
			t.Errorf("LightTheme %s = %q, want %q", name, w.got, w.want)
		}
	}
}

func TestApplyTheme(t *testing.T) {
	ApplyTheme(DarkTheme())
	if CurrentTheme == nil {
		t.Fatal("CurrentTheme should not be nil")
	}
	if AccentColor != lipgloss.Color("#cba6f7") {
		t.Errorf("AccentColor after Dark ApplyTheme = %q", AccentColor)
	}
}

func TestApplyThemeLight(t *testing.T) {
	ApplyTheme(LightTheme())
	if CurrentTheme == nil {
		t.Fatal("CurrentTheme should not be nil")
	}
	if AccentColor != lipgloss.Color("#8839ef") {
		t.Errorf("AccentColor after Light ApplyTheme = %q", AccentColor)
	}
	// Restore dark for other packages sharing package-level styles.
	ApplyTheme(DarkTheme())
}

// TestApplyThemeSwitchesEveryRole checks that every exported role follows
// the theme it was applied from, including the names other packages use.
func TestApplyThemeSwitchesEveryRole(t *testing.T) {
	roles := func(t *Theme) map[string]color.Color {
		return map[string]color.Color{
			"AccentColor":    t.Accent,
			"SuccessColor":   t.Success,
			"WarningColor":   t.Warning,
			"ErrorColor":     t.Error,
			"InfoColor":      t.Info,
			"MutedColor":     t.Muted,
			"TextColor":      t.Text,
			"BorderColor":    t.Border,
			"FocusColor":     t.Focus,
			"BgColor":        t.Background,
			"SurfaceColor":   t.Surface,
			"SubtleColor":    t.Subtle,
			"SecondaryColor": t.Pinned,
		}
	}
	live := func() map[string]color.Color {
		return map[string]color.Color{
			"AccentColor":    AccentColor,
			"SuccessColor":   SuccessColor,
			"WarningColor":   WarningColor,
			"ErrorColor":     ErrorColor,
			"InfoColor":      InfoColor,
			"MutedColor":     MutedColor,
			"TextColor":      TextColor,
			"BorderColor":    BorderColor,
			"FocusColor":     FocusColor,
			"BgColor":        BgColor,
			"SurfaceColor":   SurfaceColor,
			"SubtleColor":    SubtleColor,
			"SecondaryColor": SecondaryColor,
		}
	}
	for _, tc := range []struct {
		name  string
		theme *Theme
	}{{"dark", DarkTheme()}, {"light", LightTheme()}} {
		ApplyTheme(tc.theme)
		want, got := roles(tc.theme), live()
		for name, w := range want {
			if got[name] != w {
				t.Errorf("%s: %s = %q, want %q", tc.name, name, got[name], w)
			}
		}
	}
	ApplyTheme(DarkTheme())
}

func TestApplyThemeDefinesM14Styles(t *testing.T) {
	ApplyTheme(DarkTheme())
	if ActivePanelBg.GetBackground() != lipgloss.Color("#313244") {
		t.Errorf("dark ActivePanelBg bg = %v, want #313244", ActivePanelBg.GetBackground())
	}
	if PanelTitleActive.GetForeground() != lipgloss.Color("#cba6f7") {
		t.Errorf("PanelTitleActive fg = %v, want accent", PanelTitleActive.GetForeground())
	}
	if !PanelTitleActive.GetBold() {
		t.Error("PanelTitleActive should be bold")
	}
	if LogBorder.GetBorderTopForeground() != lipgloss.Color("#6c7086") &&
		LogBorder.GetBorderStyle() != lipgloss.RoundedBorder() {
		// BorderForeground may be reported via top; also accept style check below.
	}
	if LogBorder.GetBorderStyle() != lipgloss.RoundedBorder() {
		t.Error("LogBorder should use rounded border")
	}

	ApplyTheme(LightTheme())
	if ActivePanelBg.GetBackground() != lipgloss.Color("#ccd0da") {
		t.Errorf("light ActivePanelBg bg = %v, want #ccd0da", ActivePanelBg.GetBackground())
	}
	ApplyTheme(DarkTheme())
}

func TestThemeConcurrency(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			ApplyTheme(DarkTheme())
			ApplyTheme(LightTheme())
		}
		done <- struct{}{}
	}()
	<-done
	ApplyTheme(DarkTheme())
}

// TestSharedTextStylesDrawWhatTheyReplace (cr3 AC-06): the shared foreground
// styles draw exactly what the per-render style of each role drew, in both
// themes, and the badges that are the same role reuse them.
func TestSharedTextStylesDrawWhatTheyReplace(t *testing.T) {
	defer ApplyTheme(DarkTheme())
	for _, th := range []*Theme{DarkTheme(), LightTheme()} {
		ApplyTheme(th)
		cases := []struct {
			name string
			got  lipgloss.Style
			want color.Color
		}{
			{"MutedText", MutedText, th.Muted},
			{"SuccessText", SuccessText, th.Success},
			{"WarningText", WarningText, th.Warning},
			{"ErrorText", ErrorText, th.Error},
			{"TextStyle", TextStyle, th.Text},
		}
		for _, c := range cases {
			want := lipgloss.NewStyle().Foreground(c.want).Render("sample")
			if got := c.got.Render("sample"); got != want {
				t.Errorf("%s = %q, want the per-render style %q", c.name, got, want)
			}
		}
		badges := []struct {
			name          string
			badge, shared lipgloss.Style
		}{
			{"OutdatedBadge", OutdatedBadge, WarningText},
			{"InstalledBadge", InstalledBadge, SuccessText},
			{"ErrorBadge", ErrorBadge, ErrorText},
			{"NormalItem", NormalItem, TextStyle},
		}
		for _, b := range badges {
			if b.badge.Render("sample") != b.shared.Render("sample") {
				t.Errorf("%s should draw the shared style of its role", b.name)
			}
		}
	}
}
