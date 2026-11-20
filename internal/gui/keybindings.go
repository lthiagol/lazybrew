package gui

type keyHint struct {
	key  string
	desc string
}

// serviceActionHints are the Services actions that apply to a service in its
// state: a running service stops and restarts, a stopped one starts and runs.
func serviceActionHints(running bool) []keyHint {
	if running {
		return []keyHint{{"S", "stop"}, {"r", "restart"}}
	}
	return []keyHint{{"s", "start"}, {"f", "run"}}
}

// servicesHints is the Services footer: the actions for the selected service
// lead, then the keys that apply in any state.
func servicesHints(p *panelData) []keyHint {
	hints := []keyHint{{"j/k", "navigate"}, {"c", "cleanup"}}
	if s := p.selectedService(); s != nil {
		hints = append(serviceActionHints(serviceRunning(*s)), hints...)
	}
	return hints
}

// panelHints lists the LIVE keys per panel — every entry here must have a
// handler in gui.go's Update, and every handled key should appear here
// (r1 finding 24).
func panelHints(id PanelID) []keyHint {
	switch id {
	case PanelFormulae:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}, {"i", "install"}, {"x", "remove"}, {"r", "reinstall"}, {"u", "upgrade"}, {"U", "upgrade all"}, {"p", "pin"}, {"L", "leaves"}, {"F", "fetch"}}
	case PanelCasks:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}, {"i", "install"}, {"x", "remove"}, {"X", "zap"}, {"r", "reinstall"}, {"u", "upgrade"}, {"U", "upgrade all"}, {"F", "fetch"}}
	case PanelServices:
		return []keyHint{{"j/k", "navigate"}, {"s", "start"}, {"S", "stop"}, {"r", "restart"}, {"f", "run"}, {"c", "cleanup"}}
	case PanelTaps:
		return []keyHint{{"j/k", "navigate"}, {"a", "add"}, {"x", "remove"}, {"t", "trust"}, {"r", "repair"}}
	case PanelStatus:
		return []keyHint{{"R", "refresh"}, {"c", "cleanup"}, {"d", "doctor"}, {"A", "autoremove"}, {"B", "brewfile"}, {"v", "vulns"}, {"m", "missing"}}
	case PanelOutdated:
		return []keyHint{{"j/k", "navigate"}, {"u", "upgrade"}, {"U", "upgrade all"}, {"Space", "select"}, {"a", "select all"}}
	case PanelSearch:
		return []keyHint{{"j/k", "navigate"}, {"Enter", "search"}, {"i", "install"}, {"I", "install"}, {"F", "fetch"}}
	default:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}}
	}
}
