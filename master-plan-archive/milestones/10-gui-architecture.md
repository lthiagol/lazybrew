# Milestone 10 — GUI Architecture Payoff

> **Status:** ✅ Complete  
> **Depends on:** Milestone 5 (Modals), Milestone 9 (Config)  
> **Enables:** Milestone 11 (Feature Completion), Milestone 12 (Test Infrastructure)

---

## Goal

Pay off the architectural shortcuts taken during initial development. Several GUI features exist as **stubs** — they render but don't function, or they load config but don't use it. This milestone connects those stubs to real behavior.

---

## Why This Comes Before New Features

Adding more features on top of broken architecture compounds debt. The tab system, progress display, and config wiring are all disconnected from their backing logic. Fixing these first makes M11 (features) and M12 (tests) dramatically simpler.

---

## Steps

### 10.1 — Fix Tab Content Switching

**Problem:** The `renderContent` method (gui.go:630-633) returns `panel.renderList()` regardless of which tab is active. The "Deps", "Used By", "Caveats", "Files", "Config", "Doctor", "Versions", "Trust", "Formulae" tabs all show identical list content. The tab bar is cosmetic.

```go
// Current broken implementation:
func (m Model) renderContent(width, height int) string {
    panel := m.panels[m.activePanel]
    return panel.renderList(width, height)  // ignores m.activeTab entirely
}
```

**Fix:** Make `renderContent` switch on `(m.activePanel, m.activeTab)` and render contextually:

| Panel | Tab 1 | Tab 2 | Tab 3 | Tab 4 | Tab 5 |
|---|---|---|---|---|---|
| Status | Dashboard (list) | Config (`brew config`) | Doctor (`brew doctor`) | — | — |
| Formulae | Info (list) | Deps (`brew deps --tree`) | Used By (`brew uses`) | Caveats | Files (`brew list`) |
| Casks | Info (list) | Deps | Caveats | — | — |
| Outdated | Info (list) | Versions (diff view) | — | — | — |
| Taps | Tap Info | Trust | Formulae list | — | — |
| Services | Status (list) | — | — | — | — |
| Search | Info | — | — | — | — |

**Files:** `internal/gui/gui.go` — rewrite `renderContent`

**Acceptance criteria:**
- [ ] Each tab shows appropriate content (not identical lists)
- [ ] "Deps" tab renders `brew deps --tree <name>` output (text, monospace)
- [ ] "Used By" tab calls `FormulaeReader.Uses()` and shows dependents
- [ ] "Config" tab shows `brew config` output
- [ ] "Doctor" tab shows `brew doctor` warnings
- [ ] "Caveats" tab shows formula caveats text
- [ ] "Versions" tab shows old → new version comparison
- [ ] "Files" tab shows installed files
- [ ] "Trust" tab shows trust status for the tap
- [ ] Tab-specific content is fetched on-demand (not preloaded)

---

### 10.2 — Wire Progress Streaming

**Problem:** The `ProgressModal.AppendLine()` method exists but is never called. During mutations (install, uninstall, upgrade), the output channel is drained silently:
```go
for range ch { }  // output discarded
}
```
The user sees a progress modal with a spinner but no streaming output.

**Fix:** Replace the synchronous drain with a goroutine that forwards each line as a Bubble Tea message:

```go
// Pattern:
ch, errCh := client.FormulaeWrite.Install(ctx, name)
go func() {
    for line := range ch {
        p.Send(ProgressLineMsg{Line: line})  // needs tea.Program reference
    }
    err := <-errCh
    p.Send(ProgressCompleteMsg{Err: err})
}()
```

**Challenge:** This requires a reference to `tea.Program` (`p.Send()`), which the Model doesn't have. Options:
- **Option A:** Pass `*tea.Program` to the Model at startup (simple, but couples Model to Program lifecycle)
- **Option B:** Use a channel-based pattern where the goroutine sends to a channel and the Bubble Tea event loop reads from it (pure, but requires a separate goroutine manager)
- **Option C:** Batch lines and return them in the command result (simplest, but no real-time streaming)

**Recommendation:** Option A. Bubble Tea's `tea.Program.Send()` is designed for this pattern. The Model already has a complex lifecycle; one extra field is acceptable.

**Files:**
- `internal/gui/gui.go` — add `program *tea.Program` field, set in `main.go` after `tea.NewProgram`
- `internal/gui/modal/progress.go` — add `AppendLine` integration via messages
- `internal/gui/task.go` — rewrite `doMutation` to use streaming pattern

**Acceptance criteria:**
- [ ] Install output streams to progress modal in real-time
- [ ] Uninstall output streams (cask zap output visible)
- [ ] Upgrade output streams
- [ ] Spinner animates during operation
- [ ] User can cancel (Esc → SIGINT → 5s → SIGKILL)
- [ ] Completed state shows success/error message

---

### 10.3 — Wire Config Consumption

