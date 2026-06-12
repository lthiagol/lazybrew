# Milestone 14 — Wire Dead Code & Fix Broken Functionality

> **Status:** 🔲 Not Started  
> **Depends on:** Milestone 13 (Critical Bug Fixes)  
> **Enables:** Milestone 15 (Architecture Cleanup)

---

## Goal

Fix features that exist in the code but don't actually work. These are not missing features — the code is there, but it's disconnected or broken.

---

## Steps

### 14.1 — Wire Brewfile Menu Result Handler

**Problem:** `brewfileMenu()` creates a `MenuModal` with 5 Brewfile options, but `handleModalResult()` has no case for `*modal.MenuResult`. The selection falls through to the default case and does nothing.

**File:** `internal/gui/commands.go`

**Fix:** Add `MenuResult` handling to `handleModalResult()`:
```go
case *modal.MenuResult:
    if !r.Cancelled {
        return m, m.executeBrewfileAction(r.SelectedIndex)
    }
```

Implement `executeBrewfileAction(index int) tea.Cmd` that runs the appropriate `brew bundle` command based on the selected index (0=dump, 1=install, 2=cleanup, 3=check, 4=list).

---

### 14.2 — Fix serviceCleanup to Wait for Confirm

**Problem:** `serviceCleanup()` sets `m.activeModal` to a `ConfirmModal` (line 205), but the returned `tea.Cmd` (line 210) immediately runs `Cleanup` without waiting for user confirmation. The confirm modal is shown but its result is irrelevant.

**File:** `internal/gui/commands.go`

**Fix:** Return `nil` command from `serviceCleanup()` and let the modal flow handle it. When the confirm modal completes, `handleModalResult()` should check if it was a services cleanup confirmation and run the cleanup.

Alternative: Use a state machine pattern where the confirm modal's result triggers the cleanup command.

---

### 14.3 — Wire runVulns/runMissing to Show Results

**Problem:** `runVulns()` calls `Doctor()` and `runMissing()` calls `Missing()`, but both discard the results (`_ = out`, `_ = missing`). The progress modal shows "Completed" with no actual information.

**Files:** `internal/gui/commands.go`

**Fix:** 
- For `runVulns()`: Format the doctor warnings and display them in the progress modal or a dedicated tab.
- For `runMissing()`: Format the missing dependencies and display them.

Consider using `TabContentMsg` to populate a "Vulns" or "Missing" tab in the Status panel.

---

### 14.4 — Fix Hard Type Assertions

**Problem:** `gui.go:172` and `gui.go:303` use `updated.(modal.Modal)` without checking if the assertion succeeds. If `updated` is nil or doesn't implement `modal.Modal`, the program panics.

**File:** `internal/gui/gui.go`

**Fix:** Use comma-ok idiom:
```go
if modal, ok := updated.(modal.Modal); ok {
    m.activeModal = modal
} else {
    m.activeModal = nil
}
```

Apply same fix to all type assertions in the codebase (modal_test.go has several).

---

### 14.5 — Fix Context Cancel Discarded

**Problem:** `serviceCleanup()` creates a context with cancel but discards the cancel function (`_ = cancel`). The context can never be cancelled, leading to resource leaks if the cleanup hangs.

**File:** `internal/gui/commands.go`

**Fix:** Store the cancel function in the Model or pass it to the progress modal so Esc can cancel the operation.

---

### 14.6 — Remove Dead Batch Selection Code

**Problem:** `batchState` has 5 methods (`selectAll`, `deselectAll`, `count`, `selectedIndices`, `selectedNames`) that are never called. The space bar toggles selection (gui.go:253) but the selections are never used.

**File:** `internal/gui/task.go`

**Fix:** Either:
- **Option A:** Remove dead code and the `batch` field from Model.
- **Option B:** Wire batch selection to "Upgrade Selected" action (press `u` with selections upgrades only selected packages).

Recommendation: Option B — the feature is useful and partially implemented.

---

### 14.7 — Remove Unused Mutex

**Problem:** `m.mu` in Model is locked in the `DataLoadedMsg` handler (gui.go:146) but never locked when reading panel data (in `View()`, `renderContent()`, etc.). The mutex provides no actual protection.

**File:** `internal/gui/gui.go`

**Fix:** Remove the mutex. Bubble Tea serializes message handling, so concurrent access to the Model is not possible within the Update/View loop. The mutex was added out of caution but is unnecessary.

---

### 14.8 — Fix ModalDoneMsg Never Sent

**Problem:** `ModalDoneMsg` is defined (messages.go:22) and handled in Update (gui.go:78-80) but never constructed or sent anywhere. Dead code.

**File:** `internal/gui/messages.go`, `internal/gui/gui.go`

**Fix:** Remove `ModalDoneMsg` from both files.

---

### 14.9 — Fix Status Panel Doctor Tab Stub

**Problem:** The Doctor tab (render.go:113) shows "Run 'd' from Status panel to check" but there is no 'd' keybinding. The `runVulns` function is bound to 'v'.

**File:** `internal/gui/render.go`

**Fix:** Either:
- Add 'd' keybinding for doctor check.
- Change the hint text to reference 'v' (vulns) or remove the tab.

---

### 14.10 — Fix Formulae "Files" Tab Placeholder

**Problem:** The Files tab (commands.go:318-320) shows a placeholder message instead of actual file list.

**File:** `internal/gui/commands.go`

**Fix:** Implement `brew list <name>` call and display the output.

---

### 14.11 — Fix Cask "Caveats" Tab Always Empty

**Problem:** The Cask Caveats tab (render.go:99) always returns "No caveats" without checking any data.

**File:** `internal/gui/render.go`

**Fix:** Check if the cask has caveats data and display it.

---

## Tests

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestBrewfileMenuFlow` | Unit | `internal/gui/commands_test.go` | Menu selection triggers correct brew bundle command |
| `TestServiceCleanupConfirm` | Unit | `internal/gui/commands_test.go` | Cleanup waits for confirmation |
| `TestRunVulnsShowsResults` | Unit | `internal/gui/commands_test.go` | Vulns results are displayed |
| `TestRunMissingShowsResults` | Unit | `internal/gui/commands_test.go` | Missing deps are displayed |
| `TestTypeAssertionSafety` | Unit | `internal/gui/gui_test.go` | Type assertions don't panic on nil |
| `TestBatchSelectionUpgrade` | Unit | `internal/gui/task_test.go` | Batch selection works with upgrade |

---

## Definition of Done

- [ ] Brewfile menu selection triggers correct action
- [ ] serviceCleanup waits for confirmation
- [ ] runVulns displays vulnerability results
- [ ] runMissing displays missing dependencies
- [ ] All type assertions use comma-ok idiom
- [ ] Context cancel functions are stored and used
- [ ] Dead batch selection code is removed or wired
- [ ] Unused mutex is removed
- [ ] ModalDoneMsg is removed
- [ ] Status panel Doctor tab works
- [ ] Formulae Files tab shows actual files
- [ ] Cask Caveats tab shows actual caveats
- [ ] All new tests pass
- [ ] No regressions in existing tests
