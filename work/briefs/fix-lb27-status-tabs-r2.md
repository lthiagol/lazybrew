# Fix brief — lb27-status-tabs, round r2

Findings: `work/inbox/reviewer-lb27-status-tabs-r1.md` (read in full). Tests first,
one commit per finding, subjects end with the finding id; every commit builds and
passes `make test-fast PKG=./internal/gui/... ./internal/brew/...`.

- **F-r1-01 (high):** exit 1 with no report on stdout, or an `Error:` line on
  stderr, is a failure: the Vulns tab shows brew's first error line (error style)
  and the retry hint; only a real report (exit 0, or exit 1 **with** report
  lines on stdout) is Done. `brew vulns` exits 1 when it finds vulnerabilities,
  so keep that path. Test with exactly `Error: Invalid usage: Unknown command:
  brew vulns` as a `BrewExitError` code 1: it fails before the fix.
- **F-r1-02 (medium):** all five Status tab names fully visible at 80x24 in both
  themes (shorter spacing between tabs, or abbreviate only if spacing is already
  minimal; say which). Frame test asserting `Vulns` (underlined when active) in
  the tab row at 80 columns, and that Formulae/other panels' tab rows are
  unchanged.
- **F-r1-03 (low):** tests for timeout → error line + hint, the running frame at
  80x24, and `y` copying the whole multi-line report ANSI-free at 80 cols.
- **F-r1-04 (low):** `SMOKE-CHECKLIST.md` names tests that exist
  (`TestStatusKeysOpenTheirTabs`, `TestStatusTabsFlow`).

Gate: `GOFLAGS=-p=1 make ci` before the receipt. Never launch the TUI. Append "Fix
r2" (≤ 8 lines) to `work/inbox/runner-lb27-status-tabs.md`, commit, do not edit
afterwards; reply with one done line.
