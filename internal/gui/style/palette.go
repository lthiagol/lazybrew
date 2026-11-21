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

// Macchiato is Catppuccin Macchiato, a dark flavour.
var Macchiato = Palette{
	Rosewater: lipgloss.Color("#f4dbd6"),
	Flamingo:  lipgloss.Color("#f0c6c6"),
	Pink:      lipgloss.Color("#f5bde6"),
	Mauve:     lipgloss.Color("#c6a0f6"),
	Red:       lipgloss.Color("#ed8796"),
	Maroon:    lipgloss.Color("#ee99a0"),
	Peach:     lipgloss.Color("#f5a97f"),
	Yellow:    lipgloss.Color("#eed49f"),
	Green:     lipgloss.Color("#a6da95"),
	Teal:      lipgloss.Color("#8bd5ca"),
	Sky:       lipgloss.Color("#91d7e3"),
	Sapphire:  lipgloss.Color("#7dc4e4"),
	Blue:      lipgloss.Color("#8aadf4"),
	Lavender:  lipgloss.Color("#b7bdf8"),
	Text:      lipgloss.Color("#cad3f5"),
	Subtext1:  lipgloss.Color("#b8c0e0"),
	Subtext0:  lipgloss.Color("#a5adcb"),
	Overlay2:  lipgloss.Color("#939ab7"),
	Overlay1:  lipgloss.Color("#8087a2"),
	Overlay0:  lipgloss.Color("#6e738d"),
	Surface2:  lipgloss.Color("#5b6078"),
	Surface1:  lipgloss.Color("#494d64"),
	Surface0:  lipgloss.Color("#363a4f"),
	Base:      lipgloss.Color("#24273a"),
	Mantle:    lipgloss.Color("#1e2030"),
	Crust:     lipgloss.Color("#181926"),
}

// Frappe is Catppuccin Frappé, a dark flavour.
var Frappe = Palette{
	Rosewater: lipgloss.Color("#f2d5cf"),
	Flamingo:  lipgloss.Color("#eebebe"),
	Pink:      lipgloss.Color("#f4b8e4"),
	Mauve:     lipgloss.Color("#ca9ee6"),
	Red:       lipgloss.Color("#e78284"),
	Maroon:    lipgloss.Color("#ea999c"),
	Peach:     lipgloss.Color("#ef9f76"),
	Yellow:    lipgloss.Color("#e5c890"),
	Green:     lipgloss.Color("#a6d189"),
	Teal:      lipgloss.Color("#81c8be"),
	Sky:       lipgloss.Color("#99d1db"),
	Sapphire:  lipgloss.Color("#85c1dc"),
	Blue:      lipgloss.Color("#8caaee"),
	Lavender:  lipgloss.Color("#babbf1"),
	Text:      lipgloss.Color("#c6d0f5"),
	Subtext1:  lipgloss.Color("#b5bfe2"),
	Subtext0:  lipgloss.Color("#a5adce"),
	Overlay2:  lipgloss.Color("#949cbb"),
	Overlay1:  lipgloss.Color("#838ba7"),
	Overlay0:  lipgloss.Color("#737994"),
	Surface2:  lipgloss.Color("#626880"),
	Surface1:  lipgloss.Color("#51576d"),
	Surface0:  lipgloss.Color("#414559"),
	Base:      lipgloss.Color("#303446"),
	Mantle:    lipgloss.Color("#292c3c"),
	Crust:     lipgloss.Color("#232634"),
}
