package gui

import (
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/gui/task"
)

// Task message re-exports — types defined in internal/gui/task/
type TaskStartedMsg = task.TaskStartedMsg
type TaskOutputMsg = task.TaskOutputMsg
type TaskCompletedMsg = task.TaskCompletedMsg
type TaskRejectedMsg = task.TaskRejectedMsg

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

type RefreshMsg struct{}

type CleanupPreviewMsg struct {
	Lines []string
}

type AutoremovePreviewMsg struct {
	Lines []string
}

type UpdateTickMsg struct{}

type StartUpdateMsg struct{}

type UpdateCompleteMsg struct {
	Output []string
	Err    error
}

type DepCheckMsg struct {
	MutType mutationType
	Name    string
	Label   string
	Message string
}
