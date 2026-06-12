package gui

type DataLoadedMsg struct {
	PanelID PanelID
	Items   []string
	RawData interface{}
	Err     error
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
