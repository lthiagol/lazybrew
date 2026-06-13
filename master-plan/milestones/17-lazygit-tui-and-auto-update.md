# Milestone 17 — Lazygit-Inspired TUI & Auto-Update

> **Status:** 🔜 Planned
> **Depends on:** Milestone 15, Milestone 16
> **Enables:** Better UX for all subsequent milestones

---

## Goal

Redesign the TUI to feel more like lazygit — each sidebar panel rendered as its own bordered "box", accordion height distribution, and a `brew update` flow on startup so data is always fresh.

---

## Design Decisions

| Decision | Chosen Approach | Rationale |
|---|---|---|
| **Title placement** | Bold title INSIDE the box (first content line), not in the border line | Avoiding complex mixed-style line construction; still achieves the "boxed" look |
| **Sidebar border** | Each panel is an individual `lipgloss.RoundedBorder` box | This is the core "boxes" visual — current single outer border feels flat |
| **Accordion** | Proportional: active panel gets 40% of content area, inactive share 60% equally; minimum 4 rows active, 2 inactive | Balances visibility of active content with context from other panels |
| **renderBox** | Simple wrapper: `lipgloss.NewStyle().Border().Width(w).Height(h).BorderForeground(c).Render(content)` | No custom border-string construction; reliable, tested lipgloss behavior |
| **Auto-update blocking** | Non-blocking — UI renders immediately with loading indicators, update runs in background | No progress modal; user sees the app instantly |
| **Update indicator** | Bottom bar right section: `⟳ Updating...` / `⟳ Updated 12s ago` / `⟳ Never updated` | Minimal intrusion, always visible |
| **update_on_start default** | `false` (unchanged) | Backward compatible; opt-in via config |
| **R key behavior** | If `update_on_start=true`, `R` runs brew update first; otherwise just refreshes | Consistent with startup behavior |
| **Testing** | `newTestModel()` uses `UpdateOnStart=false` by default — all existing tests pass unchanged | No regression risk |

---

## Target Visual

```
┌─────────────────────────────────────────────────────────────────────┐
│ ┌─ Status ──────────────────┐ ┌─ Status ──────────────────────────┐ │
│ │ Dashboard                 │ │ Homebrew 4.2.22                   │ │
│ │ 45 formulae               │ │ Prefix: /opt/homebrew             │ │
│ │ 8 casks                   │ │                                   │ │
│ └───────────────────────────┘ │ ○ Dashboard  ● Config  ○ Doctor   │ │
│ ┌─ Formulae ── 45 ────────┐  │ ──────────────────────────────────│ │
│ │ ▸ ripgrep 14.1.1        │  │                                   │ │
│ │   neovim 0.10.0         │  │ HOMEBREW_VERSION: 4.2.22          │ │
│ │   python@3.12 3.12.8    │  │ HOMEBREW_PREFIX: /opt/homebrew    │ │
│ └───────────────────────────┘  │                                   │ │
│ ┌─ Casks ──── 8 ─────────┐  │                                   │ │
│ └───────────────────────────┘  └───────────────────────────────────┘ │
│ ┌─ Outdated ─ 3 ─────────┐  │
│ └───────────────────────────┘  │
│ ┌─ Taps ──── 5 ──────────┐  │
│ └───────────────────────────┘  │
├─────────────────────────────────────────────────────────────────────┤
│ j/k: ▲▼  [ ]: tabs  ?: help  q: quit       │       ⟳ Updated 12s ago │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Steps

### 17.1 — New Messages and Model Fields for Auto-Update

**What:** Add the message types and model fields that the auto-update flow needs.

**Files:**
- `internal/gui/messages.go` — 2 new message types
- `internal/gui/gui.go` — 3 new Model fields, 1 new import (`time`)

**Model changes** (`gui.go`):
```go
type Model struct {
    // ... existing fields ...

    lastUpdate     time.Time   // zero = never updated
    isUpdating     bool        // true while brew update is running
    updateOutput   []string    // lines captured during current update
}
```

**New messages** (`messages.go`):
```go
// StartUpdateMsg triggers brew update (sent from Init or R key)
type StartUpdateMsg struct{}

