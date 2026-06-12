# Milestone 6 — Package Mutations (Install / Uninstall / Upgrade)

> **Status:** ✅ Done  
> **Depends on:** Milestone 4 (Read-Only Panels), Milestone 5 (Modals & Search)  
> **Enables:** Milestone 7 (Taps & Trust)

---

## Goal

Implement the core package lifecycle actions: install, uninstall, and upgrade — for both formulae and casks. This includes the async task manager for long-running operations, progress streaming, post-mutation refresh, and all the confirmation flows. After this milestone, lazybrew is a fully functional package manager UI for day-to-day use.

---

## Steps

### 6.1 — Task Manager

**What:** A central manager for running brew mutations sequentially with progress tracking.

**File:** `internal/gui/task/manager.go`

**Design:**
```go
type TaskManager struct {
    current    *Task
    queue      []*Task
    mu         sync.Mutex
}

type Task struct {
    ID          string
    Title       string            // "Installing ripgrep"
    Command     string            // "brew install ripgrep"
    Status      TaskStatus        // pending, running, success, failed, cancelled
    Output      []string          // captured stdout/stderr lines
    StartedAt   time.Time
    CompletedAt time.Time
    Error       error
    Cancel      context.CancelFunc
}

type TaskStatus int
const (
    TaskPending TaskStatus = iota
    TaskRunning
    TaskSuccess
    TaskFailed
    TaskCancelled
)
```

**Behavior:**
- Only one **write** task runs at a time (brew uses locks, so concurrent mutations fail anyway)
- **Read operations are NOT blocked** — the task manager only queues write operations. Panel reads (List, Get, Outdated) continue to work via the cache and direct brew calls, even while a write task is running.
- Tasks queue if submitted while another is running
- Each task streams output to the progress modal
- Tasks can be cancelled (sends SIGINT via context, then SIGKILL after 5s timeout)
- After task completion, trigger cache invalidation + panel refresh
- **Homebrew 6.0.0:** All brew subprocesses must run non-interactively. The runner (M3 §3.10) must set `HOMEBREW_NO_ASK=1` or pipe a non-TTY stdin to suppress the new default ask-mode confirmation prompts.
- **Automatic retry:** If a task fails with a transient error (network timeout, "resource temporarily unavailable"), retry once after 2s. Non-transient errors (formula not found, permission denied) are not retried. Use the typed errors from M1 §1.3 to distinguish.

**Bubble Tea integration:**
- TaskManager produces messages: `TaskStartedMsg`, `TaskOutputMsg`, `TaskCompletedMsg`
- Root model listens for these messages and updates the progress modal
- The task manager holds a `context.Context` + `CancelFunc` for the active task, enabling cancel via `Esc` in the progress modal

**Acceptance criteria:**
- [ ] Sequential task execution
- [ ] Queuing works (submit while one is running)
- [ ] Task cancellation with SIGINT
- [ ] Output streaming via channels → Bubble Tea messages
- [ ] Post-completion cache invalidation
- [ ] Thread-safe

---

### 6.2 — Install Flow

**What:** Install a formula or cask from search results.

**Flow:**
```
Search panel → user selects a package → presses `i`
  → ConfirmModal: "Install neovim-qt?"
  → User confirms
  → TaskManager: brew install neovim-qt
  → ProgressModal: streaming output
  → On success: toast "✓ neovim-qt installed", refresh Formulae panel
  → On error: error display with stderr
```

**Keybinding:** `i` in Search panel (on an uninstalled result)

**Implementation:**
1. Check if already installed → toast warning "Already installed"
2. Show confirmation modal with package name + description
3. On confirm, submit task to TaskManager
4. TaskManager calls `client.Formulae.Install()` or `client.Casks.Install()`
5. Stream output to progress modal
6. On success: invalidate cache, refresh panels, show toast
7. On error: show error in progress modal, keep it open for review

**Acceptance criteria:**
- [ ] Install from search results works for formulae
- [ ] Install from search results works for casks
- [ ] Already-installed check prevents double install
- [ ] Confirmation modal shown
- [ ] Progress streams in real time
- [ ] Success toast + panel refresh
- [ ] Error display with full stderr
- [ ] Cache invalidated on success

---

### 6.3 — Uninstall Flow

**What:** Uninstall a formula or cask from the Formulae/Casks panel.

**Flow:**
```
Formulae panel → user selects a formula → presses `x` or `d`
  → ConfirmModal: "Uninstall ripgrep 14.1.2?"
  → Optional warning: "3 packages depend on this" (if has dependents)
  → User confirms
  → TaskManager: brew uninstall ripgrep
  → On success: remove from list, toast "✓ ripgrep uninstalled"
```

