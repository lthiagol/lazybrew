package gui

import "github.com/lthiagol/lazybrew/internal/brew"

// keyHint is one key a context answers to, and what each surface says about it.
// key is the footer's spelling of the key, helpKey the help screen's when that
// differs, and keys the spellings the key dispatch matches (key itself when
// empty, space-separated otherwise). desc is the footer's word, and an empty
// desc leaves the key off the footer; help is the help screen's line, and an
// empty help leaves it off the help screen.
type keyHint struct {
	key     string
	helpKey string
	keys    string
	desc    string
	help    string
}

// helpKeyText is the key as the help screen shows it.
func (h keyHint) helpKeyText() string {
	if h.helpKey != "" {
		return h.helpKey
	}
	return h.key
}

// The key tables. Every key a context answers to is declared once, here, with
// its footer word and its help line. The footer and the help screen list these
// entries in their own order, and the key dispatch in gui.go answers to them;
// keymap_test.go holds the three to one another.

// Keys every panel answers to.
var (
	keyArrows  = keyHint{key: "←↑↓→", keys: "up down left right", desc: "navigate"}
	keyNext    = keyHint{key: "Tab", keys: "tab", desc: "next", help: "Next panel"}
	keyPrev    = keyHint{key: "S-Tab", helpKey: "Shift+Tab", keys: "shift+tab", help: "Previous panel"}
	keyJump    = keyHint{key: "1-7", keys: "1 2 3 4 5 6 7", desc: "jump", help: "Jump to panel"}
	keyScroll  = keyHint{key: "j/k", keys: "j k down up", help: "Scroll list"}
	keyEnds    = keyHint{key: "Home/End", keys: "home end g G", help: "Top/bottom (g/G outside Search)"}
	keyTabs    = keyHint{key: "h/l", keys: "h l left right", help: "Switch tabs"}
	keySearch  = keyHint{key: "/", desc: "search", help: "Search packages"}
	keyFilter  = keyHint{key: "/", desc: "filter", help: "Filter the list"}
	keyHelp    = keyHint{key: "?", desc: "help", help: "This help"}
	keyRefresh = keyHint{key: "R", desc: "refresh", help: "Refresh all data"}
	keyLog     = keyHint{key: "C", desc: "log", help: "Toggle command log"}
	keyQuit    = keyHint{key: "q", keys: "q ctrl+c", desc: "quit", help: "Quit lazybrew"}
	keyClose   = keyHint{key: "esc", desc: "close"}
	keyHelpOff = keyHint{key: "?/esc", keys: "? esc", desc: "close"}
)

// Keys the panels share, with the same words wherever they appear.
var (
	keyUpgradeSelected = keyHint{key: "u", keys: "u", desc: "upgrade", help: "Upgrade selected"}
	keyUpgradeAll      = keyHint{key: "U", desc: "upgrade all", help: "Upgrade all"}
	keyCopy            = keyHint{key: "y", desc: "copy", help: "Copy the shown text"}
)

// Formulae.
var (
	fInstall   = keyHint{key: "i", desc: "install", help: "Install selected"}
	fRemove    = keyHint{key: "x", desc: "remove", help: "Uninstall selected"}
	fReinstall = keyHint{key: "r", desc: "reinstall", help: "Reinstall selected"}
	fPin       = keyHint{key: "p", desc: "pin", help: "Pin/Unpin selected"}
	fLeaves    = keyHint{key: "L", desc: "leaves", help: "Toggle leaves filter"}
	fFetch     = keyHint{key: "F", desc: "fetch", help: "Fetch selected"}
)

// Casks.
var (
	cInstall   = keyHint{key: "i", desc: "install", help: "Install"}
	cRemove    = keyHint{key: "x", desc: "remove", help: "Uninstall"}
	cZap       = keyHint{key: "X", desc: "zap", help: "Zap uninstall"}
	cReinstall = keyHint{key: "r", desc: "reinstall", help: "Reinstall"}
	cUpgrade   = keyHint{key: "u", desc: "upgrade", help: "Upgrade"}
	cFetch     = keyHint{key: "F", desc: "fetch", help: "Fetch"}
)

// Outdated.
var (
	oSelect    = keyHint{key: "Space", keys: "space", desc: "select", help: "Toggle selection"}
	oSelectAll = keyHint{key: "a", desc: "select all", help: "Select/deselect all"}
)

// Taps. The trust key's footer word depends on the tap, so it has an entry for
// each word; the help line sits with the first.
var (
	tAdd       = keyHint{key: "a", desc: "add tap", help: "Add tap"}
	tUntap     = keyHint{key: "x", desc: "untap", help: "Untap"}
	tTrust     = keyHint{key: "t", desc: "trust", help: "Trust or untrust tap"}
	tUntrust   = keyHint{key: "t", desc: "untrust"}
	tTrustItem = keyHint{key: "T", help: "Trust a formula or cask"}
	tRepair    = keyHint{key: "r", desc: "repair", help: "Repair tap"}
)