// UpdateCompleteMsg signals brew update finished
type UpdateCompleteMsg struct {
    Output []string  // raw lines from brew update
    Err    error
}
```

**Why separate StartUpdateMsg from Init():**
- Lets `R` key also trigger the update without duplicating the Init logic
- Makes the update flow testable as a discrete message handler

**Edge cases:**
- `lastUpdate` zero value = no update has run this session
- `isUpdating` prevents concurrent update calls (ignore R if already updating)

**Acceptance criteria:**
- [ ] `Model` compiles with new fields
- [ ] `StartUpdateMsg` and `UpdateCompleteMsg` defined
- [ ] Zero-value `lastUpdate` correctly indicates "never updated"

---

### 17.2 — Auto-Update Init and Message Handling

**What:** Wire the startup update flow and handle update completion.

**File:** `internal/gui/gui.go` — `Init()` and `Update()` switch cases

**Init flow** (`gui.go:59`):
```go
func (m Model) Init() tea.Cmd {
    if m.cfg.Brew.UpdateOnStart {
        return func() tea.Msg { return StartUpdateMsg{} }
    }
    return tea.Batch(
        fetchPanelData(m.client, PanelFormulae),
        fetchPanelData(m.client, PanelCasks),
        fetchPanelData(m.client, PanelOutdated),
        fetchPanelData(m.client, PanelTaps),
        fetchPanelData(m.client, PanelServices),
        fetchStatusData(m.client),
    )
}
```

**Key nuance**: When `UpdateOnStart = true`, we return just `StartUpdateMsg{}`, NOT the panel fetches. The update handler will trigger fetches after update completes.

**Update handler** in `Update()` (new case):
```go
case StartUpdateMsg:
    if m.isUpdating || !m.cfg.Brew.UpdateOnStart {
        return m, nil
    }
    m.isUpdating = true
    m.updateOutput = nil

    ctx, cancel := context.WithCancel(context.Background())
    return m, func() tea.Msg {
        ch, errCh := m.client.DiagnosticsWrite.Update(ctx)
        var lines []string
        if ch != nil {
            for line := range ch {
                lines = append(lines, line)
            }
        }
        var err error
        if errCh != nil {
            err = <-errCh
        }
        _ = cancel
        return UpdateCompleteMsg{Output: lines, Err: err}
    }
```

**UpdateCompleteMsg handler** in `Update()`:
```go
case UpdateCompleteMsg:
    m.isUpdating = false
    m.lastUpdate = time.Now()
    m.updateOutput = msg.Output

    // Parse for toast
    if msg.Err != nil {
        m.toast = modal.NewToast("Update: "+msg.Err.Error(), modal.ToastWarning)
    } else {
        summary := parseUpdateSummary(msg.Output)
        if summary != "" {
            m.toast = modal.NewToast(summary, modal.ToastSuccess)
        }
    }

    // Now fetch all panel data
    return m, tea.Batch(
        fetchPanelData(m.client, PanelFormulae),
        fetchPanelData(m.client, PanelCasks),
        fetchPanelData(m.client, PanelOutdated),
        fetchPanelData(m.client, PanelTaps),
        fetchPanelData(m.client, PanelServices),
        fetchStatusData(m.client),
    )
```

**R key modification** (`gui.go`, the `"R"` case):
```go
case "R":
    if m.isBusy || m.isUpdating {
        m.toast = modal.NewToast("A brew operation is already running", modal.ToastWarning)
        return m, nil
    }
    if m.cfg.Brew.UpdateOnStart {
        return m, func() tea.Msg { return StartUpdateMsg{} }
    }
    return m, func() tea.Msg { return RefreshMsg{} }
