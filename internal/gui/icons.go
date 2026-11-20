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

// panelIcon is the glyph a panel's sidebar title carries.
func (m Model) panelIcon(id PanelID) string {
	switch id {
	case PanelStatus:
		return m.icons.Status
	case PanelFormulae:
		return m.icons.Formula
	case PanelCasks:
		return m.icons.Cask
	case PanelOutdated:
		return m.icons.Outdated
	case PanelTaps:
		return m.icons.Tap
	case PanelServices:
		return m.icons.Service
	case PanelSearch:
		return m.icons.Search
	default:
		return ""
	}
}
