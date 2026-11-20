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
	PanelID  PanelID
	Items    []string
	Formulae []brew.Formula
	Casks    []brew.Cask
	Taps     []brew.Tap
	// TrustUnknown is set when the Taps read got no tap-info: Taps holds names
	// only, so their trust state is not known (v2-10 AC-00).
	TrustUnknown bool
	Services     []brew.Service
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

type SearchDoneMsg struct {
	Results []string
	Raw     []brew.SearchResult
	Err     error
}

type SearchInfoLoadedMsg struct {
	Content string
	Err     error
}

// RefreshMsg asks the app to refresh its panels.
//
// LoadOutdated marks the one path allowed to shell out to `brew outdated`:
// the user pressing R. Every other emitter (the auto-refresh tick, the
// install/uninstall paths, task completion) leaves it false, so those
// refreshes read the outdated entries from the cache only and the dashboard
// never triggers a `brew outdated` call on its own.
type RefreshMsg struct {
	LoadOutdated bool
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

type DepCheckMsg struct {
	MutType mutationType
	Name    string
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
// back into a follow-up autoRefreshCmd() so the next tick re-evaluates.
type autoRefreshPausedMsg struct{}
