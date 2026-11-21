package gui

import (
	"context"
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

func (m *Model) executeTrustMenuAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	tapName := panel.selectedName()
	tap := panel.selectedTap()

	// Dispatch on the indices recorded when the menu was built, not on
	// hardcoded positions: the submenu entries only exist when the tap
	// has formulae/casks, so their positions vary (r4 finding R4-3).
	formulaIdx := m.pendingTrustFormulaIdx
	caskIdx := m.pendingTrustCaskIdx
	// The tracked indices belong to this dispatch only.
	m.pendingTrustFormulaIdx = -1
	m.pendingTrustCaskIdx = -1

	switch {
	case index == 0:
		return m.executeTrustAction(0)
	case index == 1:
		// Untrust is always confirmed, here as on t (v2-10 AC-02.1).
		return m.openUntrustConfirm()
	case index == formulaIdx && formulaIdx >= 0:
		if tap != nil && len(tap.FormulaNames) > 0 {
			menuModal := modal.NewMenuModal("Trust formula in "+tapName, tap.FormulaNames)
			m.activeModal = menuModal
			m.pendingAction = "trust-formula"
			return menuModal.Init()
		}
	case index == caskIdx && caskIdx >= 0:
		if tap != nil && len(tap.CaskNames) > 0 {
			menuModal := modal.NewMenuModal("Trust cask in "+tapName, tap.CaskNames)
			m.activeModal = menuModal
			m.pendingAction = "trust-cask"
			return menuModal.Init()
		}
	}
	return nil
}

func (m *Model) executeTrustAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	name := panel.selectedName()
	if name == "" {
		return nil
	}

	verb := "Trust"
	trustID := "trust:" + name
	if index != 0 {
		// The overlay title and completion toast must name the action
		// actually taken (r3 finding R3-8): index != 0 is the untrust
		// arm, not a second trust path.
		verb = "Untrust"
		trustID = "untrust:" + name
	}
	client := m.client
	t := errTask(trustID, verb+": "+name, func(ctx context.Context) error {
		if index == 0 {
			return client.TrustWrite.TrustTap(ctx, name)
		}
		return client.TrustWrite.UntrustTap(ctx, name)
	})

	return m.enqueue(t)
}

func (m *Model) executeTrustFormulaAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	tap := panel.selectedTap()
	if tap == nil || index >= len(tap.FormulaNames) {
		return nil
	}
	return m.trustItemCmd(tap.FormulaNames[index], "formula")
}

func (m *Model) executeTrustCaskAction(index int) tea.Cmd {
	panel := m.panels[PanelTaps]
	tap := panel.selectedTap()
	if tap == nil || index >= len(tap.CaskNames) {
		return nil
	}
	return m.trustItemCmd(tap.CaskNames[index], "cask")
}

func (m *Model) trustItemCmd(fullName, itemType string) tea.Cmd {
	client := m.client
	t := errTask("trust:"+fullName, "Trust: "+fullName, func(ctx context.Context) error {
		if itemType == "formula" {
			return client.TrustWrite.TrustFormula(ctx, fullName)
		}
		return client.TrustWrite.TrustCask(ctx, fullName)
	})

	return m.enqueue(t)
}

func (m Model) executeUntap() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	name := panel.selectedName()

	client := m.client
	t := errTask("untap:"+name, "Untap "+name, func(ctx context.Context) error {
		return client.TapsWrite.Untap(ctx, name)
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) executeRepair() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	name := panel.selectedName()

	client := m.client
	t := errTask("repair:"+name, "Repair "+name, func(ctx context.Context) error {
		return client.TapsWrite.Repair(ctx, name)
	})

	c := m.enqueue(t)
	return m, c
}

func (m *Model) startTapAdd() (tea.Model, tea.Cmd) {
	m.pendingAction = "tap-add"
	inputModal := modal.NewInputModal("Add tap", "Tap repository (user/repo):")
	m.activeModal = inputModal
	return m, inputModal.Init()
}

// executeTapAdd runs `brew tap <name>` through the task manager. It used
// to be unreachable: the tap-add input result was routed to the search
// flow instead (r1 finding 2).
func (m Model) executeTapAdd(name string) (tea.Model, tea.Cmd) {
	client := m.client
	t := errTask("tap-"+name, "Tap "+name, func(ctx context.Context) error {
		return client.TapsWrite.Tap(ctx, name)
	})

	c := m.enqueue(t)
	return m, c
}

// trustStateMissingToast is what t and T say while the taps were read without
// their trust state: nothing is enqueued, since the state they would act on is
// not known (v2-10 AC-00.2).
const trustStateMissingToast = "trust state not loaded — press R"

// officialTrustToast says an official tap needs no trust. Trust is refused on
// it, and t and T say so rather than act (v2-10 AC-02.1).
func officialTrustToast(name string) string {
	return name + " is official — no trust needed"
}

// toggleTrust is t on the Taps list (v2-10 AC-02.1): an untrusted tap is trusted
// at once, a trusted one is untrusted after a confirm, and an official one needs
// no trust, so it enqueues nothing.
func (m *Model) toggleTrust() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	if panel.trustUnknown {
		toast := m.setToast(trustStateMissingToast, modal.ToastInfo)
		return m, toast
	}
	tap := panel.selectedTap()
	if tap == nil {
		return m, nil
	}
	switch {
	case tap.IsOfficial:
		toast := m.setToast(officialTrustToast(tap.Name), modal.ToastInfo)
		return m, toast
	case tap.Trusted:
		c := m.openUntrustConfirm()
		return m, c
	default:
		c := m.executeTrustAction(0)
		return m, c
	}
}

