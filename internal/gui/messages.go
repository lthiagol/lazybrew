package gui

import (
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
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
	Services []brew.Service
	Err      error
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

type UpdateTickMsg struct{}

type StartUpdateMsg struct{}

type DepCheckMsg struct {
	MutType mutationType
	Name    string
	Label   string
	Message string
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
