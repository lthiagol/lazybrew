verdict: changes-required

# Review r1 — lb32-vulns-view (392a00b..f299494)

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 typed report + fallback | changes-required | Python on `vulns-real.json` (= `testdata/vulns_v1.json`, `cmp` same): 7 pkgs, 24 vulns, 90 checked, 37 skipped, 14 patched (libssh2), HIGH 8 / MEDIUM 5 / UNKNOWN 11 — tests match. Mutations: `ReleasedFix` → `v != ""` fails `TestReleasedFix`; `Parsed=false` fails `TestVulnsReportRealFixture`, `…EmptyFindings`. Exit-1 rule wrong: F-r1-01. |
| AC-02 report view | ok | Own harness, dark/light/dracula/nord × 80x24/100x30/140x40: frames exact size; chips HIGH=error, MEDIUM=warning, UNKNOWN=muted by `Render` match in all 4 themes. Raw view at content widths 47–140: no line over width, every summary word present. Sort matches Python (libsndfile, openssl@4, libxml2, ffmpeg, expat, gitleaks, libssh2). Footer muted, skipped names only in copy. No-findings and fallback views drawn. |
| AC-03 next steps, y, docs, flow | ok (one low) | Own cases: `1.2.3`→`fix in 1.2.3`; `1.9`,`1.10`→`fix in 1.10`; 40-hex→`fixed upstream, not released` (muted); `[]`/nil→`no fix yet`; SHA+`2.0`→`fix in 2.0`; Outdated loaded→`upgrade → 9.9` success; not loaded→`fix in 1.0`; all-UNKNOWN muted. `y` copy: 51 lines, no ANSI, full SHAs, summaries whole, patched IDs, 37 skipped names. Pre-release order: F-r1-02. |

## Gate
- `GOFLAGS=-p=1 make ci` → exit 0 (fmt, vet, `test -race`, floors: gui 92.3≥83, brew 90.6≥78, others unchanged; build).
- `GOOS=darwin go build ./...` ok. deadcode `./cmd/...` empty. staticcheck: 5 findings, identical on base `ded9e17` (keymap_test, launch_check_test).
- Regressions: own harness on base vs HEAD, keys `1 d m c A 2..7 1 v` at 80x24/140x40/70x20: all 39 frames byte-identical. `TestVulnsTabOpenRunsNothing`, `…RunsOnceWhileRunning`, `…SurvivesRefreshAndPanelSwitch` pass. Scroll: `G` reaches footer (`TestVulnsReportScrolls`). Running/not-run/error frames unchanged.
- One commit per AC, in order; no goldens changed; floors not lowered; no new `go func`/`program.Send`.

## Findings

F-r1-01 (medium) — `internal/brew/doctor.go` `VulnsReport` — exit 1 with stdout that is not JSON (and no `Error:` line) is returned as a Done text fallback. The brief (Behaviour): "exit 1 with no/invalid JSON is the failure body with brew's first error line"; AC-01: "exit 1 + no JSON = error; invalid JSON = typed fallback" (i.e. on exit 0). Evidence: own frame, mock `{"findings":[` + exit 1 → `Could not read brew's JSON; showing its text.`; `TestDiagnosticsVulnsExitCode1WithReportIsResult` (doctor_test.go) asserts the text fallback for `CVE-2026-1: openssl@3` + exit 1. Fix must achieve: exit 1 + unparseable/empty stdout → the exit error (failure body), exit 1 + JSON → report, exit 0 + unparseable → text fallback; tests for all three, the old test changed to the new rule and named in the receipt.

F-r1-02 (low) — `internal/gui/vulns_report.go` `versionAfter` — a pre-release outranks its release: fixes `["2.0.0-rc1","2.0.0"]` → `fix in 2.0.0-rc1`; `["v2.0.0","2.0.0-rc1"]` → `fix in 2.0.0-rc1`. Fix must achieve: `2.0.0` (and `v2.0.0`) is later than `2.0.0-rc1`; cases added to `TestVersionAfter`.

F-r1-03 (low) — `internal/gui/vulns_report_test.go` `TestVulnsReportColours` — asserts CRITICAL, MEDIUM, LOW, UNKNOWN but not HIGH (error role), which AC-02 requires ("assert SGR per role"). Fix must achieve: the HIGH chip and HIGH table cell asserted in `style.ErrorText`.

## Notes (no change required)
- A 7-hex short SHA (`abc1234`) reads as released (`fix in abc1234`); it follows the brief's 40-hex rule. brew gave only 40-hex in the real data.
- At 80x24 the content column is 48 wide; the 22-cell word `wavlike_subchunk_parse` is split across two lines in a 21-cell summary column (by design for words wider than the column). Below 47 the table would overflow, but the minimum terminal is 80x24 (70x30 shows the "too small" screen).
- libsndfile: 1 of 6 advisories has a SHA fix, so the package reads `fixed upstream, not released`; that is the brief's rule.
- Five real advisories have empty summaries (libsndfile); the rows show ID and severity only.