**Keybinding:** `x` or `d` in Formulae/Casks panels

**Extra logic for casks:**
- `x` = normal uninstall (`brew uninstall --cask`)
- `X` = zap uninstall (`brew uninstall --zap --cask`) with stronger warning

**Dependency warning:**
- Before confirming, check `brew uses --installed <name>`
- If dependents exist, show them in the confirmation: "The following packages depend on ripgrep: ... Continue?"

**Acceptance criteria:**
- [ ] Uninstall from Formulae panel
- [ ] Uninstall from Casks panel (normal + zap)
- [ ] Confirmation modal with package name + version
- [ ] Dependency warning when package has dependents
- [ ] List updates after successful uninstall
- [ ] Selection moves to next item (or previous if last)
- [ ] Cache invalidated

---

### 6.4 — Upgrade Flow

**What:** Upgrade individual packages or all outdated packages.

**Keybindings:**

| Key | Context | Action |
|---|---|---|
| `u` | Formulae/Casks/Outdated panel (on selected item) | Upgrade this package |
| `U` | Formulae/Casks/Outdated/Status panel | Upgrade ALL outdated |

**Single upgrade flow:**
```
Outdated panel → select ripgrep → press `u`
  → ConfirmModal: "Upgrade ripgrep 14.1.1 → 14.1.2?"
  → User confirms
  → TaskManager: brew upgrade ripgrep
  → ProgressModal: streaming output
  → On success: remove from outdated, update version in formulae list
```

**Upgrade-all flow:**
```
Any panel → press `U`
  → ConfirmModal: "Upgrade all 7 outdated packages?
    (2 pinned packages will be skipped)"
  → User confirms
  → TaskManager: brew upgrade (no args)
  → ProgressModal: streaming output (shows each package being upgraded)
  → On success: clear outdated list, refresh all panels
```

> **Note:** The upgrade-all confirmation message must reflect the actual count excluding pinned packages. If 9 packages are outdated but 2 are pinned, the message says "Upgrade 7 outdated packages? (2 pinned skipped)".

**Acceptance criteria:**
- [ ] Single formula upgrade works
- [ ] Single cask upgrade works
- [ ] Upgrade all works
- [ ] Version changes reflected in panels after upgrade
- [ ] Outdated panel updates (item removed or list cleared)
- [ ] Progress shows for each package during upgrade-all
- [ ] Confirmation includes version change (old → new)

---

### 6.5 — Reinstall Flow (Gap from Audit)

**What:** Reinstall a formula or cask (uninstall + install) for fixing broken installs.

**Keybinding:** `r` in Formulae/Casks panels (on an installed item)

**Flow:**
```
Formulae panel → select a formula → press `r`
  → ConfirmModal: "Reinstall ripgrep 14.1.2?"
    "This will uninstall and reinstall the package."
  → User confirms
  → TaskManager: brew reinstall ripgrep
  → ProgressModal: streaming output
  → On success: refresh panel, toast "✓ ripgrep reinstalled"
```

**Acceptance criteria:**
- [ ] Reinstall from Formulae panel works
- [ ] Reinstall from Casks panel works
- [ ] Confirmation modal with package name + version
- [ ] Progress streams in real time
- [ ] Panel refreshes on success
- [ ] Cache invalidated on success

---

### 6.6 — Update Homebrew Metadata

**What:** Run `brew update` to fetch latest formulae/cask definitions.

**Keybinding:** `u` in Status panel

**Flow:**
```
Status panel → press `u`
  → ProgressModal: "Updating Homebrew..."
  → TaskManager: brew update
  → On success: invalidate all caches, refresh all panels, show toast
```

**Acceptance criteria:**
- [ ] `brew update` runs with progress
- [ ] All caches invalidated after update
- [ ] All panels refresh with fresh data
- [ ] "Last update" timestamp refreshes in Status panel

---

### 6.7 — Batch Selection (Outdated Panel)

**What:** Allow selecting multiple packages for batch upgrade.

**Keybinding:** `Space` in Outdated panel toggles selection

**Visual:**
```
  ✓ 📦 ripgrep        14.1.1  →  14.1.2
    📦 node           22.5.0  →  22.6.1
  ✓ 🖥  firefox        134.0   →  135.0
    📦 python@3.12    3.12.7  →  3.12.8
```

- `✓` = selected for batch operation
- `Space` toggles
- `u` with selections → upgrade only selected packages
- `a` = select all, `A` = deselect all

