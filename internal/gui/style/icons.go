package style

// Icons is one glyph set. A field the set leaves empty draws nothing, so the
// zero Icons is the show_icons: false set: every renderer prefixes a glyph
// only when it is non-empty.
type Icons struct {
	App, Status, Formula, Cask, Outdated, Tap, Service, Search string
	Installed, Absent, Upgradable, Pinned, Running, Stopped    string
	Error, Warn, OK, Info, Marker, Refresh                     string
}

// NerdIcons draws with Nerd Font glyphs. They sit in the Font Awesome range
// of the Unicode Private Use Area, so they need a patched font; the escapes
// keep the source readable and the glyphs invisible in review.
var NerdIcons = Icons{
	App:        "\uf0fc", // beer
	Status:     "\uf0e4", // tachometer
	Formula:    "\uf0c3", // flask
	Cask:       "\uf187", // archive
	Outdated:   "\uf0aa", // arrow-circle-up
	Tap:        "\uf126", // code-fork
	Service:    "\uf013", // cog
	Search:     "\uf002", // search
	Installed:  "\uf058", // check-circle
	Absent:     "\uf10c", // circle-o
	Upgradable: "\uf062", // arrow-up
	Pinned:     "\uf08d", // thumb-tack
	Running:    "\uf144", // play-circle
	Stopped:    "\uf28d", // stop-circle
	Error:      "\uf057", // times-circle
	Warn:       "\uf071", // exclamation-triangle
	OK:         "\uf00c", // check
	Info:       "\uf05a", // info-circle
	Marker:     "\uf054", // chevron-right
	Refresh:    "\uf021", // refresh
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
	Absent:     "○",
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
