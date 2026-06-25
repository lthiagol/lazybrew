# Milestone 24 — Smoke Test Fixes

> **Status:** ⚠️ In Progress
> **Size estimate:** M (2–3 days)
> **Depends on:** M23 ✅ (command log, spinner, debug logging)
> **Enables:** v0.2.0 release (unblocks M22.4)
> **Parallel track:** —
> **Gate criteria:** All smoke checklist items pass; manual smoke signed off

---

## Goal

Resolve all 13 findings from the M22.4 smoke test so that the TUI is reliable, the output pattern is consistent (operations render in the right panel), and mutation workflows (search → info → install, batch upgrade) work end-to-end without UI breakage.

---

## Why Now

These are all user-facing bugs or missing flows that block the v0.2.0 release. Together they degrade first-time UX: search breaks the layout, mutations show output in a detached pane, selection states don't render, and refreshes/deps appear broken.

---

## Out of Scope

- **Persistent command history** — covered by debug.log (M23)
- **Alternative themes / layouts** — structural design changes beyond fixing the output pane
- **Log rotation** — debug log is append-only
- **Cask upgrade in batch** — batch upgrade should handle both, but cask support depends on brew data; fix formula-only batch first
- **Build/install from source flags** — only basic `brew install/uninstall` flow

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|----|----------|----------------------|----------|
| D24-1 | Operation output renders in a right-panel tab instead of a floating ProgressModal overlay | Keep floating modal | User explicitly expects output "inside the right pane keeping the visual pattern"; overlay breaks the layout feel |
| D24-2 | `/` key always activates PanelSearch (focuses the search input directly) | Keep `/` as modal + panel toggle | Current two-path logic (modal if not on search, input focus if on search) is confusing and creates a broken intermediate state |

---

## Phase A — Output & Operation UX

### 24.1 — Render Operation Output in Right Panel Instead of Floating Modal

**Size:** M
**Phase:** A
**Depends on:** —
**Blocks:** 24.2

**Context:** Currently, task streaming output (install, upgrade, doctor, etc.) renders in a floating `ProgressModal` overlay. The user expects output to appear in the right panel, consistent with the rest of the TUI layout. The floating modal also caused confusion when cancelling (red error) and when multiple operations run.

**Implementation checklist:**
1. Add a dedicated output tab or area in the right panel that can show live task output
2. Replace `ProgressModal` with an inline renderer: when a task is running, the right panel shows the task title, streaming output, and a cancel/done status line
3. Wire `TaskOutputMsg` to append to this inline area instead of (or in addition to) a modal
4. Keep the modal for confirmations only (confirm before install/upgrade/uninstall)
5. Ensure cancel (Esc) works inline without showing raw context.Canceled errors

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — replace ProgressModal references with inline output state |
| `internal/gui/render.go` | Modify — render right panel with output area when task active |
| `internal/gui/modal/progress.go` | Modify — deprecate or keep for confirmation-only use |
| `internal/gui/task/manager.go` | Modify — adjust if needed for inline output |

**Acceptance criteria:**
- [ ] Install/upgrade/doctor output renders in the right panel, not in a floating overlay
- [ ] Output streams live as lines appear
- [ ] Cancel shows a clean "Cancelled" message, not a raw error toast
- [ ] After completion, right panel reverts to normal content

**Tests:**
- [ ] `TestOperationOutputRendersInline` — task output appears in render output

---

### 24.2 — Fix Command Log to Show Brew Commands

**Size:** S
**Phase:** A
**Depends on:** 24.1

**Context:** The command log pane (added in M23.2) shows nothing. Brew commands executed via the TaskManager should appear in the bottom portion of the right panel.

**Implementation checklist:**
1. Verify `CommandLogCallback` is wired to all task types (install, uninstall, upgrade, etc.)
2. Check that `brewCommandString` is correctly generated and passed to `CommandLog.Append`
3. Ensure command log rendering is visible (right panel bottom area ~1/5 height)
4. Add a test that fires a mock task and asserts the log entry appears

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — verify/attach command log callbacks to all write operations |
| `internal/gui/command_log.go` | Modify — fix any bugs in ring buffer or `SetStatus` matching |
| `internal/gui/render.go` | Modify — ensure command log area renders with entries |