```

**Edge cases:**
- `StartUpdateMsg` while `isUpdating` or `!UpdateOnStart` → no-op
- `R` while `isUpdating` → no-op with toast warning
- Update fails → toast warning, still loads current data
- Initial `lastUpdate` is zero → bottom bar shows "⟳ Never updated"

**Acceptance criteria:**
- [ ] `update_on_start=true` sends `StartUpdateMsg` from Init
- [ ] `update_on_start=false` sends panel fetches directly (current behavior)
- [ ] Update complete triggers all panel fetches
- [ ] Update failure shows toast warning and still loads data
- [ ] `R` key with `update_on_start=true` runs update first
- [ ] `R` key with `update_on_start=false` just refreshes (current behavior)
- [ ] Duplicate `StartUpdateMsg` is ignored

---

### 17.3 — Update Output Parsing

**What:** Extract a human-readable summary from `brew update` output.

**File:** `internal/gui/commands.go` — new function `parseUpdateSummary()`

```go
func parseUpdateSummary(lines []string) string {
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        // "Already up-to-date." → "Already up to date"
        if strings.Contains(line, "Already up-to-date") {
            return "Already up to date"
        }
        // "Updated N formulae (M casks)." → "Updated N formulae"
        // "Updated N formulae." → "Updated N formulae"
        if strings.HasPrefix(line, "Updated") {
            // Strip trailing period
            clean := strings.TrimRight(line, ".")
            return clean
        }
        // "Error: ..." → "Update failed"
        if strings.HasPrefix(line, "Error:") {
            return ""
        }
    }
    return ""
}
```

**For non-English Homebrew output:** brew hardcodes English update messages internally, so this is safe. If brew ever localizes, the fallback is an empty string (no toast at all).

**Edge cases:**
- No lines → `""` (no toast shown)
- Multiple "Updated" lines → first one wins
- Already up-to-date → proper toast
- Non-English → empty string, no toast

**Acceptance criteria:**
- [ ] `"Already up-to-date."` → `"Already up to date"`
- [ ] `"Updated 3 formulae."` → `"Updated 3 formulae"`
- [ ] `"Updated 5 formulae (2 casks)."` → `"Updated 5 formulae (2 casks)"`
- [ ] `"Error: ..."` → `""`
- [ ] Empty input → `""`

---

### 17.4 — Sidebar Content Helper (`renderSidebarContent`)

**What:** Add a compact list renderer tailored for the sidebar boxes. Separated from the existing `renderList` (which stays for main panel use).

**File:** `internal/gui/panel.go` — new method `renderSidebarContent(width, maxRows int) string`

```go
// renderSidebarContent renders the panel's items compactly for sidebar boxes.
// width: content width inside the box (excluding border)
// maxRows: maximum number of item rows to show (not including title)
func (p *panelData) renderSidebarContent(width, maxRows int) string {
    if p.loading {
        return lipgloss.NewStyle().Width(width).Render(style.SubtleText.Render("..."))
    }
    if p.err != nil {
        return lipgloss.NewStyle().Width(width).Render(style.ErrorBadge.Render("!"))
    }

    count := len(p.items)
    if count == 0 {
        return style.SubtleText.Render("(empty)")
    }

    // Clamp selected/offset safely
    if p.selected >= count {
        p.selected = max(0, count-1)
    }
    if p.offset >= count {
        p.offset = max(0, count-maxRows)
    }

    visible := min(maxRows, count-p.offset)
    if visible <= 0 {
        return ""
    }

    end := p.offset + visible
    slice := p.items[p.offset:end]

    lines := make([]string, 0, visible)
    for i, item := range slice {
        idx := p.offset + i
        prefix := "  "
        itemStyle := style.NormalItem
        if idx == p.selected {
            prefix = "▸ "
            itemStyle = style.SelectedItem
        }
        // Truncate to fit width
        text := truncateWithEllipsis(prefix+item, width)
        lines = append(lines, itemStyle.Render(text))
    }

    return lipgloss.JoinVertical(lipgloss.Top, lines...)
}

