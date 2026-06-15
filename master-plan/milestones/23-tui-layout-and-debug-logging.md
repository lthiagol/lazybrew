# Milestone 23 — TUI Layout Rework & Debug Logging

> **Status:** 🔜 Planned  
> **Size estimate:** M (2–3 days)  
> **Depends on:** —  
> **Enables:** daily-driver use, easier debugging  
> **Parallel track:** —  
> **Gate criteria:** Right panel fills remaining space; bottom bar shows two lines; `-debug` flag writes to `~/.config/lazybrew/debug.log`

Execute **phases A → D in order**. Phase E (verification) last.

| Phase | Steps | Theme |
|---|---|---|
| **A — Core layout** | 23.1, 23.2 | Right panel fill, command log |
| **B — Bottom bar** | 23.3, 23.4 | Two-line layout, styled tags |
| **C — UI polish** | 23.5, 23.6 | Number prefixes, spinner |
| **D — Debug logging** | 23.7 | File logging |
| **E — Verification** | 23.8 | Smoke test |

---

## Goal

The TUI right panel always fills the available space with a consistent border, shows a command log at the bottom, and has a two-line bottom bar with styled key tags. Sidebar titles are numbered for quick reference, tab loading shows an animated spinner. A `-debug` flag writes all brew command executions to `~/.config/lazybrew/debug.log` for troubleshooting.

---

## Why Now

These are user-facing quality-of-life improvements that make the TUI feel polished and debuggable. The command log and debug logging directly help diagnose brew failures, which have been observed in the wild.

---

## Out of Scope

- **Persistent command history across sessions** — log file covers this
- **XDG env var support** — config path stays hardcoded to `~/.config/lazybrew/`
- **Log rotation** — single append file is sufficient for debug use
- **Alternative layout modes** — no config toggles for layout variants

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|---|---|---|---|
| D23-1 | Command log stored as in-memory ring buffer in Model | Persistent file for command log | Log file covers persistence; in-memory is simpler for TUI display |
| D23-2 | `bubbles/spinner` for loading animation | Manual frame counter | Already in dependency tree, handles ticks cleanly |
| D23-3 | Bottom bar: global keys line 1, panel keys line 2 | Single line, or 3 lines | Matches awesome-stow pattern; fits most terminals |
| D23-4 | Debug log to `~/.config/lazybrew/debug.log` | `~/.local/state/` | Consistent with existing config path; user chose this |

---

## Phase A — Core Layout

### 23.1 — Fix Right Panel Sizing

**Size:** S  
**Phase:** A  
**Depends on:** —  
**Blocks:** 23.2

**Context:** The right panel's `ActiveBorder` wraps to fit content. When content is short ("Loading...", "No selection") the border box shrinks instead of filling the allocated space.

**Preconditions:**
- [ ] `make test` passes

**Implementation checklist:**
1. In `renderMainPanel()` (`render.go`), ensure the content string passed to `ActiveBorder.Render()` is padded to `mw × mh` using `lipgloss.NewStyle().Width(mw).Height(mh)` before the border is applied
2. Alternatively, apply `.Width(mw)` and `.Height(mh)` directly to `ActiveBorder` in `style/theme.go`, or pad content in `renderContent()` to fill its allocated height
3. Verify border spans full width/height regardless of content ("Loading...", info text, lists)

**Files:**

| File | Action |
|---|---|
| `internal/gui/render.go` | Modify `renderMainPanel()` to pad content to fill allocated space |

**Acceptance criteria:**
- [ ] Right panel border always fills the space between sidebar and terminal edge
- [ ] No visual difference when switching between panels with varying content lengths

**Tests (same change set as implementation):**
- [ ] `TestRightPanelFillsAllocatedSpace` — mock `renderMainPanel` output width equals expected

**Out of scope for this step:**
- Command log pane (23.2)

---

### 23.2 — Add Command Log Pane

**Size:** M  
**Phase:** A  
**Depends on:** 23.1  
**Blocks:** 23.8

