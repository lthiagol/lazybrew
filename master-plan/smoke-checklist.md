# Lazybrew — Manual Smoke Checklist

> **Status:** Signed off M08 (2026-08-02) via **agent hybrid** — automated tests + real brew CLI + operator approval. Full interactive TUI mutations intentionally covered by teatest, not live brew writes.  
> **Requires:** Real Homebrew installation (macOS or Linux).  
> **Duration:** ~15 minutes (full human) / hybrid agent path used this run

---

## Environment

| Field | Value |
|---|---|
| Date | 2026-08-02 |
| Tester | hybrid (runner agent + operator approval) |
| lazybrew version | `lazybrew v1.0.0-rc1` (`make build` → `./bin/lazybrew --version`) |
| brew version | Homebrew 6.0.14-38-g1f3abf4 |
| Terminal size | 80 × 24 |
| OS | Darwin 25.5.0 arm64 |

---

## Launch & Navigation

- [x] `lazybrew` starts without error — `make build`; `go test -race ./internal/app/... ./internal/gui/...`; binary builds
- [x] Sidebar panels available (M12 lazy-load: Status on start; others on first visit within normal use) — `TestNewModel`, `TestInitDoesNotFetchOutdated`, `TestSwitchToOutdatedLazyFetchesOnFirstVisit`
- [x] Tab / Shift+Tab cycles panels — `TestPanelNavigation` (Tab + Shift+Tab)
- [x] Keys 1–7 jump to correct panel — `TestPanelJump` (all 1–7)
- [x] `[` / `]` switch tabs within panel — `TestTabSwitching`
- [x] `?` help opens; Esc closes — `TestHelpToggle`, `TestHelpEscClose`, `flows.TestHelpFlow`

---

## Data Display (M20)

- [x] Formulae **Info** tab shows package details (not duplicate sidebar list) — `presentation.TestFormatFormulaInfoSnapshot`; render path `FormatFormulaInfo`
- [x] Select formula A → Deps tab → j/k to B → Deps updates for B — **PASS** (was FAIL 2026-08-01 / M13). `TestDepsTabLoads`, `TestDepsSelectionChangeUpdatesContent`; real `brew deps --tree ripgrep` returns promptly
- [x] Outdated panel: empty state OR list with correct items — `TestEmptyStateMessages`, `TestOutdatedPanelTypedData`
- [x] Empty formulae edge case (if testable): correct message — `TestEmptyStateMessages`, `TestNoCrashOnEmptyPanel`

---

## Mutations (M19 TaskManager)

> Hybrid: no live brew install/uninstall on dev machine. Covered by teatest + unit TaskManager tests.

- [x] Search `/` → install a small package (`i`) — progress modal streams output — `flows.TestInstallFlow`, progress modal unit tests
- [x] Cancel long operation (Esc in progress modal) — UI recovers, not stuck — `TestProgressModalCancel`, `TestOperationCancelShowsCancelled`, `task.TestManager_CancelCurrent`
- [x] Second operation while first runs — queued or rejected with toast — `TestDoMutationRejectedWhenRunning`, `TestModelTaskRejectedToast`, TaskManager queue tests
- [x] Uninstall with confirmation (`x`) — dependency warning if applicable — `flows.TestUninstallFlow`, `TestConfirmUninstall`
- [x] Pin/unpin (`p`) — correct behavior for pinned vs unpinned — `TestPinRespectsPinnedFlag`, brew pin/unpin writer tests

---

## Outdated & Batch (M20.3)

> Hybrid: no live `brew upgrade`. Selection + batch API covered in unit tests.

- [x] Space toggles selection indicator on outdated items — `TestSpaceTogglesOutdatedBatchSelection`, `TestBatchSelectionShowsIndicator`
- [x] `u` upgrades single item — key path → `confirmMutation(mutUpgrade)`; batch path covered; brew upgrade not live-run
- [x] Multi-select + `u` upgrades selected sequentially — `TestBatchUpgradeUpgradesAllSelected`, `TestBatchUpgradeCallsBrewWithSelectedNames`

---

## Status Panel Actions

- [x] `d` doctor — output in modal / status path — `TestRunDoctor`, doctor tab/dashboard tests; real `brew doctor` completes
- [x] `m` missing — output lines shown — `TestRunMissing`, `TestDiagnosticsMissing`; real `brew missing` completes
- [x] `v` vulns — runs (or clear error if tap missing) — `TestRunVulns` returns cmd
- [x] `R` refresh — panels reload — `TestRefreshKey`, tiered refresh suite, `flows.TestRefreshFlow`

---

## Config (M20.8)

- [x] Custom `brew.path` in config (if tested) — `config.TestLoadValidConfig` path `/custom/brew`
- [x] `auto_refresh_seconds: 30` triggers refresh (observe ~30s) — config load asserts 30; tick/pause suite covers auto-refresh machinery

---

## Terminal Size (M20.7)

- [x] Resize to 79×24 — warning shown — `TestSmallTerminalWarning`, `TestSmallTerminalWarningInView`
- [x] Resize to 80×24 — warning gone — same

---

## Exit

- [x] `q` quits cleanly — `TestQuitKey`; all teatest flows end with `q` + `WaitFinished`
- [x] No panic in terminal after exit — race suite green (`go test -race ./...`); teatest finish without panic

---

## Result

| | |
|---|---|
| **Pass / Fail** | **PASS** (hybrid) — 28/28 checklist items resolved pass via automated evidence + real brew CLI probes; no live brew write mutations |
| **Blocking issues** | None. Prior M13 Deps hang fixed and regression-tested. |
| **Notes** | Operator chose agent hybrid sign-off over full interactive TUI. Checklist item “all 7 panels populate in 10s” reinterpreted under M12 lazy-load (Status first; others on visit). Self-review fixed: AutoRefreshSeconds config assert; Shift+Tab; keys 1–7; Space batch toggle; terminal warning View assert; Help flow hard assert. |
| **M08 state** | Complete (hybrid sign-off 2026-08-02) |
