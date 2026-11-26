package gui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
)

// previewCache is what the Cleanup and Unused tabs read. It is typed, not text:
// the path column is cut to the width each draw has, so the tab cannot be a
// string made once. A nil read has not landed yet. R clears it with tabContent.
type previewCache struct {
	cleanup    *brew.CleanupPreview
	cleanupErr error
	unused     *brew.AutoremovePreview
	unusedErr  error
	// cleanupFetching and unusedFetching mark a read out, so a look at the tab
	// while it is out starts no second brew dry run (FB9).
	cleanupFetching bool
	unusedFetching  bool
	// gen numbers the cache. Clearing it or dropping a read moves it on, and a
	// read that lands from an older gen is dropped (FB9).
	gen int
}

func (p previewCache) cleanupLoaded() bool { return p.cleanup != nil || p.cleanupErr != nil }
func (p previewCache) unusedLoaded() bool  { return p.unused != nil || p.unusedErr != nil }

// isPreviewTab is whether a tab is one of the two previews of brew's dry runs.
func isPreviewTab(panel PanelID, tab int) bool {
	return panel == PanelStatus && (tab == statusCleanupTab || tab == statusUnusedTab)
}

// loadPreviewTab starts the read of the Cleanup or Unused tab when it has none.
// They read brew, not the dashboard, so they load before the dashboard has.
func (m *Model) loadPreviewTab() tea.Cmd {
	switch m.activeTab {
	case statusCleanupTab:
		if !m.previews.cleanupLoaded() && !m.previews.cleanupFetching {
			m.previews.cleanupFetching = true
			return cleanupPreviewCmd(m.client, m.previews.gen)
		}
	case statusUnusedTab:
		if !m.previews.unusedLoaded() && !m.previews.unusedFetching {
			m.previews.unusedFetching = true
			return autoremovePreviewCmd(m.client, m.previews.gen)
		}
	}
	return nil
}

// cleanupPreviewCmd runs `brew cleanup -n` as a bounded read. It changes nothing,
// so it does not go through the task manager. gen is the cache's gen it was started
// under.
func cleanupPreviewCmd(client *brew.Client, gen int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTabContentTimeout)
		defer cancel()
		p, err := client.Diagnostics.CleanupPreview(ctx)
		return CleanupPreviewMsg{Preview: p, Err: timeoutError(ctx, "cleanup", fetchTabContentTimeout, err), gen: gen}
	}
}

// autoremovePreviewCmd runs `brew autoremove -n` as a bounded read.
func autoremovePreviewCmd(client *brew.Client, gen int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTabContentTimeout)
		defer cancel()
		p, err := client.Diagnostics.AutoremovePreview(ctx)
		return AutoremovePreviewMsg{Preview: p, Err: timeoutError(ctx, "autoremove", fetchTabContentTimeout, err), gen: gen}
	}
}

// landCleanup and landUnused store a read's result. A failed read keeps its error
// and no preview, so the tab shows the error rather than a spinner. A read from an
// older gen is stale and is dropped; it leaves the current read's flag alone.
func (p *previewCache) landCleanup(msg CleanupPreviewMsg) {
	if msg.gen != p.gen {
		return
	}
	p.cleanupFetching = false
	if msg.Err != nil {
		p.cleanup, p.cleanupErr = nil, msg.Err
		return
	}
	p.cleanup, p.cleanupErr = &msg.Preview, nil
}

func (p *previewCache) landUnused(msg AutoremovePreviewMsg) {
	if msg.gen != p.gen {
		return
	}
	p.unusedFetching = false
	if msg.Err != nil {
		p.unused, p.unusedErr = nil, msg.Err
		return
	}
	p.unused, p.unusedErr = &msg.Preview, nil
}

// dropPreview forgets one tab's read, so opening the tab reads brew again. Every
// read out goes stale with it, so both flags clear: the other tab reads again when
// it is next opened.
func (p *previewCache) dropPreview(tab int) {
	p.gen++
	p.cleanupFetching, p.unusedFetching = false, false
	switch tab {
	case statusCleanupTab:
		p.cleanup, p.cleanupErr = nil, nil
	case statusUnusedTab:
		p.unused, p.unusedErr = nil, nil
	}
}

