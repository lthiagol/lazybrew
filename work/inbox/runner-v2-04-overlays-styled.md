# runner receipt — v2-04-overlays-styled

Commits, in AC order: AC-01 `f8a8b10`, AC-02 `54fe613`, AC-03 `aa432d6`, AC-04 `b8c110d`, AC-05 `bc23c23` (frames, result.md, STATE row).

- **AC-01** Confirm call sites and kinds. Danger: uninstall (`x`, via DepCheckMsg), zap (`X`), untap (`x` on Taps), cleanup and autoremove previews (gui.go), services cleanup (`c` on Services). Upgrade: upgrade (`u`), upgrade all (`U`), reinstall (`r`), batch upgrade. Install: install (`i`/`I`), repair (`r` on Taps). Neutral: none. Not confirms: `S` (services stop) runs with no dialog, and tap add is an input modal; neither has a kind. Commit `f8a8b10` says "tap-add-style prompts install", which is loose: tap add opens no confirm. Evidence: `TestConfirmKindsByCallSite` (10 cases), `TestUninstallConfirmIsDanger`, `TestActionKindAccent`.
- **AC-02** Face: thick border in the kind's accent, chip, bold question, subtle detail, key line with reverse-video selection. Box is 72 wide, which is min(frame-8, 72) because no frame is drawn under 80. Body text wraps rather than cuts, so the untap warning survives. Confirm chip is upper-cased; input and menu chips keep the title as written (the brief names only the confirm chip). Detail: uninstall/zap `removes X; N packages depend on it` (timeout: `dependents unknown`), untap the existing warning. Evidence: `TestConfirmFace*`, `TestInputAndMenuShareTheFrame`, `TestRemovalDetail`, `TestUntapConfirmShowsWarning`; frames 07, 12, 14.
- **AC-03** `dimBackdrop` in render.go mutes every row but the footer, at compose time. The toast layer sits above the dimmed base and is not muted. `charmbracelet/x/ansi` moved to a direct require (no new module; `go mod tidy -diff` clean). Evidence: `TestDimBackdropMutesEveryRowButTheFooter`, `TestDimBackdropKeepsText`; `.ansi` has 53 muted runs in 07, 0 in `before/`.
- **AC-04** Stack of up to 3, newest lowest, bottom-right, bottom-aligned on the last body row. Fourth evicts oldest. One tick chain (`toastTicking`). Lifetime 3 s, errors 6 s. x/esc do not dismiss; a left click does. Info toasts draw in the Info role (blue), not the old accent. Evidence: `toast_stack_test.go` (10 tests), `toast_float_test.go`, `modal/modal_test.go` lifetime and ellipsis tests; frame 13.
- **AC-05** Harness copied; `before/` = v2-03b `after/`; screens 12–14 added. 01–06, 08, 09, 11 are identical in text, and in ANSI at 140x40 (`cmp`). Details in `result.md`.

Expectation changes (named; none deleted, skipped or weakened):
- `modal_float_test.go`: expected overlay no longer wrapped in `ActiveBorder` (outer wrap removed).
- `toast_float_test.go`: rewritten from one text row to measuring the bordered box; same cell-isolation check.
- `modal/modal_test.go`: `TestToastDismiss` and `TestToastRearmsTick` replaced by lifetime and ellipsis tests (dismissal moved to the GUI stack). `NewConfirmModal` calls gained the kind argument.
- `gui_test`, `frame_size_test`, `review_r1..r4_test`: `m.toast` became `m.newestToast()` / `m.toasts`. Mechanical; assertions unchanged.

Gate on `bc23c23`: `make build` ok; `make lint` ok; `CGO_ENABLED=1 go test ./... -race -count=1`: 674 PASS lines, 0 FAIL; `make cover-check`: brew 72.9 (≥62), presentation 96.2 (≥91), gui 78.9 (≥36), modal 86.4 (≥41); `gofmt -l` clean.

For the next session:
- Frame 13's warning toast needs an official tap. This machine has none (`0 official`), so the harness injects `homebrew/cask-versions` into Taps, then presses `x`. The frame shows success, error and warning; no info toast, because the info producers are not reachable without brew writes.
- The `Install bash completed` toast from 10 stays up in 12 and 14, since the harness does not drain ticks (same as v2-03b).
- No question appended to `work/questions.md`. Not pushed. No brew write command run. Worktree clean.
