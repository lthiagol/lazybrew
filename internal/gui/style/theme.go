package style

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	CurrentTheme *Theme

	// Semantic roles. The GUI draws with these; the names below are the
	// exported roles other packages already use, mapped onto the same values.
	AccentColor  color.Color
	SuccessColor color.Color
	WarningColor color.Color
	ErrorColor   color.Color
	InfoColor    color.Color
	MutedColor   color.Color
	TextColor    color.Color
	BorderColor  color.Color
	FocusColor   color.Color
	BgColor      color.Color
	SurfaceColor color.Color
	// SubtleColor is the Subtle text role (subtext0).
	SubtleColor color.Color
	// SecondaryColor is the Pinned role (teal), kept under its old name.
	SecondaryColor color.Color

	SelectedItem lipgloss.Style
	NormalItem   lipgloss.Style
	// The foreground styles of the text roles, built once per theme: a render
	// reuses them rather than building one per row (cr3 AC-06).
	TextStyle        lipgloss.Style
	MutedText        lipgloss.Style
	SuccessText      lipgloss.Style
	WarningText      lipgloss.Style
	ErrorText        lipgloss.Style
	TabActive        lipgloss.Style
	TabInactive      lipgloss.Style
	PanelTitleActive lipgloss.Style
	OutdatedBadge    lipgloss.Style
	PinnedBadge      lipgloss.Style
	InstalledBadge   lipgloss.Style
	ErrorBadge       lipgloss.Style
	SubtleText       lipgloss.Style
	AccentText       lipgloss.Style
)

// Theme maps the semantic roles onto one palette. Each theme carries its own
// surface token, so no theme is detected by comparing its background.
type Theme struct {
	Accent     color.Color // focus and selection
	Success    color.Color
	Warning    color.Color
	Error      color.Color
	Info       color.Color
	Pinned     color.Color
	Muted      color.Color // de-emphasised text
	Subtle     color.Color // secondary text
	Text       color.Color
	Border     color.Color // unfocused borders
	Focus      color.Color // focused borders and titles
	Surface    color.Color // the active sidebar row's background
	Background color.Color
}

// themeFrom assigns the roles for one flavour.
func themeFrom(p Palette) *Theme {
	return &Theme{
		Accent:     p.Mauve,
		Success:    p.Green,
		Warning:    p.Yellow,
		Error:      p.Red,
		Info:       p.Blue,
		Pinned:     p.Teal,
		Muted:      p.Overlay0,
		Subtle:     p.Subtext0,
		Text:       p.Text,
		Border:     p.Surface1,
		Focus:      p.Mauve,
		Surface:    p.Surface0,
		Background: p.Base,
	}
}

// DarkTheme is Catppuccin Mocha.
func DarkTheme() *Theme { return themeFrom(Mocha) }

// LightTheme is Catppuccin Latte.
func LightTheme() *Theme { return themeFrom(Latte) }

// MacchiatoTheme is Catppuccin Macchiato.
func MacchiatoTheme() *Theme { return themeFrom(Macchiato) }

// FrappeTheme is Catppuccin Frappé. Its Muted is overlay1, not overlay0:
// overlay0 is 2.87:1 on the base, under the 3:1 floor; overlay1 is 3.65:1.
func FrappeTheme() *Theme {
	t := themeFrom(Frappe)
	t.Muted = Frappe.Overlay1
	return t
}

// DraculaTheme is Dracula, from its official palette (draculatheme.com).
// Its roles are not Catppuccin-shaped, so each is set here, with the colour
// and the reason:
//
//	role        colour    why
//	Accent      #bd93f9   purple, the focus and selection colour
//	Success     #50fa7b   green
//	Warning     #ffb86c   orange, 8.4:1 on the background
//	Error       #ff5555   red; 4.53:1 on the background, so it clears 4.5:1
//	Info        #8be9fd   cyan
//	Pinned      #ff79c6   pink
//	Muted       #6272a4   comment; 3.03:1 on the background, so it clears 3:1
//	Subtle      #bababa   derived: foreground at 70 % over the background
//	                      (7.3:1). Dracula has no dimmer text colour.
//	Text        #f8f8f2   foreground
//	Border      #6272a4   comment; unfocused borders
//	Focus       #bd93f9   purple, as the accent
//	Surface     #44475a   selection, the active sidebar row
//	Background  #282a36   background
func DraculaTheme() *Theme {
	return &Theme{
		Accent:     lipgloss.Color("#bd93f9"),
		Success:    lipgloss.Color("#50fa7b"),
		Warning:    lipgloss.Color("#ffb86c"),
		Error:      lipgloss.Color("#ff5555"),
		Info:       lipgloss.Color("#8be9fd"),
		Pinned:     lipgloss.Color("#ff79c6"),
		Muted:      lipgloss.Color("#6272a4"),
		Subtle:     lipgloss.Color("#bababa"),
		Text:       lipgloss.Color("#f8f8f2"),
		Border:     lipgloss.Color("#6272a4"),
		Focus:      lipgloss.Color("#bd93f9"),
		Surface:    lipgloss.Color("#44475a"),
		Background: lipgloss.Color("#282a36"),
	}
}

