# Milestone 21 — Test Strategy v2

> **Status:** ⚠️ Partial (~80% done)  
> **Size estimate:** S remaining (2 teatest flows)  
> **Depends on:** M19.5 (done), M20 phase A (done)  
> **Enables:** M22 full gates, M17 safe refactor  
> **Parallel track:** D (Quality) — T2 remaining

See [planning-challenge-2026-06-13.md](../archive/planning-challenge-2026-06-13.md) — do not wait for all of M20.

---

## Goal

Build a **layered test pyramid** where failures indicate real regressions. Coverage percentages are gates on critical packages, not the primary metric.

---

## Tier Schedule

| Tier | Steps | Start when |
|---|---|---|
| **T0 Safety** | 21.0, 21.6, 21.7 | Immediately / during M19 |
| **T1 Infra** | 21.1, 21.3 | M19.5 |
| **T2 Behavior** | 21.2, 21.4 | M20 phase A |
| **T3 Gates** | 21.5 | M21 T2 complete |

---

## Out of Scope

- 80% overall line coverage target (misleading for GUI)
- Visual screenshot testing
- Performance benchmarks (backlog B-02)

---

## Reality Check (2026-06-14)

Implemented:

- **21.0** `MockRunner` records calls (used by batch upgrade regression test).
- **21.1** `internal/gui/testutil/program.go` helper exists.
- **21.3** `internal/brew/runner_integration_test.go` with 5 integration tests.
- **21.4** `internal/gui/regression_test.go` covers C1/C3/H1–H3 findings.
- **21.5** `make cover-check` + `scripts/check-coverage.sh` exist.
- **21.2** 6 teatest flows exist in `internal/gui/flows/` (navigation, search, help, refresh, modal, tabs).

Remaining: 2 flow tests from the original 8 (install, uninstall).

---

## Architecture Decisions (ADRs)

| ID | Decision |
|---|---|
| D21-1 | teatest for E2E; keep unit tests for pure logic |
| D21-2 | Integration tag `integration`; never run in default CI |
| D21-3 | Golden snapshots only for presentation strings (stable) |
| D21-4 | MockRunner records args for command verification |

---

## Step Index

| Step | Title | Size | Tier | Status | Depends |
|---|---|---|---|---|---|
| 21.0 | MockRunner call recorder | S | T0 | Done | — |
| 21.6 | Concurrent cache expiry test | S | T0 | Done | — |
| 21.7 | (Moved to M19.0) TypedCache | — | T0 | Done | M19.0 |
| 21.1 | teatest helper + fixtures | M | T1 | Done | M19.5 |
| 21.3 | Integration test suite | M | T1 | Done | — |
| 21.2 | Core E2E flows | S remaining | T2 | **Partial** | 21.1, M20.1 |
| 21.4 | Regression tests from review | M | T2 | Done | M19.6, M20 |
| 21.5 | Coverage gates in Makefile | S | T3 | Done | 21.2 |

### 21.2 breakdown (remaining work)

| Sub-step | Title | Size | Assertion |
|---|---|---|---|
| 21.2a | Install flow teatest | S | Search `/` → `i` → mock install called |
| 21.2b | Uninstall flow teatest | S | `x` → confirm → mock uninstall called |

---

## Steps

### 21.0 — MockRunner Call Recorder

**Size:** S · **Tier:** T0

**File:** `internal/brew/runner_test.go` or extend `MockRunner` in `runner.go`

**Implementation:**
```go
type RecordedCall struct { Args []string; At time.Time }
type RecordingRunner struct { MockRunner; Calls []RecordedCall }
```

Helper: `AssertCalled(t, "install", "ripgrep")`

**Acceptance criteria:**
- [ ] Batch upgrade tests can assert brew args without real brew

**Tests:** Self-test `TestRecordingRunnerCapturesCalls`

---

### 21.6 — Concurrent Cache Expiry Test

**Size:** S · **Tier:** T0 · **M13 remainder**

**File:** `cache_test.go`

**Implementation:**
1. Cache with 1ms TTL
2. Set key; spawn 20 goroutines calling Get after sleep
3. Run with `-race`; expect no fatal

**Acceptance criteria:**
- [ ] `go test -race -run TestCacheConcurrentExpiry` passes 100x (optional stress loop)

---

### 21.1 — teatest Helper + Fixtures

**Size:** M · **Tier:** T1 · **Depends on:** M19.5

