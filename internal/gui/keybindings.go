package gui

type keyHint struct {
	key  string
	desc string
}

func panelHints(id PanelID) []keyHint {
	switch id {
	case PanelFormulae:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}, {"x", "remove"}, {"r", "reinstall"}, {"u", "upgrade"}, {"p", "pin"}, {"L", "leaves"}}
	case PanelCasks:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}, {"x", "remove"}, {"X", "zap"}, {"p", "pin"}}
	case PanelServices:
		return []keyHint{{"j/k", "navigate"}, {"s", "start"}, {"S", "stop"}, {"r", "restart"}, {"f", "run"}, {"c", "cleanup"}}
	case PanelTaps:
		return []keyHint{{"j/k", "navigate"}, {"a", "add"}, {"x", "remove"}, {"t", "trust"}, {"r", "repair"}}
	case PanelStatus:
		return []keyHint{{"R", "refresh"}, {"c", "cleanup"}, {"d", "doctor"}, {"A", "autoremove"}, {"B", "brewfile"}, {"v", "vulns"}, {"m", "missing"}}
	case PanelOutdated:
		return []keyHint{{"j/k", "navigate"}, {"u", "upgrade"}, {"Space", "select"}}
	default:
		return []keyHint{{"j/k", "navigate"}, {"[ ]", "tabs"}}
	}
}