// truncateWithEllipsis shortens a string to fit width, adding "..." if truncated.
// Must handle multi-byte runes (Japanese, emoji in names).
func truncateWithEllipsis(s string, maxWidth int) string {
    runes := []rune(s)
    if len(runes) <= maxWidth {
        return s
    }
    if maxWidth <= 3 {
        return string(runes[:maxWidth])
    }
    return string(runes[:maxWidth-3]) + "..."
}
```

**This does NOT replace** `renderList` — that method is unchanged for the main panel content area.

**Acceptance criteria:**
- [ ] Returns "..." for loading state
- [ ] Returns "!" for error state  
- [ ] Returns "(empty)" for 0 items
- [ ] Selected item has `▸` prefix
- [ ] Items truncated with `...` when exceeding width
- [ ] At most `maxRows` items returned
- [ ] Multi-byte runes handled correctly (no broken UTF-8)

---

### 17.5 — Accordion Height Computation

**What:** Compute content heights for the sidebar's accordion layout.

**File:** `internal/gui/render.go` — new method `computeContentHeights() []int`

```go
// computeContentHeights returns the number of CONTENT rows (inside border)
// that each sidebar panel should get. Border rows are NOT included.
// Total rows = sum(contentHeights) + 2*N = sidebarHeight
func (m Model) computeContentHeights() []int {
    n := len(m.panels)
    if n == 0 { return nil }

    // Sidebar content area = total sidebar height - border overhead
    sidebarHeight := m.height - 4  // current: app padding (top + bottom bar)
    borderOverhead := n * 2        // top + bottom border per panel
    availableRows := sidebarHeight - borderOverhead

    if availableRows < n {
        availableRows = n  // at least 1 content row per panel
    }

    minActive := 4    // title + 3 items
    minInactive := 2  // title + 1 item

    // Verify minimums fit
    needed := minActive + (n-1)*minInactive
    if availableRows < needed {
        // Shrink inactive to 1 each, active gets rest
        heights := make([]int, n)
        for i := range heights {
            if i == int(m.activePanel) {
                heights[i] = max(1, availableRows-(n-1))
            } else {
                heights[i] = 1
            }
        }
        return heights
    }

    // Distribute: active 40%, inactive share 60%
    activeRows := availableRows * 40 / 100
    if activeRows < minActive { activeRows = minActive }

    remaining := availableRows - activeRows
    inactiveRows := remaining / max(1, n-1)
    if inactiveRows < minInactive { inactiveRows = minInactive }

    heights := make([]int, n)
    for i := range heights {
        if i == int(m.activePanel) {
            heights[i] = activeRows
        } else {
            heights[i] = inactiveRows
        }
    }

    return heights
}
```

**Test matrix:**

| Terminal (cols x rows) | Panels | Active rows | Inactive rows | Total including borders |
|---|---|---|---|---|
| 120x40 | 7 | 7 | 3 | 7+2 + 6*(3+2) = 9 + 30 = 39 ✓ |
| 80x24 | 7 | 4 | 2 | 6 + 6*4 = 30 > 24 → shrink |
| 80x20 | 7 | too small → active=2, inactive=1 | 2+2 + 6*(1+2) = 4+18 = 22 ≈ 20 |

**The minInactive=2 case:** "title + 1 item" means the box shows the title line and one item. Even in collapsed state, the user can see what data is in the panel.

**Acceptance criteria:**
- [ ] Returns correct count (n elements)
- [ ] Active panel gets ~40% of available content rows
- [ ] Inactive panels share remaining rows equally
- [ ] Minimum enforced: active ≥ 4, inactive ≥ 2
- [ ] Degrades gracefully when terminal too small (minimum 1 per panel)
- [ ] All 7 panels render within sidebar height

---

### 17.6 — Sidebar Per-Panel Boxes (The Big Visual Change)

**What:** Replace the single-border sidebar with per-panel bordered boxes.

**File:** `internal/gui/render.go` — rewrite `renderSidebar()`
**New file:** `internal/gui/box.go` — `renderBox()` helper

**`renderBox`** (`box.go`):
```go
package gui

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/thiago/lazybrew/internal/gui/style"
)