// previewKey is c (Cleanup) and A (Unused) on Status. Off its tab, the key opens
// the tab and reads brew again, as d and m do. On its tab, the key asks to run the
// real command, and the run goes through the task manager after the confirm.
func (m Model) previewKey(tab int) (tea.Model, tea.Cmd) {
	if m.activeTab != tab {
		m.activeTab = tab
		m.previews.dropPreview(tab)
		cmd := m.loadTabContent()
		return m, cmd
	}
	if tab == statusCleanupTab {
		return m.confirmCleanup()
	}
	return m.confirmAutoremove()
}

// confirmCleanup asks before brew cleanup runs. It names brew's total, and says
// that cleanup also removes the unused formulae the tab lists.
func (m Model) confirmCleanup() (tea.Model, tea.Cmd) {
	m.pendingAction = "cleanup"
	question := "Run brew cleanup?"
	detail := "Deletes the files this tab lists."
	if p := m.previews.cleanup; p != nil {
		if p.Total != "" {
			question = "Run brew cleanup, freeing about " + p.Total + "?"
		}
		if n := len(p.Unused); n > 0 {
			detail += fmt.Sprintf(" Also removes %d unused %s.", n, plural(n, "formula", "formulae"))
		}
	}
	m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Cleanup", question).WithDetail(detail)
	return m, m.activeModal.Init()
}

// confirmAutoremove asks before brew autoremove uninstalls the unused formulae.
func (m Model) confirmAutoremove() (tea.Model, tea.Cmd) {
	m.pendingAction = "autoremove"
	question := "Uninstall the unused formulae?"
	if p := m.previews.unused; p != nil && len(p.Formulae) > 0 {
		n := len(p.Formulae)
		question = fmt.Sprintf("Uninstall %d unused %s?", n, plural(n, "formula", "formulae"))
	}
	m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Autoremove", question).
		WithDetail("Runs brew autoremove: removes the formulae this tab lists.")
	return m, m.activeModal.Init()
}

// cleanupTabView is the Cleanup tab: brew's total, then the paths biggest first.
func (m Model) cleanupTabView(width int) string {
	switch {
	case m.previews.cleanupErr != nil:
		return "Error: " + m.previews.cleanupErr.Error()
	case m.previews.cleanup == nil:
		return m.spinner.View() + style.SubtleText.Render(" Loading...")
	}
	return cleanupView(*m.previews.cleanup, width, m.icons, homeDir(), m.status.Prefix)
}

// unusedTabView is the Unused tab: the formulae autoremove would uninstall.
func (m Model) unusedTabView() string {
	switch {
	case m.previews.unusedErr != nil:
		return "Error: " + m.previews.unusedErr.Error()
	case m.previews.unused == nil:
		return m.spinner.View() + style.SubtleText.Render(" Loading...")
	}
	return unusedView(*m.previews.unused, m.icons)
}

// cleanupView draws a cleanup preview at width cells. Paths are cut in the
// middle to the room the path column has; the copy draws it wide, so it is whole.
func cleanupView(p brew.CleanupPreview, width int, icons style.Icons, home, prefix string) string {
	hint := style.SubtleText.Render(textutil.Clip(fmt.Sprintf("Press %s again to run brew cleanup", stCleanup.key), width))
	var out []string
	switch {
	case len(p.Items) > 0:
		if p.Total != "" {
			out = append(out, style.AccentText.Render("Would free ~"+p.Total))
		}
		out = append(out, hint, "")
		out = append(out, cleanupTable(p.Items, width, home, prefix)...)
	case len(p.Raw) > 0:
		out = append(out, p.Raw...)
		out = append(out, hint)
	default:
		out = append(out, style.SuccessText.Render(presentation.WithIcon(icons.OK, "Nothing to clean up")))
	}
	if n := len(p.Unused); n > 0 {
		note := fmt.Sprintf("Also removes %d unused %s (Unused tab)", n, plural(n, "formula", "formulae"))
		out = append(out, "", style.SubtleText.Render(textutil.Clip(note, width)))
	}
	return strings.Join(out, "\n")
}