// Services. The action keys' footer words depend on the service's state.
var (
	svStart     = keyHint{key: "s", desc: "start", help: "Start"}
	svStop      = keyHint{key: "S", desc: "stop", help: "Stop"}
	svRestart   = keyHint{key: "r", desc: "restart", help: "Restart"}
	svRun       = keyHint{key: "f", desc: "run", help: "Run (foreground)"}
	svCleanup   = keyHint{key: "c", desc: "cleanup", help: "Cleanup stale services"}
	svLogStream = keyHint{key: "e", desc: "stream", help: "Logs tab: stdout or stderr"}
)

// Status.
var (
	stCleanup    = keyHint{key: "c", desc: "cleanup", help: "Run brew cleanup"}
	stDoctor     = keyHint{key: "d", desc: "doctor", help: "Open the Doctor tab and refresh it"}
	stAutoremove = keyHint{key: "A", desc: "autoremove", help: "Run brew autoremove"}
	stBrewfile   = keyHint{key: "B", desc: "brewfile", help: "Brewfile actions"}
	stVulns      = keyHint{key: "v", desc: "vulns", help: "Open the Vulns tab and run the check"}
	stMissing    = keyHint{key: "m", desc: "missing", help: "Open the Missing tab and refresh it"}
	stTabs       = keyHint{key: "h/l", keys: "h l left right", help: "Switch tabs"}
)

// Search.
var (
	sExecute   = keyHint{key: "Enter", keys: "enter", desc: "search", help: "Execute search"}
	sType      = keyHint{key: "/", help: "Type search query"}
	sInstall   = keyHint{key: "i", desc: "install", help: "Install selected result"}
	sFetch     = keyHint{key: "F", desc: "fetch", help: "Fetch selected result"}
	sUninstall = keyHint{key: "x", desc: "uninstall", help: "Uninstall installed result"}
	sZap       = keyHint{key: "X", help: "Zap installed cask result"}
)

// serviceActionHints are the Services actions that apply to a service in its
// state. A service that is not running starts and runs. One that stops and
// restarts (serviceStoppable) also offers stop and restart.
func serviceActionHints(s brew.Service) []keyHint {
	var hints []keyHint
	if !serviceRunning(s) {
		hints = append(hints, svStart)
	}
	if serviceStoppable(s) {
		hints = append(hints, svStop, svRestart)
	}
	if !serviceRunning(s) {
		hints = append(hints, svRun)
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
		hints = append(hints, tUntrust)
	default:
		hints = append(hints, tTrust)
	}
	if !t.IsOfficial {
		hints = append(hints, tUntap)
	}
	return append(hints, tRepair)
}

// tapsHints is the Taps footer: the actions for the selected tap lead, then the
// keys that apply to any tap.
func tapsHints(p *panelData) []keyHint {
	hints := []keyHint{tAdd, keyCopy}
	if t := p.selectedTap(); t != nil {
		hints = append(tapActionHints(*t, !p.trustUnknown), hints...)
	}
	return hints
}

// servicesHints is the Services footer: the actions for the selected service
// lead, then the keys that apply in any state.
func servicesHints(p *panelData) []keyHint {
	hints := []keyHint{svCleanup, keyCopy}
	if s := p.selectedService(); s != nil {
		hints = append(serviceActionHints(*s), hints...)
	}
	return hints
}

// panelHints lists the footer keys of a panel with no state applied: the Taps
// and Services footers are made by tapsHints and servicesHints, which apply it.
// The arrow hint and the list keys are on every footer, so no panel lists them.
func panelHints(id PanelID) []keyHint {
	switch id {
	case PanelFormulae:
		return []keyHint{fInstall, fRemove, fReinstall, keyUpgradeSelected, keyUpgradeAll, fPin, fLeaves, fFetch, keyCopy}
	case PanelCasks:
		return []keyHint{cInstall, cRemove, cZap, cReinstall, cUpgrade, keyUpgradeAll, cFetch, keyCopy}
	case PanelServices:
		return []keyHint{svStart, svStop, svRestart, svRun, svCleanup, keyCopy}
	case PanelTaps:
		return []keyHint{tAdd, tUntap, tTrust, tRepair, keyCopy}
	case PanelStatus:
		return []keyHint{keyRefresh, stCleanup, stDoctor, stAutoremove, stBrewfile, stVulns, stMissing, keyCopy}
	case PanelOutdated:
		return []keyHint{keyUpgradeSelected, keyUpgradeAll, oSelect, oSelectAll, keyCopy}
	case PanelSearch:
		return []keyHint{sExecute, sInstall, sUninstall, sFetch}
	default:
		return nil
	}
}
