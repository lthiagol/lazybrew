# Receipt — runner lb12-cask-token

Commits, one per AC:
- AC-01 `93d0659` parseCask reads token [LB-12]
- AC-02 `6a8b551` cask consumers use the token [LB-12]

## AC-01 — parseCask reads token

Failing first (`DisplayName` field added alone, parser unchanged):

```
--- FAIL: TestCaskNameIsTokenInRealBrewShape
    casks_token_test.go:26: Name = "Claude Code", want the token "claude-code"
--- FAIL: TestCaskNameFallsBackToDisplayName
    casks_token_test.go:46: DisplayName = "", want "Google Chrome"
```

Fix: `caskJSON.Token`; `Name` = token, else first display name; `DisplayName` = first display name. Both pass after.
Fixture: `internal/brew/testdata/casks_installed_token.json` (new).
`casks_depends_test.go` keys changed from display names to `gcloud-cli` / `claude-code`; assertions unchanged.

## AC-02 — consumers

- Rows, `nameAt`, confirms, upgrade/uninstall/zap/fetch all read `Name`, now the token.
- `depviews.indexInstalled` keys on the token, matching `depends_on`.
- `FormatCaskInfo`: Name row = token; "Display" row only when `DisplayName` differs.
- Outdated ↔ Casks: no join in code; both carry the token.
- Gui test `TestCaskUpgradeSendsTokenFromBrewPayload` asserts `upgrade --cask claude-code`.

Failing run, `parseCask` mutated back to `string(c.Name)`:

```
--- FAIL: TestCaskUpgradeSendsTokenFromBrewPayload
    confirm must name the token, got view: ... Upgrade Claude Code? ...
```

Restored (`cmp`), passes. The run stopped before the brew-args check.
Goldens: none changed.

## Gate

- Per commit: `make test-fast` passed before each commit.
- `make ci`: exit 0. Coverage brew 87.0 ≥ 78, gui 89.9 ≥ 83, presentation 100.0 ≥ 96.
- `GOOS=darwin go build ./...`: ok.
- `git diff --shortstat a0aaecc 6a8b551`: 8 files, 248+, 16-.
- No brew write command run.

## Questions filed

None.

## Seen, not done

- Installed casks have empty `FullName`: payload has `full_token`, not `full_name`, so `installedFromTap` never sees them. Predates LB-12.
- Outdated cask rows have no `DisplayName`.
- Orchestrator: strike LB-12 in `work/backlog.md`.
