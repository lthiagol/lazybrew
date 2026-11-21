package style

import (
	"image/color"

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
