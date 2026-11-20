verdict: changes-required

# reviewer r3 — v2-04-overlays-styled

Reviewed `60763e4..da63856`: fix f4b4127 (F-r2-01), frames 35e7090, receipt da63856.

## Verify

- **V1 gate:** `make build` ok, `make lint` ok, `CGO_ENABLED=1 go test ./... -race -count=1` 9/9 packages ok, `make cover-check` 72.9/96.2/79.0/87.7, floors met.
- **V2 frames:** I ran the committed harness against real Homebrew into scratch and removed `internal/gui/dump` afterwards. 42/42 files match `after/` (`cmp`). All 28 `.txt` frames are exactly W×H.
- **V8:** only `internal/gui/modal/{confirm,frame,input,menu}.go` and a test changed. Test funcs went from 511 to 512, with none removed. `internal/brew/**` is untouched. No keybinding or semantics change. One fix commit plus one frames commit, as the r2 brief asks. No brew write command was run.

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r1-01..05 | still fixed | The fix touches only the dialog layout. The frames for 13 are byte-identical to r2, where my cell-by-cell check found 0 differing cells outside the boxes. |
| F-r2-01 | fixed | I re-ran the scratch probe (removed afterwards). `input.View width=33` (was 43), and every input row is ≤ 42 inner cells. The chip `ADD TAP` sits 20 cells left and 20 right, the prompt 10/11, and the key line 11/12. The field block starts at column 4 of 42, with 33 cells and 5 cells after it. That holds with a 55-character value too, which scrolls inside the field. The menu is still centred as one block. `TestDialogsFitAndCentreTheirRows` covers confirm, input and menu, plain and long. Frame 14 is centred at both sizes. |

## Findings

**F-r3-01 (minor)** — `internal/gui/modal/frame.go` (`dialog`, the per-row `PlaceHorizontal`), `internal/gui/modal/confirm.go` (body through `proseLines`). New in f4b4127.
- **What:** every row is now centred on its own, so the uninstall/zap dependents list (`The following depend on bash:` followed by one `  name` per line) loses its alignment. Each dependent name is centred separately, and the list reads ragged. The two-space indent in the message no longer means anything.
- **Evidence:** in `after/80x24/07-confirm-modal.txt`, `direnv` moved from column 23 to column 33, 10 cells right. It sits under the middle of the heading instead of indented below it. With three dependents, the probe gives `direnv` at column 33, `zsh-completions-long-name` at 23 and `git` at 34.
- **`result.md` is wrong here:** the r2 section says rows in 07 and 12 move "one cell further left or right … the content is unchanged". That is not true for the dependents row. The receipt makes the same claim.
- **Fix must achieve:**
  - The confirm body that follows the question (the dependents heading and list, and the timeout note) keeps its lines aligned as one block, the way the menu pads its items. Chip, question, detail and key line stay centred per row.
  - Add a test: an uninstall confirm with dependents of different lengths draws every dependent name at the same start column.
  - Correct the `result.md` line for 07 and re-capture 07.

## ANSI read-through (140x40)

Frames 12 and 14 read well; 14 is now centred. 13 is clean. 07 reads well apart from the lone centred `direnv` (F-r3-01).
