# Fix brief — lb32-vulns-view, round r2 (orchestrator-verified)

Findings: `work/inbox/reviewer-lb32-vulns-view-r1.md` (read in full). Tests first, one commit per
finding, subjects end `[LB-30] F-r1-0N`.

- **F-r1-01 (medium):** exit 1 + unparseable or empty stdout → the exit error (failure body with
  brew's first error line + the retry hint); exit 1 + JSON → the report; exit 0 + unparseable → the
  text fallback. Tests for all three; change `TestDiagnosticsVulnsExitCode1WithReportIsResult` to the
  new rule and name it in the receipt.
- **F-r1-02 (low):** `2.0.0` and `v2.0.0` rank later than `2.0.0-rc1` in `versionAfter`; add the two
  cases from the review to `TestVersionAfter`.
- **F-r1-03 (low):** `TestVulnsReportColours` asserts the HIGH chip and HIGH table cell in
  `style.ErrorText` as well.

Gate: `GOFLAGS=-p=1 make ci`. Never launch the TUI; do NOT run `brew vulns`. Append "Fix r2" (≤ 6
lines) to `work/inbox/runner-lb32-vulns-view.md`, commit, do not edit afterwards; reply with one done
line.
