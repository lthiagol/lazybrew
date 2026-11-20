# runner fix receipt r2 — v2-04-overlays-styled

- **F-r2-01 (minor)** commit `f4b4127`. The input field was 43 cells in a 42-cell box. It is now a fixed 30 cells of text (33 with prompt and cursor), so the dialog centres it. Wrapped prose is styled one line at a time: styling a paragraph padded every line to the longest. Each row is centred on its own, and the menu pads its items to one width so the list stays one block. `TestDialogsFitAndCentreTheirRows` covers the confirm, input and menu dialogs, plain and long: every content row fits the inner width, the chip is centred within one cell, the block is centred, and the input field sits at its exact column. A mutation restoring the 40-cell field fails it.
- **Frames** commit `35e7090`. 14 changed at both sizes (the field and its centring). 07 and 12 changed by one cell of centring, since each row is now centred on its own; the content is unchanged. The other screens are byte-identical. `result.md` has a fix-round r2 section.

Gate on `35e7090`: `make build` ok; `make lint` ok; `CGO_ENABLED=1 go test ./... -race -count=1`: 690 PASS lines, 0 FAIL, 9 packages ok; `make cover-check`: brew 72.9 (≥62), presentation 96.2 (≥91), gui 79.0 (≥36), modal 87.7 (≥41); `gofmt -l` clean.

Notes for the next session:
- No API or expectation change in this round. The input field is a constant 30 cells, not `innerWidth - 3`.
- Worktree clean. Not pushed. No brew write command run.
