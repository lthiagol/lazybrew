# Milestone 17 — Lazygit-Inspired TUI & Auto-Update

> **Status:** ⚠️ Partial (~95% done)  
> **Size estimate:** S remaining (17.3 only)  
> **Depends on:** M19 ✅, M20 ✅, M21 T2 (baseline done)  
> **Enables:** v0.3.0 UX release, `ShowIcons` (backlog B-10)  
> **Parallel track:** F (Polish)  
> **Gate criteria:** Update summary toast shown after `brew update`; no new `program.Send`  
> **Format:** Refined 2026-06-13 ([templates/milestone.md](../templates/milestone.md))

See [planning-challenge-2026-06-13.md](../archive/planning-challenge-2026-06-13.md) — M17 runs after correctness + CI, not after M15/M16 alone.

---

## Goal

Redesign the TUI to feel more like lazygit: per-panel bordered sidebar boxes, accordion height distribution, main-panel breadcrumb, bottom-bar update status, optional non-blocking `brew update` on startup, and search result info in the main panel.

---

## Why Now

Visual polish on top of broken concurrency or stale tab content wastes rework. This milestone runs last so layout changes sit on TaskManager (M19), correct Info/tabs (M20), and teatest coverage (M21).

---

## Reality Check (2026-06-14)

Implemented and verified:

- **17.1/17.2** Auto-update messages, model fields, Init/Update handlers, `R` key behavior.
- **17.4/17.5/17.6/17.9/17.10** Sidebar per-panel boxes, accordion height math, compact sidebar renderer, loading/empty/error states, visual integration.
- **17.7** Two-line bottom bar with update status + 10s ticker.
- **17.8** Main panel breadcrumb (`Panel › Tab`).
- **17.11** Search result info preview in main panel.

Remaining:

- **17.3** `parseUpdateSummary` + toast after `brew update` completes. The code collects `updateOutput` but never parses it.

---

## Challenged Assumptions

| Assumption | Challenge | Decision |
|---|---|---|
| Depends on M15/M16 | M16 coverage targets unmet; M15 incomplete | **Depends on M19–M22** |
| Switch to gocui | Large rewrite | **Keep Bubble Tea + Lip Gloss** (D17-1) |
| Auto-update via raw goroutine | Same anti-pattern as M19 | **Use TaskManager / tea.Cmd** for update (D17-2) |
| M2 small-terminal collapse here | Scope creep | **M20.7 warning only**; full collapse → backlog B-10 / post-M17 |

---

## Out of Scope

- Full M2 sidebar collapse / Ctrl+B toggle — backlog **B-10**
- `ShowIcons` in sidebar titles — backlog **B-10**
- Lazy panel loading — backlog **B-02**
- Rewriting M1–M16 legacy steps

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|---|---|---|---|
| D17-1 | Stay on Bubble Tea + Lip Gloss | gocui rewrite | 7k+ lines; Charm ecosystem |
| D17-2 | Auto-update through TaskManager | `go func` + `program.Send` | M19 concurrency rules |
| D17-3 | Title inside box (bold first line) | Title in border line | Simpler lipgloss; same visual goal |
| D17-4 | Accordion 40/60 split; min 4 active / 2 inactive rows | Equal heights | lazygit-like focus |
| D17-5 | `renderBox` = lipgloss border wrapper only | Custom border strings | Reliable sizing |
| D17-6 | Update timestamp ticker **10s** | 1s tick | Enough precision; less CPU |
| D17-7 | Search info: section rules, not nested boxes | Nested `renderBox` in main | Avoid double borders |
| D17-8 | `update_on_start` default `false` | default true | Backward compatible |

---

## Phases

Execute **in order**. Complete phase gate before next phase.

| Phase | Steps | Theme | Phase gate |
|---|---|---|---|
| **A — Auto-update** | 17.1, 17.2, 17.3, 17.7 | Messages, startup/R refresh, bottom bar | Update flow uses TaskManager; bottom bar states correct |
| **B — Sidebar boxes** | 17.4, 17.5, 17.6, 17.9, 17.10 | Accordion, per-panel borders, alignment | 80×24 renders; no clip |
| **C — Main panel** | 17.8 | Breadcrumb title | Tab bar + content fit |
| **D — Search preview** | 17.11 | Package info on selection | Info loads on j/k; `i` still installs |

---

## Step Index

