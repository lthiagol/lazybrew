# runner fix receipt r1 — v2-04-overlays-styled

Commits, one per finding, then the frame re-capture:
F-r1-01 `6b01ecb`, F-r1-02 `215aadd`, F-r1-03 `23dbb79`, F-r1-04 `40588cf`, F-r1-05 `f6807c4`, frames and result `6b35b67`.

- **F-r1-01 (major)** Each toast is its own layer at its own place (`toastPlacements` is the one layout the draw and `toastAt` read). The padded layer is gone. Test `TestToastStackLeavesCellsBesideBoxesAlone`: mixed widths at 80x24 and 140x40; every cell outside the boxes equals the base frame, and each box's rows are as drawn. A mutation restoring the padded layer fails it.
- **F-r1-02 (major)** Every confirm call site sets a consequence line. Reused: cleanup keeps `This will remove old versions.`; repair says it runs `brew tap --repair`; services cleanup says what it removes. Install, reinstall, upgrade, upgrade all and batch upgrade name what they install (`mutationDetail`). `TestConfirmKindsByCallSite` asserts a non-empty detail at all ten sites; `TestUninstallConfirmIsDanger` covers the dependency-check path.
- **F-r1-03 (minor)** The chip is the action: UNINSTALL, ZAP, UNTAP, REPAIR, INSTALL, REINSTALL, UPGRADE, UPGRADE ALL, BATCH UPGRADE, CLEANUP, AUTOREMOVE, CLEANUP SERVICES. The target stays in the question. `ConfirmModal`'s title became `action` with an `Action()` accessor. Tests assert each site's chip and that its question names the package or tap.
- **F-r1-04 (minor)** Applies Q-v2-04-wrap. Prose (question, consequence, input prompt) wraps; the chip, key line and menu items are single lines, cut with `…`. Every dialog is at most 20 rows, which fits the 22 body rows of an 80x24 frame; a body too long for that ends in `…` and the key line stays. `NewInputModal(action, prompt)` takes the action as its chip, so the prompt is prose. Chips are upper-case in every modal, where the input and menu chips were title case before. Tests: `TestConfirmQuestionWrapsAndBoxFits80x24`, `TestLongMenuItemAndChipAreCut`, and the shared-frame test.
- **F-r1-05 (minor)** `TestModalBoxKeepsItsOwnStyling`: at 140x40, 80x24 and 200x60, every modal row's border is in the danger accent and its text sits at its column unchanged.

Frames: re-captured with the unchanged harness. 07, 12, 13 and 14 changed at both sizes (`.ansi` at 140x40); the other screens are byte-identical to the last capture (`cmp`). `result.md` has a fix-round section for each changed frame.

Gate on `6b35b67`: `make build` ok; `make lint` ok; `CGO_ENABLED=1 go test ./... -race -count=1`: 683 PASS lines, 0 FAIL; `make cover-check`: brew 72.9 (≥62), presentation 96.2 (≥91), gui 79.0 (≥36), modal 86.3 (≥41); `gofmt -l` clean.

Expectation and API changes to name:
- `NewInputModal` takes `(action, prompt)`. Its test call sites gained the action argument; no assertion changed.
- `NewConfirmModal`'s second argument is the action, not the title. Test fixtures pass the action name.
- `TestConfirmKindsByCallSite` gained chip and target assertions; its earlier assertions are unchanged.

For the next session: question F-r1-04 was closed by the ruling, so nothing was appended to `work/questions.md`. Worktree clean. Not pushed. No brew write command run.