// openUntrustConfirm asks before a tap is untrusted. The confirm acts on the
// tap the cursor is on, as the untap confirm does.
func (m *Model) openUntrustConfirm() tea.Cmd {
	name := m.panels[PanelTaps].selectedName()
	m.pendingAction = "untrust"
	m.activeModal = modal.NewConfirmModal(modal.ActionUpgrade, "Untrust", "Untrust "+name+"?").
		WithDetail("Homebrew will refuse to load its formulae and casks until you trust it again.")
	return m.activeModal.Init()
}

func (m *Model) startTrustMenu() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	if panel.trustUnknown {
		toast := m.setToast(trustStateMissingToast, modal.ToastInfo)
		return m, toast
	}
	if tap := panel.selectedTap(); tap != nil && tap.IsOfficial {
		toast := m.setToast(officialTrustToast(tap.Name), modal.ToastInfo)
		return m, toast
	}
	tapName := panel.selectedName()
	tap := panel.selectedTap()
	formulaCount := 0
	caskCount := 0
	if tap != nil {
		formulaCount = len(tap.FormulaNames)
		caskCount = len(tap.CaskNames)
	}

	menuItems := []string{
		"Trust entire tap: " + tapName,
		"Untrust tap: " + tapName,
	}
	// Track where the submenu entries actually landed: on a tap with
	// only one of the two kinds the hardcoded positions 2/3 pointed at
	// the wrong (or a missing) entry, making "Trust specific cask..."
	// dead on cask-only taps (r4 finding R4-3).
	m.pendingTrustFormulaIdx = -1
	m.pendingTrustCaskIdx = -1
	if formulaCount > 0 {
		m.pendingTrustFormulaIdx = len(menuItems)
		menuItems = append(menuItems, "Trust specific formula...")
	}
	if caskCount > 0 {
		m.pendingTrustCaskIdx = len(menuItems)
		menuItems = append(menuItems, "Trust specific cask...")
	}

	m.pendingAction = "trust-menu"
	menuModal := modal.NewMenuModal("Trust: "+tapName, menuItems)
	m.activeModal = menuModal
	return m, menuModal.Init()
}

// tapsReload asks for the Taps list and its trust state again when a taps write
// fails or is cancelled (v2-10 AC-04). It is nil for every other task; a
// successful one reloads through RefreshMsg, as every task does.
func tapsReload(id string) tea.Cmd {
	if !isTapTask(id) {
		return nil
	}
	return func() tea.Msg { return tapsReloadMsg{} }
}

// isTapTask is whether a task is a taps write or a trust write on a tap, which
// the Taps panel reloads when it ends.
func isTapTask(id string) bool {
	for _, prefix := range []string{"tap-", "untap:", "repair:", "trust:", "untrust:"} {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

// tapFailureText is the error toast of a failed taps write: its title, then the
// line of brew's stderr that says what went wrong. brew's error is only its
// exit code, so the stderr line is the message (v2-10 AC-03.2).
func tapFailureText(title string, err error) string {
	var exit *brew.BrewExitError
	if errors.As(err, &exit) {
		if line := brewMessage(exit.Stderr); line != "" {
			return title + ": " + line
		}
	}
	return title + ": " + err.Error()
}

func (m Model) confirmUntap() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	name := panel.selectedName()
	if name == "" {
		return m, nil
	}
	if isOfficialTap(name) {
		c := m.setToast("Cannot untap official taps", modal.ToastWarning)
		return m, c
	}
	question := "Untap " + name + "?"
	// Packages installed from the tap block the untap, and brew says so only
	// once it is run. The list comes from the Formulae and Casks lists already
	// loaded, so the confirm makes no brew call (v2-10 AC-03.1).
	if installed := installedFromTap(m, name); len(installed) > 0 {
		question += "\nInstalled from this tap; brew will refuse to untap while they are installed:\n  " +
			strings.Join(installed, "\n  ")
	}
	m.pendingAction = "untap"
	m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Untap", question).
		WithDetail("The tap's formulae and casks are no longer available.")
	return m, m.activeModal.Init()
}

// installedFromTap names the installed formulae and casks whose full name is
// under the tap: user/repo/<name>. A filtered list (leaves, or the query) is read
// from the full list it was filtered from, so a package the filter hides still counts.
func installedFromTap(m Model, tap string) []string {
	prefix := tap + "/"
	var names []string
	for _, f := range m.panels[PanelFormulae].fullFormulae() {
		if strings.HasPrefix(f.FullName, prefix) {
			names = append(names, f.Name)
		}
	}
	for _, c := range m.panels[PanelCasks].fullCasks() {
		if strings.HasPrefix(c.FullName, prefix) {
			names = append(names, c.Name)
		}
	}
	return names
}

func isOfficialTap(name string) bool {
	return strings.HasPrefix(name, "homebrew/")
}

func (m Model) confirmRepair() (tea.Model, tea.Cmd) {
	panel := m.panels[PanelTaps]
	name := panel.selectedName()
	if name == "" {
		return m, nil
	}
	m.pendingAction = "repair"
	m.activeModal = modal.NewConfirmModal(modal.ActionInstall, "Repair", "Repair tap "+name+"?").
		WithDetail("Runs brew tap --repair, which migrates the tap to the current layout.")
	return m, m.activeModal.Init()
}