// renderBox wraps content in a lipgloss rounded-border box.
// width, height: content area dimensions (excluding border)
// active: controls border color (accent vs subtle)
func renderBox(content string, width, height int, active bool) string {
    borderColor := style.SubtleColor
    if active {
        borderColor = style.AccentColor
    }

    return lipgloss.NewStyle().
        Width(width).
        Height(height).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(borderColor).
        Render(content)
}
```

**`renderSidebar`** rewrite (`render.go`):
```go
func (m Model) renderSidebar() string {
    sw := sidebarWidth(m.cfg, m.width)
    contentWidth := sw - 2  // minus left/right border chars inside

    heights := m.computeContentHeights()

    var boxes []string
    for i, p := range m.panels {
        // Build title line
        title := p.title
        if p.loading {
            title += style.SubtleText.Render("  …")
        } else if count := p.itemCount(); count > 0 {
            title += style.SubtleText.Render("  " + strconv.Itoa(count))
        }
        titleLine := style.PanelTitle.Render(title)

        // Build items content
        itemsMaxRows := max(0, heights[i]-1)  // reserve 1 row for title
        itemsContent := p.renderSidebarContent(contentWidth, itemsMaxRows)

        // Full content = title + items
        fullContent := lipgloss.JoinVertical(lipgloss.Top, titleLine, itemsContent)

        // Render box
        box := renderBox(fullContent, contentWidth, heights[i], i == int(m.activePanel))
        boxes = append(boxes, box)
    }

    return lipgloss.JoinVertical(lipgloss.Top, boxes...)
}
```

**What this changes from the current layout:**
- Current: single outer border around all sidebar items
- New: each panel is individually bordered
- Current: `renderSidebarItem()` formats as `▸ Formulae (45)`
- New: `titleLine` inside box, items rendered separately
- Current: all panels same height
- New: accordion heights

**What stays the same:**
- `sidebarWidth()` calculation unchanged
- `PanelID` order unchanged (matches number keys 1-7)
- Item selection and offset logic unchanged

**Interaction with existing rendering:**
- `renderSidebar` no longer calls `p.renderSidebarItem()`
- `renderSidebarItem` can remain (it's not called from anywhere else currently) or be removed
- `renderContent` and `renderMainPanel` are unchanged
- The `p.width`/`p.height` fields on panelData are no longer used by sidebar; they're still used by `renderList` for the main panel width

**Acceptance criteria:**
- [ ] Each sidebar panel renders as a bordered box
- [ ] Active panel has accent-colored border
- [ ] Inactive panels have subtle-colored border
- [ ] Title line shows panel name + item count
- [ ] Active panel shows more items (accordion)
- [ ] Selected item has `▸` prefix inside its box
- [ ] Sidebar total height matches `m.height - 4`
- [ ] No visual clipping at any terminal size ≥ 80x24

---

### 17.7 — Bottom Bar with Update Status

**What:** Redesign the bottom bar to show key hints on the left and update status on the right.

**File:** `internal/gui/render.go` — rewrite `renderBottomBar()`

```go
func (m Model) renderBottomBar() string {
    // Left section: compact hints
    hints := panelHints(m.activePanel)
    globalHints := []keyHint{
        {"j/k", "▲▼"}, {"[ ]", "tabs"}, {"?", "help"}, {"q", "quit"},
    }

    var leftParts []string
    for _, h := range hints {
        leftParts = append(leftParts,
            style.HintKey.Render(h.key)+" "+style.HintDesc.Render(h.desc))
    }
    leftParts = append(leftParts, style.SubtleText.Render("│"))
    for _, h := range globalHints {
        leftParts = append(leftParts,
            style.HintKey.Render(h.key)+" "+style.HintDesc.Render(h.desc))
    }
    leftBar := lipgloss.JoinHorizontal(lipgloss.Top, leftParts...)

    // Right section: update status
    rightBar := m.renderUpdateStatus()

    // Combine left + right with flexible space
    // Use lipgloss.JoinHorizontal with a spacer or use Width to push right
    combined := lipgloss.JoinHorizontal(lipgloss.Left,
        leftBar,
        lipgloss.NewStyle().Width(m.width-2-lipgloss.Width(leftBar)-lipgloss.Width(rightBar)).Render(""),
        rightBar,
    )

    return lipgloss.NewStyle().
        Width(m.width - 2).
        Padding(0, 1).
        Render(combined)
}

