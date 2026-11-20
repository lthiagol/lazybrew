verdict: changes-required

# Review r2 — slice v2-06-animations (wip 4e3bde8..b9fa4d9)

**V1 gate (CGO_ENABLED=1, HEAD b9fa4d9):** build, lint, `go test ./... -race -count=1` (12 pkgs ok), cover-check (brew 76.5, presentation 98.2, gui 84.1, modal 90.1): exit 0.
**Every fix commit passes:** `go vet` + `go test ./... -count=1` in a worktree at ee7d782, 789caf6, 7a66ae2, a000af9: all four exit 0, no failing test lines.
**V2/V4 frames:** harness re-run against real brew: `diff -r` vs `v2-05s-services/after/` empty. Motion dump re-run: `diff -r` vs `motion/` empty (fix changed only the 5 `motion-modal-*.ansi`). 55 frames exactly W×H. moshi-hook MainPID=800, NRestarts=0, ActiveEnter 2026-11-18 11:37:48 before and after. `internal/gui/dump` removed; worktree clean.

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r1-01 (major) | fixed | `testutil.NewModel` = `gui.New(...).WithoutMotion()`; both helpers use it, so motion is off whatever the env. `TestHelpersBuildStillModels` (testutil) and `TestFlowModelsAreStill` (flows) set `LAZYBREW_ANIMATIONS=1` and fail if the model animates. `internal/app` has a `TestMain`: HOME → empty temp dir, animations off; `TestAppModelIsStill`. |
| F-r1-02 (minor) | fixed | Off: `Update` returns `m.update(msg)` directly. Re-measured `BenchmarkUpdateUnhandled` (now in repo) ×3: off **2 allocs, 3.0–5.0 µs, 21.8 KB** (314dd3c: 2 allocs, 2.4 µs, 19.5 KB); on **20 allocs, 10–13 µs, 44.5 KB** (was 178). pprof (on): `motionTargets` no longer on top; `Set.Clone`+`NewSet` are 46 % of allocs. `tabBarRows` = 2 and `tabWidth` = name+4 without a render, held by tests; I also checked `tabWidth` against `TabActive` (all panelTabs): equal. |
| F-r1-03 (minor) | fixed | `result.md` § "Commit history deviation" names 7c79a51, the failing test and message, why, and 0ebc253. |
| F-r1-04 (minor) | partly | Text runs no longer appear in frame 01 (`205;214;244` 51 → 0). The fade now runs subtle→muted on a 0..100 scale; `TestModalFadeSpringIsMonotone` shows no snap from ≈0.8. Border cells still brighten: see F-r2-01. |

## Staleness of the target caching

Only the toast slot layout is cached (`motionStack`, re-laid when the stack pointers or the window change). Tab, sidebar, modal and progress targets are still recomputed on every update, so they cannot go stale. Toasts are immutable after `NewToast`; slots depend only on the stack, `m.width`/`m.height` (`bottomBarRows` is const) and the icon set (fixed per model). Tabs are static per panel (`panelTabs`); no data load changes a tab's width. Sidebar total is `bodyH` (height only), so log collapse (`L`) cannot skew it.
Throwaway test (removed): 8 sequences at 140×40 and 80×24 — tab switch then panel switch mid-slide, rapid `[`/`]`, focus change then resize mid-flight, `L` toggles mid-focus, toasts added across a resize then evicted, same stack across two resizes, modal open/esc/reopen then resize, modal + toast + resize. Every step: boxes sum to the sidebar, frame W×H. After settle: frame equals the same model drawn with motion off, heights and every panel's `visibleRows` equal the resting ones. All passed. Sanity: snapping `tab:x` and `sidebar:0` off-target makes the same check fail, so the check is not vacuous. I found no stale sequence.

## New findings

**F-r2-01 (minor, visual) — internal/gui/render.go:199-213** — The first fade frame is still brighter than the frame it replaces, for the borders. In the frame behind the prompt (`after/140x40/06-taps.ansi`) the inactive borders are `38;2;69;71;90` (68 runs). `motion-modal-01.ansi` paints the whole backdrop in the subtle role, `166;173;200`, so those border cells jump up about 100 points of brightness. They then drop to `127;132;156` (frame 05) and rest at `108;112;134`. The flash moved from the text to the borders. `TestBackdropFadeNeverBrightens` compares the fade steps with each other, starting at subtle. It never compares step 1 with the cell's own colour in the replaced frame, which is what the fix brief asks for ("no brighter than the frame it replaces"). With one colour for the whole backdrop, any start brighter than the muted role brightens the borders. Fix must achieve: no backdrop cell is brighter in fade frame 1 than in the frame it replaces, plus a test on a real backdrop cell that was a border. Two ways: blend each cell from its own colour to muted, or the orchestrator accepts the subtle start as the brief's allowed option and closes this. Resting frame unchanged either way.

## Notes (not findings)

- `TestBackdropFadeNeverBrightens` checks luminance order only for the theme loaded at test time (dark). In Latte, text < subtle < muted, so the premise check would `t.Fatal` there. The comment's "both themes" claim is not tested.
- `TestHelpersBuildStillModels` makes the same call twice; the second message says "with a runner" but calls the same thing. Harmless.
- `TestAppModelIsStill` asserts with the env set to 0, the same value TestMain sets. HOME isolation is real (TestMain) but untested. Fine for the brief.
- On: `Set.Clone` now dominates (20 allocs). A copy-on-write or value-slice Set would remove it; not required.
