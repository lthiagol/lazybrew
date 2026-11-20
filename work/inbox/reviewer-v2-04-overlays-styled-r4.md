verdict: ok

# reviewer r4 — v2-04-overlays-styled

Reviewed `a3dd443..1835bae`: fix 64bc9d8 (F-r3-01), frames 222ae13 (07 re-captured, 15 added), receipt 1835bae.

## Verify

- **V1 gate:** `make build` ok, `make lint` ok, `CGO_ENABLED=1 go test ./... -race -count=1` 9/9 packages ok, `make cover-check` 72.9/96.2/79.0/88.3, floors met.
- **V2 frames:** I ran the committed harness (now with step 15) against real Homebrew into scratch and removed `internal/gui/dump` afterwards. 45/45 committed files match (`cmp`). All `.txt` frames are exactly W×H.
- **Harness step 15 is safe:** `1` then `B` on Status calls `brewfileMenu`, which only opens a `MenuModal`. Esc returns a cancelled `MenuResult`, and `handleModalResult` runs nothing on a cancelled result (`commands.go`, the `MenuResult` case). No Enter or digit is sent while the menu is open. `executeBrewfileAction` is unreachable from the harness.
- **V8:** only `modal/confirm.go`, `modal/frame.go` and a test changed. Test funcs went from 512 to 513, with none removed. `internal/brew/**` is untouched. No semantics or keybinding change. One fix commit plus one frames commit. No brew write command was run.

## Probe: every dialog (scratch test, removed afterwards)

All 72-wide confirms and 50-wide forms have every row inside the box.

| dialog | result |
|---|---|
| confirm with 3 dependents of different lengths | `direnv`, `zsh-completions-long-name` and `git` all start at column 22, two cells in from `The following depend on bash:` (column 20), as one centred block. |
| dependents plus timeout note | The heading, the list and the wrapped note form one left-aligned block starting at column 3; the indents are kept. |
| timeout note only | The note wraps as a left-aligned block, within the box. |
| confirm without dependents (upgrade) | Chip, question, consequence and key line are each centred. |
| 20 dependents | The body is cut to fit `dialogMaxRows` and ends in `…`. The box is 20 rows, so it fits 80×24. |
| input | `ADD TAP` sits 20/20, the prompt 10/11, and the 33-cell field 4/5. |
| menu (Brewfile) | The items form one left-aligned block; item 1 is cut with `…`. |

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r1-01..05 | still fixed | Frames 12, 13 and 14 are byte-identical to r3. My r2 cell-by-cell check of frame 13 still holds. |
| F-r2-01 | still fixed | Input probe above. |
| F-r3-01 | fixed | `leftBlock` pads the body to one width, and `TestDependentsAlignAsOneBlock` locks it. `after/80x24/07-confirm-modal.txt` shows `direnv` at column 22 under its heading at column 20. `result.md` corrects the r2 claim and adds lines for 07 and 15. |

## AC table (final)

| AC | verdict |
|---|---|
| AC-01 | ok: every confirm call site has a kind; see r1. |
| AC-02 | ok: the confirm face, the consequence line at every site, the action chip, and wrap-or-cut per Q-v2-04-wrap. |
| AC-03 | ok: the backdrop is dimmed at compose time; the modal and the footer are not. |
| AC-04 | ok: the stack, eviction, 3 s/6 s lifetimes and click dismiss; only the toast boxes cover the frame. |
| AC-05 | ok: frames 01–15 and `result.md` are consistent with the harness output. |

## Findings

None.

## ANSI read-through (140x40)

07, 12, 13, 14 and 15 read aligned and clear. 15 is muted behind the menu, and the toasts from 13 are still visible because the harness does not drain ticks, as `result.md` notes.

One observation, not a finding and possibly for the backlog: when a dependents list is cut, the closing `…` row is centred on its own rather than placed in the list's column. This comes from the r1 `dialogMaxRows` cut in `dialog()` and only shows with more than about 12 body rows.
