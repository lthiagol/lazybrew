verdict: ok

# Review r3 — slice v2-06-animations (wip 7f4e562..2c38479)

**V1 gate (CGO_ENABLED=1, HEAD 2c38479):** build, lint, `go test ./... -race -count=1` (all pkgs ok), cover-check (brew 76.5, presentation 98.2, gui 84.0, modal 90.1): exit 0. The only code commit is 2a135d8; the HEAD after it adds only the receipt, so the gate at HEAD covers it.
**V2/V4 frames:** harness re-run against real brew: `diff -r` vs `v2-05s-services/after/` empty. Motion dump re-run: `diff -r` vs committed `motion/` empty. 55 frames exactly W×H. moshi-hook MainPID=800, NRestarts=0, ActiveEnter 2026-11-18 11:37:48 before and after. `internal/gui/dump` removed; worktree clean.

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r1-01..03 | fixed (r2) | unchanged by 2a135d8 |
| F-r1-04 / F-r2-01 (minor) | fixed (fallback) | No fade now: `render` calls `dimBackdrop` on the modal's first frame; the 3-row drop is kept. `motion-modal-01..05.ansi` re-dumped. Every frame's backdrop is `38;2;108;112;134` (muted), the same as resting `after/140x40/14-input-modal.ansi`. Rows 1–6 (backdrop above the modal) of modal-01 are byte-identical to frame 14. No backdrop cell ever passes through a third colour. |

## The fallback's reasoning

It holds, and it holds in both themes, not only Mocha:
- **Mocha:** inactive border Surface1 `#45475a` is darker than muted Overlay0 `#6c7086`. The resting dim (v2-04, unchanged) already brightens border cells. Any blend from a cell's own colour to muted brightens them step by step.
- **Latte:** text `#4c4f69` is darker than muted `#9ca0b0` (and border `#bcc0cc` is lighter). Text cells lighten on the way to muted.

The rest is one colour that sits between the frame's extremes, so in either theme some cells must move against "never brighter than the frame it replaces". That criterion cannot be met while resting frames stay unchanged. Caveat for the record: a per-cell blend would still have been monotone per cell, with no overshoot flash. So the fallback trades the fade for simplicity; it was not forced by a flash. The fix brief permits this, and the orchestrator accepts it.

## Dead code / tests

- `grep` for `modalFade|backdropFade|fadeBackdrop|backdropColor|Blend1D|luma(` over `*.go`: no hits. `dimBackdrop` is byte-identical to 314dd3c's. `modalDrop`/`modalDropRows` are the only modal motion left, and both are used. No orphan imports (build and vet clean). The README makes no fade claim.
- Removed: `TestBackdropFadesIntoTheMutedRole`, `TestBackdropFadeNeverBrightens`, `TestModalFadeSpringIsMonotone`. All three were added in this slice and tested the removed code. `TestBackdropIsDimAtOnce` replaces them. It runs under `DarkTheme` and `LightTheme`, with `defer` restoring dark; no `t.Parallel` in package gui, so the global theme swap is safe. It passes. Not a weakening of any pre-slice test.

## Notes (not findings)

- **Contract deviation, authorized:** slice AC-04.1 asked that the backdrop dim "fades in"; it now dims at once. `result.md` § "Fix round r2" records why; the orchestrator should carry this into the STATE row/verdict.
- `TestBackdropIsDimAtOnce` compares only frame row 0, and without a fade it is close to tautological. Comparing every backdrop row outside the modal box would be stronger. Optional.
- The r2 report double commit (645ee3a / 7f4e562) is noted by the orchestrator; not a finding.

No findings left.
