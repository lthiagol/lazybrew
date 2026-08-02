package style

import "github.com/charmbracelet/lipgloss"

var (
	CurrentTheme   *Theme
	AccentColor    lipgloss.Color
	SecondaryColor lipgloss.Color
	SuccessColor   lipgloss.Color
	WarningColor   lipgloss.Color
	ErrorColor     lipgloss.Color
	SubtleColor    lipgloss.Color
	TextColor      lipgloss.Color
	BgColor        lipgloss.Color

	ActiveBorder     lipgloss.Style
	InactiveBorder   lipgloss.Style
	SelectedItem     lipgloss.Style
	NormalItem       lipgloss.Style
	TabActive        lipgloss.Style
	TabInactive      lipgloss.Style
	HintKey          lipgloss.Style
	HintDesc         lipgloss.Style
	PanelTitle       lipgloss.Style
	PanelTitleActive lipgloss.Style
	ActivePanelBg    lipgloss.Style
	LogBorder        lipgloss.Style
	OutdatedBadge    lipgloss.Style
	PinnedBadge      lipgloss.Style
	InstalledBadge   lipgloss.Style
	ErrorBadge       lipgloss.Style
	SubtleText       lipgloss.Style
	AccentText       lipgloss.Style
	DocStyle         lipgloss.Style
)

type Theme struct {
	AccentColor    lipgloss.Color
	SecondaryColor lipgloss.Color
	SuccessColor   lipgloss.Color
	WarningColor   lipgloss.Color
	ErrorColor     lipgloss.Color
	SubtleColor    lipgloss.Color
	TextColor      lipgloss.Color
	BgColor        lipgloss.Color
}

// DarkTheme is Catppuccin Mocha.
func DarkTheme() *Theme {
	return &Theme{
		AccentColor:    lipgloss.Color("#cba6f7"),
		SecondaryColor: lipgloss.Color("#94e2d5"),
		SuccessColor:   lipgloss.Color("#a6e3a1"),
		WarningColor:   lipgloss.Color("#f9e2af"),
		ErrorColor:     lipgloss.Color("#f38ba8"),
		SubtleColor:    lipgloss.Color("#6c7086"),
		TextColor:      lipgloss.Color("#cdd6f4"),
		BgColor:        lipgloss.Color("#1e1e2e"),
	}
}

// LightTheme keeps pre-M14 hex values (not Catppuccin Latte).
func LightTheme() *Theme {
	return &Theme{
		AccentColor:    lipgloss.Color("#34548a"),
		SecondaryColor: lipgloss.Color("#0891b2"),
		SuccessColor:   lipgloss.Color("#33635c"),
		WarningColor:   lipgloss.Color("#8f5e15"),
		ErrorColor:     lipgloss.Color("#8c4351"),
		SubtleColor:    lipgloss.Color("#9699a3"),
		TextColor:      lipgloss.Color("#343b58"),
		BgColor:        lipgloss.Color("#f5f5f5"),
	}
}

func ApplyTheme(t *Theme) {
	CurrentTheme = t
	AccentColor = t.AccentColor
	SecondaryColor = t.SecondaryColor
	SuccessColor = t.SuccessColor
	WarningColor = t.WarningColor
	ErrorColor = t.ErrorColor
	SubtleColor = t.SubtleColor
	TextColor = t.TextColor
	BgColor = t.BgColor

	surface0 := lipgloss.Color("#313244")
	if t.BgColor == LightTheme().BgColor {
		surface0 = lipgloss.Color("#ccd0da")
	}

	ActiveBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.AccentColor)
	InactiveBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.SubtleColor)
	SelectedItem = lipgloss.NewStyle().Foreground(t.AccentColor).Bold(true)
	NormalItem = lipgloss.NewStyle().Foreground(t.TextColor)
	TabActive = lipgloss.NewStyle().Foreground(t.AccentColor).Bold(true).Underline(true)
	TabInactive = lipgloss.NewStyle().Foreground(t.SubtleColor)
	HintKey = lipgloss.NewStyle().Foreground(t.AccentColor).Bold(true).Padding(0, 1)
	HintDesc = lipgloss.NewStyle().Foreground(t.SubtleColor)
	PanelTitle = lipgloss.NewStyle().Foreground(t.TextColor).Bold(true)
	PanelTitleActive = lipgloss.NewStyle().Foreground(t.AccentColor).Bold(true)
	ActivePanelBg = lipgloss.NewStyle().Background(surface0)
	LogBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.SubtleColor)
	OutdatedBadge = lipgloss.NewStyle().Foreground(t.WarningColor)
	PinnedBadge = lipgloss.NewStyle().Foreground(t.SecondaryColor)
	InstalledBadge = lipgloss.NewStyle().Foreground(t.SuccessColor)
	ErrorBadge = lipgloss.NewStyle().Foreground(t.ErrorColor)
	SubtleText = lipgloss.NewStyle().Foreground(t.SubtleColor)
	AccentText = lipgloss.NewStyle().Foreground(t.AccentColor).Bold(true)
	DocStyle = lipgloss.NewStyle().Padding(1, 2)
}

func init() {
	ApplyTheme(DarkTheme())
}