**Acceptance criteria:**
- [ ] After any brew command (install, upgrade, doctor, etc.), the command appears in the log pane
- [ ] Log shows "⟳ brew install foo" during execution, "✓ brew install foo" on success, "✗ brew install foo" on failure

**Tests:**
- [ ] `TestCommandLogShowsExecutedCommands` — assert log entry present after task

---

## Phase B — Search & Navigation

### 24.3 — Fix `/` Search Keybind

**Size:** S
**Phase:** B
**Depends on:** —

**Context:** The `/` key has two behaviours: opens a modal input when not on PanelSearch, or focuses the text input when on PanelSearch. This creates a broken intermediate state where the modal appears as a separate broken pane. Remove the modal path: `/` always switches to PanelSearch and focuses the search input.

**Implementation checklist:**
1. Change `/` handler in `gui.go` to always `switchPanel(PanelSearch)` and focus the search input, regardless of current panel
2. Remove the `m.startSearch()` modal path entirely (or keep only as a fallback)
3. Update global key hint from `{"/", "search"}` if needed

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — unify `/` handler to always activate PanelSearch |
| `internal/gui/commands.go` | Modify — clean up `startSearch()` if no longer needed |

**Acceptance criteria:**
- [ ] Pressing `/` from any panel switches to Search panel and focuses the input
- [ ] No intermediate modal or broken pane appears
- [ ] Existing search results populate correctly

**Tests:**
- [ ] `TestSearchKeyActivatesPanel` — `/` switches to PanelSearch

---

### 24.4 — Add Info & Install Keybinds to Search Results

**Size:** S
**Phase:** B
**Depends on:** 24.3

**Context:** From the Search panel, the user can find packages but there's no obvious way to view detailed info or install a result. Add keybindings (`i` for info, `I` for install) that operate on the selected search result.

**Implementation checklist:**
1. Add `i` keybind in Search panel: calls `fetchSelectedSearchInfo` (already exists at `commands.go:306`)
2. Add `I` (shift-i) keybind in Search panel: calls `doMutation` with `mutationInstall` on the selected search result name
3. Update Search panel key hints to show `{"i", "info"}, {"I", "install"}`

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — handle `i`/`I` in Search panel |
| `internal/gui/keybindings.go` | Modify — add Search panel key hints |
| `internal/gui/commands.go` | Modify — wire `fetchSelectedSearchInfo` and `doMutation` for search results |

**Acceptance criteria:**
- [ ] Pressing `i` on a search result shows detailed info in the right panel
- [ ] Pressing `I` on a search result triggers install with confirmation prompt
- [ ] Search panel hints show `i` and `I`

**Tests:**
- [ ] `TestSearchInfoKeybind` — `i` triggers info fetch
- [ ] `TestSearchInstallKeybind` — `I` triggers install mutation

---

## Phase C — Panel Data Reliability

### 24.5 — Fix Deps Tab Loading

**Size:** M
**Phase:** C
**Depends on:** —

**Context:** The Deps tab in the Formulae panel never loads. The fetch calls `client.Formulae.Deps(ctx, name)` which runs `brew deps <name>`. Need to diagnose why it hangs or returns no results, and ensure the UI shows loading state and error handling.

**Implementation checklist:**
1. Run `brew deps <formula>` manually and verify output format
2. Check `Formulae.Deps` in `internal/brew/` — does it parse the output correctly?
3. Check caching in `tabContent` — verify key format matches between fetch and render
4. Ensure error from Deps fetch is surfaced in the UI (toast or inline error)
5. Add loading spinner while deps are being fetched (the tab shows "Loading...", verify spinner from M23.6 works)

**Files:**

| File | Action |
|---|---|
| `internal/brew/` | Investigate — `Deps` method, output parsing |
| `internal/gui/commands.go` | Modify — fix `fetchTabContentCmd` for Deps |
| `internal/gui/gui.go` | Modify — fix caching or message handling |