**Problem:** `internal/config/config.go` loads YAML from disk, but `gui.go` never reads any config values. The theme field, sidebar width, mouse toggle, and auto-refresh interval are all stored but ignored.

**Current state:**
```go
// config.go defines the field:
type GUIConfig struct {
    Theme string `yaml:"theme"`           // default "dark"
    SidebarWidth int `yaml:"sidebar_width"` // default 30
    ...
}

// style/theme.go is hardcoded:
var AccentColor = lipgloss.Color("#7C3AED")  // always purple
```

**Fix:**
- Create `internal/gui/style/config.go` that reads from `*config.Config` and builds a `Theme` struct dynamically
- Support two themes: `dark` (Tokyo Night inspired) and `light` (clean light)
- Apply theme at app startup in `app.go`
- Apply `SidebarWidth` in `sidebarWidth()` function
- Apply `Mouse` in `tea.NewProgram` options

**Files:**
- `internal/gui/style/config.go` — theme builder from config
- `internal/gui/gui.go` — `sidebarWidth()` reads cfg, `New()` applies mouse option
- `internal/app/app.go` — pass config to style system

**Acceptance criteria:**
- [ ] `theme: dark` in config → Tokyo Night colors
- [ ] `theme: light` in config → clean light colors
- [ ] `sidebar_width: 25` in config → sidebar is 25% width
- [ ] `mouse: false` in config → mouse disabled
- [ ] Missing config file → defaults work
- [ ] Invalid color in config → fallback to dark theme with error toast

---

### 10.4 — Wire Help Overlay Esc Close

**Problem:** The help overlay documentation reads "Press ? or Esc to close" but the `Esc` key is not handled. Only `?` toggles it.

**Fix:** Add `case "esc":` to the keybinding switch that sets `m.showHelp = false` when help is active.

**Also:** Add missing help pages for Status and Search panels.

**Files:** `internal/gui/gui.go`, `internal/gui/help.go`

**Acceptance criteria:**
- [ ] `Esc` closes help overlay when active
- [ ] `?` toggles help overlay when not active
- [ ] Help includes Status panel (R refresh, Doctor, Cleanup, Vulns, Missing)
- [ ] Help includes Search panel (type query, Enter to search, i to install)

---

### 10.5 — Decompose gui.go (Structural)

**Problem:** `gui.go` is 792 lines handling message routing, keybinding dispatch, data fetching, rendering, and panel metadata. This makes it hard to test, navigate, and extend.

**Fix:** Split into focused files within `internal/gui/`:

| New File | Contents | Lines Moved |
|---|---|---|
| `messages.go` | All message types (`DataLoadedMsg`, `SearchDoneMsg`, `MutationResultMsg`, `RefreshMsg`, `ProgressLineMsg`, `ProgressCompleteMsg`) | ~30 |
| `commands.go` | All `fetchPanelData`, `fetchStatusData`, `executeSearch`, `doMutation`, `serviceAction`, `togglePin` functions | ~200 |
| `keybindings.go` | Keybinding dispatch table, `panelHints()`, `keyHint` type | ~80 |
| `render.go` | `View()`, `renderSidebar()`, `renderMainPanel()`, `renderContent()`, `renderTabBar()`, `renderBottomBar()` | ~150 |
| `gui.go` | `Model` struct, `New()`, `Init()`, `Update()`, navigation methods (`nextPanel`, `prevPanel`, `switchPanel`, `nextTab`, `prevTab`) | ~330 |

**Result:** No file exceeds 350 lines. Each file has a single responsibility.

**Acceptance criteria:**
- [ ] App compiles and runs identically before and after decomposition
- [ ] No circular dependencies between extracted files
- [ ] Each file has a clear `package gui` header comment describing its responsibility

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestTabContentSwitching` | E2E | `internal/gui/gui_test.go` | Each panel/tab combo shows distinct content |
| `TestHelpEscClose` | E2E | `internal/gui/gui_test.go` | Esc closes help overlay |
| `TestProgressStreaming` | Unit | `internal/gui/task_test.go` | Output lines forwarded as messages |
| `TestThemeDark` | Snapshot | `internal/gui/style/config_test.go` | Dark theme colors correct |
| `TestThemeLight` | Snapshot | `internal/gui/style/config_test.go` | Light theme colors correct |
| `TestConfigLoading` | Unit | `internal/config/config_test.go` | YAML parses correctly, defaults used |

---

## Definition of Done

- [ ] Tab content is contextual (not identical lists)
- [ ] Progress modal shows real-time output during mutations
- [ ] `theme: dark|light` in config affects actual colors
- [ ] `sidebar_width` in config affects layout
- [ ] `Esc` closes help overlay
- [ ] Help overlay covers all 7 panels
- [ ] `gui.go` is decomposed into ≤350 lines per file
- [ ] All tests pass
- [ ] No regressions in existing functionality
