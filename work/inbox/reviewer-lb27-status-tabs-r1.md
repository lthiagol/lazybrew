verdict: changes-required

# reviewer lb27-status-tabs r1 (1416c23..48f5378)

| AC | Commit | Verdict |
|---|---|---|
| AC-01 Missing tab | 6ab7914 | ok |
| AC-02 Vulns tab | cdbcb03 + 3310921 | changes-required (F-r1-01, F-r1-02, F-r1-03) |
| AC-03 d/m/v, runners gone | f535d23 | ok (F-r1-04 low) |

## Evidence
- V1: `GOFLAGS=-p=1 make ci` exit 0 (gui 91.7 ≥ 83, all floors met, build ok); `GOOS=darwin go build ./...` exit 0. (A first run with TMPDIR under ~/.cache failed 3 Logs-tab path tests on the long path: environment, not the slice.)
- V2: real `brew missing` exit 0, empty stdout. Scratch test: exit 1 + `ffmpeg: libvpx x264\nnode: icu4c` → `⚠ ffmpeg: missing libvpx x264` / `⚠ node: missing icu4c`; `y` copies both rows. Reverting the `statusMissingTab` fetch case fails `TestStatusKeysOpenTheirTabs`, `TestMissingTabFetchesRows`, `…RealErrorShowsError`, `…LoadsOnceAndStays`, `…Frames`.
- V3: real `brew vulns` (run once): exit 1, 52 lines on stdout, stderr empty → Done; wrapped by word at 80 and 140; `copyText()` equals the whole report at both widths. Timeout (cap 50 ms, scratch): `brew vulns timed out after 50ms` + hint. Zero calls on open, one on v, none on R/panel switch: tests pass and the flow asserts it.
- V4: d/m/v switch tabs (scratch: d→2, v→4, m→3); flow cycles all five tabs. Dropping `stMissing` from the Status footer fails `TestPanelKeysAreOnFooterAndHelp`. deadcode clean; staticcheck base vs HEAD identical (pre-existing SA1019/SA4006 only).
- V5: own frame dump, base 1416c23 vs HEAD, panels 1–7 + help, 80x24 and 140x40, dark and light: only Status (tab row, tabs) and the help screen's three Status lines differ. Panel 1 tabs row at 80x24 — see F-r1-02.
- V6: no `go func`/`program.Send` added; vulns read goes through `boundedRead` (2 min); floors unchanged.

## Findings
- F-r1-01 (high) — `internal/brew/doctor.go:180` `Vulns` + `internal/gui/vulns.go:68` — exit 1 with an empty stdout is read as a clean run. A brew without `vulns` exits 1 with `Error: Invalid usage: Unknown command: brew vulns` on stderr (verified: `brew nosuchcmd` exits 1 the same way); scratch test with that `BrewExitError` gives status Done and the tab shows `No known vulnerabilities` in the success style. The brief says a missing command shows brew's first error line + the retry hint. — Fix must achieve: exit 1 with no report on stdout (or an `Error:` stderr) reaches the Failed state showing brew's error line and the hint; a test with exactly that error fails before the fix.
- F-r1-02 (medium) — `internal/gui/panel.go:51` tab row — at 80x24 the Status tab row reads `Dashboard    Config    Doctor    Missing    Vu`; with Vulns active the underline sits under `Vu`. The fifth tab's name is unreadable at the brief's own frame size, and `TestMissingTabFrames`/`TestVulnsTabFrames` only check frame size. — Fix must achieve: all five Status tab names fully visible at 80x24 (both themes), with a frame test asserting `Vulns` in the tab row at 80 cols.
- F-r1-03 (low) — `internal/gui/status_vulns_test.go` — no test for the timeout state, the running frame at 80x24, or `y` on the Vulns tab (receipt "seen, not done"); behaviour verified by hand above. — Fix must achieve: tests for timeout → error line + hint, and `y` copies the whole multi-line report ANSI-free at 80 cols.
- F-r1-04 (low) — `SMOKE-CHECKLIST.md:132-134` — cites the removed `TestRunDoctor`/`TestRunMissing`/`TestRunVulns`. — Fix must achieve: the checklist names tests that exist (`TestStatusKeysOpenTheirTabs`, `TestStatusTabsFlow`).

## Process (record only, no fix)
- cdbcb03 did not compile its tests (per-commit gate broken), repaired by 3310921; AC-02 was not written tests-first (receipt says so). AC-01..03 lack the trailer.
