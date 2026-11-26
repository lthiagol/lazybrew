package modal

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
)

type MenuModal struct {
	title     string
	items     []string
	selected  int
	done      bool
	cancelled bool
}

func NewMenuModal(title string, items []string) *MenuModal {
	return &MenuModal{
		title: title,
		items: items,
	}
}

func (m *MenuModal) Init() tea.Cmd {
	return nil
}

func (m *MenuModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			m.cancelled = true
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "enter":
			m.done = true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// The key's text can be empty, so the digit comes from msg.String(),
			// which the case above has checked is one digit.
			idx, _ := strconv.Atoi(msg.String())
			if idx < len(m.items) {
				m.selected = idx
				m.done = true
			}
		}
	}
	return m, nil
}

// window is the items the box shows: all of them when they fit the body, else
// a run of dialogBodyRows-2 that holds the cursor. The two rows the run leaves
// are the marks for the items hidden above and below.
func (m *MenuModal) window() (start, end int) {
	if len(m.items) <= dialogBodyRows {
		return 0, len(m.items)
	}
	size := dialogBodyRows - 2
	start = min(max(m.selected-size/2, 0), len(m.items)-size)
	return start, start + size
}

// scrollMark is the row that says how many items are hidden on one side of the
// window. It is blank when none are, so the box keeps its height.
func scrollMark(arrow string, hidden int) string {
	if hidden == 0 {
		return ""
	}
	return style.SubtleText.Render(arrow + " " + strconv.Itoa(hidden) + " more")
}

// row is the line of item i: its digit label, the cursor and the name, clipped
// to the box.
func (m *MenuModal) row(i int) string {
	prefix := "  "
	itemStyle := style.NormalItem
	if i == m.selected {
		prefix = "▸ "
		itemStyle = style.SelectedItem
	}
	// Only items 0-9 have a digit key; the rest keep the label's column blank.
	shortcut := "   "
	if i < 10 {
		shortcut = style.SubtleText.Render(strconv.Itoa(i) + ": ")
	}
	return textutil.Clip(itemStyle.Render(shortcut+prefix+m.items[i]), innerWidth(formWidth))
}

func (m *MenuModal) View() tea.View {
	start, end := m.window()
	scrolling := end-start < len(m.items)
	// Every row is padded to the width of the widest row of any item, and of the
	// widest mark, not only the visible ones. The list then keeps its column as
	// the window moves, and the box keeps its width.
	width := 0
	for i := range m.items {
		width = max(width, lipgloss.Width(m.row(i)))
	}
	var lines []string
	if scrolling {
		width = max(width, lipgloss.Width(scrollMark("↓", len(m.items))))
		lines = append(lines, scrollMark("↑", start))
	}
	for i := start; i < end; i++ {
		lines = append(lines, m.row(i))
	}
	if scrolling {
		lines = append(lines, scrollMark("↓", len(m.items)-end))
	}

	keys := textutil.Clip(keyLine(hintViews(style.AccentColor, m.Hints())...), innerWidth(formWidth))
	// The list reads as one block, left aligned, when each row is centred in the box.
	list := strings.Join(padBlock(lines, width), "\n")
	return tea.NewView(dialog(style.AccentColor, m.title, formWidth, []string{list}, keys))
}

// Hints are the keys the menu answers to: the arrows, enter, and esc.
func (m *MenuModal) Hints() []Hint {
	return []Hint{{Key: "↑↓", Desc: "move", Keys: "up k down j"}, {Key: "enter", Desc: "choose", Keys: "enter"}, hintCancel}
}

func (m *MenuModal) Done() bool      { return m.done }
func (m *MenuModal) Cancelled() bool { return m.cancelled }
func (m *MenuModal) Result() interface{} {
	return &MenuResult{SelectedIndex: m.selected, Cancelled: m.cancelled}
}
