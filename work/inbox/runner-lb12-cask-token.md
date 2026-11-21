# Receipt — runner lb12-cask-token

**Commits** (one per AC, in order)
- AC-01 `93d0659` lb12 AC-01: parseCask reads token [LB-12]
- AC-02 `6a8b551` lb12 AC-02: cask consumers use the token [LB-12]

## AC-01 — parseCask reads token

Test first. Added only the `Cask.DisplayName` field, then ran the new tests on the unfixed parser:

```
$ go test ./internal/brew/ -run 'TestCaskName' -count=1
--- FAIL: TestCaskNameIsTokenInRealBrewShape (0.00s)
    casks_token_test.go:26: Name = "Claude Code", want the token "claude-code"
--- FAIL: TestCaskNameFallsBackToDisplayName (0.00s)
    casks_token_test.go:46: DisplayName = "", want "Google Chrome"
FAIL
```

Fix: `caskJSON.Token`; `Name` = token, else first display name; `DisplayName` = first display name. Both pass after.
Fixture: `internal/brew/testdata/casks_installed_token.json` (new).
`casks_depends_test.go` keyed by display name; keys changed to `gcloud-cli` / `claude-code`, assertions unchanged.

## AC-02 — consumers

- Casks and Outdated rows, `nameAt`, confirms, upgrade/uninstall/zap/fetch: all read `Name`, now the token.
- `depviews.indexInstalled` keys on the token, matching `depends_on` (`TestCaskInstalledMapKeysByToken`).
- `FormatCaskInfo`: Name row = token; "Display" row only when `DisplayName` differs.
- Outdated ↔ Casks: no join in code; both carry the token (`TestOutdatedAndInstalledCaskNamesAgree`).
- New gui test `TestCaskUpgradeSendsTokenFromBrewPayload` (asserts `upgrade --cask claude-code`).

Failing run, with `parseCask` mutated back to `name := string(c.Name)`:

```
--- FAIL: TestCaskUpgradeSendsTokenFromBrewPayload
    cask_token_test.go:56: confirm must name the token, got view: ... Upgrade Claude Code? ...
```

Restored (`cmp`), passes again. The run stopped at the confirm-text check, before the brew-args check, so the args check was not shown failing.

Goldens: none changed (`git diff --name-only a0aaecc 6a8b551` has no golden or snapshot file).

## Gate

- Per commit: `make test-fast` (all packages) passed before each commit.
- `make ci`: exit 0. Coverage: brew 87.0 ≥ 78, gui 89.9 ≥ 83, presentation 100.0 ≥ 96.
- `GOOS=darwin go build ./...`: ok.
- `git diff --shortstat a0aaecc 6a8b551`: 8 files, 248 insertions(+), 16 deletions(-).
- No brew write command run.

## Questions filed

None.

## Seen, not done

- Installed casks have empty `FullName`: payload has `full_token`, not `full_name`, so `installedFromTap` never sees them. Predates LB-12.
- Outdated cask rows have no `DisplayName` (outdated payload has only the token).
- LB-12 in `work/backlog.md` is closed by this slice; orchestrator to strike it.
