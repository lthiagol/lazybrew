verdict: changes-required

# Reviewer — v2-08-preload, round r1

Range: `2a7e2a5..a8a1c38` (v2-08 commits 1471c38, 2237be9, fd55e63, receipt
06c198c/d23bb9a/c5b37c9, recapture a8a1c38). Brew ran at normal speed during this
run (`brew tap-info` 0.88 s, `brew outdated --formula` 0.32 s). No brew write ran.

## Verify

| Check | Result | Evidence |
|---|---|---|
| V1 gate | pass | `make build` ok; `make lint` (go vet) ok; `CGO_ENABLED=1 go test ./... -race -count=1` 11 packages `ok`; `make cover-check` "All coverage floors met" (gui 85.1%) |
| V2 frames | pass | harness copied to `internal/gui/dump`, real brew, `ok` in 27.8 s, dir removed. 40/40 frames exactly W×H (cell width). Output byte-identical to committed `after/`. `before/` identical to `v2-07-op-dismiss/after/`. Main-area diffs only on 01, 05, 15 (140x40) and 01, 05 (80x24) plus header row 0 (`1 outdated`), all listed in result.md |
| V3 AC-01 | met (see F-r1-01) | `preload.go`: Status first (started on first Status `DataLoadedMsg`), `preloadOrder` F→O→C→T→S, `preloadWidth`=2, reads via `fetchPanelData`→`boundedRead`/`brewReadTimeout`, no goroutine/`program.Send`. Cap tested in `TestPreloadOrderAndTwoInFlight` (pending ≤ 2 after each landing). Queued switch is promoted, not re-read (`TestPreloadDedupesWithAUserSwitch`). In-flight switch: `loading` set, so `lazyLoadPanel` returns nil. `update_on_start`: Init sends no Status; Status comes from the RefreshMsg after `brew-update` completes (`TestPreloadWaitsForUpdateOnStart`). Preload off: `TestPreloadOffStaysLazy`. Services fetch reads list + info |
| V4 AC-02 | met | `TestPreloadKeysWorkMidPreload` (View asserts), `TestPreloadMouseWorksMidPreload`, `TestPreloadFailedReadShowsItsErrorOnItsPanelOnly`, `TestPreloadFailedReadDoesNotStopTheOthers` |
| V5 AC-03 | met | spinner while queued/in flight (`TestPreloadSidebarSaysLoadingNotNotLoaded`); counts after; `not loaded — press R` only with preload off or a failed read (`TestPreloadOffDashboardSaysPressR`). Spinner instead of Subtle text is allowed by the brief and noted in result.md |
| V6 invariants | met | no new brew command; test helpers changed only `config.Default()`→`testConfig()` (preload off, reason in result.md), harness sets preload on, `TestProductionDefaultPreloads` + `TestPreloadDefaultIsOn` pin the default; no test deleted/skipped; one commit per AC (a8a1c38 is a frames-only recapture after the receipt) |

## AC table

| AC | Commit | Verdict |
|---|---|---|
| AC-01 | 1471c38 | met |
| AC-02 | 2237be9 | met, except edge F-r1-01 |
| AC-03 | fd55e63 (+ a8a1c38 recapture) | met; records stale, F-r1-02 |

## Findings

F-r1-01 (minor) — internal/gui/gui.go:599, internal/gui/preload.go:81 — with preload
on, a panel the user opens before Status lands is read by `lazyLoadPanel` as a plain
`DataLoadedMsg`, and its count is never absorbed into the dashboard. For Outdated the
dashboard then reads `not loaded — press R` after the panel has loaded. Evidence: a
temporary test (removed after) pressed `4` before `statusLoaded()`, landed the lazy read
and the preload: `status.Outdated.Known=false`, `panels[Outdated].loaded=true`, dashboard
`Outdated:  not loaded — press R`. The real Status read only reads outdated from the
cache, which is still empty at that point. — Fix must achieve: with preload on, every
successful panel landing (lazy or preload) fills the dashboard count, so `press R` never
shows for a panel that has loaded. Add a test for the early-switch case.

F-r1-02 (minor) — work/STATE.md:296-298, work/inbox/runner-v2-08-preload.md:28-36 —
the resume record and the receipt still describe the first capture: "Brew is slow …
tap-info ~87 s", Taps frames 06/14 degraded, "sidebar changed on 16 screens", main area
changed on 06 and 14. After a8a1c38 the committed frames show full Taps data and
result.md lists 18 screens and main changes on 01, 05, 15. — Fix must achieve: the STATE
paragraph matches the committed frames (degraded-Taps text removed or marked as the
superseded first capture); Q-v2-08-b stays open as a question.
