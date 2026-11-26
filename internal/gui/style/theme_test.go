package style

import (
	"image/color"
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/config"
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

var macchiatoHex = map[string]struct {
	got color.Color
	hex string
}{
	"Rosewater": {Macchiato.Rosewater, "#f4dbd6"}, "Flamingo": {Macchiato.Flamingo, "#f0c6c6"},
	"Pink": {Macchiato.Pink, "#f5bde6"}, "Mauve": {Macchiato.Mauve, "#c6a0f6"},
	"Red": {Macchiato.Red, "#ed8796"}, "Maroon": {Macchiato.Maroon, "#ee99a0"},
	"Peach": {Macchiato.Peach, "#f5a97f"}, "Yellow": {Macchiato.Yellow, "#eed49f"},
	"Green": {Macchiato.Green, "#a6da95"}, "Teal": {Macchiato.Teal, "#8bd5ca"},
	"Sky": {Macchiato.Sky, "#91d7e3"}, "Sapphire": {Macchiato.Sapphire, "#7dc4e4"},
	"Blue": {Macchiato.Blue, "#8aadf4"}, "Lavender": {Macchiato.Lavender, "#b7bdf8"},
	"Text": {Macchiato.Text, "#cad3f5"}, "Subtext1": {Macchiato.Subtext1, "#b8c0e0"},
	"Subtext0": {Macchiato.Subtext0, "#a5adcb"}, "Overlay2": {Macchiato.Overlay2, "#939ab7"},
	"Overlay1": {Macchiato.Overlay1, "#8087a2"}, "Overlay0": {Macchiato.Overlay0, "#6e738d"},
	"Surface2": {Macchiato.Surface2, "#5b6078"}, "Surface1": {Macchiato.Surface1, "#494d64"},
	"Surface0": {Macchiato.Surface0, "#363a4f"}, "Base": {Macchiato.Base, "#24273a"},
	"Mantle": {Macchiato.Mantle, "#1e2030"}, "Crust": {Macchiato.Crust, "#181926"},
}

var frappeHex = map[string]struct {
	got color.Color
	hex string
}{
	"Rosewater": {Frappe.Rosewater, "#f2d5cf"}, "Flamingo": {Frappe.Flamingo, "#eebebe"},
	"Pink": {Frappe.Pink, "#f4b8e4"}, "Mauve": {Frappe.Mauve, "#ca9ee6"},
	"Red": {Frappe.Red, "#e78284"}, "Maroon": {Frappe.Maroon, "#ea999c"},
	"Peach": {Frappe.Peach, "#ef9f76"}, "Yellow": {Frappe.Yellow, "#e5c890"},
	"Green": {Frappe.Green, "#a6d189"}, "Teal": {Frappe.Teal, "#81c8be"},
	"Sky": {Frappe.Sky, "#99d1db"}, "Sapphire": {Frappe.Sapphire, "#85c1dc"},
	"Blue": {Frappe.Blue, "#8caaee"}, "Lavender": {Frappe.Lavender, "#babbf1"},
	"Text": {Frappe.Text, "#c6d0f5"}, "Subtext1": {Frappe.Subtext1, "#b5bfe2"},
	"Subtext0": {Frappe.Subtext0, "#a5adce"}, "Overlay2": {Frappe.Overlay2, "#949cbb"},
	"Overlay1": {Frappe.Overlay1, "#838ba7"}, "Overlay0": {Frappe.Overlay0, "#737994"},
	"Surface2": {Frappe.Surface2, "#626880"}, "Surface1": {Frappe.Surface1, "#51576d"},
	"Surface0": {Frappe.Surface0, "#414559"}, "Base": {Frappe.Base, "#303446"},
	"Mantle": {Frappe.Mantle, "#292c3c"}, "Crust": {Frappe.Crust, "#232634"},
}

// TestPaletteTokensAreOfficialHex checks every token of both flavours: each
// is a valid #rrggbb literal and carries exactly the official value.
func TestPaletteTokensAreOfficialHex(t *testing.T) {
	hex := regexp.MustCompile(`^#[0-9a-f]{6}$`)
	for name, set := range map[string]map[string]struct {
		got color.Color
		hex string
	}{"Mocha": mochaHex, "Latte": latteHex, "Macchiato": macchiatoHex, "Frappe": frappeHex} {
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
		"pinned":  {theme.Pinned, lipgloss.Color("#179299")},
		"error":   {theme.Error, lipgloss.Color("#d20f39")},
	}
	for name, w := range want {
		if w.got != w.want {
			t.Errorf("LightTheme %s = %q, want %q", name, w.got, w.want)
		}
	}
}

// TestLightRolesAreTheChosenShades pins the five Latte roles that differ from
// their flavour's default: Muted and Subtle take official tokens, Success,
// Warning and Info take the darker shades (theme.go gives each ratio).
func TestLightRolesAreTheChosenShades(t *testing.T) {
	theme := LightTheme()
	want := map[string]struct{ got, want color.Color }{
		"muted":   {theme.Muted, lipgloss.Color("#7c7f93")},   // overlay2
		"subtle":  {theme.Subtle, lipgloss.Color("#5c5f77")},  // subtext1
		"success": {theme.Success, lipgloss.Color("#317a21")}, // green, darkened
		"warning": {theme.Warning, lipgloss.Color("#965f13")}, // yellow, darkened
		"info":    {theme.Info, lipgloss.Color("#135ef5")},    // blue, darkened
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
	if SurfaceColor != lipgloss.Color("#313244") {
		t.Errorf("dark active sidebar row surface = %v, want #313244", SurfaceColor)
	}
	if PanelTitleActive.GetForeground() != lipgloss.Color("#cba6f7") {
		t.Errorf("PanelTitleActive fg = %v, want accent", PanelTitleActive.GetForeground())
	}
	if !PanelTitleActive.GetBold() {
		t.Error("PanelTitleActive should be bold")
	}

	ApplyTheme(LightTheme())
	if SurfaceColor != lipgloss.Color("#ccd0da") {
		t.Errorf("light active sidebar row surface = %v, want #ccd0da", SurfaceColor)
	}
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

// roleSet is every role of a theme, by its name.
func roleSet(t *Theme) map[string]color.Color {
	return map[string]color.Color{
		"Accent":     t.Accent,
		"Success":    t.Success,
		"Warning":    t.Warning,
		"Error":      t.Error,
		"Info":       t.Info,
		"Pinned":     t.Pinned,
		"Muted":      t.Muted,
		"Subtle":     t.Subtle,
		"Text":       t.Text,
		"Border":     t.Border,
		"Focus":      t.Focus,
		"Surface":    t.Surface,
		"Background": t.Background,
	}
}

// registryNames lists the names in the registry, in order.
func registryNames() []string {
	names := make([]string, len(themes))
	for i, t := range themes {
		names[i] = t.name
	}
	return names
}

// TestEveryThemeSetsEveryRole: each registered theme has all 13 roles non-nil.
func TestEveryThemeSetsEveryRole(t *testing.T) {
	for _, name := range registryNames() {
		th, ok := ThemeByName(name)
		if !ok {
			t.Fatalf("ThemeByName(%q) not found", name)
		}
		for role, c := range roleSet(th) {
			if c == nil {
				t.Errorf("%s: role %s is nil", name, role)
			}
		}
	}
}

// TestThemeRegistryNames: the names gui.theme accepts, in display order, and
// each one resolves; the lookup ignores case; an unknown name is not found.
func TestThemeRegistryNames(t *testing.T) {
	want := []string{
		"dark", "light",
		"catppuccin-mocha", "catppuccin-macchiato", "catppuccin-frappe", "catppuccin-latte",
		"dracula", "nord",
	}
	if got := registryNames(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("registry names = %v, want %v", got, want)
	}
	for _, name := range want {
		if _, ok := ThemeByName(name); !ok {
			t.Errorf("ThemeByName(%q) not found", name)
		}
		if _, ok := ThemeByName(strings.ToUpper(name)); !ok {
			t.Errorf("ThemeByName(%q) should ignore case", strings.ToUpper(name))
		}
	}
	if _, ok := ThemeByName("solarized"); ok {
		t.Error("ThemeByName(solarized) should not be found")
	}
}

// TestRegistryMatchesConfig: the names gui.theme accepts (config) and the
// names the registry resolves are the same list, in the same order.
func TestRegistryMatchesConfig(t *testing.T) {
	if got, want := strings.Join(registryNames(), ","), strings.Join(config.ThemeNames(), ","); got != want {
		t.Errorf("registry names = %s, want config's %s", got, want)
	}
}

// TestDarkAndLightResolveToMochaAndLatte: the old names keep exactly the
// themes they had, and the Catppuccin names resolve to their flavours.
func TestDarkAndLightResolveToMochaAndLatte(t *testing.T) {
	cases := map[string]*Theme{
		"dark":                 DarkTheme(),
		"catppuccin-mocha":     DarkTheme(),
		"light":                LightTheme(),
		"catppuccin-latte":     LightTheme(),
		"catppuccin-frappe":    FrappeTheme(),
		"catppuccin-macchiato": MacchiatoTheme(),
	}
	for name, want := range cases {
		got, ok := ThemeByName(name)
		if !ok {
			t.Fatalf("ThemeByName(%q) not found", name)
		}
		if *got != *want {
			t.Errorf("ThemeByName(%q) = %+v, want %+v", name, *got, *want)
		}
	}
}

// TestCatppuccinConstructorsFollowTheirPalette: each flavour's theme maps the
// same roles as Mocha does, from its own palette. Frappé is the one exception:
// its Muted is overlay1 (see FrappeTheme).
func TestCatppuccinConstructorsFollowTheirPalette(t *testing.T) {
	if *MacchiatoTheme() != *themeFrom(Macchiato) {
		t.Error("Macchiato theme does not map its palette's roles")
	}
	want := themeFrom(Frappe)
	want.Muted = Frappe.Overlay1
	if *FrappeTheme() != *want {
		t.Error("Frappé theme does not map its palette's roles")
	}
}

// TestDraculaRoles and TestNordRoles pin each role to the colour the comment
// table in theme.go gives it: the official value, or the derived one.
func TestDraculaRoles(t *testing.T) {
	want := map[string]string{
		"Accent":     "#bd93f9", // purple
		"Success":    "#50fa7b", // green
		"Warning":    "#ffb86c", // orange
		"Error":      "#ff5555", // red
		"Info":       "#8be9fd", // cyan
		"Pinned":     "#ff79c6", // pink
		"Muted":      "#6272a4", // comment
		"Subtle":     "#bababa", // derived: foreground at 70 % over background
		"Text":       "#f8f8f2", // foreground
		"Border":     "#6272a4", // comment
		"Focus":      "#bd93f9", // purple
		"Surface":    "#44475a", // selection
		"Background": "#282a36", // background
	}
	checkRoles(t, "Dracula", DraculaTheme(), want)
}

func TestNordRoles(t *testing.T) {
	want := map[string]string{
		"Accent":     "#88c0d0", // nord8
		"Success":    "#a3be8c", // nord14
		"Warning":    "#ebcb8b", // nord13
		"Error":      "#d19aa1", // derived: nord11 at 60 % over nord6
		"Info":       "#81a1c1", // nord9
		"Pinned":     "#b48ead", // nord15
		"Muted":      "#848c9d", // derived: nord3 at 60 % over nord4
		"Subtle":     "#d8dee9", // nord4
		"Text":       "#eceff4", // nord6
		"Border":     "#4c566a", // nord3
		"Focus":      "#88c0d0", // nord8
		"Surface":    "#3b4252", // nord1
		"Background": "#2e3440", // nord0
	}
	checkRoles(t, "Nord", NordTheme(), want)
}

// checkRoles compares each role's colour with the hex it must carry.
func checkRoles(t *testing.T, name string, th *Theme, want map[string]string) {
	t.Helper()
	got := roleSet(th)
	for role, hex := range want {
		if g := got[role]; g != lipgloss.Color(hex) {
			t.Errorf("%s %s = %v, want %s", name, role, g, hex)
		}
	}
}
