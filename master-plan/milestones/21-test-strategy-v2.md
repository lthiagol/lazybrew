# Milestone 21 — Test Strategy v2

> **Status:** 🔜 Planned  
> **Size estimate:** M–L (4–5 days, can split across M19–M20)  
> **Depends on:** M19.5 (teatest against stable Model), M20 phase A (tab tests)  
> **Enables:** M22 full gates, M17 safe refactor  
> **Parallel track:** D (Quality) — tiers T0–T3 start at different times

See [planning-challenge-2026-06-13.md](../planning-challenge-2026-06-13.md) — do not wait for all of M20.

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

## Architecture Decisions (ADRs)

| ID | Decision |
|---|---|
| D21-1 | teatest for E2E; keep unit tests for pure logic |
| D21-2 | Integration tag `integration`; never run in default CI |
| D21-3 | Golden snapshots only for presentation strings (stable) |
| D21-4 | MockRunner records args for command verification |

---

## Step Index

| Step | Title | Size | Tier | Depends |
|---|---|---|---|---|
| 21.0 | MockRunner call recorder | S | T0 | — |
| 21.6 | Concurrent cache expiry test | S | T0 | — |
| 21.7 | (Moved to M19.0) TypedCache | — | T0 | done in M19 |
| 21.1 | teatest helper + fixtures | M | T1 | M19.5 |
| 21.3 | Integration test suite | M | T1 | — |
| 21.2 | Core E2E flows | L | T2 | 21.1, M20.1 |
| 21.4 | Regression tests from review | M | T2 | M19.6, M20 |
| 21.5 | Coverage gates in Makefile | S | T3 | 21.2 |

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

**Size:** L · **Tier:** T2 · **Depends on:** 21.1, M20.1

**Directory:** `internal/gui/flows/`

| # | Test file | Flow | Key assertions |
|---|---|---|---|
| 1 | `navigation_test.go` | Tab between panels | View contains "Formulae" |
| 2 | `search_test.go` | `/` query Enter | Search panel active |
| 3 | `install_test.go` | Search → i | Mock install called |
| 4 | `uninstall_test.go` | x → confirm | Mock uninstall |
| 5 | `refresh_test.go` | R | Mock list called again |
| 6 | `modal_test.go` | Modal open | `q` doesn't quit |
| 7 | `tabs_test.go` | `]` Deps tab | View contains mock deps string |
| 8 | `help_test.go` | `?` | View contains help text |

**Pattern:**
```go
tm := teatest.NewTestModel(t, newTestModel(), teatest.WithInitialTermSize(120, 40))
tm.Send(tea.KeyMsg{...})
tm.WaitTeatestExpectedTerminals(t, expectedViewSubstring)
```

**Acceptance criteria:**
- [ ] ≥8 flow tests with View assertions (not just model field checks)
- [ ] All pass `-race`

---

### 21.4 — Regression Tests from Architecture Review

**Size:** M · **Tier:** T2

Map findings to tests:

| Finding ID | Test | Step dependency |
|---|---|---|
| C1 | `TestDoMutationInvalidSelectionNotStuck` | M19.6 |
| C3 | `TestTabContentChangesWithSelection` | M20.1 |
| H1 | `TestBatchUpgradeCallsBrewWithSelectedNames` | M20.3 |
| H2 | `TestPinRespectsPinnedFlag` | M20.4 |
| H3 | `TestOutdatedFetchSurfacesError` | M20.9 |

**Acceptance criteria:**
- [ ] Each critical/high finding from architecture-review has a linked test name in review doc appendix

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

- [ ] T0–T3 complete (21.7 satisfied by M19.0)
- [ ] ≥8 teatest flows
- [ ] ≥5 integration tests
- [ ] Regression tests for C1, C3, H1–H3
- [ ] `make cover-check` exists
- [ ] Test tiers documented in DESIGN + AGENTS

---

## Post-Milestone Gate

- [ ] M22.2 integration workflow
- [ ] M22.5 release checklist includes test tier verification