**Context:** Users need visibility into what brew commands are being executed. A command log pane at the bottom of the right panel (~1/5 height) shows recent commands with status.

**Implementation checklist:**
1. **Data structure** — Create `internal/gui/command_log.go`:
   - Ring buffer of `LogEntry{Command string, Timestamp, Status (running/success/error)}`
   - Max 20 entries
   - Thread-safe append (or use channel + update loop)
2. **Wire into Model** — Add `commandLog *CommandLog` to `Model` in `gui.go`
3. **Hook into command execution** — Create a `RecordingRunner` wrapper in `internal/brew/runner.go` that wraps `Runner` and calls a callback on each `Execute`/`ExecuteStream`/`ExecuteJSON` invocation:
   ```go
   type RecordingRunner struct {
       inner    Runner
       OnExec   func(args []string)
   }
   ```
   Or simpler: log commands via the existing `task.Task` lifecycle (add `Command` field to `Task` and log on `TaskStartedMsg`/`TaskCompletedMsg`)
4. **Render** — In `renderMainPanel()` (`render.go`):
   - Calculate split: content gets ~4/5 of `mh`, command log gets ~1/5
   - Render content normally for the content portion
   - Render command log entries for the log portion
   - Join vertically with lipgloss
5. **Style** — Use `style.SubtleText` for log entries, `style.AccentText` for command name, color-coded status icon

**Files:**

| File | Action |
|---|---|
| `internal/gui/command_log.go` | Create — ring buffer data structure |
| `internal/gui/gui.go` | Modify — add `commandLog` field, wire recording |
| `internal/gui/render.go` | Modify — split right panel content/log vertically |
| `internal/brew/runner.go` | Modify — add `RecordingRunner` |

**Acceptance criteria:**
- [ ] After running a brew command, the command appears in the log pane
- [ ] Log pane shows at most 20 entries (oldest evicted)
- [ ] Log pane occupies ~1/5 of right panel height

**Tests:**
- [ ] `TestCommandLogRingBuffer` — append until full, verify eviction
- [ ] `TestCommandLogRendersInPanel` — render output contains log entries

**Out of scope for this step:**
- Debug file logging (23.7)

---

## Phase B — Bottom Bar

### 23.3 — Rework Bottom Bar to Two-Line Layout

**Size:** M  
**Phase:** B  
**Depends on:** —  
**Blocks:** 23.4

**Context:** The bottom bar currently mixes global and panel-specific key hints in a single line, making it crowded and hard to scan.

**Implementation checklist:**
1. In `renderBottomBar()` (`render.go`), split hints into two groups:
   - **Global** (always shown): `Tab`, `S-Tab`, `1-7`, `/`, `?`, `R`, `q`
   - **Panel-specific** (changes with `m.activePanel`): from `panelHints()`
2. Render as two separate lines:
   - Line 1: `lipgloss.JoinHorizontal` of global hints
   - Line 2: `lipgloss.JoinHorizontal` of panel-specific hints
3. Join lines vertically: `lipgloss.JoinVertical(lipgloss.Top, line1, line2)`
4. Apply `.Width(m.w-2)` to the combined result
5. If terminal is too narrow to fit both, show only line 2 (panel-specific)

**Files:**

| File | Action |
|---|---|
| `internal/gui/render.go` | Modify `renderBottomBar()` — two-line layout |
| `internal/gui/keybindings.go` | Modify — possibly split global/panel exports |

**Acceptance criteria:**
- [ ] Global keys on line 1, panel-specific keys on line 2
- [ ] Switching panels updates line 2
- [ ] Narrow terminal falls back to line 2 only

**Tests:**
- [ ] `TestBottomBarTwoLines` — rendered output contains two lines
- [ ] `TestBottomBarNarrowTerminal` — only panel hints shown when terminal < threshold

---

### 23.4 — Styled Key Tags

**Size:** S  
**Phase:** B  
**Depends on:** 23.3

**Context:** Inspired by awesome-stow's `key_span()` pattern, key names should be visually distinct with a tag-like appearance.

