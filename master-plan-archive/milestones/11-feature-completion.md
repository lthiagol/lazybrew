# Milestone 11 — Feature Completion

> **Status:** ⚠️ Partial  
> **Remaining:** Verify services run, brewfile, vulns, missing in smoke  
> **Depends on:** Milestone 10 (GUI Architecture — tab content must work before new features land correctly)  
> **Enables:** Milestone 12 (Test Infrastructure — features need tests)

---

## Active Work Routing

> **Format:** Legacy. Coverage audit targets; verify don't re-implement.

| Open item | Execute in |
|---|---|
| Feature wiring verification | [M20.11](../smoke-checklist.md) |
| Search info preview | [M17.11](17-lazygit-tui-and-auto-update.md) |

---

## Goal

Wire the remaining Homebrew commands that are defined in the coverage audit as gaps but never received keybindings or GUI integration. After this milestone, the coverage audit reads 100% for P0/P1 commands.

---

## Steps

### 11.1 — Wire Services Run (`f`)

**What:** The plan defines `f` for "Run service" (foreground, no auto-start). The `ServicesService.Run()` method exists but is never called from the GUI.

**Keybinding:** `f` in Services panel

**Implementation:** Add `case "f":` to the keybinding switch that calls `ServicesService.Run()`. No confirmation needed — run is non-destructive.

**File:** `internal/gui/keybindings.go` (post-M10 decomposition)

**Acceptance criteria:**
- [ ] `f` calls `brew services run <name>`
- [ ] Status indicator updates after run
- [ ] `f` is disabled for stopped services (cannot run a stopped service)
- [ ] Bottom bar shows `f` hint in Services panel

---

### 11.2 — Wire Services Cleanup (`c`)

**What:** `brew services cleanup` removes stale service files. The `ServicesService` has no `Cleanup` method yet.

**Keybinding:** `c` in Services panel

**Implementation:**
- Add `Cleanup(ctx context.Context) error` to `ServicesService` (calls `brew services cleanup --all`)
- Wire `case "c":` to call it with confirmation modal

**Files:**
- `internal/brew/services.go` — add `Cleanup` method
- `internal/gui/keybindings.go` — wire keybinding

**Acceptance criteria:**
- [ ] Cleanup runs with confirmation
- [ ] Success toast on completion
- [ ] Error handling for failed cleanup

---

### 11.3 — Wire Brewfile Menu (`B`)

**What:** The plan defines `B` in Status panel for Brewfile operations (dump, install, cleanup, check, list). None of these are wired.

**Keybinding:** `B` in Status panel

**Implementation:**
- On `B` press, show a `MenuModal` with options:
  1. Export to Brewfile (`brew bundle dump`)
  2. Install from Brewfile (`brew bundle install`)
  3. Cleanup (`brew bundle cleanup`)
  4. Check (`brew bundle check`)
  5. List (`brew bundle list`)
- Each option fires the appropriate brew command via `DefaultRunner.Execute` or `ExecuteStream`
- Results shown in the main panel or via toast

**Files:**
- `internal/gui/gui.go` — add `brewfileAction()` method
- `internal/gui/keybindings.go` — wire `B`

**Design note — no BrewfileService:** The brew bundle commands are simple enough to call directly via the runner without a dedicated service interface. Create a small helper in `internal/brew/bundle.go` if the commands need more than one line each.

**Acceptance criteria:**
- [ ] `B` opens Brewfile menu modal
- [ ] Export creates a Brewfile at specified path
- [ ] Install runs with progress modal
- [ ] Cleanup shows dry-run first, then confirmed run
- [ ] Check verifies Brewfile satisfaction
- [ ] Lists entries in main panel
- [ ] All operations use `HOMEBREW_NO_ASK=1` (runner already does this)

---

### 11.4 — Wire `brew uses` to "Used By" Tab

**What:** The plan defines a "Used By" tab in the Formulae panel showing `brew uses --installed <name>`. The `FormulaeReader.Uses()` method exists but is never called from the GUI.

**Implementation:** In `renderContent`, when `activePanel == PanelFormulae && activeTab == 2`, call `FormulaeReader.Uses()` and display results.

**Data flow:**
- `FormulaeReader.Uses()` returns `[]string` (names of dependents)
- Format as a simple list in the main panel
- Cache the result so switching tabs doesn't re-fetch

**Files:** `internal/gui/gui.go` (renderContent switch) or `internal/gui/render.go` (post-M10)

**Acceptance criteria:**
- [ ] "Used By" tab shows `brew uses --installed <name>`
- [ ] Loading state shown while fetching
- [ ] Empty state: "No dependents"
- [ ] Error state handled

---

### 11.5 — Wire `brew vulns` and `brew missing` (from M8 Plan)

**What:** The plan defines `v` for vulnerability check and `m` for missing deps check in the Status panel. These exist in milestone documents but were never wired.

**Keybinding:** `v` and `m` in Status panel

**Implementation:**
- `v`: Show progress modal, run `brew vulns`, display results in main panel
- `m`: Show progress modal, run `brew missing` (already in `DiagnosticsReader.Missing()`), display results

**Files:** `internal/gui/keybindings.go`

**Acceptance criteria:**
- [ ] `v` runs vulnerability check with progress indicator
- [ ] Results shown in main panel (CVE list or "No vulnerabilities")
- [ ] `m` runs missing deps check
- [ ] Results shown in main panel (formula: missing dep list or "All satisfied")

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestServicesRun` | E2E | `internal/gui/flows/services_test.go` | `f` runs service |
| `TestServicesCleanup` | E2E | `internal/gui/flows/services_test.go` | `c` cleans up services |
| `TestBrewfileExport` | E2E | `internal/gui/flows/bundle_test.go` | Export creates Brewfile |
| `TestBrewfileCheck` | E2E | `internal/gui/flows/bundle_test.go` | Check verifies |
| `TestBrewfileList` | E2E | `internal/gui/flows/bundle_test.go` | List displays entries |
| `TestUsedByTab` | E2E | `internal/gui/gui_test.go` | Tab shows dependents |
| `TestVulnsCheck` | E2E | `internal/gui/flows/vulns_test.go` | Vulns check runs |
| `TestMissingCheck` | E2E | `internal/gui/flows/missing_test.go` | Missing check runs |
| `TestBrewfileCleanup` | Unit | `internal/brew/bundle_test.go` | Cleanup dry-run logic |

---

## Definition of Done

- [ ] Services run (`f`) and cleanup (`c`) wired
- [ ] Brewfile menu (`B`) with all 5 sub-actions
- [ ] "Used By" tab shows reverse dependencies
- [ ] Vulns check (`v`) and missing deps check (`m`) wired
- [ ] All keybindings shown in bottom bar hints
- [ ] Help overlay includes new keybindings
- [ ] Coverage audit updated to reflect newly covered commands
- [ ] All tests pass