**Acceptance criteria:**
- [ ] Selecting a formula and switching to Deps tab shows dependency list within 5s
- [ ] Error if Deps fetch fails is shown to user
- [ ] Spinner visible while loading

**Tests:**
- [ ] `TestDepsTabLoadsContent` — mock Deps returns expected content

---

### 24.6 — Fix R Refresh with Visual Feedback

**Size:** S
**Phase:** C
**Depends on:** —

**Context:** Pressing `R` triggers a data refresh but gives no visual indication. The user reported "I saw no refresh." Add a toast or status message on refresh, and ensure panel data actually reloads.

**Implementation checklist:**
1. Verify `RefreshMsg` handler (`gui.go:385-398`) correctly clears cache and re-fetches all panels
2. Add a brief toast "Refreshed" or show a loading state on panels during refresh
3. If `UpdateOnStart` is enabled and `update` is triggered first, ensure the update progress is visible (if 24.1 is done, it shows in right panel inline)

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — add toast on refresh complete |

**Acceptance criteria:**
- [ ] Pressing `R` reloads all panel data
- [ ] A toast "Data refreshed" appears briefly
- [ ] Panels show loading state during refresh

**Tests:**
- [ ] `TestRefreshShowsToast` — refresh triggers toast message

---

### 24.7 — Handle Doctor/Vulns Exit Codes Gracefully

**Size:** S
**Phase:** C
**Depends on:** 24.1

**Context:** `brew doctor` exits with code 1 when it finds warnings (not errors). `brew vulns` (or equivalent) may also exit non-zero. The TUI should display the output regardless of exit code, not fail silently or show an error toast for non-zero exits that still produce useful output.

**Implementation checklist:**
1. Identify where `doctor` and `vulns` commands are executed in the codebase
2. Change the success/failure logic: if the command produced stdout output, show it regardless of exit code
3. Display exit code info in the output if non-zero but output was captured
4. Ensure doctor output shows warnings even when brew exits 1

**Files:**

| File | Action |
|---|---|
| `internal/gui/commands.go` | Modify — doctor/vulns task success handling |
| `internal/gui/gui.go` | Modify — `TaskCompletedMsg` handler for doctor/vulns |

**Acceptance criteria:**
- [ ] `brew doctor` output shows even when brew exits with code 1
- [ ] Non-zero exit is displayed as a note ("Exit code 1"), not as a failure toast
- [ ] Vulns output shows on success or non-zero with output

**Tests:**
- [ ] `TestDoctorShowsWarningsOnExit1` — mock exits 1 with stdout, output is shown

---

## Phase D — Selection & Mutation UX

### 24.8 — Fix Space Selection Indicator in Outdated Panel

**Size:** S
**Phase:** D
**Depends on:** —

**Context:** Pressing Space on an outdated item should toggle a selection indicator (bullet) but the indicator isn't visible. Need to debug why the batch prefix in `panel.go` isn't rendering.

**Implementation checklist:**
1. Check `panel.renderList` batch prefix logic (`panel.go:173-185`) — verify `batch map[int]bool` is correctly passed from render path
2. Check `gui.go:576-579` — verify `m.batch.toggle(p.selected)` uses the correct index
3. Add visual debugging or check that the batch map is non-empty when Space is pressed
4. Ensure prefix characters aren't being stripped or overwritten by styling

**Files:**

| File | Action |
|---|---|
| `internal/gui/panel.go` | Modify — fix batch prefix rendering |
| `internal/gui/render.go` | Modify — ensure batch map is passed correctly for Outdated panel list |
| `internal/gui/gui.go` | Modify — fix Space handler if index incorrect |

**Acceptance criteria:**
- [ ] Pressing Space on an outdated item shows a bullet (`●`) prefix
- [ ] Pressing Space again removes the bullet
- [ ] Moving cursor preserves selection state

**Tests:**
- [ ] `TestBatchSelectionShowsIndicator` — list render includes bullet for selected index

---

### 24.9 — Fix Multi-Select Batch Upgrade

**Size:** M
**Phase:** D
**Depends on:** 24.8

**Context:** Multi-select followed by `u` should upgrade all selected items sequentially. Currently `u` only upgrades the highlighted item. Need to ensure `batch.selected` is populated correctly and `batchUpgrade` uses it.

