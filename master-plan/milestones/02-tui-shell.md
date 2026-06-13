# Milestone 2 — TUI Shell & Layout

> **Status:** ⚠️ Partial  
> **Remaining:** Small terminal — warning in M20.7; full lazygit collapse deferred to M17/backlog B-10  
> **Depends on:** Milestone 1 (Foundation)  
> **Enables:** Milestone 4 (Read-Only Panels), Milestone 5 (Modals & Search)

---

## Active Work Routing

> **Format:** Legacy ([milestone-legacy-index.md](../milestone-legacy-index.md)). Do not execute open items from steps below.

| Open item | Execute in |
|---|---|
| Minimum terminal warning | [M20.7](20-functional-completeness-and-ux.md) |
| Sidebar collapse / Ctrl+B / accordion at small width | [M17](17-lazygit-tui-and-auto-update.md) / backlog B-10 |

---

## Goal

Build the core TUI application with the lazygit-style layout: a vertical sidebar of navigable panels on the left, a tabbed content area on the right, and a keybinding hint bar at the bottom. All panels display **hardcoded mock data** — no brew integration yet. The focus is entirely on getting the layout, navigation, and visual foundation right.

After this milestone, you can run lazybrew and navigate between panels, switch tabs, scroll lists — it just shows fake data.

---

## Why This Milestone Matters

The layout and navigation system is the skeleton that everything hangs on. Getting it wrong means painful refactors later. By using mock data, we can iterate on the UI feel independently of the brew data layer. This is also where we establish the Bubble Tea model/update/view architecture that all future features plug into.

---

## Steps

### 2.1 — Bubble Tea App Skeleton

**What:** Create the root Bubble Tea model with the main event loop.

**File:** `internal/gui/gui.go`

**Implementation:**
```go
// Model is the root Bubble Tea model
type Model struct {
    width       int
    height      int
    activePanel PanelID
    panels      []Panel
    mainTabs    []Tab
    activeTab   int
    ready       bool
}

type PanelID int
const (
    PanelStatus PanelID = iota
    PanelFormulae
    PanelCasks
    PanelOutdated
    PanelTaps
    PanelServices
    PanelSearch
)
```

**Key decisions:**
- The root `Model` owns all state and delegates to sub-models (panels)
- Each panel is a separate model implementing a `Panel` interface
- The root model handles global keybindings; panels handle local ones
- Use Lip Gloss for all styling from day one (no raw ANSI)

**Input routing priority (keyboard conflict resolution):**
When multiple layers could handle a keypress, the priority is:
1. **Modal** — if a modal is active, ALL input goes to the modal. Background panels receive nothing.
2. **Text input** — if a text input field is focused (e.g., search modal), ALL keystrokes go to the input (including `j`, `k`, `/`, etc.)
3. **Panel-local** — the active panel handles its keybindings (`j/k` for scroll, `i` for install, etc.)
4. **Global** — the root model handles global keybindings (`Tab`, `1-7`, `/`, `?`, `q`)

> This means `j/k` in a panel scrolls the list, but `j/k` in a text input modal types characters. No ambiguity.

**Sidebar behavior on small terminals:**
- At terminal width < 100 chars: sidebar collapses to icons only (no text labels)
- At terminal width < 60 chars: sidebar hides entirely; user navigates via `1-7` number keys or a hamburger menu (`=` key)
- The sidebar can be toggled visible/hidden at any size with `Ctrl+B`

**Acceptance criteria:**
- [ ] `bubbletea.NewProgram(gui.New())` launches and renders
- [ ] Terminal resize is handled (`tea.WindowSizeMsg`)
- [ ] `q` quits the app cleanly
- [ ] App fills the terminal

---

### 2.2 — Panel Interface & Sub-Models

**What:** Define the `Panel` interface that all sidebar panels implement, and create stub implementations.

**File:** `internal/gui/panel.go`

**Interface:**
```go
type Panel interface {
    // Identity
    ID() PanelID
    Title() string
    Icon() string          // e.g., "📦", "🖥", "⏫"
    
    // Bubble Tea lifecycle
    Init() tea.Cmd
    Update(msg tea.Msg) (Panel, tea.Cmd)
    View() string
    
    // List behavior
    ItemCount() int
    SelectedIndex() int
    SetActive(active bool)
    
    // Dimensions
    SetSize(width, height int)
}
```

**Create stub panels** (each in its own file under `internal/gui/controllers/`):
- `status_panel.go` — renders a fake dashboard
- `formulae_panel.go` — renders a hardcoded list of formulae
- `casks_panel.go` — renders a hardcoded list of casks
- `outdated_panel.go` — renders a hardcoded list of outdated packages
- `taps_panel.go` — renders a hardcoded list of taps
- `services_panel.go` — renders a hardcoded list of services
- `search_panel.go` — renders "press / to search" placeholder

Each panel uses `bubbles/list` or a custom list component for scrolling behavior.

