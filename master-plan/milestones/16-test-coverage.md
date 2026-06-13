# Milestone 16 — Test Coverage Completion

> **Status:** ⚠️ Partial  
> **Tests:** ~162 functions (not 186); gui/ ~31.5% coverage (not 70% target)  
> **Depends on:** Milestone 15 (Architecture Cleanup)  
> **Enables:** Superseded by M21 for release gate

---

## Active Work Routing

> **Format:** Legacy. **Do not treat this file's coverage targets as current gate** — use M21.

| Open item | Execute in |
|---|---|
| E2E + integration | [M21](21-test-strategy-v2.md) |
| Coverage floors (brew 75%, gui 55%, …) | [M21.5](21-test-strategy-v2.md) |
| Package tests started here | Keep; extend in M21 |

---

## Goal

Achieve comprehensive test coverage across all packages. Currently 4 packages have zero tests, and many functions in tested packages are untested.

---

## Current Coverage

| Package | Files | Test Files | Coverage |
|---|---|---|---|
| `internal/brew/` | 11 | 5 | ~65% |
| `internal/gui/modal/` | 5 | 1 | ~60% |
| `internal/gui/presentation/` | 2 | 2 | ~83% |
| `internal/config/` | 1 | **0** | **0%** |
| `internal/app/` | 1 | **0** | **0%** |
| `internal/gui/` | 7 | **0** | **0%** |
| `internal/gui/style/` | 1 | **0** | **0%** |

---

## Steps

### 16.1 — Config Package Tests

**File:** `internal/config/config_test.go`

**Tests:**
| Test | What It Validates |
|---|---|
| `TestDefaultConfig` | Default values are correct |
| `TestLoadMissingFile` | Returns defaults when file doesn't exist |
| `TestLoadValidConfig` | Parses valid YAML correctly |
| `TestLoadInvalidYAML` | Returns error for malformed YAML |
| `TestLoadInvalidTheme` | Falls back to default for invalid theme |
| `TestLoadInvalidSidebarWidth` | Clamps to valid range |
| `TestConfigPathFromEnv` | Uses LAZYBREW_CONFIG env var |

---

### 16.2 — App Package Tests

**File:** `internal/app/app_test.go`

**Tests:**
| Test | What It Validates |
|---|---|
| `TestNewWithDefaults` | Creates model with default options |
| `TestNewWithDebug` | Debug flag enables logging |
| `TestNewWithConfigPath` | Custom config path is used |
| `TestNewBrewNotFound` | Returns error when brew not found |
| `TestNewInvalidConfig` | Returns error for invalid config |

---

### 16.3 — GUI Package Tests

**File:** `internal/gui/gui_test.go`

**Tests:**
| Test | What It Validates |
|---|---|
| `TestNewModel` | Model initializes with correct defaults |
| `TestPanelNavigation` | Tab/Shift+Tab cycles through panels |
| `TestPanelJump` | Number keys 1-7 jump to panels |
| `TestTabSwitching` | [/] switches tabs within panel |
| `TestSearchFlow` | / opens search, Enter triggers search |
| `TestHelpToggle` | ? shows/hides help overlay |
| `TestHelpEscClose` | Esc closes help overlay |
| `TestRefreshKey` | R triggers data refresh |
| `TestQuitKey` | q quits the app |
| `TestWindowSizeMsg` | Layout adjusts to terminal size |
| `TestModalInputRouting` | Modal captures input, background doesn't |
| `TestProgressStreaming` | ProgressLineMsg updates modal |
| `TestMutationResultRefresh` | Mutation completion triggers refresh |

---

### 16.4 — Style Package Tests

**File:** `internal/gui/style/theme_test.go`

**Tests:**
| Test | What It Validates |
|---|---|
| `TestDarkThemeColors` | Dark theme has correct color values |
| `TestLightThemeColors` | Light theme has correct color values |
| `TestApplyTheme` | ApplyTheme updates all package variables |
| `TestThemeConcurrency` | Theme replacement is race-free |

