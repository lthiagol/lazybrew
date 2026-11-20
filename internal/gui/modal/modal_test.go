package modal

import (
	"image/color"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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

// TestToastLifetimeByType locks AC-04.2: a toast stays up 3 s, and an error
// 6 s, measured from when it began.
func TestToastLifetimeByType(t *testing.T) {
	base := time.Now()
	cases := []struct {
		typ     ToastType
		upTo    time.Duration
		expired bool
	}{
		{ToastSuccess, 2900 * time.Millisecond, false},
		{ToastSuccess, 3100 * time.Millisecond, true},
		{ToastInfo, 3100 * time.Millisecond, true},
		{ToastWarning, 3100 * time.Millisecond, true},
		{ToastError, 5900 * time.Millisecond, false},
		{ToastError, 6100 * time.Millisecond, true},
	}
	for _, tc := range cases {
		toast := NewToastAt("msg", tc.typ, base)
		if got := toast.Expired(base.Add(tc.upTo)); got != tc.expired {
			t.Errorf("type %v at %v: expired = %v, want %v", tc.typ, tc.upTo, got, tc.expired)
		}
	}
}

// TestToastViewCutsWithEllipsis locks AC-04.1: the message is cut with an
// ellipsis to the width it is given, and the icon sits before it.
func TestToastViewCutsWithEllipsis(t *testing.T) {
	toast := NewToast("Install a-very-long-formula-name completed successfully", ToastSuccess)
	view := toast.View("OK", 20)
	plain := ansi.Strip(view)
	if !strings.Contains(plain, "OK ") || !strings.Contains(plain, "…") {
		t.Fatalf("toast lacks its icon or its ellipsis:\n%s", plain)
	}
	if strings.Contains(plain, "successfully") {
		t.Fatalf("message was not cut:\n%s", plain)
	}
	if !strings.Contains(view, sgrOf(style.SuccessColor)) {
		t.Fatal("the toast is not drawn in its role colour")
	}
}

// sgrOf is the SGR sequence that opens a foreground colour.
func sgrOf(c color.Color) string {
	sample := lipgloss.NewStyle().Foreground(c).Render("x")
	return sample[:strings.Index(sample, "x")]
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
