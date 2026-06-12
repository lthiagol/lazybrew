# Milestone 4 — Read-Only Panels (Real Data)

> **Status:** 🔲 Not Started  
> **Depends on:** Milestone 2 (TUI Shell), Milestone 3 (Brew Data Layer)  
> **Enables:** Milestone 5 (Modals & Search), Milestone 6 (Mutations)

---

## Goal

Replace mock data with real brew output in all sidebar panels and the main content area. After this milestone, lazybrew is a **read-only dashboard** — you can browse your installed formulae, casks, taps, outdated packages, and status, with detailed info in the right panel. No mutations yet.

This is the first milestone where the app feels genuinely useful.

---

## Steps

### 4.1 — Wire Brew Client into the App

**What:** Initialize the real `brew.Client` at app startup and inject it into the GUI model.

**File:** `internal/app/app.go`, `internal/gui/gui.go`

**Implementation:**
- `app.New()` creates the brew runner, client, and GUI model
- The `gui.Model` receives a `*brew.Client` in its constructor
- On startup (`Init()`), the model dispatches async commands to fetch initial data
- Use a `tea.Cmd` that calls `client.Formulae.List()` in a goroutine and returns a `FormulaeLoadedMsg`

**Bubble Tea async pattern:**
```go
type FormulaeLoadedMsg struct {
    Formulae []brew.Formula
    Err      error
}

func fetchFormulae(client *brew.Client) tea.Cmd {
    return func() tea.Msg {
        formulae, err := client.Formulae.List(context.Background())
        return FormulaeLoadedMsg{Formulae: formulae, Err: err}
    }
}
```

