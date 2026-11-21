# Receipt: lb20-cask-followups

## Commits

- AC-01 (LB-17): `c05ebf0` lb20 AC-01: installed casks carry full_token [LB-17]
- AC-02 (LB-19): `179adf2` lb20 AC-02: Search info asks brew for the row's kind [LB-19]
- Receipt, state row, question: `6ba4360` (this file trimmed in a later commit)

## AC-01 — full_token

- Read-only `brew info --json=v2 --cask claude-code`: `"full_token": "claude-code"`, no `full_name`. Matches fixture.
- Fails before fix (`git show HEAD:` swapped in):
  `casks_token_test.go:96: FullName = "", want the full_token "claude-code"`
  `tap_untap_test.go:155: confirm lacks "gcloud"`
- Fix: `caskJSON.FullToken`; FullName = full_token, else full_name.
- Goldens changed: none.
- Trust-menu part: not done as written (Q-lb20-trust-menu).

## AC-02 — info kind flag

- Fails before fix:
  `search_info_test.go:225: row 0 (formula=true): brew args [info --json=v2 same], want [info --json=v2 --formula same]`
  `search_info_test.go:225: row 1 (cask=true): brew args [info --json=v2 same], want [info --json=v2 --cask same]`
- Fix: `fetchSelectedSearchInfo` adds `--cask`/`--formula`; unknown row keeps no flag.
- Added `TestSearchInfoCaskRowParsesBrewCaskFixture`: passes.
- Frames changed: none.

## Gate

- `make test-fast PKG="./internal/brew/... ./internal/gui/..."`: ok after each AC.
- `make ci`: exit 0; all packages ok; all coverage floors met (gui 90.7 ≥ 83, brew 87.8 ≥ 78).
- `GOOS=darwin go build ./...`: ok.
- No floors touched, no new deps, no TUI, no brew writes.

## Questions

- `Q-lb20-trust-menu`: the trust menu reads no installed packages. Tested the untap confirm instead. Ruling needed.

## Seen, not done

- `work/backlog.md` still lists LB-17 and LB-19 as parked; orchestrator to close.
- `work/STATE.md` rows lb12, lb16, lb18 still say `implemented`; not mine.