| Step | Title | Size | Phase | Status | Deliverable |
|---|---|---|---|---|---|
| 17.1 | Messages + model fields | S | A | Done | `StartUpdateMsg`, `UpdateCompleteMsg`, fields |
| 17.2 | Auto-update Init + handlers | M | A | Done | Non-blocking update; R key behavior |
| 17.3 | `parseUpdateSummary` | S | A | **Remaining** | Toast summary strings |
| 17.4 | `renderSidebarContent` | M | B | Done | Compact sidebar list renderer |
| 17.5 | `computeContentHeights` | M | B | Done | Accordion row math |
| 17.6 | Per-panel sidebar boxes | L | B | Done | `box.go`, rewrite `renderSidebar` |
| 17.7 | Bottom bar + update ticker | M | A | Done | Hints left; status right; 10s tick |
| 17.8 | Main panel breadcrumb | S | C | Done | `Panel › Tab` line |
| 17.9 | Sidebar loading/empty/error | S | B | Done | States inside boxes |
| 17.10 | Visual integration / alignment | M | B | Done | No gaps; height match main panel |
| 17.11 | Search info preview | L | D | Done | `SearchInfoLoadedMsg`, render |

---

## Design Decisions (detail)

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

**Size:** S · **Phase:** A · **Status:** Remaining

**What:** Extract a human-readable summary from `brew update` output and show it as a toast when the update completes successfully.

**Files:**

| File | Action |
|---|---|
| `internal/gui/commands.go` | Add `parseUpdateSummary(lines []string) string` |
| `internal/gui/gui.go` | In `UpdateCompleteMsg` handler, call parser and set toast |
| `internal/gui/commands_test.go` (new) | Unit tests for parser |

**Implementation detail:**

1. Add parser function:
   ```go
   func parseUpdateSummary(lines []string) string {
       for _, line := range lines {
           line = strings.TrimSpace(line)
           if line == "" { continue }
           if strings.Contains(line, "Already up-to-date") {
               return "Already up to date"
           }
           if strings.HasPrefix(line, "Updated") {
               return strings.TrimRight(line, ".")
           }
           if strings.HasPrefix(line, "Error:") {
               return ""
           }
       }
       return ""
   }
   ```

2. In `gui.go` `UpdateCompleteMsg` handler, after `m.lastUpdate = time.Now()`:
   ```go
   summary := parseUpdateSummary(m.updateOutput)
   if summary != "" {
       m.toast = modal.NewToast(summary, modal.ToastSuccess)
   }
   ```

**Edge cases:**
- No lines → no toast.
- Multiple "Updated" lines → first one wins.
- `Error:` prefix → return empty; existing error toast already handles `msg.Err != nil`.
- Non-English output → empty string (safe fallback).

**Acceptance criteria:**
- [ ] Parser unit tests pass for all listed cases.
- [ ] Successful update shows a toast matching parser output.
- [ ] Failed update still shows the existing error toast, not the summary.
- [ ] No new `program.Send` introduced.

**Tests:**
- `TestParseUpdateSummary` table-driven unit test in `internal/gui/commands_test.go`.
- `TestUpdateCompleteShowsSummary` model-level test (optional; depends on existing mock setup).

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

### 17.11 — Search Result Info Preview in Main Panel

**What:** When the Search panel is active and a result is selected, show `brew info` for that package in the main panel. No tab switching needed — it auto-loads on selection.

**Files:**
- `internal/gui/gui.go` — Model field `searchResults`, store on `SearchDoneMsg`, info fetch handler
- `internal/gui/messages.go` — new `SearchInfoLoadedMsg`
- `internal/gui/render.go` — update `renderContent()` for `PanelSearch`
- `internal/gui/commands.go` — new `fetchSearchInfo()` command

**Visual target (Search selected in sidebar):**
```
┌─ Search ─────────────────┐ ┌─ Search › stow ──────────────────────────────────────┐
│ ▸ stow    formula   ✓    │ │ ┌─ Package Info ────────────────────────────────────┐ │
│   lazygit formula   ✓    │ │ │ Name:      stow                                   │ │
│   stowage cask           │ │ │ Version:   2.4.1 (bottled)                        │ │
│   stowaway formula       │ │ │ Type:      formula                                │ │
│                           │ │ │ Status:    installed                              │ │
│                           │ │ │ License:   GPL-3.0-or-later                      │ │
│                           │ │ │ Homepage:  https://www.gnu.org/software/stow/    │ │
│                           │ │ ├───────────────────────────────────────────────────┤ │
│                           │ │ │ Description:                                      │ │
│                           │ │ │   Organize software neatly under ~/stow           │ │
│                           │ │ ├───────────────────────────────────────────────────┤ │
│                           │ │ │ Dependencies: perl                                │ │
│                           │ │ ├───────────────────────────────────────────────────┤ │
│                           │ │ │ Caveats:                                          │ │
│                           │ │ │   Emacs users: add (require 'stow) to .emacs     │ │
│                           │ │ └───────────────────────────────────────────────────┘ │
│                           │ │                                                       │
│                           │ │  i: install    j/k: navigate  Enter: re-search       │
│                           │ └───────────────────────────────────────────────────────┘
├───────────────────────────┤
│ j/k: ▲▼  i: install  ?: help  q: quit            │  ⟳ Updated 5m ago │
└───────────────────────────┘
```