**Acceptance criteria:**
- [ ] App initializes brew client on startup
- [ ] Loading state shown while fetching ("Loading formulae...")
- [ ] Error state shown if brew is not installed
- [ ] Data fetched asynchronously (UI doesn't freeze)

---

### 4.2 — Formulae Panel (Real Data)

**What:** Display real installed formulae in the sidebar list, with info in the main panel.

**Files:** `internal/gui/controllers/formulae_panel.go`, `internal/gui/presentation/formulae.go`

**List item format:**
```
  neovim                    0.10.4     ✓ bottled
▸ ripgrep                   14.1.1     ✓ bottled   ⬆ 14.1.2
  tree-sitter               0.25.3     ✓ bottled
  ⊘ python@3.12             3.12.8     pinned
```

**Presentation formatter:**
- Name: left-aligned, fixed width
- Version: right of name
- Badges: `✓ bottled`, `⬆ x.y.z` (outdated), `⊘ pinned`
- Colors: outdated in warning color, pinned in secondary color
- Count in panel title: `📦 Formulae (124)`

**Main panel tabs when Formulae is active:**
1. **Info** — name, version, description, homepage, license, install path, size, installed date; **6.0.0 adds:** binaries (executables), installed dependents, upgrade hints, shadowing warnings
2. **Deps** — output of `brew deps --tree <name>` (rendered as-is, monospace)
3. **Caveats** — caveats text (or "No caveats" if empty)
4. **Files** — output of `brew list <name>` (files installed by the formula)
5. **Used By** — `brew uses --installed <name>` (reverse dependencies; gap from coverage audit)

**Acceptance criteria:**
- [ ] Panel shows real installed formulae
- [ ] Correct formatting with aligned columns
- [ ] Outdated formulae highlighted with badge
- [ ] Pinned formulae marked
- [ ] Selecting a formula updates the main panel
- [ ] All 4 tabs populate with real data
- [ ] Loading state while fetching info
- [ ] Empty state: "No formulae installed"

---

### 4.3 — Casks Panel (Real Data)

**What:** Display real installed casks.

**Files:** `internal/gui/controllers/casks_panel.go`, `internal/gui/presentation/casks.go`

**List item format:**
```
  google-chrome             132.0    auto-update
▸ visual-studio-code        1.96.2   ⬆ 1.97.0
  iterm2                    3.5.10
  docker                    4.37.1   auto-update
```

**Main panel tabs:**
1. **Info** — name, version, description, homepage, tap, artifacts
2. **Deps** — dependencies (if any)
3. **Caveats** — caveats text

**Acceptance criteria:**
- [ ] Panel shows real installed casks
- [ ] Auto-updating casks marked
- [ ] Outdated casks highlighted
- [ ] Main panel info populated on selection
- [ ] Empty state: "No casks installed"
- [ ] Cask-specific fields shown (artifacts, auto_updates)

---

### 4.4 — Outdated Panel (Real Data)

**What:** A combined view of all outdated packages (formulae + casks).

**Files:** `internal/gui/controllers/outdated_panel.go`, `internal/gui/presentation/outdated.go`

**List item format:**
```
  📦 ripgrep        14.1.1  →  14.1.2
  📦 node           22.5.0  →  22.6.1
  🖥  firefox        134.0   →  135.0
  📦 python@3.12    3.12.7  →  3.12.8
```

- `📦` = formula, `🖥` = cask
- Shows current → new version with arrow
- Count in panel title: `⏫ Outdated (7)`

**Main panel tabs:**
1. **Info** — same info view as Formulae/Casks
2. **Versions** — current version, new version, changelog link (if available)

**Acceptance criteria:**
- [ ] Shows both outdated formulae and casks
- [ ] Visual distinction between formulae and casks
- [ ] Version comparison shown (old → new)
- [ ] Count is accurate
- [ ] Empty state: "Everything up to date! 🎉"

---

### 4.5 — Taps Panel (Real Data)

**What:** Display tapped repositories with trust indicators.

**Files:** `internal/gui/controllers/taps_panel.go`, `internal/gui/presentation/taps.go`

**List item format:**
```
  homebrew/core              ✓ official   API
  homebrew/cask              ✓ official   API
  homebrew/services          ✓ official   clone
▸ nicknisi/tap               🔓 trusted    clone
  some-org/formulas          ⚠ untrusted  clone
```

**Trust indicators:**
- `✓ official` (green) — homebrew/* taps
- `🔓 trusted` (blue) — explicitly trusted via `brew trust`
- `⚠ untrusted` (amber) — third-party, not yet trusted
- `API` / `clone` — how the tap is sourced

**Main panel tabs:**
1. **Tap Info** — name, remote URL, formula count, cask count, last commit; **6.0.0:** `brew tap-info --json` now includes `formula_names` and `cask_names` arrays
2. **Trust** — current trust status, available trust actions; **6.0.0:** `trusted` field in tap-info JSON
3. **Formulae** — list of formulae from this tap

**Acceptance criteria:**
- [ ] All taps listed with trust indicators
- [ ] Official taps auto-detected
- [ ] Trust status fetched and displayed
- [ ] Tap info panel populated from `brew tap-info --json`
- [ ] Formulae list from tap shown in tab
- [ ] API vs clone distinction shown

---

### 4.6 — Status Panel (Real Data)

**What:** Dashboard summary panel.

**Files:** `internal/gui/controllers/status_panel.go`, `internal/gui/presentation/status.go`

**Content:**
```
  lazybrew v0.1.0

  Homebrew 6.0.0
  Prefix: /opt/homebrew
  Last update: 2 hours ago
  
  ────────────────────────
  
  📦 Formulae     124 installed
  🖥  Casks         31 installed
  ⏫ Outdated       7 packages
  🔌 Taps           4 (2 official, 2 third-party)
  ⚙  Services       3 (2 running, 1 stopped)
  
  ────────────────────────
  
  Doctor: ✓ No issues
```

**Main panel tabs:**
1. **Config** — full `brew config` output
2. **Doctor** — doctor results (fetched on demand)

**Acceptance criteria:**
- [ ] All counts accurate (formulae, casks, outdated, taps, services)
- [ ] Brew version shown
- [ ] Prefix detected (macOS vs Linux)
- [ ] Doctor status shown (green check or warning count)
- [ ] Config tab shows full brew config

---

### 4.7 — Services Panel (Real Data, Read-Only)

**What:** Display brew services status (no start/stop yet — that's Milestone 8).

**Files:** `internal/gui/controllers/services_panel.go`, `internal/gui/presentation/services.go`

**List item format:**
```
  postgresql@16       ● started    thiago
▸ redis               ○ stopped
  nginx               ✗ error      thiago   exit: 1
```

- `●` green = started, `○` gray = stopped, `✗` red = error

**Main panel tab:**
1. **Status** — service details (name, status, user, file, exit code)

**Acceptance criteria:**
- [ ] Services listed with status indicators
- [ ] Color-coded status
- [ ] Graceful handling when no services exist
- [ ] Graceful handling on systems without service support

---

### 4.8 — Data Refresh Mechanism

**What:** Implement periodic and on-demand data refresh.

**Implementation:**
- Auto-refresh every 60 seconds (configurable, on by default)
- Manual refresh with `R` key (global keybinding)
- Refresh shows a subtle spinner in the panel title
- Only the active panel refreshes on auto-cycle; others refresh when navigated to if stale (cache TTL)
- Use `tea.Tick` for the refresh timer

**Acceptance criteria:**
- [ ] Auto-refresh updates data periodically
- [ ] `R` triggers manual refresh
- [ ] Spinner shown during refresh
- [ ] Cache prevents redundant brew calls
- [ ] Stale panels refresh when navigated to

---

### 4.9 — Presentation / Snapshot Tests

**What:** Create snapshot tests for all presentation formatters.

**Implementation:**
- Each formatter function takes domain types and returns styled strings
- Snapshot tests render a set of known inputs and compare against golden files
- Golden files stored in `testdata/snapshots/`
- Update golden files with `go test -update` flag

**Acceptance criteria:**
- [ ] Snapshot tests for formulae list items
- [ ] Snapshot tests for casks list items
- [ ] Snapshot tests for taps list items (with each trust status)
- [ ] Snapshot tests for outdated items
- [ ] Snapshot tests for services items
- [ ] Snapshot tests for status dashboard
- [ ] All snapshots committed and passing

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestFormulaePresentation` | Snapshot | `internal/gui/presentation/formulae_test.go` | Formatted output matches golden file |
| `TestFormulaeOutdatedBadge` | Snapshot | `internal/gui/presentation/formulae_test.go` | Outdated badge rendering |
| `TestFormulaePinnedBadge` | Snapshot | `internal/gui/presentation/formulae_test.go` | Pinned badge rendering |
| `TestCasksPresentation` | Snapshot | `internal/gui/presentation/casks_test.go` | Cask list formatting |
| `TestTapsPresentation` | Snapshot | `internal/gui/presentation/taps_test.go` | Trust indicator rendering |
| `TestOutdatedPresentation` | Snapshot | `internal/gui/presentation/outdated_test.go` | Version diff rendering |
| `TestServicesPresentation` | Snapshot | `internal/gui/presentation/services_test.go` | Status indicator colors |
| `TestStatusDashboard` | Snapshot | `internal/gui/presentation/status_test.go` | Dashboard summary |
| `TestDataLoading` | E2E (teatest) | `internal/gui/gui_test.go` | Panels populate with mock client data |
| `TestRefreshKey` | E2E (teatest) | `internal/gui/gui_test.go` | `R` triggers data refresh |
| `TestEmptyStates` | E2E (teatest) | `internal/gui/gui_test.go` | Empty panels show friendly messages |
| `TestErrorState` | E2E (teatest) | `internal/gui/gui_test.go` | Brew not found shows error |

---

## Definition of Done

- [ ] All panels show real data from brew
- [ ] Main panel updates contextually when selecting items
- [ ] Loading, empty, and error states handled gracefully
- [ ] Data refresh works (auto + manual)
- [ ] Cache prevents redundant brew calls
- [ ] All presentation snapshot tests pass
- [ ] All E2E tests pass
- [ ] App feels responsive — no UI freezing during data loads
