# Receipt: lb32-vulns-view (runner) — LB-30

Commits, one per AC: `392a00b` AC-01 typed report + text fallback;
`ba894cb` AC-02 report view; `a6f496f` AC-03 next steps, `y` copy, README, flow.

## Evidence

- **AC-01.** Red first (`undefined: VulnReport`). Green: `go test ./internal/brew/`
  → `ok`. Fixture `testdata/vulns_v1.json` (copy of `~/.cache/vulns-real.json`):
  `TestVulnsReportRealFixture` checks 7 findings, 24 vulns, 14 patched, 90 checked,
  37 skipped, HIGH 8 / MEDIUM 5 / UNKNOWN 11.
- **AC-02.** Red: 9 `TestVulnsReport*` failed on the placeholder view. Green:
  `go test ./internal/gui/ -run Vulns` → `ok`. Frames 80x24 and 140x40, both
  themes; colours by role in `TestVulnsReportColours`.
- **AC-03.** Red: `vulnNextStep undefined`; flow failed on its first wait. Green:
  `TestVulnsNextStep*`, `TestVersionAfter`, `TestVulnsCopyHoldsTheWholeReport`,
  `flows.TestVulnsReportFlow` (decodes the OSC 52 payload).

## Gate

- `GOFLAGS=-p=1 make ci` → exit 0 (fmt, vet, `test -race`, cover-check, build).
  Lowest margins: gui 92.3 ≥ 83, brew 90.6 ≥ 78.
- Per commit: `GOFLAGS=-p=1 go test ./internal/brew/... ./internal/gui/...` → `ok`.
- `GOOS=darwin go build ./...` ok; deadcode `./cmd/...` no output.
- staticcheck: 5 findings, none in touched files (`keymap_test.go`, `launch_check_test.go`).

## Decisions the brief left open

- Own table renderer: `renderTable` cuts and does not wrap.
- Summaries wrap at spaces only: `ansi.Wrap` and lipgloss break at hyphens.
- Next step is the phrase only; `(press 4, then u)` is not drawn.
- All-UNKNOWN package: whole phrase muted, `upgrade →` included.
- Copy: `fixed versions:` under each row; patched IDs and skipped names in footer.
- Scroll: existing G/g/home/end/wheel; `TestVulnsReportScrolls`.

## Test edits (no goldens changed, none deleted)

- `status_vulns_test.go`: two checks read `vulns.report.Text`.
- `doctor_test.go`: `vulnsOnce` returns `VulnReport`.

## Seen, not done

- `unscanned_formulae` / `outdated_without_sbom` shapes seen only as `{}` / `[]`.
- `complete` parsed, not drawn.

## Fix r2

- F-r1-01 `8d2a0f2`, `6ddde2f`: exit 1 is a report only with JSON. `TestDiagnosticsVulnsExitCode1WithReportIsResult` now uses JSON; new `TestDiagnosticsVulnsExitCode1WithTextIsError`; `TestVulnsResultShapes` follows the rule (`8d2a0f2` alone left that case red).
- F-r1-02 `d835549`: a release outranks its pre-releases; `2.0.0` and `v2.0.0` > `2.0.0-rc1`; a missing part counts as 0 (`2.0` = `2.0.0`). Cases in `TestVersionAfter`.
- F-r1-03 `db0ab4f`: `TestVulnsReportColours` asserts the HIGH chip and HIGH cell in `style.ErrorText`; a WarningText mutation fails it.
- Gate: `GOFLAGS=-p=1 make ci` exit 0 (gui 92.4 ≥ 83, brew 90.6 ≥ 78); `GOOS=darwin go build ./...` ok; deadcode empty; staticcheck: the same 5 findings as before.
