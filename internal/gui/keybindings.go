package gui

import "github.com/lthiagol/lazybrew/internal/brew"

type keyHint struct {
	key  string
	desc string
}

// serviceActionHints are the Services actions that apply to a service in its
// state. A service that is not running starts and runs. One that stops and
// restarts (serviceStoppable) also offers stop and restart.
func serviceActionHints(s brew.Service) []keyHint {
	var hints []keyHint
	if !serviceRunning(s) {
		hints = append(hints, keyHint{"s", "start"})
	}
	if serviceStoppable(s) {
		hints = append(hints, keyHint{"S", "stop"}, keyHint{"r", "restart"})
	}
	if !serviceRunning(s) {
		hints = append(hints, keyHint{"f", "run"})
	}
	return hints
}

// tapActionHints are the Taps actions that apply to a tap in its state (v2-10
// AC-01). An untrusted tap is trusted with t, a trusted one untrusted with t,
// and an official one needs no trust, so it has no t. x untaps, but lazybrew
// refuses to untap an official tap (Q-v2-10-a), so an official tap gets no x.
// r repairs in every state. With the trust state unread, t is left out: it would
// act on a state nobody has read.
func tapActionHints(t brew.Tap, trustKnown bool) []keyHint {
	var hints []keyHint
	switch {
	case t.IsOfficial || !trustKnown:
	case t.Trusted:
		hints = append(hints, keyHint{"t", "untrust"})
	default:
		hints = append(hints, keyHint{"t", "trust"})
	}
	if !t.IsOfficial {
		hints = append(hints, keyHint{"x", "untap"})
	}
	return append(hints, keyHint{"r", "repair"})
}

// tapsHints is the Taps footer: the actions for the selected tap lead, then the
// keys that apply to any tap.
func tapsHints(p *panelData) []keyHint {
	hints := []keyHint{{"j/k", "navigate"}, {"a", "add tap"}}
	if t := p.selectedTap(); t != nil {
		hints = append(tapActionHints(*t, !p.trustUnknown), hints...)
	}
	return hints
}

// servicesHints is the Services footer: the actions for the selected service
// lead, then the keys that apply in any state.
func servicesHints(p *panelData) []keyHint {
	hints := []keyHint{{"j/k", "navigate"}, {"c", "cleanup"}}
	if s := p.selectedService(); s != nil {
		hints = append(serviceActionHints(*s), hints...)
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
		return []keyHint{{"j/k", "navigate"}, {"a", "add tap"}, {"x", "untap"}, {"t", "trust"}, {"T", "trust item"}, {"r", "repair"}}
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
