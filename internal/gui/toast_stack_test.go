package gui

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// colOf is the cell column where text starts in a frame row. Columns count
// cells, so the multi-byte box borders before it are counted once each.
func colOf(row []rune, text string) int {
	s := string(row)
	return utf8.RuneCountInString(s[:strings.Index(s, text)])
}

// rowOf is the index of the first frame row that shows text, or -1.
func rowOf(rows [][]rune, text string) int {
	for i, row := range rows {
		if strings.Contains(string(row), text) {
			return i
		}
	}
	return -1
}

// stackedModel is a model at w x h with three toasts pushed in order.
func stackedModel(t *testing.T, w, h int) *Model {
	t.Helper()
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
	m.icons = style.UnicodeIcons
	m.setToast("first done", modal.ToastSuccess)
	m.setToast("second note", modal.ToastInfo)
	m.setToast("third warn", modal.ToastWarning)
	return m
}

// TestToastStackNewestAtBottom locks AC-04.2: three toasts stack, the newest
// at the bottom, all flush right above the footer.
func TestToastStackNewestAtBottom(t *testing.T) {
	m := stackedModel(t, 140, 40)
	rows := plainRows(m.View().Content)
	first, second, third := rowOf(rows, "first done"), rowOf(rows, "second note"), rowOf(rows, "third warn")
	if first < 0 || second < 0 || third < 0 {
		t.Fatalf("a toast is missing from the frame: %d %d %d", first, second, third)
	}
	if !(first < second && second < third) {
		t.Fatalf("toasts are not stacked oldest-first: rows %d, %d, %d", first, second, third)
	}
	if third != 40-bottomBarRows-2 {
		t.Fatalf("the newest toast's text is on row %d, want the row above its bottom border on row %d", third, 40-bottomBarRows-2)
	}
	for _, r := range []int{first, second, third} {
		if r == 0 || r >= 40-bottomBarRows {
			t.Fatalf("a toast covers the header or footer at row %d", r)
		}
	}
}

// TestFourthToastEvictsOldest locks AC-04.2: the stack holds three, and a
// fourth pushes the oldest out.
func TestFourthToastEvictsOldest(t *testing.T) {
	m := stackedModel(t, 140, 40)
	m.setToast("fourth info", modal.ToastInfo)
	if len(m.toasts) != maxToasts {
		t.Fatalf("stack holds %d toasts, want %d", len(m.toasts), maxToasts)
	}
	got := []string{}
	for _, toast := range m.toasts {
		got = append(got, toast.Message)
	}
	want := []string{"second note", "third warn", "fourth info"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("stack = %q, want %q", got, want)
	}
}

// TestToastStackOneTickChain locks AC-04.2 and the tick rule: a stack has one
// dismissal chain, armed by the first toast, and the chain stops when the
// stack is empty.
func TestToastStackOneTickChain(t *testing.T) {
	m := newTestModel()
	if cmd := m.setToast("one", modal.ToastSuccess); cmd == nil {
		t.Fatal("the first toast must arm the tick")
	}
	if cmd := m.setToast("two", modal.ToastSuccess); cmd != nil {
		t.Fatal("a second toast must not start another tick chain")
	}
	nm, cmd := m.Update(modal.ToastTickMsg{})
	if cmd == nil {
		t.Fatal("the tick must re-arm while toasts are alive")
	}
	m = asModel(nm)
	m.toasts = nil
	nm, cmd = m.Update(modal.ToastTickMsg{})
	if cmd != nil || asModel(nm).toastTicking {
		t.Fatal("an empty stack must stop the tick chain")
	}
}

// TestToastTickEvictsByLifetime locks AC-04.2: each toast leaves on its own
// lifetime, and a live one stays.
func TestToastTickEvictsByLifetime(t *testing.T) {
	m := newTestModel()
	old := time.Now().Add(-4 * time.Second)
	m.toasts = []*modal.Toast{
		modal.NewToastAt("warn too old", modal.ToastWarning, old),
		modal.NewToastAt("error still up", modal.ToastError, old),
		modal.NewToast("success fresh", modal.ToastSuccess),
	}
	nm, cmd := m.Update(modal.ToastTickMsg{})
	got := asModel(nm).toasts
	if len(got) != 2 || got[0].Message != "error still up" || got[1].Message != "success fresh" {
		var msgs []string
		for _, toast := range got {
			msgs = append(msgs, toast.Message)
		}
		t.Fatalf("after the tick the stack = %q, want the error (6 s) and the fresh success", msgs)
	}
	if cmd == nil {
		t.Fatal("live toasts must keep the tick running")
	}
}