**Acceptance criteria:**
- [ ] All 7 panels implement the `Panel` interface
- [ ] Each panel renders mock data (3–5 items each)
- [ ] Panels handle `j/k` navigation internally
- [ ] Selected item is visually highlighted
- [ ] Active vs inactive panel has distinct border styling

---

### 2.3 — Layout Engine

**What:** Implement the two-column layout with dynamic sizing.

**File:** `internal/gui/layout.go`

**Layout specification:**
```
┌──────────────────────────────────────────────────────┐
│  ┌───sidebar───┐ ┌──────────main──────────────────┐  │
│  │  Status     │ │  [Tab1] [Tab2] [Tab3]          │  │
│  ├─────────────│ │────────────────────────────────│  │
│  │▸ Formulae   │ │                                │  │
│  ├─────────────│ │   Content area                 │  │
│  │  Casks      │ │                                │  │
│  ├─────────────│ │                                │  │
│  │  Outdated   │ │                                │  │
│  ├─────────────│ │                                │  │
│  │  Taps       │ │                                │  │
│  ├─────────────│ │                                │  │
│  │  Services   │ │                                │  │
│  ├─────────────│ │                                │  │
│  │  Search     │ │                                │  │
│  └─────────────┘ └────────────────────────────────┘  │
│  ──────────────────────────────────────────────────── │
│   x: uninstall  u: upgrade  /: search  ?: help       │
└──────────────────────────────────────────────────────┘
```

**Sizing rules:**
- Sidebar width: 30% of terminal width, min 20, max 40 chars
- Main panel: remaining width
- Sidebar panels: height divided evenly among panels, with the active panel getting extra space (like lazygit)
- Bottom bar: 1–2 lines fixed
- Border characters: rounded box-drawing (Lip Gloss)

**Implementation:**
- Use Lip Gloss `lipgloss.JoinHorizontal`, `lipgloss.JoinVertical` for composition
- Define a `Layout` struct that computes coordinates from terminal dimensions
- The active sidebar panel gets a highlighted border (accent color)
- Inactive panels get a subtle border

**Acceptance criteria:**
- [ ] Two-column layout renders correctly at different terminal sizes
- [ ] Sidebar panels stack vertically with dynamic height
- [ ] Active panel has highlighted border
- [ ] Responsive to terminal resize
- [ ] Minimum terminal size check (e.g., 80x24) with error message

---

### 2.4 — Panel Navigation (Sidebar)

**What:** Implement sidebar panel switching — the core navigation mechanic.