// NordTheme is Nord, from its official palette (nordtheme.com). As with
// Dracula, each role is set here:
//
//	role        colour    why
//	Accent      #88c0d0   nord8, the focus and selection colour
//	Success     #a3be8c   nord14
//	Warning     #ebcb8b   nord13
//	Error       #d19aa1   derived: nord11 at 60 % over nord6. Official nord11
//	                      is 3.05:1 on the background; this is 5.26:1. (At 70 %
//	                      it would be 4.6:1, too close to the 4.5:1 bound.)
//	Info        #81a1c1   nord9; 4.64:1 on the background
//	Pinned      #b48ead   nord15
//	Muted       #848c9d   derived: nord3 at 60 % over nord4. Official nord3 is
//	                      1.69:1 on the background; this is 3.7:1.
//	Subtle      #d8dee9   nord4
//	Text        #eceff4   nord6
//	Border      #4c566a   nord3; unfocused borders
//	Focus       #88c0d0   nord8, as the accent
//	Surface     #3b4252   nord1, the active sidebar row
//	Background  #2e3440   nord0
func NordTheme() *Theme {
	return &Theme{
		Accent:     lipgloss.Color("#88c0d0"),
		Success:    lipgloss.Color("#a3be8c"),
		Warning:    lipgloss.Color("#ebcb8b"),
		Error:      lipgloss.Color("#d19aa1"),
		Info:       lipgloss.Color("#81a1c1"),
		Pinned:     lipgloss.Color("#b48ead"),
		Muted:      lipgloss.Color("#848c9d"),
		Subtle:     lipgloss.Color("#d8dee9"),
		Text:       lipgloss.Color("#eceff4"),
		Border:     lipgloss.Color("#4c566a"),
		Focus:      lipgloss.Color("#88c0d0"),
		Surface:    lipgloss.Color("#3b4252"),
		Background: lipgloss.Color("#2e3440"),
	}
}

// themes maps each name gui.theme accepts to its constructor, so every lookup
// gets its own copy of the theme. The names are listed twice: here, and in
// config.themeNames. TestRegistryMatchesConfig ties the two lists together.
var themes = []struct {
	name  string
	theme func() *Theme
}{
	{"dark", DarkTheme},
	{"light", LightTheme},
	{"catppuccin-mocha", DarkTheme},
	{"catppuccin-macchiato", MacchiatoTheme},
	{"catppuccin-frappe", FrappeTheme},
	{"catppuccin-latte", LightTheme},
	{"dracula", DraculaTheme},
	{"nord", NordTheme},
}

// ThemeByName returns the theme for a name, matched without regard to case.
func ThemeByName(name string) (*Theme, bool) {
	for _, t := range themes {
		if strings.EqualFold(t.name, name) {
			return t.theme(), true
		}
	}
	return nil, false
}

func ApplyTheme(t *Theme) {
	CurrentTheme = t
	AccentColor = t.Accent
	SuccessColor = t.Success
	WarningColor = t.Warning
	ErrorColor = t.Error
	InfoColor = t.Info
	MutedColor = t.Muted
	TextColor = t.Text
	BorderColor = t.Border
	FocusColor = t.Focus
	BgColor = t.Background
	SurfaceColor = t.Surface
	SubtleColor = t.Subtle
	SecondaryColor = t.Pinned

	SelectedItem = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	TextStyle = lipgloss.NewStyle().Foreground(t.Text)
	MutedText = lipgloss.NewStyle().Foreground(t.Muted)
	SuccessText = lipgloss.NewStyle().Foreground(t.Success)
	WarningText = lipgloss.NewStyle().Foreground(t.Warning)
	ErrorText = lipgloss.NewStyle().Foreground(t.Error)
	NormalItem = TextStyle
	TabActive = lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true).
		Border(lipgloss.Border{Bottom: "━"}, false, false, true, false).
		BorderForeground(t.Accent).
		Padding(0, 1)
	TabInactive = lipgloss.NewStyle().
		Foreground(t.Subtle).
		Padding(0, 1)
	PanelTitleActive = lipgloss.NewStyle().Foreground(t.Focus).Bold(true)
	OutdatedBadge = WarningText
	PinnedBadge = lipgloss.NewStyle().Foreground(t.Pinned)
	InstalledBadge = SuccessText
	ErrorBadge = ErrorText
	SubtleText = lipgloss.NewStyle().Foreground(t.Subtle)
	AccentText = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
}

func init() {
	ApplyTheme(DarkTheme())
}