**Dependency:**
```bash
go get github.com/charmbracelet/x/exp/teatest@latest
```
Pin version in commit; verify compatible with bubbletea v1.3.x.

**Files:**
- `internal/gui/testutil/program.go`
- `internal/gui/testutil/fixtures.go`

**API:**
```go
func NewTestModel(t *testing.T, opts ...TestOption) *Model
func RunTest(t *testing.T, m *Model, keys ...string) string // returns final View
```

**Fixtures:** Load formulae JSON from `testdata/formulae_installed.json` into MockRunner responses.

**Acceptance criteria:**
- [ ] Helper produces stable View for `newTestModel` + WindowSizeMsg
- [ ] Documented in AGENTS.md

**Tests:** `TestTestutilHelperRenders`

---

### 21.3 — Integration Test Suite

**Size:** M · **Tier:** T1 · **Can start early**

**File:** `internal/brew/runner_integration_test.go`

```go
//go:build integration
```

| Test | Command | Skip condition |
|---|---|---|
| `TestIntegrationBrewVersion` | `--version` | `findBrewPath` fails |
| `TestIntegrationBrewListFormulaJSON` | `list --formula --json=v2` | parse error |
| `TestIntegrationBrewSearchJSON` | `search --json=v2 git` | network optional |
| `TestIntegrationBrewDoctor` | `doctor` | none |
| `TestIntegrationBrewConfig` | `config` | none |

**Helper:**
```go
func requireBrew(t *testing.T) *DefaultRunner
```

**Acceptance criteria:**
- [ ] `make test-integration` runs ≥5 tests on machine with brew
- [ ] `make test` does NOT run them

**Tests:** self

---

### 21.2 — Core E2E Flows

**Size:** L · **Tier:** T2 · **Depends on:** 21.1, M20.1 · **Status:** Partial (6 done, 2 remaining)

**Directory:** `internal/gui/flows/`

| # | Test file | Flow | Status | Key assertions |
|---|---|---|---|---|
| 1 | `navigation_test.go` | Tab between panels | Done | View contains "Formulae" |
| 2 | `search_test.go` | `/` query Enter | Done | Search panel active |
| 3 | `install_test.go` | Search → `i` | **Remaining** | Mock install called |
| 4 | `uninstall_test.go` | `x` → confirm | **Remaining** | Mock uninstall called |
| 5 | `refresh_test.go` | `R` | Done | Mock list called again |
| 6 | `modal_test.go` | Modal open | Done | `q` doesn't quit |
| 7 | `tabs_test.go` | `]` Deps tab | Done | View contains mock deps string |
| 8 | `help_test.go` | `?` | Done | View contains help text |

**Pattern:**
```go
tm := teatest.NewTestModel(t, testutil.NewTestModel(t), teatest.WithInitialTermSize(120, 40))
tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
out := readOutput(t, tm.FinalOutput(t))
// assert on `out`
```

**Acceptance criteria:**
- [ ] ≥8 flow tests with View assertions (not just model field checks)
- [ ] All pass `-race`

---

#### 21.2a — Install Flow Teatest

**Size:** S · **Tier:** T2

**What:** Verify that searching for a package and pressing `i` triggers a mock install.

**File:** `internal/gui/flows/install_test.go`

**Implementation plan:**
1. Create a custom `MockRunner` that records `install` calls.
2. Build a `*brew.Client` with that runner.
3. Use `testutil.NewTestModel(t, ...)` with the custom client.
4. Send keys: `/`, type a query, `Enter`, wait for Search panel, press `i`.
5. Wait for model to finish or timeout.
6. Assert that the runner received `install <query>`.

**Edge cases / considerations:**
- The test should not require real Homebrew.
- Use a short final timeout (2–3s) so failures are fast.
- The search modal closes on `Enter`; the search results panel becomes active.
- `i` on Search panel uses `m.executeSearchInstall()` or similar path; verify the exact handler name before writing the test.

**Acceptance criteria:**
- [ ] Test compiles and passes with `go test -race ./internal/gui/flows/...`
- [ ] Test asserts the install command was recorded by the mock runner.

**Tests:** `TestInstallFlow`

---

#### 21.2b — Uninstall Flow Teatest

**Size:** S · **Tier:** T2

**What:** Verify that pressing `x` on an installed formula opens a confirm modal and confirming triggers a mock uninstall.

**File:** `internal/gui/flows/uninstall_test.go`

