package modal

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmModalYes(t *testing.T) {
	m := NewConfirmModal("Test", "Are you sure?")
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if !m.Done() {
		t.Error("expected Done after y")
	}
	res := m.Result().(*ConfirmResult)
	if !res.Confirmed {
		t.Error("expected Confirmed=true")
	}
}

func TestConfirmModalNo(t *testing.T) {
	m := NewConfirmModal("Test", "Are you sure?")
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if !m.Done() {
		t.Error("expected Done after n")
	}
	res := m.Result().(*ConfirmResult)
	if res.Confirmed {
		t.Error("expected Confirmed=false")
	}
}

func TestConfirmModalEsc(t *testing.T) {
	m := NewConfirmModal("Test", "Are you sure?")
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.Cancelled() {
		t.Error("expected Cancelled after Esc")
	}
}

func TestConfirmModalDefaultNo(t *testing.T) {
	m := NewConfirmModal("Test", "Are you sure?")
	if m.selected != 1 {
		t.Error("expected default selection No (index 1)")
	}
}

func TestInputModalSubmit(t *testing.T) {
	m := NewInputModal("Enter name:")

	for _, r := range "test-package" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.Done() {
		t.Error("expected Done after Enter")
	}
	res := m.Result().(*InputResult)
	if res.Value != "test-package" {
		t.Errorf("expected test-package, got %s", res.Value)
	}
}

func TestInputModalEsc(t *testing.T) {
	m := NewInputModal("Enter name:")
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !m.Cancelled() {
		t.Error("expected Cancelled after Esc")
	}
}

func TestMenuModalNavigation(t *testing.T) {
	items := []string{"Option A", "Option B", "Option C"}
	m := NewMenuModal("Test", items)

	if m.selected != 0 {
		t.Errorf("expected selected 0, got %d", m.selected)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.selected != 1 {
		t.Errorf("expected selected 1 after j, got %d", m.selected)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.selected != 2 {
		t.Errorf("expected selected 2 after second j, got %d", m.selected)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.selected != 1 {
		t.Errorf("expected selected 1 after k, got %d", m.selected)
	}
}

func TestMenuModalEnterSelects(t *testing.T) {
	items := []string{"Option A", "Option B"}
	m := NewMenuModal("Test", items)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.Done() {
		t.Error("expected Done after Enter")
	}
	res := m.Result().(*MenuResult)
	if res.SelectedIndex != 1 {
		t.Errorf("expected selected index 1, got %d", res.SelectedIndex)
	}
}

func TestMenuModalEsc(t *testing.T) {
	m := NewMenuModal("Test", []string{"A", "B"})
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.Cancelled() {
		t.Error("expected Cancelled after Esc")
	}
}

func TestProgressModalAppendLine(t *testing.T) {
	m := NewProgressModal("Working", nil)

	defer func() {
		if r := recover(); r != nil {
			t.Fatal("AppendLine panicked:", r)
		}
	}()

	m.AppendLine("line 1")
	m.AppendLine("line 2")
}

func TestProgressModalSetDone(t *testing.T) {
	m := NewProgressModal("Working", nil)
	m.SetDone(nil)

	if !m.Done() && !m.Cancelled() {
		t.Error("expected Done or Cancelled after SetDone(nil)")
	}
}

func TestProgressModalCancel(t *testing.T) {
	cancelled := false
	cancel := func() { cancelled = true }
	m := NewProgressModal("Working", cancel)

	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !cancelled {
		t.Error("expected cancel func to be called on Esc")
	}
}

func TestToastDismiss(t *testing.T) {
	toast := NewToast("Test message", ToastSuccess)
	if toast.Dismissed() {
		t.Error("new toast should not be dismissed")
	}
	toast.created = toast.created.Add(-4 * time.Second)
	toast.Update(ToastTickMsg{})
	if !toast.Dismissed() {
		t.Error("toast should be dismissed after 3 seconds")
	}
}