**Acceptance criteria:**
- [ ] Space toggles selection marker
- [ ] `a` selects all, `A` deselects all
- [ ] `u` with selections upgrades only selected
- [ ] `u` without selections upgrades the cursor item
- [ ] Selection state is visual (checkmark)
- [ ] Selection count shown: "3 of 7 selected"
- [ ] **Selection persists when switching panels and returning** (stored in the root model, not the panel)
- [ ] Selection is cleared after a successful batch upgrade

---

### 6.8 — Fetch (Pre-Download Without Installing)

**What:** Pre-download a formula or cask without installing it. Useful for offline preparation or slow networks.

**Keybinding:** `F` in Formulae/Casks/Search panels (on an uninstalled item)

**Flow:**
```
Search panel → select a package → press `F`
  → ProgressModal: "Fetching neovim..."
  → TaskManager: brew fetch neovim [--all-platforms]
  → On success: toast "✓ neovim downloaded"
```

**Options:**
- `--all-platforms` for casks: download for all platforms
- Useful before going offline: fetch now, install later

**Acceptance criteria:**
- [ ] Fetch from Search/formulae/casks panels works
- [ ] `--all-platforms` flag available for casks
- [ ] Progress streams in real time
- [ ] Success/failure toast
- [ ] Already-installed check (no need to fetch installed packages)

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestTaskManager_Sequential` | Unit | `internal/gui/task/manager_test.go` | Tasks run one at a time |
| `TestTaskManager_Queue` | Unit | `internal/gui/task/manager_test.go` | Tasks queue when one is running |
| `TestTaskManager_Cancel` | Unit | `internal/gui/task/manager_test.go` | Cancel stops the running task |
| `TestTaskManager_Output` | Unit | `internal/gui/task/manager_test.go` | Output lines stream to channel |
| `TestInstallFlow` | E2E (teatest) | `internal/gui/flows/install_test.go` | Search → select → i → confirm → progress → success |
| `TestInstallAlreadyInstalled` | E2E (teatest) | `internal/gui/flows/install_test.go` | Already installed shows warning |
| `TestUninstallFlow` | E2E (teatest) | `internal/gui/flows/uninstall_test.go` | Select → x → confirm → success → item removed |
| `TestUninstallCancel` | E2E (teatest) | `internal/gui/flows/uninstall_test.go` | Cancel confirmation aborts |
| `TestUninstallWithDeps` | E2E (teatest) | `internal/gui/flows/uninstall_test.go` | Dependency warning shown |
| `TestUpgradeSingle` | E2E (teatest) | `internal/gui/flows/upgrade_test.go` | Single package upgrade |
| `TestUpgradeAll` | E2E (teatest) | `internal/gui/flows/upgrade_test.go` | Upgrade all flow |
| `TestBatchSelect` | E2E (teatest) | `internal/gui/flows/batch_test.go` | Space toggles, a selects all |
| `TestBatchUpgrade` | E2E (teatest) | `internal/gui/flows/batch_test.go` | Only selected packages upgraded |
| `TestUpdateBrew` | E2E (teatest) | `internal/gui/flows/update_test.go` | brew update runs and refreshes |
| `TestReinstallFormula` | E2E (teatest) | `internal/gui/flows/reinstall_test.go` | Select → r → confirm → progress → success |
| `TestReinstallCask` | E2E (teatest) | `internal/gui/flows/reinstall_test.go` | Reinstall cask works |
| `TestFetchFormula` | E2E (teatest) | `internal/gui/flows/fetch_test.go` | Fetch formula works |
| `TestFetchCask` | E2E (teatest) | `internal/gui/flows/fetch_test.go` | Fetch cask with --all-platforms |
| `TestCacheInvalidationAfterInstall` | Unit | `internal/brew/cache_test.go` | Install invalidates formulae cache |
| `TestCacheInvalidationAfterUpgrade` | Unit | `internal/brew/cache_test.go` | Upgrade invalidates outdated cache |

---

## Definition of Done

- [ ] Task manager handles sequential execution with queuing
- [ ] Install works for formulae and casks (from search)
- [ ] Uninstall works with confirmation and dependency warning
- [ ] Reinstall works for formulae and casks (gap from coverage audit)
- [ ] Fetch works for formulae and casks with --all-platforms (gap from coverage audit)
- [ ] Upgrade works (single + all + batch)
- [ ] `brew update` works from Status panel
- [ ] Progress modal streams output in real-time
- [ ] Post-mutation: cache invalidated, panels refreshed, toast shown
- [ ] Batch selection in Outdated panel
- [ ] Brew runs non-interactively (HOMEBREW_NO_ASK set; 6.0.0 requirement)
- [ ] All tests pass
- [ ] App is now genuinely useful for daily homebrew management