**Implementation plan:**
1. Create a custom `MockRunner` that records `uninstall` calls.
2. Seed the Formulae panel with one item via the mock client.
3. Start the test model.
4. Send `2` to switch to Formulae panel, then `x`.
5. Confirm the modal appears in the View output.
6. Send `Enter` or `y` to confirm.
7. Assert that the runner received `uninstall <formula>`.

**Edge cases / considerations:**
- The modal blocks `q` until dismissed; use `Enter` to confirm.
- The selected formula name must be parseable from the seeded item string.
- If the handler uses a confirmation modal, wait for the modal state in the output.

**Acceptance criteria:**
- [ ] Test compiles and passes with `go test -race ./internal/gui/flows/...`
- [ ] Test asserts the uninstall command was recorded by the mock runner.

**Tests:** `TestUninstallFlow`

---

### 21.4 — Regression Tests from Architecture Review

**Size:** M · **Tier:** T2

Map each finding to a regression test that prevents re-introduction:

| Finding ID | Finding (short) | What the test must verify | Test name | Step dep. |
|---|---|---|---|---|
| C1 | `doMutation` sets `isBusy=true` then returns without clearing on invalid selection / empty name / nil program | `doMutation` with invalid selection (index out of range, empty name) returns nil cmd, does NOT leave manager stuck in running state (naturally solved by TaskManager — no `isBusy` to get stuck) | `TestDoMutationRejectedWhenRunning` | M19.6 |
| C3 | Tab content cache keyed by `panel:tab` only — ignores selected item; stale Deps/Used By/Files on selection change | Selecting different item on same panel produces different tab cache key; `j`/`k` triggers refetch for tabs that depend on selected item | `TestTabContentChangesWithSelection` | M20.1 |
| H1 | Batch select (`Space`) toggles selection but no batch upgrade implementation existed | Batch upgrade with 2+ selected outdated items enqueues one upgrade Task per item; recorded brew args match selected names | `TestBatchUpgradeCallsBrewWithSelectedNames` | M20.3 |
| H2 | Pin toggle tries `Unpin` first, then `Pin` on any error — wrong semantics for already-pinned packages | Pin toggle calls `Pin` when `Pinned` is false, calls `Unpin` when `Pinned` is true (never both) | `TestPinRespectsPinnedFlag` | M20.4 |
| H3 | `fetchPanelData` Outdated silently ignores errors (`formulae, _ := ...`) | Mock runner error on Outdated fetch → `DataLoadedMsg.Err` is set, panel shows error state | `TestOutdatedFetchSurfacesError` | M20.9 |

**Acceptance criteria:**
- [ ] Each critical/high finding from architecture-review has a linked test name in review doc appendix
- [ ] Each test explicitly asserts the bug scenario (not just a happy-path variant)

---

### 21.5 — Coverage Gates in Makefile

**Size:** S · **Tier:** T3

**File:** `Makefile`

```makefile
cover-check:
	go test ./... -coverprofile=coverage.out -count=1
	@go tool cover -func=coverage.out | ./scripts/check-coverage.sh
```

**`scripts/check-coverage.sh` floors (initial):**

| Package | Min % |
|---|---|
| `internal/brew` | 75 |
| `internal/gui/presentation` | 90 |
| `internal/gui` | 55 |
| `internal/gui/modal` | 60 |

Raise floors only when tests land — script reads actual and compares.

**Acceptance criteria:**
- [ ] `make cover-check` fails if brew drops below floor
- [ ] Documented in AGENTS.md

---

## Test Pyramid (target end state)

```
        ┌─────────────┐
        │  8+ teatest │  flows/
        ├─────────────┤
        │  5 integr.  │  runner_integration_test.go
        ├─────────────┤
        │  snapshots  │  presentation/
        ├─────────────┤
        │  unit       │  brew/, task/, gui unit
        └─────────────┘
```

---

## Definition of Done

- [x] T0–T3 complete (21.7 satisfied by M19.0)
- [ ] ≥8 teatest flows (6 done, 2 remaining)
- [x] ≥5 integration tests
- [x] Regression tests for C1, C3, H1–H3
- [x] `make cover-check` exists
- [x] Test tiers documented in DESIGN
- [ ] Test tiers documented in AGENTS (remaining M18.8)

---

## Post-Milestone Gate

- [ ] M22.2 integration workflow
- [ ] M22.5 release checklist includes test tier verification