// TestToastClickDismisses locks AC-04.3: a click on a toast dismisses that
// toast, and a click elsewhere does not.
func TestToastClickDismisses(t *testing.T) {
	m := stackedModel(t, 140, 40)
	m.cfg.GUI.Mouse = true
	rows := plainRows(m.View().Content)
	row := rowOf(rows, "second note")
	col := colOf(rows[row], "second note")

	nm, _ := m.Update(tea.MouseClickMsg{X: col + 2, Y: row, Button: tea.MouseLeft})
	if got := asModel(nm).toasts; len(got) != 2 || got[0].Message != "first done" || got[1].Message != "third warn" {
		t.Fatalf("clicking the middle toast left %d toasts, want the other two", len(got))
	}

	nm, _ = m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if got := asModel(nm).toasts; len(got) != 3 {
		t.Fatalf("a click outside the toasts dismissed one: %d left", len(got))
	}
}

// TestToastsIgnoreEscAndX locks AC-04.3: x and esc do not dismiss toasts.
func TestToastsIgnoreEscAndX(t *testing.T) {
	m := stackedModel(t, 140, 40)
	m = asModel(sendKeyTo(m, tea.KeyPressMsg{Code: tea.KeyEsc}))
	m = asModel(sendKeyTo(m, tea.KeyPressMsg{Code: 'x', Text: "x"}))
	if len(m.toasts) != 3 {
		t.Fatalf("x or esc dismissed a toast: %d left", len(m.toasts))
	}
}

// sendKeyTo delivers one key press and returns the model it leaves.
func sendKeyTo(m *Model, k tea.KeyPressMsg) tea.Model {
	nm, _ := m.Update(k)
	return nm
}

// TestToastStackFitsAt80x24 locks AC-04.4-5: three toasts at 80x24 keep the
// frame exactly 80x24, leave the header and footer alone, and sit on the body.
func TestToastStackFitsAt80x24(t *testing.T) {
	w, h := 80, 24
	base := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
	m := stackedModel(t, w, h)
	view := m.View().Content
	assertFrameSize(t, view, w, h)

	before := plainRows(base.View().Content)
	after := plainRows(view)
	if string(after[0]) != string(before[0]) {
		t.Fatalf("the header changed with toasts showing:\nwant %q\n got %q", string(before[0]), string(after[0]))
	}
	if string(after[h-1]) != string(before[h-1]) {
		t.Fatalf("the footer changed with toasts showing:\nwant %q\n got %q", string(before[h-1]), string(after[h-1]))
	}
	if rowOf(after, "third warn") < 0 || rowOf(after, "first done") < 0 {
		t.Fatal("the stack is missing from the 80x24 frame")
	}
}

// TestSuccessToastDrawsBrewsCheckmark (lb25 fix r2): a success toast draws the
// ✔ brew prints, in both sets, never ✓.
func TestSuccessToastDrawsBrewsCheckmark(t *testing.T) {
	for name, icons := range map[string]style.Icons{"nerd": style.NerdIcons, "unicode": style.UnicodeIcons} {
		m := newTestModel()
		m.width = 80
		m.icons = icons
		m.toasts = []*modal.Toast{modal.NewToast("msg", modal.ToastSuccess)}
		got := strings.Join(plainLinesOf(m.toastBoxes()[0]), "")
		if !strings.Contains(got, "✔ msg") || strings.Contains(got, "✓") {
			t.Errorf("%s set: success toast %q should draw ✔ and no ✓", name, got)
		}
	}
}