**Implementation checklist:**
1. Enhance `HintKey` and `HintDesc` styles in `style/theme.go`:
   - `HintKey`: add inline padding (1 space left/right via `.Padding(0, 1)`), keep accent color + bold
   - `HintDesc`: add space before, keep subtle color
2. Update `renderBottomBar()` to format each hint as:
   ```go
   style.HintKey.Render(h.key) + " " + style.HintDesc.Render(h.desc)
   ```
   (This already exists — verify consistency)
3. Ensure the two-line layout applies the same formatting to both lines

**Files:**

| File | Action |
|---|---|
| `internal/gui/style/theme.go` | Modify — enhance `HintKey` with padding |
| `internal/gui/render.go` | Modify — verify formatting in bottom bar |

**Acceptance criteria:**
- [ ] Key names appear with colored background or padding, visually distinct from descriptions
- [ ] Consistent look between global and panel-specific hint lines

**Tests:**
- [ ] `TestKeyTagStyling` — rendered spans have expected padding/color

---

## Phase C — UI Polish

### 23.5 — Add Number Prefixes to Sidebar Titles

**Size:** S  
**Phase:** C  
**Depends on:** —  

**Context:** Sidebar panels (Status, Formulae, Casks...) are accessible via `1-7` keys but the titles don't show which number maps to which panel.

**Implementation checklist:**
1. In `renderSidebar()` (`render.go`), modify the title assembly:
   ```go
   prefix := strconv.Itoa(int(p.id) + 1)
   title := prefix + " " + p.title
   ```
   (PanelStatus=0 → "1", PanelFormulae=1 → "2", etc.)
2. Keep existing item count / loading indicator suffix
3. Style the number prefix with `style.SubtleText` or `style.AccentText` (subtle for inactive, accent for active)

**Files:**

| File | Action |
|---|---|
| `internal/gui/render.go` | Modify `renderSidebar()` — add number prefix to title |

**Acceptance criteria:**
- [ ] Sidebar shows "1 Status", "2 Formulae", etc.
- [ ] Still shows item count / loading indicator

**Tests:**
- [ ] `TestSidebarNumberPrefixes` — rendered output contains numbered titles

---

### 23.6 — Add Loading Spinner to Tabs

**Size:** M  
**Phase:** C  
**Depends on:** —  

**Context:** Tab content shows static "Loading..." text while fetching. An animated spinner would indicate activity.

**Implementation checklist:**
1. Add `spinner.Model` to the GUI `Model` in `gui.go`:
   ```go
   import "github.com/charmbracelet/bubbles/spinner"
   
   type Model struct {
       // ... existing fields
       spinner spinner.Model
   }
   ```
2. Initialize in `New()`: `spinner.New(spinner.WithStyle(style.SubtleText))`
3. In `Init()`, return `m.spinner.Tick` as a tea.Cmd for continuous animation
4. In `Update()`, handle spinner tick:
   ```go
   case spinner.TickMsg:
       m.spinner, cmd = m.spinner.Update(msg)
       return m, cmd
   ```
5. In `renderContent()` (`render.go`), replace all `"Loading..."` strings with `m.spinner.View() + " Loading..."` or similar

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — add `spinner.Model`, init, update handler |
| `internal/gui/render.go` | Modify — replace `"Loading..."` with spinner |

**Acceptance criteria:**
- [ ] Animated spinner visible while data/tab content loads
- [ ] Spinner stops when content loads

**Tests:**
- [ ] `TestSpinnerRendersWhileLoading` — loading state shows spinner characters

---

## Phase D — Debug Logging

### 23.7 — Debug Logging to File

**Size:** S  
**Phase:** D  
**Depends on:** —  

**Context:** The `-debug` flag already enables `slog` debug output to stderr. We need file logging so users can share logs for troubleshooting.

**Implementation checklist:**
1. In `internal/brew/logger.go`:
   - Add a function `EnableFileLogging(path string)` that creates the directory (`os.MkdirAll`) and adds a file handler to `slog`
   - The file handler should be a `slog.NewTextHandler` writing to a file at `~/.config/lazybrew/debug.log`
