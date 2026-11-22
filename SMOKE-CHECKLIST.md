# Lazybrew — Manual Smoke Checklist

> **Status:** Signed off M08 (2026-08-02) via **agent hybrid** — automated tests + real brew CLI + operator approval. Full interactive TUI mutations intentionally covered by teatest, not live brew writes.  
> **Requires:** Real Homebrew installation (macOS or Linux).  
> **Duration:** ~15 minutes (full human) / hybrid agent path used this run
> **Provenance:** copied out of the frozen, untracked, local-only `master-plan` archive when planning detached from it. This file is the tracked release gate artifact; the archive is not an input to any decision.

---

## Run 3 — before v1.1.0 (2026-11-22)

> **Status:** signed off by the user (human, interactive) on `wip` after rc5:
> Taps & Trust tree and lock icons, tap counts, tap-add trim, Doctor warnings
> and the Log marks, one Log row per command, sidebar kept during the refresh
> after an upgrade, Search. Result: pass.

---

## Run 2 — v2 overhaul (LB-05, before the `wip` → `stable` PR)

> **Status:** run 2026-11-21 (human, interactive; result below). Covers LB-08
> (v2-01 … v2-11). Run 1 (M08, hybrid) is kept below for history.
> Build: `git pull && make run`. Tick each box; note anything odd under Result.

| Field | Value |
|---|---|
| Date | 2026-11-21 |
| Tester | user (human, interactive) |
| lazybrew commit | `4bf1584` |
| brew version | Homebrew 7.0.15-111-ge40c695 |
| Terminal / size | herdr, xterm-256color, 140x38 |
| OS | Linux 6.12 arm64 |

### Look and startup
- [ ] Starts without error; header, sidebar boxes, main area and footer fill the terminal
- [ ] Every sidebar box loads by itself after Status (spinner, then rows and a count) — no `not loaded` (v2-08)
- [ ] Status dashboard shows Outdated and the other counts without pressing `R`
- [ ] Keys work while panels are still loading
- [ ] `LAZYBREW_ICONS=unicode make run` looks right; `theme: light` in config looks right
- [ ] Motion feels short and subtle (tabs `h` `l`, focus moves, a dialog opening); `LAZYBREW_ANIMATIONS=0` turns it off

### Navigation and main area (v2-03b, v2-09)
- [ ] Tab / Shift+Tab, `1`–`7`, `h` `l`, mouse click on boxes / rows / tabs, wheel scroll
- [ ] Main area has padding; nothing touches the border
- [ ] Formulae → Info at your usual width: two columns when wide enough, nothing cut mid-word
- [ ] Formulae → Deps / Used By: full-width table, long descriptions end with `…`
- [ ] Log box is small when idle and grows while something runs
- [ ] `?` help scrolls (`j`/`k`); `esc` closes; paste into Search (`/`) works

### Operations (live — use a small, unneeded formula, e.g. `hello`)
- [ ] Search → install `hello` (`i`): steps stream, `✓ Completed in …`, then the view **closes itself after ~2 s** (v2-07)
- [ ] Uninstall it (`x`): confirm names the consequence; completes and closes
- [ ] Start a longer install/upgrade and press **Esc** mid-run: it shows `Cancelled · Esc to close` promptly (v2-11 — no waiting for a download to finish)
- [ ] A failed operation stays on screen with `Esc to close`; switching panel closes it
- [ ] (optional) Batch upgrade in Outdated (`space`, `u`): progress bar `n/total`

### Services (v2-05s)
- [ ] Services → Info shows status, PID, paths, and the actions that apply
- [ ] `r` restart on a service → confirm → completes → main area returns to Info by itself (the bug you hit)
- [ ] `s` on a running service / `S` on a stopped one → info toast, nothing runs

### Taps (v2-10)
- [ ] Tap Info shows the actions that apply: untrusted `t trust`, trusted `t untrust`, official `r repair` only
- [ ] `t` on an untrusted tap trusts it at once; `t` again opens the **untrust confirm** → confirm → back to untrusted
- [ ] `T` opens the per-formula / per-cask trust menu
- [ ] `x` on a tap opens the untap confirm with its consequence (and installed packages, if any) — `esc` unless you mean it
- [ ] After a tap action the cursor stays on the same tap

### Exit
- [ ] `q` quits cleanly; terminal is restored

### Result (Run 2)

| | |
|---|---|
| **Pass / Fail** | **PASS** (human, partial: items below not reported) |
| **Blocking issues** | None. |
| **Notes** | Passed: startup and loading, header update badge, casks as tokens, counts match brew (2 casks, 3 outdated, 3 taps), `Tab`/`Shift+Tab`/digits, mouse, main-area layout, Info/Deps, `?` help, install and uninstall `hello`, cancel mid-run (`hunk`), failed uninstall stays with Esc, Services restart (`moshi-hook`) and start/stop (`pulseaudio`), Taps trust/untrust/untap, `q`. **Bugs found and parked as lb19:** `Tab` stuck on Search (focused input), search rows all `unknown`/never `installed`. **Decision:** `[` `]` and PgUp/PgDn are removed (lb19 AC-03); tabs switch with arrows and `h`/`l`. **Not reported (not run):** `LAZYBREW_ICONS=unicode` and light theme, `LAZYBREW_ANIMATIONS=0`, batch upgrade, `s`/`S` info toast, the stale-launch `touch` check. The `wip` → `stable` PR is not planned: work stays on `wip`. |

---

# Run 1 — M08 (history)

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
- [x] `h` / `l` (and arrows) switch tabs within panel — `TestTabSwitching`
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
- [x] Cancel long operation (Esc in operation overlay) — UI recovers, not stuck — `TestOperationCancelShowsCancelled`, `task.TestManager_CancelCurrent`
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

- [x] `d` doctor — output in modal / status path — `TestStatusKeysOpenTheirTabs`, `TestStatusDoctorKeyRefreshes`; real `brew doctor` completes
- [x] `m` missing — output lines shown — `TestStatusKeysOpenTheirTabs`, `TestMissingTabFetchesRows`, `TestDiagnosticsMissing`; real `brew missing` completes
- [x] `v` vulns — runs (or clear error if tap missing) — `TestVulnsKeyRunsOnceWhileRunning`, `TestVulnsResultShapes`, `flows.TestStatusTabsFlow`
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
