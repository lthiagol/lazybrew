package gui

import "github.com/thiago/lazybrew/internal/brew"

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
	Content  string
	Err      error
}

type SearchDoneMsg struct {
	Results []string
	Err     error
}

type RefreshMsg struct{}

type CleanupPreviewMsg struct {
	Lines []string
}

type AutoremovePreviewMsg struct {
	Lines []string
}

type DepCheckMsg struct {
	MutType mutationType
	Name    string
	Label   string
	Message string
}
