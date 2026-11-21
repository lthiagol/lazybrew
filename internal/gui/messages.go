package gui

import (
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// Task message re-exports — types defined in internal/gui/task/
type TaskStartedMsg = task.TaskStartedMsg
type TaskOutputMsg = task.TaskOutputMsg
type TaskCompletedMsg = task.TaskCompletedMsg

type DataLoadedMsg struct {
	// launch marks a read the launch started (Init). Once a stale update's
	// refresh has been issued, such a read is dropped: it holds the metadata the
	// update replaces, and the refresh reads the panels again.
	launch bool

	PanelID  PanelID
	Items    []string
	Formulae []brew.Formula
	Casks    []brew.Cask
	Taps     []brew.Tap
	// TrustUnknown is set when the Taps read got no tap-info: Taps holds names
	// only, so their trust state is not known (v2-10 AC-00).
	TrustUnknown bool
	// Trusted is the Taps list's trusted formulae, casks and commands (trustedRows
	// order). TrustErr is the trust read's error: the taps still load without it.
	Trusted  []brew.TrustEntry
	TrustErr error
	Services []brew.Service
	// ServiceInfo is the Services Info view's data: `brew services info`. It is
	// empty when that read failed, and the Info view then shows the list's fields.
	ServiceInfo []brew.ServiceInfo
	// Summary is set by the status dashboard load only: the typed facts the
	// sidebar rows, the header counts and the dashboard cards are built from.
	Summary statusSummary
	Err     error
}

type TabContentMsg struct {
	PanelID  PanelID
	TabIndex int
	ItemName string
	Content  string
	Err      error
}

// CleanupPreviewMsg is the Cleanup tab's read: `brew cleanup -n`, parsed.
type CleanupPreviewMsg struct {
	Preview brew.CleanupPreview
	Err     error
}

// AutoremovePreviewMsg is the Unused tab's read: `brew autoremove -n`, parsed.
type AutoremovePreviewMsg struct {
	Preview brew.AutoremovePreview
	Err     error
}

// VulnsMsg is the result of the Vulns tab's check: brew's report, or the error
// that stopped it.
type VulnsMsg struct {
	Report brew.VulnReport
	Err    error
}

type SearchDoneMsg struct {
	Results []string
	Raw     []brew.SearchResult
	Err     error
}

// SearchMarksMsg carries the installed marks re-read for the rows of the last
// search, in the rows' order. gen is the remarkGen the read was issued under.
type SearchMarksMsg struct {
	Results []brew.SearchResult
	gen     int
}

type SearchInfoLoadedMsg struct {
	Content string
	Err     error
}

// RefreshMsg asks the app to refresh its panels.
//
// LoadOutdated marks the paths allowed to shell out to `brew outdated`: the
// user pressing R, and the end of a brew update, which leaves the count stale
// otherwise. Every other emitter (the auto-refresh tick, the install/uninstall
// paths, other task completions) leaves it false, so those refreshes read the
// outdated entries from the cache only and the dashboard never triggers a
// `brew outdated` call on its own.
//
// ReloadDiagnostics works the same way for `brew doctor` and `brew config`,
// the slowest reads: only the R key sets it. Other refreshes reuse the cached
// results.
type RefreshMsg struct {
	LoadOutdated      bool
	ReloadDiagnostics bool
}

// servicesReloadMsg reloads the Services panel alone, after a services task
// fails or is cancelled. Unlike RefreshMsg it shows no "Data refreshed" toast,
// so the error toast stays the last word (F-r1-02).
type servicesReloadMsg struct{}

// tapsReloadMsg reloads the Taps panel alone, after a taps write fails or is
// cancelled. Like servicesReloadMsg it shows no "Data refreshed" toast.
type tapsReloadMsg struct{}

// opAutoCloseMsg is the timer of a successful operation. It names the operation
// it is for, so a timer that fires after a newer operation started closes
// nothing.
type opAutoCloseMsg struct {
	op *Operation
}

type UpdateTickMsg struct{}

type StartUpdateMsg struct{}

// lastUpdateMsg is the launch read of when brew last updated. err is set when
// the read failed; auto mode then treats the update as stale.
type lastUpdateMsg struct {
	at  time.Time
	err error
}

// SudoRunDoneMsg is a failed step run again in the terminal, once the command
// has exited. Err is nil on success.
type SudoRunDoneMsg struct {
	ID    string
	Title string
	Args  []string
	Err   error
}

type DepCheckMsg struct {
	MutType mutationType
	Label   string
	Message string
	Kind    modal.ActionKind
	Detail  string
}

// autoRefreshTickMsg is delivered by tea.Tick so Update can evaluate
// pause-on-interaction against the *current* model (not a stale copy
// captured when the tick was scheduled).
type autoRefreshTickMsg struct {
	at time.Time
}

// autoRefreshPausedMsg signals that the auto-refresh tick observed a
// recent key press and rescheduled itself. The handler converts this
// back into a follow-up tick (armAutoRefresh) so the next tick re-evaluates.
type autoRefreshPausedMsg struct{}
