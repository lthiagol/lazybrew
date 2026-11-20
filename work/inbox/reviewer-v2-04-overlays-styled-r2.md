verdict: changes-required

# reviewer r2 — v2-04-overlays-styled

Reviewed `2a54b4c..c27a7aa`: fix commits 6b01ecb (F-r1-01), 215aadd (F-r1-02), 23dbb79 (F-r1-03), 40588cf (F-r1-04), f6807c4 (F-r1-05), frames 6b35b67, receipt c27a7aa. AC-02.5 is read under the ruling Q-v2-04-wrap.

## Verify

- **V1 gate:** `make build` ok, `make lint` ok, `CGO_ENABLED=1 go test ./... -race -count=1` 9/9 packages ok, `make cover-check` 72.9/96.2/79.0/86.3, floors met. In a scratch worktree, each fix commit builds and passes `go test ./internal/gui/...`.
- **V2 frames:** I ran the committed harness against real Homebrew into scratch and removed `internal/gui/dump` afterwards. 42/42 files match `after/` (`cmp`).
- **Frame 13, cell by cell:** a scratch copy of the harness takes a second shot of the same state after a 6.5 s wait and a `ToastTickMsg`, so every toast has aged out. I found the three boxes by their corners and compared every cell outside them. At 80x24 (boxes at rows 14/17/20, cols 52/41/48), 0 cells differ. At 140x40 (rows 30/33/36, cols 112/101/108), 0 cells differ. The Log text `No commands execute` and the border runs beside the narrower boxes are back.
- **V8:** `handleModalResult` and `ConfirmModal.Update` are unchanged. Test funcs went from 507 to 511, with none removed. The expectation changes (the action argument on `NewConfirmModal`/`NewInputModal`) are named in the receipt. `internal/brew/**` is untouched. No keybinding diff. There is one commit per finding, in order. No brew write command was run.

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r1-01 | fixed | `toastPlacements` and `toastLayers` give each box its own layer, and `toastAt` reads the same placements. `TestToastStackLeavesCellsBesideBoxesAlone` covers mixed widths at 80x24 and 140x40. The independent frame-13 check above found 0 differing cells. |
| F-r1-02 | fixed | All 10 confirm sites plus uninstall and zap set a detail line. `TestConfirmKindsByCallSite` asserts `Detail() != ""` at every site. Frame 12 shows `Installs the newest version of bash.` |
| F-r1-03 | fixed | The chips are UNINSTALL, ZAP, UNTAP, REPAIR, INSTALL, REINSTALL, UPGRADE, UPGRADE ALL, BATCH UPGRADE, CLEANUP, AUTOREMOVE and CLEANUP SERVICES. The test asserts `Action()` and that the question names the target. Frames 07 and 12 show this. |
| F-r1-04 | fixed (per ruling) | Prose wraps through `wrapProse`; the chip, key line and menu items are cut through `fitLine`; `dialogMaxRows`=20. `TestConfirmQuestionWrapsAndBoxFits80x24` and `TestLongMenuItemAndChipAreCut` cover it. See F-r2-01 for a side effect. |
| F-r1-05 | fixed | `TestModalBoxKeepsItsOwnStyling` checks every box row for the danger-accent border and the unchanged text, at three sizes. |

## Findings

**F-r2-01 (minor)** — `internal/gui/modal/input.go` (`View`), `internal/gui/modal/frame.go` (`dialog`). New in 40588cf.
- **What:** `dialog` no longer wraps or cuts every body part, and the textinput line is 43 cells wide (`SetWidth(40)`, plus the prompt and the cursor). The box's inner width is 42. Because the centred block is wider than the box, the input modal's content is off-centre.
- **Evidence:** a scratch probe (since removed) shows `input.View width=43`. The rendered rows are `┃                             ADD TAP            ┃` (29 cells left, 12 right), `┃              Tap repository (user/repo):       ┃` and `┃                enter submit   esc cancel       ┃`. Frame `after/80x24/14-input-modal.txt` rows 10–16 show the same, and so does the 140x40 frame. In r1, frame 14 was centred. `result.md` does not mention the shift. The menu modal is not affected; it centres correctly.
- **Fix must achieve:** every line of the input modal fits the inner width, for example by fitting the textinput to `innerWidth(formWidth)` including its prompt and cursor, or by cutting it with `fitLine`. The chip, prompt and key line are centred again. Add a test that asserts every content row of the input modal is at most the inner width and the chip is centred (left and right gaps within 1). Re-capture frame 14.

## ANSI read-through (140x40)

Frames 07, 12 and 13 read well. In 13, the stack sits cleanly over the Log box and the main panel, with the cells beside it intact. In 14, the chip and the key line visibly lean right (F-r2-01).
