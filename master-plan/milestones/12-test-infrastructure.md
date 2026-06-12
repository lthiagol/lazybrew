# Milestone 12 — Test Infrastructure & Quality

> **Status:** ⚠️ In Progress  
> **Done:** Modal tests (14), fuzz tests (2), snapshot tests (5)  
> **Remaining:** E2E tests (0), integration tests (0), gui/ package has zero tests  
> **Depends on:** Milestone 10 (GUI Architecture — decomposed code is testable), Milestone 11 (Feature Completion — features must exist before they can be tested)  
> **Enables:** v0.2.0 Release

---

## Goal

Add systematic test coverage for the GUI layer, modal system, and data layer edge cases. Currently the project has 71 tests all in `internal/brew/` and `internal/gui/presentation/`. The GUI layer (`internal/gui/`, `internal/gui/modal/`) has **zero tests**. This milestone adds ~40+ tests across all uncovered areas.

---

## Why This Milestone Exists

The project was built feature-first with tests only for the data layer. The GUI was developed by visual inspection only. This means:
- Refactoring the 792-line `gui.go` (M10) is risky — no safety net
- Modal bugs are found by users, not by CI
- JSON parsing regressions are undetected (no fuzz tests)
- Snapshot changes from formatter updates are invisible

---

## Steps

### 12.1 — Modal Unit Tests

**What:** Test each modal type (confirm, input, menu, progress) in isolation using Bubble Tea's `teatest` framework.

**Challenge:** The modals implement `tea.Model` but they're tested through the root model, not independently. `teatest` sends key messages and asserts on output.

**Files to create:** `internal/gui/modal/confirm_test.go`, `input_test.go`, `menu_test.go`, `progress_test.go`

**Test cases per modal:**

| Modal | Test | What It Validates |
|---|---|---|
| Confirm | `TestConfirmYes` | `y` key returns `Confirmed: true` |
| Confirm | `TestConfirmNo` | `n` key returns `Confirmed: false` |
| Confirm | `TestConfirmEsc` | `Esc` returns `Cancelled: true` |
| Confirm | `TestConfirmDefaultNo` | Default selection is No (safe default) |
| Input | `TestInputSubmit` | `Enter` returns typed text |
| Input | `TestInputEsc` | `Esc` returns cancelled |
| Input | `TestInputPlaceholder` | Placeholder shown when empty |
| Menu | `TestMenuNavigation` | `j/k` moves selection |
| Menu | `TestMenuShortcuts` | Number keys select option |
| Menu | `TestMenuEsc` | `Esc` cancels |
| Progress | `TestProgressAppend` | Lines appended, auto-scroll |
| Progress | `TestProgressDone` | Done state shows success |
| Progress | `TestProgressCancel` | Esc calls cancel func |
| Toast | `TestToastDismiss` | Toast auto-dismisses after 3s |
| Toast | `TestToastTypes` | Different styles render correctly |

**Acceptance criteria:**
- [ ] All modal types have unit tests
- [ ] All keybinding paths covered (Enter, Esc, y, n, j, k, number keys)
- [ ] Edge cases: empty input, long text, rapid key presses

---

### 12.2 — Fuzz Tests for JSON Parsing

**What:** Add Go fuzz tests for the `parseFormula` and `parseCask` functions to ensure they don't panic on malformed input.

**Files to create:**
- `internal/brew/formulae_fuzz_test.go`
- `internal/brew/casks_fuzz_test.go`

**Test approach:**
```go
func FuzzParseFormula(f *testing.F) {
    f.Add(`{"name":"test","versions":{"stable":"1.0"}}`)
    f.Add(`{"name":"","installed":null}`)
    f.Fuzz(func(t *testing.T, data string) {
        var fj formulaJSON
        json.Unmarshal([]byte(data), &fj)  // ignore error
        parseFormula(fj)  // must not panic
    })
}
```

**Acceptance criteria:**
- [ ] Fuzz tests run for at least 30 seconds without panic
- [ ] Edge cases: empty strings, null values, deeply nested JSON, unicode
- [ ] `go test -fuzz` works for both formulae and casks

---

### 12.3 — Presentation Snapshot Tests

**What:** Create snapshot/golden-file tests for all presentation formatters. When formatters change, the golden files update, making changes visible in diffs.

**Files to create:** `internal/gui/presentation/snapshots_test.go`

**Implementation:**
- Use `go-snaps` or a simple custom approach: render known inputs, write to `testdata/snapshots/`, compare on subsequent runs
- Run with `-update` flag to regenerate golden files

**Test cases:**

