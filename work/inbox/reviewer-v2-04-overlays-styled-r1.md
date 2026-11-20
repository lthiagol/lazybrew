verdict: changes-required

# reviewer r1 — v2-04-overlays-styled

Reviewed `3c61e42..c6cc350` (f8a8b10 AC-01, 54fe613 AC-02, aa432d6 AC-03, b8c110d AC-04, bc23c23 AC-05, c6cc350 receipt).

## Verify

- **V1 gate:** `make build` ok, `make lint` ok, `CGO_ENABLED=1 go test ./... -race -count=1` all 9 packages ok (0 FAIL), `make cover-check` brew 72.9 / presentation 96.2 / gui 78.9 / modal 86.4, all floors met. Each AC commit also builds and passes `go test ./internal/gui/...` in a scratch worktree.
- **V2 frames:** I ran the committed harness against real Homebrew into scratch, then removed `internal/gui/dump`. All 42 files match `after/` (`cmp`); all 28 `.txt` frames are exactly W×H. `before/` matches `v2-03b-sidebar-lists/after/` (md5).
- **Harness safety (untap path):** `x` on Taps → `gui.go:836` `confirmUntap` → `isOfficialTap("homebrew/cask-versions")` → warning toast, returns before setting `pendingAction` or opening a modal. `executeUntap` (the only `Untap` write) is reached only through `handleModalResult` with a confirmed `ConfirmResult` and `pendingAction=="untap"` (`commands.go:162`). The harness never presses y/enter after a confirm opens. The `DataLoadedMsg`/`TaskCompletedMsg` commands are dropped (`nm, _`). `UpdateOnStart` is false by default. No path reaches a brew write.
- **Replaced toast tests:** these are equal or stronger. `TestToastDismiss`, which checked that a fresh toast is live and that a tick after 4 s removes it, is covered by `TestToastLifetimeByType` (2.9 s live, 3.1 s expired; errors at 5.9 s and 6.1 s) and by `TestToastTickEvictsByLifetime` (a tick drops the 4 s warning and keeps the 4 s error and the fresh one). `TestToastRearmsTick` is covered by `TestToastStackOneTickChain` (a live stack re-arms, an empty stack stops) and by the kept `review_r1` `TestToastTickChainKeepsRearmingUntilExpired`. Test funcs went from 487 to 507; only those two were removed.
- **V8:** `ConfirmModal.Update` is unchanged (keys, default No). `handleModalResult` is unchanged. `internal/brew/**` is untouched. Toast texts and triggers are unchanged. No keybinding diff. There are 5 AC commits in order. The STATE row is present.

## AC table

| AC | verdict | evidence |
|---|---|---|
| AC-01 | ok | 9 `NewConfirmModal` call sites (grep), all with an explicit kind, and none inferred from the title. `mutationKind` maps uninstall/zap→danger, install→install, upgrade/upgrade-all/reinstall→upgrade. Cleanup, autoremove and services cleanup are danger, repair is install, batch is upgrade. This matches the contract and the receipt's list. |
| AC-02 | changes | The thick border in accent, the chip, the bold question, the key line, the reverse selection, the 72 width (minTerminalWidth=80 makes `min(W-8,72)`=72) and the input/menu frame are all there. The consequence line is missing on most confirms (F-r1-02). Chip text: F-r1-03. Wrap vs cut: F-r1-04. |
| AC-03 | ok | `dimBackdrop` in `render.go` runs at compose time. The test compares each row with and without the modal: the text is equal and the muted SGR is present. The footer is not muted. In `07-confirm-modal.ansi` the base is muted and the box and footer are not. |
| AC-04 | changes | Stack, eviction, 3 s/6 s lifetime, click dismiss, a single tick chain, and no goroutines are all in place and tested. The stack layer erases frame cells outside the toast boxes (F-r1-01). |
| AC-05 | changes | The harness, screens 12–14 and `result.md` are present. `result.md` does not explain the erased cells in frame 13 (F-r1-01). |

## Findings

**F-r1-01 (major)** — `internal/gui/toasts.go:84`
- **What:** `toastLayer` builds the stack with `lipgloss.JoinVertical(lipgloss.Right, …)`. That pads each narrower box with spaces to the width of the widest box. The layer is opaque, so the padding blanks frame cells to the left of each narrower toast.
- **Evidence:** `after/80x24/13-toast-stack.txt` row 21: the Log text `No commands executed ye` becomes `No commands        `. Row 22: the Log bottom border is broken (`╰────────────       ╰──…`). At 140x40, rows 31–39 blank part of the main panel's bottom border and part of the Log title and bottom border. This breaks AC-04.4 (only a toast's own bordered box may cover the frame), and `result.md` does not explain it.
- **Fix must achieve:** only the cells of each toast box change. Every cell outside the union of the boxes equals the base frame, with a test that uses a stack of mixed widths at 80x24 and 140x40. `toastAt` must stay consistent with the new layout. The frames are re-captured.

**F-r1-02 (major)** — `internal/gui/modal/confirm.go:101` and the call sites
- **What:** AC-02.3 asks for one subtle line under every confirm question saying what will happen. Only uninstall, zap and untap set `WithDetail`. Upgrade, upgrade all, batch upgrade, reinstall, install, repair, cleanup, autoremove and services cleanup show none.
- **Evidence:** frame `12-confirm-upgrade` has only `Upgrade bash?` and the key line. The receipt does not name the gap.
- **Fix must achieve:** every confirm call site gives a consequence line; existing text such as cleanup's "This will remove old versions." may be reused. A test covers each call site, for example as an extension of `TestConfirmKindsByCallSite`.

**F-r1-03 (minor)** — `internal/gui/modal/confirm.go` (`View`, the chip label)
- **What:** AC-02.2 asks for the action's name upper-cased as the chip (` UNINSTALL `). The chip is the whole title upper-cased instead: ` UNINSTALL BASH `, ` CONFIRM UPGRADE `, ` CONFIRM BATCH UPGRADE `. "CONFIRM" carries no action name.
- **Fix must achieve:** the chip names the action (UNINSTALL, UPGRADE, UNTAP, CLEANUP…), and the package or tap name stays in the question.

**F-r1-04 (minor)** — `internal/gui/modal/frame.go:31`
- **What:** AC-02.5 says every line is cut to the inner width. The body is wrapped instead. The receipt discloses this, but nothing about it was written to `work/questions.md`. The contradiction should be recorded for the orchestrator, not just decided by the lane.
- **Fix must achieve:** either cut with `…` as written, or append the question to `work/questions.md` and keep the wrap until the orchestrator answers.

**F-r1-05 (minor)** — `internal/gui/backdrop_test.go`
- **What:** the AC-03 test checks the footer but never checks that the modal box itself is not muted. The frame shows it is not muted, but nothing locks that.
- **Fix must achieve:** an assertion that the cells inside the modal box carry the modal's own styling (for example the accent border SGR) and not the muted role.

## ANSI read-through (140x40)

Frames 07, 12 and 14 read well: the dim is clear, and the box and footer stand out. Frame 13 reads as broken because of the blanked border and text cells (F-r1-01). Frame 10 is acceptable: the toast's border sits over the Log bottom border row, and `result.md` explains it.
