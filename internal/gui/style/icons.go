package style

// Icons is one glyph set. A field the set leaves empty draws nothing, so the
// zero Icons is the show_icons: false set: every renderer prefixes a glyph
// only when it is non-empty.
type Icons struct {
	App, Status, Formula, Cask, Outdated, Tap, Service, Search string
	Installed, Upgradable, Pinned, Running, Stopped            string
	Error, Warn, OK, Info, Marker, Refresh                     string
}

// NerdIcons draws with Nerd Font glyphs. They sit in the Font Awesome range
// of the Unicode Private Use Area, so they need a patched font; the escapes
// keep the source readable and the glyphs invisible in review.
var NerdIcons = Icons{
	App:        "", // beer
	Status:     "", // tachometer
	Formula:    "", // flask
	Cask:       "", // archive
	Outdated:   "", // arrow-circle-up
	Tap:        "", // code-fork
	Service:    "", // cog
	Search:     "", // search
	Installed:  "", // check-circle
	Upgradable: "", // arrow-up
	Pinned:     "", // thumb-tack
	Running:    "", // play-circle
	Stopped:    "", // stop-circle
	Error:      "", // times-circle
	Warn:       "", // exclamation-triangle
	OK:         "", // check
	Info:       "", // info-circle
	Marker:     "", // chevron-right
	Refresh:    "", // refresh
}

// UnicodeIcons draws with plain Unicode for terminals without a Nerd Font.
var UnicodeIcons = Icons{
	App:        "◆",
	Status:     "◉",
	Formula:    "∘",
	Cask:       "▣",
	Outdated:   "↑",
	Tap:        "⊕",
	Service:    "◎",
	Search:     "⌕",
	Installed:  "✓",
	Upgradable: "↑",
	Pinned:     "•",
	Running:    "▶",
	Stopped:    "■",
	Error:      "✗",
	Warn:       "⚠",
	OK:         "✓",
	Info:       "i",
	Marker:     "▸",
	Refresh:    "⟳",
}

// IconsNamed returns the set for a gui.icons value: "nerd" or "unicode".
// The second result is false for any other name.
func IconsNamed(name string) (Icons, bool) {
	switch name {
	case "nerd":
		return NerdIcons, true
	case "unicode":
		return UnicodeIcons, true
	default:
		return Icons{}, false
	}
}
