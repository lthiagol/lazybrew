package modal

import (
	"image/color"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

func TestConfirmModalYes(t *testing.T) {
	m := NewConfirmModal(ActionNeutral, "Test", "Are you sure?")
	m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if !m.Done() {
		t.Error("expected Done after y")
	}
	res := m.Result().(*ConfirmResult)
	if !res.Confirmed {
		t.Error("expected Confirmed=true")
	}
}

func TestConfirmModalNo(t *testing.T) {
	m := NewConfirmModal(ActionNeutral, "Test", "Are you sure?")
	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if !m.Done() {
		t.Error("expected Done after n")
	}
	res := m.Result().(*ConfirmResult)
	if res.Confirmed {
		t.Error("expected Confirmed=false")
	}
}

func TestConfirmModalEsc(t *testing.T) {
	m := NewConfirmModal(ActionNeutral, "Test", "Are you sure?")
	m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !m.Cancelled() {
		t.Error("expected Cancelled after Esc")
	}
}

func TestConfirmModalDefaultNo(t *testing.T) {
	m := NewConfirmModal(ActionNeutral, "Test", "Are you sure?")
	if m.selected != 1 {
		t.Error("expected default selection No (index 1)")
	}
}

func TestInputModalSubmit(t *testing.T) {
	m := NewInputModal("Enter name:")

	for _, r := range "test-package" {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

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
	m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

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

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.selected != 1 {
		t.Errorf("expected selected 1 after j, got %d", m.selected)
	}

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.selected != 2 {
		t.Errorf("expected selected 2 after second j, got %d", m.selected)
	}

	m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if m.selected != 1 {
		t.Errorf("expected selected 1 after k, got %d", m.selected)
	}
}

func TestMenuModalEnterSelects(t *testing.T) {
	items := []string{"Option A", "Option B"}
	m := NewMenuModal("Test", items)

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

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
	m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !m.Cancelled() {
		t.Error("expected Cancelled after Esc")
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

// TestToastRearmsTick locks the auto-dismiss chain (r1 finding 4): a toast
// that is not yet expired must return another tick cmd so the chain never
// stops after the first tick.
func TestToastRearmsTick(t *testing.T) {
	toast := NewToast("Test message", ToastInfo)
	_, cmd := toast.Update(ToastTickMsg{})
	if cmd == nil {
		t.Fatal("expected a re-armed tick cmd from an unexpired toast")
	}
	if toast.Dismissed() {
		t.Fatal("unexpired toast must not be dismissed")
	}
}

// TestActionKindAccent locks AC-01.1: each action kind is framed in its theme
// role, and the dialog reports the kind it was opened with.
func TestActionKindAccent(t *testing.T) {
	cases := []struct {
		kind ActionKind
		want color.Color
	}{
		{ActionDanger, style.ErrorColor},
		{ActionUpgrade, style.WarningColor},
		{ActionInstall, style.InfoColor},
		{ActionNeutral, style.AccentColor},
	}
	for _, tc := range cases {
		if got := tc.kind.accent(); got != tc.want {
			t.Errorf("kind %v: accent = %v, want %v", tc.kind, got, tc.want)
		}
		if got := NewConfirmModal(tc.kind, "T", "m").Kind(); got != tc.kind {
			t.Errorf("Kind() = %v, want %v", got, tc.kind)
		}
	}
}