// cleanupTable is the item rows: size (right-aligned), file count, then the path,
// biggest first. The largest third of the sizes is in the warning colour.
func cleanupTable(items []brew.CleanupItem, width int, home, prefix string) []string {
	sorted := append([]brew.CleanupItem(nil), items...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Bytes > sorted[j].Bytes })

	sizeW, filesW := lipgloss.Width("Size"), lipgloss.Width("Files")
	for _, it := range sorted {
		sizeW = max(sizeW, lipgloss.Width(sizeText(it)))
		filesW = max(filesW, lipgloss.Width(fileCountText(it)))
	}
	pathW := max(1, width-sizeW-filesW-2*tableGap)

	rows := []string{style.MutedText.Render(
		padLeft("Size", sizeW) + strings.Repeat(" ", tableGap) + padRight("Files", filesW) + strings.Repeat(" ", tableGap) + "Path")}
	large := (len(sorted) + 2) / 3
	for i, it := range sorted {
		size := padLeft(sizeText(it), sizeW)
		if it.Bytes > 0 && i < large {
			size = style.WarningText.Render(size)
		}
		path := cutPath(shortPath(it.Path, home, prefix), pathW)
		rows = append(rows, size+strings.Repeat(" ", tableGap)+padRight(fileCountText(it), filesW)+strings.Repeat(" ", tableGap)+path)
	}
	return rows
}

// sizeText is brew's size for a path, or a dash when brew gave none.
func sizeText(it brew.CleanupItem) string {
	if it.Size == "" {
		return "—"
	}
	return it.Size
}

// fileCountText is the file count, or for a path brew removes without a size, its
// label ("broken link"). A dash is a path with no count at all.
func fileCountText(it brew.CleanupItem) string {
	switch {
	case it.Kind != "":
		return it.Kind
	case it.Files == 1:
		return "1 file"
	case it.Files > 1:
		return fmt.Sprintf("%d files", it.Files)
	}
	return "—"
}

// unusedView draws the formulae autoremove would uninstall, one per line.
func unusedView(p brew.AutoremovePreview, icons style.Icons) string {
	if len(p.Formulae) == 0 {
		if len(p.Raw) > 0 {
			return strings.Join(p.Raw, "\n")
		}
		return style.SuccessText.Render(presentation.WithIcon(icons.OK, "Nothing unused"))
	}
	n := len(p.Formulae)
	out := []string{
		style.AccentText.Render(fmt.Sprintf("%d unused %s would be removed", n, plural(n, "formula", "formulae"))),
		style.SubtleText.Render(fmt.Sprintf("Press %s again to run brew autoremove", stAutoremove.key)),
	}
	for _, name := range p.Formulae {
		out = append(out, style.NormalItem.Render(name))
	}
	return strings.Join(out, "\n")
}

// shortPath writes a path under the Homebrew prefix as <prefix>/… and under the
// home directory as ~/…. Any other path is left whole.
func shortPath(path, home, prefix string) string {
	if prefix != "" && strings.HasPrefix(path, prefix+"/") {
		return "<prefix>" + path[len(prefix):]
	}
	if home != "" && strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

// cutPath fits a shortened path to w cells. When the last segment is wider than w,
// the directories go and that segment is cut alone, so two long file names keep
// their starts and ends. Otherwise the whole path is cut.
func cutPath(p string, w int) string {
	last := p[strings.LastIndex(p, "/")+1:]
	if len([]rune(last)) > w {
		return cutMiddle(last, w)
	}
	return cutMiddle(p, w)
}

// cutSnap is how far, in cells, a cut may move to reach a "/". Farther than that
// the cut falls where the width ends, so the room is used and not thrown away.
const cutSnap = 4

// cutMiddle cuts s to w cells with … in the middle, keeping its start and end. The
// cut moves to a "/" only when one is within cutSnap cells, so a path keeps whole
// segments at each end without losing the room it has. A cut that already falls
// after a "/" keeps the segment that follows it. A string that fits is returned as
// it is.
func cutMiddle(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w <= 1 {
		return string(r[:max(0, w)])
	}
	keep := w - 1
	headN := keep / 2
	tailN := keep - headN

	head := r[:headN]
	if i := lastRune(head, '/'); i > 0 && headN-(i+1) <= cutSnap {
		head = head[:i+1]
	}
	tailStart := len(r) - tailN
	tail := r[tailStart:]
	if i := firstRune(tail, '/'); i >= 0 && i <= cutSnap && r[tailStart-1] != '/' {
		tail = tail[i:]
	}
	return string(head) + "…" + string(tail)
}

// lastRune is the index of the last r in s, or -1.
func lastRune(s []rune, r rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == r {
			return i
		}
	}
	return -1
}

// firstRune is the index of the first r in s, or -1.
func firstRune(s []rune, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}

// padLeft and padRight pad s with spaces to w cells.
func padLeft(s string, w int) string {
	return strings.Repeat(" ", max(0, w-lipgloss.Width(s))) + s
}

func padRight(s string, w int) string {
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
}

// homeDir is the user's home directory, or "" when it has none.
func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
