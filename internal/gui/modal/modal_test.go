package modal

import (
	"fmt"
	"image/color"
	"regexp"
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
	m := NewInputModal("Name", "Enter name:")

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
	m := NewInputModal("Name", "Enter name:")
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

// TestMenuModalDigitWithoutTextSelects (LB-51 FC5): a digit key that carries no
// text selects its item; it does not panic on the empty text.
func TestMenuModalDigitWithoutTextSelects(t *testing.T) {
	m := NewMenuModal("Test", []string{"A", "B", "C"})
	m.Update(tea.KeyPressMsg{Code: '1'})
	if !m.Done() || m.selected != 1 {
		t.Fatalf("a digit with no text: done %v, selected %d, want done and 1", m.Done(), m.selected)
	}
}

// scrollUpMark and scrollDownMark are the menu's marks for hidden items; the key
// line's arrows do not match them.
var (
	scrollUpMark   = regexp.MustCompile(`↑ \d+ more`)
	scrollDownMark = regexp.MustCompile(`↓ \d+ more`)
)

// menuItems is n items named formula-0 up to formula-(n-1).
func menuItems(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("formula-%d", i)
	}
	return items
}

// TestMenuModalWindowFollowsTheCursor (LB-51 FC3): a menu longer than its box
// shows a window that holds the cursor, with a mark for each side it hides items
// on. The box keeps its height as the cursor moves.
func TestMenuModalWindowFollowsTheCursor(t *testing.T) {
	m := NewMenuModal("Trust formula in tap", menuItems(15))
	top := plainLines(m.View().Content)
	if joined := strings.Join(top, "\n"); !strings.Contains(joined, "↓ 5 more") || scrollUpMark.MatchString(joined) {
		t.Fatalf("at the top the menu should mark only the five hidden items below:\n%s", joined)
	}
	for range 13 {
		m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	}
	bottom := plainLines(m.View().Content)
	joined := strings.Join(bottom, "\n")
	if !strings.Contains(joined, "▸ formula-13") {
		t.Fatalf("the cursor's item is not drawn:\n%s", joined)
	}
	if !scrollUpMark.MatchString(joined) || scrollDownMark.MatchString(joined) {
		t.Fatalf("at the bottom the menu should mark only the five hidden items above:\n%s", joined)
	}
	if strings.Contains(joined, "formula-0") {
		t.Fatalf("an item scrolled out of the window is still drawn:\n%s", joined)
	}
	if len(bottom) != len(top) || len(bottom) > dialogMaxRows {
		t.Fatalf("the box is %d rows at the bottom, %d at the top, want the same and at most %d",
			len(bottom), len(top), dialogMaxRows)
	}
}

// TestMenuModalShortcutsCoverTheFirstTen (LB-51 FC3): only items 0-9 carry a
// digit label, and the items past them keep the label column, so the names
// still start in one column.
func TestMenuModalShortcutsCoverTheFirstTen(t *testing.T) {
	m := NewMenuModal("Test", menuItems(15))
	for range 13 {
		m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	}
	lines := plainLines(m.View().Content)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "9: ") {
		t.Fatalf("item 9 has no digit label:\n%s", joined)
	}
	if strings.Contains(joined, "10:") || strings.Contains(joined, "14:") {
		t.Fatalf("items past the ninth carry a digit label:\n%s", joined)
	}
	col := -1
	for _, line := range lines {
		i := strings.Index(line, "formula-")
		if i < 0 {
			continue
		}
		at := lipgloss.Width(line[:i])
		if col >= 0 && at != col {
			t.Fatalf("item names start at columns %d and %d:\n%s", col, at, joined)
		}
		col = at
	}
}

// TestMenuModalListKeepsItsColumnAsItScrolls (LB-51 D1): a long name past the
// first window does not move the cursor's column when the window scrolls to it,
// because every row is padded to the widest item, not the widest visible one.
func TestMenuModalListKeepsItsColumnAsItScrolls(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("f%d", i)
	}
	items[15] = "a-very-long-formula-name-here"
	m := NewMenuModal("Test", items)
	cursorCol := func() int {
		for _, line := range plainLines(m.View().Content) {
			if at := strings.Index(line, "▸"); at >= 0 {
				return lipgloss.Width(line[:at])
			}
		}
		t.Fatalf("the cursor's row is not drawn:\n%s", strings.Join(plainLines(m.View().Content), "\n"))
		return -1
	}
	top := cursorCol()
	for range 19 {
		m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	}
	if bottom := cursorCol(); bottom != top {
		t.Fatalf("the cursor moved from column %d at item 0 to %d at item 19", top, bottom)
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
