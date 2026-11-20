package modal

import (
	tea "charm.land/bubbletea/v2"
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
