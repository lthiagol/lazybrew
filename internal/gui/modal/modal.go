package modal

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Modal interface {
	tea.Model
	Done() bool
	Cancelled() bool
	Result() interface{}
}

type ConfirmResult struct {
	Confirmed bool
}

type InputResult struct {
	Value     string
	Cancelled bool
}

type MenuResult struct {
	SelectedIndex int
	Cancelled     bool
}