func (m Model) renderUpdateStatus() string {
    if m.isUpdating {
        return style.HintKey.Render("⟳") + " " + style.HintDesc.Render("Updating...")
    }
    if m.lastUpdate.IsZero() {
        return style.SubtleText.Render("⟳ Never updated")
    }

    elapsed := time.Since(m.lastUpdate)
    var timeStr string
    switch {
    case elapsed < time.Minute:
        timeStr = fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
    case elapsed < time.Hour:
        timeStr = fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
    case elapsed < 24*time.Hour:
        timeStr = fmt.Sprintf("%dh ago", int(elapsed.Hours()))
    default:
        timeStr = fmt.Sprintf("%dd ago", int(elapsed.Hours()/24))
    }
    return style.SuccessColor.Render("⟳ Updated " + timeStr)
}
```

Wait — `SuccessColor` is a `lipgloss.Color`, not a style. Let me fix:

```go
return lipgloss.NewStyle().Foreground(style.SuccessColor).Render("⟳ Updated " + timeStr)
```

**Timestamp ticker:** We need a periodic tick to update the "12s ago" text. Without it, the timestamp would only update on keypress (when View() is called). Add a background tick:

```go
// New message type
type UpdateTimestampTickMsg struct{}

// Start ticker when update completes
case UpdateCompleteMsg:
    // ... existing code ...
    return m, tea.Batch(
        // ... panel fetches ...
        updateTimestampTick(),
    )

// Ticker command
func updateTimestampTick() tea.Cmd {
    return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
        return UpdateTimestampTickMsg{}
    })
}

// Handler — triggers re-render
case UpdateTimestampTickMsg:
    if !m.lastUpdate.IsZero() {
        return m, updateTimestampTick()  // re-arm
    }
    return m, nil