**Implementation checklist:**
1. After 24.8 fixes Space indicator, verify `m.batch.selected` has entries after Space toggles
2. Check `gui.go:564-569` — when `u` is pressed, verify the condition checks `len(m.batch.selected) > 0`
3. Check `batchUpgrade()` (`commands.go:395-453`) — verify it iterates selected indices and creates tasks for each
4. Ensure batch state is cleared after upgrade starts
5. Handle edge case: selecting items, then navigating away and back — should selection persist? (Keep current behaviour: selection persists until action or explicit clear)

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — fix `u` handler for batch |
| `internal/gui/commands.go` | Modify — fix `batchUpgrade` task creation |

**Acceptance criteria:**
- [ ] Selecting multiple outdated items with Space, then pressing `u`, upgrades all selected items sequentially
- [ ] Upgraded items show progress in right panel per 24.1
- [ ] Selection cleared after upgrade starts
- [ ] `a` toggles all selection

**Tests:**
- [ ] `TestBatchUpgradeUpgradesAllSelected` — batch with 2+ items creates tasks for each

---

### 24.10 — Add Confirmation Prompts for Mutations

**Size:** S
**Phase:** D
**Depends on:** 24.1

**Context:** Install (`i`), uninstall (`x`), upgrade (`u`) should show a confirmation modal before executing. Currently install runs immediately without confirmation. Reuse the existing confirm modal pattern.

**Implementation checklist:**
1. In `doMutation` (`commands.go:321-393`), before creating the task, return a `tea.Cmd` that sets a confirm modal: `"Install <name>?"` / `"Upgrade <name>?"` / `"Uninstall <name> and its dependencies?"`
2. On confirm (Enter), proceed with task creation and execution
3. On cancel (Esc), do nothing and return to previous state
4. Ensure uninstall shows dependency warning if applicable

**Files:**

| File | Action |
|---|---|
| `internal/gui/commands.go` | Modify — add confirmation before mutation tasks |
| `internal/gui/gui.go` | Modify — handle confirm modal result |

**Acceptance criteria:**
- [ ] `i` on a package shows "Install <name>?" confirm modal
- [ ] `x` shows "Uninstall <name>?" confirm modal
- [ ] `u` on single or batch shows confirm modal
- [ ] Esc cancels, Enter proceeds

**Tests:**
- [ ] `TestMutationShowsConfirmModal` — mutation sets confirm modal state
- [ ] `TestMutationConfirmProceeds` — confirm triggers task

---

## Phase E — Error Handling & Loading States

### 24.11 — Graceful Cancel (No Red Error Toast)

**Size:** S
**Phase:** E
**Depends on:** 24.1

**Context:** When the user cancels a running operation (Esc), the UI shows a raw red error message from `context.Canceled`. Cancel should show a clean neutral message: "Cancelled" or "Operation cancelled by user."

**Implementation checklist:**
1. In `TaskCompletedMsg` handler (`gui.go:236-283`), check if `msg.Err` is `context.Canceled`
2. If cancelled, skip the error toast; show a neutral "Cancelled" status in the output area
3. Update `ProgressModal` (or inline output) status line to show "Cancelled" gracefully

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — handle `context.Canceled` without error toast |
| `internal/gui/modal/progress.go` | Modify — if kept, show "Cancelled" not error |

**Acceptance criteria:**
- [ ] Esc during operation shows "Cancelled" — no red error toast
- [ ] UI recovers cleanly after cancel (right panel shows normal content)

**Tests:**
- [ ] `TestCancelShowsCancelledNotError` — cancelled task shows neutral message

---

### 24.12 — Add Loading Spinner to Panels on Data Load

**Size:** M
**Phase:** E
**Depends on:** —

**Context:** When the app starts or refreshes, panels show static "Loading..." text with no activity indicator. Add the spinner from M23.6 to panel sidebar items when they're in loading state.