| Test | Input | Golden File |
|---|---|---|
| `TestSnapshotFormula` | Normal, pinned, outdated, keg-only, no-bottle | `formula_normal.snap` |
| `TestSnapshotCask` | Normal, outdated, auto-update, pinned | `cask_normal.snap` |
| `TestSnapshotTap` | Official, trusted, untrusted, API/clone | `tap_official.snap` |
| `TestSnapshotService` | Started, stopped, error | `service_started.snap` |
| `TestSnapshotDashboard` | Various counts | `dashboard.snap` |
| `TestSnapshotDoctor` | Clean, 1 warning, 3 warnings | `doctor_clean.snap` |

**Acceptance criteria:**
- [ ] All formatters have snapshot coverage
- [ ] CI command (`make test`) fails on snapshot mismatch
- [ ] `make test-update` regenerates golden files
- [ ] Snapshots committed to repo

---

### 12.4 — E2E Flow Tests

**What:** Use `teatest` to simulate full user flows: key presses → UI state changes → panel content updates.

**Challenge:** The current GUI has no `teatest`-compatible tests. We need to set up the test infrastructure first. `teatest` requires `tea.Program` with `tea.WithInput()` for sending key presses and `tea.WithoutSignals()` for test environments.

**Files to create:** `internal/gui/gui_test.go`, `internal/gui/flows/search_test.go`, `internal/gui/flows/install_test.go`, `internal/gui/flows/uninstall_test.go`

**Test cases:**

| Test | Flow | What It Validates |
|---|---|---|
| `TestPanelNavigation` | Tab → Tab → 3 → 1 | Active panel changes |
| `TestTabSwitching` | `[` → `]` | Active tab index changes |
| `TestSearchFlow` | `/` → type "neovim" → Enter | Results appear in Search panel |
| `TestHelpToggle` | `?` → `?` | Help overlay shows/hides |
| `TestRefreshKey` | `R` | Data refreshes |
| `TestInstallFlow` | Search → select → `i` → confirm → progress | Install flow starts |

**Acceptance criteria:**
- [ ] teatest infrastructure is set up (import, helpers)
- [ ] Mock client used (not real brew)
- [ ] At least 5 E2E flows tested
- [ ] Tests run in CI (no brew required)
- [ ] Tests use `t.Parallel()` where safe

---

### 12.5 — Integration Test Setup

**What:** Add build-tagged integration tests that call real `brew` commands. These don't run in normal CI but can be run manually.

**File to create:** `internal/brew/runner_integration_test.go`

**Tests:**
- `TestDefaultRunnerVersion` — `brew --version` returns "Homebrew"
- `TestDefaultRunnerList` — `brew info --json=v2 --installed` is valid JSON
- `TestDefaultRunnerConfig` — `brew config` parses correctly

**Tag:** `//go:build integration`

**Acceptance criteria:**
- [ ] `make test` skips integration tests
- [ ] `make test-integration` runs them
- [ ] Tests fail gracefully with clear message if brew not installed

---

## Tests for This Milestone

| Test | Type | File | Count |
|---|---|---|---|
| Modal confirm/input/menu/progress | Unit | `internal/gui/modal/*_test.go` | ~14 |
| Fuzz: formulae and cask parsing | Fuzz | `internal/brew/*_fuzz_test.go` | 2 |
| Presentation snapshots | Snapshot | `internal/gui/presentation/snapshots_test.go` | ~8 |
| E2E flows | E2E (teatest) | `internal/gui/*_test.go`, `internal/gui/flows/*_test.go` | ~10 |
| Integration (real brew) | Integration | `internal/brew/runner_integration_test.go` | ~3 |
| **Total added** | | | **~37** |

---

## Definition of Done

- [ ] All 4 modal types have unit tests covering all key paths
- [ ] Fuzz tests run for 30s without panic
- [ ] Snapshot tests exist for all presentation formatters
- [ ] E2E tests cover navigation, search, help, and at least one mutation flow
- [ ] Integration tests exist with build tag
- [ ] `make test` runs all unit/snapshot/E2E tests (no brew required)
- [ ] `make test-integration` runs integration tests
- [ ] `make fuzz` runs fuzz tests for 30s
- [ ] Total test count ≥ 100
- [ ] Coverage ≥ 70% (brew) and ≥ 80% (presentation)

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| teatest incompatible with current Bubble Tea version | Medium | High — E2E tests blocked | Pin `teatest` version in go.mod, test compatibility early |
| Fuzz tests find real parsing bugs | High | Medium — fix bugs as found | Expected; fuzzing is designed to find bugs |
| Snapshot golden files rot if formatters change | Medium | Low — CI fails, developer updates goldens | Document `make test-update` in CONTRIBUTING |
| Integration tests require brew v6.0.0 | Low | Low — skip if version mismatch | Check `brew --version` in test setup |
