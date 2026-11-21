verdict: ok

# Review — lb12-cask-token, round r1

Range reviewed: `a0aaecc..HEAD` (`93d0659` AC-01, `6a8b551` AC-02, plus `work/`-only commits).

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 parseCask reads `token` | met | `caskJSON.Token`; `Name` = token, falls back to `name`; `DisplayName` added to `brew.Cask` (`json:"-"`). Fixture `testdata/casks_installed_token.json` matches the live shape. With `casks.go` reset to `a0aaecc` in a temp worktree: `TestCaskNameIsTokenInRealBrewShape` FAIL `Name = "Claude Code"`, `TestCaskNameFallsBackToDisplayName` FAIL `DisplayName = ""`, `TestOutdatedAndInstalledCaskNamesAgree` FAIL, `TestCaskDependsOnRealBrewShape` FAIL. |
| AC-02 consumers use the token | met | Same revert: `TestCaskInstalledMapKeysByToken` FAIL (`keys = map[Claude Code:…]`). `TestCaskUpgradeSendsTokenFromBrewPayload` with the confirm-text check turned off fails **on the args check alone**: `expected brew args [upgrade --cask claude-code], got [[upgrade --cask Claude Code]]`. With `formatters.go` reset as well, `TestCaskInfoShowsDisplayNameBesideToken` FAILs. |

## Verify

- **V1 gate:** `make ci` exit 0 (gofmt, vet, `-race`, floors: brew 87.0 ≥ 78, gui 89.9 ≥ 83, presentation 100.0 ≥ 96, the rest met, build). `GOOS=darwin go build ./...` exit 0.
- **V2 regressions:** every new test fails with the fix reverted (table above). The open point in the receipt is closed: the brew-args assertion fails by itself.
- **V3 real shape:** live `brew info --json=v2 --installed --cask` has `token`, `full_token`, `name` (a list), `tap`, and **no `full_name`**, for `claude-code` ("Claude Code") and `gcloud-cli` ("Google Cloud CLI"). Live `brew outdated --json=v2 --cask` has `name: "claude-code"`. Driving the live payloads through `fetchPanelData`, the Casks panel lists `claude-code` and `gcloud-cli`, the tokens of every installed cask.
- **V4 consumers:** `sidebar_rows.go` rows, `panel.go` `nameAt`/`selectedCask`, `commands.go` items for Casks and Outdated, `depviews.go` `indexInstalled`, `tasks_packages.go` pkgInfo (search/deps), `details.go` Outdated details, `installedFromTap` (it returns `c.Name`, now the token), upgrade/uninstall/reinstall/zap in `casks.go` all take `Name`. Cask trust takes the tap's `CaskNames`, not `Cask`. No display name reaches brew or keys a map.
- **V5 frames:** my own harness (package `gui` test; live payloads, mock runner) rendered `View()` at `a0aaecc` and at HEAD, 80x24 and 140x40. Diffs: Casks rows `Claude Code`→`claude-code` and `Google Cloud CLI`→`gcloud-cli`; the Casks Info pane `Name:` row shows the token, with a new `Display:` row below it; the Search info `Name:` shows the token. Outdated panel: no change except the Casks rows in the sidebar. Nothing else changed.
- **V6 invariants:** one commit per AC, in order, with `[LB-12]`; no `program.Send`, no goroutines, no write path added; floors and Makefile untouched; no new dependency.

## Findings

None.

## Notes (not findings; for the orchestrator)

- The receipt's "seen, not done" is right and is a real bug from before LB-12: installed casks get an empty `FullName`, because brew writes `full_token`, not `full_name` (confirmed live). So `installedFromTap` never lists casks, and the untap warning misses them. This needs a backlog entry.
- `work/inbox/runner-lb12-cask-token.md` already had uncommitted changes when this review started. They are not mine, I left them alone, and they are not staged.
- Scratch: `/home/agent/.cache/lb12-r1/` (ci.log, frames `out-wt-pre/` and `out-wt/`, harness).