**Implementation checklist:**
1. Check each panel's loading state — does it set a `loading bool` flag during data fetch?
2. Where panels render their item list, if `loading` is true, show the spinner instead of or alongside the list
3. Ensure spinner ticks correctly through the existing `spinner.TickMsg` handler (added in M23.6)
4. Verify panels show spinner during initial load and during `R` refresh

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Modify — ensure panel loading flags are correct |
| `internal/gui/render.go` | Modify — add spinner to panel sidebar when loading |
| `internal/gui/panel.go` | Modify — add loading state rendering |

**Acceptance criteria:**
- [ ] All panels show an animated spinner in the sidebar while data is loading
- [ ] Spinner stops when data loads
- [ ] Refresh (`R`) shows spinner on panels during reload

**Tests:**
- [ ] `TestPanelShowsSpinnerWhileLoading` — loading state renders spinner character

---

## Phase F — Verification

### 24.13 — Manual Smoke Test

**Size:** S
**Phase:** F
**Depends on:** 24.1–24.12

**Context:** Verify all fixes work together in a real terminal.

**Implementation checklist:**
1. Run `make test` — all tests pass
2. Run `go vet ./...` — clean
3. Run `make lint` — clean
4. Launch TUI:
   - [ ] `/` switches to Search panel, focuses input
   - [ ] Search → Enter → results in sidebar; `i` shows info; `I` shows install confirm
   - [ ] Install with confirm: output in right panel, command in log pane
   - [ ] Cancel shows "Cancelled", UI recovers
   - [ ] Outdated: Space toggles bullet, `u` batch upgrades all selected
   - [ ] Formulae → Deps tab loads dependencies
   - [ ] `R` refreshes with toast feedback
   - [ ] Doctor/vulns show output even on non-zero exit
   - [ ] Panels show spinner on load/refresh
   - [ ] Resize to 79×24: warning shown
   - [ ] `q` quits cleanly

**Acceptance criteria:**
- [ ] All checklist items pass

---

## Test Plan (milestone-level)

| Test | Tier | Step | Proves |
|---|---|---|---|
| `TestOperationOutputRendersInline` | unit | 24.1 | Output in right panel, not overlay |
| `TestCommandLogShowsExecutedCommands` | unit | 24.2 | Log entry after task executed |
| `TestSearchKeyActivatesPanel` | unit | 24.3 | `/` switches to PanelSearch |
| `TestSearchInfoKeybind` | unit | 24.4 | `i` triggers info |
| `TestSearchInstallKeybind` | unit | 24.4 | `I` triggers install |
| `TestDepsTabLoadsContent` | unit | 24.5 | Deps tab shows dependency list |
| `TestRefreshShowsToast` | unit | 24.6 | Toast on refresh |
| `TestDoctorShowsWarningsOnExit1` | unit | 24.7 | Doctor output shown on exit 1 |
| `TestBatchSelectionShowsIndicator` | unit | 24.8 | Bullet prefix on Space |
| `TestBatchUpgradeUpgradesAllSelected` | unit | 24.9 | Multi-select upgrade creates N tasks |
| `TestMutationShowsConfirmModal` | unit | 24.10 | Confirm modal before install |
| `TestMutationConfirmProceeds` | unit | 24.10 | Confirm proceeds |
| `TestCancelShowsCancelledNotError` | unit | 24.11 | Cancelled not error toast |
| `TestPanelShowsSpinnerWhileLoading` | unit | 24.12 | Spinner during data load |

**Verification commands:**

```bash
go build ./cmd/lazybrew
make test
go vet ./...
make lint
```

---

## Definition of Done

- [ ] All steps 24.1–24.13 complete; acceptance criteria checked
- [ ] Every Test Plan row has a passing test
- [ ] Verification commands pass (build, test, vet, lint)
- [ ] `master-plan/status.md` updated; this file header Status matches
- [ ] No open critical/high findings in this milestone's scope
- [ ] Smoke checklist signed off (Tester row filled, Pass)

---

## Post-Milestone Gate

- [ ] Header gate criteria satisfied
- [ ] Release checklist M22.4 signed off
- [ ] Tag and release proceed

---

## Version History

| Date | Change |
|---|---|
| 2026-06-15 | Created from [templates/milestone.md](../templates/milestone.md) based on smoke test findings |