**Implementation:**

**1. Store raw search results** (`gui.go`):

Add to Model:
```go
searchResults []brew.SearchResult
```

Modify `SearchDoneMsg` handler in `Update()`:
```go
case SearchDoneMsg:
    p := m.panels[PanelSearch]
    p.loading = false
    if msg.Err != nil {
        p.err = msg.Err
    } else {
        p.items = msg.Items       // formatted strings (keep for sidebar)
        m.searchResults = msg.Raw // NEW: raw data for info fetching
    }
    m.switchPanel(PanelSearch)
    return m, nil
```

**2. Update `SearchDoneMsg`** (`messages.go`):
```go
type SearchDoneMsg struct {
    Items   []string          // formatted display strings
    Raw     []brew.SearchResult // raw results for info
    Err     error
}
```

**3. Update `executeSearch`** (`commands.go`):
```go
items := make([]string, len(results))
for i, r := range results {
    // ... existing formatting ...
}
return SearchDoneMsg{Items: items, Raw: results, Err: nil}
```

**4. Info loading on selection change** (`gui.go`):

Add a case to detect selection changes in the Search panel. The simplest approach: load info whenever `j`/`k` is pressed and the Search panel is active. Or more elegantly: check in a new `searchSelectionChanged()` method.

Simplest approach — in the `j`/`k` handlers, after the existing `p.down()`/`p.up()`:
```go
case "j", "down":
    m.panels[m.activePanel].down()
    if m.activePanel == PanelSearch {
        return m, m.fetchSelectedSearchInfo()
    }
case "k", "up":
    m.panels[m.activePanel].up()
    if m.activePanel == PanelSearch {
        return m, m.fetchSelectedSearchInfo()
    }
```

And also when search results first arrive:
```go
case SearchDoneMsg:
    // ... store results ...
    return m, m.fetchSelectedSearchInfo()  // load info for first result
```

**5. `SearchInfoLoadedMsg`** (`messages.go`):
```go
type SearchInfoLoadedMsg struct {
    Content string
    Err     error
}
```

**6. `fetchSelectedSearchInfo`** (`commands.go`):
```go
func (m Model) fetchSelectedSearchInfo() tea.Cmd {
    if m.activePanel != PanelSearch || m.panels[PanelSearch].selected >= len(m.searchResults) {
        return nil
    }
    r := m.searchResults[m.panels[PanelSearch].selected]
    name := r.Name

    return func() tea.Msg {
        ctx := context.Background()
        output, err := m.client.Runner.Execute(ctx, "info", "--json=v2", name)
        if err != nil {
            return SearchInfoLoadedMsg{Err: err}
        }
        return SearchInfoLoadedMsg{
            Content: string(output),  // raw JSON, formatted in renderContent
        }
    }
}
```

**7. Store info content** (`gui.go`):

Add to Model:
```go
searchInfoContent string  // cached info content for current selection
```

Handler:
```go
case SearchInfoLoadedMsg:
    if msg.Err != nil {
        m.searchInfoContent = "Error: " + msg.Err.Error()
    } else {
        m.searchInfoContent = msg.Content
    }
    return m, nil
```

**8. Render info in main panel** (`render.go`):

Update `renderContent()` — add a `PanelSearch` case that renders info instead of a generic list:
```go
case PanelSearch:
    if m.searchInfoContent == "" {
        return style.SubtleText.Render("No package selected")
    }
    return m.renderSearchInfo(mw, height)
```

**9. `renderSearchInfo`** new method (`render.go`):
```go
func (m Model) renderSearchInfo(width, height int) string {
    // Parse the JSON from searchInfoContent
    // Format into sub-box display
    // For JSON response shape:
    // { "formulae": [{...}], "casks": [{...}] }
    //
    // Use lipgloss boxes to show:
    //   ┌─ Package Info ─────────────┐
    //   │ Name:   stow              │
    //   │ ...                       │
    //   ├───────────────────────────┤
    //   │ Description               │
    //   ├───────────────────────────┤
    //   │ Dependencies              │
    //   └───────────────────────────┘
}
```

