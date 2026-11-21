# Receipt — runner lb12-cask-token

**Commits** (one per AC, in order)
- AC-01 `93d0659` lb12 AC-01: parseCask reads token, display name kept apart [LB-12]
- AC-02 `6a8b551` lb12 AC-02: cask consumers use the token; display name beside it [LB-12]

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

Fix: `caskJSON.Token`; `Name` = token, else the first display name; `DisplayName` = first display name. Both tests pass after it.
The no-token case's `Name` assertion already passed before the fix; only its `DisplayName` half failed.
Fixture: `internal/brew/testdata/casks_installed_token.json` (new, one claude-code entry in brew's current shape).
`casks_depends_test.go` keyed its lookups by display name. I changed the keys to `gcloud-cli` / `claude-code`; the assertions are unchanged.

## AC-02 — consumers

Every reader of `Cask.Name` now gets the token:
- Casks and Outdated rows, `nameAt`/`selectedName`, confirms, upgrade/uninstall/zap/fetch/reinstall: all read `Name`.
- `depviews.indexInstalled` keys on the token, matching the tokens in `depends_on` (`TestCaskInstalledMapKeysByToken`).
- `installedFromTap` and the `ParseInfo` search-info path: token.
- `FormatCaskInfo`: Name row = token. A "Display" row shows `DisplayName` only when it differs from `Name`.
- Outdated ↔ Casks: no join exists in code. Both lists now carry the token (`TestOutdatedAndInstalledCaskNamesAgree`).

Failing run on the fixed-up tree, mutating `parseCask` back to `name := string(c.Name)`:

```
--- FAIL: TestCaskUpgradeSendsTokenFromBrewPayload
    cask_token_test.go:56: confirm must name the token, got view: ... Upgrade Claude Code? ...
```

The file was restored byte-for-byte (`cmp`) and the test passes again. The run stopped at the confirm-text check, before the brew-args check, so the args assertion itself was not shown failing.

Goldens: none changed. `git diff --name-only a0aaecc 6a8b551` lists no golden or snapshot file. Test fixtures use struct literals, so no existing frame carried a token that differs from its name. The Casks list shows the token for real payloads only.

## Gate

- Per commit: `make test-fast` (all packages, no `-race`) passed before AC-01 and before AC-02.
- `make ci`, exit 0: gofmt check, `go vet`, `go test ./... -race -count=1`, coverage floors all met (brew 87.0 ≥ 78, gui 89.9 ≥ 83, presentation 100.0 ≥ 96), `go build`.
- `GOOS=darwin go build ./...`: ok.
- `git diff --shortstat a0aaecc 6a8b551`: 8 files changed, 248 insertions(+), 16 deletions(-).
- No brew write command was run. `bin/lazybrew` was built by `make ci` and not run.

## Questions filed

None.

## Seen, not done

- Installed casks have an empty `FullName`. The `brew info` payload has `full_token`, not `full_name`, so `installedFromTap` never sees installed casks in the tap-trust menu. This predates LB-12 and is outside its scope.
- Outdated cask rows carry no `DisplayName`, because the outdated payload has only the token. Their details show the token alone.
- LB-12 in `work/backlog.md` is closed by this slice. The orchestrator should strike it.
