verdict: ok

# reviewer: cr4-gui-dedup r3b (fresh session; base db7d464, HEAD 7de2010)
## Gate
- `make ci` exit 0 (gofmt, vet, race tests, 10 floors OK, build); `GOOS=darwin go build ./...` exit 0.
  deadcode/staticcheck not run (cannot be fetched here).
- Since r2 only keymap_test.go changed outside work/ (508df05). Goldens/testdata/flows: unchanged.

## Earlier findings — each fix reverted in a scratch worktree
| id | status | evidence |
|---|---|---|
| F-r1-01 | fixed | task_build.go from fe72f96^: TestLinesTaskCancelMidStream (`errCh delivered nothing within 2s`), ...CancelledBeforeStart, ...CancelThroughManager ×3 (`no TaskCompletedMsg within 2s`) FAIL; HEAD: pass, -race ×3 |
| F-r1-02 | fixed | M1 below passes with keymap_test.go from 187a8cb^, fails at HEAD |
| F-r2-01 | fixed | M6, M7 pass with keymap_test.go from 508df05^, fail at HEAD |

Mutations at HEAD (`go test ./internal/gui/ ./internal/gui/modal/`; each restored, `git status` clean):
| # | mutation | result |
|---|---|---|
| M1 | stDoctor off the Status footer | FAIL :357 Status "d" |
| M6 | keyNavigate off the Outdated footer | FAIL :357 Outdated "j" "k" "up" ... |
| M7 | keyPanelTabs off the Formulae footer | FAIL :357 Formulae "[" "left" "right" ... |
| M4/M5 | `n` off ConfirmModal.Hints / `j` off MenuModal Keys | FAIL :524 (+ look_test.go:367) |
| M8/M9 | keyNavigate off tapsHints / tRepair off tapActionHints | FAIL :357 Taps |

## footerExempt against today's footers
- Matches: Status's panelHints has neither keyNavigate nor keyPanelTabs, so it exempts j/k/down/up
  and [ ]/left/right. Outdated, Taps and Search show keyNavigate but not keyPanelTabs, so they exempt
  [ ]/left/right only. Formulae, Casks and Services (servicesHints) show both and have no entry.
  The comments are accurate (Search: panelTabs[PanelSearch] is empty).
- Minimal and live: drop Search's entry or Status j/k → :357 fails; stale `u` on Outdated → :365;
  `d` on Taps → :362. The omission itself is LB-13 (ruled out of scope), not a finding.

## Frames and brew args (own harness, gui.New/Update/View over a replay runner)
- Reads recorded once from real read-only brew calls, then replayed; every other call recorded and
  faked, never run. services info's log paths point at a frozen copy of moshi-hook.log.
- 849 frames per theme×size (dark+light × 80x24+140x40 = 3396): panels 1-7, every tab, j/k, help +
  paging, log toggle, tab/shift-tab, every panel key's modal answered y/n/esc, the overlay, tap-add
  input, T menu, the four Brewfile entries, search + i/I/F, and R with update_on_start.
- Brew args: 61 write calls per theme×size, identical at base and HEAD (install, uninstall, zap, reinstall,
  upgrade cask/batch, fetch, pin, (un)tap, repair, (un)trust, services, cleanup, autoremove, bundle, vulns, update).
- Frames: the Log box's row order differs from run to run at base too, because the client loads in
  parallel (base-vs-base gives 1989 raw diffs). With the Log box's text masked (other rows compared
  byte for byte, ANSI included), base vs HEAD gives 0 diffs, except R-update 02/04. In those frames
  the "1 outdated" count lands or does not; base-vs-base shows the same split, and with no command
  dropped, base lands it 5/6 runs and HEAD 6/6. That is a pre-existing race between `brew update` and
  the R refresh, not a cr4 change. The header line 0 matched in every frame (all "0s ago").

## New problems
None. Seen, not a finding (dead before cr4, cr5 scope): panelHints' PanelTaps/PanelServices cases
are never shown (footerHints routes them to tapsHints/servicesHints, also at db7d464).