Format the info as sub-boxes:
```go
func (m Model) renderSearchInfo(width, height int) string {
    info, err := parsePackageInfo(m.searchInfoContent)
    if err != nil {
        return style.ErrorBadge.Render("Parse error: " + err.Error())
    }

    // ┌─ Package Info ──────────────┐
    var lines []string
    lines = append(lines, fmt.Sprintf("Name:     %s", info.Name))
    lines = append(lines, fmt.Sprintf("Version:  %s%s", info.Version, bottledSuffix(info)))
    lines = append(lines, fmt.Sprintf("Type:     %s", info.Type))
    status := "not installed"
    if info.Installed {
        status = fmt.Sprintf("installed (%s)", info.InstallPath)
    }
    lines = append(lines, fmt.Sprintf("Status:   %s", status))
    if info.License != "" {
        lines = append(lines, fmt.Sprintf("License:  %s", info.License))
    }
    if info.Homepage != "" {
        lines = append(lines, fmt.Sprintf("Homepage: %s", info.Homepage))
    }

    infoBox := renderBox(lipgloss.JoinVertical(lipgloss.Top, lines...),
        width-2, len(lines)+1, true)

    // Description sub-box
    descBox := ""
    if info.Description != "" {
        descBox = renderBox(style.NormalItem.Render(info.Description),
            width-2, 2, false)
    }

    // Dependencies sub-box
    depsBox := ""
    if len(info.Dependencies) > 0 {
        deps := lipgloss.JoinVertical(lipgloss.Top,
            mapToStyled(info.Dependencies, style.NormalItem)...)
        depsBox = renderBox(deps, width-2, len(info.Dependencies)+1, false)
    }

    return lipgloss.JoinVertical(lipgloss.Top, infoBox, descBox, depsBox)
}
```

**Important simplification**: Rather than rendering sub-boxes with `renderBox` (which adds a border), we can use a simple approach:
- Main panel already has an outer border (from 17.8)
- Info content is rendered as styled text inside the main border
- Use horizontal rules (`─`) as visual separators between sections
- Section headers in bold/accent

This is simpler and avoids nested borders:
```
┌─ Search › stow ────────────────────────────────────┐
│ Name:      stow                                      │
│ Version:   2.4.1 (bottled)                           │
│ Type:      formula                                   │
│ Status:    not installed                             │
│ ─────────────────────────────────────────────────── │
│ Description:                                         │
│   Organize software neatly under ~/stow               │
│ ─────────────────────────────────────────────────── │
│ Dependencies:                                        │
│   perl                                               │
│ ─────────────────────────────────────────────────── │
│ Caveats:                                             │
│   Emacs users: add (require 'stow) to .emacs        │
└──────────────────────────────────────────────────────┘
```

**10. `parsePackageInfo`** helper (`commands.go`):
```go
type pkgInfo struct {
    Name         string
    Version      string
    Type         string     // "formula" or "cask"
    Bottled      bool
    Installed    bool
    InstallPath  string
    License      string
    Description  string
    Homepage     string
    Dependencies []string
    Caveats      string
}

func parsePackageInfo(rawJSON string) (*pkgInfo, error) {
    // Parse brew info --json=v2 output
    // Structure: { "formulae": [...], "casks": [...] }
    // At most one array has one element
    var result struct {
        Formulae []brew.Formula `json:"formulae"`
        Casks    []brew.Cask    `json:"casks"`
    }
    if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
        return nil, err
    }

    if len(result.Formulae) > 0 {
        f := result.Formulae[0]
        return &pkgInfo{
            Name:        f.Name,
            Version:     f.Version,
            Type:        "formula",
            Bottled:     f.Bottled,
            Installed:   f.InstalledOnReq || f.InstalledAsDep,
            InstallPath: f.InstallPath,
            License:     f.License,
            Description: f.Description,
            Homepage:    f.Homepage,
            Dependencies: append(f.Dependencies, f.BuildDeps...),
            Caveats:     f.Caveats,
        }, nil
    }
    if len(result.Casks) > 0 {
        c := result.Casks[0]
        return &pkgInfo{
            Name:         c.Name,
            Version:      c.Version,
            Type:         "cask",
            Installed:    false, // installed detection from search result
            Description:  c.Description,
            Homepage:     c.Homepage,
            Dependencies: c.DependsOn,
        }, nil
    }

    return nil, fmt.Errorf("no package info found")
}
```