2. In `internal/app/app.go` or `cmd/lazybrew/main.go`:
   - When `-debug` is set, call the file logging function after config is loaded
   - Pass the resolved path from config location (or compute it from `os.UserHomeDir() + "/.config/lazybrew/debug.log"`)
3. Ensure existing `Logger().Debug(...)` calls in `runner.go` capture:
   - Command args before execution
   - Duration and success on completion
   - Exit code and stderr on failure
4. Consider adding richer logging: stdout size, full stderr on failure

**Files:**

| File | Action |
|---|---|
| `internal/brew/logger.go` | Modify — add `EnableFileLogging()` |
| `internal/app/app.go` | Modify — call file logging when debug enabled |
| `cmd/lazybrew/main.go` | Modify — pass config path or home dir for log file location |

**Acceptance criteria:**
- [ ] `lazybrew -debug` creates `~/.config/lazybrew/debug.log`
- [ ] All brew command executions are logged with args, duration, status
- [ ] No file created when `-debug` is not set

**Tests:**
- [ ] `TestDebugLogFileCreated` — `-debug` flag creates log file with expected entries
- [ ] `TestDebugLogNotCreatedWithoutFlag` — no file when flag absent

---

## Phase E — Verification

### 23.8 — Manual Smoke Test

**Size:** S  
**Phase:** E  
**Depends on:** 23.1–23.7

**Context:** Verify all changes work together in a real terminal.

**Implementation checklist:**
1. Run `make test` — all existing tests pass
2. Run `go build ./cmd/lazybrew` — compiles clean
3. Launch TUI (`./lazybrew`):
   - [ ] Right panel fills remaining space (no shrinking border)
   - [ ] Command log pane visible at bottom-right
   - [ ] Bottom bar shows two lines (globals + panel hints)
   - [ ] Key tags are visually distinct (colored, padded)
   - [ ] Sidebar shows "1 Status", "2 Formulae"...
   - [ ] Loading spinner animates during data fetch
4. Launch with `-debug`:
   - [ ] `~/.config/lazybrew/debug.log` created
   - [ ] Log file contains brew command entries
5. Switch panels, run commands — verify no regressions

**Acceptance criteria:**
- [ ] All checklist items pass

---

## Test Plan (milestone-level)

| Test | Tier | Step | Proves |
|---|---|---|---|
| `TestRightPanelFillsAllocatedSpace` | unit | 23.1 | Border fills remaining space |
| `TestCommandLogRingBuffer` | unit | 23.2 | Ring buffer eviction logic |
| `TestCommandLogRendersInPanel` | unit | 23.2 | Log visible in render output |
| `TestBottomBarTwoLines` | unit | 23.3 | Two-line layout |
| `TestBottomBarNarrowTerminal` | unit | 23.3 | Graceful degradation |
| `TestKeyTagStyling` | unit | 23.4 | Key tag appearance |
| `TestSidebarNumberPrefixes` | unit | 23.5 | Numbered titles |
| `TestSpinnerRendersWhileLoading` | unit | 23.6 | Spinner visible during load |
| `TestDebugLogFileCreated` | unit | 23.7 | File written with flag |
| `TestDebugLogNotCreatedWithoutFlag` | unit | 23.7 | No file without flag |

**Verification commands:**

```bash
go build ./cmd/lazybrew
make test
go vet ./...
```

---

## Definition of Done

- [ ] All steps 23.1–23.8 complete; acceptance criteria checked
- [ ] Every Test Plan row has a passing test
- [ ] Verification commands pass (including vet)
- [ ] `master-plan/status.md` updated; this file header Status matches
- [ ] No open critical/high findings in this milestone's scope

---

## Post-Milestone Gate

Before starting dependent work:

- [ ] Header gate criteria satisfied
- [ ] Smoke checklist signed off

---

## Version History

| Date | Change |
|---|---|
| 2026-06-14 | Created from [templates/milestone.md](../templates/milestone.md) |
