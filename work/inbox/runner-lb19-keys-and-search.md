# Receipt: runner lb19-keys-and-search

## Commits (one per AC)
- AC-01 cee6eb5 Tab/Shift+Tab cycle past focused search
- AC-02 e00a573 Search rows: real type and installed state
- AC-03 3b0677a h/l switch tabs; [ ] and PgUp/PgDn removed
- AC-04 28aa5a4 Search uninstall (x, X on cask)

## AC-01
- Red: `search_tab_test.go:19: Tab from focused Search: activePanel = Search, want PanelStatus`.
- Green: `go test ./internal/gui/ -run TestTabCycle` ok.
- Weak: `flows.TestSearchTabCycleFlow` checks "Services", always in the sidebar.

## AC-02
- Red: `search_installed_test.go:72: row "claude-code  cask    ", want prefix "claude-code  cask  installed  "`.
- Two typed reads (`search --formula`, `--cask`); `brew.MarkInstalled` by name/token.
- Unreadable lists: `TestSearchInstalledListUnreadableKeepsTypes` ok. Changed mock only: `TestSearchServiceSearch`.

## AC-03
- Tests: `TestRetiredTabAndPageKeysDoNothing` (keymap_test.go), `TestHLSwitchTabs`, `TestHLTypeIntoSearch`.
- No red run against old code for this AC.
- Changed expectations: footer `[ ] tabs` -> `h/l tabs`; help loses PgUp/PgDn; `TestScrollKeysAreInTheHelpAndFooter`; footerExempt; services footer-agreement skip.
- PgDn tests -> End or wheel; `BenchmarkFilesTabPageDown` -> `BenchmarkFilesTabWheelDown`.
- `pageRows` removed; no test deleted. SMOKE-CHECKLIST 31, 34, 39, 69, 94; README.

## AC-04
- Red: `x opened no confirm`; `toasts = [], want ripgrep is not installed`.
- Green: `go test ./internal/gui/ -run TestSearch` ok.
- Writes: `uninstall --cask claude-code`, `uninstall jq`, `uninstall --zap --cask claude-code`.
- i/I on installed row: toast "already installed, x to uninstall", nothing runs.
- Success re-marks rows (`remarkSearch`). Footer shows `x uninstall`; X help-only.
- No golden pinned Search footer text. README added.

## Gate
- `make ci`: exit 0. gui 90.6% (floor 83), brew 87.2% (floor 78).
- `GOOS=darwin go build ./...` ok; `make test-fast` passed per commit. `make test-integration` not run (needs Homebrew).

## Seen, not done
- Success path re-reads installed lists after each successful task while Search has rows (cached).
- AC-01 flow assertion is weak; a footer-level check would fix it.

## Fix r2 (findings F-r1-01 … F-r1-06, one commit each)
- F-r1-01 85290cd: `TestSearchTabCyclesAllPanelsFlow` (flows). Red with the AC-01 tab cases removed from keys.go: `after 2 keys: main title is not Status`. Green with the fix; 3/3 on repeat. Per-key `View()` is read from a teatest program run per key prefix: the renderer writes cursor-positioned diffs, so a panel title is never contiguous in the output stream. `TestSearchTabCycleFlow` kept as is.
- F-r1-02 b0039d1: `TestSearchMarkClearsAfterUninstall` runs the batch its TaskCompletedMsg update returns. Red with `remarkSearch()` removed from it: `the row still says installed after the uninstall`.
- F-r1-03 4f70534: `TestStaleSearchMarksDoNotLandOnNewRows`. Red without the name guard: `row "wget" became installed from marks for another search`.
- F-r1-04 4c78f98: `TestRemarkKeepsMarksOfUnreadableList`. Red before the fix: `formula list unreadable: jq Installed = false, want true`. `markInstalled` keeps the marks of an unreadable type; doc comments in tasks_packages.go and brew/search.go corrected.
- F-r1-05 ee85b8f: `TestSearchKeepsTheReadThatWorked` (ruling Q-lb19-search-partial). Red before the fix: `--cask failed: Search err = test error, want the other read's rows`. Both reads run; error only when both fail (the first read's error is returned).
- F-r1-06 42696cb: `TestRetiredTabAndPageKeysDoNothing` extended to the open help screen. Red with a `pgdown` case in helpKey: `PgUp/PgDn scrolled the help screen: offset 2->7`.
- Gate: `make ci` exit 0 (gui 90.6% ≥ 83, brew 87.8% ≥ 78, all floors met). `GOOS=darwin go build ./...` ok. `make test-fast` passed per commit.
- No brew write commands run; the TUI was not launched. STATE.md row not changed (orchestrator's after re-review).
