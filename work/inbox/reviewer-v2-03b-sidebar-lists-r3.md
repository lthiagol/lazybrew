verdict: ok

# reviewer r3 — v2-03b-sidebar-lists (reviewed `2ea6a0d..6368ac7`: `3fefcf8`, `9c91b5d`, `b501d1f`, frames `9deb75e`)

## Verify

- **Gate** (HEAD `6368ac7`): all four commands green.
  - `make build` ok; `make lint` ok.
  - `CGO_ENABLED=1 go test ./... -race -count=1`: every package ok.
  - `make cover-check`: brew 72.9, presentation 96.2, gui 78.1, modal 50.5; all floors met.
  - Each fix commit (`3fefcf8`, `9c91b5d`, `b501d1f`) passes `go build ./...` and `go test ./internal/gui/...` in a detached worktree.
- **Frames**: re-ran the committed harness against real Homebrew.
  - Every `.txt` frame at both sizes and every 140×40 `.ansi` frame matches `after/` byte-for-byte, and each is exactly W×H.
  - Only 08-help changed this round, as the `result.md` "Fix round r2" section says.
  - `internal/gui/dump` removed; worktree clean.
- **Invariants**: no change to `internal/brew/**`. No test removed; the top-level `func Test` count went 484 → 487. One commit per finding, each subject naming its id.

## Type-to-search probe

I ran a throwaway test, deleted after the run, at `fe12a58` (the slice start) and at HEAD. For each of the 94 printable ASCII keys (`!`..`~`) it pressed `7`, then `Esc` (Search panel, input unfocused), then the key, and recorded the input's focus, its value, and the active panel.

- The two outputs are **identical**. `g` → focused=true value="g"; `G` → focused=true value="G"; `z` → focused=true value="z".
- The only key cases this slice adds (`git diff fe12a58..HEAD`, non-test) are `pgup`, `pgdown`, `home`, `end` and `g`/`G`. The first four have no `Text`, so they can never reach type-to-search. On Search, the `g`/`G` case does nothing and falls out of the switch into the type-to-search block.

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r2-01 (major) | fixed | `gui.go` `case "g", "G"` is guarded by `m.activePanel != PanelSearch`. `TestLetterKeysOnUnfocusedSearchTypeToSearch` and `TestScrollLettersOutsideSearch` cover it, along with the probe above. The help row reads `Home/End  Top/bottom (g/G outside Search)` (frame 08-help, 140×40, row 10), and `TestScrollKeysAreInTheHelpAndFooter` was updated. |
| F-r2-02 (minor) | fixed | `plural()` in `details.go`. `TestTrustStoryPluralises` covers 1 formula/0 casks, 2 formulae/1 cask and 0/0. The Trust tab test now expects `1 cask`. |
| F-orch-01..03, F-r1-01..03 | fixed (r2) | unchanged since r2 apart from the g/G guard above. |

## Findings

None.

Note, not a finding: at 80×24 the help's `(g/G outside Search)` wraps onto its own row in the narrow column. `result.md` explains this, and the help's layout at that size is Q-v2-help / v2-05.
