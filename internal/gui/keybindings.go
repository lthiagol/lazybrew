package gui

type keyHint struct {
	key  string
	desc string
}

func panelHints(id PanelID) []keyHint {
	switch id {
	case PanelFormulae, PanelCasks:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}}
	case PanelServices:
		return []keyHint{{"j/k", "navigate"}, {"s", "start"}, {"S", "stop"}, {"r", "restart"}, {"f", "run"}, {"c", "cleanup"}}
	case PanelTaps:
		return []keyHint{{"j/k", "navigate"}, {"a", "add"}, {"x", "remove"}}
	case PanelStatus:
		return []keyHint{{"R", "refresh"}, {"B", "brewfile"}, {"v", "vulns"}, {"m", "missing"}}
	case PanelOutdated:
		return []keyHint{{"j/k", "navigate"}, {"u", "upgrade"}, {"Space", "select"}}
	default:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}}
	}
}
