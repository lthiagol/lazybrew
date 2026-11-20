package gui

import (
	"os"

	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// iconSet resolves the glyphs a model draws with. gui.show_icons: false draws
// none. Otherwise gui.icons picks the set, and LAZYBREW_ICONS overrides it.
// config.Load has already rejected unknown names, so the nerd fallback only
// serves a config built in code.
func iconSet(cfg *config.Config) style.Icons {
	if !cfg.GUI.ShowIcons {
		return style.Icons{}
	}
	name := cfg.GUI.Icons
	if env := os.Getenv(config.IconsEnv); env != "" {
		name = env
	}
	icons, ok := style.IconsNamed(name)
	if !ok {
		return style.NerdIcons
	}
	return icons
}

// withIcon prefixes text with its glyph and a space. An empty glyph, as with
// show_icons: false, leaves the text alone, with no stray leading space.
func withIcon(icon, text string) string {
	if icon == "" {
		return text
	}
	return icon + " " + text
}
