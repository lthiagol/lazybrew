# Receipt: lb20-cask-followups

## Commits

- AC-01 (LB-17): `c05ebf0` lb20 AC-01: installed casks carry full_token [LB-17]
- AC-02 (LB-19): `179adf2` lb20 AC-02: Search info asks brew for the row's kind [LB-19]
- Receipt, state row, question: this commit (`work/inbox/runner-lb20-cask-followups.md`,
  `work/STATE.md`, `work/questions.md`)

## AC-01 — installed casks carry full_token

- Shape check, read-only: `brew info --json=v2 --cask claude-code` has
  `"full_token": "claude-code"` and no `full_name`; the fixture matches.
- Failing before the fix (fix swapped out with `git show HEAD:`):
  ```
  casks_token_test.go:96: FullName = "", want the full_token "claude-code"
  casks_token_test.go:112: FullName = "", want the full_token "user/tap/bar"
  tap_untap_test.go:155: confirm lacks "brew will refuse to untap while they are installed"
  tap_untap_test.go:155: confirm lacks "gcloud"
  ```
- Fix: `caskJSON.FullToken`; `FullName` = full_token, else full_name.
- Passing after: `go test ./internal/brew/ -run TestCaskFullName` ok; gui untap test ok.
- Goldens changed: none (`FullName` is rendered nowhere except `installedFromTap`).
- Trust-menu part: not done as written. See Q-lb20-trust-menu.

## AC-02 — Search info passes the row's kind

- Failing before the fix:
  ```
  search_info_test.go:225: row 0 (formula=true cask=false): brew args [info --json=v2 same], want [info --json=v2 --formula same]
  search_info_test.go:225: row 1 (formula=false cask=true): brew args [info --json=v2 same], want [info --json=v2 --cask same]
  ```
- Fix: `fetchSelectedSearchInfo` adds `--cask` / `--formula` from the row; unknown row keeps no flag.
- Added `TestSearchInfoCaskRowParsesBrewCaskFixture`: the `--cask` payload
  (`casks_installed_token.json`) parses as type cask, name claude-code. Passes.
- Frames changed: none.

## Gate

- `make test-fast PKG="./internal/brew/... ./internal/gui/..."`: all ok (after each AC).
- `make ci`: exit 0. Output: all packages ok; coverage floors all met
  (app 70.0 ≥ 64, brew 87.8 ≥ 78, config 92.6 ≥ 85, gui 90.7 ≥ 83, anim 88.2 ≥ 86,
  modal 89.2 ≥ 88, presentation 100.0 ≥ 96, style 89.5 ≥ 88, task 96.2 ≥ 93,
  logtail 92.3 ≥ 90). Log: scratchpad `ci.log`.
- `GOOS=darwin go build ./...`: ok.
- Floors not touched. No dependency change. No TUI launched. No brew write commands.

## Questions filed

- `Q-lb20-trust-menu`: the brief asks that "the tap-trust menu lists" an installed
  cask. The trust menu reads no installed packages. I tested the untap confirm
  instead (`TestUntapConfirmListsInstalledCaskFromBrewPayload`). Ruling needed.

## Seen, not done

- `work/backlog.md` still lists LB-17 and LB-19 as parked. Left for the orchestrator to close.
- `work/STATE.md` has older rows at `lb12-cask-token`, `lb16-update-when-stale`,
  `lb18-config-sample` still marked `implemented`. Not mine; left alone.