```

**Why 10 seconds instead of 1:** The timestamp resolution is seconds, but nobody needs "up-to-the-second" accuracy. 10s interval reduces CPU/tick overhead.

**Acceptance criteria:**
- [ ] Left bar shows hints for current panel + global hints
- [ ] Right bar shows update status
- [ ] "Never updated" when `lastUpdate` is zero
- [ ] "Updating..." when update is in progress
- [ ] "Updated 12s ago" after update completes
- [ ] Timestamp text updates every 10 seconds via ticker
- [ ] Ticker stops when model is destroyed (Bubble Tea handles this)

---

### 17.8 — Main Panel Title Line

**What:** Add a panel name + tab name title line at the top of the main content area.

**File:** `internal/gui/render.go` — modify `renderMainPanel()`

```go
func (m Model) renderMainPanel() string {
    sw := sidebarWidth(m.cfg, m.width)
    mw := m.width - sw - 4
    mh := m.height - 4

    // Breadcrumb: PanelName › TabName
    panelName := m.panels[m.activePanel].title
    tabName := ""
    if len(m.tabs) > 0 && m.activeTab < len(m.tabs) {
        tabName = m.tabs[m.activeTab].name
    }
    breadcrumb := style.PanelTitle.Render(panelName)
    if tabName != "" {
        breadcrumb += style.SubtleText.Render(" › ") + style.AccentText.Render(tabName)
    }

    tabBar := m.renderTabBar(mw)
    content := m.renderContent(mw, mh-4)  // -4 for breadcrumb, tab bar, and gaps

    panel := lipgloss.JoinVertical(lipgloss.Top,
        breadcrumb,
        tabBar,
        content,
    )

    return lipgloss.NewStyle().
        Width(mw + 2).
        Height(mh).
        Render(style.ActiveBorder.Render(panel))
}
```

The height adjustment: previously `mh-3` (tab bar + content). Now `mh-4` (breadcrumb + tab bar + content). The breadcrumb takes 1 row.

**Acceptance criteria:**
- [ ] Main panel shows `PanelName › TabName` at top
- [ ] Breadcrumb fits with tab bar below it
- [ ] Content area shrinks by 1 row (breadcrumb takes space)
- [ ] Tab bar still renders and switches correctly

---

### 17.9 — Empty, Loading, and Error States Inside Sidebar Boxes

**What:** Polish what each sidebar box shows when the panel is loading, has no data, or has an error.

**Files:** `internal/gui/panel.go` — `renderSidebarContent()`, already covered above but add explicit state tests.

**States:**

| State | Title line | Content |
|---|---|---|
| Loading | `Formulae  …` | `...` (animated in future) |
| Loaded, data | `Formulae  45` | Item list |
| Loaded, empty | `Formulae` | `(empty)` in subtle |
| Error | `Formulae` | `!` in error color |
| Updating after update | Active state, shows `Updating...` | Same as loading |

**Acceptance criteria:**
- [ ] Loading: count shows `…`, content shows `...`
- [ ] Empty: no count in title, content shows `(empty)`
- [ ] Error: no count in title, content shows `!`
- [ ] All states render inside the bordered box without breaking layout

---

### 17.10 — Visual Integration (No gaps, proper alignment)

**What:** Fix any visual gaps between the sidebar and main panel, ensure the outer app frame lines up correctly.

**File:** `internal/gui/gui.go` — `View()`

**Current `View()`:**
```go
body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
```

This should still work. The sidebar now has `sw` width (like before), and the main panel takes the rest.

But there's a potential alignment issue: the sidebar boxes each have their own border. When joined horizontally with the main panel, the top borders may not align perfectly. Let me check:

Sidebar: `renderBox(fullContent, contentWidth, height, active)` produces:
```
┌─────────────────────┐  ← top border line
│ title               │
│ item 1              │
└─────────────────────┘  ← bottom border
┌─────────────────────┐  ← top border
│ title               │
└─────────────────────┘
```

Main panel: `style.ActiveBorder.Render(panel)` produces:
```
┌─────────────────────┐  ← top border
│ content             │
└─────────────────────┘
```

When joined horizontally, the tops align at the same row. But the sidebar boxes have multiple top borders, creating "staggered" tops. This is the intended lazygit look — each panel is its own box, so they visually stack.

The join should work correctly: lipgloss will align the tops by default (lipgloss.Top).

**Check:** The sidebar uses `sidebarWidth()` and the main panel calculates `mw := m.width - sw - 4`. The `-4` accounts for the sidebar's border width (2) and the main panel's border width (2) + a 2-char gap. Wait, let me re-read:

Current code:
```go
sw := sidebarWidth(m.cfg, m.width)
// sidebar renders at width sw
// sidebar is rendered with: lipgloss.NewStyle().Width(sw).Height(sh).Render(border.Render(list))
// border.Render(list) adds 2 to both width and height (border chars)

mw := m.width - sw - 4
// main panel renders at width mw
// then: lipgloss.NewStyle().Width(mw + 2).Height(mh).Render(style.ActiveBorder.Render(panel))
```

Current width accounting:
- Sidebar total visual width: `sw` (includes border chars, since border is rendered inside)
- Main panel total visual width: `mw + 2` (includes border chars)
- Total: `sw + mw + 2 = sw + (m.width - sw - 4) + 2 = m.width - 2`

Hmm, that's 2 characters short. Let me check... Actually, let me look more carefully:

```go
// renderSidebar returns:
lipgloss.NewStyle().Width(sw).Height(sh).Render(border.Render(list))

