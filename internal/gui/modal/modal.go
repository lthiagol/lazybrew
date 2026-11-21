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

// Hint is one key a modal answers to, and what it does. Keys are the
// msg.String() spellings the modal answers with that key, space separated;
// keymap_test.go holds them to the modal's Update.
type Hint struct {
	Key  string
	Desc string
	Keys string
}

// Hinter is a modal that lists the keys it answers to. Its key line and the
// footer are both drawn from Hints, so the two cannot disagree.
type Hinter interface {
	Hints() []Hint
}

// hintCancel is the key every modal answers with a cancel.
var hintCancel = Hint{Key: "esc", Desc: "cancel", Keys: "esc"}