**File:** Update `internal/gui/gui.go` (root model's `Update`)

**Keybindings:**
| Key | Action |
|---|---|
| `Tab` | Next panel (wraps around) |
| `Shift+Tab` | Previous panel (wraps around) |
| `1`–`7` | Jump to panel by number |

**Behavior:**
- Switching panels changes the active border highlight
- The main panel content updates to show the context for the newly active panel
- The bottom bar hints update to show keybindings for the new context
- Previous panel's selected item is preserved (not reset)

**Acceptance criteria:**
- [ ] Tab cycles through all 7 panels
- [ ] Shift+Tab cycles backward
- [ ] Number keys jump directly
- [ ] Active panel has distinct visual styling
- [ ] Main panel content changes when panel switches
- [ ] Bottom bar hints update per panel

---

### 2.5 — Main Panel with Tabs

**What:** Implement the right-side main panel with tab switching.

**File:** `internal/gui/main_panel.go`

**Tab structure (per active sidebar panel):**

| Sidebar Panel | Tab 1 | Tab 2 | Tab 3 | Tab 4 | Tab 5 |
|---|---|---|---|---|---|
| Status | Config | Doctor | — | — | — |
| Formulae | Info | Deps | Used By | Caveats | Files |
| Casks | Info | Deps | Caveats | — | — |
| Outdated | Info | Versions | — | — | — |
| Taps | Tap Info | Trust | Formulae | — | — |
| Services | Status | — | — | — | — |
| Search | Info | — | — | — | — |

> **Note:** Formulae has 5 tabs — the "Used By" tab shows `brew uses --installed <name>` (reverse dependencies). This was added during the 6.0.0 review.

**Keybindings:**
| Key | Action |
|---|---|
| `[` | Previous tab |
| `]` | Next tab |

**Implementation:**
- Tabs rendered as a styled header: `[Info]  Deps   Caveats   Files`
- Active tab is highlighted/underlined
- Tab content area is a scrollable viewport (`bubbles/viewport`)
- Content is mock text for now (e.g., "Formula info will appear here")

**Acceptance criteria:**
- [ ] Tab header renders with active tab highlighted
- [ ] `[` / `]` switch tabs
- [ ] Tabs change when sidebar panel changes
- [ ] Tab count matches the panel context
- [ ] Content area scrolls for long content (PgUp/PgDn or Ctrl+u/Ctrl+d)

---

### 2.6 — Bottom Bar (Keybinding Hints)

**What:** Render context-sensitive keybinding hints at the bottom of the screen.

**File:** `internal/gui/bottom_bar.go`

**Design:**
```
 x: uninstall  u: upgrade  U: upgrade all  /: search  ?: help  q: quit
```

**Implementation:**
- Each panel provides a `KeyHints() []KeyHint` method returning its context-specific hints
- Global hints (`/: search`, `?: help`, `q: quit`) are always appended
- Hints are styled: key in bold/accent, description in subtle text
- Truncated gracefully if terminal is too narrow

**Acceptance criteria:**
- [ ] Bottom bar renders with styled key hints
- [ ] Hints change when active panel changes
- [ ] Global hints always visible
- [ ] Graceful truncation on narrow terminals

---

### 2.7 — Styling Foundation

**What:** Define the color palette, theme, and reusable Lip Gloss styles.

**File:** `internal/gui/style/theme.go`

**Design tokens:**
```go
var (
    // Colors
    AccentColor     = lipgloss.Color("#7C3AED")  // purple
    SecondaryColor  = lipgloss.Color("#06B6D4")  // cyan
    SuccessColor    = lipgloss.Color("#10B981")  // green
    WarningColor    = lipgloss.Color("#F59E0B")  // amber
    ErrorColor      = lipgloss.Color("#EF4444")  // red
    SubtleColor     = lipgloss.Color("#6B7280")  // gray
    TextColor       = lipgloss.Color("#E5E7EB")  // light gray
    BgColor         = lipgloss.Color("#1F2937")  // dark gray
    
    // Component styles
    ActiveBorder    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(AccentColor)
    InactiveBorder  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(SubtleColor)
    SelectedItem    = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
    NormalItem      = lipgloss.NewStyle().Foreground(TextColor)
    TabActive       = lipgloss.NewStyle().Foreground(AccentColor).Bold(true).Underline(true)
    TabInactive     = lipgloss.NewStyle().Foreground(SubtleColor)
    HintKey         = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
    HintDesc        = lipgloss.NewStyle().Foreground(SubtleColor)
    PanelTitle      = lipgloss.NewStyle().Foreground(TextColor).Bold(true)
    
    // Status badges
    OutdatedBadge   = lipgloss.NewStyle().Foreground(WarningColor)
    PinnedBadge     = lipgloss.NewStyle().Foreground(SecondaryColor)
    InstalledBadge  = lipgloss.NewStyle().Foreground(SuccessColor)
    ErrorBadge      = lipgloss.NewStyle().Foreground(ErrorColor)
)
```

**Acceptance criteria:**
- [ ] All styles defined as exported Lip Gloss styles
- [ ] Consistent visual language across all panels
- [ ] Works on 256-color and truecolor terminals
- [ ] Degrades gracefully on terminals with fewer colors

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestPanelInterface` | Unit | `internal/gui/panel_test.go` | All panels implement the Panel interface |
| `TestLayoutComputation` | Unit | `internal/gui/layout_test.go` | Correct sizing at various terminal dimensions |
| `TestLayoutMinSize` | Unit | `internal/gui/layout_test.go` | Error shown when terminal too small |
| `TestPanelNavigation` | E2E (teatest) | `internal/gui/gui_test.go` | Tab/Shift+Tab/number keys change active panel |
| `TestTabSwitching` | E2E (teatest) | `internal/gui/gui_test.go` | `[`/`]` switch tabs, tabs change per panel |
| `TestBottomBarHints` | Unit | `internal/gui/bottom_bar_test.go` | Correct hints per panel context |
| `TestQuitKey` | E2E (teatest) | `internal/gui/gui_test.go` | `q` quits the program |
| `TestResize` | E2E (teatest) | `internal/gui/gui_test.go` | Layout adapts to WindowSizeMsg |
| `TestStyleContrast` | Unit | `internal/gui/style/theme_test.go` | Verify styles are non-zero (catch missing definitions) |

### teatest Usage Pattern

```go
func TestPanelNavigation(t *testing.T) {
    m := gui.New()
    tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 40))
    
    // Should start on Status panel
    out := tm.FinalOutput(t)
    // assert Status panel is active
    
    // Press Tab → should move to Formulae
    tm.Send(tea.KeyMsg{Type: tea.KeyTab})
    // assert Formulae panel is active
    
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
    tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
```

---

## Definition of Done

- [ ] App launches and renders the full layout with mock data
- [ ] All 7 sidebar panels visible and navigable
- [ ] Tab/Shift+Tab/number keys switch panels
- [ ] `[`/`]` switch main panel tabs
- [ ] j/k scroll within panel lists
- [ ] Bottom bar shows context-sensitive hints
- [ ] Styling looks polished (not raw terminal text)
- [ ] All tests pass
- [ ] Terminal resize works
- [ ] `q` quits