// TestToastIconsFollowTypeAndSetting locks AC-04.1: each type draws its icon,
// and under show_icons: false the toast draws none.
func TestToastIconsFollowTypeAndSetting(t *testing.T) {
	cases := []struct {
		typ  modal.ToastType
		icon string
	}{
		{modal.ToastSuccess, style.UnicodeIcons.OK},
		{modal.ToastError, style.UnicodeIcons.Error},
		{modal.ToastWarning, style.UnicodeIcons.Warn},
		{modal.ToastInfo, style.UnicodeIcons.Info},
	}
	for _, tc := range cases {
		m := newTestModel()
		m.width = 80
		m.icons = style.UnicodeIcons
		m.toasts = []*modal.Toast{modal.NewToast("msg", tc.typ)}
		if got := strings.Join(plainLinesOf(m.toastBoxes()[0]), ""); !strings.Contains(got, tc.icon+" msg") {
			t.Errorf("type %v: box %q lacks its icon %q", tc.typ, got, tc.icon)
		}
	}

	m := newTestModel()
	m.width = 80
	m.icons = style.Icons{}
	m.toasts = []*modal.Toast{modal.NewToast("msg", modal.ToastError)}
	if got := strings.Join(plainLinesOf(m.toastBoxes()[0]), ""); strings.Contains(got, "✗") || !strings.Contains(got, "msg") {
		t.Errorf("show_icons false: box %q still draws an icon", got)
	}
}

// TestToastMessageIsCutWithEllipsis locks AC-04.1: a long message is cut with
// an ellipsis to min(frame/2, 60) cells, so at 80 columns it is 40.
func TestToastMessageIsCutWithEllipsis(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.toasts = []*modal.Toast{modal.NewToast(strings.Repeat("long ", 30), modal.ToastInfo)}
	text := strings.Join(plainLinesOf(m.toastBoxes()[0]), "\n")
	if !strings.Contains(text, "…") {
		t.Fatalf("long message was not cut:\n%s", text)
	}
	if got := toastMessageWidth(80); got != 40 {
		t.Fatalf("toastMessageWidth(80) = %d, want 40", got)
	}
	if got := toastMessageWidth(200); got != 60 {
		t.Fatalf("toastMessageWidth(200) = %d, want 60", got)
	}
}

// plainLinesOf is a rendered box as plain text, one string per row.
func plainLinesOf(box string) []string {
	return strings.Split(sgr.ReplaceAllString(box, ""), "\n")
}

// TestToastStackLeavesCellsBesideBoxesAlone locks F-r1-01: with toasts of
// mixed widths, only the cells inside a toast box differ from the base frame.
// A narrower toast must not blank the cells beside it.
func TestToastStackLeavesCellsBesideBoxesAlone(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 40}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			w, h := sz[0], sz[1]
			base := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
			withToasts := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
			withToasts.toasts = []*modal.Toast{
				modal.NewToast("ok", modal.ToastSuccess),
				modal.NewToast("a much longer message that is wider than the others", modal.ToastError),
				modal.NewToast("mid", modal.ToastWarning),
			}
			after := plainRows(withToasts.View().Content)
			before := plainRows(base.View().Content)
			assertFrameSize(t, withToasts.View().Content, w, h)

			placements := withToasts.toastPlacements()
			inBox := func(r, c int) bool {
				for _, p := range placements {
					bw, bh := lipgloss.Width(p.box), lipgloss.Height(p.box)
					if r >= p.y && r < p.y+bh && c >= p.x && c < p.x+bw {
						return true
					}
				}
				return false
			}
			for r := range before {
				for c := range before[r] {
					if !inBox(r, c) && before[r][c] != after[r][c] {
						t.Fatalf("cell (%d,%d) outside every toast box changed: %q -> %q", c, r, before[r][c], after[r][c])
					}
				}
			}
			for _, p := range placements {
				for i, row := range plainRows(p.box) {
					if got, want := string(after[p.y+i][p.x:p.x+len(row)]), string(row); got != want {
						t.Fatalf("toast box row %d is %q, want %q", i, got, want)
					}
				}
			}
		})
	}
}

// newestToast is the toast drawn lowest in the stack, or nil when there is none.
func (m Model) newestToast() *modal.Toast {
	if len(m.toasts) == 0 {
		return nil
	}
	return m.toasts[len(m.toasts)-1]
}