---

### 16.5 — Untested Functions in Tested Packages

**File:** `internal/brew/*_test.go`

**Tests:**
| Test | File | What It Validates |
|---|---|---|
| `TestSetDebug` | `logger_test.go` | SetDebug changes log level |
| `TestBrewNotFoundError` | `errors_test.go` | Error message format |
| `TestBrewExitError` | `errors_test.go` | Error message format |
| `TestJSONParseError` | `errors_test.go` | Error message format, Unwrap |
| `TestTimeoutError` | `errors_test.go` | Error message format |
| `TestDefaultRunnerExecute` | `runner_test.go` | Executes brew command |
| `TestDefaultRunnerExecuteJSON` | `runner_test.go` | Parses JSON output |
| `TestDefaultRunnerExecuteStream` | `runner_test.go` | Streams output line by line |
| `TestFindBrewPath` | `runner_test.go` | Finds brew in standard locations |
| `TestServicesCleanup` | `services_test.go` | Cleanup calls correct command |
| `TestTrustUntrustFormula` | `trust_test.go` | Untrust formula calls correct command |
| `TestTrustUntrustCask` | `trust_test.go` | Untrust cask calls correct command |
| `TestFormulaeGetError` | `formulae_test.go` | Get handles brew errors |
| `TestLeavesEmpty` | `formulae_test.go` | Handles empty output |
| `TestDepsError` | `formulae_test.go` | Deps handles errors |
| `TestUsesError` | `formulae_test.go` | Uses handles errors |

**File:** `internal/gui/modal/*_test.go`

**Tests:**
| Test | File | What It Validates |
|---|---|---|
| `TestConfirmModalView` | `modal_test.go` | View renders correctly |
| `TestInputModalView` | `modal_test.go` | View renders correctly |
| `TestMenuModalView` | `modal_test.go` | View renders correctly |
| `TestProgressModalView` | `modal_test.go` | View renders correctly |
| `TestProgressModalAppendLineReady` | `modal_test.go` | AppendLine with viewport ready |
| `TestToastTypes` | `toast_test.go` | Different toast types render differently |
| `TestToastView` | `toast_test.go` | View renders correctly |

**File:** `internal/gui/presentation/*_test.go`

**Tests:**
| Test | File | What It Validates |
|---|---|---|
| `TestFormatOutdatedFormula` | `snapshot_test.go` | Formats outdated formula correctly |
| `TestFormatOutdatedCask` | `snapshot_test.go` | Formats outdated cask correctly |

---

## Test Count Summary

| Package | New Tests |
|---|---|
| `internal/config/` | 7 |
| `internal/app/` | 5 |
| `internal/gui/` | 13 |
| `internal/gui/style/` | 4 |
| `internal/brew/` (untested) | 15 |
| `internal/gui/modal/` (untested) | 7 |
| `internal/gui/presentation/` (untested) | 2 |
| **Total** | **53** |

Combined with existing 102 tests: **155 total tests**

---

## Coverage Targets

| Package | Current | Target |
|---|---|---|
| `internal/brew/` | 65% | 85% |
| `internal/gui/modal/` | 60% | 90% |
| `internal/gui/presentation/` | 83% | 95% |
| `internal/config/` | 0% | 90% |
| `internal/app/` | 0% | 80% |
| `internal/gui/` | 0% | 70% |
| `internal/gui/style/` | 0% | 90% |
| **Overall** | ~50% | **80%** |

---

## Definition of Done

- [ ] All 4 zero-coverage packages have tests
- [ ] All untested functions in tested packages have tests
- [ ] Total test count ≥ 150
- [ ] Overall coverage ≥ 80%
- [ ] All packages individually ≥ 70% coverage
- [ ] `go test -race ./...` passes
- [ ] `go test -cover ./...` shows coverage report
- [ ] No regressions in existing tests