// border.Render(list) produces content that is 2 cols wider than list (border chars)
// lipgloss.NewStyle().Width(sw) clips or pads to sw
```

So the sidebar visual width is `sw`. The main panel visual width is `mw + 2`. And the `body` is `lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)`.

Total visual width: `sw + (mw + 2) = sw + (m.width - sw - 4) + 2 = m.width - 2`

And then in `View()`:
```go
full = lipgloss.JoinVertical(lipgloss.Top, body, bottomBar)
```

The bottom bar is `m.width - 2`. So body is `m.width - 2` and bottom bar is `m.width - 2`. They should align.

Wait, but there's `sidebarWidth` which returns a percentage of total width. If `sidebarWidth` returns `sw`, and the sidebar visual is `sw`, and the main panel is `m.width - sw - 4` (then +2 for border), the main panel visual is `m.width - sw - 2`. So total = `sw + m.width - sw - 2 = m.width - 2`.

But the bottom bar is `m.width - 2`. And `body` joins sidebar (sw) + main (m.width - sw - 2) = `m.width - 2`. They match. Good.

For the new sidebar, each box uses `renderBox` with `Width(contentWidth)` where `contentWidth = sw - 2`. The box visual is `contentWidth + 2 = sw`. Same as before. So the width accounting is unchanged.

For height: each box uses `Height(heights[i])` where `heights[i]` is the content height. The box visual is `heights[i] + 2`. When we join N boxes, total visual = sum(heights[i] + 2) for i in 0..N-1 = sum(heights) + 2N.

And `sum(heights) = m.height - 4 - 2N` (from the accordion algorithm). Wait no, `availableRows = sidebarHeight - borderOverhead = (m.height - 4) - 2N`. And `sum(heights) = availableRows = m.height - 4 - 2N`.

So total visual = sum(heights) + 2N = (m.height - 4 - 2N) + 2N = m.height - 4. Which matches `sh = m.height - 4` from the current code.

The main panel is `mh = m.height - 4`. So both sidebar and main panel have the same "visual height" of `m.height - 4`. Perfect.

**Acceptance criteria:**
- [ ] Sidebar and main panel have the same height at any terminal size
- [ ] No gaps or overflow between boxes in the sidebar
- [ ] Bottom bar aligns with the outer frame
- [ ] Renders correctly at 80x24 minimum size

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestStartUpdateMsg` | Unit | `gui_test.go` | `update_on_start=true` sends `StartUpdateMsg` from Init |
| `TestNoUpdateOnStart` | Unit | `gui_test.go` | `update_on_start=false` sends panel fetches from Init |
| `TestUpdateCompleteRefreshes` | Unit | `gui_test.go` | After `UpdateCompleteMsg`, all panel fetches are queued |
| `TestDuplicateUpdateIgnored` | Unit | `gui_test.go` | `StartUpdateMsg` while `isUpdating` is no-op |
| `TestParseUpdateSummary` | Unit | `gui_test.go` | All brew update output variants parsed correctly |
| `TestParseUpdateEmpty` | Unit | `gui_test.go` | Empty/no-match input returns empty string |
| `TestAccordionHeights` | Unit | `gui_test.go` | Heights sum to available space, active gets ~40% |
| `TestAccordionMinimums` | Unit | `gui_test.go` | Minimums enforced at small terminal sizes |
| `TestRenderSidebarContent` | Unit | `gui_test.go` | Items rendered correctly with truncation |
| `TestRenderSidebarContentStates` | Unit | `gui_test.go` | Loading, error, empty states rendered properly |
| `TestRenderBox` | Unit | `gui_test.go` | Box has correct border, active/inactive colors |
| `TestRenderBottomBarUpdate` | Unit | `gui_test.go` | Bottom bar shows correct update state |
| `TestRenderBreadcrumb` | Unit | `gui_test.go` | Main panel shows `PanelName › TabName` |
| `TestExistingNavigation` | Unit | `gui_test.go` | All existing navigation tests still pass |
| `TestExistingTabSwitching` | Unit | `gui_test.go` | All existing tab tests still pass |

---

## Definition of Done

- [ ] Auto-update runs on startup when `brew.update_on_start=true` (non-blocking, bottom bar indicator)
- [ ] `R` key with `update_on_start=true` runs update first
- [ ] Bottom bar shows hints (left) + update status (right)  
- [ ] Each sidebar panel renders as an individual bordered box
- [ ] Active panel has accent border and ~40% sidebar height
- [ ] Inactive panels have subtle borders and share remaining height
- [ ] Main panel shows `PanelName › TabName` title line
- [ ] Empty, loading, error states render correctly in sidebar boxes
- [ ] All existing tests pass
- [ ] Renders correctly at 80x24 minimum terminal size
- [ ] No regressions in mutation/panel navigation/tab switching
