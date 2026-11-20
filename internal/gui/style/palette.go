package style

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette is one Catppuccin flavour: its named colours at the official hex
// values. Themes map semantic roles onto these tokens; the GUI draws with
// roles, never with tokens.
type Palette struct {
	Rosewater, Flamingo, Pink, Mauve, Red, Maroon, Peach, Yellow, Green,
	Teal, Sky, Sapphire, Blue, Lavender color.Color
	Text, Subtext1, Subtext0, Overlay2, Overlay1, Overlay0,
	Surface2, Surface1, Surface0, Base, Mantle, Crust color.Color
}

// Mocha is Catppuccin Mocha, the dark flavour.
var Mocha = Palette{
	Rosewater: lipgloss.Color("#f5e0dc"),
	Flamingo:  lipgloss.Color("#f2cdcd"),
	Pink:      lipgloss.Color("#f5c2e7"),
	Mauve:     lipgloss.Color("#cba6f7"),
	Red:       lipgloss.Color("#f38ba8"),
	Maroon:    lipgloss.Color("#eba0ac"),
	Peach:     lipgloss.Color("#fab387"),
	Yellow:    lipgloss.Color("#f9e2af"),
	Green:     lipgloss.Color("#a6e3a1"),
	Teal:      lipgloss.Color("#94e2d5"),
	Sky:       lipgloss.Color("#89dceb"),
	Sapphire:  lipgloss.Color("#74c7ec"),
	Blue:      lipgloss.Color("#89b4fa"),
	Lavender:  lipgloss.Color("#b4befe"),
	Text:      lipgloss.Color("#cdd6f4"),
	Subtext1:  lipgloss.Color("#bac2de"),
	Subtext0:  lipgloss.Color("#a6adc8"),
	Overlay2:  lipgloss.Color("#9399b2"),
	Overlay1:  lipgloss.Color("#7f849c"),
	Overlay0:  lipgloss.Color("#6c7086"),
	Surface2:  lipgloss.Color("#585b70"),
	Surface1:  lipgloss.Color("#45475a"),
	Surface0:  lipgloss.Color("#313244"),
	Base:      lipgloss.Color("#1e1e2e"),
	Mantle:    lipgloss.Color("#181825"),
	Crust:     lipgloss.Color("#11111b"),
}

// Latte is Catppuccin Latte, the light flavour.
var Latte = Palette{
	Rosewater: lipgloss.Color("#dc8a78"),
	Flamingo:  lipgloss.Color("#dd7878"),
	Pink:      lipgloss.Color("#ea76cb"),
	Mauve:     lipgloss.Color("#8839ef"),
	Red:       lipgloss.Color("#d20f39"),
	Maroon:    lipgloss.Color("#e64553"),
	Peach:     lipgloss.Color("#fe640b"),
	Yellow:    lipgloss.Color("#df8e1d"),
	Green:     lipgloss.Color("#40a02b"),
	Teal:      lipgloss.Color("#179299"),
	Sky:       lipgloss.Color("#04a5e5"),
	Sapphire:  lipgloss.Color("#209fb5"),
	Blue:      lipgloss.Color("#1e66f5"),
	Lavender:  lipgloss.Color("#7287fd"),
	Text:      lipgloss.Color("#4c4f69"),
	Subtext1:  lipgloss.Color("#5c5f77"),
	Subtext0:  lipgloss.Color("#6c6f85"),
	Overlay2:  lipgloss.Color("#7c7f93"),
	Overlay1:  lipgloss.Color("#8c8fa1"),
	Overlay0:  lipgloss.Color("#9ca0b0"),
	Surface2:  lipgloss.Color("#acb0be"),
	Surface1:  lipgloss.Color("#bcc0cc"),
	Surface0:  lipgloss.Color("#ccd0da"),
	Base:      lipgloss.Color("#eff1f5"),
	Mantle:    lipgloss.Color("#e6e9ef"),
	Crust:     lipgloss.Color("#dce0e8"),
}
