verdict: ok

# reviewer lb27-status-tabs r2 (fix commits b548db4..afccf07)

| Finding | Commit | Verdict |
|---|---|---|
| F-r1-01 (high) exit 1 without report | b548db4 | fixed |
| F-r1-02 (medium) tab row at 80x24 | b450312 | fixed (spacing, no abbreviation) |
| F-r1-03 (low) timeout/running/y tests | 05f7aa5 + 1610d81 | fixed |
| F-r1-04 (low) smoke checklist names | d99091e | fixed |

| AC | Verdict |
|---|---|
| AC-01 Missing tab | ok (unchanged since r1) |
| AC-02 Vulns tab | ok |
| AC-03 d/m/v, runners gone | ok |

## Evidence
- Gate: `GOFLAGS=-p=1 make ci` exit 0 (gui 91.7 ≥ 83, all floors met, no floor file touched); `GOOS=darwin go build ./...` ok. Per commit, `make test-fast PKG="./internal/gui/... ./internal/brew/..."` ok at b548db4, b450312, 05f7aa5, d99091e, 1610d81 (detached worktree).
- F-r1-01, real outputs through `DefaultRunner` (fake brew scripts, scratch test):
  - real `brew vulns` (run once this round): exit 1, 52 lines on stdout, stderr empty → Done, report shown wrapped at 80 and 140, `y` copy equals the report byte-for-byte.
  - unknown command: real `brew nosuchcmd` stderr (usage text, then `Error: Invalid usage: Unknown command: brew nosuchcmd`) with `vulns` substituted, empty stdout, exit 1 → Failed; tab shows `Invalid usage: Unknown command: brew vulns` plus the `Press v…` hint at 80 and 140.
  - revert `internal/brew/doctor.go` to b548db4^: `TestDiagnosticsVulnsExitCode1WithoutReportIsError`, `…ErrorLineIsError` and `TestVulnsResultShapes/exit_1_no_report` fail; the scratch unknown-command case goes back to Done, `No known vulnerabilities`.
- F-r1-02: tab rows dumped at 80x24 and 140x40, dark and light, each of the five tabs active: `Dashboard  Config  Doctor  Missing  Vulns` whole every time; the underline is under the active name (Vulns: 7 cells at col 37). Formulae/Casks/Outdated/Taps/Services rows keep their old gap (`Info    Deps    Used By …`). Mutations: `tabGap` always 1 → `TestStatusTabRowShowsEveryName` fails; always 0 → `TestOtherTabBarsKeepTheirGap` fails. Mouse: clicking the middle of each Status tab at 80 cols selects that tab (`handleTabBarClick` uses the same gap).
- F-r1-03 mutations: drop the hint from the failed view → `TestVulnsTimeoutShowsErrorAndHint` fails; drop `Checking…` → `TestVulnsRunningFrameFitsAt80x24` fails; hard-cut report lines to 60 cols (`cutLines`) → `TestYCopiesTheVulnsReportWhole` fails. (The vulns view ignores width, so `copyWidth` mutations do not reach it; by design.)
- F-r1-04: all seven cited tests exist (`status_keys_test.go`, `status_missing_test.go`, `status_vulns_test.go`, `brew/doctor_test.go`, `flows/status_tabs_test.go`).
- No `go func`/`program.Send` added. go vet clean (in ci). staticcheck not rerun this round (not installed; `go run …@latest` denied by the session's permission check). The receipt says no new hit.

## Findings
None.

## Notes (no fix required)
- `doctor.go:193`: exit 1 with a report **and** an `Error:` stderr line drops the report, as the fix brief asks. The real run has empty stderr, so the normal path is unaffected.
- `SMOKE-CHECKLIST.md:132` `d` line still says "output in modal / status path" (no overlay now). Wording only; flagged in the receipt too. Worth fixing in a later slice.
- Status tabs are also tighter at 140 cols (single gap set for all widths). Accepted under F-r1-02.
