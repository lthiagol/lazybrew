verdict: changes-required

# reviewer r2 — v2-03b-sidebar-lists (reviewed `78ae6e6..78452de`: fix commits `01fb663`..`752583b`, frames `008d344`)

## Verify

- **Gate** (HEAD `78452de`): `make build` ok, `make lint` ok, `CGO_ENABLED=1 go test ./... -race -count=1` every package ok. `make cover-check`: brew 72.9, presentation 96.2, gui 77.9, modal 50.5, all floors met. Each fix commit (`01fb663`, `9ba0f80`, `d68d66f`, `3c38d91`, `8d2c171`, `752583b`) also passes `go build ./...` and `go test ./internal/gui/...` in a detached worktree.
- **Frames**: re-ran the committed harness against real Homebrew. All `.txt` frames match `after/` at both sizes, all 140×40 `.ansi` frames match byte-for-byte, and every frame is exactly W×H. `internal/gui/dump` removed afterwards. The fix-round section of `result.md` accounts for every changed row: the Homebrew row, the 06/09 footer tail, and the 08 help rows.
- **Invariants**: no change to `internal/brew/**` or `internal/config/**`. No test removed; the top-level `func Test` count went 479 → 484. One commit per finding, each subject naming its id.

## Earlier findings and orchestrator items

| id | status | evidence |
|---|---|---|
| F-orch-01 | fixed | `render.go` passes the batch only for `PanelOutdated`. `TestBatchMarksOnlyTheOutdatedList` checks the Outdated box shows `●` and the Formulae box at the same indices does not. |
| F-orch-02 | fixed (F-r2-02 minor) | `tapTrustStory` (`details.go:85`) shows the state, its meaning for the tap's formulae and casks, the counts, and the `t` hint. Tap Info keeps `tapDetails`. `TestTapTrustTabShowsTheTrustStory` covers untrusted, trusted, and Tap Info. |
| F-orch-03 | **not met** (F-r2-01) | PgUp/PgDn/Home/End/g/G work and are clamped at both ends (`TestScrollKeysMoveAndClamp` asserts offsets and drawn rows; `TestScrollIsPerItem`). Help lists them, and `globalHints` gains `PgUp/Dn scroll`. However, `g`/`G` were not unused; see F-r2-01. |
| F-r1-01 | fixed | The fix receipt states that `0fa39ff` is AC-01. |
| F-r1-02 | fixed | `status.go:35` uses `%-10s`. Frames show `Homebrew  7.0.15-…` in the label column. The test expects `Homebrew  4.5.2`. |
| F-r1-03 | fixed | `gui_test.go:1336` now names Info. The `panel.go:74` comment describes `unfilteredFormulae`. |

## Findings

- **F-r2-01 (major) — `internal/gui/gui.go:793-796`, `g`/`G` taken from type-to-search.**
  - **What:** the brief allows `g`/`G` "only if unbound". They were not missing from every path: when the Search panel is focused and its input is not, any printable key that no case catches falls through to the type-to-search block (`gui.go`, `m.activePanel == PanelSearch && !m.searchInput.Focused()`). That block focuses the input and types the key. Before this round, `g`/`G` took that path. Now the new `case "home", "g"` / `case "end", "G"` return first, so `g`/`G` only scroll the content. A user who presses Esc on Search and types `git` or `gh` loses the first letter. This is a remapped key, which the contract invariant forbids.
  - **Evidence:** a throwaway test (deleted after the run). After pressing `7` then `Esc`, sending `g`, `G` and `z`:
    - at `78ae6e6`: `g` → focused=true value="g"; `G` → focused=true value="G"; `z` → focused=true value="z".
    - at HEAD: `g` → focused=false value=""; `G` → focused=false value=""; `z` → focused=true value="z".
  - **Fix must achieve:**
    - On the Search panel with the input unfocused, `g`/`G` start a search with that letter again.
    - Either keep `g`/`G` for scrolling only outside Search, or drop them and keep Home/End.
    - Add a test for each of `g` and `G` on the unfocused Search panel.
    - If `g`/`G` are dropped, update the help row `Home/End, g/G` and `TestScrollKeysMoveAndClamp` to match.
- **F-r2-02 (minor) — `internal/gui/details.go`, `tapTrustStory`.**
  - **What:** `"Its %d formulae and %d casks follow that state."` reads `Its 2 formulae and 1 casks` (asserted verbatim in `TestTapTrustTabShowsTheTrustStory`).
  - **Fix must achieve:** a singular count reads `1 cask` / `1 formula`. Update the test.

Note: the race-gate run overlapped with that throwaway test file in `internal/gui`. The file only logged values and asserted nothing. It was removed before `make cover-check` finished, and the worktree is clean.