**Edge cases:**
- No selection → main panel shows "No package selected"
- Fetch error → "Error: failed to load package info"
- Not installed → show "not installed" in accent/warning
- No description → skip description section
- No dependencies → skip dependencies section
- Package name changed between search and info fetch (rare) → handle gracefully

**Caching:**
- `searchInfoContent` caches the last fetched info
- When selection changes, re-fetch
- A simple map cache (`map[string]string`) on the Model could avoid redundant fetches for frequently-viewed packages

**Keybindings update:**
- `i` still installs (unchanged)
- `Enter` could re-run search (future)
- Tab switching is irrelevant — Search has only one tab

**Acceptance criteria:**
- [ ] Selecting a search result in sidebar loads package info in main panel
- [ ] Formula info shows name, version, type, status, license, homepage
- [ ] Cask info shows name, version, type, status, homepage
- [ ] Not installed shows "not installed" clearly
- [ ] Description section shown when available
- [ ] Dependencies section shown when available
- [ ] Caveats section shown when available
- [ ] "No package selected" when search is empty or no selection
- [ ] Error state shown when info fetch fails
- [ ] Info updates when navigating to a different result
- [ ] `i` still installs from the info view
- [ ] JSON parse errors render gracefully (not a crash)

---

## Tests for This Milestone

Consolidated test plan — add/update in **same step** as implementation.

| Test | Tier | Step | Proves |
|---|---|---|---|
| `TestStartUpdateMsg` | unit | 17.2 | Init sends update when configured |
| `TestNoUpdateOnStart` | unit | 17.2 | Default Init fetches panels |
| `TestUpdateCompleteRefreshes` | unit | 17.2 | Refresh after update |
| `TestDuplicateUpdateIgnored` | unit | 17.2 | No concurrent updates |
| `TestParseUpdateSummary` | unit | 17.3 | Summary strings |
| `TestParseUpdateEmpty` | unit | 17.3 | Empty brew output |
| `TestAccordionHeights` | unit | 17.5 | 40/60 split |
| `TestAccordionMinimums` | unit | 17.5 | Small terminal degradation |
| `TestRenderSidebarContent` | unit | 17.4 | Truncation, selection |
| `TestRenderSidebarContentStates` | unit | 17.9 | Loading/empty/error |
| `TestRenderBox` | unit | 17.6 | Border colors |
| `TestRenderBottomBarUpdate` | unit | 17.7 | Status strings |
| `TestRenderBreadcrumb` | unit | 17.8 | Panel › Tab |
| `TestExistingNavigation` | unit | 17.10 | No regressions |
| `TestExistingTabSwitching` | unit | 17.10 | No regressions |
| `TestSearchInfoLoadsOnSelection` | unit | 17.11 | Info fetch on j/k |
| `TestSearchInfoRenders` | unit/e2e | 17.11 | Main panel content |
| `TestParsePackageInfo` | unit | 17.11 | JSON parsing |
| `TestParsePackageInfoCask` | unit | 17.11 | Cask branch |
| `TestParsePackageInfoInvalid` | unit | 17.11 | Error handling |

**Verification commands:**

```bash
make test
go test -race ./internal/gui/...
# After M21: teatest flows for navigation + search info
```

---

## Definition of Done

- [x] Phases A–D complete except 17.3
- [x] Auto-update uses TaskManager (D17-2) — no new `program.Send`
- [ ] 17.3 `parseUpdateSummary` + toast implemented and tested
- [x] All other tests in Test Plan exist and pass
- [x] `go test -race ./...` passes
- [x] [smoke-checklist.md](../smoke-checklist.md) re-run after visual changes
- [x] Renders at 80×24 without clip (17.10)
- [x] [status.md](../status.md) updated

---

## Post-Milestone Gate

Before v0.3.0 tag:

- [ ] M19–M22 still green (no regressions)
- [ ] teatest navigation + search flows pass (M21)
- [ ] Manual lazygit visual check on real terminal

---

## Rollback Plan

If accordion layout fails at common sizes:

1. Ship Phase A (auto-update + bottom bar) without Phase B box rewrite
2. Keep old `renderSidebar` behind config flag `gui.legacy_sidebar` (optional emergency — remove before v1.0)

---

## Version History

| Date | Change |
|---|---|
| 2026-06-11 | Initial milestone |
| 2026-06-13 | Refined to current template; deps M19–M22; phases; TaskManager for update |
